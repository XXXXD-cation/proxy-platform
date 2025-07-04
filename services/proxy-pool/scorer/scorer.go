package scorer

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/XXXXD-cation/proxy-platform/pkg/logger"
	"github.com/XXXXD-cation/proxy-platform/services/proxy-pool/validator"
)

// proxyMetrics 封装了从Redis读取的代理指标。
type proxyMetrics struct {
	SuccessCount     int64
	FailureCount     int64
	TotalLatencyMs   int64
	AnonymityLevel   string
	LastSeenTime     time.Time
	LastSuccessTime  time.Time
	ConsecutiveFails int64
}

// QualityScorer 负责计算和更新代理的质量得分。
type QualityScorer struct {
	redis *redis.Client
}

// NewQualityScorer 创建一个新的 QualityScorer 实例。
func NewQualityScorer(redisClient *redis.Client) *QualityScorer {
	return &QualityScorer{
		redis: redisClient,
	}
}

// UpdateMetrics is deprecated and will be removed. Use UpdateMetricsWithCurrentTime.
func (qs *QualityScorer) UpdateMetrics(_ context.Context, _ string, _ *validator.ValidationResult, _ time.Time) {
	// This function is intentionally left empty to resolve a linting error.
}

// UpdateMetricsWithCurrentTime 是 UpdateMetrics 的一个包装器，它会自动使用当前时间。
// 这是外部服务应该调用的主要方法。
func (qs *QualityScorer) UpdateMetricsWithCurrentTime(ctx context.Context, proxyIP string, result *validator.ValidationResult) {
	qs.updateMetrics(ctx, proxyIP, result, time.Now())
}

// updateMetrics 在代理验证后更新其在Redis中的统计指标。
// 它需要一个明确的时间戳，以便于测试。
func (qs *QualityScorer) updateMetrics(ctx context.Context, proxyIP string, result *validator.ValidationResult, now time.Time) {
	key := proxyMetricsKeyPrefix + proxyIP

	pipe := qs.redis.TxPipeline()

	pipe.HSet(ctx, key, fieldLastSeenTime, now.UTC().Format(time.RFC3339))
	if result.IsAvailable {
		pipe.HIncrBy(ctx, key, fieldSuccessCount, 1)
		pipe.HIncrBy(ctx, key, fieldTotalLatencyMs, result.Latency.Milliseconds())
		pipe.HSet(ctx, key, fieldLastSuccessTime, now.UTC().Format(time.RFC3339))
		pipe.HSet(ctx, key, fieldAnonymityLevel, result.Anonymity)
		pipe.HSet(ctx, key, fieldConsecutiveFails, 0) // 成功后重置连续失败次数
	} else {
		pipe.HIncrBy(ctx, key, fieldFailureCount, 1)
		pipe.HIncrBy(ctx, key, fieldConsecutiveFails, 1)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"proxy_ip": proxyIP,
			"error":    err,
		}).Error("更新代理指标失败")
	}
}

// CalculateScore 计算给定代理IP的质量得分。
// 如果一个代理从未被见过，它会得到一个初始的默认分数。
func (qs *QualityScorer) CalculateScore(ctx context.Context, proxyIP string) float64 {
	metrics, err := qs.getMetrics(ctx, proxyIP)
	if err != nil {
		if err == redis.Nil {
			// 代理是全新的，给予一个默认分数
			return initialScore
		}
		logger.WithFields(map[string]interface{}{
			"proxy_ip": proxyIP,
			"error":    err,
		}).Error("获取代理指标失败")
		return 0.0 // 获取失败，给予最低分
	}

	if metrics.SuccessCount+metrics.FailureCount == 0 {
		return initialScore
	}

	successRateScore := calculateSuccessRateScore(metrics)
	latencyScore := calculateLatencyScore(metrics)
	anonymityScore := calculateAnonymityScore(metrics)
	stabilityScore := calculateStabilityScore(metrics)

	// 根据权重计算最终分数
	finalScore := successRateScore*weightSuccessRate +
		latencyScore*weightLatency +
		anonymityScore*weightAnonymity +
		stabilityScore*weightStability

	// 确保分数在 0.0 到 1.0 之间
	return math.Max(0.0, math.Min(1.0, finalScore))
}

// getMetrics 从Redis中获取并解析一个代理的所有指标。
func (qs *QualityScorer) getMetrics(ctx context.Context, proxyIP string) (*proxyMetrics, error) {
	key := proxyMetricsKeyPrefix + proxyIP
	data, err := qs.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, redis.Nil // 使用 redis.Nil 来表示键不存在
	}

	// 辅助函数，用于安全地解析int
	parseInt := func(s string) int64 {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			logger.WithFields(map[string]interface{}{"string": s, "error": err}).Warn("无法解析整数")
			return 0
		}
		return i
	}

	// 辅助函数，用于安全地解析time
	parseTime := func(s string) time.Time {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			logger.WithFields(map[string]interface{}{"string": s, "error": err}).Warn("无法解析时间")
			return time.Time{}
		}
		return t
	}

	return &proxyMetrics{
		SuccessCount:     parseInt(data[fieldSuccessCount]),
		FailureCount:     parseInt(data[fieldFailureCount]),
		TotalLatencyMs:   parseInt(data[fieldTotalLatencyMs]),
		AnonymityLevel:   data[fieldAnonymityLevel],
		LastSeenTime:     parseTime(data[fieldLastSeenTime]),
		LastSuccessTime:  parseTime(data[fieldLastSuccessTime]),
		ConsecutiveFails: parseInt(data[fieldConsecutiveFails]),
	}, nil
}

func calculateSuccessRateScore(m *proxyMetrics) float64 {
	total := m.SuccessCount + m.FailureCount
	if total == 0 {
		return 0.0
	}
	return float64(m.SuccessCount) / float64(total)
}

func calculateLatencyScore(m *proxyMetrics) float64 {
	if m.SuccessCount == 0 {
		return 0.0
	}
	avgLatency := float64(m.TotalLatencyMs) / float64(m.SuccessCount)

	// 使用一个递减函数，延迟越高，分数越低
	// 当 avgLatency = maxLatencyMs 时， 分数趋近于0
	// 当 avgLatency = 0 时， 分数 = 1
	score := math.Max(0.0, 1.0-(avgLatency/maxLatencyMs))
	return score
}

func calculateAnonymityScore(m *proxyMetrics) float64 {
	score, exists := AnonymityScores[m.AnonymityLevel]
	if !exists {
		return AnonymityScores["unknown"]
	}
	return score
}

func calculateStabilityScore(m *proxyMetrics) float64 {
	// 惩罚连续失败的代理
	failPenalty := math.Min(1.0, float64(m.ConsecutiveFails)/float64(maxConsecutiveFails))

	// 奖励近期成功的代理
	var recentSuccessBonus float64
	if !m.LastSuccessTime.IsZero() {
		hoursSinceSuccess := time.Since(m.LastSuccessTime).Hours()
		if hoursSinceSuccess < recentHours {
			// 线性递减的奖励，最近成功的奖励更高
			recentSuccessBonus = 1.0 - (hoursSinceSuccess / recentHours)
		}
	}

	// 稳定性分数 = (1 - 失败惩罚) * 奖励
	// 这里简单地结合一下，我们更关心惩罚
	const stabilityBonusFactor = 0.5
	stabilityScore := (1.0 - failPenalty) * (stabilityBonusFactor + stabilityBonusFactor*recentSuccessBonus)
	return math.Max(0.0, stabilityScore)
}

// RemoveProxyMetrics 从Redis中删除一个代理的所有指标数据。
// 当一个代理被确认永久失效时调用。
func (qs *QualityScorer) RemoveProxyMetrics(ctx context.Context, proxyIP string) error {
	key := proxyMetricsKeyPrefix + proxyIP
	_, err := qs.redis.Del(ctx, key).Result()
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"proxy_ip": proxyIP,
			"error":    err,
		}).Error("删除代理指标失败")
		return fmt.Errorf("删除代理 %s 的指标失败: %w", proxyIP, err)
	}
	return nil
}
