package handler

import (
	"errors"
	client "live-interact-engine/services/api-service/internal/grpc_clients"
	"live-interact-engine/services/api-service/internal/utils/reskit/apicodes"
	"live-interact-engine/services/api-service/internal/utils/reskit/response"
	pb "live-interact-engine/shared/proto/danmaku"
	"net/http"
	"time"

	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
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
		Content         string `json:"content" binding:"required"`
		Type            int32  `json:"type"`
		MentionedUserID string `json:"mentioned_user_id"`
	}

	var req SendDanmakuReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, err)
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
	roomID := ctx.Query("room_id")
	if roomID == "" {
		response.Error(ctx, apicodes.ErrDanmakuNeedRoomID)
		return
	}

	reqCtx := ctx.Request.Context()

	// 获取 danmaku channel
	danmakuChan, err := h.danmakuClient.SubscribeDanmaku(ctx.Request.Context(), roomID)
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
	const writeDeadlineWindow = 30 * time.Second

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
			return

		case <-heartbeatTicker.C:
			// SSE 注释心跳，维持链路活性
			refreshWriteDeadline()
			if _, err := ctx.Writer.WriteString(": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()

		case danmaku, ok := <-danmakuChan:
			if !ok {
				return
			}
			if danmaku == nil {
				continue
			}

			refreshWriteDeadline()
			if err := sse.Encode(ctx.Writer, sse.Event{
				Event: "danmaku",
				Data:  danmaku,
			}); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
