package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// 创建临时配置文件
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")
	
	configContent := `
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30
  write_timeout: 30
  idle_timeout: 120

database:
  host: "localhost"
  port: 3306
  user: "test_user"
  password: "test_pass"
  dbname: "test_db"
  charset: "utf8mb4"
  max_idle: 10
  max_open: 50
  max_life: 3600

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  pool_size: 10
  min_idle: 5
  max_retries: 3
  dial_timeout: 5

log:
  level: "info"
  format: "json"
  output: "stdout"
  filename: ""
  max_size: 100
  max_age: 7
  max_backups: 10
  compress: true
`
	
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("创建配置文件失败: %v", err)
	}
	
	// 测试加载配置
	config, err := Load(configFile)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	
	// 验证配置内容
	if config.Server.Port != 8080 {
		t.Errorf("期望服务器端口为 8080，实际为 %d", config.Server.Port)
	}
	
	if config.Database.Host != "localhost" {
		t.Errorf("期望数据库主机为 localhost，实际为 %s", config.Database.Host)
	}
	
	if config.Redis.Host != "localhost" {
		t.Errorf("期望Redis主机为 localhost，实际为 %s", config.Redis.Host)
	}
	
	if config.Log.Level != "info" {
		t.Errorf("期望日志级别为 info，实际为 %s", config.Log.Level)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	_, err := Load("non_existent_file.yaml")
	if err == nil {
		t.Error("期望加载不存在的文件失败")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid_config.yaml")
	
	invalidContent := `invalid yaml content: [[[`
	err := os.WriteFile(configFile, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("创建无效配置文件失败: %v", err)
	}
	
	_, err = Load(configFile)
	if err == nil {
		t.Error("期望加载无效YAML文件失败")
	}
}

func TestLoadFromDir(t *testing.T) {
	tempDir := t.TempDir()
	serviceDir := filepath.Join(tempDir, "test_service")
	
	err := os.MkdirAll(serviceDir, 0755)
	if err != nil {
		t.Fatalf("创建服务目录失败: %v", err)
	}
	
	configFile := filepath.Join(serviceDir, "config.yaml")
	configContent := `
server:
  host: "0.0.0.0"
  port: 9090
database:
  host: "localhost"
  port: 3306
  user: "test"
  password: "test"
  dbname: "test"
redis:
  host: "localhost" 
  port: 6379
log:
  level: "debug"
`
	
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("创建配置文件失败: %v", err)
	}
	
	config, err := LoadFromDir(tempDir, "test_service")
	if err != nil {
		t.Fatalf("从目录加载配置失败: %v", err)
	}
	
	if config.Server.Port != 9090 {
		t.Errorf("期望服务器端口为 9090，实际为 %d", config.Server.Port)
	}
}

func TestConfigValidate(t *testing.T) {
	// 测试有效配置
	validConfig := &Config{
		Server: ServerConfig{
			Port: 8080,
		},
		Database: DatabaseConfig{
			Host:   "localhost",
			Port:   3306,
			User:   "test",
			DBName: "test_db",
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
	}
	
	err := validConfig.Validate()
	if err != nil {
		t.Errorf("有效配置验证失败: %v", err)
	}
	
	// 测试无效服务器端口
	invalidServerConfig := *validConfig
	invalidServerConfig.Server.Port = 0
	err = invalidServerConfig.Validate()
	if err == nil {
		t.Error("期望无效服务器端口验证失败")
	}
	
	// 测试无效数据库配置
	invalidDBConfig := *validConfig
	invalidDBConfig.Database.Host = ""
	err = invalidDBConfig.Validate()
	if err == nil {
		t.Error("期望无效数据库主机验证失败")
	}
	
	// 测试无效Redis配置
	invalidRedisConfig := *validConfig
	invalidRedisConfig.Redis.Host = ""
	err = invalidRedisConfig.Validate()
	if err == nil {
		t.Error("期望无效Redis主机验证失败")
	}
}

func TestGetDSN(t *testing.T) {
	config := &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     3306,
			User:     "testuser",
			Password: "testpass",
			DBName:   "testdb",
			Charset:  "utf8mb4",
		},
	}
	
	expectedDSN := "testuser:testpass@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	actualDSN := config.GetDSN()
	
	if actualDSN != expectedDSN {
		t.Errorf("期望DSN为 %s，实际为 %s", expectedDSN, actualDSN)
	}
	
	// 测试默认字符集
	config.Database.Charset = ""
	expectedDSN = "testuser:testpass@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	actualDSN = config.GetDSN()
	
	if actualDSN != expectedDSN {
		t.Errorf("期望默认字符集DSN为 %s，实际为 %s", expectedDSN, actualDSN)
	}
}

func TestGetRedisAddr(t *testing.T) {
	config := &Config{
		Redis: RedisConfig{
			Host: "redis.example.com",
			Port: 6380,
		},
	}
	
	expectedAddr := "redis.example.com:6380"
	actualAddr := config.GetRedisAddr()
	
	if actualAddr != expectedAddr {
		t.Errorf("期望Redis地址为 %s，实际为 %s", expectedAddr, actualAddr)
	}
}

func TestGetServerAddr(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 9090,
		},
	}
	
	expectedAddr := "0.0.0.0:9090"
	actualAddr := config.GetServerAddr()
	
	if actualAddr != expectedAddr {
		t.Errorf("期望服务器地址为 %s，实际为 %s", expectedAddr, actualAddr)
	}
}

func TestGlobalConfig(t *testing.T) {
	// 重置全局配置
	globalConfig = nil
	
	// 测试未初始化时panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("期望Get()在未初始化时panic")
		}
	}()
	Get()
} 