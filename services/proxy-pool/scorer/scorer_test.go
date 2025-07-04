package scorer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"

	"github.com/XXXXD-cation/proxy-platform/services/proxy-pool/validator"
)

const testProxyIP = "1.2.3.4"

var fixedTime = time.Date(2025, 7, 4, 10, 0, 0, 0, time.UTC)
var fixedTimeStr = fixedTime.Format(time.RFC3339)

func TestQualityScorer_UpdateMetrics_Success(t *testing.T) {
	db, mock := redismock.NewClientMock()
	scorer := NewQualityScorer(db)
	ctx := context.Background()
	proxyIP := testProxyIP
	key := proxyMetricsKeyPrefix + proxyIP

	validationResult := &validator.ValidationResult{
		IsAvailable: true,
		Latency:     100 * time.Millisecond,
		Anonymity:   "elite",
	}

	mock.Regexp().ExpectTxPipeline()
	mock.ExpectHSet(key, fieldLastSeenTime, fixedTimeStr).SetVal(1)
	mock.ExpectHIncrBy(key, fieldSuccessCount, 1).SetVal(1)
	mock.ExpectHIncrBy(key, fieldTotalLatencyMs, 100).SetVal(100)
	mock.ExpectHSet(key, fieldLastSuccessTime, fixedTimeStr).SetVal(1)
	mock.ExpectHSet(key, fieldAnonymityLevel, validator.AnonymityLevel("elite")).SetVal(1)
	mock.ExpectHSet(key, fieldConsecutiveFails, 0).SetVal(1)
	mock.Regexp().ExpectTxPipelineExec()

	scorer.updateMetrics(ctx, proxyIP, validationResult, fixedTime)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQualityScorer_UpdateMetrics_Failure(t *testing.T) {
	db, mock := redismock.NewClientMock()
	scorer := NewQualityScorer(db)
	ctx := context.Background()
	proxyIP := testProxyIP
	key := proxyMetricsKeyPrefix + proxyIP

	validationResult := &validator.ValidationResult{
		IsAvailable: false,
	}

	mock.Regexp().ExpectTxPipeline()
	mock.ExpectHSet(key, fieldLastSeenTime, fixedTimeStr).SetVal(1)
	mock.ExpectHIncrBy(key, fieldFailureCount, 1).SetVal(1)
	mock.ExpectHIncrBy(key, fieldConsecutiveFails, 1).SetVal(1)
	mock.Regexp().ExpectTxPipelineExec()

	scorer.updateMetrics(ctx, proxyIP, validationResult, fixedTime)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQualityScorer_CalculateScore(t *testing.T) {
	db, mock := redismock.NewClientMock()
	scorer := NewQualityScorer(db)
	ctx := context.Background()
	proxyIP := testProxyIP
	key := proxyMetricsKeyPrefix + proxyIP

	t.Run("New Proxy", func(t *testing.T) {
		mock.ExpectHGetAll(key).RedisNil()
		score := scorer.CalculateScore(ctx, proxyIP)
		assert.Equal(t, initialScore, score)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Good Proxy", func(t *testing.T) {
		lastSuccessTime := time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)
		mock.ExpectHGetAll(key).SetVal(map[string]string{
			fieldSuccessCount:     "100",
			fieldFailureCount:     "5",
			fieldTotalLatencyMs:   "20000", // 200ms avg
			fieldAnonymityLevel:   "elite",
			fieldLastSuccessTime:  lastSuccessTime,
			fieldConsecutiveFails: "0",
		})
		score := scorer.CalculateScore(ctx, proxyIP)
		// Expected score breakdown:
		// SuccessRate: 100/105 = 0.952 -> 0.952 * 0.5 = 0.476
		// Latency: avg = 200ms. 1 - (200/5000) = 0.96 -> 0.96 * 0.3 = 0.288
		// Anonymity: elite = 1.0 -> 1.0 * 0.15 = 0.15
		// Stability: bonus = 1 - (1/24) = 0.958. score = (1-0)*(0.5+0.5*0.958) = 0.979 -> 0.979 * 0.05 = 0.049
		// Total: 0.476 + 0.288 + 0.15 + 0.049 = 0.963
		assert.InDelta(t, 0.963, score, 0.01)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Bad Proxy", func(t *testing.T) {
		lastSuccessTime := time.Now().UTC().Add(-48 * time.Hour).Format(time.RFC3339)
		mock.ExpectHGetAll(key).SetVal(map[string]string{
			fieldSuccessCount:     "10",
			fieldFailureCount:     "90",
			fieldTotalLatencyMs:   "40000", // 4000ms avg
			fieldAnonymityLevel:   "transparent",
			fieldLastSuccessTime:  lastSuccessTime,
			fieldConsecutiveFails: "8",
		})
		score := scorer.CalculateScore(ctx, proxyIP)
		// Expected score breakdown:
		// SuccessRate: 10/100 = 0.1 -> 0.1 * 0.5 = 0.05
		// Latency: avg = 4000ms. 1 - (4000/5000) = 0.2 -> 0.2 * 0.3 = 0.06
		// Anonymity: transparent = 0.3 -> 0.3 * 0.15 = 0.045
		// Stability: failPenalty=1.0. bonus=0. score=(1-1)*(...)=0 -> 0 * 0.05 = 0
		// Total: 0.05 + 0.06 + 0.045 + 0 = 0.155
		assert.InDelta(t, 0.155, score, 0.01)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestQualityScorer_RemoveProxyMetrics(t *testing.T) {
	db, mock := redismock.NewClientMock()
	scorer := NewQualityScorer(db)
	ctx := context.Background()
	proxyIP := testProxyIP
	key := proxyMetricsKeyPrefix + proxyIP

	mock.ExpectDel(key).SetVal(1)

	err := scorer.RemoveProxyMetrics(ctx, proxyIP)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQualityScorer_RemoveProxyMetrics_Error(t *testing.T) {
	db, mock := redismock.NewClientMock()
	scorer := NewQualityScorer(db)
	ctx := context.Background()
	proxyIP := testProxyIP
	key := proxyMetricsKeyPrefix + proxyIP

	mock.ExpectDel(key).SetErr(fmt.Errorf("redis error"))

	err := scorer.RemoveProxyMetrics(ctx, proxyIP)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis error")
	assert.NoError(t, mock.ExpectationsWereMet())
}
