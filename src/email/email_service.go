package email

import (
	"crypto/tls"
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
			Message: "测试邮箱地址未配置",
			Error:   fmt.Errorf("test email not configured"),
		}
	}

	message := Message{
		To:      testEmail,
		Subject: "生存确认服务测试邮件",
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
			Message: "测试邮箱地址未配置",
			Error:   fmt.Errorf("test email not configured"),
		}
	}

	message := Message{
		To:      testEmail,
		Subject: "生存确认服务 - 签到提醒",
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

	// 处理遗书内容：如果是Markdown则转换为HTML
	emailContent := es.processPosthumousContent(content)

	// 发送给所有收件人
	var lastError error
	successCount := 0

	for _, recipient := range recipients {
		message := Message{
			To:      recipient,
			Subject: "遗书通知 - 重要信息",
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
			Message: "所有邮件发送失败",
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
		Subject: "【测试】遗书通知 - 重要信息",
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
		es.logger.Error("Retry manager is not initialized")
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
			Message: "邮件配置验证失败",
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
		MinVersion:         tls.VersionTLS12,
	}

	// 连接SMTP服务器
	addr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)
	client, err := smtp.Dial(addr)
	if err != nil {
		return &Result{
			Success: false,
			Message: "连接SMTP服务器失败",
			Error:   err,
		}
	}
	defer client.Close()

	// 启用TLS
	if err := client.StartTLS(tlsConfig); err != nil {
		return &Result{
			Success: false,
			Message: "启用TLS失败",
			Error:   err,
		}
	}

	// 认证
	if err := client.Auth(auth); err != nil {
		return &Result{
			Success: false,
			Message: "SMTP认证失败",
			Error:   err,
		}
	}

	// 设置发件人
	if err := client.Mail(fromEmail); err != nil {
		return &Result{
			Success: false,
			Message: "设置发件人失败",
			Error:   err,
		}
	}

	// 设置收件人
	if err := client.Rcpt(message.To); err != nil {
		return &Result{
			Success: false,
			Message: "设置收件人失败",
			Error:   err,
		}
	}

	// 发送邮件内容
	wc, err := client.Data()
	if err != nil {
		return &Result{
			Success: false,
			Message: "获取邮件写入器失败",
			Error:   err,
		}
	}
	defer wc.Close()

	// 写入邮件内容
	if _, err := fmt.Fprint(wc, content); err != nil {
		return &Result{
			Success: false,
			Message: "写入邮件内容失败",
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
		return fmt.Errorf("SMTP host is required")
	}
	if smtpPort <= 0 {
		return fmt.Errorf("invalid SMTP port")
	}
	if username == "" {
		return fmt.Errorf("username is required")
	}
	if password == "" {
		return fmt.Errorf("password is required")
	}
	if fromEmail == "" {
		return fmt.Errorf("from email is required")
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
	builder.WriteString("MIME-Version: 1.0\r\n")
	builder.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
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
	builder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	builder.WriteString("\r\n")
	builder.WriteString(content)

	return builder.String()
}

// generateTestEmailContent 生成测试邮件内容
func (es *Service) generateTestEmailContent() string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>生存确认服务测试邮件</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #f4f4f4; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #fff; }
        .footer { background-color: #f4f4f4; padding: 10px; text-align: center; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>生存确认服务测试邮件</h1>
        </div>
        <div class="content">
            <p>您好！</p>
            <p>这是一封来自生存确认服务的测试邮件。</p>
            <p>如果您收到这封邮件，说明邮件服务配置正常。</p>
            <p>发送时间：%s</p>
        </div>
        <div class="footer">
            <p>生存确认服务 - 自动化邮件通知系统</p>
        </div>
    </div>
</body>
</html>
`, time.Now().Format("2006-01-02 15:04:05"))
}

// generateReminderEmailContent 生成提醒邮件内容
func (es *Service) generateReminderEmailContent(reminderTime time.Duration, willSendTime time.Time) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>生存确认服务 - 签到提醒</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #fff3cd; padding: 20px; text-align: center; border: 1px solid #ffeaa7; }
        .content { padding: 20px; background-color: #fff; border: 1px solid #ddd; border-top: none; }
        .warning { background-color: #f8d7da; padding: 15px; border: 1px solid #f5c6cb; border-radius: 4px; margin: 20px 0; }
        .footer { background-color: #f8f9fa; padding: 15px; text-align: center; font-size: 12px; border: 1px solid #ddd; border-top: none; }
        .btn { display: inline-block; padding: 10px 20px; background-color: #007bff; color: white; text-decoration: none; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="color: #856404;">🔔 生存确认服务 - 签到提醒</h1>
        </div>
        <div class="content">
            <p>您好！</p>
            <p>这是一封来自生存确认服务的<strong>签到提醒邮件</strong>。</p>
            
            <div class="warning">
                <h3>⚠️ 重要提醒</h3>
                <p>您已经 <strong>%s</strong> 没有进行签到操作。</p>
                <p>如果您在 <strong>%s</strong> 之前仍然没有签到，系统将自动发送您的遗书邮件。</p>
            </div>
            
            <h4>请立即采取以下操作：</h4>
            <ol>
                <li>访问生存确认服务</li>
                <li>进行签到操作以确认您的安全状态</li>
                <li>确保定期签到以避免误触发遗书发送</li>
            </ol>
            
            <p><strong>提醒：</strong>请勿忽视此提醒，一旦遗书邮件发送，将无法撤销。</p>
        </div>
        <div class="footer">
            <p>生存确认服务 - 自动化邮件通知系统</p>
            <p>发送时间：%s</p>
        </div>
    </div>
</body>
</html>
`, reminderTime.String(), willSendTime.Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"))
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
	markdownIndicators := []string{
		"# ",   // 标题
		"## ",  // 二级标题
		"### ", // 三级标题
		"- ",   // 无序列表
		"* ",   // 无序列表
		"1. ",  // 有序列表
		"[",    // 链接
		"![",   // 图片
		"```",  // 代码块
		"`",    // 行内代码
		"> ",   // 引用
		"**",   // 粗体
		"*",    // 斜体
		"---",  // 分割线
	}

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
	policy.AllowElements("h1", "h2", "h3", "h4", "h5", "h6", "p", "br", "hr",
		"ul", "ol", "li", "blockquote", "pre", "code",
		"strong", "em", "i", "b", "a", "img",
		"table", "thead", "tbody", "tr", "th", "td")

	safeHTML := policy.SanitizeBytes(unsafeHTML)

	// 添加基本的HTML文档结构
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>遗书</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            line-height: 1.6; 
            color: #333; 
            max-width: 800px; 
            margin: 0 auto; 
            padding: 20px;
            background-color: #fff;
        }
        h1, h2, h3, h4, h5, h6 { color: #2c3e50; margin-top: 2em; margin-bottom: 1em; }
        h1 { border-bottom: 2px solid #3498db; padding-bottom: 0.3em; }
        h2 { border-bottom: 1px solid #bdc3c7; padding-bottom: 0.2em; }
        p { margin-bottom: 1em; }
        pre { 
            background-color: #f8f9fa; 
            border: 1px solid #e9ecef; 
            border-radius: 4px; 
            padding: 16px; 
            overflow-x: auto;
            font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
        }
        code { 
            background-color: #f8f9fa; 
            border: 1px solid #e9ecef; 
            border-radius: 3px; 
            padding: 2px 4px; 
            font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
        }
        blockquote { 
            border-left: 4px solid #3498db; 
            margin: 1em 0; 
            padding-left: 1em; 
            color: #7f8c8d; 
        }
        table { 
            border-collapse: collapse; 
            width: 100%%; 
            margin: 1em 0; 
        }
        th, td { 
            border: 1px solid #ddd; 
            padding: 8px; 
            text-align: left; 
        }
        th { 
            background-color: #f8f9fa; 
            font-weight: bold; 
        }
        img { max-width: 100%%; height: auto; }
        a { color: #3498db; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    %s
</body>
</html>`, string(safeHTML))
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
		Subject: "邮件配置测试",
		Content: es.generateTestEmailContent(),
		IsHTML:  true,
	}

	return es.sendEmail(message)
}
