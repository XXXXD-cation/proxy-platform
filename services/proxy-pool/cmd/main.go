package main

import (
	"log"
	"net/http"
	"time"

	"github.com/XXXXD-cation/proxy-platform/pkg/config"
	"github.com/XXXXD-cation/proxy-platform/pkg/logger"
)

const (
	// ServerReadTimeout 服务器读取超时时间
	ServerReadTimeout = 10 * time.Second
	// ServerWriteTimeout 服务器写入超时时间
	ServerWriteTimeout = 10 * time.Second
	// ServerIdleTimeout 服务器空闲超时时间
	ServerIdleTimeout = 60 * time.Second
)

func main() {
	// 加载配置
	cfg, err := config.LoadFromDir("configs", "proxy-pool")
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// 初始化日志记录器
	if err := logger.Init(&cfg.Log); err != nil {
		log.Fatalf("❌ Failed to initialize logger: %v", err)
	}

	logger.Info("🚀 Proxy Pool Service starting...")

	// 简单的健康检查端点
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok","service":"proxy-pool"}`)); err != nil {
			logger.WithField("error", err).Warn("Failed to write health check response")
		}
	})

	// 基础路由
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"message":"Proxy Pool Service","version":"1.0.0"}`)); err != nil {
			logger.WithField("error", err).Warn("Failed to write root response")
		}
	})

	// 代理池API端点
	http.HandleFunc("/api/proxies", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		response := `{"proxies":[],"total":0,"message":"Service ready for proxy pool management"}`
		if _, err := w.Write([]byte(response)); err != nil {
			logger.WithField("error", err).Warn("Failed to write proxies API response")
		}
	})

	addr := cfg.GetServerAddr()
	logger.Infof("🎯 Proxy Pool server listening on %s", addr)

	// 配置HTTP服务器的超时设置
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  ServerReadTimeout,
		WriteTimeout: ServerWriteTimeout,
		IdleTimeout:  ServerIdleTimeout,
	}

	logger.Fatal(server.ListenAndServe())
}
