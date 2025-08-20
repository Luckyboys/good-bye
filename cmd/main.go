package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Luckyboys/good-bye/src/config"
	"github.com/Luckyboys/good-bye/src/email"
	"github.com/Luckyboys/good-bye/src/state"
	"github.com/Luckyboys/good-bye/src/web"
	"github.com/sirupsen/logrus"
)

func main() {
	// 解析命令行参数
	var (
		configPath = flag.String("config", "", "配置文件路径")
		port       = flag.Int("port", 0, "服务端口（覆盖配置文件）")
		version    = flag.Bool("version", false, "显示版本信息")
	)
	flag.Parse()

	if *version {
		fmt.Println("Good-Bye Service v1.0.0")
		os.Exit(0)
	}

	// 初始化日志
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// 加载配置
	configMgr := config.NewConfigManager(logger)
	if *configPath != "" {
		configMgr.Viper.SetConfigFile(*configPath)
	}

	if err := configMgr.LoadConfig(); err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	// 验证配置
	if err := configMgr.ValidateConfig(); err != nil {
		logger.WithError(err).Fatal("Invalid configuration")
	}

	// 验证遗书文件是否存在
	if err := configMgr.ValidatePosthumousPapersFile(); err != nil {
		logger.WithError(err).Fatal("Posthumous papers file validation failed")
	}

	// 设置日志级别
	if level := configMgr.GetString("log.level"); level != "" {
		switch level {
		case "debug":
			logger.SetLevel(logrus.DebugLevel)
		case "info":
			logger.SetLevel(logrus.InfoLevel)
		case "warn":
			logger.SetLevel(logrus.WarnLevel)
		case "error":
			logger.SetLevel(logrus.ErrorLevel)
		}
	}

	// 获取服务器配置
	serverConfig := configMgr.GetServerConfig()
	if *port != 0 {
		serverConfig.Port = *port
	}

	// 不再需要数据库

	// 初始化服务
	dataDir := configMgr.GetString("deployment.data_dir")
	posthumousPapersFile := configMgr.GetString("deployment.posthumous_papers_file")
	stateMgr := state.NewStateManager(configMgr, logger, dataDir, posthumousPapersFile)
	emailSvc := email.NewEmailService(configMgr, stateMgr, logger)

	// 创建路由器
	router := web.NewRouter(stateMgr, emailSvc, configMgr, logger)

	// 启动时默认刷新签到时间
	logger.Info("Performing startup status refresh...")
	if err := stateMgr.UpdateStatus(); err != nil {
		logger.WithError(err).Error("Failed to update status on startup")
	} else {
		logger.Info("Startup status refresh completed successfully")
	}

	// 启动后台任务
	go startBackgroundTasks(stateMgr, emailSvc, configMgr, logger)

	// 启动HTTP服务器
	addr := fmt.Sprintf("%s:%d", serverConfig.Host, serverConfig.Port)
	logger.WithFields(logrus.Fields{
		"host": serverConfig.Host,
		"port": serverConfig.Port,
	}).Info("Starting HTTP server")

	// 优雅关闭处理
	go func() {
		if err := router.Run(addr); err != nil {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 执行清理工作
	if err := cleanup(logger); err != nil {
		logger.WithError(err).Error("Failed to cleanup")
	}

	logger.Info("Server exited")
}

// startBackgroundTasks 启动后台任务
func startBackgroundTasks(stateMgr *state.StateManager, emailSvc *email.EmailService, configMgr *config.ConfigManager, logger *logrus.Logger) {
	// 等待服务启动完成
	time.Sleep(5 * time.Second)

	logger.Info("Starting background tasks")

	// 状态检查任务
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // 每小时检查一次
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 检查是否应该发送遗书
				shouldSend, err := stateMgr.ShouldSendWillMessage()
				if err != nil {
					logger.WithError(err).Error("Failed to check if should send will message")
					continue
				}

				if shouldSend {
					logger.Info("Sending will messages due to inactivity")
					hasWill, err := stateMgr.GetUnsentWillMessages()
					if err != nil {
						logger.WithError(err).Error("Failed to check unsent will messages")
						continue
					}

					if hasWill {
						result := emailSvc.SendWillMessage()
						if result.Success {
							if err := stateMgr.MarkWillAsSent(); err != nil {
								logger.WithError(err).Error("Failed to mark will as sent")
							}
							logger.Info("Will message sent successfully")
						} else {
							logger.WithError(result.Error).Error("Failed to send will message")
						}
					}
				}
			}
		}
	}()

	// 邮件重试任务
	go func() {
		ticker := time.NewTicker(30 * time.Minute) // 每30分钟重试一次
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := emailSvc.RetryFailedEmails(); err != nil {
					logger.WithError(err).Error("Failed to retry failed emails")
				} else {
					logger.Info("Email retry task completed")
				}
			}
		}
	}()

	// 配置重载任务
	go func() {
		ticker := time.NewTicker(5 * time.Minute) // 每5分钟检查一次配置变化
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := configMgr.ReloadConfig(); err != nil {
					logger.WithError(err).Error("Failed to reload config")
				}
			}
		}
	}()
}

// cleanup 清理资源
func cleanup(logger *logrus.Logger) error {
	logger.Info("Performing cleanup...")
	logger.Info("Cleanup completed")
	return nil
}
