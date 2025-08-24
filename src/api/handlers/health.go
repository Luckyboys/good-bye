package handlers

import (
	"net/http"

	"github.com/Luckyboys/good-bye/src/api/middleware"
	"github.com/Luckyboys/good-bye/src/api/response"
	"github.com/Luckyboys/good-bye/src/state"
	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	stateMgr *state.Manager
}

// NewHealthHandler 创建新的健康检查处理器
func NewHealthHandler(stateMgr *state.Manager) *HealthHandler {
	return &HealthHandler{
		stateMgr: stateMgr,
	}
}

// HealthCheck 健康检查
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	_, exists := middleware.GetAPIContext(c)
	if !exists {
		response.InternalServerError(c, "API context not found", nil)
		return
	}

	health := h.stateMgr.GetHealthStatus()

	response.Success(c, "Service is healthy", health)
}

// Ready 就绪检查
func (h *HealthHandler) Ready(c *gin.Context) {
	_, exists := middleware.GetAPIContext(c)
	if !exists {
		response.InternalServerError(c, "API context not found", nil)
		return
	}

	// 检查服务是否完全就绪
	health := h.stateMgr.GetHealthStatus()

	// 可以添加更多就绪检查逻辑
	isReady := health["status"] == "healthy"

	if isReady {
		response.Success(c, "Service is ready", map[string]string{
			"status": "ready",
		})
	} else {
		response.ServiceUnavailable(c, "Service is not ready")
	}
}

// Live 存活检查
func (h *HealthHandler) Live(c *gin.Context) {
	// 简单的存活检查，只返回200状态码
	c.JSON(http.StatusOK, map[string]string{
		"status": "alive",
	})
}
