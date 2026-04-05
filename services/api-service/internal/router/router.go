package router

import (
	"live-interact-engine/services/api-service/internal/grpc_clients"
	"live-interact-engine/services/api-service/internal/handler"
	"live-interact-engine/shared/env"
	"log"

	"github.com/gin-gonic/gin"
)

// Client 通用客户端接口
type Client interface {
	Close() error
}

// 全局变量保存所有客户端
var (
	clients = make([]Client, 0)
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.RouterGroup) {
	// 注册弹幕相关路由
	registerDanmuka(r)
}

func registerDanmuka(r *gin.RouterGroup) {
	danmakuServiceURL := env.GetString("DANMAKU_SERVICE_URL", "danmaku-service:9093")

	// 创建 danmaku 客户端
	danmakuClient, err := grpc_clients.NewDanmakuClient(danmakuServiceURL)
	if err != nil {
		log.Fatalf("创建 danmaku 客户端失败: %v", err)
	}

	// 加入管理列表
	clients = append(clients, danmakuClient)

	// 创建 handler
	danmakuHandler := handler.NewDanmakuHandler(danmakuClient)

	// 注册路由
	r.POST("/danmaku/send", danmakuHandler.SendDanmaku)
	r.GET("/danmaku/subscribe", danmakuHandler.SubscribeDanmaku)
}

func CloseClients() {
	for _, client := range clients {
		if err := client.Close(); err != nil {
			log.Printf("关闭客户端失败: %v", err)
			// 可以继续关闭其他客户端，或者立即返回错误
		}
	}
}
