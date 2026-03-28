package router

import (
	"live-interact-engine/services/api-service/internal/handler"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.RouterGroup) {
	// 用户相关
	userGroup := r.Group("/users")
	{
		userGroup.GET("/:id", handler.GetUserHandler)
	}

	// 任务相关
	taskGroup := r.Group("/tasks")
	{
		taskGroup.GET("", handler.CreateTaskHandler)
	}

	// 健康检查
	r.GET("/health", handler.HealthHandler)
}
