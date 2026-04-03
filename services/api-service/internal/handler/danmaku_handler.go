package handler

import (
	"context"
	client "live-interact-engine/services/api-service/internal/grpc_clients"
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
func (h *DanmakuHandler) SendDanmaku(c *gin.Context) {
	type SendDanmakuReq struct {
		RoomID          string `json:"room_id" binding:"required"`
		UserID          string `json:"user_id" binding:"required"`
		Username        string `json:"username" binding:"required"`
		Content         string `json:"content" binding:"required"`
		Type            int32  `json:"type"`
		MentionedUserID string `json:"mentioned_user_id"`
	}

	var req SendDanmakuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用 danmaku-service
	resp, err := h.danmakuClient.SendDanmaku(
		context.Background(),
		req.RoomID,
		req.UserID,
		req.Username,
		req.Content,
		pb.DanmakuType(req.Type),
		req.MentionedUserID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    resp.Danmaku,
		"message": resp.Message,
	})
}

// SubscribeDanmaku WebSocket 订阅弹幕
func (h *DanmakuHandler) SubscribeDanmaku(c *gin.Context) {
	roomID := c.Query("room_id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_id required"})
		return
	}

	// 获取 danmaku channel
	danmakuChan, err := h.danmakuClient.SubscribeDanmaku(c.Request.Context(), roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 通过 Server-Sent Events (SSE) 推送弹幕
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	for danmaku := range danmakuChan {
		// 发送 SSE 格式
		c.Writer.WriteString("data: " + danmaku.String() + "\n\n")
		flusher.Flush()
	}
}
