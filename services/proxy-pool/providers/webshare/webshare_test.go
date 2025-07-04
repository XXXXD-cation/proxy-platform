package webshare

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/XXXXD-cation/proxy-platform/pkg/config"
	"github.com/XXXXD-cation/proxy-platform/services/proxy-pool/providers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider_GetProxyList(t *testing.T) {
	// --- Test Cases Definition ---
	testCases := []struct {
		name           string
		mockResponse   string
		mockStatusCode int
		params         providers.ProxyParams
		wantCount      int
		wantErr        bool
		wantErrMsg     string
	}{
		{
			name: "正常成功获取代理列表",
			mockResponse: `{
				"count": 3,
				"next": null,
				"previous": null,
				"results": [
					{"id": "p1", "proxy_address": "1.1.1.1", "port": 8080, "username": "u1", "password": "p1", "valid": true, "country_code": "US"},
					{"id": "p2", "proxy_address": "2.2.2.2", "port": 8081, "username": "u2", "password": "p2", "valid": true, "country_code": "DE"},
					{"id": "p3", "proxy_address": "3.3.3.3", "port": 8082, "username": "u3", "password": "p3", "valid": false, "country_code": "JP"}
				]
			}`,
			mockStatusCode: http.StatusOK,
			params:         providers.ProxyParams{Page: 1, PageSize: 10},
			wantCount:      2, // p3 is invalid, so it should be filtered out
			wantErr:        false,
		},
		{
			name:           "API返回非200状态码",
			mockResponse:   `{"error": "Internal Server Error"}`,
			mockStatusCode: http.StatusInternalServerError,
			params:         providers.ProxyParams{Page: 1, PageSize: 10},
			wantCount:      0,
			wantErr:        true,
			wantErrMsg:     "webshare: received non-200 status code 500",
		},
		{
			name:           "API返回无效的JSON",
			mockResponse:   `{ "results": [ ...`,
			mockStatusCode: http.StatusOK,
			params:         providers.ProxyParams{Page: 1, PageSize: 10},
			wantCount:      0,
			wantErr:        true,
			wantErrMsg:     "webshare: failed to decode response",
		},
		{
			name: "API返回空结果列表",
			mockResponse: `{
				"count": 0,
				"next": null,
				"previous": null,
				"results": []
			}`,
			mockStatusCode: http.StatusOK,
			params:         providers.ProxyParams{Page: 1, PageSize: 10},
			wantCount:      0,
			wantErr:        false,
		},
	}

	// --- Test Execution ---
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. 创建模拟服务器
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// 验证请求参数是否正确传递
				assert.Equal(t, fmt.Sprintf("/proxy/list/?mode=direct&page=%d&page_size=%d", tc.params.Page, tc.params.PageSize), r.URL.String())
				assert.Equal(t, "Token test-api-key", r.Header.Get("Authorization"))

				w.WriteHeader(tc.mockStatusCode)
				_, _ = w.Write([]byte(tc.mockResponse))
			}))
			defer server.Close()

			// 2. 创建 Provider 实例，并将其指向模拟服务器
			mockConfig := config.WebshareConfig{
				Enabled: true,
				APIKey:  "test-api-key",
			}
			provider := NewProvider(mockConfig)
			provider.baseURL = server.URL // Override baseURL to point to our test server

			// 3. 调用被测试的方法
			proxies, err := provider.GetProxyList(context.Background(), tc.params)

			// 4. 断言结果
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrMsg)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, proxies)
			}

			assert.Len(t, proxies, tc.wantCount)
			if tc.wantCount > 0 {
				// 验证第一个代理的数据是否正确映射
				proxy := proxies[0]
				assert.Equal(t, "p1", proxy.ID)
				assert.Equal(t, "1.1.1.1", proxy.Address)
				assert.Equal(t, uint16(8080), proxy.Port)
				assert.Equal(t, "u1", proxy.Username)
				assert.Equal(t, "p1", proxy.Password)
				assert.Equal(t, "US", proxy.Country)
			}
		})
	}
}
