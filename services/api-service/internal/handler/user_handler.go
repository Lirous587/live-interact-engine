package handler

import (
	client "live-interact-engine/services/api-service/internal/grpc_clients"
	"live-interact-engine/services/api-service/internal/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type UserHandler struct {
	userClient *client.UserClient
	logger     *zap.Logger
}

func NewUserHandler(userClient *client.UserClient) *UserHandler {
	logger, _ := zap.NewProduction()
	return &UserHandler{
		userClient: userClient,
		logger:     logger,
	}
}

// Register 用户注册 API
func (h *UserHandler) Register(ctx *gin.Context) {
	type RegisterReq struct {
		Username string `json:"username" binding:"required,min=1,max=50"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6,max=100"`
	}

	var req RegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	// 调用 user-service
	resp, err := h.userClient.Register(ctx.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		h.logger.Error("failed to register user", zap.Error(err), zap.String("email", req.Email))
		response.Error(ctx, err)
		return
	}

	type RegisterResp struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	response.Success(ctx, RegisterResp{
		UserID:   resp.User.UserId,
		Username: resp.User.Username,
		Email:    resp.User.Email,
	})
}

// Login 用户登录 API
func (h *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6,max=100"`
	}

	var req LoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	// 从请求头获取 device_id
	deviceID := ctx.GetHeader("X-Device-ID")
	if deviceID == "" {
		deviceID = "default"
	}

	// 调用 user-service
	resp, err := h.userClient.Login(ctx.Request.Context(), req.Email, req.Password, deviceID)
	if err != nil {
		h.logger.Error("failed to login user", zap.Error(err), zap.String("email", req.Email))
		response.Error(ctx, err)
		return
	}

	type LoginResp struct {
		UserID       string `json:"user_id"`
		Username     string `json:"username"`
		Email        string `json:"email"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	response.Success(ctx, LoginResp{
		UserID:       resp.User.UserId,
		Username:     resp.User.Username,
		Email:        resp.User.Email,
		AccessToken:  resp.TokenPair.AccessToken,
		RefreshToken: resp.TokenPair.RefreshToken,
	})
}

// GetUser 获取用户信息 API
func (h *UserHandler) GetUser(ctx *gin.Context) {
	userID := ctx.Param("user_id")
	if userID == "" {
		response.InvalidParams(ctx, errors.New("user_id required"))
		return
	}

	// 调用 user-service
	user, err := h.userClient.GetUser(ctx.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get user info", zap.Error(err), zap.String("user_id", userID))
		response.Error(ctx, err)
		return
	}

	type UserResp struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		IsActive bool   `json:"is_active"`
	}

	response.Success(ctx, UserResp{
		UserID:   user.UserId,
		Username: user.Username,
		Email:    user.Email,
		IsActive: user.IsActive,
	})
}
