package state

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Luckyboys/good-bye/src/config"
	"github.com/sirupsen/logrus"
)

// Manager 状态管理器
type Manager struct {
	configMgr        *config.Manager
	logger           *logrus.Logger
	posthumousFile   string
	lastSeen         time.Time
	mu               sync.RWMutex
	checkingStopped  bool
	willSent         bool
	reminderSent     bool
	lastReminderTime time.Time
}

// NewStateManager 创建新的状态管理器
func NewStateManager(configMgr *config.Manager, logger *logrus.Logger, posthumousFile string) *Manager {
	return &Manager{
		configMgr:        configMgr,
		logger:           logger,
		posthumousFile:   posthumousFile,
		lastSeen:         time.Now(), // 初始化为当前时间
		checkingStopped:  false,
		willSent:         false,
		reminderSent:     false,
		lastReminderTime: time.Time{}, // 零值表示未发送过提醒
	}
}

// UpdateStatus 更新存活状态
func (sm *Manager) UpdateStatus() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.lastSeen = time.Now()
	// 重置提醒状态
	sm.ResetReminderSent()
	return nil
}

// Status 表示存活状态
type Status struct {
	LastSeen time.Time `json:"last_seen"`
}

// GetStatus 获取当前状态
func (sm *Manager) GetStatus() (*Status, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return &Status{
		LastSeen: sm.lastSeen,
	}, nil
}

// IsInactive 检查是否处于不活跃状态
func (sm *Manager) IsInactive() (bool, error) {
	systemConfig := sm.configMgr.GetSystemConfig()

	sm.mu.RLock()
	defer sm.mu.RUnlock()
	now := time.Now()
	duration := now.Sub(sm.lastSeen)
	return duration > systemConfig.MaxInactiveTime, nil
}

// IsInactiveWithStatus 检查给定状态是否处于不活跃状态
func (sm *Manager) IsInactiveWithStatus(status *Status, maxTime time.Duration) bool {
	now := time.Now()
	duration := now.Sub(status.LastSeen)
	return duration > maxTime
}

// GetInactiveDuration 获取不活跃时长
func (sm *Manager) GetInactiveDuration() (time.Duration, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return time.Since(sm.lastSeen), nil
}

// GetLastSeenTime 获取最后活跃时间
func (sm *Manager) GetLastSeenTime() (time.Time, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.lastSeen, nil
}

// GetSystemSettings 获取系统设置
func (sm *Manager) GetSystemSettings() (*config.SystemConfig, error) {
	systemConfig := sm.configMgr.GetSystemConfig()
	return &systemConfig, nil
}

// UpdateSystemSettings 更新系统设置
func (sm *Manager) UpdateSystemSettings(settings *config.SystemConfig) error {
	// 更新配置文件
	sm.configMgr.Viper.Set("system.check_interval", settings.CheckInterval)
	sm.configMgr.Viper.Set("system.max_inactive_time", settings.MaxInactiveTime)
	sm.configMgr.Viper.Set("system.timezone", settings.Timezone)

	configPath := sm.configMgr.Viper.ConfigFileUsed()
	return sm.configMgr.Viper.WriteConfigAs(configPath)
}

// ReadPosthumousPapers 从文件读取遗书内容
func (sm *Manager) ReadPosthumousPapers() (string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(sm.posthumousFile); os.IsNotExist(err) {
		return "", fmt.Errorf("posthumous papers file not found: %s", sm.posthumousFile)
	}

	// 读取文件内容
	content, err := os.ReadFile(sm.posthumousFile)
	if err != nil {
		return "", fmt.Errorf("failed to read posthumous papers file: %w", err)
	}

	return string(content), nil
}

// ShouldSendWillMessage 检查是否应该发送遗书消息
func (sm *Manager) ShouldSendWillMessage() (bool, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// 如果检查已停止，不再发送遗书
	if sm.checkingStopped {
		sm.logger.Info("Status checking has been stopped, will not send will message")
		return false, nil
	}

	// 如果遗书已发送，不再重复发送
	if sm.willSent {
		sm.logger.Info("Will message has already been sent, will not send again")
		return false, nil
	}

	// 检查是否处于不活跃状态
	isInactive, err := sm.IsInactive()
	if err != nil {
		return false, err
	}

	// 检查遗书文件是否存在
	if _, err := os.Stat(sm.posthumousFile); os.IsNotExist(err) {
		sm.logger.Warn("Posthumous papers file not found, will not send will message")
		return false, nil
	}

	return isInactive, nil
}

// ShouldSendReminder 检查是否应该发送提醒邮件
func (sm *Manager) ShouldSendReminder() (bool, time.Time, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// 如果检查已停止，不再发送提醒
	if sm.checkingStopped {
		sm.logger.Info("Status checking has been stopped, will not send reminder")
		return false, time.Time{}, nil
	}

	// 如果遗书已发送，不再发送提醒
	if sm.willSent {
		sm.logger.Info("Will message has already been sent, will not send reminder")
		return false, time.Time{}, nil
	}

	// 获取系统配置
	systemConfig := sm.configMgr.GetSystemConfig()

	// 如果提醒时间未配置（小于等于0），不发送提醒
	if systemConfig.ReminderTime <= 0 {
		return false, time.Time{}, nil
	}

	// 如果已经发送过提醒，不再重复发送
	if sm.reminderSent {
		sm.logger.Info("Reminder has already been sent, will not send again")
		return false, time.Time{}, nil
	}

	// 检查遗书文件是否存在
	if _, err := os.Stat(sm.posthumousFile); os.IsNotExist(err) {
		sm.logger.Warn("Posthumous papers file not found, will not send reminder")
		return false, time.Time{}, nil
	}

	// 计算应该发送遗书的时间
	now := time.Now()
	willSendTime := sm.lastSeen.Add(systemConfig.MaxInactiveTime)

	// 计算应该发送提醒的时间（发送遗书前reminder_time）
	reminderSendTime := willSendTime.Add(-systemConfig.ReminderTime)

	// 如果当前时间已经过了提醒发送时间，但还没有到发送遗书的时间，则需要发送提醒
	if now.After(reminderSendTime) && now.Before(willSendTime) {
		return true, willSendTime, nil
	}

	return false, willSendTime, nil
}

// MarkReminderAsSent 标记提醒为已发送
func (sm *Manager) MarkReminderAsSent() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.logger.Info("Reminder has been sent")
	sm.reminderSent = true
	sm.lastReminderTime = time.Now()

	return nil
}

// ResetReminderSent 重置提醒发送状态（在签到后调用）
func (sm *Manager) ResetReminderSent() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.reminderSent {
		sm.logger.Info("Reminder status reset due to user check-in")
		sm.reminderSent = false
		sm.lastReminderTime = time.Time{}
	}
}

// IsReminderSent 检查提醒是否已发送
func (sm *Manager) IsReminderSent() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.reminderSent
}

// GetLastReminderTime 获取最后提醒时间
func (sm *Manager) GetLastReminderTime() (time.Time, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.lastReminderTime, nil
}

// MarkWillAsSent 标记遗书为已发送
func (sm *Manager) MarkWillAsSent() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.logger.Info("Will message has been sent, stopping status checking")
	sm.willSent = true
	sm.checkingStopped = true

	return nil
}

// GetUnsentWillMessages 检查是否有未发送的遗书文件
func (sm *Manager) GetUnsentWillMessages() (bool, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// 如果检查已停止或遗书已发送，不再检查
	if sm.checkingStopped || sm.willSent {
		return false, nil
	}

	// 检查文件是否存在
	if _, err := os.Stat(sm.posthumousFile); os.IsNotExist(err) {
		return false, fmt.Errorf("posthumous papers file not found")
	}
	return true, nil
}

// StopStatusChecking 停止状态检查
func (sm *Manager) StopStatusChecking() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.logger.Info("Status checking stopped manually")
	sm.checkingStopped = true
}

// ResumeStatusChecking 恢复状态检查
func (sm *Manager) ResumeStatusChecking() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.logger.Info("Status checking resumed")
	sm.checkingStopped = false
	// 注意：不重置 willSent 标志，因为遗书发送后不应该恢复检查
}

// IsCheckingStopped 检查状态检查是否已停止
func (sm *Manager) IsCheckingStopped() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.checkingStopped
}

// IsWillSent 检查遗书是否已发送
func (sm *Manager) IsWillSent() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.willSent
}

// GetHealthStatus 获取健康状态
func (sm *Manager) GetHealthStatus() map[string]any {
	systemConfig := sm.configMgr.GetSystemConfig()

	sm.mu.RLock()
	defer sm.mu.RUnlock()
	inactiveDuration := time.Since(sm.lastSeen)
	now := time.Now()
	duration := now.Sub(sm.lastSeen)
	isInactive := duration > systemConfig.MaxInactiveTime

	health := map[string]any{
		"status":            "healthy",
		"last_seen":         sm.lastSeen.Format(time.RFC3339),
		"inactive_duration": inactiveDuration.String(),
		"is_inactive":       isInactive,
		"max_inactive_time": systemConfig.MaxInactiveTime,
		"check_interval":    systemConfig.CheckInterval,
	}

	return health
}

// LogStateChange 记录状态变化
func (sm *Manager) LogStateChange(action, details string) {
	sm.logger.WithFields(logrus.Fields{
		"action":  action,
		"details": details,
		"time":    time.Now().Format(time.RFC3339),
	}).Info("State changed")
}

// GetStats 获取统计信息
func (sm *Manager) GetStats() (map[string]any, error) {
	systemConfig := sm.configMgr.GetSystemConfig()

	// 检查遗书文件是否存在
	hasPosthumousFile := false
	if _, err := os.Stat(sm.posthumousFile); err == nil {
		hasPosthumousFile = true
	}

	var totalWills int64
	var unsentWills int64

	if hasPosthumousFile {
		totalWills = 1
		unsentWills = 1
	}

	sm.mu.RLock()
	defer sm.mu.RUnlock()
	stats := map[string]any{
		"last_seen":         sm.lastSeen.Format(time.RFC3339),
		"inactive_duration": time.Since(sm.lastSeen).String(),
		"system_settings":   systemConfig,
		"will_stats": map[string]any{
			"total":  totalWills,
			"sent":   int64(0),
			"unsent": unsentWills,
		},
		"email_stats": map[string]any{
			"total":  int64(0),
			"sent":   int64(0),
			"failed": int64(0),
		},
	}

	return stats, nil
}
