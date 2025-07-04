// Package config 提供配置管理功能，支持从文件和环境变量加载配置
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/XXXXD-cation/proxy-platform/pkg/logger"
	"gopkg.in/yaml.v3"
)

// Config 应用程序配置结构
type Config struct {
	Server    ServerConfig     `yaml:"server"`
	Database  DatabaseConfig   `yaml:"database"`
	Redis     RedisConfig      `yaml:"redis"`
	Log       logger.LogConfig `yaml:"log"`
	ProxyPool ProxyPoolConfig  `yaml:"proxy_pool"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	IdleTimeout  int    `yaml:"idle_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	Charset  string `yaml:"charset"`
	MaxIdle  int    `yaml:"max_idle"`
	MaxOpen  int    `yaml:"max_open"`
	MaxLife  int    `yaml:"max_life"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Password    string `yaml:"password"`
	DB          int    `yaml:"db"`
	PoolSize    int    `yaml:"pool_size"`
	MinIdle     int    `yaml:"min_idle"`
	MaxRetries  int    `yaml:"max_retries"`
	DialTimeout int    `yaml:"dial_timeout"`
}

// ProxyPoolConfig 代理池服务配置
type ProxyPoolConfig struct {
	Providers ProviderConfigs `yaml:"providers"`
}

// ProviderConfigs 代理提供商配置
type ProviderConfigs struct {
	Webshare WebshareConfig `yaml:"webshare"`
	// 在此可以添加其他提供商，例如 BrightData, Oxylabs等
}

// WebshareConfig Webshare提供商特定配置
type WebshareConfig struct {
	Enabled bool   `yaml:"enabled"`
	APIKey  string `yaml:"api_key"`
}

var globalConfig *Config

// validateConfigPath 验证配置文件路径，防止路径遍历攻击
func validateConfigPath(configPath string) error {
	// 清理路径
	cleanPath := filepath.Clean(configPath)

	// 检查是否包含路径遍历字符
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("配置文件路径不安全: %s", configPath)
	}

	// 获取绝对路径
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %v", err)
	}

	// 检查文件扩展名
	if !strings.HasSuffix(absPath, ".yaml") && !strings.HasSuffix(absPath, ".yml") {
		return fmt.Errorf("配置文件必须是YAML格式: %s", configPath)
	}

	return nil
}

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	// 验证配置文件路径安全性
	if err := validateConfigPath(configPath); err != nil {
		return nil, err
	}

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath) // #nosec G304 - path已经过安全验证
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	// 解析YAML
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}

	globalConfig = &config
	return &config, nil
}

// LoadFromDir 从目录加载配置文件
func LoadFromDir(configDir, serviceName string) (*Config, error) {
	configPath := filepath.Join(configDir, serviceName, "config.yaml")
	return Load(configPath)
}

// Get 获取全局配置
func Get() *Config {
	if globalConfig == nil {
		panic("配置未初始化，请先调用 Load 或 LoadFromDir")
	}
	return globalConfig
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证服务器配置
	if c.Server.Port <= 0 {
		return fmt.Errorf("服务器端口配置无效: %d", c.Server.Port)
	}

	// 验证数据库配置
	if c.Database.Host == "" {
		return fmt.Errorf("数据库主机地址不能为空")
	}
	if c.Database.Port <= 0 {
		return fmt.Errorf("数据库端口配置无效: %d", c.Database.Port)
	}
	if c.Database.User == "" {
		return fmt.Errorf("数据库用户名不能为空")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("数据库名不能为空")
	}

	// 验证Redis配置
	if c.Redis.Host == "" {
		return fmt.Errorf("Redis主机地址不能为空")
	}
	if c.Redis.Port <= 0 {
		return fmt.Errorf("redis端口配置无效: %d", c.Redis.Port)
	}

	// 验证日志配置
	if c.Log.Level == "" {
		c.Log.Level = "info" // 设置默认值
	}
	if c.Log.Format == "" {
		c.Log.Format = "json" // 设置默认值
	}

	return nil
}

// GetDSN 获取MySQL连接字符串
func (c *Config) GetDSN() string {
	charset := c.Database.Charset
	if charset == "" {
		charset = "utf8mb4"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.DBName,
		charset,
	)
}

// GetRedisAddr 获取Redis连接地址
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// GetServerAddr 获取服务器监听地址
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
