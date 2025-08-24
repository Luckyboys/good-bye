package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse 标准API响应结构
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Builder 响应构建器
type Builder struct {
	response APIResponse
}

// NewResponseBuilder 创建新的响应构建器
func NewResponseBuilder() *Builder {
	return &Builder{
		response: APIResponse{
			Success: true,
		},
	}
}

// WithSuccess 设置成功状态
func (rb *Builder) WithSuccess(success bool) *Builder {
	rb.response.Success = success
	return rb
}

// WithMessage 设置消息
func (rb *Builder) WithMessage(message string) *Builder {
	rb.response.Message = message
	return rb
}

// WithData 设置数据
func (rb *Builder) WithData(data any) *Builder {
	rb.response.Data = data
	return rb
}

// WithError 设置错误
func (rb *Builder) WithError(err error) *Builder {
	rb.response.Success = false
	if err != nil {
		rb.response.Error = err.Error()
	}
	return rb
}

// WithErrorMessage 设置错误消息
func (rb *Builder) WithErrorMessage(message string) *Builder {
	rb.response.Success = false
	rb.response.Error = message
	return rb
}

// Build 构建响应
func (rb *Builder) Build() APIResponse {
	return rb.response
}

// SendResponse 发送响应
func SendResponse(c *gin.Context, statusCode int, response APIResponse) {
	c.JSON(statusCode, response)
}

// Success 成功响应
func Success(c *gin.Context, message string, data any) {
	response := NewResponseBuilder().
		WithSuccess(true).
		WithMessage(message).
		WithData(data).
		Build()
	SendResponse(c, http.StatusOK, response)
}

// Created 创建成功响应
func Created(c *gin.Context, message string, data any) {
	response := NewResponseBuilder().
		WithSuccess(true).
		WithMessage(message).
		WithData(data).
		Build()
	SendResponse(c, http.StatusCreated, response)
}

// BadRequest 400错误响应
func BadRequest(c *gin.Context, message string, err error) {
	response := NewResponseBuilder().
		WithSuccess(false).
		WithMessage(message).
		WithError(err).
		Build()
	SendResponse(c, http.StatusBadRequest, response)
}

// Unauthorized 401错误响应
func Unauthorized(c *gin.Context, message string) {
	response := NewResponseBuilder().
		WithSuccess(false).
		WithMessage(message).
		Build()
	SendResponse(c, http.StatusUnauthorized, response)
}

// Forbidden 403错误响应
func Forbidden(c *gin.Context, message string) {
	response := NewResponseBuilder().
		WithSuccess(false).
		WithMessage(message).
		Build()
	SendResponse(c, http.StatusForbidden, response)
}

// NotFound 404错误响应
func NotFound(c *gin.Context, message string, err error) {
	response := NewResponseBuilder().
		WithSuccess(false).
		WithMessage(message).
		WithError(err).
		Build()
	SendResponse(c, http.StatusNotFound, response)
}

// InternalServerError 500错误响应
func InternalServerError(c *gin.Context, message string, err error) {
	response := NewResponseBuilder().
		WithSuccess(false).
		WithMessage(message).
		WithError(err).
		Build()
	SendResponse(c, http.StatusInternalServerError, response)
}

// ServiceUnavailable 503错误响应
func ServiceUnavailable(c *gin.Context, message string) {
	response := NewResponseBuilder().
		WithSuccess(false).
		WithMessage(message).
		Build()
	SendResponse(c, http.StatusServiceUnavailable, response)
}
