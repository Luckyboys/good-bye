package handlers

import (
	"github.com/Luckyboys/good-bye/src/api/middleware"
	"github.com/Luckyboys/good-bye/src/api/response"
	"github.com/Luckyboys/good-bye/src/config"
	"github.com/Luckyboys/good-bye/src/state"
	"github.com/gin-gonic/gin"
)

// SettingsHandler 系统设置处理器
type SettingsHandler struct {
	stateMgr  *state.Manager
	configMgr *config.Manager
}

// NewSettingsHandler 创建新的设置处理器
func NewSettingsHandler(stateMgr *state.Manager, configMgr *config.Manager) *SettingsHandler {
	return &SettingsHandler{
		stateMgr:  stateMgr,
		configMgr: configMgr,
	}
}

// GetSystemSettings 获取系统设置
func (h *SettingsHandler) GetSystemSettings(c *gin.Context) {
	ctx, exists := middleware.GetAPIContext(c)
	if !exists {
		response.InternalServerError(c, "API context not found", nil)
		return
	}

	settings, err := h.stateMgr.GetSystemSettings()
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to get system settings")
		response.InternalServerError(c, "获取系统设置失败", err)
		return
	}

	response.Success(c, "获取系统设置成功", settings)
}

// UpdateSystemSettings 更新系统设置
func (h *SettingsHandler) UpdateSystemSettings(c *gin.Context) {
	ctx, exists := middleware.GetAPIContext(c)
	if !exists {
		response.InternalServerError(c, "API context not found", nil)
		return
	}

	var settings config.SystemConfig
	if err := c.ShouldBindJSON(&settings); err != nil {
		response.BadRequest(c, "无效的请求数据", err)
		return
	}

	// 验证设置
	if settings.CheckInterval <= 0 {
		response.BadRequest(c, "检查间隔必须大于0", nil)
		return
	}

	if settings.MaxInactiveTime <= 0 {
		response.BadRequest(c, "最大不活跃天数必须大于0", nil)
		return
	}

	if err := h.stateMgr.UpdateSystemSettings(&settings); err != nil {
		ctx.Logger.WithError(err).Error("Failed to update system settings")
		response.InternalServerError(c, "更新系统设置失败", err)
		return
	}

	// 同步到配置文件
	h.configMgr.Viper.Set("system.check_interval", settings.CheckInterval)
	h.configMgr.Viper.Set("system.max_inactive_time", settings.MaxInactiveTime)
	h.configMgr.Viper.Set("system.timezone", settings.Timezone)

	configPath := h.configMgr.Viper.ConfigFileUsed()
	if err := h.configMgr.Viper.WriteConfigAs(configPath); err != nil {
		ctx.Logger.WithError(err).Error("Failed to save config file")
		// 不影响主要功能，只记录错误
	}

	h.stateMgr.LogStateChange("update_settings", "System settings updated")

	response.Success(c, "系统设置更新成功", settings)
}
