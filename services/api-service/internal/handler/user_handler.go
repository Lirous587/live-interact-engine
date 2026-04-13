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
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags User
// @Accept json
// @Produce json
// @Param request body mapper.RegisterReq true "注册请求"
// @Success 200 {object} mapper.RegisterResp "注册成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /v1/user/register [post]
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
// @Summary 用户登录
// @Description 用户使用邮箱和密码登录，返回 Token
// @Tags User
// @Accept json
// @Produce json
// @Param request body mapper.LoginReq true "登录请求"
// @Param X-Device-ID header string false "设备ID"
// @Success 200 {object} mapper.LoginResp "登录成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "认证失败"
// @Router /v1/user/login [post]
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
// @Summary 获取用户信息
// @Description 根据用户ID获取用户详细信息
// @Tags User
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} mapper.UserResp "用户信息"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 404 {object} map[string]interface{} "用户不存在"
// @Router /v1/user/{user_id} [get]
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
