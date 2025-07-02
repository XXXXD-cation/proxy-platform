// Package utils 提供通用工具函数，包括字符串、数字、时间、加密、验证、JSON和切片操作
package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// 常量定义
const (
	// MathConstants 数学相关常量
	DecimalBase      = 10
	FloatBitSize     = 64
	PowerBase        = 10
	HoursPerDay      = 24
	BitsPerByte      = 8
	UUIDLength       = 16
	DefaultPrecision = 2

	// String constants 字符串相关常量
	DefaultTruncateSuffix = "..."

	// Crypto constants 加密相关常量
	DefaultRandomStringCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// StringUtils 字符串工具函数
type StringUtils struct{}

// IsEmpty 检查字符串是否为空
func (s StringUtils) IsEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}

// IsNotEmpty 检查字符串是否不为空
func (s StringUtils) IsNotEmpty(str string) bool {
	return !s.IsEmpty(str)
}

// Trim 去除字符串首尾空格
func (s StringUtils) Trim(str string) string {
	return strings.TrimSpace(str)
}

// Contains 检查字符串是否包含子串
func (s StringUtils) Contains(str, substr string) bool {
	return strings.Contains(str, substr)
}

// ContainsIgnoreCase 忽略大小写检查是否包含子串
func (s StringUtils) ContainsIgnoreCase(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

// StartsWith 检查字符串是否以指定前缀开始
func (s StringUtils) StartsWith(str, prefix string) bool {
	return strings.HasPrefix(str, prefix)
}

// EndsWith 检查字符串是否以指定后缀结束
func (s StringUtils) EndsWith(str, suffix string) bool {
	return strings.HasSuffix(str, suffix)
}

// Reverse 反转字符串
func (s StringUtils) Reverse(str string) string {
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Truncate 截断字符串到指定长度
func (s StringUtils) Truncate(str string, length int) string {
	if len(str) <= length {
		return str
	}
	return str[:length] + "..."
}

// PadLeft 左填充字符串
func (s StringUtils) PadLeft(str string, length int, pad string) string {
	if len(str) >= length {
		return str
	}
	padding := strings.Repeat(pad, length-len(str))
	return padding + str
}

// PadRight 右填充字符串
func (s StringUtils) PadRight(str string, length int, pad string) string {
	if len(str) >= length {
		return str
	}
	padding := strings.Repeat(pad, length-len(str))
	return str + padding
}

// CamelToSnake 驼峰转蛇形
func (s StringUtils) CamelToSnake(str string) string {
	var result strings.Builder
	for i, r := range str {
		if unicode.IsUpper(r) && i > 0 {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// SnakeToCamel 蛇形转驼峰
func (s StringUtils) SnakeToCamel(str string) string {
	parts := strings.Split(str, "_")
	var result strings.Builder
	for i, part := range parts {
		if i == 0 {
			result.WriteString(strings.ToLower(part))
		} else if part != "" {
			// 首字母大写，其余小写 (替代已弃用的strings.Title)
			result.WriteString(strings.ToUpper(string(part[0])) + strings.ToLower(part[1:]))
		}
	}
	return result.String()
}

// NumberUtils 数字工具函数
type NumberUtils struct{}

// IsNumber 检查字符串是否为数字
func (n NumberUtils) IsNumber(str string) bool {
	_, err := strconv.ParseFloat(str, 64)
	return err == nil
}

// IsInteger 检查字符串是否为整数
func (n NumberUtils) IsInteger(str string) bool {
	_, err := strconv.ParseInt(str, 10, 64)
	return err == nil
}

// ToInt 字符串转整数
func (n NumberUtils) ToInt(str string) (int, error) {
	return strconv.Atoi(str)
}

// ToInt64 字符串转长整数
func (n NumberUtils) ToInt64(str string) (int64, error) {
	return strconv.ParseInt(str, 10, 64)
}

// ToFloat64 字符串转浮点数
func (n NumberUtils) ToFloat64(str string) (float64, error) {
	return strconv.ParseFloat(str, 64)
}

// Round 浮点数四舍五入
func (n NumberUtils) Round(num float64, precision int) float64 {
	ratio := math.Pow(PowerBase, float64(precision))
	return math.Round(num*ratio) / ratio
}

// Max 返回最大值
func (n NumberUtils) Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min 返回最小值
func (n NumberUtils) Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Abs 返回绝对值
func (n NumberUtils) Abs(num int) int {
	if num < 0 {
		return -num
	}
	return num
}

// TimeUtils 时间工具函数
type TimeUtils struct{}

// Now 获取当前时间
func (t TimeUtils) Now() time.Time {
	return time.Now()
}

// NowUnix 获取当前Unix时间戳
func (t TimeUtils) NowUnix() int64 {
	return time.Now().Unix()
}

// NowUnixMilli 获取当前毫秒时间戳
func (t TimeUtils) NowUnixMilli() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// Format 格式化时间
func (t TimeUtils) Format(tm time.Time, layout string) string {
	return tm.Format(layout)
}

// Parse 解析时间字符串
func (t TimeUtils) Parse(layout, value string) (time.Time, error) {
	return time.Parse(layout, value)
}

// FormatDate 格式化日期 (YYYY-MM-DD)
func (t TimeUtils) FormatDate(tm time.Time) string {
	return tm.Format("2006-01-02")
}

// FormatDateTime 格式化日期时间 (YYYY-MM-DD HH:mm:ss)
func (t TimeUtils) FormatDateTime(tm time.Time) string {
	return tm.Format("2006-01-02 15:04:05")
}

// ParseDate 解析日期字符串
func (t TimeUtils) ParseDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}

// ParseDateTime 解析日期时间字符串
func (t TimeUtils) ParseDateTime(value string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", value)
}

// AddDays 增加天数
func (t TimeUtils) AddDays(tm time.Time, days int) time.Time {
	return tm.AddDate(0, 0, days)
}

// AddHours 增加小时数
func (t TimeUtils) AddHours(tm time.Time, hours int) time.Time {
	return tm.Add(time.Duration(hours) * time.Hour)
}

// AddMinutes 增加分钟数
func (t TimeUtils) AddMinutes(tm time.Time, minutes int) time.Time {
	return tm.Add(time.Duration(minutes) * time.Minute)
}

// DiffDays 计算两个时间之间的天数差
func (t TimeUtils) DiffDays(t1, t2 time.Time) int {
	return int(t1.Sub(t2).Hours() / HoursPerDay)
}

// CryptoUtils 加密工具函数
type CryptoUtils struct{}

// SHA256 计算SHA256哈希 (推荐使用，安全性更高)
func (c CryptoUtils) SHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GenerateRandomString 生成加密安全的随机字符串
func (c CryptoUtils) GenerateRandomString(length int) string {
	const charset = DefaultRandomStringCharset
	b := make([]byte, length)

	// 使用加密安全的随机数生成器
	if _, err := rand.Read(b); err != nil {
		// 如果crypto/rand失败，使用时间戳作为后备方案
		now := time.Now().UnixNano()
		for i := range b {
			b[i] = byte(now >> (i % BitsPerByte))
		}
	} else {
		// 将随机字节映射到字符集
		for i := range b {
			b[i] = charset[int(b[i])%len(charset)]
		}
	}
	return string(b)
}

// GenerateUUID 生成UUID (简化版)
func (c CryptoUtils) GenerateUUID() string {
	b := make([]byte, UUIDLength)
	if _, err := rand.Read(b); err != nil {
		// 如果crypto/rand失败，使用当前时间戳作为种子生成替代方案
		now := time.Now().UnixNano()
		for i := range b {
			b[i] = byte(now >> (i % BitsPerByte))
		}
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// ValidatorUtils 验证工具函数
type ValidatorUtils struct{}

// IsEmail 验证邮箱格式
func (v ValidatorUtils) IsEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(pattern, email)
	if err != nil {
		return false
	}
	return matched
}

// IsPhone 验证手机号格式 (中国大陆)
func (v ValidatorUtils) IsPhone(phone string) bool {
	pattern := `^1[3-9]\d{9}$`
	matched, err := regexp.MatchString(pattern, phone)
	if err != nil {
		return false
	}
	return matched
}

// IsURL 验证URL格式
func (v ValidatorUtils) IsURL(url string) bool {
	pattern := `^https?://[^\s/$.?#].[^\s]*$`
	matched, err := regexp.MatchString(pattern, url)
	if err != nil {
		return false
	}
	return matched
}

// IsIP 验证IP地址格式
func (v ValidatorUtils) IsIP(ip string) bool {
	pattern := `^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	matched, err := regexp.MatchString(pattern, ip)
	if err != nil {
		return false
	}
	return matched
}

// JSONUtils JSON工具函数
type JSONUtils struct{}

// ToJSON 转换为JSON字符串
func (j JSONUtils) ToJSON(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromJSON 从JSON字符串解析
func (j JSONUtils) FromJSON(jsonStr string, v interface{}) error {
	return json.Unmarshal([]byte(jsonStr), v)
}

// ToJSONPretty 转换为格式化的JSON字符串
func (j JSONUtils) ToJSONPretty(data interface{}) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// IsValidJSON 检查是否为有效的JSON字符串
func (j JSONUtils) IsValidJSON(jsonStr string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(jsonStr), &js) == nil
}

// SliceUtils 切片工具函数
type SliceUtils struct{}

// Contains 检查切片是否包含元素
func (s SliceUtils) Contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// ContainsInt 检查整数切片是否包含元素
func (s SliceUtils) ContainsInt(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// Remove 从切片中移除元素
func (s SliceUtils) Remove(slice []string, item string) []string {
	result := make([]string, 0)
	for _, v := range slice {
		if v != item {
			result = append(result, v)
		}
	}
	return result
}

// RemoveInt 从整数切片中移除元素
func (s SliceUtils) RemoveInt(slice []int, item int) []int {
	result := make([]int, 0)
	for _, v := range slice {
		if v != item {
			result = append(result, v)
		}
	}
	return result
}

// Unique 去除切片中的重复元素
func (s SliceUtils) Unique(slice []string) []string {
	keys := make(map[string]bool)
	result := make([]string, 0)
	for _, v := range slice {
		if !keys[v] {
			keys[v] = true
			result = append(result, v)
		}
	}
	return result
}

// UniqueInt 去除整数切片中的重复元素
func (s SliceUtils) UniqueInt(slice []int) []int {
	keys := make(map[int]bool)
	result := make([]int, 0)
	for _, v := range slice {
		if !keys[v] {
			keys[v] = true
			result = append(result, v)
		}
	}
	return result
}

// 全局工具实例
var (
	String    = StringUtils{}
	Number    = NumberUtils{}
	Time      = TimeUtils{}
	Crypto    = CryptoUtils{}
	Validator = ValidatorUtils{}
	JSON      = JSONUtils{}
	Slice     = SliceUtils{}
)

// GetEnv 获取环境变量，如果不存在则返回默认值
func GetEnv(key, defaultValue string) string {
	if value := strings.TrimSpace(key); value != "" {
		return value
	}
	return defaultValue
}

// Ternary 三元运算符
func Ternary(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

// TernaryString 字符串三元运算符
func TernaryString(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}

// TernaryInt 整数三元运算符
func TernaryInt(condition bool, trueVal, falseVal int) int {
	if condition {
		return trueVal
	}
	return falseVal
}

// Retry 重试函数
func Retry(maxAttempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < maxAttempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		if i < maxAttempts-1 {
			time.Sleep(delay)
		}
	}
	return err
}
