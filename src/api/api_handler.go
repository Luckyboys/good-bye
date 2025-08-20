package api

import (
	"net/http"
	"time"

	"github.com/Luckyboys/good-bye/src/config"
	"github.com/Luckyboys/good-bye/src/email"
	"github.com/Luckyboys/good-bye/src/state"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type APIHandler struct {
	stateMgr  *state.StateManager
	emailSvc  *email.EmailService
	configMgr *config.ConfigManager
	logger    *logrus.Logger
}

func NewAPIHandler(stateMgr *state.StateManager, emailSvc *email.EmailService, configMgr *config.ConfigManager, logger *logrus.Logger) *APIHandler {
	return &APIHandler{
		stateMgr:  stateMgr,
		emailSvc:  emailSvc,
		configMgr: configMgr,
		logger:    logger,
	}
}

// APIResponse 标准API响应结构
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// 健康检查
func (h *APIHandler) HealthCheck(c *gin.Context) {
	health := h.stateMgr.GetHealthStatus()

	response := APIResponse{
		Success: true,
		Message: "Service is healthy",
		Data:    health,
	}

	c.JSON(http.StatusOK, response)
}

// 签到接口
func (h *APIHandler) CheckIn(c *gin.Context) {
	if err := h.stateMgr.UpdateStatus(); err != nil {
		h.logger.WithError(err).Error("Failed to update status")
		response := APIResponse{
			Success: false,
			Message: "签到失败",
			Error:   err.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	h.stateMgr.LogStateChange("check_in", "User checked in successfully")

	response := APIResponse{
		Success: true,
		Message: "签到成功",
		Data: map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusOK, response)
}

// 获取状态信息
func (h *APIHandler) GetStatus(c *gin.Context) {
	status, err := h.stateMgr.GetStatus()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get status")
		response := APIResponse{
			Success: false,
			Message: "获取状态失败",
			Error:   err.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	inactiveDuration, err := h.stateMgr.GetInactiveDuration()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get inactive duration")
		response := APIResponse{
			Success: false,
			Message: "获取不活跃时长失败",
			Error:   err.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	settings, err := h.stateMgr.GetSystemSettings()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get system settings")
		response := APIResponse{
			Success: false,
			Message: "获取系统设置失败",
			Error:   err.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	data := map[string]interface{}{
		"last_seen":         status.LastSeen.Format(time.RFC3339),
		"inactive_duration": inactiveDuration.String(),
		"is_inactive":       h.stateMgr.IsInactiveWithStatus(status, settings.MaxInactiveDays),
		"max_inactive_days": settings.MaxInactiveDays,
		"check_interval":    settings.CheckInterval,
	}

	response := APIResponse{
		Success: true,
		Message: "获取状态成功",
		Data:    data,
	}

	c.JSON(http.StatusOK, response)
}

// 获取系统设置
func (h *APIHandler) GetSystemSettings(c *gin.Context) {
	settings, err := h.stateMgr.GetSystemSettings()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get system settings")
		response := APIResponse{
			Success: false,
			Message: "获取系统设置失败",
			Error:   err.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := APIResponse{
		Success: true,
		Message: "获取系统设置成功",
		Data:    settings,
	}

	c.JSON(http.StatusOK, response)
}

// 更新系统设置
func (h *APIHandler) UpdateSystemSettings(c *gin.Context) {
	var settings config.SystemConfig
	if err := c.ShouldBindJSON(&settings); err != nil {
		response := APIResponse{
			Success: false,
			Message: "无效的请求数据",
			Error:   err.Error(),
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := h.stateMgr.UpdateSystemSettings(&settings); err != nil {
		h.logger.WithError(err).Error("Failed to update system settings")
		response := APIResponse{
			Success: false,
			Message: "更新系统设置失败",
			Error:   err.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// 同步到配置文件
	h.configMgr.Viper.Set("system.check_interval", settings.CheckInterval)
	h.configMgr.Viper.Set("system.max_inactive_days", settings.MaxInactiveDays)
	h.configMgr.Viper.Set("system.enable_notification", settings.EnableNotification)
	h.configMgr.Viper.Set("system.timezone", settings.Timezone)

	configPath := h.configMgr.Viper.ConfigFileUsed()
	if err := h.configMgr.Viper.WriteConfigAs(configPath); err != nil {
		h.logger.WithError(err).Error("Failed to save config file")
	}

	h.stateMgr.LogStateChange("update_settings", "System settings updated")

	response := APIResponse{
		Success: true,
		Message: "系统设置更新成功",
		Data:    settings,
	}

	c.JSON(http.StatusOK, response)
}

// 获取遗书状态（从文件读取）
func (h *APIHandler) GetWillMessages(c *gin.Context) {
	// 检查遗书文件是否存在
	content, err := h.stateMgr.ReadPosthumousPapers()
	if err != nil {
		response := APIResponse{
			Success: false,
			Message: "遗书文件不存在或无法读取",
			Error:   err.Error(),
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	data := map[string]interface{}{
		"file_exists": true,
		"content":     content,
		"message":     "遗书内容已从文件读取",
	}

	response := APIResponse{
		Success: true,
		Message: "获取遗书状态成功",
		Data:    data,
	}

	c.JSON(http.StatusOK, response)
}

// 发送测试邮件
func (h *APIHandler) SendTestEmail(c *gin.Context) {
	result := h.emailSvc.SendTestEmail()

	if result.Success {
		h.stateMgr.LogStateChange("send_test_email", "Test email sent successfully")
		response := APIResponse{
			Success: true,
			Message: result.Message,
		}
		c.JSON(http.StatusOK, response)
	} else {
		h.logger.WithError(result.Error).Error("Failed to send test email")
		response := APIResponse{
			Success: false,
			Message: result.Message,
			Error:   result.Error.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
	}
}

// 获取邮件配置
func (h *APIHandler) GetEmailConfig(c *gin.Context) {
	emailConfig := map[string]interface{}{
		"smtp_host":  h.configMgr.GetString("email.smtp_host"),
		"smtp_port":  h.configMgr.GetInt("email.smtp_port"),
		"username":   h.configMgr.GetString("email.username"),
		"password":   h.configMgr.GetString("email.password"),
		"from_email": h.configMgr.GetString("email.from_email"),
		"test_email": h.configMgr.GetString("email.test_email"),
	}

	response := APIResponse{
		Success: true,
		Message: "获取邮件配置成功",
		Data:    emailConfig,
	}

	c.JSON(http.StatusOK, response)
}

// 更新邮件配置
func (h *APIHandler) UpdateEmailConfig(c *gin.Context) {
	var emailConfig map[string]interface{}
	if err := c.ShouldBindJSON(&emailConfig); err != nil {
		response := APIResponse{
			Success: false,
			Message: "无效的请求数据",
			Error:   err.Error(),
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	smtpHost, _ := emailConfig["smtp_host"].(string)
	smtpPort, _ := emailConfig["smtp_port"].(int)
	username, _ := emailConfig["username"].(string)
	password, _ := emailConfig["password"].(string)
	fromEmail, _ := emailConfig["from_email"].(string)
	testEmail, _ := emailConfig["test_email"].(string)

	if err := h.emailSvc.UpdateEmailConfig(smtpHost, smtpPort, username, password, fromEmail, testEmail); err != nil {
		h.logger.WithError(err).Error("Failed to update email config")
		response := APIResponse{
			Success: false,
			Message: "更新邮件配置失败",
			Error:   err.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	h.stateMgr.LogStateChange("update_email_config", "Email configuration updated")

	response := APIResponse{
		Success: true,
		Message: "邮件配置更新成功",
		Data:    emailConfig,
	}

	c.JSON(http.StatusOK, response)
}

// 测试邮件配置
func (h *APIHandler) TestEmailConfig(c *gin.Context) {
	var emailConfig map[string]interface{}
	if err := c.ShouldBindJSON(&emailConfig); err != nil {
		response := APIResponse{
			Success: false,
			Message: "无效的请求数据",
			Error:   err.Error(),
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	smtpHost, _ := emailConfig["smtp_host"].(string)
	smtpPort, _ := emailConfig["smtp_port"].(int)
	username, _ := emailConfig["username"].(string)
	password, _ := emailConfig["password"].(string)
	fromEmail, _ := emailConfig["from_email"].(string)
	testEmail, _ := emailConfig["test_email"].(string)

	result := h.emailSvc.TestEmailConfig(smtpHost, smtpPort, username, password, fromEmail, testEmail)

	if result.Success {
		response := APIResponse{
			Success: true,
			Message: result.Message,
		}
		c.JSON(http.StatusOK, response)
	} else {
		response := APIResponse{
			Success: false,
			Message: result.Message,
			Error:   result.Error.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
	}
}

// 获取统计信息
func (h *APIHandler) GetStats(c *gin.Context) {
	stats, err := h.stateMgr.GetStats()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get stats")
		response := APIResponse{
			Success: false,
			Message: "获取统计信息失败",
			Error:   err.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := APIResponse{
		Success: true,
		Message: "获取统计信息成功",
		Data:    stats,
	}

	c.JSON(http.StatusOK, response)
}
