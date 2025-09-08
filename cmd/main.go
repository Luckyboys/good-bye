package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Luckyboys/good-bye/src/config"
	"github.com/Luckyboys/good-bye/src/email"
	"github.com/Luckyboys/good-bye/src/scheduler"
	"github.com/Luckyboys/good-bye/src/state"
	"github.com/Luckyboys/good-bye/src/web"
	"github.com/sirupsen/logrus"
)

// 版本信息（在构建时通过 ldflags 设置）
var (
	Version    = "v1.2.0"
	BuildTime  = "unknown"
	CommitHash = "unknown"
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
		fmt.Printf("Good-Bye Service %s\n", Version)
		os.Exit(0)
	}

	// 初始化日志
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{})
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

	// 根据配置重新配置日志设置
	setLogConfig(logger, configMgr)

	// 验证遗书文件是否存在
	if err := configMgr.ValidatePosthumousPapersFile(); err != nil {
		logger.WithError(err).Fatal("Posthumous papers file validation failed")
	}

	// 获取服务器配置
	serverConfig := configMgr.GetServerConfig()
	if *port != 0 {
		serverConfig.Port = *port
	}

	// 不再需要数据库

	// 初始化服务
	posthumousPapersFile := configMgr.GetString("deployment.posthumous_papers_file")
	stateMgr := state.NewStateManager(configMgr, logger, posthumousPapersFile)
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

	// 创建退出信号通道
	exitChan := make(chan struct{})

	// 启动后台任务
	taskScheduler := scheduler.NewTaskScheduler(stateMgr, emailSvc, configMgr, logger, exitChan)
	go taskScheduler.Start()

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
	cleanup(logger, exitChan)

	logger.Info("Server exited")
}

// cleanup 清理资源
func cleanup(logger *logrus.Logger, exitChan chan struct{}) {
	logger.Info("Performing cleanup...")

	// 关闭退出通道，通知所有后台任务退出
	close(exitChan)

	logger.Info("Cleanup completed")
}

func setLogConfig(logger *logrus.Logger, configMgr *config.Manager) {

	logConfig := configMgr.GetLogConfig()
	// 设置日志级别
	if level, err := logrus.ParseLevel(logConfig.Level); err == nil {
		logger.SetLevel(level)
	}

	// 设置日志格式
	switch logConfig.Format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{})
	default:
		logger.SetFormatter(&logrus.TextFormatter{})
	}

	// 设置日志输出
	switch logConfig.Output {
	case "stdout":
		logger.SetOutput(os.Stdout)
	case "stderr":
		logger.SetOutput(os.Stderr)
	}
}
