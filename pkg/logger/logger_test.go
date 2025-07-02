package logger

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNew(t *testing.T) {
	config := LogConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("创建日志器失败: %v", err)
	}

	if logger == nil {
		t.Error("日志器不能为nil")
	}

	if logger.GetLevel() != logrus.DebugLevel {
		t.Error("日志级别设置错误")
	}
}

func TestNewWithInvalidLevel(t *testing.T) {
	config := LogConfig{
		Level:  "invalid",
		Format: "json",
		Output: "stdout",
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("创建日志器失败: %v", err)
	}

	// 应该使用默认的Info级别
	if logger.GetLevel() != logrus.InfoLevel {
		t.Error("无效级别时应该使用默认Info级别")
	}
}

func TestNewWithFileOutput(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := LogConfig{
		Level:    "info",
		Format:   "text",
		Output:   "file",
		Filename: logFile,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("创建文件日志器失败: %v", err)
	}

	if logger == nil {
		t.Error("日志器不能为nil")
	}

	// 写入一条日志
	logger.Info("测试消息")

	// 检查文件是否存在
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("日志文件未创建")
	}
}

func TestInit(t *testing.T) {
	// 重置全局日志器
	globalLogger = nil

	config := LogConfig{
		Level:  "warn",
		Format: "json",
		Output: "stdout",
	}

	err := Init(config)
	if err != nil {
		t.Fatalf("初始化日志器失败: %v", err)
	}

	if globalLogger == nil {
		t.Error("全局日志器未初始化")
	}

	if globalLogger.GetLevel() != logrus.WarnLevel {
		t.Error("全局日志器级别设置错误")
	}
}

func TestGet(t *testing.T) {
	// 重置全局日志器
	globalLogger = nil

	// 第一次调用Get应该创建默认日志器
	logger := Get()
	if logger == nil {
		t.Error("Get()返回nil")
	}

	if logger.GetLevel() != logrus.InfoLevel {
		t.Error("默认日志器级别应该是Info")
	}
}

func TestWithContext(t *testing.T) {
	logger := Get()
	ctx := context.WithValue(context.Background(), "request_id", "test-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")
	ctx = context.WithValue(ctx, "trace_id", "trace-789")

	entry := logger.WithContext(ctx)
	if entry == nil {
		t.Error("WithContext返回nil")
	}

	// 验证上下文字段是否被添加
	if entry.Data["request_id"] != "test-123" {
		t.Error("request_id未正确添加到日志条目")
	}
	if entry.Data["user_id"] != "user-456" {
		t.Error("user_id未正确添加到日志条目")
	}
	if entry.Data["trace_id"] != "trace-789" {
		t.Error("trace_id未正确添加到日志条目")
	}
}

func TestWithFields(t *testing.T) {
	logger := Get()
	fields := logrus.Fields{
		"component": "test",
		"version":   "1.0.0",
	}

	entry := logger.WithFields(fields)
	if entry == nil {
		t.Error("WithFields返回nil")
	}

	if entry.Data["component"] != "test" {
		t.Error("component字段未正确添加")
	}
	if entry.Data["version"] != "1.0.0" {
		t.Error("version字段未正确添加")
	}
}

func TestWithField(t *testing.T) {
	logger := Get()
	entry := logger.WithField("test_key", "test_value")

	if entry == nil {
		t.Error("WithField返回nil")
	}

	if entry.Data["test_key"] != "test_value" {
		t.Error("字段未正确添加")
	}
}

func TestWithError(t *testing.T) {
	logger := Get()
	testErr := &testError{msg: "test error"}
	entry := logger.WithError(testErr)

	if entry == nil {
		t.Error("WithError返回nil")
	}

	if entry.Data["error"] != testErr {
		t.Error("错误未正确添加")
	}
}

func TestGlobalLogMethods(t *testing.T) {
	// 创建一个缓冲区来捕获日志输出
	var buf bytes.Buffer
	
	// 重置并配置全局日志器
	config := LogConfig{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	}
	
	logger, _ := New(config)
	logger.SetOutput(&buf)
	globalLogger = logger

	// 测试各种日志级别
	Debug("debug message")
	if !strings.Contains(buf.String(), "debug message") {
		t.Error("Debug消息未输出")
	}

	buf.Reset()
	Info("info message")
	if !strings.Contains(buf.String(), "info message") {
		t.Error("Info消息未输出")
	}

	buf.Reset()
	Warn("warn message")
	if !strings.Contains(buf.String(), "warn message") {
		t.Error("Warn消息未输出")
	}

	buf.Reset()
	Error("error message")
	if !strings.Contains(buf.String(), "error message") {
		t.Error("Error消息未输出")
	}

	buf.Reset()
	Debugf("debug formatted %s", "message")
	if !strings.Contains(buf.String(), "debug formatted message") {
		t.Error("Debugf消息未输出")
	}

	buf.Reset()
	Infof("info formatted %s", "message")
	if !strings.Contains(buf.String(), "info formatted message") {
		t.Error("Infof消息未输出")
	}

	buf.Reset()
	Warnf("warn formatted %s", "message")
	if !strings.Contains(buf.String(), "warn formatted message") {
		t.Error("Warnf消息未输出")
	}

	buf.Reset()
	Errorf("error formatted %s", "message")
	if !strings.Contains(buf.String(), "error formatted message") {
		t.Error("Errorf消息未输出")
	}
}

func TestGlobalWithMethods(t *testing.T) {
	entry := WithField("global_key", "global_value")
	if entry == nil {
		t.Error("全局WithField返回nil")
	}

	fields := logrus.Fields{"global_field": "value"}
	entry = WithFields(fields)
	if entry == nil {
		t.Error("全局WithFields返回nil")
	}

	testErr := &testError{msg: "global error"}
	entry = WithError(testErr)
	if entry == nil {
		t.Error("全局WithError返回nil")
	}

	ctx := context.WithValue(context.Background(), "global_id", "123")
	entry = WithContext(ctx)
	if entry == nil {
		t.Error("全局WithContext返回nil")
	}
}

func TestSetGlobalLevel(t *testing.T) {
	err := SetGlobalLevel("error")
	if err != nil {
		t.Errorf("设置全局日志级别失败: %v", err)
	}

	if Get().GetLevel() != logrus.ErrorLevel {
		t.Error("全局日志级别未正确设置为Error")
	}

	// 测试无效级别
	err = SetGlobalLevel("invalid_level")
	if err == nil {
		t.Error("设置无效日志级别应该返回错误")
	}
}

func TestGetGlobalLevel(t *testing.T) {
	SetGlobalLevel("debug")
	level := GetGlobalLevel()
	if level != "debug" {
		t.Errorf("期望日志级别为debug，实际为%s", level)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "logger_config.yaml")

	configContent := `
level: "debug"
format: "json"
output: "file"
filename: "test.log"
max_size: 100
max_age: 7
max_backups: 10
compress: true
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("创建配置文件失败: %v", err)
	}

	config, err := LoadConfigFromFile(configFile)
	if err != nil {
		t.Fatalf("加载配置文件失败: %v", err)
	}

	if config.Level != "debug" {
		t.Errorf("期望日志级别为debug，实际为%s", config.Level)
	}
	if config.Format != "json" {
		t.Errorf("期望日志格式为json，实际为%s", config.Format)
	}
	if config.Output != "file" {
		t.Errorf("期望输出为file，实际为%s", config.Output)
	}
}

func TestLoadConfigFromNonExistentFile(t *testing.T) {
	_, err := LoadConfigFromFile("non_existent_config.yaml")
	if err == nil {
		t.Error("期望加载不存在的配置文件失败")
	}
}

// 辅助测试错误类型
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
} 