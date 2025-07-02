package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// 网关服务 - 基础骨架
	fmt.Println("🚀 Proxy Platform Gateway starting...")
	
	// 简单的健康检查端点
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"gateway"}`))
	})
	
	// 基础路由
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Proxy Platform API Gateway","version":"1.0.0"}`))
	})
	
	fmt.Println("🌐 Gateway server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
} 