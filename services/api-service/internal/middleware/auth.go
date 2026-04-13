package middleware

import (
	"net/http"
	"strings"

	pb "live-interact-engine/shared/proto/user"
	"live-interact-engine/shared/telemetry"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type authMiddleware struct {
	conn   *grpc.ClientConn
	client pb.TokenServiceClient
}

func NewAuthMiddleware(userServiceAddr string) (*authMiddleware, error) {
	dialOptions := append(
		telemetry.SetupGRPCClientTracing(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	// 连接到 user-service
	conn, err := grpc.NewClient(
		userServiceAddr,
		dialOptions...,
	)
	if err != nil {
		return nil, err
	}

	// 创建 gRPC 客户端
	client := pb.NewTokenServiceClient(conn)

	return &authMiddleware{
		conn:   conn,
		client: client,
	}, nil
}

func getAccessToken(ctx *gin.Context) string {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && parts[0] == "Bearer" {
		return parts[1]
	}
	return authHeader
}

func (auth *authMiddleware) Validate() func(c *gin.Context) {
	return func(c *gin.Context) {
		accessToken := getAccessToken(c)

		if accessToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token为空",
			})
			c.Abort()
			return
		}

		validateReq := &pb.ValidateTokenRequest{
			AccessToken: accessToken,
		}

		validateRes, err := auth.client.ValidateToken(c, validateReq)
		if err != nil {
			zap.L().Error("校验Token GRPC 调用失败", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "服务器内部错误",
			})
			c.Abort()
			return
		}

		switch validateRes.Status {
		case pb.TokenStatus_TOKEN_STATUS_VALID:
			// 只有 Valid 时，放行并注入 payload
			setPayloadToGinCtx(c, validateRes.Payload)
			c.Next()
		case pb.TokenStatus_TOKEN_STATUS_EXPIRED:
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    4014, // 约定业务状态码：Token 已过期，需要使用 refreshToken
				"message": "Token已过期",
			})
			c.Abort()
		default:
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token无效",
			})
			c.Abort()
		}
	}
}

func setPayloadToGinCtx(c *gin.Context, payload *pb.TokenPayload) {
	if payload != nil && payload.UserIdentity != nil {
		c.Set("user_id", payload.UserIdentity.UserId)
	}
}
