package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/Luckyboys/good-bye/src/config"
	"github.com/Luckyboys/good-bye/src/state"
	"github.com/sirupsen/logrus"
)

// Service 邮件服务
type Service struct {
	config *config.Manager
	state  *state.Manager
	logger *logrus.Logger
}

// NewEmailService 创建新的邮件服务
func NewEmailService(cfg *config.Manager, stateMgr *state.Manager, logger *logrus.Logger) *Service {
	return &Service{
		config: cfg,
		state:  stateMgr,
		logger: logger,
	}
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
		Subject: "遗书服务测试邮件",
		Content: es.generateTestEmailContent(),
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

	// 生成邮件内容
	emailContent := es.generateWillEmailContentFromFile(content)

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

// sendEmail 发送邮件
func (es *Service) sendEmail(message Message) *Result {
	return es.doSendEmail(message)
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
    <title>遗书服务测试邮件</title>
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
            <h1>遗书服务测试邮件</h1>
        </div>
        <div class="content">
            <p>您好！</p>
            <p>这是一封来自遗书服务的测试邮件。</p>
            <p>如果您收到这封邮件，说明邮件服务配置正常。</p>
            <p>发送时间：%s</p>
        </div>
        <div class="footer">
            <p>遗书服务 - 自动化邮件通知系统</p>
        </div>
    </div>
</body>
</html>
`, time.Now().Format("2006-01-02 15:04:05"))
}

// generateWillEmailContentFromFile 从文件内容生成遗书邮件内容
func (es *Service) generateWillEmailContentFromFile(content string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>遗书通知 - 重要信息</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #ff6b6b; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #fff; border: 1px solid #ddd; }
        .footer { background-color: #f4f4f4; padding: 10px; text-align: center; font-size: 12px; }
        .will-content { background-color: #f9f9f9; padding: 15px; border-left: 4px solid #ff6b6b; margin: 20px 0; }
        .will-content pre { white-space: pre-wrap; word-wrap: break-word; margin: 0; font-family: inherit; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>遗书通知</h1>
        </div>
        <div class="content">
            <p>尊敬的收件人：</p>
            <p>此邮件是由遗书服务自动发送的重要通知。</p>
            <p>由于长时间未检测到用户活动，系统认为需要发送以下重要信息。</p>
            
            <div class="will-content">
                <h2>遗书内容</h2>
                <pre>%s</pre>
            </div>
            
            <p><strong>发送时间：</strong>%s</p>
            
            <p style="color: #666; font-size: 14px;">
                请注意：这是一封非常重要的通知邮件。如果您对此有任何疑问，请及时联系相关人员。
            </p>
        </div>
        <div class="footer">
            <p>遗书服务 - 自动化邮件通知系统</p>
        </div>
    </div>
</body>
</html>
`, content, time.Now().Format("2006-01-02 15:04:05"))
}

// getWillRecipients 获取遗书收件人列表
func (es *Service) getWillRecipients() []string {
	recipients := make([]string, 0)

	// 从配置中获取收件人
	testEmail := es.config.GetString("email.test_email")
	if testEmail != "" {
		recipients = append(recipients, testEmail)
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
