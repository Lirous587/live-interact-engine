package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"live-interact-engine/services/api-service/internal/adapter/mapper"
	"live-interact-engine/services/api-service/internal/utils/ctxutil"
)

const (
	wsWriteWait  = 10 * time.Second
	wsPingPeriod = 30 * time.Second
)

type DanmakuHandler struct {
	danmakuMapper *mapper.DanmakuMapper
}

func NewDanmakuHandler(danmakuMapper *mapper.DanmakuMapper) *DanmakuHandler {
	return &DanmakuHandler{
		danmakuMapper: danmakuMapper,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	// 生产环境按 Origin 白名单收紧
	CheckOrigin: func(r *http.Request) bool { return true },
}

// wsInMsg Client → Server 消息体
type wsInMsg struct {
	Type            string `json:"type"`              // 目前仅 "send"
	Username        string `json:"username"`          // 发送者用户名
	Content         string `json:"content"`           // 弹幕内容
	DanmakuType     int32  `json:"danmaku_type"`      // 弹幕类型（0=普通 等）
	MentionedUserID string `json:"mentioned_user_id"` // 被@用户 ID（可选）
}

// wsOutMsg Server → Client 消息体
type wsOutMsg struct {
	Type    string              `json:"type"`              // "danmaku" | "error"
	Data    *mapper.DanmakuResp `json:"data,omitempty"`    // 弹幕数据（type="danmaku" 时有值）
	Message string              `json:"message,omitempty"` // 错误信息（type="error" 时有值）
}

// ConnectDanmaku WebSocket 弹幕全双工端点
// @Summary WebSocket 弹幕连接
// @Description 客户端通过 WebSocket 同时收发弹幕；连接建立后立即推送历史消息，随后实时推送新消息。
// @Description 客户端发送: {"type":"send","username":"xxx","content":"xxx","danmaku_type":0}
// @Description 服务端推送: {"type":"danmaku","data":{...}} 或 {"type":"error","message":"..."}
// @Tags Danmaku
// @Param Authorization header string true "Bearer Token"
// @Param room_id query string true "房间ID"
// @Router /v1/danmaku/ws [get]
func (h *DanmakuHandler) ConnectDanmaku(c *gin.Context) {
	roomID := c.Query("room_id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "room_id is required"})
		return
	}

	userID, ok := ctxutil.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.L().Error("[ConnectDanmaku] WebSocket upgrade failed", zap.Error(err))
		return
	}
	defer conn.Close()

	// 派生 context，取消时同步关闭 gRPC 订阅流
	connCtx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	tracer := otel.Tracer("api-service")
	connCtx, span := tracer.Start(connCtx, "ws.danmaku.session",
		trace.WithAttributes(
			attribute.String("room_id", roomID),
			attribute.String("user_id", userID),
		),
	)
	defer span.End()

	danmakuChan, err := h.danmakuMapper.SubscribeDanmaku(connCtx, &mapper.SubscribeDanmakuReq{
		RoomID: roomID,
		UserID: userID,
	})
	if err != nil {
		zap.L().Error("[ConnectDanmaku] subscribe failed", zap.Error(err))
		_ = conn.WriteJSON(wsOutMsg{Type: "error", Message: "subscribe failed: " + err.Error()})
		return
	}

	// writeCh 汇聚来自读循环的错误回写，统一由写 goroutine 发出，避免并发写 WS
	writeCh := make(chan wsOutMsg, 16)
	done := make(chan struct{})

	// 写 goroutine：独占所有 WS 写操作
	go func() {
		defer func() {
			close(done)
			// 强制读循环退出（conn.ReadMessage 会立即返回错误）
			conn.Close()
		}()

		pingTicker := time.NewTicker(wsPingPeriod)
		defer pingTicker.Stop()

		for {
			select {
			case danmaku, ok := <-danmakuChan:
				if !ok {
					// gRPC 订阅流正常结束（订阅者被熔断或服务关闭）。
					// 发送 WS CloseFrame 让客户端感知到优雅关闭，而非 TCP reset。
					zap.L().Debug("[ConnectDanmaku] danmaku channel closed, sending ws close frame",
						zap.String("room_id", roomID))
					_ = conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
					_ = conn.WriteMessage(websocket.CloseMessage,
						websocket.FormatCloseMessage(websocket.CloseNormalClosure, "stream closed"))
					return
				}
				_ = conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
				if err := conn.WriteJSON(wsOutMsg{Type: "danmaku", Data: danmaku}); err != nil {
					zap.L().Debug("[ConnectDanmaku] write danmaku error", zap.Error(err))
					return
				}

			case msg, ok := <-writeCh:
				if !ok {
					return
				}
				_ = conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
				if err := conn.WriteJSON(msg); err != nil {
					zap.L().Debug("[ConnectDanmaku] write msg error", zap.Error(err))
					return
				}

			case <-pingTicker.C:
				_ = conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					zap.L().Debug("[ConnectDanmaku] ping error", zap.Error(err))
					return
				}

			case <-connCtx.Done():
				return
			}
		}
	}()

	// 读循环（主 goroutine）：WS 消息 → gRPC SendDanmaku
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			zap.L().Debug("[ConnectDanmaku] read error, closing session",
				zap.String("room_id", roomID),
				zap.String("user_id", userID),
				zap.Error(err))
			break
		}

		var inMsg wsInMsg
		if jsonErr := json.Unmarshal(raw, &inMsg); jsonErr != nil || inMsg.Type != "send" {
			continue
		}

		sendCtx, sendSpan := tracer.Start(connCtx, "ws.danmaku.send",
			trace.WithAttributes(
				attribute.String("room_id", roomID),
				attribute.String("user_id", userID),
				attribute.Int("content_length", len(inMsg.Content)),
			),
		)
		_, sendErr := h.danmakuMapper.SendDanmaku(sendCtx, &mapper.SendDanmakuReq{
			RoomID:          roomID,
			UserID:          userID,
			Username:        inMsg.Username,
			Content:         inMsg.Content,
			Type:            inMsg.DanmakuType,
			MentionedUserID: inMsg.MentionedUserID,
		})
		sendSpan.End()

		if sendErr != nil {
			select {
			case writeCh <- wsOutMsg{Type: "error", Message: sendErr.Error()}:
			default:
			}
		}
	}

	// 取消 context → 触发 gRPC 流关闭 → danmakuChan 关闭 → 写 goroutine 退出
	cancel()
	<-done
}
