package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Luckyboys/good-bye/src/api/response"
	"github.com/Luckyboys/good-bye/src/config"
	"github.com/Luckyboys/good-bye/src/email"
	"github.com/Luckyboys/good-bye/src/state"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Context API上下文
type Context struct {
	StateMgr  *state.Manager
	EmailSvc  *email.Service
	ConfigMgr *config.Manager
	Logger    *logrus.Logger
}

// ErrorHandler 错误处理中间件
func ErrorHandler(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		// 只记录错误请求
		if c.Writer.Status() >= 400 {
			end := time.Now()
			latency := end.Sub(start)

			if raw != "" {
				path = path + "?" + raw
			}

			logger.WithFields(logrus.Fields{
				"status":     c.Writer.Status(),
				"method":     c.Request.Method,
				"path":       path,
				"ip":         c.ClientIP(),
				"latency":    latency,
				"user_agent": c.Request.UserAgent(),
				"error":      c.Errors.String(),
			}).Error("HTTP request failed")
		}
	}
}

// Logger 日志中间件
func Logger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		if raw != "" {
			path = path + "?" + raw
		}

		logger.WithFields(logrus.Fields{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       path,
			"ip":         c.ClientIP(),
			"latency":    latency,
			"user_agent": c.Request.UserAgent(),
		}).Info("HTTP request")
	}
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Timeout 超时中间件
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置超时上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 替换请求上下文
		c.Request = c.Request.WithContext(ctx)

		done := make(chan bool, 1)
		go func() {
			c.Next()
			done <- true
		}()

		select {
		case <-done:
			return
		case <-ctx.Done():
			c.JSON(http.StatusRequestTimeout, response.APIResponse{
				Success: false,
				Message: "Request timeout",
				Error:   "Request processing timeout",
			})
			c.Abort()
			return
		}
	}
}

// Recovery 恢复中间件
func Recovery(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic错误
				logger.WithField("error", err).Error("Panic recovered in HTTP handler")

				// 发送错误响应
				response.InternalServerError(c, "Internal server error", fmt.Errorf("panic: %v", err))
				c.Abort()
			}
		}()

		c.Next()
	}
}

// WithAPIContext 为路由添加API上下文
func WithAPIContext(ctx *Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("apiContext", ctx)
		c.Next()
	}
}

// GetAPIContext 从gin上下文中获取API上下文
func GetAPIContext(c *gin.Context) (*Context, bool) {
	value, exists := c.Get("apiContext")
	if !exists {
		return nil, false
	}

	ctx, ok := value.(*Context)
	return ctx, ok
}
