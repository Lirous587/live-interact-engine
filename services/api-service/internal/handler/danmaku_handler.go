package handler

import (
	"errors"
	client "live-interact-engine/services/api-service/internal/grpc_clients"
	"live-interact-engine/services/api-service/internal/utils/reskit/response"
	pb "live-interact-engine/shared/proto/danmaku"
	"net/http"
	"time"

	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type DanmakuHandler struct {
	danmakuClient *client.DanmakuClient
}

func NewDanmakuHandler(danmakuClient *client.DanmakuClient) *DanmakuHandler {
	return &DanmakuHandler{
		danmakuClient: danmakuClient,
	}
}

// SendDanmaku 发送弹幕 API
func (h *DanmakuHandler) SendDanmaku(ctx *gin.Context) {
	type SendDanmakuReq struct {
		RoomID          string `json:"room_id" binding:"required"`
		UserID          string `json:"user_id" binding:"required"`
		Username        string `json:"username" binding:"required"`
		Content         string `json:"content" binding:"required,min=1,max=500"`
		Type            int32  `json:"type" binding:"required,gte=0,lte=3"`
		MentionedUserID string `json:"mentioned_user_id"`
	}

	var req SendDanmakuReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	// 调用 danmaku-service
	// 使用 ctx.Request.Context() 来传播链路追踪信息
	resp, err := h.danmakuClient.SendDanmaku(
		ctx.Request.Context(),
		req.RoomID,
		req.UserID,
		req.Username,
		req.Content,
		pb.DanmakuType(req.Type),
		req.MentionedUserID,
	)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, resp.Danmaku)
}

// SubscribeDanmaku SSE 订阅弹幕
func (h *DanmakuHandler) SubscribeDanmaku(ctx *gin.Context) {
	type SubscribeDanmakuReq struct {
		RoomID string `form:"room_id" binding:"required"`
		UserID string `form:"user_id" binding:"required"`
	}
	req := new(SubscribeDanmakuReq)

	if err := ctx.BindQuery(req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	reqCtx := ctx.Request.Context()
	tracer := otel.Tracer("api-service")

	// 用 reqCtx 发起订阅（包含 otelgin 的 root span）
	danmakuChan, err := h.danmakuClient.SubscribeDanmaku(reqCtx, req.RoomID, req.UserID)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// SSE 响应头
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("X-Accel-Buffering", "no")

	flusher, ok := ctx.Writer.(http.Flusher)
	if !ok {
		response.Error(ctx, errors.New("streaming not supported"))
		return
	}

	// 动态续期写超时：每次发送前刷新 deadline
	rc := http.NewResponseController(ctx.Writer)
	const writeDeadlineWindow = 15 * time.Second

	refreshWriteDeadline := func() {
		// 若底层不支持 SetWriteDeadline，这里会返回错误，忽略不影响功能
		_ = rc.SetWriteDeadline(time.Now().Add(writeDeadlineWindow))
	}

	// 先发送一个注释，尽快建立流
	refreshWriteDeadline()
	if _, err := ctx.Writer.WriteString(": connected\n\n"); err != nil {
		return
	}
	flusher.Flush()

	heartbeatTicker := time.NewTicker(10 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-reqCtx.Done():
			zap.L().Debug("[SubscribeDanmaku] Request context done, exiting loop")
			return

		case <-heartbeatTicker.C:
			zap.S().Debug("[SubscribeDanmaku] Heartbeat triggered - still active")
			if _, err := ctx.Writer.WriteString(": ping\n\n"); err != nil {
				zap.L().Error("[SubscribeDanmaku] Heartbeat write error", zap.Error(err))
				return
			}
			flusher.Flush()

			refreshWriteDeadline()
		case danmaku, ok := <-danmakuChan:
			zap.S().Debugf("[SubscribeDanmaku] Received from danmakuChan, ok=%v", ok)
			if !ok {
				zap.S().Debug("[SubscribeDanmaku] Danmaku channel closed, exiting loop")
				return
			}
			if danmaku == nil {
				continue
			}

			zap.S().Debugf("[SubscribeDanmaku] Sending danmaku to client, id=%s", danmaku.Id)

			// 为每条弹幕创建 child span
			_, childSpan := tracer.Start(reqCtx, "send_danmaku_to_client",
				trace.WithAttributes(
					attribute.String("danmaku_id", danmaku.Id),
					attribute.String("room_id", req.RoomID),
					attribute.String("user_id", req.UserID),
					attribute.Int("content_length", len(danmaku.Content)),
				),
			)

			refreshWriteDeadline()
			if err := sse.Encode(ctx.Writer, sse.Event{
				Event: "danmaku",
				Data:  danmaku,
			}); err != nil {
				zap.L().Error("[SubscribeDanmaku] SSE encode failed", zap.Error(err))
				childSpan.RecordError(err)
				childSpan.End()
				return
			}
			childSpan.End()
			flusher.Flush()
		}
	}
}
