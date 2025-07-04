package scorer

import (
	"context"

	"github.com/XXXXD-cation/proxy-platform/services/proxy-pool/validator"
)

// Redis键和字段的常量
const (
	// proxyMetricsKeyPrefix 是存储代理指标的Redis哈希键的前缀。
	// 格式: proxy:metrics:{ip_address}
	proxyMetricsKeyPrefix = "proxy:metrics:"

	// --- 哈希字段 ---
	fieldSuccessCount     = "success_count"
	fieldFailureCount     = "failure_count"
	fieldTotalLatencyMs   = "total_latency_ms"
	fieldAnonymityLevel   = "anonymity_level"
	fieldLastSeenTime     = "last_seen_time"
	fieldLastSuccessTime  = "last_success_time"
	fieldConsecutiveFails = "consecutive_fails"
)

// 评分参数
const (
	// initialScore 是新发现代理的初始分数。
	initialScore = 0.5
	// maxLatencyMs 是计算延迟分数时可接受的最大延迟（毫秒）。
	// 超过此延迟的代理将获得非常低的延迟分数。
	maxLatencyMs = 5000.0 // 5 seconds
	// recentHours 表示我们更关注最近多少小时内的成功率。
	recentHours = 24
	// maxConsecutiveFails 是一个代理在被大幅降分或暂时禁用之前允许的最大连续失败次数。
	maxConsecutiveFails = 5
)

// 评分权重
const (
	weightSuccessRate = 0.50
	weightLatency     = 0.30
	weightAnonymity   = 0.15
	weightStability   = 0.05 // 基于近期成功和低连续失败
)

// AnonymityLevelScores holds the score values for different anonymity levels.
const (
	ScoreElite       = 1.0
	ScoreAnonymous   = 0.7
	ScoreTransparent = 0.3
	ScoreUnknown     = 0.1
)

// AnonymityScores 定义了不同匿名级别的基础分值。
var AnonymityScores = map[string]float64{
	"elite":       ScoreElite,
	"anonymous":   ScoreAnonymous,
	"transparent": ScoreTransparent,
	"unknown":     ScoreUnknown,
}

// Scorer 定义了质量评分器需要实现的接口。
type Scorer interface {
	UpdateMetricsWithCurrentTime(ctx context.Context, proxyIP string, result *validator.ValidationResult)
	CalculateScore(ctx context.Context, proxyIP string) float64
	RemoveProxyMetrics(ctx context.Context, proxyIP string) error
}
