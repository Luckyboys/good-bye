package email

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/Luckyboys/good-bye/src/config"
	"github.com/Luckyboys/good-bye/src/state"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
	"github.com/sirupsen/logrus"
)

// Service 邮件服务
type Service struct {
	config       *config.Manager
	state        *state.Manager
	logger       *logrus.Logger
	retryManager *RetryManager
	sendFunc     func(Message) *Result
}

// NewEmailService 创建新的邮件服务
func NewEmailService(cfg *config.Manager, stateMgr *state.Manager, logger *logrus.Logger) *Service {
	service := &Service{
		config: cfg,
		state:  stateMgr,
		logger: logger,
	}

	service.retryManager = NewRetryManager(logger)
	service.retryManager.SetSendFunc(service.doSendEmail)
	service.sendFunc = service.doSendEmail

	return service
}

// Message 邮件消息结构
type Message struct {
	To      string
	Subject string
	Content string
	IsHTML  bool
}

// Result 邮件发送结果
type Result struct {
	Success bool
	Message string
	Error   error
}

// SendTestEmail 发送测试邮件
func (es *Service) SendTestEmail() *Result {
	testEmail := es.config.GetString("email.test_email")
	if testEmail == "" {
		return &Result{
			Success: false,
			Message: MessageTestEmailNotConfigured,
			Error:   errors.New(ErrorTestEmailNotConfigured),
		}
	}

	message := Message{
		To:      testEmail,
		Subject: SubjectTestEmail,
		Content: es.generateTestEmailContent(),
		IsHTML:  true,
	}

	return es.sendEmail(message)
}

// SendReminderEmail 发送提醒签到邮件
func (es *Service) SendReminderEmail(reminderTime time.Duration, willSendTime time.Time) *Result {
	// 获取测试邮箱地址
	testEmail := es.config.GetString("email.test_email")
	if testEmail == "" {
		return &Result{
			Success: false,
			Message: MessageTestEmailNotConfigured,
			Error:   errors.New(ErrorTestEmailNotConfigured),
		}
	}

	message := Message{
		To:      testEmail,
		Subject: SubjectReminderEmail,
		Content: es.generateReminderEmailContent(reminderTime, willSendTime),
		IsHTML:  true,
	}

	return es.sendEmail(message)
}

// SendWillMessage 发送遗书邮件（从文件读取内容）
func (es *Service) SendWillMessage() *Result {
	// 从文件读取遗书内容
	content, err := es.state.ReadPosthumousPapers()
	if err != nil {
		es.logger.WithError(err).Error("Failed to read posthumous papers file")
		return &Result{
			Success: false,
			Message: MessageReadPosthumousFailed,
			Error:   err,
		}
	}

	// 获取收件人邮箱列表
	recipients := es.getWillRecipients()
	if len(recipients) == 0 {
		return &Result{
			Success: false,
			Message: MessageWillRecipientsNotConfigured,
			Error:   errors.New(ErrorNoRecipientsConfigured),
		}
	}

	// 处理遗书内容：如果是Markdown则转换为HTML
	emailContent := es.processPosthumousContent(content)

	// 发送给所有收件人
	var lastError error
	successCount := 0

	for _, recipient := range recipients {
		message := Message{
			To:      recipient,
			Subject: SubjectWillEmail,
			Content: emailContent,
			IsHTML:  true,
		}

		result := es.sendEmail(message)
		if result.Success {
			successCount++
		} else {
			lastError = result.Error
			es.logger.WithError(result.Error).WithField("recipient", recipient).Error("Failed to send will email")
		}
	}

	if successCount == 0 {
		return &Result{
			Success: false,
			Message: MessageAllEmailsFailed,
			Error:   lastError,
		}
	}

	return &Result{
		Success: true,
		Message: fmt.Sprintf("成功发送 %d/%d 封邮件", successCount, len(recipients)),
		Error:   nil,
	}
}

// SendWillToFirstRecipient 发送遗书到第一个收件人（用于测试）
func (es *Service) SendWillToFirstRecipient() *Result {
	// 从文件读取遗书内容
	content, err := es.state.ReadPosthumousPapers()
	if err != nil {
		es.logger.WithError(err).Error("Failed to read posthumous papers file")
		return &Result{
			Success: false,
			Message: "读取遗书文件失败",
			Error:   err,
		}
	}

	// 获取收件人邮箱列表
	recipients := es.getWillRecipients()
	if len(recipients) == 0 {
		return &Result{
			Success: false,
			Message: "未配置收件人邮箱",
			Error:   fmt.Errorf("no recipients configured"),
		}
	}

	// 获取第一个收件人
	firstRecipient := recipients[0]

	// 处理遗书内容：如果是Markdown则转换为HTML
	emailContent := es.processPosthumousContent(content)

	message := Message{
		To:      firstRecipient,
		Subject: SubjectTestWillEmail,
		Content: emailContent,
		IsHTML:  true,
	}

	result := es.sendEmail(message)
	if result.Success {
		return &Result{
			Success: true,
			Message: fmt.Sprintf("测试遗书已发送到 %s", firstRecipient),
			Error:   nil,
		}
	}
	return &Result{
		Success: false,
		Message: fmt.Sprintf("发送测试遗书到 %s 失败", firstRecipient),
		Error:   result.Error,
	}
}

// sendEmail 发送邮件
func (es *Service) sendEmail(message Message) *Result {
	// 立即尝试发送一次
	result := es.doSendEmail(message)
	if result.Success {
		return result
	}

	// 如果发送失败，启动重试机制
	es.logger.WithError(result.Error).
		WithField("to", message.To).
		WithField("subject", message.Subject).
		Warn("Email send failed, starting retry mechanism")

	// 检查重试管理器是否初始化
	if es.retryManager == nil {
		es.logger.Error(MessageRetryManagerNotInit)
		return result
	}

	// 启动重试
	es.logger.Info("Starting retry manager...")
	retrySuccess := make(chan bool, 1)
	retryError := make(chan error, 1)

	es.retryManager.StartRetry(message,
		func() {
			es.logger.Info("Retry succeeded!")
			retrySuccess <- true
		},
		func(err error) {
			es.logger.WithError(err).Error("Retry failed")
			retryError <- err
		},
	)

	es.logger.Info("Retry manager started successfully")

	// 等待重试结果（这里不阻塞，直接返回初始结果）
	// 实际的重试是异步进行的
	return result
}

// doSendEmail 实际发送邮件
func (es *Service) doSendEmail(message Message) *Result {
	// 验证邮件配置
	if err := es.validateEmailConfig(); err != nil {
		return &Result{
			Success: false,
			Message: MessageEmailConfigValidationFailed,
			Error:   err,
		}
	}

	// 构建邮件内容
	var content string
	if message.IsHTML {
		content = es.buildHTMLEmail(message.Subject, message.Content, message.To)
	} else {
		content = es.buildTextEmail(message.Subject, message.Content, message.To)
	}

	// 设置SMTP认证
	smtpHost := es.config.GetString("email.smtp_host")
	smtpPort := es.config.GetInt("email.smtp_port")
	username := es.config.GetString("email.username")
	password := es.config.GetString("email.password")
	fromEmail := es.config.GetString("email.from_email")

	auth := smtp.PlainAuth("", username, password, smtpHost)

	// 设置TLS配置
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         smtpHost,
		MinVersion:         TLSMinVersion,
	}

	// 连接SMTP服务器
	addr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)
	client, err := smtp.Dial(addr)
	if err != nil {
		return &Result{
			Success: false,
			Message: MessageConnectSMTPFailed,
			Error:   err,
		}
	}
	defer client.Close()

	// 启用TLS
	if err := client.StartTLS(tlsConfig); err != nil {
		return &Result{
			Success: false,
			Message: MessageEnableTLSFailed,
			Error:   err,
		}
	}

	// 认证
	if err := client.Auth(auth); err != nil {
		return &Result{
			Success: false,
			Message: MessageSMTPAuthFailed,
			Error:   err,
		}
	}

	// 设置发件人
	if err := client.Mail(fromEmail); err != nil {
		return &Result{
			Success: false,
			Message: MessageSetFromFailed,
			Error:   err,
		}
	}

	// 设置收件人
	if err := client.Rcpt(message.To); err != nil {
		return &Result{
			Success: false,
			Message: MessageSetToFailed,
			Error:   err,
		}
	}

	// 发送邮件内容
	wc, err := client.Data()
	if err != nil {
		return &Result{
			Success: false,
			Message: MessageGetDataFailed,
			Error:   err,
		}
	}
	defer wc.Close()

	// 写入邮件内容
	if _, err := fmt.Fprint(wc, content); err != nil {
		return &Result{
			Success: false,
			Message: MessageWriteContentFailed,
			Error:   err,
		}
	}

	es.logger.WithFields(logrus.Fields{
		"to":      message.To,
		"subject": message.Subject,
	}).Info("Email sent successfully")

	return &Result{
		Success: true,
		Message: "sent",
		Error:   nil,
	}
}

// validateEmailConfig 验证邮件配置
func (es *Service) validateEmailConfig() error {
	smtpHost := es.config.GetString("email.smtp_host")
	smtpPort := es.config.GetInt("email.smtp_port")
	username := es.config.GetString("email.username")
	password := es.config.GetString("email.password")
	fromEmail := es.config.GetString("email.from_email")

	if smtpHost == "" {
		return errors.New(ErrorSMTPHostRequired)
	}
	if smtpPort <= 0 {
		return errors.New(ErrorInvalidSMTPPort)
	}
	if username == "" {
		return errors.New(ErrorUsernameRequired)
	}
	if password == "" {
		return errors.New(ErrorPasswordRequired)
	}
	if fromEmail == "" {
		return errors.New(ErrorFromEmailRequired)
	}
	return nil
}

// buildHTMLEmail 构建HTML邮件
func (es *Service) buildHTMLEmail(subject, content, toEmail string) string {
	var builder strings.Builder

	fromEmail := es.config.GetString("email.from_email")
	builder.WriteString(fmt.Sprintf("From: %s\r\n", fromEmail))
	builder.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	builder.WriteString("MIME-Version: " + MIMEVersion + "\r\n")
	builder.WriteString("Content-Type: " + ContentTypeHTML + "\r\n")
	builder.WriteString("\r\n")
	builder.WriteString(content)

	return builder.String()
}

// buildTextEmail 构建纯文本邮件
func (es *Service) buildTextEmail(subject, content, toEmail string) string {
	var builder strings.Builder

	fromEmail := es.config.GetString("email.from_email")
	builder.WriteString(fmt.Sprintf("From: %s\r\n", fromEmail))
	builder.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	builder.WriteString("Content-Type: " + ContentTypeText + "\r\n")
	builder.WriteString("\r\n")
	builder.WriteString(content)

	return builder.String()
}

// generateTestEmailContent 生成测试邮件内容
func (es *Service) generateTestEmailContent() string {
	return fmt.Sprintf(TestEmailTemplate, time.Now().Format(TimeFormatRFC3339))
}

// generateReminderEmailContent 生成提醒邮件内容
func (es *Service) generateReminderEmailContent(reminderTime time.Duration, willSendTime time.Time) string {
	return fmt.Sprintf(ReminderEmailTemplate, reminderTime.String(), willSendTime.Format(TimeFormatRFC3339), time.Now().Format(TimeFormatRFC3339))
}

// processPosthumousContent 处理遗书内容，如果是Markdown则转换为HTML
func (es *Service) processPosthumousContent(content string) string {
	// 检查是否是Markdown文件（通过文件扩展名或内容特征）
	if es.isMarkdownContent(content) {
		return es.convertMarkdownToHTML(content)
	}

	// 如果不是Markdown，直接返回原内容（纯文本或HTML）
	return content
}

// isMarkdownContent 检查内容是否为Markdown格式
func (es *Service) isMarkdownContent(content string) bool {
	// 简单的Markdown特征检测
	markdownIndicators := MarkdownIndicators

	for _, indicator := range markdownIndicators {
		if strings.Contains(content, indicator) {
			return true
		}
	}

	return false
}

// convertMarkdownToHTML 将Markdown转换为安全的HTML
func (es *Service) convertMarkdownToHTML(markdown string) string {
	// 将Markdown转换为HTML
	unsafeHTML := blackfriday.Run([]byte(markdown), blackfriday.WithNoExtensions(), blackfriday.WithRenderer(blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags | blackfriday.HrefTargetBlank,
	})))

	// 使用bluemonday清理HTML，确保安全性
	policy := bluemonday.UGCPolicy()
	policy.AllowStandardURLs()
	policy.AllowStandardAttributes()
	policy.AllowElements(AllowedHTMLTags...)

	safeHTML := policy.SanitizeBytes(unsafeHTML)

	// 添加基本的HTML文档结构
	return fmt.Sprintf(WillEmailTemplateWrapper, WillEmailStyles, string(safeHTML))
}

// getWillRecipients 获取遗书收件人列表
func (es *Service) getWillRecipients() []string {
	recipients := make([]string, 0)

	// 从配置中获取收件人列表
	emailConfig := es.config.GetEmailConfig()
	for _, recipient := range emailConfig.Recipients {
		if recipient.Email != "" {
			recipients = append(recipients, recipient.Email)
		}
	}

	// 如果没有配置收件人，使用测试邮箱作为后备
	if len(recipients) == 0 {
		testEmail := es.config.GetString("email.test_email")
		if testEmail != "" {
			recipients = append(recipients, testEmail)
		}
	}

	return recipients
}

// RetryFailedEmails 重试发送失败的邮件
func (es *Service) RetryFailedEmails() error {
	// 由于不再存储邮件记录，此方法不再需要
	return nil
}

// GetEmailStats 获取邮件统计信息
func (es *Service) GetEmailStats() (map[string]any, error) {
	// 由于不再存储邮件记录，返回空统计
	stats := map[string]any{
		"total_emails":   int64(0),
		"sent_emails":    int64(0),
		"failed_emails":  int64(0),
		"pending_emails": int64(0),
		"success_rate":   float64(0),
	}

	return stats, nil
}

// GetRetryStatus 获取重试状态
func (es *Service) GetRetryStatus() map[string]any {
	if es.retryManager != nil {
		return es.retryManager.GetRetryStatus()
	}
	return nil
}

// StopAllRetries 停止所有重试
func (es *Service) StopAllRetries() {
	if es.retryManager != nil {
		es.retryManager.StopAllRetries()
		es.logger.Info("All email retries stopped")
	}
}

// UpdateEmailConfig 更新邮件配置
func (es *Service) UpdateEmailConfig(smtpHost string, smtpPort int, username, password, fromEmail, testEmail string) error {
	// 更新配置
	es.config.Viper.Set("email.smtp_host", smtpHost)
	es.config.Viper.Set("email.smtp_port", smtpPort)
	es.config.Viper.Set("email.username", username)
	es.config.Viper.Set("email.password", password)
	es.config.Viper.Set("email.from_email", fromEmail)
	es.config.Viper.Set("email.test_email", testEmail)

	// 保存配置
	configPath := es.config.Viper.ConfigFileUsed()
	if err := es.config.Viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to save email config: %w", err)
	}

	es.logger.Info("Email configuration updated successfully")
	return nil
}

// TestEmailConfig 测试邮件配置
func (es *Service) TestEmailConfig(smtpHost string, smtpPort int, username, password, fromEmail, testEmail string) *Result {
	// 临时使用新的配置发送测试邮件
	oldSMTPHost := es.config.GetString("email.smtp_host")
	oldSMTPPort := es.config.GetInt("email.smtp_port")
	oldUsername := es.config.GetString("email.username")
	oldPassword := es.config.GetString("email.password")
	oldFromEmail := es.config.GetString("email.from_email")
	oldTestEmail := es.config.GetString("email.test_email")

	// 临时设置新配置
	es.config.Viper.Set("email.smtp_host", smtpHost)
	es.config.Viper.Set("email.smtp_port", smtpPort)
	es.config.Viper.Set("email.username", username)
	es.config.Viper.Set("email.password", password)
	es.config.Viper.Set("email.from_email", fromEmail)
	es.config.Viper.Set("email.test_email", testEmail)

	// 恢复旧配置
	defer func() {
		es.config.Viper.Set("email.smtp_host", oldSMTPHost)
		es.config.Viper.Set("email.smtp_port", oldSMTPPort)
		es.config.Viper.Set("email.username", oldUsername)
		es.config.Viper.Set("email.password", oldPassword)
		es.config.Viper.Set("email.from_email", oldFromEmail)
		es.config.Viper.Set("email.test_email", oldTestEmail)
	}()

	message := Message{
		To:      testEmail,
		Subject: SubjectConfigTest,
		Content: es.generateTestEmailContent(),
		IsHTML:  true,
	}

	return es.sendEmail(message)
}
