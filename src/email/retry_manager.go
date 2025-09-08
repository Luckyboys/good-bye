package email

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// RetryManager 重试管理器
type RetryManager struct {
	logger           *logrus.Logger
	maxRetryCount    int
	maxRetryDuration time.Duration
	activeRetries    map[string]*RetryContext
	mu               sync.RWMutex
	sendFunc         func(Message) *Result
}

// RetryContext 重试上下文
type RetryContext struct {
	Message         Message
	CurrentRetry    int
	NextRetryTime   time.Time
	StartTime       time.Time
	LastError       error
	Cancel          context.CancelFunc
	SuccessCallback func()
	FailureCallback func(error)
}

// NewRetryManager 创建新的重试管理器
func NewRetryManager(logger *logrus.Logger) *RetryManager {
	return &RetryManager{
		logger:           logger,
		maxRetryCount:    DefaultMaxRetryCount,
		maxRetryDuration: DefaultMaxRetryDuration,
		activeRetries:    make(map[string]*RetryContext),
	}
}

// SetSendFunc 设置邮件发送函数
func (rm *RetryManager) SetSendFunc(sendFunc func(Message) *Result) {
	rm.sendFunc = sendFunc
}

// StartRetry 开始重试发送邮件
func (rm *RetryManager) StartRetry(message Message, successCallback func(), failureCallback func(error)) {
	retryKey := fmt.Sprintf("%s_%d", message.To, time.Now().Unix())

	ctx, cancel := context.WithCancel(context.Background())

	retryContext := &RetryContext{
		Message:         message,
		CurrentRetry:    0,
		StartTime:       time.Now(),
		Cancel:          cancel,
		SuccessCallback: successCallback,
		FailureCallback: failureCallback,
	}

	rm.mu.Lock()
	rm.activeRetries[retryKey] = retryContext
	rm.mu.Unlock()

	rm.logger.WithField("retry_key", retryKey).
		WithField("to", message.To).
		WithField("subject", message.Subject).
		Info("Starting email retry with exponential backoff")

	go rm.executeRetry(ctx, retryKey, retryContext)
}

// executeRetry 执行重试逻辑
func (rm *RetryManager) executeRetry(ctx context.Context, retryKey string, retryContext *RetryContext) {
	for {
		select {
		case <-ctx.Done():
			rm.logger.WithField("retry_key", retryKey).Info("Retry cancelled")
			rm.removeRetryContext(retryKey)
			return
		default:
			// 检查是否超过最大重试次数
			if retryContext.CurrentRetry >= rm.maxRetryCount {
				rm.logger.WithField("retry_key", retryKey).
					WithField("retry_count", retryContext.CurrentRetry).
					WithField("duration", time.Since(retryContext.StartTime)).
					Error("Max retry count reached, stopping retry attempts")

				if retryContext.FailureCallback != nil {
					retryContext.FailureCallback(fmt.Errorf(ErrorMaxRetryCountReached, rm.maxRetryCount))
				}
				rm.removeRetryContext(retryKey)
				return
			}

			// 检查是否超过最大重试时长
			if time.Since(retryContext.StartTime) > rm.maxRetryDuration {
				rm.logger.WithField("retry_key", retryKey).
					WithField("duration", time.Since(retryContext.StartTime)).
					Error("Max retry duration reached, stopping retry attempts")

				if retryContext.FailureCallback != nil {
					retryContext.FailureCallback(fmt.Errorf(ErrorMaxRetryDurationReached, rm.maxRetryDuration))
				}
				rm.removeRetryContext(retryKey)
				return
			}

			// 等待到下次重试时间
			now := time.Now()
			if now.Before(retryContext.NextRetryTime) {
				select {
				case <-time.After(retryContext.NextRetryTime.Sub(now)):
				case <-ctx.Done():
					rm.removeRetryContext(retryKey)
					return
				}
			}

			// 执行重试
			rm.logger.WithField("retry_key", retryKey).
				WithField("attempt", retryContext.CurrentRetry+1).
				WithField("delay", rm.calculateRetryDelay(retryContext.CurrentRetry)).
				Info("Attempting email retry")

			// 调用邮件发送函数
			if rm.sendFunc != nil {
				result := rm.sendFunc(retryContext.Message)
				if result.Success {
					rm.logger.WithField("retry_key", retryKey).
						WithField("attempt", retryContext.CurrentRetry+1).
						WithField("total_duration", time.Since(retryContext.StartTime)).
						Info("Email retry succeeded")

					if retryContext.SuccessCallback != nil {
						retryContext.SuccessCallback()
					}
					rm.removeRetryContext(retryKey)
					return
				}

				retryContext.LastError = result.Error
				rm.logger.WithField("retry_key", retryKey).
					WithField("attempt", retryContext.CurrentRetry+1).
					WithError(result.Error).
					Warn("Email retry attempt failed")
			}

			retryContext.CurrentRetry++

			// 计算下次重试时间
			retryContext.NextRetryTime = time.Now().Add(rm.calculateRetryDelay(retryContext.CurrentRetry))

			// 暂停一下，避免过于频繁的重试
			time.Sleep(RetryPauseDelay)
		}
	}
}

// calculateRetryDelay 计算重试延迟（指数退避）
func (rm *RetryManager) calculateRetryDelay(retryCount int) time.Duration {
	if retryCount == 0 {
		return InitialRetryDelay
	}

	// 指数退避：1s, 2s, 4s, 8s, 16s, 32s, 64s, 128s, 256s, 512s, 1024s, ...
	delay := time.Duration(math.Pow(2, float64(retryCount))) * time.Second

	// 最大延迟不超过1天
	if delay > MaxRetryDelay {
		delay = MaxRetryDelay
	}

	return delay
}

// removeRetryContext 移除重试上下文
func (rm *RetryManager) removeRetryContext(retryKey string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.activeRetries, retryKey)
}

// CancelRetry 取消重试
func (rm *RetryManager) CancelRetry(retryKey string) {
	rm.mu.RLock()
	retryContext, exists := rm.activeRetries[retryKey]
	rm.mu.RUnlock()

	if exists {
		retryContext.Cancel()
		rm.logger.WithField("retry_key", retryKey).Info("Retry cancelled by request")
	}
}

// GetRetryStatus 获取重试状态
func (rm *RetryManager) GetRetryStatus() map[string]any {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	status := make(map[string]any)
	status["active_retries"] = len(rm.activeRetries)
	status["max_retry_count"] = rm.maxRetryCount
	status["max_retry_duration"] = rm.maxRetryDuration.String()

	retries := make([]map[string]any, 0)
	for key, ctx := range rm.activeRetries {
		retryInfo := map[string]any{
			"retry_key":       key,
			"to":              ctx.Message.To,
			"subject":         ctx.Message.Subject,
			"current_retry":   ctx.CurrentRetry,
			"next_retry_time": ctx.NextRetryTime.Format(time.RFC3339),
			"start_time":      ctx.StartTime.Format(time.RFC3339),
			"duration":        time.Since(ctx.StartTime).String(),
		}
		retries = append(retries, retryInfo)
	}
	status["retries"] = retries

	return status
}

// StopAllRetries 停止所有重试
func (rm *RetryManager) StopAllRetries() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for key, ctx := range rm.activeRetries {
		ctx.Cancel()
		rm.logger.WithField("retry_key", key).Info("Retry cancelled during shutdown")
	}

	rm.activeRetries = make(map[string]*RetryContext)
}
