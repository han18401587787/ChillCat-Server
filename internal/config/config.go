package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Log      LogConfig      `yaml:"log"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port string `yaml:"port" default:"8080"`
	Mode string `yaml:"mode" default:"debug"` // debug / release / test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `yaml:"host" default:"localhost"`
	Port     int    `yaml:"port" default:"5432"`
	User     string `yaml:"user" default:"chillcat"`
	Password string `yaml:"password" default:"chillcat_pass"`
	DBName   string `yaml:"dbname" default:"chillcat"`
	SSLMode  string `yaml:"sslmode" default:"disable"`
	MaxOpen  int    `yaml:"max_open" default:"25"`
	MaxIdle  int    `yaml:"max_idle" default:"10"`
}

// DSN 获取数据库连接字符串
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `yaml:"host" default:"localhost"`
	Port     int    `yaml:"port" default:"6379"`
	Password string `yaml:"password" default:""`
	DB       int    `yaml:"db" default:"0"`
}

// Addr 获取 Redis 地址
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret     string `yaml:"secret" default:"chillcat-jwt-secret"`
	ExpireHour int    `yaml:"expire_hour" default:"72"`
	Issuer     string `yaml:"issuer" default:"chillcat"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `yaml:"level" default:"info"`    // debug / info / warn / error
	Format string `yaml:"format" default:"console"` // console / json
}

// Load 加载配置，优先级：环境变量 > 配置文件 > 默认值
func Load() (*Config, error) {
	cfg := &Config{}

	// 尝试读取配置文件
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("解析配置文件失败: %w", err)
		}
	}

	// 环境变量覆盖（仅覆盖关键字段）
	cfg.Server.Port = getEnv("SERVER_PORT", cfg.Server.Port)
	cfg.Server.Mode = getEnv("SERVER_MODE", cfg.Server.Mode)
	cfg.Database.Host = getEnv("DB_HOST", cfg.Database.Host)
	cfg.Database.Port = getEnvInt("DB_PORT", cfg.Database.Port)
	cfg.Database.User = getEnv("DB_USER", cfg.Database.User)
	cfg.Database.Password = getEnv("DB_PASSWORD", cfg.Database.Password)
	cfg.Database.DBName = getEnv("DB_NAME", cfg.Database.DBName)
	cfg.Redis.Host = getEnv("REDIS_HOST", cfg.Redis.Host)
	cfg.Redis.Port = getEnvInt("REDIS_PORT", cfg.Redis.Port)
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", cfg.Redis.Password)
	cfg.JWT.Secret = getEnv("JWT_SECRET", cfg.JWT.Secret)

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
