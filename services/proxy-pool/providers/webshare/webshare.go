package webshare

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/XXXXD-cation/proxy-platform/pkg/config"
	"github.com/XXXXD-cation/proxy-platform/services/proxy-pool/providers"
)

const (
	defaultBaseURL = "https://proxy.webshare.io/api/v2"
)

// Provider 实现了 providers.ProxyProvider 接口，用于从 Webshare 获取代理。
type Provider struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

// NewProvider 创建一个新的 Webshare Provider 实例。
func NewProvider(cfg config.WebshareConfig) *Provider {
	return &Provider{
		client:  &http.Client{},
		apiKey:  cfg.APIKey,
		baseURL: defaultBaseURL,
	}
}

// GetProxyList 从 Webshare API 获取代理列表。
func (p *Provider) GetProxyList(ctx context.Context, params providers.ProxyParams) ([]providers.ProxyIP, error) {
	// 1. 构建请求URL
	reqURL := fmt.Sprintf("%s/proxy/list/?mode=direct&page=%d&page_size=%d", p.baseURL, params.Page, params.PageSize)

	// 2. 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("webshare: failed to create request: %w", err)
	}

	// 3. 设置认证头
	req.Header.Set("Authorization", "Token "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// 4. 发送请求
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("webshare: failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 5. 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("webshare: received non-200 status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 6. 解析JSON响应
	var webshareResp Response
	if err := json.NewDecoder(resp.Body).Decode(&webshareResp); err != nil {
		return nil, fmt.Errorf("webshare: failed to decode response: %w", err)
	}

	// 7. 将 Webshare 代理格式转换为内部标准格式
	return toProxyIPs(webshareResp.Results), nil
}

// toProxyIPs 是一个辅助函数，用于将 []webshare.Proxy 转换为 []providers.ProxyIP。
func toProxyIPs(proxies []Proxy) []providers.ProxyIP {
	result := make([]providers.ProxyIP, 0, len(proxies))
	for _, p := range proxies {
		// Webshare API 返回的 valid 字段可用于过滤掉不可用的代理
		if !p.Valid {
			continue
		}
		result = append(result, providers.ProxyIP{
			ID:            p.ID,
			Address:       p.ProxyAddress,
			Port:          p.Port,
			Username:      p.Username,
			Password:      p.Password,
			Country:       p.CountryCode,
			City:          p.CityName,
			LastCheckedAt: p.LastVerification,
			CreatedAt:     p.CreatedAt,
		})
	}
	return result
}
