package middleware

import (
	"net/http"
	"strings"

	"live-interact-engine/services/api-service/internal/utils/ctxutil"
	pb "live-interact-engine/shared/proto/user"
	"live-interact-engine/shared/telemetry"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
		span := trace.SpanFromContext(c.Request.Context())
		accessToken := getAccessToken(c)

		if accessToken == "" {
			if span.IsRecording() {
				span.SetAttributes(
					attribute.Bool("auth.valid", false),
					attribute.String("auth.error", "missing_token"),
				)
			}
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

		validateRes, err := auth.client.ValidateToken(c.Request.Context(), validateReq)
		if err != nil {
			if span.IsRecording() {
				span.SetAttributes(
					attribute.Bool("auth.valid", false),
					attribute.String("auth.error", "validate_rpc_error"),
				)
			}
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
			if span.IsRecording() {
				span.SetAttributes(attribute.Bool("auth.valid", true))
				if validateRes.Payload != nil && validateRes.Payload.UserIdentity != nil {
					span.SetAttributes(attribute.String("auth.user_id", validateRes.Payload.UserIdentity.UserId))
				}
			}
			c.Next()
		case pb.TokenStatus_TOKEN_STATUS_EXPIRED:
			if span.IsRecording() {
				span.SetAttributes(
					attribute.Bool("auth.valid", false),
					attribute.String("auth.error", "token_expired"),
				)
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    4014, // 约定业务状态码：Token 已过期，需要使用 refreshToken
				"message": "Token已过期",
			})
			c.Abort()
		default:
			if span.IsRecording() {
				span.SetAttributes(
					attribute.Bool("auth.valid", false),
					attribute.String("auth.error", "token_invalid"),
				)
			}
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
		ctxutil.SetUserID(c, payload.UserIdentity.UserId)
	}
}
