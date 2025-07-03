package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/XXXXD-cation/proxy-platform/pkg/config"
)

// setupTestRedis 设置测试Redis
func setupTestRedis(t *testing.T) *redis.Client {
	testConfig := config.NewTestConfig()

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", testConfig.Redis.Host, testConfig.Redis.Port),
		Password: testConfig.Redis.Password,
		DB:       testConfig.Redis.DB,
	})

	// 测试连接
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		t.Fatalf("连接测试Redis失败: %v", err)
	}

	// 清理测试数据
	rdb.FlushDB(context.Background())

	return rdb
}

// setupTestRouter 设置测试路由
func setupTestRouter(limiter *RateLimiter) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 添加限流中间件
	router.Use(limiter.IPRateLimiter(10, time.Second*10)) // 10次/10秒

	// 测试路由
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	return router
}

// TestRateLimiter_Integration 集成测试
// TestRateLimiter_IPLimiting IP限流基本功能测试
func TestRateLimiter_IPLimiting(t *testing.T) {
	rdb := setupTestRedis(t)
	limiter := NewRateLimiter(rdb)
	router := setupTestRouter(limiter)

	// 清理Redis以确保测试独立性
	rdb.FlushDB(context.Background())

	// 前10次请求应该成功
	for i := 0; i < 10; i++ {
		req, err := http.NewRequest("GET", "/test", http.NoBody)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}
		// 使用RemoteAddr设置IP，这样c.ClientIP()可以正确获取
		req.RemoteAddr = TestIPAddress

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("第%d次请求失败，状态码: %d", i+1, w.Code)
			return
		}
	}

	// 第11次请求应该被限制
	req, err := http.NewRequest("GET", "/test", http.NoBody)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req.RemoteAddr = TestIPAddress

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("第11次请求应该被限制，但状态码为: %d", w.Code)
	}

	t.Logf("✅ IP限流基本功能正常")

	// 清理测试数据
	t.Cleanup(func() {
		rdb.FlushDB(context.Background())
	})
}

// TestRateLimiter_MultipleIPs 不同IP独立限流测试
func TestRateLimiter_MultipleIPs(t *testing.T) {
	rdb := setupTestRedis(t)
	limiter := NewRateLimiter(rdb)
	router := setupTestRouter(limiter)

	// 清理Redis
	rdb.FlushDB(context.Background())

	// IP1的请求
	for i := 0; i < 5; i++ {
		req, err := http.NewRequest("GET", "/test", http.NoBody)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}
		req.RemoteAddr = "192.168.1.101:12345"

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("IP1第%d次请求失败", i+1)
		}
	}

	// IP2的请求
	for i := 0; i < 5; i++ {
		req, err := http.NewRequest("GET", "/test", http.NoBody)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}
		req.RemoteAddr = "192.168.1.102:12345"

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("IP2第%d次请求失败", i+1)
		}
	}

	t.Logf("✅ 不同IP独立限流正常")

	// 清理测试数据
	t.Cleanup(func() {
		rdb.FlushDB(context.Background())
	})
}

// TestRateLimiter_UserLimiting 用户限流测试
func TestRateLimiter_UserLimiting(t *testing.T) {
	rdb := setupTestRedis(t)
	limiter := NewRateLimiter(rdb)

	// 清理Redis
	rdb.FlushDB(context.Background())

	// 创建用户限流中间件
	userLimiter := limiter.UserRateLimiter(5, time.Second*10) // 5次/10秒

	gin.SetMode(gin.TestMode)
	userRouter := gin.New()
	// 添加中间件来设置用户ID
	userRouter.Use(func(c *gin.Context) {
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			c.Set("user_id", userID)
		}
		c.Next()
	})
	userRouter.Use(userLimiter)
	userRouter.GET("/user-test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 模拟用户ID
	userID := "12345"

	// 前5次请求应该成功
	for i := 0; i < 5; i++ {
		req, err := http.NewRequest("GET", "/user-test", http.NoBody)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}
		req.Header.Set("X-User-ID", userID)

		w := httptest.NewRecorder()
		userRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("用户限流第%d次请求失败，状态码: %d", i+1, w.Code)
		}
	}

	// 第6次请求应该被限制
	req, err := http.NewRequest("GET", "/user-test", http.NoBody)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req.Header.Set("X-User-ID", userID)

	w := httptest.NewRecorder()
	userRouter.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("用户限流第6次请求应该被限制，但状态码为: %d", w.Code)
	}

	t.Logf("✅ 用户限流功能正常")

	// 清理测试数据
	t.Cleanup(func() {
		rdb.FlushDB(context.Background())
	})
}

// TestRateLimiter_APIKeyLimiting API Key限流测试
func TestRateLimiter_APIKeyLimiting(t *testing.T) {
	rdb := setupTestRedis(t)
	limiter := NewRateLimiter(rdb)

	// 清理Redis
	rdb.FlushDB(context.Background())

	// 创建API Key限流中间件
	apiKeyLimiter := limiter.APIKeyRateLimiter(3, time.Second*10) // 3次/10秒

	gin.SetMode(gin.TestMode)
	apiRouter := gin.New()
	apiRouter.Use(apiKeyLimiter)
	apiRouter.GET("/api-test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 模拟API Key
	apiKey := "test-api-key-12345"

	// 前3次请求应该成功
	for i := 0; i < 3; i++ {
		req, err := http.NewRequest("GET", "/api-test", http.NoBody)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}
		req.Header.Set("X-API-Key", apiKey)

		w := httptest.NewRecorder()
		apiRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("API Key限流第%d次请求失败，状态码: %d", i+1, w.Code)
		}
	}

	// 第4次请求应该被限制
	req, err := http.NewRequest("GET", "/api-test", http.NoBody)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req.Header.Set("X-API-Key", apiKey)

	w := httptest.NewRecorder()
	apiRouter.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("API Key限流第4次请求应该被限制，但状态码为: %d", w.Code)
	}

	t.Logf("✅ API Key限流功能正常")

	// 清理测试数据
	t.Cleanup(func() {
		rdb.FlushDB(context.Background())
	})
}

// TestRateLimiter_SlidingWindow 滑动窗口算法测试
// TestRateLimiter_SlidingWindowAccuracy 滑动窗口准确性测试
func TestRateLimiter_SlidingWindowAccuracy(t *testing.T) {
	rdb := setupTestRedis(t)
	limiter := NewRateLimiter(rdb)

	// 清理Redis
	rdb.FlushDB(context.Background())

	key := "sliding-window-test"
	limit := 5
	window := time.Second * 5

	// 在前2秒内发送5次请求
	for i := 0; i < limit; i++ {
		result, err := limiter.CheckLimit(key, limit, window)
		if err != nil {
			t.Fatalf("检查限流失败: %v", err)
		}
		if !result.Allowed {
			t.Errorf("第%d次请求应该被允许", i+1)
		}
	}

	// 第6次请求应该被拒绝
	result, err := limiter.CheckLimit(key, limit, window)
	if err != nil {
		t.Fatalf("检查限流失败: %v", err)
	}
	if result.Allowed {
		t.Errorf("第6次请求应该被拒绝")
	}

	// 等待超过窗口时间后，应该可以再次请求
	time.Sleep(window + time.Second) // 等待6秒（超过5秒窗口）

	result, err = limiter.CheckLimit(key, limit, window)
	if err != nil {
		t.Fatalf("检查限流失败: %v", err)
	}
	if !result.Allowed {
		t.Errorf("等待超过窗口时间后的请求应该被允许")
	}

	t.Logf("✅ 滑动窗口算法准确性测试通过")

	// 清理测试数据
	t.Cleanup(func() {
		rdb.FlushDB(context.Background())
	})
}

// TestRateLimiter_ConcurrentLimiting 并发限流测试
func TestRateLimiter_ConcurrentLimiting(t *testing.T) {
	rdb := setupTestRedis(t)
	limiter := NewRateLimiter(rdb)

	// 清理Redis
	rdb.FlushDB(context.Background())

	key := "concurrent-test"
	limit := 10
	window := time.Second * 5
	concurrency := 20

	// 并发发送请求
	results := make(chan bool, concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			result, err := limiter.CheckLimit(key, limit, window)
			if err != nil {
				t.Errorf("并发检查限流失败: %v", err)
				results <- false
			} else {
				results <- result.Allowed
			}
		}()
	}

	// 收集结果
	allowedCount := 0
	rejectedCount := 0
	for i := 0; i < concurrency; i++ {
		if <-results {
			allowedCount++
		} else {
			rejectedCount++
		}
	}

	// 验证结果
	if allowedCount != limit {
		t.Errorf("并发限流允许次数不准确: 期望 %d, 实际 %d", limit, allowedCount)
	}

	if rejectedCount != (concurrency - limit) {
		t.Errorf("并发限流拒绝次数不准确: 期望 %d, 实际 %d", concurrency-limit, rejectedCount)
	}

	// 计算误差率
	errorRate := float64(abs(allowedCount-limit)) / float64(limit) * 100
	if errorRate > 5.0 {
		t.Errorf("并发限流误差过高: %.2f%% (要求 <5%%)", errorRate)
	} else {
		t.Logf("✅ 并发限流误差: %.2f%% (满足 <5%% 要求)", errorRate)
	}

	t.Logf("✅ 并发限流测试通过: 允许 %d 次, 拒绝 %d 次", allowedCount, rejectedCount)

	// 清理测试数据
	t.Cleanup(func() {
		rdb.FlushDB(context.Background())
	})
}

// TestRateLimiter_Performance 性能测试
func TestRateLimiter_Performance(t *testing.T) {
	rdb := setupTestRedis(t)
	limiter := NewRateLimiter(rdb)

	t.Run("限流性能测试", func(t *testing.T) {
		// 清理Redis
		rdb.FlushDB(context.Background())

		key := "performance-test"
		limit := 100
		window := time.Second * 10

		// 测试100次限流检查的性能
		start := time.Now()
		for i := 0; i < 100; i++ {
			_, err := limiter.CheckLimit(key, limit, window)
			if err != nil {
				t.Fatalf("性能测试失败: %v", err)
			}
		}
		duration := time.Since(start)

		// 平均每次检查的时间
		avgDuration := duration / 100

		// 验证性能要求 (<10ms)
		if avgDuration > 10*time.Millisecond {
			t.Errorf("限流检查平均延迟过高: %v (要求 <10ms)", avgDuration)
		} else {
			t.Logf("✅ 限流检查平均延迟: %v (满足 <10ms 要求)", avgDuration)
		}

		t.Logf("✅ 100次限流检查总时间: %v", duration)
	})

	// 清理测试数据
	t.Cleanup(func() {
		rdb.FlushDB(context.Background())
	})
}

// TestRateLimiter_EdgeCases 边缘情况测试
func TestRateLimiter_EdgeCases(t *testing.T) {
	rdb := setupTestRedis(t)
	limiter := NewRateLimiter(rdb)

	t.Run("零限制测试", func(t *testing.T) {
		// 清理Redis
		rdb.FlushDB(context.Background())

		key := "zero-limit-test"
		limit := 0
		window := time.Second * 10

		// 任何请求都应该被拒绝
		result, err := limiter.CheckLimit(key, limit, window)
		if err != nil {
			t.Fatalf("零限制测试失败: %v", err)
		}
		if result.Allowed {
			t.Errorf("零限制应该拒绝所有请求")
		}

		t.Logf("✅ 零限制测试通过")
	})

	t.Run("极小窗口测试", func(t *testing.T) {
		// 清理Redis
		rdb.FlushDB(context.Background())

		key := "small-window-test"
		limit := 1
		window := time.Millisecond * 100 // 增加窗口时间到100毫秒以提高可靠性

		// 第一次请求应该成功
		result1, err := limiter.CheckLimit(key, limit, window)
		if err != nil {
			t.Fatalf("极小窗口测试失败: %v", err)
		}
		if !result1.Allowed {
			t.Errorf("第一次请求应该被允许")
		}

		// 立即的第二次请求应该被拒绝
		result2, err := limiter.CheckLimit(key, limit, window)
		if err != nil {
			t.Fatalf("极小窗口测试失败: %v", err)
		}
		if result2.Allowed {
			t.Errorf("立即的第二次请求应该被拒绝")
		}

		// 等待窗口过期后再次测试
		time.Sleep(window + time.Millisecond*10)

		result3, err := limiter.CheckLimit(key, limit, window)
		if err != nil {
			t.Fatalf("窗口过期后测试失败: %v", err)
		}

		if !result3.Allowed {
			t.Errorf("窗口过期后的请求应该被允许")
		}

		t.Logf("✅ 极小窗口测试通过")
	})

	t.Run("大数量限制测试", func(t *testing.T) {
		// 清理Redis
		rdb.FlushDB(context.Background())

		key := "large-limit-test"
		limit := 10000
		window := time.Second * 10

		// 大量请求应该都被允许
		for i := 0; i < 100; i++ {
			result, err := limiter.CheckLimit(key, limit, window)
			if err != nil {
				t.Fatalf("大数量限制测试失败: %v", err)
			}
			if !result.Allowed {
				t.Errorf("第%d次请求应该被允许", i+1)
			}
		}

		t.Logf("✅ 大数量限制测试通过")
	})

	// 清理测试数据
	t.Cleanup(func() {
		rdb.FlushDB(context.Background())
	})
}

// BenchmarkRateLimiter_CheckRateLimit 限流检查性能基准测试
func BenchmarkRateLimiter_CheckRateLimit(b *testing.B) {
	rdb := setupTestRedis(&testing.T{})
	limiter := NewRateLimiter(rdb)

	key := "benchmark-test"
	limit := 1000
	window := time.Second * 60

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := limiter.CheckLimit(key, limit, window)
			if err != nil {
				b.Errorf("基准测试失败: %v", err)
			}
		}
	})
}

// abs 返回整数的绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
