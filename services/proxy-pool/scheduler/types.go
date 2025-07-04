package scheduler

import (
	"time"

	"github.com/XXXXD-cation/proxy-platform/services/proxy-pool/providers"
)

// ScheduleRequest 定义了选择代理时的请求参数。
type ScheduleRequest struct {
	// CountryCode 是期望的代理国家/地区代码 (ISO 3166-1 alpha-2)。
	// 如果为空，则不按国家筛选。
	CountryCode string
	// AllowInsecure 是否允许选择非高匿名(Elite)级别的代理。
	AllowInsecure bool
}

// UsageResult 封装了一次代理使用的结果，用于反馈给评分系统。
type UsageResult struct {
	// Proxy 是被使用的代理。
	Proxy *providers.ProxyIP
	// Success 指示代理是否成功完成了任务。
	Success bool
	// Latency 是完成任务所花费的时间。
	Latency time.Duration
	// ErrorMessage 如果使用失败，这里会包含错误信息。
	ErrorMessage string
}

// 调度器中使用的常量
const (
	// availableProxiesZSetKey 是存储可用代理的Redis有序集合的键。
	// Score: 代理的质量得分, Member: 代理的 "IP:Port" 字符串。
	availableProxiesZSetKey = "proxies:available"
)
