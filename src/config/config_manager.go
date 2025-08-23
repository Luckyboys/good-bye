package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Manager 配置管理器
type Manager struct {
	Viper  *viper.Viper
	logger *logrus.Logger
}

// NewConfigManager 创建新的配置管理器
func NewConfigManager(logger *logrus.Logger) *Manager {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	// 设置环境变量支持
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 设置默认值
	setDefaults(v)

	return &Manager{
		Viper:  v,
		logger: logger,
	}
}

// setDefaults 设置默认配置值
func setDefaults(v *viper.Viper) {
	// 服务器配置
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.read_timeout", 30)
	v.SetDefault("server.write_timeout", 30)

	// 系统配置
	v.SetDefault("system.check_interval", time.Hour*24)
	v.SetDefault("system.max_inactive_days", 7)
	v.SetDefault("system.timezone", "Asia/Shanghai")

	// 日志配置
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "text")
	v.SetDefault("log.output", "stdout")

	// 邮件配置
	v.SetDefault("email.smtp_host", "smtp.gmail.com")
	v.SetDefault("email.smtp_port", 587)
	v.SetDefault("email.username", "")
	v.SetDefault("email.password", "")
	v.SetDefault("email.from_email", "")
	v.SetDefault("email.test_email", "")

	// 默认收件人列表
	v.SetDefault("email.recipients", []map[string]string{
		{"email": "", "name": ""},
	})

	// 部署配置
	v.SetDefault("deployment.posthumous_papers_file", "./data/posthumous_papers.md")
}

// LoadConfig 加载配置文件
func (cm *Manager) LoadConfig() error {
	// 确保配置目录存在
	if err := os.MkdirAll("./config", 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 尝试读取配置文件
	if err := cm.Viper.ReadInConfig(); err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			// 配置文件不存在，创建默认配置
			cm.logger.Info("Config file not found, creating default config")
			if err := cm.createDefaultConfig(); err != nil {
				return fmt.Errorf("failed to create default config: %w", err)
			}
		} else {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	cm.logger.Info("Configuration loaded successfully")
	return nil
}

// createDefaultConfig 创建默认配置文件
func (cm *Manager) createDefaultConfig() error {
	configPath := "./config/config.yaml"

	// 创建默认配置内容
	defaultConfig := `# 服务器配置
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: 30
  write_timeout: 30

# 系统配置
system:
  check_interval: "24h"     # 检查间隔（时间间隔格式）
  max_inactive_days: 7      # 最大不活跃天数
  timezone: "Asia/Shanghai"  # 时区

# 日志配置
log:
  level: "info"
  format: "text"
  output: "stdout"

# 邮件配置
email:
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: ""
  password: ""
  from_email: ""
  test_email: ""
  
  # 收件人列表
  recipients:
    - email: ""
      name: ""

# 部署配置
deployment:
  posthumous_papers_file: "./data/posthumous_papers.md"  # 遗书文件路径
`

	// 写入配置文件
	if err := os.WriteFile(configPath, []byte(defaultConfig), 0600); err != nil {
		return fmt.Errorf("failed to write default config file: %w", err)
	}

	// 重新读取配置文件
	return cm.Viper.ReadInConfig()
}

// GetServerConfig 获取服务器配置
func (cm *Manager) GetServerConfig() ServerConfig {
	return ServerConfig{
		Port:         cm.Viper.GetInt("server.port"),
		Host:         cm.Viper.GetString("server.host"),
		ReadTimeout:  cm.Viper.GetInt("server.read_timeout"),
		WriteTimeout: cm.Viper.GetInt("server.write_timeout"),
	}
}

// GetSystemConfig 获取系统配置
func (cm *Manager) GetSystemConfig() SystemConfig {
	return SystemConfig{
		CheckInterval:   cm.Viper.GetDuration("system.check_interval"),
		MaxInactiveDays: cm.Viper.GetInt("system.max_inactive_days"),
		Timezone:        cm.Viper.GetString("system.timezone"),
	}
}

// GetLogConfig 获取日志配置
func (cm *Manager) GetLogConfig() LogConfig {
	return LogConfig{
		Level:  cm.Viper.GetString("log.level"),
		Format: cm.Viper.GetString("log.format"),
		Output: cm.Viper.GetString("log.output"),
	}
}

// GetDeploymentConfig 获取部署配置
func (cm *Manager) GetDeploymentConfig() DeploymentConfig {
	return DeploymentConfig{
		PosthumousPapersFile: cm.Viper.GetString("deployment.posthumous_papers_file"),
	}
}

// GetEmailConfig 获取邮件配置（不包含敏感信息）
func (cm *Manager) GetEmailConfig() EmailConfig {
	var recipients []EmailRecipient
	if err := cm.Viper.UnmarshalKey("email.recipients", &recipients); err != nil {
		cm.logger.WithError(err).Warn("Failed to unmarshal email recipients, using empty list")
		recipients = []EmailRecipient{}
	}

	cm.logger.WithField("recipients_count", len(recipients)).Info("Email recipients loaded")

	return EmailConfig{
		SMTPHost:   cm.Viper.GetString("email.smtp_host"),
		SMTPPort:   cm.Viper.GetInt("email.smtp_port"),
		Username:   cm.Viper.GetString("email.username"),
		Password:   "", // 不返回密码
		FromEmail:  cm.Viper.GetString("email.from_email"),
		TestEmail:  cm.Viper.GetString("email.test_email"),
		Recipients: recipients,
	}
}

// UpdateConfig 更新配置
func (cm *Manager) UpdateConfig(key string, value any) error {
	cm.Viper.Set(key, value)

	// 保存到配置文件
	configPath := cm.Viper.ConfigFileUsed()
	if err := cm.Viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	cm.logger.WithField("key", key).Info("Configuration updated")
	return nil
}

// GetConfig 获取配置值
func (cm *Manager) GetConfig(key string) any {
	return cm.Viper.Get(key)
}

// GetString 获取字符串配置
func (cm *Manager) GetString(key string) string {
	return cm.Viper.GetString(key)
}

// GetInt 获取整数配置
func (cm *Manager) GetInt(key string) int {
	return cm.Viper.GetInt(key)
}

// GetBool 获取布尔配置
func (cm *Manager) GetBool(key string) bool {
	return cm.Viper.GetBool(key)
}

// SyncToDatabase 同步配置到数据库
func (cm *Manager) SyncToDatabase() error {
	// 不再需要数据库同步
	return nil
}

// ValidatePosthumousPapersFile 验证遗书文件是否存在
func (cm *Manager) ValidatePosthumousPapersFile() error {
	posthumousPapersFile := cm.Viper.GetString("deployment.posthumous_papers_file")
	if posthumousPapersFile == "" {
		return fmt.Errorf("posthumous papers file path is not configured")
	}

	if _, err := os.Stat(posthumousPapersFile); os.IsNotExist(err) {
		return fmt.Errorf("posthumous papers file does not exist: %s", posthumousPapersFile)
	}

	return nil
}

// ValidateConfig 验证配置
func (cm *Manager) ValidateConfig() error {
	// 验证端口
	port := cm.Viper.GetInt("server.port")
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid server port: %d", port)
	}

	// 验证SMTP端口
	smtpPort := cm.Viper.GetInt("email.smtp_port")
	if smtpPort < 1 || smtpPort > 65535 {
		return fmt.Errorf("invalid SMTP port: %d", smtpPort)
	}

	// 验证检查间隔
	checkInterval := cm.Viper.GetDuration("system.check_interval")
	if checkInterval < time.Minute {
		return fmt.Errorf("invalid check interval: %v", checkInterval)
	}

	// 验证最大不活跃天数
	maxInactiveDays := cm.Viper.GetInt("system.max_inactive_days")
	if maxInactiveDays < 1 {
		return fmt.Errorf("invalid max inactive days: %d", maxInactiveDays)
	}

	// 验证遗书文件路径
	posthumousPapersFile := cm.Viper.GetString("deployment.posthumous_papers_file")
	if posthumousPapersFile == "" {
		return fmt.Errorf("posthumous papers file path is required")
	}

	return nil
}

// GetConfigPath 获取配置文件路径
func (cm *Manager) GetConfigPath() string {
	return cm.Viper.ConfigFileUsed()
}

// ReloadConfig 重新加载配置
func (cm *Manager) ReloadConfig() error {
	if err := cm.Viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	// 验证配置
	if err := cm.ValidateConfig(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	cm.logger.Info("Configuration reloaded successfully")
	return nil
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// SystemConfig 系统配置
type SystemConfig struct {
	CheckInterval   time.Duration `mapstructure:"check_interval"`
	MaxInactiveDays int           `mapstructure:"max_inactive_days"`
	Timezone        string        `mapstructure:"timezone"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// DeploymentConfig 部署配置
type DeploymentConfig struct {
	PosthumousPapersFile string `mapstructure:"posthumous_papers_file"`
}

// EmailRecipient 邮件收件人
type EmailRecipient struct {
	Email string `mapstructure:"email"`
	Name  string `mapstructure:"name"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost   string           `mapstructure:"smtp_host"`
	SMTPPort   int              `mapstructure:"smtp_port"`
	Username   string           `mapstructure:"username"`
	Password   string           `mapstructure:"password"`
	FromEmail  string           `mapstructure:"from_email"`
	TestEmail  string           `mapstructure:"test_email"`
	Recipients []EmailRecipient `mapstructure:"recipients"`
}
