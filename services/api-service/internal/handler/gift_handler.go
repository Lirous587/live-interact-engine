package handler

import (
	"live-interact-engine/services/api-service/internal/adapter/mapper"
	"live-interact-engine/services/api-service/internal/utils/response"

	"github.com/gin-gonic/gin"
)

// ==================== Gift Handler ====================

// GiftHandler 礼物业务处理器
type GiftHandler struct {
	giftMapper *mapper.GiftMapper
}

// NewGiftHandler 创建礼物处理器
func NewGiftHandler(giftMapper *mapper.GiftMapper) *GiftHandler {
	return &GiftHandler{
		giftMapper: giftMapper,
	}
}

// SendGift godoc
//
//	@Summary		发送礼物
//	@Description	用户向主播发送礼物，扣除用户钱包余额，生成礼物赠送记录
//	@Tags			Gifts
//	@Accept			json
//	@Produce		json
//	@Param			body	body		mapper.SendGiftReq	true	"请求体"
//	@Success		200	{object}	mapper.GiftRecordResp
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
//	@Router			/v1/gift/send [post]
func (h *GiftHandler) SendGift(ctx *gin.Context) {
	var req mapper.SendGiftReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	resp, err := h.giftMapper.SendGift(ctx.Request.Context(), &req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, resp)
}

// ListGifts godoc
//
//	@Summary		获取礼物列表
//	@Description	获取平台全量礼物列表
//	@Tags			Gifts
//	@Produce		json
//	@Success		200	{object}	mapper.ListGiftsResp
//	@Failure		500	{object}	map[string]interface{}
//	@Router			/v1/gift/list [get]
func (h *GiftHandler) ListGifts(ctx *gin.Context) {
	resp, err := h.giftMapper.ListGifts(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, resp)
}

// ==================== Wallet Handler ====================

// WalletHandler 钱包业务处理器
type WalletHandler struct {
	walletMapper *mapper.WalletMapper
}

// NewWalletHandler 创建钱包处理器
func NewWalletHandler(walletMapper *mapper.WalletMapper) *WalletHandler {
	return &WalletHandler{
		walletMapper: walletMapper,
	}
}

// GetWalletBalance godoc
//
//	@Summary		获取钱包余额
//	@Description	获取指定用户的钱包余额
//	@Tags			Wallet
//	@Produce		json
//	@Param			user_id	path		string	true	"用户 ID (UUID)"
//	@Success		200	{object}	mapper.WalletResp
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		404	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
//	@Router			/v1/wallet/{user_id}/balance [get]
func (h *WalletHandler) GetWalletBalance(ctx *gin.Context) {
	var req mapper.GetWalletBalanceReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	resp, err := h.walletMapper.GetWalletBalance(ctx.Request.Context(), &req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, resp)
}

// ==================== Leaderboard Handler ====================

// LeaderboardHandler 排行榜业务处理器
type LeaderboardHandler struct {
	leaderboardMapper *mapper.LeaderboardMapper
}

// NewLeaderboardHandler 创建排行榜处理器
func NewLeaderboardHandler(leaderboardMapper *mapper.LeaderboardMapper) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderboardMapper: leaderboardMapper,
	}
}

// GetLeaderboard godoc
//
//	@Summary		获取房间排行榜
//	@Description	获取指定房间的礼物赠送排行榜 (按累计送礼金额降序)
//	@Tags			Leaderboard
//	@Produce		json
//	@Param			room_id	path		string	true	"房间 ID (UUID)"
//	@Param			top_n	query		integer	false	"返回前 N 名 (默认 100, 最大 1000)"	default(100)
//	@Success		200	{object}	mapper.LeaderboardResp
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
//	@Router			/v1/leaderboard/{room_id} [get]
func (h *LeaderboardHandler) GetLeaderboard(ctx *gin.Context) {
	var req mapper.GetLeaderboardReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	// 从查询参数绑定 TopN
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	resp, err := h.leaderboardMapper.GetLeaderboard(ctx.Request.Context(), &req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, resp)
}
