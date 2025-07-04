package validator

import "time"

// AnonymityLevel 匿名级别
type AnonymityLevel string

const (
	// Transparent 透明代理，目标服务器知道你的真实IP
	Transparent AnonymityLevel = "transparent"
	// Anonymous 匿名代理，目标服务器不知道你的真实IP，但知道你正在使用代理
	Anonymous AnonymityLevel = "anonymous"
	// Elite 高匿名代理，目标服务器不知道你的真实IP，也不知道你正在使用代理
	Elite AnonymityLevel = "elite"
	// Unknown 未知级别
	Unknown AnonymityLevel = "unknown"
)

// CheckTarget 定义了验证代理时所使用的目标。
type CheckTarget struct {
	// URL 是验证请求将发送到的地址。
	URL string
	// MustContain 是一个字符串，验证器会检查目标URL的响应体中是否包含此字符串。
	// 这有助于确保代理没有返回一个虚假的或错误页面。
	MustContain string
}

// ValidationResult 封装了一次代理验证的完整结果。
type ValidationResult struct {
	// IsAvailable 指示代理是否能够成功连接并获取响应。
	IsAvailable bool
	// Latency 是从发送请求到接收到响应头所花费的时间。
	Latency time.Duration
	// Anonymity 是代理的匿名级别。
	Anonymity AnonymityLevel
	// ErrorMessage 如果验证失败，这里会包含错误信息。
	ErrorMessage string
}
