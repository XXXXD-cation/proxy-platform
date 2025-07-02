package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
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
	// 网关服务 - 基础骨架
	fmt.Println("🚀 Proxy Platform Gateway starting...")

	// 简单的健康检查端点
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok","service":"gateway"}`)); err != nil {
			log.Printf("Failed to write response: %v", err)
		}
	})

	// 基础路由
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"message":"Proxy Platform API Gateway","version":"1.0.0"}`)); err != nil {
			log.Printf("Failed to write response: %v", err)
		}
	})

	fmt.Println("🌐 Gateway server listening on :8080")

	// 配置HTTP服务器的超时设置
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  ServerReadTimeout,
		WriteTimeout: ServerWriteTimeout,
		IdleTimeout:  ServerIdleTimeout,
	}

	log.Fatal(server.ListenAndServe())
}
