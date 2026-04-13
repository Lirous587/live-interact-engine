package router

import (
	"live-interact-engine/services/api-service/internal/adapter/mapper"
	"live-interact-engine/services/api-service/internal/grpc_clients"
	"live-interact-engine/services/api-service/internal/handler"
	"live-interact-engine/services/api-service/internal/middleware"
	"live-interact-engine/shared/env"
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "live-interact-engine/api/openapi"
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
	// 注册 Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 创建 v1 路由组
	v1 := r.Group("/v1")

	// 注册弹幕相关路由
	registerDanmaka(v1)

	// 注册用户相关路由
	registerUser(v1)

	// 注册房间相关路由
	registerRoom(v1)
}

func registerDanmaka(r *gin.RouterGroup) {
	danmakuServiceURL := env.GetString("DANMAKU_SERVICE_URL", "danmaku-service:9093")

	// 创建 danmaku 客户端
	danmakuClient, err := grpc_clients.NewDanmakuClient(danmakuServiceURL)
	if err != nil {
		log.Fatalf("创建 danmaku 客户端失败: %v", err)
	}

	// 加入管理列表
	clients = append(clients, danmakuClient)

	// 创建 danmaku mapper（业务适配层）
	danmakuMapper := mapper.NewDanmakuMapper(danmakuClient)

	// 创建 handler
	danmakuHandler := handler.NewDanmakuHandler(danmakuMapper)

	// 注册路由
	dg := r.Group("/danmaku")
	{
		dg.POST("/send", danmakuHandler.SendDanmaku)
		dg.GET("/subscribe", danmakuHandler.SubscribeDanmaku)
	}
}

func registerUser(r *gin.RouterGroup) {
	userServiceURL := env.GetString("USER_SERVICE_URL", "user-service:9094")

	// 创建 user 客户端
	userClient, err := grpc_clients.NewUserClient(userServiceURL)
	if err != nil {
		log.Fatalf("创建 user 客户端失败: %v", err)
	}

	// 加入管理列表
	clients = append(clients, userClient)

	// 创建 user mapper（业务适配层）
	userMapper := mapper.NewUserMapper(userClient)

	// 创建 handler
	userHandler := handler.NewUserHandler(userMapper)

	// 注册路由
	ug := r.Group("/user")
	{
		ug.POST("/register", userHandler.Register)
		ug.POST("/login", userHandler.Login)
		ug.GET("/:user_id", userHandler.GetUser)
	}
}

func registerRoom(r *gin.RouterGroup) {
	userServiceAddr := env.GetString("USER_SERVICE_URL", "user-service:9094")
	authMiddleware, err := middleware.NewAuthMiddleware(userServiceAddr)
	if err != nil {
		log.Fatalf("初始化 auth 中间件失败: %v", err)
	}

	roomServiceURL := env.GetString("ROOM_SERVICE_URL", "room-service:9095")

	// 创建 room 客户端
	roomClient, err := grpc_clients.NewRoomClient(roomServiceURL)
	if err != nil {
		log.Fatalf("创建 room 客户端失败: %v", err)
	}

	// 加入管理列表
	clients = append(clients, roomClient)

	// 创建 room mapper（业务适配层）
	roomMapper := mapper.NewRoomMapper(roomClient)

	// 创建 handler
	roomHandler := handler.NewRoomHandler(roomMapper)

	// 注册路由
	rg := r.Group("/room")
	{
		rg.POST("/create", authMiddleware.Validate(), roomHandler.CreateRoom)
		rg.GET("/:room_id", roomHandler.GetRoom)
		rg.POST("/assign-role", authMiddleware.Validate(), roomHandler.AssignRole)
		rg.GET("/:room_id/user/:user_id/role", roomHandler.GetUserRoomRole)
	}
}

func CloseClients() {
	for _, client := range clients {
		if err := client.Close(); err != nil {
			log.Printf("关闭客户端失败: %v", err)
			// 可以继续关闭其他客户端，或者立即返回错误
		}
	}
}
