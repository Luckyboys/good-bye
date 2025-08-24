package handlers

import (
	"time"

	"github.com/Luckyboys/good-bye/src/api/middleware"
	"github.com/Luckyboys/good-bye/src/api/response"
	"github.com/Luckyboys/good-bye/src/state"
	"github.com/gin-gonic/gin"
)

// StatusHandler 状态管理处理器
type StatusHandler struct {
	stateMgr *state.Manager
}

// NewStatusHandler 创建新的状态处理器
func NewStatusHandler(stateMgr *state.Manager) *StatusHandler {
	return &StatusHandler{
		stateMgr: stateMgr,
	}
}

// GetStatus 获取状态信息
func (h *StatusHandler) GetStatus(c *gin.Context) {
	ctx, exists := middleware.GetAPIContext(c)
	if !exists {
		response.InternalServerError(c, "API context not found", nil)
		return
	}

	status, err := h.stateMgr.GetStatus()
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to get status")
		response.InternalServerError(c, "获取状态失败", err)
		return
	}

	inactiveDuration, err := h.stateMgr.GetInactiveDuration()
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to get inactive duration")
		response.InternalServerError(c, "获取不活跃时长失败", err)
		return
	}

	settings, err := h.stateMgr.GetSystemSettings()
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to get system settings")
		response.InternalServerError(c, "获取系统设置失败", err)
		return
	}

	data := map[string]any{
		"last_seen":         status.LastSeen.Format(time.RFC3339),
		"inactive_duration": inactiveDuration.String(),
		"is_inactive":       h.stateMgr.IsInactiveWithStatus(status, settings.MaxInactiveDays),
		"max_inactive_days": settings.MaxInactiveDays,
		"check_interval":    settings.CheckInterval,
	}

	response.Success(c, "获取状态成功", data)
}

// CheckIn 签到接口
func (h *StatusHandler) CheckIn(c *gin.Context) {
	ctx, exists := middleware.GetAPIContext(c)
	if !exists {
		response.InternalServerError(c, "API context not found", nil)
		return
	}

	if err := h.stateMgr.UpdateStatus(); err != nil {
		ctx.Logger.WithError(err).Error("Failed to update status")
		response.InternalServerError(c, "签到失败", err)
		return
	}

	h.stateMgr.LogStateChange("check_in", "User checked in successfully")

	data := map[string]any{
		"timestamp": time.Now().Format(time.RFC3339),
	}

	response.Success(c, "签到成功", data)
}

// GetStats 获取统计信息
func (h *StatusHandler) GetStats(c *gin.Context) {
	ctx, exists := middleware.GetAPIContext(c)
	if !exists {
		response.InternalServerError(c, "API context not found", nil)
		return
	}

	stats, err := h.stateMgr.GetStats()
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to get stats")
		response.InternalServerError(c, "获取统计信息失败", err)
		return
	}

	response.Success(c, "获取统计信息成功", stats)
}
