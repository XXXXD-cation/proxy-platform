package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// TestConfig 测试环境配置
type TestConfig struct {
	MySQL TestMySQLConfig
	Redis TestRedisConfig
	JWT   TestJWTConfig
}

// TestMySQLConfig 测试MySQL配置
type TestMySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	DSN      string
}

// TestRedisConfig 测试Redis配置
type TestRedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// TestJWTConfig 测试JWT配置
type TestJWTConfig struct {
	SecretKey string
	Expiry    time.Duration
}

// NewTestConfig 创建测试配置
func NewTestConfig() *TestConfig {
	config := &TestConfig{
		MySQL: TestMySQLConfig{
			Host:     getEnv("TEST_MYSQL_HOST", "localhost"),
			Port:     getEnvInt("TEST_MYSQL_PORT", DefaultMySQLPort),
			User:     getEnv("TEST_MYSQL_USER", "proxy_user"),
			Password: getEnv("TEST_MYSQL_PASSWORD", "proxy_pass123"),
			Database: getEnv("TEST_MYSQL_DATABASE", "proxy_platform"),
		},
		Redis: TestRedisConfig{
			Host:     getEnv("TEST_REDIS_HOST", "localhost"),
			Port:     getEnvInt("TEST_REDIS_PORT", DefaultRedisPort),
			Password: getEnv("TEST_REDIS_PASSWORD", ""),
			DB:       getEnvInt("TEST_REDIS_DB", 1),
		},
		JWT: TestJWTConfig{
			SecretKey: getEnv("TEST_JWT_SECRET", "test-jwt-secret-key-for-integration-tests"),
			Expiry:    time.Hour * DefaultTokenCacheHours,
		},
	}

	// 构建MySQL DSN
	config.MySQL.DSN = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.MySQL.User,
		config.MySQL.Password,
		config.MySQL.Host,
		config.MySQL.Port,
		config.MySQL.Database,
	)

	return config
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取环境变量并转换为整数
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
