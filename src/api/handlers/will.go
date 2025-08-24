package handlers

import (
	"github.com/Luckyboys/good-bye/src/api/middleware"
	"github.com/Luckyboys/good-bye/src/api/response"
	"github.com/Luckyboys/good-bye/src/email"
	"github.com/Luckyboys/good-bye/src/state"
	"github.com/gin-gonic/gin"
)

// WillHandler 遗书管理处理器
type WillHandler struct {
	stateMgr *state.Manager
	emailSvc *email.Service
}

// NewWillHandler 创建新的遗书处理器
func NewWillHandler(stateMgr *state.Manager, emailSvc *email.Service) *WillHandler {
	return &WillHandler{
		stateMgr: stateMgr,
		emailSvc: emailSvc,
	}
}

// GetWillMessages 获取遗书状态（从文件读取）
func (h *WillHandler) GetWillMessages(c *gin.Context) {
	_, exists := middleware.GetAPIContext(c)
	if !exists {
		response.InternalServerError(c, "API context not found", nil)
		return
	}

	// 检查遗书文件是否存在
	content, err := h.stateMgr.ReadPosthumousPapers()
	if err != nil {
		response.NotFound(c, "遗书文件不存在或无法读取", err)
		return
	}

	data := map[string]any{
		"file_exists": true,
		"content":     content,
		"message":     "遗书内容已从文件读取",
	}

	response.Success(c, "获取遗书状态成功", data)
}

// SendTestWill 发送测试遗书到第一个收件人
func (h *WillHandler) SendTestWill(c *gin.Context) {
	ctx, exists := middleware.GetAPIContext(c)
	if !exists {
		response.InternalServerError(c, "API context not found", nil)
		return
	}

	result := h.emailSvc.SendWillToFirstRecipient()

	if result.Success {
		h.stateMgr.LogStateChange("send_test_will", "Test will sent successfully")
		response.Success(c, result.Message, nil)
	} else {
		ctx.Logger.WithError(result.Error).Error("Failed to send test will")
		response.InternalServerError(c, result.Message, result.Error)
	}
}

// SendWillMessage 发送遗书邮件（从文件读取内容）
func (h *WillHandler) SendWillMessage(c *gin.Context) {
	ctx, exists := middleware.GetAPIContext(c)
	if !exists {
		response.InternalServerError(c, "API context not found", nil)
		return
	}

	result := h.emailSvc.SendWillMessage()

	if result.Success {
		h.stateMgr.LogStateChange("send_will", "Will message sent successfully")
		response.Success(c, result.Message, nil)
	} else {
		ctx.Logger.WithError(result.Error).Error("Failed to send will message")
		response.InternalServerError(c, result.Message, result.Error)
	}
}
