// Package middleware 提供HTTP中间件功能
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter 限流器
type RateLimiter struct {
	redis *redis.Client
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Limit    int                       // 限制数量
	Window   time.Duration             // 时间窗口
	KeyFunc  func(*gin.Context) string // 键生成函数
	Message  string                    // 限流消息
	SkipFunc func(*gin.Context) bool   // 跳过检查函数
}

// RateLimitResult 限流结果
type RateLimitResult struct {
	Allowed    bool          // 是否允许
	Remaining  int           // 剩余请求数
	Reset      time.Time     // 重置时间
	RetryAfter time.Duration // 重试延迟
}

// NewRateLimiter 创建限流器
func NewRateLimiter(redis *redis.Client) *RateLimiter {
	return &RateLimiter{
		redis: redis,
	}
}

// Allow 检查是否允许请求
func (r *RateLimiter) Allow(userID int64, limit int, window time.Duration) bool {
	key := fmt.Sprintf("rate_limit:user:%d", userID)
	result, err := r.CheckLimit(key, limit, window)
	if err != nil {
		// 发生错误时，为了系统稳定性，默认允许请求
		return true
	}
	return result.Allowed
}

// CheckLimit 检查限流
func (r *RateLimiter) CheckLimit(key string, limit int, window time.Duration) (*RateLimitResult, error) {
	ctx := context.Background()
	now := time.Now()
	windowStart := now.Add(-window)

	// 使用Lua脚本确保原子性
	script := `
		local key = KEYS[1]
		local window_start = tonumber(ARGV[1])
		local now = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local window_seconds = tonumber(ARGV[4])
		
		-- 清理过期的记录
		redis.call('ZREMRANGEBYSCORE', key, '-inf', window_start)
		
		-- 获取当前窗口内的请求数
		local current_requests = redis.call('ZCARD', key)
		
		-- 检查是否超过限制
		if current_requests < limit then
			-- 添加当前请求（使用唯一值避免重复）
			local member = tostring(now) .. ':' .. tostring(math.random(10000, 99999))
			redis.call('ZADD', key, now, member)
			-- 设置过期时间（向上取整，最少1秒）
			local expire_seconds = math.max(1, math.ceil(window_seconds))
			redis.call('EXPIRE', key, expire_seconds)
			-- 返回允许的结果
			return {1, limit - current_requests - 1, current_requests + 1}
		else
			-- 返回拒绝的结果
			return {0, 0, current_requests}
		end
	`

	// 使用毫秒时间戳以提高精度
	result, err := r.redis.Eval(ctx, script, []string{key},
		windowStart.UnixMilli(), now.UnixMilli(), limit, int(window.Seconds())).Result()

	if err != nil {
		return nil, fmt.Errorf("执行限流脚本失败: %v", err)
	}

	// 解析结果
	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) != 3 {
		return nil, fmt.Errorf("限流脚本返回格式错误")
	}

	// 安全地进行类型断言
	allowedVal, ok := resultSlice[0].(int64)
	if !ok {
		return nil, fmt.Errorf("限流脚本返回格式错误: allowed值类型错误")
	}
	allowed := allowedVal == 1

	remainingVal, ok := resultSlice[1].(int64)
	if !ok {
		return nil, fmt.Errorf("限流脚本返回格式错误: remaining值类型错误")
	}
	remaining := int(remainingVal)

	// 计算重置时间
	resetTime := now.Add(window)
	retryAfter := time.Duration(0)

	if !allowed {
		// 计算下次可以请求的时间
		retryAfter = r.calculateRetryAfter(key, window)
	}

	return &RateLimitResult{
		Allowed:    allowed,
		Remaining:  remaining,
		Reset:      resetTime,
		RetryAfter: retryAfter,
	}, nil
}

// calculateRetryAfter 计算重试延迟
func (r *RateLimiter) calculateRetryAfter(key string, window time.Duration) time.Duration {
	ctx := context.Background()

	// 获取最早的请求时间
	earliest, err := r.redis.ZRange(ctx, key, 0, 0).Result()
	if err != nil || len(earliest) == 0 {
		return window
	}

	// 解析时间戳（毫秒）
	timestampStr := earliest[0]
	// 时间戳格式是 "timestamp:random"，需要提取时间戳部分
	if colonIndex := strings.Index(timestampStr, ":"); colonIndex != -1 {
		timestampStr = timestampStr[:colonIndex]
	}

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return window
	}

	// 使用毫秒时间戳
	earliestTime := time.UnixMilli(timestamp)
	retryAfter := time.Until(earliestTime.Add(window))

	if retryAfter < 0 {
		return time.Second
	}

	return retryAfter
}

// Middleware 限流中间件
func (r *RateLimiter) Middleware(config RateLimitConfig) gin.HandlerFunc {
	// 设置默认值
	if config.KeyFunc == nil {
		config.KeyFunc = func(c *gin.Context) string {
			return c.ClientIP()
		}
	}

	if config.Message == "" {
		config.Message = "请求过于频繁，请稍后再试"
	}

	return func(c *gin.Context) {
		// 检查是否跳过
		if config.SkipFunc != nil && config.SkipFunc(c) {
			c.Next()
			return
		}

		// 获取键
		key := config.KeyFunc(c)

		// 检查限流
		result, err := r.CheckLimit(key, config.Limit, config.Window)
		if err != nil {
			// 发生错误时继续处理请求
			c.Next()
			return
		}

		// 设置响应头
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(result.Reset.Unix(), 10))

		if !result.Allowed {
			c.Header("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limit_exceeded",
				"message":     config.Message,
				"retry_after": result.RetryAfter.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// IPRateLimiter IP限流中间件
func (r *RateLimiter) IPRateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	return r.Middleware(RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyFunc: func(c *gin.Context) string {
			return fmt.Sprintf("rate_limit:ip:%s", c.ClientIP())
		},
		Message: "IP访问过于频繁，请稍后再试",
	})
}

// UserRateLimiter 用户限流中间件
func (r *RateLimiter) UserRateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	return r.Middleware(RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyFunc: func(c *gin.Context) string {
			userID := c.GetString("user_id")
			if userID == "" {
				userID = c.ClientIP()
			}
			return fmt.Sprintf("rate_limit:user:%s", userID)
		},
		Message: "用户访问过于频繁，请稍后再试",
	})
}

// APIKeyRateLimiter API Key限流中间件
func (r *RateLimiter) APIKeyRateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	return r.Middleware(RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyFunc: func(c *gin.Context) string {
			apiKey := c.GetHeader("X-API-Key")
			if apiKey == "" {
				apiKey = c.ClientIP()
			}
			return fmt.Sprintf("rate_limit:apikey:%s", apiKey)
		},
		Message: "API Key访问过于频繁，请稍后再试",
	})
}

// EndpointRateLimiter 端点限流中间件
func (r *RateLimiter) EndpointRateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	return r.Middleware(RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyFunc: func(c *gin.Context) string {
			endpoint := c.Request.Method + ":" + c.Request.URL.Path
			userID := c.GetString("user_id")
			if userID == "" {
				userID = c.ClientIP()
			}
			return fmt.Sprintf("rate_limit:endpoint:%s:%s", endpoint, userID)
		},
		Message: "此端点访问过于频繁，请稍后再试",
	})
}

// GlobalRateLimiter 全局限流中间件
func (r *RateLimiter) GlobalRateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	return r.Middleware(RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyFunc: func(_ *gin.Context) string {
			return "rate_limit:global"
		},
		Message: "系统繁忙，请稍后再试",
	})
}

// GetRateLimitStatus 获取限流状态
func (r *RateLimiter) GetRateLimitStatus(key string, limit int, window time.Duration) (*RateLimitResult, error) {
	ctx := context.Background()
	now := time.Now()
	windowStart := now.Add(-window)

	// 清理过期记录
	r.redis.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(windowStart.Unix(), 10))

	// 获取当前请求数
	currentRequests, err := r.redis.ZCard(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("获取限流状态失败: %v", err)
	}

	remaining := limit - int(currentRequests)
	if remaining < 0 {
		remaining = 0
	}

	return &RateLimitResult{
		Allowed:    remaining > 0,
		Remaining:  remaining,
		Reset:      now.Add(window),
		RetryAfter: 0,
	}, nil
}

// ResetRateLimit 重置限流
func (r *RateLimiter) ResetRateLimit(key string) error {
	ctx := context.Background()
	return r.redis.Del(ctx, key).Err()
}

// ResetUserRateLimit 重置用户限流
func (r *RateLimiter) ResetUserRateLimit(userID int64) error {
	key := fmt.Sprintf("rate_limit:user:%d", userID)
	return r.ResetRateLimit(key)
}

// ResetIPRateLimit 重置IP限流
func (r *RateLimiter) ResetIPRateLimit(ip string) error {
	key := fmt.Sprintf("rate_limit:ip:%s", ip)
	return r.ResetRateLimit(key)
}

// GetLimitedKeys 获取被限流的键
func (r *RateLimiter) GetLimitedKeys(pattern string) ([]string, error) {
	ctx := context.Background()
	return r.redis.Keys(ctx, pattern).Result()
}

// CleanupExpiredKeys 清理过期的限流键
func (r *RateLimiter) CleanupExpiredKeys(pattern string, olderThan time.Duration) error {
	ctx := context.Background()
	keys, err := r.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-olderThan)

	for _, key := range keys {
		// 检查键是否有数据
		count, err := r.redis.ZCard(ctx, key).Result()
		if err != nil {
			continue
		}

		if count == 0 {
			// 删除空键
			r.redis.Del(ctx, key)
			continue
		}

		// 清理过期记录
		r.redis.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(cutoff.Unix(), 10))

		// 如果清理后没有数据，删除键
		count, err = r.redis.ZCard(ctx, key).Result()
		if err == nil && count == 0 {
			r.redis.Del(ctx, key)
		}
	}

	return nil
}

// GetKeyFromContext 从上下文获取限流键
func GetKeyFromContext(c *gin.Context, prefix string) string {
	// 优先使用用户ID
	if userID := c.GetString("user_id"); userID != "" {
		return fmt.Sprintf("%s:user:%s", prefix, userID)
	}

	// 其次使用API Key
	if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
		return fmt.Sprintf("%s:apikey:%s", prefix, apiKey)
	}

	// 最后使用IP
	return fmt.Sprintf("%s:ip:%s", prefix, c.ClientIP())
}

// FormatDuration 格式化时间间隔
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Nanoseconds()/NanosecondsToMilliseconds)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}
