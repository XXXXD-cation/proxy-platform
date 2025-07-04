package validator

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/XXXXD-cation/proxy-platform/services/proxy-pool/providers"
	"github.com/stretchr/testify/assert"
)

// --- Mocks and Test Setup ---

// a mock target server that our proxy will try to connect to.
func newMockTargetServer(t *testing.T, responseBody string, desiredStatusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(desiredStatusCode)
		_, err := w.Write([]byte(responseBody))
		assert.NoError(t, err)
	}))
}

// a mock proxy server.
func newMockProxyServer(_ *testing.T, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// --- Test Cases ---

type testCase struct {
	name                 string
	setupProxyServer     func(t *testing.T, targetURL string) *httptest.Server
	setupTargetServer    func(t *testing.T) *httptest.Server
	checkTarget          CheckTarget
	expectedAvailability bool
	expectedAnonymity    AnonymityLevel
	expectedErrorPart    string
}

func createSuccessTestCase(name string, anonymity AnonymityLevel, proxyRespBody string) testCase {
	return testCase{
		name: name,
		setupProxyServer: func(t *testing.T, _ string) *httptest.Server {
			return newMockProxyServer(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(proxyRespBody))
				assert.NoError(t, err)
			})
		},
		setupTargetServer: func(t *testing.T) *httptest.Server {
			return newMockTargetServer(t, ``, http.StatusOK)
		},
		checkTarget:          CheckTarget{URL: "", MustContain: `"origin"`},
		expectedAvailability: true,
		expectedAnonymity:    anonymity,
	}
}

func TestValidator_Validate(t *testing.T) {
	const fakePublicIP = "1.2.3.4"

	goodProxy := &providers.ProxyIP{
		Address:  "127.0.0.1",
		Port:     0,
		Username: "user",
		Password: "password",
	}

	testCases := []testCase{
		createSuccessTestCase("成功验证 - 高匿名", Elite, `{"origin":"5.5.5.5", "headers": {}}`),
		createSuccessTestCase("成功验证 - 普通匿名", Anonymous, `{"origin":"5.g.g.g", "headers": {"Via": "1.1 CachingProxy"}}`),
		createSuccessTestCase("成功验证 - 透明代理", Transparent, `{"origin":"5.5.5.5", "headers": {"X-Forwarded-For": "`+fakePublicIP+`"}}`),
		{
			name: "代理服务器超时",
			setupProxyServer: func(_ *testing.T, _ string) *httptest.Server {
				return newMockProxyServer(nil, func(_ http.ResponseWriter, _ *http.Request) {
					time.Sleep(150 * time.Millisecond)
				})
			},
			setupTargetServer:    func(t *testing.T) *httptest.Server { return newMockTargetServer(t, "", http.StatusOK) },
			checkTarget:          CheckTarget{URL: ""},
			expectedAvailability: false,
			expectedErrorPart:    "context deadline exceeded",
		},
		{
			name: "目标服务器返回错误状态码",
			setupProxyServer: func(t *testing.T, targetURL string) *httptest.Server {
				return newMockProxyServer(t, func(w http.ResponseWriter, _ *http.Request) {
					// #nosec G107
					resp, err := http.Get(targetURL)
					assert.NoError(t, err)
					defer resp.Body.Close()
					w.WriteHeader(resp.StatusCode)
				})
			},
			setupTargetServer:    func(t *testing.T) *httptest.Server { return newMockTargetServer(t, "not found", http.StatusNotFound) },
			checkTarget:          CheckTarget{URL: ""},
			expectedAvailability: false,
			expectedErrorPart:    "bad status code: 404",
		},
		{
			name: "响应内容不匹配",
			setupProxyServer: func(t *testing.T, _ string) *httptest.Server {
				return newMockProxyServer(t, func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte(`{"some_other_content":"nothing"}`))
					assert.NoError(t, err)
				})
			},
			setupTargetServer:    func(t *testing.T) *httptest.Server { return newMockTargetServer(t, ``, http.StatusOK) },
			checkTarget:          CheckTarget{URL: "", MustContain: `"origin"`},
			expectedAvailability: false,
			expectedErrorPart:    "response content mismatch",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runValidationTestCase(t, &tc, goodProxy, fakePublicIP)
		})
	}
}

func runValidationTestCase(t *testing.T, tc *testCase, baseProxy *providers.ProxyIP, fakePublicIP string) {
	targetServer := tc.setupTargetServer(t)
	defer targetServer.Close()
	proxyServer := tc.setupProxyServer(t, targetServer.URL)
	defer proxyServer.Close()

	validator := &Validator{
		timeout:  100 * time.Millisecond,
		publicIP: fakePublicIP,
	}

	proxyURL, err := url.Parse(proxyServer.URL)
	assert.NoError(t, err)
	portInt, err := strconv.Atoi(proxyURL.Port())
	assert.NoError(t, err)

	if portInt <= 0 || portInt > 65535 {
		t.Fatalf("无效的端口号: %d", portInt)
	}
	port := uint16(portInt)

	testProxy := *baseProxy // copy
	testProxy.Address = proxyURL.Hostname()
	testProxy.Port = port

	checkTarget := tc.checkTarget
	checkTarget.URL = targetServer.URL

	result := validator.Validate(context.Background(), &testProxy, checkTarget)

	assert.Equal(t, tc.expectedAvailability, result.IsAvailable)
	if tc.expectedErrorPart != "" {
		assert.Contains(t, result.ErrorMessage, tc.expectedErrorPart)
	}
	if result.IsAvailable {
		assert.Equal(t, tc.expectedAnonymity, result.Anonymity)
	}
}
