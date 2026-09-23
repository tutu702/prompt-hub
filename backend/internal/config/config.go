// Package config 提供基于 Viper 的应用配置加载能力。
//
// 优先级：环境变量 > .env > 配置文件 > 默认值
// 配置文件默认路径：./configs/config.yaml，可通过 --config 指定。
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 是整个应用配置的根结构。
type Config struct {
	App    App
	Server Server
	Log    Logger
	DB     Database
	JWT    JWT
	LLM    LLMConfig
}

// App 应用级别元信息。
type App struct {
	Name        string
	Env         string
	Version     string
	Description string
}

// Server HTTP 服务配置。
type Server struct {
	Host            string
	Port            int
	Mode            string        // gin 运行模式：debug / release / test
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// Logger 日志配置。
type Logger struct {
	Level      string // debug / info / warn / error
	Format     string // json / console
	OutputPath string
}

// Database 数据库配置。
type Database struct {
	Driver          string        // postgres / mysql
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string        // postgres: disable / require / verify-full
	Charset         string        // mysql: utf8mb4
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	LogLevel        string        // silent / error / warn / info
}

// JWT JWT 配置（预留）。
type JWT struct {
	Secret         string
	AccessExpire   time.Duration
	RefreshExpire  time.Duration
	Issuer         string
}

// LLMConfig 大模型配置（P08）：API Key 加密密钥（AES-GCM）。
type LLMConfig struct {
	AESKey string `mapstructure:"aes_key"`
}

// Load 按优先级加载配置。
func Load() (*Config, error) {
	v := viper.New()

	// 默认值
	v.SetDefault("app.name", "PromptHub")
	v.SetDefault("app.env", "development")
	v.SetDefault("app.version", "0.1.0")
	v.SetDefault("app.description", "AI 提示词资产管理工具后端")

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.read_timeout", "15s")
	v.SetDefault("server.write_timeout", "15s")
	v.SetDefault("server.shutdown_timeout", "10s")

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "console")
	v.SetDefault("log.output_path", "stdout")

	v.SetDefault("db.driver", "postgres")
	v.SetDefault("db.host", "127.0.0.1")
	v.SetDefault("db.port", 5432)
	v.SetDefault("db.user", "postgres")
	v.SetDefault("db.password", "postgres")
	v.SetDefault("db.dbname", "prompthub")
	v.SetDefault("db.sslmode", "disable")
	v.SetDefault("db.charset", "utf8mb4")
	v.SetDefault("db.max_open_conns", 50)
	v.SetDefault("db.max_idle_conns", 10)
	v.SetDefault("db.conn_max_lifetime", "1h")
	v.SetDefault("db.log_level", "warn")

	v.SetDefault("jwt.secret", "change-me-in-production")
	v.SetDefault("jwt.access_expire", "2h")
	v.SetDefault("jwt.refresh_expire", "168h")
	v.SetDefault("jwt.issuer", "prompthub")

	// P08 LLM 配置
	v.SetDefault("llm.aes_key", "")

	// 配置文件
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")
	if err := v.ReadInConfig(); err != nil {
		// 配置文件可省略，加载 .env 即可
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	// 环境变量
	v.SetEnvPrefix("PROMPTHUB")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 修复：Viper v1.21 + mapstructure/v2 默认不会把 "2h" 解析成 time.Duration，
	// 导致 cfg.JWT.AccessExpire / RefreshExpire 始终为 0，进而让登录响应的
	// expires_in 也变成 0，前端 token 立即"过期"。
	// 用 viper 自带的 GetDuration 显式覆盖一次。
	cfg.JWT.AccessExpire = v.GetDuration("jwt.access_expire")
	cfg.JWT.RefreshExpire = v.GetDuration("jwt.refresh_expire")

	return cfg, nil
}
