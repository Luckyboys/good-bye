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

// Handler API处理器
type Handler struct {
	stateMgr  *state.Manager
	emailSvc  *email.Service
	configMgr *config.Manager
	logger    *logrus.Logger
}

// NewAPIHandler 创建新的API处理器
func NewAPIHandler(stateMgr *state.Manager, emailSvc *email.Service, configMgr *config.Manager, logger *logrus.Logger) *Handler {
	return &Handler{
		stateMgr:  stateMgr,
		emailSvc:  emailSvc,
		configMgr: configMgr,
		logger:    logger,
	}
}

// Response 标准API响应结构
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// HealthCheck 健康检查
func (h *Handler) HealthCheck(c *gin.Context) {
	health := h.stateMgr.GetHealthStatus()

	response := Response{
		Success: true,
		Message: "Service is healthy",
		Data:    health,
	}

	c.JSON(http.StatusOK, response)
}

// CheckIn 签到接口
func (h *Handler) CheckIn(c *gin.Context) {
	if err := h.stateMgr.UpdateStatus(); err != nil {
		h.logger.WithError(err).Error("Failed to update status")
		response := Response{
			Success: false,
			Message: "签到失败",
			Error:   err.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	h.stateMgr.LogStateChange("check_in", "User checked in successfully")

	response := Response{
		Success: true,
		Message: "签到成功",
		Data: map[string]any{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetStatus 获取状态信息
func (h *Handler) GetStatus(c *gin.Context) {
	status, err := h.stateMgr.GetStatus()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get status")
		response := Response{
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
		response := Response{
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
		response := Response{
			Success: false,
			Message: "获取系统设置失败",
			Error:   err.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	data := map[string]any{
		"last_seen":         status.LastSeen.Format(time.RFC3339),
		"inactive_duration": inactiveDuration.String(),
		"is_inactive":       h.stateMgr.IsInactiveWithStatus(status, settings.MaxInactiveDays),
		"max_inactive_days": settings.MaxInactiveDays,
		"check_interval":    settings.CheckInterval,
	}

	response := Response{
		Success: true,
		Message: "获取状态成功",
		Data:    data,
	}

	c.JSON(http.StatusOK, response)
}

// GetSystemSettings 获取系统设置
func (h *Handler) GetSystemSettings(c *gin.Context) {
	settings, err := h.stateMgr.GetSystemSettings()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get system settings")
		response := Response{
			Success: false,
			Message: "获取系统设置失败",
			Error:   err.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := Response{
		Success: true,
		Message: "获取系统设置成功",
		Data:    settings,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateSystemSettings 更新系统设置
func (h *Handler) UpdateSystemSettings(c *gin.Context) {
	var settings config.SystemConfig
	if err := c.ShouldBindJSON(&settings); err != nil {
		response := Response{
			Success: false,
			Message: "无效的请求数据",
			Error:   err.Error(),
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := h.stateMgr.UpdateSystemSettings(&settings); err != nil {
		h.logger.WithError(err).Error("Failed to update system settings")
		response := Response{
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
	h.configMgr.Viper.Set("system.timezone", settings.Timezone)

	configPath := h.configMgr.Viper.ConfigFileUsed()
	if err := h.configMgr.Viper.WriteConfigAs(configPath); err != nil {
		h.logger.WithError(err).Error("Failed to save config file")
	}

	h.stateMgr.LogStateChange("update_settings", "System settings updated")

	response := Response{
		Success: true,
		Message: "系统设置更新成功",
		Data:    settings,
	}

	c.JSON(http.StatusOK, response)
}

// GetWillMessages 获取遗书状态（从文件读取）
func (h *Handler) GetWillMessages(c *gin.Context) {
	// 检查遗书文件是否存在
	content, err := h.stateMgr.ReadPosthumousPapers()
	if err != nil {
		response := Response{
			Success: false,
			Message: "遗书文件不存在或无法读取",
			Error:   err.Error(),
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	data := map[string]any{
		"file_exists": true,
		"content":     content,
		"message":     "遗书内容已从文件读取",
	}

	response := Response{
		Success: true,
		Message: "获取遗书状态成功",
		Data:    data,
	}

	c.JSON(http.StatusOK, response)
}

// SendTestEmail 发送测试邮件
func (h *Handler) SendTestEmail(c *gin.Context) {
	result := h.emailSvc.SendTestEmail()

	if result.Success {
		h.stateMgr.LogStateChange("send_test_email", "Test email sent successfully")
		response := Response{
			Success: true,
			Message: result.Message,
		}
		c.JSON(http.StatusOK, response)
	} else {
		h.logger.WithError(result.Error).Error("Failed to send test email")
		response := Response{
			Success: false,
			Message: result.Message,
			Error:   result.Error.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
	}
}

// SendTestWill 发送测试遗书到第一个收件人
func (h *Handler) SendTestWill(c *gin.Context) {
	result := h.emailSvc.SendWillToFirstRecipient()

	if result.Success {
		h.stateMgr.LogStateChange("send_test_will", "Test will sent successfully")
		response := Response{
			Success: true,
			Message: result.Message,
		}
		c.JSON(http.StatusOK, response)
	} else {
		h.logger.WithError(result.Error).Error("Failed to send test will")
		response := Response{
			Success: false,
			Message: result.Message,
			Error:   result.Error.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
	}
}

// GetEmailConfig 获取邮件配置
func (h *Handler) GetEmailConfig(c *gin.Context) {
	emailConfig := h.configMgr.GetEmailConfig()

	response := Response{
		Success: true,
		Message: "获取邮件配置成功",
		Data:    emailConfig,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateEmailConfig 更新邮件配置
func (h *Handler) UpdateEmailConfig(c *gin.Context) {
	var emailConfig map[string]any
	if err := c.ShouldBindJSON(&emailConfig); err != nil {
		response := Response{
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
		response := Response{
			Success: false,
			Message: "更新邮件配置失败",
			Error:   err.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	h.stateMgr.LogStateChange("update_email_config", "Email configuration updated")

	response := Response{
		Success: true,
		Message: "邮件配置更新成功",
		Data:    emailConfig,
	}

	c.JSON(http.StatusOK, response)
}

// TestEmailConfig 测试邮件配置
func (h *Handler) TestEmailConfig(c *gin.Context) {
	var emailConfig map[string]any
	if err := c.ShouldBindJSON(&emailConfig); err != nil {
		response := Response{
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
		response := Response{
			Success: true,
			Message: result.Message,
		}
		c.JSON(http.StatusOK, response)
	} else {
		response := Response{
			Success: false,
			Message: result.Message,
			Error:   result.Error.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
	}
}

// GetStats 获取统计信息
func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.stateMgr.GetStats()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get stats")
		response := Response{
			Success: false,
			Message: "获取统计信息失败",
			Error:   err.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := Response{
		Success: true,
		Message: "获取统计信息成功",
		Data:    stats,
	}

	c.JSON(http.StatusOK, response)
}
