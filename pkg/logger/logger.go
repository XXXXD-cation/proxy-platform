package logger

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Logger 日志管理器
type Logger struct {
	*logrus.Logger
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"max_size"`
	MaxAge     int    `yaml:"max_age"`
	MaxBackups int    `yaml:"max_backups"`
	Compress   bool   `yaml:"compress"`
}

var globalLogger *Logger

// New 创建新的日志实例
func New(config LogConfig) (*Logger, error) {
	logger := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// 设置日志格式
	switch config.Format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	default:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	}

	// 设置输出目标
	var output io.Writer
	switch config.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	case "file":
		if config.Filename == "" {
			config.Filename = "app.log"
		}
		
		// 确保日志目录存在
		dir := filepath.Dir(config.Filename)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
		
		file, err := os.OpenFile(config.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		output = file
	default:
		output = os.Stdout
	}
	logger.SetOutput(output)

	return &Logger{Logger: logger}, nil
}

// Init 初始化全局日志器
func Init(config LogConfig) error {
	logger, err := New(config)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// Get 获取全局日志器
func Get() *Logger {
	if globalLogger == nil {
		// 使用默认配置创建日志器
		defaultConfig := LogConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		}
		logger, _ := New(defaultConfig)
		globalLogger = logger
	}
	return globalLogger
}

// WithContext 添加上下文信息
func (l *Logger) WithContext(ctx context.Context) *logrus.Entry {
	entry := l.Logger.WithContext(ctx)
	
	// 从上下文中提取常用字段
	if requestID := ctx.Value("request_id"); requestID != nil {
		entry = entry.WithField("request_id", requestID)
	}
	
	if userID := ctx.Value("user_id"); userID != nil {
		entry = entry.WithField("user_id", userID)
	}
	
	if traceID := ctx.Value("trace_id"); traceID != nil {
		entry = entry.WithField("trace_id", traceID)
	}
	
	return entry
}

// WithFields 添加字段信息
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.Logger.WithFields(fields)
}

// WithField 添加单个字段
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

// WithError 添加错误信息
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

// 全局便捷方法
func Debug(args ...interface{}) {
	Get().Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	Get().Debugf(format, args...)
}

func Info(args ...interface{}) {
	Get().Info(args...)
}

func Infof(format string, args ...interface{}) {
	Get().Infof(format, args...)
}

func Warn(args ...interface{}) {
	Get().Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	Get().Warnf(format, args...)
}

func Error(args ...interface{}) {
	Get().Error(args...)
}

func Errorf(format string, args ...interface{}) {
	Get().Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	Get().Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	Get().Fatalf(format, args...)
}

func Panic(args ...interface{}) {
	Get().Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	Get().Panicf(format, args...)
}

// 带上下文的全局便捷方法
func WithContext(ctx context.Context) *logrus.Entry {
	return Get().WithContext(ctx)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return Get().WithFields(fields)
}

func WithField(key string, value interface{}) *logrus.Entry {
	return Get().WithField(key, value)
}

func WithError(err error) *logrus.Entry {
	return Get().WithError(err)
}

// LoadConfigFromFile 从文件加载日志配置
func LoadConfigFromFile(configPath string) (LogConfig, error) {
	var config LogConfig
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, err
	}
	
	err = yaml.Unmarshal(data, &config)
	return config, err
}

// SetGlobalLevel 设置全局日志级别
func SetGlobalLevel(level string) error {
	l, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	Get().SetLevel(l)
	return nil
}

// GetGlobalLevel 获取当前全局日志级别
func GetGlobalLevel() string {
	return Get().GetLevel().String()
} 