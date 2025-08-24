package api

import (
	"github.com/Luckyboys/good-bye/src/api/handlers"
	"github.com/Luckyboys/good-bye/src/api/middleware"
	"github.com/Luckyboys/good-bye/src/api/response"
	"github.com/Luckyboys/good-bye/src/config"
	"github.com/Luckyboys/good-bye/src/email"
	"github.com/Luckyboys/good-bye/src/state"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Handler API处理器（重构后的主处理器）
type Handler struct {
	healthHandler   *handlers.HealthHandler
	statusHandler   *handlers.StatusHandler
	settingsHandler *handlers.SettingsHandler
	emailHandler    *handlers.EmailHandler
	willHandler     *handlers.WillHandler
	apiContext      *middleware.Context
	router          *gin.Engine
	logger          *logrus.Logger
}

// NewAPIHandler 创建新的API处理器
func NewAPIHandler(stateMgr *state.Manager, emailSvc *email.Service, configMgr *config.Manager, logger *logrus.Logger) *Handler {
	// 创建API上下文
	apiContext := &middleware.Context{
		StateMgr:  stateMgr,
		EmailSvc:  emailSvc,
		ConfigMgr: configMgr,
		Logger:    logger,
	}

	// 创建各功能处理器
	healthHandler := handlers.NewHealthHandler(stateMgr)
	statusHandler := handlers.NewStatusHandler(stateMgr)
	settingsHandler := handlers.NewSettingsHandler(stateMgr, configMgr)
	emailHandler := handlers.NewEmailHandler(stateMgr, emailSvc, configMgr)
	willHandler := handlers.NewWillHandler(stateMgr, emailSvc)

	return &Handler{
		healthHandler:   healthHandler,
		statusHandler:   statusHandler,
		settingsHandler: settingsHandler,
		emailHandler:    emailHandler,
		willHandler:     willHandler,
		apiContext:      apiContext,
		logger:          logger,
	}
}

// SetupRoutes 设置API路由
func (h *Handler) SetupRoutes(router *gin.Engine) {
	h.router = router

	// 创建API路由组
	apiGroup := router.Group("/api/v1")

	// 应用中间件
	apiGroup.Use(
		middleware.CORSMiddleware(),
		middleware.WithAPIContext(h.apiContext),
	)

	// 健康检查路由
	apiGroup.GET("/health", h.healthHandler.HealthCheck)
	apiGroup.GET("/ready", h.healthHandler.Ready)
	apiGroup.GET("/live", h.healthHandler.Live)

	// 状态管理路由
	apiGroup.POST("/checkin", h.statusHandler.CheckIn)
	apiGroup.GET("/status", h.statusHandler.GetStatus)
	apiGroup.GET("/stats", h.statusHandler.GetStats)

	// 系统设置路由
	apiGroup.GET("/settings", h.settingsHandler.GetSystemSettings)
	apiGroup.PUT("/settings", h.settingsHandler.UpdateSystemSettings)

	// 遗书管理路由
	apiGroup.GET("/wills", h.willHandler.GetWillMessages)
	apiGroup.POST("/wills/send", h.willHandler.SendWillMessage)
	apiGroup.POST("/wills/test-send", h.willHandler.SendTestWill)

	// 邮件服务路由
	apiGroup.POST("/email/test", h.emailHandler.SendTestEmail)
	apiGroup.GET("/email/config", h.emailHandler.GetEmailConfig)
	apiGroup.PUT("/email/config", h.emailHandler.UpdateEmailConfig)
	apiGroup.POST("/email/config/test", h.emailHandler.TestEmailConfig)
}

// SetupMiddleware 设置全局中间件
func (h *Handler) SetupMiddleware(router *gin.Engine) {
	// 基础中间件
	router.Use(middleware.Recovery(h.logger))
	router.Use(middleware.Logger(h.logger))
	router.Use(middleware.ErrorHandler(h.logger))

	// 可选：超时中间件（根据需要启用）
	// router.Use(middleware.Timeout(30 * time.Second))
}

// GetRouter 获取路由器（用于兼容性）
func (h *Handler) GetRouter() *gin.Engine {
	return h.router
}

// Response 兼容性类型别名（为了与现有代码兼容）
type Response = response.APIResponse
