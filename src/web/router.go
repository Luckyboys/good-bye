package web

import (
	"github.com/Luckyboys/good-bye/src/api"
	"github.com/Luckyboys/good-bye/src/config"
	"github.com/Luckyboys/good-bye/src/email"
	"github.com/Luckyboys/good-bye/src/state"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Router struct {
	engine     *gin.Engine
	apiHandler *api.Handler
	logger     *logrus.Logger
}

func NewRouter(stateMgr *state.StateManager, emailSvc *email.EmailService, configMgr *config.ConfigManager, logger *logrus.Logger) *Router {
	// 设置Gin模式
	if configMgr.GetString("log.level") == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// 添加中间件
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	// 创建CORS中间件
	engine.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	apiHandler := api.NewAPIHandler(stateMgr, emailSvc, configMgr, logger)

	router := &Router{
		engine:     engine,
		apiHandler: apiHandler,
		logger:     logger,
	}

	router.setupRoutes()
	return router
}

func (r *Router) setupRoutes() {
	// 健康检查
	r.engine.GET("/health", r.apiHandler.HealthCheck)

	// API路由组
	v1 := r.engine.Group("/api/v1")
	{
		// 状态相关
		v1.POST("/checkin", r.apiHandler.CheckIn)
		v1.GET("/status", r.apiHandler.GetStatus)

		// 系统设置
		v1.GET("/settings", r.apiHandler.GetSystemSettings)
		v1.PUT("/settings", r.apiHandler.UpdateSystemSettings)

		// 遗书状态
		v1.GET("/wills", r.apiHandler.GetWillMessages)

		// 邮件相关
		v1.POST("/email/test", r.apiHandler.SendTestEmail)
		v1.GET("/email/config", r.apiHandler.GetEmailConfig)
		v1.PUT("/email/config", r.apiHandler.UpdateEmailConfig)
		v1.POST("/email/config/test", r.apiHandler.TestEmailConfig)

		// 统计信息
		v1.GET("/stats", r.apiHandler.GetStats)
	}

	// 静态文件服务
	r.engine.Static("/static", "./static")
	r.engine.LoadHTMLGlob("templates/*.html")

	// 首页
	r.engine.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"title": "遗书服务",
		})
	})

	// 签到页面
	r.engine.GET("/checkin", func(c *gin.Context) {
		c.HTML(200, "checkin.html", gin.H{
			"title": "签到",
		})
	})

	// 配置页面
	r.engine.GET("/config", func(c *gin.Context) {
		c.HTML(200, "config.html", gin.H{
			"title": "配置管理",
		})
	})

	// 404处理
	r.engine.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error": "页面不存在",
		})
	})
}

func (r *Router) Run(addr string) error {
	r.logger.Infof("Starting server on %s", addr)
	return r.engine.Run(addr)
}

func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
