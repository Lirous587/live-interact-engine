package handler

import (
	"errors"
	"live-interact-engine/services/api-service/internal/adapter/mapper"
	"live-interact-engine/services/api-service/internal/utils/response"
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
	danmakuMapper *mapper.DanmakuMapper
}

func NewDanmakuHandler(danmakuMapper *mapper.DanmakuMapper) *DanmakuHandler {
	return &DanmakuHandler{
		danmakuMapper: danmakuMapper,
	}
}

// SendDanmaku 发送弹幕 API
// @Summary 发送弹幕
// @Description 在房间内发送弹幕消息
// @Tags Danmaku
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Bearer Token"
// @Param request body mapper.SendDanmakuReq true "发送弹幕请求"
// @Success 200 {object} mapper.DanmakuResp "弹幕信息"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /v1/danmaku/send [post]
func (h *DanmakuHandler) SendDanmaku(ctx *gin.Context) {
	var req mapper.SendDanmakuReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	// 使用 mapper 调用 danmaku-service
	resp, err := h.danmakuMapper.SendDanmaku(ctx.Request.Context(), &req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, resp.Danmaku)
}

// SubscribeDanmaku SSE 订阅弹幕
// @Summary 订阅房间弹幕
// @Description 通过 SSE 实时订阅房间内的弹幕消息
// @Tags Danmaku
// @Produce text/event-stream
// @Security Bearer
// @Param Authorization header string true "Bearer Token"
// @Param room_id query string true "房间ID"
// @Param user_id query string true "用户ID"
// @Success 200 {object} mapper.DanmakuResp "实时弹幕流"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /v1/danmaku/subscribe [get]
func (h *DanmakuHandler) SubscribeDanmaku(ctx *gin.Context) {
	var req mapper.SubscribeDanmakuReq
	if err := ctx.BindQuery(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	reqCtx := ctx.Request.Context()
	tracer := otel.Tracer("api-service")

	// 使用 mapper 发起订阅
	danmakuChan, err := h.danmakuMapper.SubscribeDanmaku(reqCtx, &req)
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

			zap.S().Debugf("[SubscribeDanmaku] Sending danmaku to client, id=%s", danmaku.ID)

			// 为每条弹幕创建 child span
			_, childSpan := tracer.Start(reqCtx, "send_danmaku_to_client",
				trace.WithAttributes(
					attribute.String("danmaku_id", danmaku.ID),
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
