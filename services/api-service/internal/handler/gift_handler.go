package handler

import (
	"errors"
	"live-interact-engine/services/api-service/internal/adapter/mapper"
	"live-interact-engine/services/api-service/internal/utils/ctxutil"
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
//	@Security		Bearer
//	@Param			Authorization	header	string	true	"Bearer Token"
//	@Param			body	body		mapper.SendGiftReq	true	"请求体"
//	@Success		200	{object}	mapper.GiftRecordResp
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		401	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
//	@Router			/v1/gift/send [post]
func (h *GiftHandler) SendGift(ctx *gin.Context) {
	var req mapper.SendGiftReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	userID, ok := ctxutil.GetUserID(ctx)
	if !ok || userID == "" {
		response.InvalidParams(ctx, errors.New("user_id not found in auth context"))
		return
	}

	req.SetUserID(userID)

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
//	@Description	获取指定用户的钱包余额（通常获取自己的钱包）
//	@Tags			Wallet
//	@Produce		json
//	@Security		Bearer
//	@Param			Authorization	header	string	true	"Bearer Token"
//	@Success		200	{object}	mapper.WalletResp
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		401	{object}	map[string]interface{}
//	@Failure		404	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
//	@Router			/v1/wallet/balance [get]
func (h *WalletHandler) GetWalletBalance(ctx *gin.Context) {
	var req mapper.GetWalletBalanceReq

	userID, ok := ctxutil.GetUserID(ctx)
	if !ok || userID == "" {
		response.InvalidParams(ctx, errors.New("user_id not found in auth context"))
		return
	}

	req.UserID = userID

	resp, err := h.walletMapper.GetWalletBalance(ctx.Request.Context(), &req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, resp)
}

// Recharge godoc
//
//	@Summary		充值
//	@Description	用户充值钱包余额
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Param			Authorization	header	string	true	"Bearer Token"
//	@Param			body	body		mapper.RechargeReq	true	"请求体"
//	@Success		200	{object}	mapper.RechargeResp
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		401	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
//	@Router			/v1/wallet/recharge [post]
func (h *WalletHandler) Recharge(ctx *gin.Context) {
	var req mapper.RechargeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	userID, ok := ctxutil.GetUserID(ctx)
	if !ok || userID == "" {
		response.InvalidParams(ctx, errors.New("user_id not found in auth context"))
		return
	}

	req.SetUserID(userID)

	resp, err := h.walletMapper.Recharge(ctx.Request.Context(), &req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, resp)
}
