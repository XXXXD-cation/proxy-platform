package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/XXXXD-cation/proxy-platform/pkg/logger"
	"github.com/XXXXD-cation/proxy-platform/services/proxy-pool/providers"
)

// httpbinResponse 是用于解析 httpbin.org/get 响应的结构体。
type httpbinResponse struct {
	Origin  string            `json:"origin"`
	Headers map[string]string `json:"headers"`
}

// Validator 用于验证代理的可用性和性能。
type Validator struct {
	timeout  time.Duration
	publicIP string // 验证器所在机器的公网IP
}

// New 创建一个新的 Validator 实例。
// 它会自动尝试获取本机的公网IP，以便进行匿名度检查。
func New(timeout time.Duration) (*Validator, error) {
	ip, err := getMyPublicIP(timeout)
	if err != nil {
		logger.WithError(err).Warn("无法获取公网IP, 匿名度检查的准确性会受影响")
		// 即使无法获取IP，也允许继续，只是匿名检查会不准
		ip = ""
	}
	return &Validator{
		timeout:  timeout,
		publicIP: ip,
	}, nil
}

// getMyPublicIP 通过访问外部服务来获取本机的公网IP。
func getMyPublicIP(timeout time.Duration) (string, error) {
	client := http.Client{Timeout: timeout}
	req, err := http.NewRequest("GET", "https://api.ipify.org", http.NoBody)
	if err != nil {
		return "", fmt.Errorf("创建获取公网IP的请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "proxy-platform-validator")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("获取公网IP失败: %w", err)
	}
	defer resp.Body.Close()
	ipBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取公网IP响应失败: %w", err)
	}
	return string(ipBytes), nil
}

// DefaultCheckTarget 是一个默认的验证目标，用于快速测试。
var DefaultCheckTarget = CheckTarget{
	URL:         "https://httpbin.org/get",
	MustContain: `"origin"`, // httpbin.org/get 的响应会包含来源IP信息
}

// Validate 对单个代理进行验证。
// proxy 参数通过指针传递，以避免复制大的结构体。
func (v *Validator) Validate(ctx context.Context, proxy *providers.ProxyIP, target CheckTarget) ValidationResult {
	// 1. 构建代理URL
	proxyURLStr := fmt.Sprintf("http://%s:%d", proxy.Address, proxy.Port)
	if proxy.Username != "" {
		proxyURLStr = fmt.Sprintf("http://%s:%s@%s:%d", proxy.Username, proxy.Password, proxy.Address, proxy.Port)
	}
	proxyURL, err := url.Parse(proxyURLStr)
	if err != nil {
		return ValidationResult{IsAvailable: false, ErrorMessage: "invalid proxy format"}
	}

	// 2. 创建带有代理和超时的HTTP客户端
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   v.timeout,
	}

	// 3. 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", target.URL, http.NoBody)
	if err != nil {
		return ValidationResult{IsAvailable: false, ErrorMessage: "failed to create request"}
	}
	// 添加用户代理头，模拟真实浏览器
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")

	// 4. 执行请求并计时
	startTime := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(startTime)

	if err != nil {
		return ValidationResult{IsAvailable: false, ErrorMessage: err.Error(), Latency: latency}
	}
	defer resp.Body.Close()

	// 5. 检查状态码
	if resp.StatusCode != http.StatusOK {
		return ValidationResult{IsAvailable: false, ErrorMessage: fmt.Sprintf("bad status code: %d", resp.StatusCode), Latency: latency}
	}

	// 6. 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ValidationResult{IsAvailable: false, ErrorMessage: "failed to read response body", Latency: latency}
	}

	// 7. 检查响应内容
	if target.MustContain != "" && !strings.Contains(string(body), target.MustContain) {
		logger.WithFields(map[string]interface{}{
			"proxy":    proxy.Address,
			"response": string(body),
		}).Warn("response body does not contain expected string")
		return ValidationResult{IsAvailable: false, ErrorMessage: "response content mismatch", Latency: latency}
	}

	// 8. 判断匿名级别
	anonymity := v.detectAnonymity(body)

	return ValidationResult{
		IsAvailable:  true,
		Latency:      latency,
		Anonymity:    anonymity,
		ErrorMessage: "",
	}
}

// detectAnonymity 通过分析httpbin的返回内容来判断代理的匿名级别。
func (v *Validator) detectAnonymity(body []byte) AnonymityLevel {
	var respData httpbinResponse
	if err := json.Unmarshal(body, &respData); err != nil {
		logger.WithError(err).Warn("无法解析httpbin响应以检测匿名性")
		return Unknown
	}

	var isTransparent bool
	var hasProxyHeaders bool

	// 如果我们知道了自己的公网IP，检查它是否出现在任何头中。
	if v.publicIP != "" {
		for _, headerValue := range respData.Headers {
			if strings.Contains(headerValue, v.publicIP) {
				isTransparent = true
				break
			}
		}
	}

	// 检查是否存在已知的代理头。
	proxyHeaderKeys := []string{"Via", "via", "X-Forwarded-For", "x-forwarded-for", "Forwarded", "forwarded", "X-Proxy-Id", "Proxy-Connection"}
	for _, key := range proxyHeaderKeys {
		if _, exists := respData.Headers[key]; exists {
			hasProxyHeaders = true
			break
		}
	}

	if isTransparent {
		return Transparent
	}
	if hasProxyHeaders {
		return Anonymous
	}

	// 如果既不透明也没有找到代理头，则认为是高匿。
	return Elite
}
