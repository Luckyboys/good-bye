package handlers

import (
	"fmt"

	"github.com/Luckyboys/good-bye/src/api/middleware"
	"github.com/Luckyboys/good-bye/src/api/response"
	"github.com/Luckyboys/good-bye/src/config"
	"github.com/Luckyboys/good-bye/src/email"
	"github.com/Luckyboys/good-bye/src/state"
	"github.com/gin-gonic/gin"
)

// EmailHandler 邮件服务处理器
type EmailHandler struct {
	stateMgr  *state.Manager
	emailSvc  *email.Service
	configMgr *config.Manager
}

// NewEmailHandler 创建新的邮件处理器
func NewEmailHandler(stateMgr *state.Manager, emailSvc *email.Service, configMgr *config.Manager) *EmailHandler {
	return &EmailHandler{
		stateMgr:  stateMgr,
		emailSvc:  emailSvc,
		configMgr: configMgr,
	}
}

// SendTestEmail 发送测试邮件
func (h *EmailHandler) SendTestEmail(c *gin.Context) {
	ctx, exists := middleware.GetAPIContext(c)
	if !exists {
		response.InternalServerError(c, "API context not found", nil)
		return
	}

	result := h.emailSvc.SendTestEmail()

	if result.Success {
		h.stateMgr.LogStateChange("send_test_email", "Test email sent successfully")
		response.Success(c, result.Message, nil)
	} else {
		ctx.Logger.WithError(result.Error).Error("Failed to send test email")
		response.InternalServerError(c, result.Message, result.Error)
	}
}

// GetEmailConfig 获取邮件配置
func (h *EmailHandler) GetEmailConfig(c *gin.Context) {
	_, exists := middleware.GetAPIContext(c)
	if !exists {
		response.InternalServerError(c, "API context not found", nil)
		return
	}

	emailConfig := h.configMgr.GetEmailConfig()

	response.Success(c, "获取邮件配置成功", emailConfig)
}

// UpdateEmailConfig 更新邮件配置
func (h *EmailHandler) UpdateEmailConfig(c *gin.Context) {
	ctx, exists := middleware.GetAPIContext(c)
	if !exists {
		response.InternalServerError(c, "API context not found", nil)
		return
	}

	var emailConfig config.EmailConfig
	if err := c.ShouldBindJSON(&emailConfig); err != nil {
		response.BadRequest(c, "无效的请求数据", err)
		return
	}

	// 验证邮件配置
	if err := h.validateEmailConfig(&emailConfig); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	if err := h.emailSvc.UpdateEmailConfig(
		emailConfig.SMTPHost,
		emailConfig.SMTPPort,
		emailConfig.Username,
		emailConfig.Password,
		emailConfig.FromEmail,
		emailConfig.TestEmail,
	); err != nil {
		ctx.Logger.WithError(err).Error("Failed to update email config")
		response.InternalServerError(c, "更新邮件配置失败", err)
		return
	}

	// 同步到配置文件
	h.configMgr.Viper.Set("email.smtp_host", emailConfig.SMTPHost)
	h.configMgr.Viper.Set("email.smtp_port", emailConfig.SMTPPort)
	h.configMgr.Viper.Set("email.username", emailConfig.Username)
	h.configMgr.Viper.Set("email.password", emailConfig.Password)
	h.configMgr.Viper.Set("email.from_email", emailConfig.FromEmail)
	h.configMgr.Viper.Set("email.test_email", emailConfig.TestEmail)

	configPath := h.configMgr.Viper.ConfigFileUsed()
	if err := h.configMgr.Viper.WriteConfigAs(configPath); err != nil {
		ctx.Logger.WithError(err).Error("Failed to save config file")
		// 不影响主要功能，只记录错误
	}

	h.stateMgr.LogStateChange("update_email_config", "Email configuration updated")

	response.Success(c, "邮件配置更新成功", emailConfig)
}

// TestEmailConfig 测试邮件配置
func (h *EmailHandler) TestEmailConfig(c *gin.Context) {
	_, exists := middleware.GetAPIContext(c)
	if !exists {
		response.InternalServerError(c, "API context not found", nil)
		return
	}

	var emailConfig config.EmailConfig
	if err := c.ShouldBindJSON(&emailConfig); err != nil {
		response.BadRequest(c, "无效的请求数据", err)
		return
	}

	// 验证邮件配置
	if err := h.validateEmailConfig(&emailConfig); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	result := h.emailSvc.TestEmailConfig(
		emailConfig.SMTPHost,
		emailConfig.SMTPPort,
		emailConfig.Username,
		emailConfig.Password,
		emailConfig.FromEmail,
		emailConfig.TestEmail,
	)

	if result.Success {
		response.Success(c, result.Message, nil)
	} else {
		response.InternalServerError(c, result.Message, result.Error)
	}
}

// validateEmailConfig 验证邮件配置
func (h *EmailHandler) validateEmailConfig(config *config.EmailConfig) error {
	if config.SMTPHost == "" {
		return fmt.Errorf("SMTP主机不能为空")
	}

	if config.SMTPPort <= 0 || config.SMTPPort > 65535 {
		return fmt.Errorf("SMTP端口必须在1-65535之间")
	}

	if config.Username == "" {
		return fmt.Errorf("用户名不能为空")
	}

	if config.Password == "" {
		return fmt.Errorf("密码不能为空")
	}

	if config.FromEmail == "" {
		return fmt.Errorf("发件人邮箱不能为空")
	}

	if config.TestEmail == "" {
		return fmt.Errorf("测试邮箱不能为空")
	}

	return nil
}
