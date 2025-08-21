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
	stateMgr  *state.StateManager
	emailSvc  *email.EmailService
	configMgr *config.ConfigManager
	logger    *logrus.Logger
	exitChan  <-chan struct{}
}

// NewTaskScheduler 创建新的任务调度器
func NewTaskScheduler(
	stateMgr *state.StateManager,
	emailSvc *email.EmailService,
	configMgr *config.ConfigManager,
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

// startStatusCheckTask 启动状态检查任务
func (ts *TaskScheduler) startStatusCheckTask() {
	systemConfig := ts.configMgr.GetSystemConfig()

	go func() {
		ticker := time.NewTicker(systemConfig.CheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 检查是否应该发送遗书
				shouldSend, err := ts.stateMgr.ShouldSendWillMessage()
				if err != nil {
					ts.logger.WithError(err).Error("Failed to check if should send will message")
					continue
				}

				if shouldSend {
					ts.logger.Info("Sending will messages due to inactivity")
					hasWill, err := ts.stateMgr.GetUnsentWillMessages()
					if err != nil {
						ts.logger.WithError(err).Error("Failed to check unsent will messages")
						continue
					}

					if hasWill {
						result := ts.emailSvc.SendWillMessage()
						if result.Success {
							if err := ts.stateMgr.MarkWillAsSent(); err != nil {
								ts.logger.WithError(err).Error("Failed to mark will as sent")
							}
							ts.logger.Info("Will message sent successfully")
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
	go func() {
		ticker := time.NewTicker(30 * time.Minute) // 每30分钟重试一次
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := ts.emailSvc.RetryFailedEmails(); err != nil {
					ts.logger.WithError(err).Error("Failed to retry failed emails")
				} else {
					ts.logger.Info("Email retry task completed")
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
		ticker := time.NewTicker(5 * time.Minute) // 每5分钟检查一次配置变化
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
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