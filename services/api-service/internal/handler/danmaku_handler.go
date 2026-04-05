package handler

import (
	"errors"
	client "live-interact-engine/services/api-service/internal/grpc_clients"
	"live-interact-engine/services/api-service/internal/utils/reskit/codes"
	"live-interact-engine/services/api-service/internal/utils/reskit/response"
	pb "live-interact-engine/shared/proto/danmaku"
	"net/http"

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

// SubscribeDanmaku WebSocket 订阅弹幕
func (h *DanmakuHandler) SubscribeDanmaku(ctx *gin.Context) {
	roomID := ctx.Query("room_id")
	if roomID == "" {
		response.Error(ctx, codes.ErrDanmakuNeedRoomID)
		return
	}

	// 获取 danmaku channel
	danmakuChan, err := h.danmakuClient.SubscribeDanmaku(ctx.Request.Context(), roomID)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 通过 Server-Sent Events (SSE) 推送弹幕
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	flusher, ok := ctx.Writer.(http.Flusher)
	if !ok {
		response.Error(ctx, errors.New("streaming not supported"))
		return
	}

	for danmaku := range danmakuChan {
		// 发送 SSE 格式
		ctx.Writer.WriteString("data: " + danmaku.String() + "\n\n")
		flusher.Flush()
	}
}
