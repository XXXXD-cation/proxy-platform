package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// setupSecurityTestRouter 设置安全测试路由
func setupSecurityTestRouter(config *SecurityConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 添加安全中间件
	router.Use(NewSecurityMiddleware(config).Middleware())

	// 测试路由
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	router.POST("/upload", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "upload success"})
	})

	return router
}

// TestSecurityMiddleware_Integration 安全中间件集成测试
// setupSecurityConfig 创建安全中间件配置
func setupSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		// CORS配置
		AllowOrigins:     []string{"https://example.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           86400,

		// 安全头配置
		ContentSecurityPolicy: "default-src 'self'",
		XFrameOptions:         "DENY",
		XContentTypeOptions:   "nosniff",
		XSSProtection:         "1; mode=block",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		HSTSMaxAge:            31536000,
		HSTSIncludeSubdomains: true,

		// CSRF配置
		CSRFEnabled:   true,
		CSRFTokenName: "X-CSRF-Token",
		CSRFSecret:    "test-csrf-secret",

		// 输入验证配置
		MaxRequestSize:    1024 * 1024, // 1MB
		AllowedFileTypes:  []string{"image/jpeg", "image/png", "text/plain"},
		BlockedUserAgents: []string{"malicious-bot"},

		// IP白名单/黑名单
		IPWhitelist: []string{"192.168.1.0/24", "10.0.0.0/8"},
		IPBlacklist: []string{"192.168.1.100"},
	}
}

// TestSecurityMiddleware_CORS CORS安全测试
func TestSecurityMiddleware_CORS(t *testing.T) {
	config := setupSecurityConfig()
	router := setupSecurityTestRouter(config)

	// CORS安全头测试
	req, err := http.NewRequest("OPTIONS", "/test", http.NoBody)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证CORS头
	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("CORS Allow-Origin头不正确")
	}

	if !strings.Contains(w.Header().Get("Access-Control-Allow-Methods"), "GET") {
		t.Errorf("CORS Allow-Methods头不正确")
	}

	if !strings.Contains(w.Header().Get("Access-Control-Allow-Headers"), "Content-Type") {
		t.Errorf("CORS Allow-Headers头不正确")
	}

	t.Logf("✅ CORS安全头测试通过")

	// 非法Origin拒绝测试
	req2, err := http.NewRequest("GET", "/test", http.NoBody)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req2.Header.Set("Origin", "https://malicious.com")

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// 不应该返回CORS头
	if w2.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("不应该对非法Origin返回CORS头")
	}

	t.Logf("✅ 非法Origin拒绝测试通过")
}

// TestSecurityMiddleware_Headers 安全头设置测试
func TestSecurityMiddleware_Headers(t *testing.T) {
	config := setupSecurityConfig()
	router := setupSecurityTestRouter(config)

	req, err := http.NewRequest("GET", "/test", http.NoBody)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证各种安全头
	securityHeaders := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Content-Security-Policy":   "default-src 'self'",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
	}

	for header, expectedValue := range securityHeaders {
		if w.Header().Get(header) != expectedValue {
			t.Errorf("安全头 %s 不正确: 期望 %s, 得到 %s", header, expectedValue, w.Header().Get(header))
		}
	}

	t.Logf("✅ 安全头设置测试通过")
}

// TestSecurityMiddleware_IPFiltering IP过滤测试
func TestSecurityMiddleware_IPFiltering(t *testing.T) {
	config := setupSecurityConfig()
	router := setupSecurityTestRouter(config)

	// IP白名单测试
	req, err := http.NewRequest("GET", "/test", http.NoBody)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req.RemoteAddr = "192.168.1.50:12345"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("白名单IP应该被允许，但状态码为: %d", w.Code)
	}

	t.Logf("✅ IP白名单测试通过")

	// IP黑名单测试
	req2, err := http.NewRequest("GET", "/test", http.NoBody)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req2.RemoteAddr = "192.168.1.100:12345"

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusForbidden {
		t.Errorf("黑名单IP应该被拒绝，但状态码为: %d", w2.Code)
	}

	t.Logf("✅ IP黑名单测试通过")
}

// TestSecurityMiddleware_UserAgentAndContent 用户代理和内容过滤测试
func TestSecurityMiddleware_UserAgentAndContent(t *testing.T) {
	config := setupSecurityConfig()
	router := setupSecurityTestRouter(config)

	// 用户代理过滤测试
	req, err := http.NewRequest("GET", "/test", http.NoBody)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("允许的User-Agent应该通过，但状态码为: %d", w.Code)
	}

	// 不允许的User-Agent
	req2, err := http.NewRequest("GET", "/test", http.NoBody)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req2.Header.Set("User-Agent", "malicious-bot/1.0")

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusForbidden {
		t.Errorf("不允许的User-Agent应该被拒绝，但状态码为: %d", w2.Code)
	}

	t.Logf("✅ 用户代理过滤测试通过")

	// 请求大小限制测试
	largeBody := bytes.Repeat([]byte("a"), 2*1024*1024) // 2MB
	req3, err := http.NewRequest("POST", "/upload", bytes.NewReader(largeBody))
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req3.Header.Set("Content-Type", "text/plain")

	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("大请求应该被拒绝，但状态码为: %d", w3.Code)
	}

	t.Logf("✅ 请求大小限制测试通过")

	// 文件类型过滤测试
	req4, err := http.NewRequest("POST", "/upload", strings.NewReader("test content"))
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req4.Header.Set("Content-Type", "text/plain")

	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)

	if w4.Code != http.StatusOK {
		t.Errorf("允许的文件类型应该通过，但状态码为: %d", w4.Code)
	}

	// 不允许的文件类型
	req5, err := http.NewRequest("POST", "/upload", strings.NewReader("test content"))
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req5.Header.Set("Content-Type", "application/javascript")

	w5 := httptest.NewRecorder()
	router.ServeHTTP(w5, req5)

	if w5.Code != http.StatusUnsupportedMediaType {
		t.Errorf("不允许的文件类型应该被拒绝，但状态码为: %d", w5.Code)
	}

	t.Logf("✅ 文件类型过滤测试通过")
}

// TestSecurityMiddleware_Performance 安全中间件性能测试
func TestSecurityMiddleware_Performance(t *testing.T) {
	config := &SecurityConfig{
		// CORS配置
		AllowOrigins:     []string{"https://example.com"},
		AllowMethods:     []string{"GET", "POST", "PUT"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,

		// IP过滤配置
		IPWhitelist:       []string{"192.168.1.0/24"},
		BlockedUserAgents: []string{"malicious-bot"},
		MaxRequestSize:    1024 * 1024,
		AllowedFileTypes:  []string{"text/plain"},

		// CSRF配置 - 在性能测试中禁用以避免token验证
		CSRFEnabled:   false,
		CSRFTokenName: "X-CSRF-Token",
	}

	router := setupSecurityTestRouter(config)

	t.Run("安全中间件性能测试", func(t *testing.T) {
		// 准备测试请求
		req, err := http.NewRequest("GET", "/test", http.NoBody)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("X-Forwarded-For", "192.168.1.50")
		req.Header.Set("User-Agent", "Mozilla/5.0")

		// 测试1000次请求的性能
		totalRequests := 1000
		start := time.Now()

		for i := 0; i < totalRequests; i++ {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("第%d次请求失败，状态码: %d", i+1, w.Code)
			}
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(totalRequests)

		t.Logf("✅ %d次请求总时间: %v", totalRequests, duration)
		t.Logf("✅ 平均每次请求时间: %v", avgDuration)

		// 验证性能要求（每次请求应该在1ms内完成）
		if avgDuration > time.Millisecond {
			t.Errorf("安全中间件性能过低: 平均 %v/请求 (要求 <1ms)", avgDuration)
		}
	})
}

// BenchmarkSecurityMiddleware 安全中间件基准测试
func BenchmarkSecurityMiddleware(b *testing.B) {
	config := &SecurityConfig{
		// CORS配置
		AllowOrigins:     []string{"https://example.com"},
		AllowMethods:     []string{"GET", "POST", "PUT"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,

		// IP过滤配置
		IPWhitelist:       []string{"192.168.1.0/24"},
		BlockedUserAgents: []string{"malicious-bot"},
		MaxRequestSize:    1024 * 1024,
		AllowedFileTypes:  []string{"text/plain"},

		// CSRF配置 - 在基准测试中禁用
		CSRFEnabled:   false,
		CSRFTokenName: "X-CSRF-Token",
	}

	router := setupSecurityTestRouter(config)

	req, err := http.NewRequest("GET", "/test", http.NoBody)
	if err != nil {
		b.Fatalf("创建请求失败: %v", err)
	}
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("X-Forwarded-For", "192.168.1.50")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("请求失败，状态码: %d", w.Code)
			}
		}
	})
}
