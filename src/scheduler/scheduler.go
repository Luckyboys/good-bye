package scheduler

import (
	"time"

	"github.com/Luckyboys/good-bye/src/config"
	"github.com/Luckyboys/good-bye/src/email"
	"github.com/Luckyboys/good-bye/src/state"
	"github.com/sirupsen/logrus"
)

// TaskScheduler 任务调度器
type TaskScheduler struct {
	stateMgr  *state.Manager
	emailSvc  *email.Service
	configMgr *config.Manager
	logger    *logrus.Logger
	exitChan  <-chan struct{}
}

// NewTaskScheduler 创建新的任务调度器
func NewTaskScheduler(
	stateMgr *state.Manager,
	emailSvc *email.Service,
	configMgr *config.Manager,
	logger *logrus.Logger,
	exitChan <-chan struct{},
) *TaskScheduler {
	return &TaskScheduler{
		stateMgr:  stateMgr,
		emailSvc:  emailSvc,
		configMgr: configMgr,
		logger:    logger,
		exitChan:  exitChan,
	}
}

// Start 启动所有后台任务
func (ts *TaskScheduler) Start() {
	ts.logger.Info("Starting background tasks")

	// 启动状态检查任务
	ts.startStatusCheckTask()

	// 启动邮件重试任务
	ts.startEmailRetryTask()

	// 启动配置重载任务
	ts.startConfigReloadTask()
}

// Stop 停止所有后台任务
func (ts *TaskScheduler) Stop() {
	ts.logger.Info("Stopping all background tasks")

	// 停止所有邮件重试
	ts.emailSvc.StopAllRetries()

	ts.logger.Info("All background tasks stopped")
}

// startStatusCheckTask 启动状态检查任务
func (ts *TaskScheduler) startStatusCheckTask() {
	systemConfig := ts.configMgr.GetSystemConfig()

	go func() {
		ts.logger.WithField("checkInterval", systemConfig.CheckInterval).Info("Starting status check task")
		defer ts.logger.Info("Status check task exiting")

		ticker := time.NewTicker(systemConfig.CheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:

				ts.logger.Debug("Checking status")

				// 检查状态检查是否已停止
				if ts.stateMgr.IsCheckingStopped() {
					ts.logger.Debug("Status checking is stopped, skipping check")
					continue
				}

				// 检查是否应该发送提醒
				shouldSendReminder, willSendTime, err := ts.stateMgr.ShouldSendReminder()
				if err != nil {
					ts.logger.WithError(err).Error("Failed to check if should send reminder")
					continue
				}
				ts.logger.
					WithField("shouldSendReminder", shouldSendReminder).
					WithField("willSendTime", willSendTime).
					Debug("checked should be send reminder")

				if shouldSendReminder {
					ts.logger.Info("Sending reminder email due to inactivity")
					systemConfig := ts.configMgr.GetSystemConfig()
					result := ts.emailSvc.SendReminderEmail(systemConfig.ReminderTime, willSendTime)
					if result.Success {
						if err := ts.stateMgr.MarkReminderAsSent(); err != nil {
							ts.logger.WithError(err).Error("Failed to mark reminder as sent")
						}
						ts.logger.Info("Reminder email sent successfully")
					} else {
						ts.logger.WithError(result.Error).Error("Failed to send reminder email")
					}
				}

				// 检查是否应该发送遗书
				shouldSendWill, err := ts.stateMgr.ShouldSendWillMessage()
				if err != nil {
					ts.logger.WithError(err).Error("Failed to check if should send will message")
					continue
				}
				ts.logger.
					WithField("shouldSendWill", shouldSendWill).
					Debug("checked should be send will")

				if shouldSendWill {
					ts.logger.Info("Sending will messages due to inactivity")
					hasWill, err := ts.stateMgr.GetUnsentWillMessages()
					if err != nil {
						ts.logger.WithError(err).Error("Failed to check unsent will messages")
						continue
					}
					ts.logger.WithField("hasWill", hasWill).Debug("checked has will")

					if hasWill {
						result := ts.emailSvc.SendWillMessage()
						if result.Success {
							if err := ts.stateMgr.MarkWillAsSent(); err != nil {
								ts.logger.WithError(err).Error("Failed to mark will as sent")
							}
							ts.logger.Info("Will message sent successfully, status checking stopped")
						} else {
							ts.logger.WithError(result.Error).Error("Failed to send will message")
						}
					}
				}
			case <-ts.exitChan:
				ts.logger.Info("Status check task exiting")
				return
			}
		}
	}()
}

// startEmailRetryTask 启动邮件重试任务
func (ts *TaskScheduler) startEmailRetryTask() {
	systemConfig := ts.configMgr.GetSystemConfig()

	go func() {
		ts.logger.WithField("checkInterval", systemConfig.CheckInterval).Info("Starting email retry task")
		defer ts.logger.Info("Email retry task exiting")

		ticker := time.NewTicker(systemConfig.CheckInterval) // 每5分钟记录一次重试状态
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ts.logger.Debug("Checking email retry status")

				// 记录重试状态
				if retryStatus := ts.emailSvc.GetRetryStatus(); retryStatus != nil {
					ts.logger.WithField("active_retries", retryStatus["active_retries"]).
						Debug("Email retry status check")
				}
			case <-ts.exitChan:
				ts.logger.Info("Email retry task exiting")
				return
			}
		}
	}()
}

// startConfigReloadTask 启动配置重载任务
func (ts *TaskScheduler) startConfigReloadTask() {
	go func() {
		ts.logger.Info("Starting config reload task")
		defer ts.logger.Info("Config reload task exiting")

		ticker := time.NewTicker(5 * time.Minute) // 每5分钟检查一次配置变化
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ts.logger.Debug("Checking config reload status")

				if err := ts.configMgr.ReloadConfig(); err != nil {
					ts.logger.WithError(err).Error("Failed to reload config")
				}
			case <-ts.exitChan:
				ts.logger.Info("Config reload task exiting")
				return
			}
		}
	}()
}
