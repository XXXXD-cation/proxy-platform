package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"

	"github.com/XXXXD-cation/proxy-platform/pkg/logger"
	"github.com/XXXXD-cation/proxy-platform/services/proxy-pool/providers"
	"github.com/XXXXD-cation/proxy-platform/services/proxy-pool/scorer"
	"github.com/XXXXD-cation/proxy-platform/services/proxy-pool/validator"
)

// IntelligentScheduler 实现了智能调度逻辑。
type IntelligentScheduler struct {
	redis  *redis.Client
	scorer scorer.Scorer
}

// NewIntelligentScheduler 创建一个新的 IntelligentScheduler 实例。
func NewIntelligentScheduler(redisClient *redis.Client, scorer scorer.Scorer) *IntelligentScheduler {
	return &IntelligentScheduler{
		redis:  redisClient,
		scorer: scorer,
	}
}

// SelectBestProxy 根据请求选择一个得分最高的可用代理。
// 目前的实现比较简单，直接从ZSET中返回分数最高的成员。
// 复杂的逻辑 (如按国家筛选) 将在后续版本中添加。
func (is *IntelligentScheduler) SelectBestProxy(ctx context.Context, _ *ScheduleRequest) (*providers.ProxyIP, error) {
	// 从有序集合中按分数从高到低取出一个代理
	result, err := is.redis.ZRevRange(ctx, availableProxiesZSetKey, 0, 0).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get best proxy from redis: %w", err)
	}

	if len(result) == 0 {
		return nil, redis.Nil // 使用 redis.Nil 表示没有可用的代理
	}

	proxyIdentifier := result[0]
	// TODO: 从 Redis 哈希中获取完整的代理元数据，而不仅仅是IP和端口。
	// 目前，我们只返回一个包含部分信息的ProxyIP结构体。
	ip, portStr, err := parseProxyIdentifier(proxyIdentifier)
	if err != nil {
		logger.WithError(err).Error("failed to parse proxy identifier from zset")
		// 尝试下一个
		return nil, err
	}

	return &providers.ProxyIP{
		Address: ip,
		Port:    portStr,
	}, nil
}

// UpdateProxyUsage 在代理被网关使用后，根据使用结果更新其质量评分。
func (is *IntelligentScheduler) UpdateProxyUsage(ctx context.Context, result *UsageResult) {
	// 将 UsageResult 转换为 scorer 可以理解的 ValidationResult
	validationRes := &validator.ValidationResult{
		IsAvailable:  result.Success,
		Latency:      result.Latency,
		ErrorMessage: result.ErrorMessage,
		// 匿名字段在实际使用中无法检测，因此留空
		Anonymity: validator.Unknown,
	}

	// 调用scorer来更新指标
	is.scorer.UpdateMetricsWithCurrentTime(ctx, result.Proxy.Address, validationRes)

	// 获取更新后的分数
	newScore := is.scorer.CalculateScore(ctx, result.Proxy.Address)

	// 更新代理在有序集合中的分数
	proxyIdentifier := formatProxyIdentifier(result.Proxy.Address, result.Proxy.Port)
	_, err := is.redis.ZAdd(ctx, availableProxiesZSetKey, &redis.Z{
		Score:  newScore,
		Member: proxyIdentifier,
	}).Result()
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"proxy": proxyIdentifier,
			"score": newScore,
			"error": err,
		}).Error("failed to update proxy score in zset")
	}
}

// AddProxyToPool 将一个新的代理添加到调度池中。
// 这个方法会被代理获取服务调用。
func (is *IntelligentScheduler) AddProxyToPool(ctx context.Context, proxy *providers.ProxyIP) {
	initialScore := is.scorer.CalculateScore(ctx, proxy.Address)
	proxyIdentifier := formatProxyIdentifier(proxy.Address, proxy.Port)

	// 将代理添加到可用集合
	_, err := is.redis.ZAdd(ctx, availableProxiesZSetKey, &redis.Z{
		Score:  initialScore,
		Member: proxyIdentifier,
	}).Result()

	if err != nil {
		logger.WithError(err).Error("failed to add proxy to available set")
	}

	// TODO: 同时存储代理的完整元数据到Redis哈希中。
}

// RemoveProxyFromPool 从调度池中移除一个代理。
func (is *IntelligentScheduler) RemoveProxyFromPool(ctx context.Context, proxy *providers.ProxyIP) {
	proxyIdentifier := formatProxyIdentifier(proxy.Address, proxy.Port)
	_, err := is.redis.ZRem(ctx, availableProxiesZSetKey, proxyIdentifier).Result()
	if err != nil {
		logger.WithError(err).Error("failed to remove proxy from available set")
	}
	// 同时删除元数据
	if err := is.scorer.RemoveProxyMetrics(ctx, proxy.Address); err != nil {
		logger.WithError(err).Error("failed to remove proxy metrics")
	}
}

// formatProxyIdentifier 创建一个 "ip:port" 格式的字符串。
func formatProxyIdentifier(ip string, port uint16) string {
	return fmt.Sprintf("%s:%d", ip, port)
}

// parseProxyIdentifier 从 "ip:port" 格式的字符串中解析出ip和port。
func parseProxyIdentifier(identifier string) (ip string, port uint16, err error) {
	// 在最后一个冒号处分割
	lastColon := strings.LastIndex(identifier, ":")
	if lastColon == -1 {
		return "", 0, fmt.Errorf("invalid identifier format: %s", identifier)
	}
	ip = identifier[:lastColon]
	portStr := identifier[lastColon+1:]
	portVal, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return "", 0, fmt.Errorf("invalid port in identifier: %s", portStr)
	}
	return ip, uint16(portVal), nil
}
