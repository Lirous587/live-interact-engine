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

	// 注册礼物相关路由
	registerGift(v1)
}

func registerDanmaka(r *gin.RouterGroup) {
	userServiceAddr := env.GetString("USER_SERVICE_URL", "user-service:9095")
	authMiddleware, err := middleware.NewAuthMiddleware(userServiceAddr)
	if err != nil {
		log.Fatalf("初始化 auth 中间件失败: %v", err)
	}

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
		dg.POST("/send", authMiddleware.Validate(), danmakuHandler.SendDanmaku)
		dg.GET("/subscribe", authMiddleware.Validate(), danmakuHandler.SubscribeDanmaku)
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
	userServiceAddr := env.GetString("USER_SERVICE_URL", "user-service:9095")
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
		rg.POST("/create", authMiddleware.Validate(), roomHandler.SaveRoom)
		rg.GET("/:room_id", roomHandler.GetRoom)
		rg.POST("/assign-role", authMiddleware.Validate(), roomHandler.AssignRole)
		rg.POST("/remove-role", authMiddleware.Validate(), roomHandler.RemoveRole)
		rg.GET("/:room_id/user/:user_id/role", roomHandler.GetUserRoomRole)
		rg.GET("/:room_id/user/:user_id/permission/:permission", authMiddleware.Validate(), roomHandler.CheckPermission)
		rg.POST("/mute", authMiddleware.Validate(), roomHandler.MuteUser)
		rg.POST("/unmute", authMiddleware.Validate(), roomHandler.UnmuteUser)
		rg.GET("/:room_id/user/:user_id/mute-status", roomHandler.IsMuted)
		rg.GET("/:room_id/user/:user_id/mute-info", roomHandler.GetMuteInfo)
		rg.GET("/:room_id/mute-list", roomHandler.GetMuteList)
	}
}

func registerGift(r *gin.RouterGroup) {
	userServiceAddr := env.GetString("USER_SERVICE_URL", "user-service:9095")
	authMiddleware, err := middleware.NewAuthMiddleware(userServiceAddr)
	if err != nil {
		log.Fatalf("初始化 auth 中间件失败: %v", err)
	}

	giftServiceURL := env.GetString("GIFT_SERVICE_URL", "gift-service:9096")

	// 创建 gift 客户端
	giftClient, err := grpc_clients.NewGiftClient(giftServiceURL)
	if err != nil {
		log.Fatalf("创建 gift 客户端失败: %v", err)
	}

	// 加入管理列表
	clients = append(clients, giftClient)

	// 创建 mappers（业务适配层）
	giftMapper := mapper.NewGiftMapper(giftClient)
	walletMapper := mapper.NewWalletMapper(giftClient)
	leaderboardMapper := mapper.NewLeaderboardMapper(giftClient)

	// 创建 handlers
	giftHandler := handler.NewGiftHandler(giftMapper)
	walletHandler := handler.NewWalletHandler(walletMapper)
	leaderboardHandler := handler.NewLeaderboardHandler(leaderboardMapper)

	// 注册路由 - Gift 路由
	gg := r.Group("/gift")
	{
		gg.POST("/send", authMiddleware.Validate(), giftHandler.SendGift)
		gg.GET("/list", giftHandler.ListGifts)
	}

	// 注册路由 - Wallet 路由
	wg := r.Group("/wallet")
	{
		wg.GET("/:user_id/balance", authMiddleware.Validate(), walletHandler.GetWalletBalance)
	}

	// 注册路由 - Leaderboard 路由
	lg := r.Group("/leaderboard")
	{
		lg.GET("/:room_id", leaderboardHandler.GetLeaderboard)
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
