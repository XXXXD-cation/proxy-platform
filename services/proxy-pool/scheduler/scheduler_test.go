package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"

	"github.com/XXXXD-cation/proxy-platform/services/proxy-pool/providers"
	"github.com/XXXXD-cation/proxy-platform/services/proxy-pool/scorer"
	"github.com/XXXXD-cation/proxy-platform/services/proxy-pool/validator"
)

// mockScorer 是 scorer.Scorer 的一个模拟实现，用于测试。
type mockScorer struct {
	scores map[string]float64
}

func newMockScorer() *mockScorer {
	return &mockScorer{
		scores: make(map[string]float64),
	}
}

func (m *mockScorer) UpdateMetricsWithCurrentTime(_ context.Context, proxyIP string, result *validator.ValidationResult) {
	// 简单的模拟逻辑：成功则分数增加，失败则减少
	if result.IsAvailable {
		m.scores[proxyIP] += 0.1
	} else {
		m.scores[proxyIP] -= 0.2
	}
	if m.scores[proxyIP] > 1.0 {
		m.scores[proxyIP] = 1.0
	}
	if m.scores[proxyIP] < 0.0 {
		m.scores[proxyIP] = 0.0
	}
}

func (m *mockScorer) CalculateScore(_ context.Context, proxyIP string) float64 {
	score, exists := m.scores[proxyIP]
	if !exists {
		return 0.5 // 默认分数
	}
	return score
}

func (m *mockScorer) RemoveProxyMetrics(_ context.Context, proxyIP string) error {
	delete(m.scores, proxyIP)
	return nil
}

// --- Test Cases ---

func TestIntelligentScheduler_SelectBestProxy(t *testing.T) {
	db, mock := redismock.NewClientMock()
	scorer := newMockScorer()
	scheduler := NewIntelligentScheduler(db, scorer)
	ctx := context.Background()

	t.Run("No proxies available", func(t *testing.T) {
		mock.ExpectZRevRange(availableProxiesZSetKey, 0, 0).SetVal([]string{})
		_, err := scheduler.SelectBestProxy(ctx, &ScheduleRequest{})
		assert.ErrorIs(t, err, redis.Nil)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Successfully select best proxy", func(t *testing.T) {
		bestProxy := "1.1.1.1:8080"
		mock.ExpectZRevRange(availableProxiesZSetKey, 0, 0).SetVal([]string{bestProxy})

		// TODO: 一旦实现了元数据获取，这里需要模拟对哈希的调用

		p, err := scheduler.SelectBestProxy(ctx, &ScheduleRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, "1.1.1.1", p.Address)
		assert.Equal(t, uint16(8080), p.Port)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestIntelligentScheduler_UpdateProxyUsage(t *testing.T) {
	db, mock := redismock.NewClientMock()
	scorer := newMockScorer()
	scheduler := NewIntelligentScheduler(db, scorer)
	ctx := context.Background()

	proxy := &providers.ProxyIP{Address: "2.2.2.2", Port: 8888}
	proxyIdentifier := formatProxyIdentifier(proxy.Address, proxy.Port)

	// 初始分数为0.5
	scorer.scores[proxy.Address] = 0.5

	// 使用成功，分数应增加
	usageResult := &UsageResult{
		Proxy:   proxy,
		Success: true,
		Latency: 100 * time.Millisecond,
	}

	// 模拟 redis ZAdd 操作
	newScore := 0.6 // 0.5 + 0.1
	mock.ExpectZAdd(availableProxiesZSetKey, &redis.Z{Score: newScore, Member: proxyIdentifier}).SetVal(1)

	scheduler.UpdateProxyUsage(ctx, usageResult)

	assert.Equal(t, newScore, scorer.scores[proxy.Address])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIntelligentScheduler_AddAndRemoveProxy(t *testing.T) {
	db, mock := redismock.NewClientMock()
	scorer := newMockScorer()
	scheduler := NewIntelligentScheduler(db, scorer)
	ctx := context.Background()

	proxy := &providers.ProxyIP{Address: "3.3.3.3", Port: 9999}
	proxyIdentifier := formatProxyIdentifier(proxy.Address, proxy.Port)
	initialScore := scorer.CalculateScore(ctx, proxy.Address)

	// Test Add
	mock.ExpectZAdd(availableProxiesZSetKey, &redis.Z{Score: initialScore, Member: proxyIdentifier}).SetVal(1)
	scheduler.AddProxyToPool(ctx, proxy)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test Remove
	mock.ExpectZRem(availableProxiesZSetKey, proxyIdentifier).SetVal(1)
	scheduler.RemoveProxyFromPool(ctx, proxy)
	_, scoreExists := scorer.scores[proxy.Address]
	assert.False(t, scoreExists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_parseProxyIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		wantIP     string
		wantPort   uint16
		wantErr    bool
	}{
		{"Valid IPv4", "192.168.1.1:8080", "192.168.1.1", 8080, false},
		{"Valid IPv6", "[::1]:8888", "[::1]", 8888, false},
		{"No port", "192.168.1.1", "", 0, true},
		{"Invalid port", "192.168.1.1:abc", "", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIP, gotPort, err := parseProxyIdentifier(tt.identifier)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantIP, gotIP)
				assert.Equal(t, tt.wantPort, gotPort)
			}
		})
	}
}

// 这确保了 mockScorer 实现了 scorer.Scorer 接口
var _ scorer.Scorer = (*mockScorer)(nil)
