package handler

import (
	"live-interact-engine/services/api-service/internal/adapter/mapper"
	"live-interact-engine/services/api-service/internal/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type UserHandler struct {
	userMapper *mapper.UserMapper
}

func NewUserHandler(userMapper *mapper.UserMapper) *UserHandler {
	return &UserHandler{
		userMapper: userMapper,
	}
}

// Register 用户注册 API
func (h *UserHandler) Register(ctx *gin.Context) {
	var req mapper.RegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	// 调用 mapper
	resp, err := h.userMapper.Register(ctx.Request.Context(), &req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, resp)
}

// Login 用户登录 API
func (h *UserHandler) Login(ctx *gin.Context) {
	var req mapper.LoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	// 从请求头获取 device_id
	deviceID := ctx.GetHeader("X-Device-ID")
	if deviceID == "" {
		deviceID = "default"
	}
	req.DeviceID = deviceID

	// 调用 mapper
	resp, err := h.userMapper.Login(ctx.Request.Context(), &req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, resp)
}

// GetUser 获取用户信息 API
func (h *UserHandler) GetUser(ctx *gin.Context) {
	userID := ctx.Param("user_id")
	if userID == "" {
		response.InvalidParams(ctx, errors.New("user_id required"))
		return
	}

	req := &mapper.GetUserReq{UserID: userID}

	// 调用 mapper
	resp, err := h.userMapper.GetUser(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, resp)
}
