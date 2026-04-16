package mapper

import (
	"context"
	"errors"

	client "live-interact-engine/services/api-service/internal/grpc_clients"
	pb "live-interact-engine/shared/proto/gift"

	"github.com/google/uuid"
)

// ==================== Gift 请求体定义 ====================

// SendGiftReq 发送礼物请求体
type SendGiftReq struct {
	userID   string `json:"-"`
	AnchorID string `json:"anchor_id" binding:"required,uuid_rfc4122"`
	RoomID   string `json:"room_id" binding:"required,uuid_rfc4122"`
	GiftID   string `json:"gift_id" binding:"required,uuid_rfc4122"`
	Amount   int64  `json:"amount" binding:"required,gt=0"`
}

func (s *SendGiftReq) SetUserID(id string) {
	s.userID = id
}

func (s *SendGiftReq) GetUserID() string {
	return s.userID
}

type SendGiftResp struct {
	UserID  string `json:"user_id"`
	Balance int64  `json:"balance"`
}

// ==================== Wallet 请求体定义 ====================

// GetWalletBalanceReq 获取钱包余额请求体
type GetWalletBalanceReq struct {
	UserID string `uri:"user_id" binding:"required,uuid_rfc4122"`
}

// ==================== Leaderboard 请求体定义 ====================

// GetLeaderboardReq 获取排行榜请求体
type GetLeaderboardReq struct {
	RoomID string `uri:"room_id" binding:"required,uuid_rfc4122"`
	TopN   int32  `form:"top_n" binding:"omitempty,gte=1,lte=1000"`
}

// ==================== Gift 响应体定义 ====================

// GiftRecordResp 礼物记录响应体
type GiftRecordResp struct {
	IdempotencyKey string `json:"idempotency_key"`
	UserID         string `json:"user_id"`
	AnchorID       string `json:"anchor_id"`
	RoomID         string `json:"room_id"`
	GiftID         string `json:"gift_id"`
	Amount         int64  `json:"amount"`
	Status         string `json:"status"` // pending, success, failed
	CreatedAt      int64  `json:"created_at"`
	UpdatedAt      int64  `json:"updated_at"`
}

// GiftResp 礼物信息响应体
type GiftResp struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url"`
	CacheKey    string `json:"cache_key"`
	Price       int64  `json:"price"`
	VIPOnly     bool   `json:"vip_only"`
	Status      string `json:"status"` // online, offline, limited_time
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// ListGiftsResp 礼物列表响应体
type ListGiftsResp struct {
	Gifts []*GiftResp `json:"gifts"`
}

// ==================== Wallet 响应体定义 ====================

// WalletResp 钱包信息响应体
type WalletResp struct {
	UserID        string `json:"user_id"`
	Balance       int64  `json:"balance"`
	VersionNumber int64  `json:"version_number"` // 乐观锁版本号
	CreatedAt     int64  `json:"created_at"`
	UpdatedAt     int64  `json:"updated_at"`
}

// ==================== Leaderboard 响应体定义 ====================

// LeaderboardEntryResp 排行榜条目响应体
type LeaderboardEntryResp struct {
	Rank   int32  `json:"rank"`
	UserID string `json:"user_id"`
	Score  int64  `json:"score"` // 累计送礼金额
}

// LeaderboardResp 排行榜响应体
type LeaderboardResp struct {
	RoomID  string                  `json:"room_id"`
	Entries []*LeaderboardEntryResp `json:"entries"`
}

// ==================== Mapper 类定义 ====================

// GiftMapper 礼物业务适配层（gRPC 客户端 + 业务转换）
type GiftMapper struct {
	giftClient *client.GiftClient
}

// NewGiftMapper 创建礼物 mapper
func NewGiftMapper(giftClient *client.GiftClient) *GiftMapper {
	return &GiftMapper{
		giftClient: giftClient,
	}
}

// WalletMapper 钱包业务适配层（gRPC 客户端 + 业务转换）
type WalletMapper struct {
	giftClient *client.GiftClient
}

// NewWalletMapper 创建钱包 mapper
func NewWalletMapper(giftClient *client.GiftClient) *WalletMapper {
	return &WalletMapper{
		giftClient: giftClient,
	}
}

// LeaderboardMapper 排行榜业务适配层（gRPC 客户端 + 业务转换）
type LeaderboardMapper struct {
	giftClient *client.GiftClient
}

// NewLeaderboardMapper 创建排行榜 mapper
func NewLeaderboardMapper(giftClient *client.GiftClient) *LeaderboardMapper {
	return &LeaderboardMapper{
		giftClient: giftClient,
	}
}

// ==================== Gift 业务方法 ====================

// SendGift 发送礼物
func (m *GiftMapper) SendGift(ctx context.Context, req *SendGiftReq) (*SendGiftResp, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	uuid, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	// 构造 pb.SendGiftRequest
	pbReq := &pb.SendGiftRequest{
		IdempotencyKey: uuid.String(),
		UserId:         req.GetUserID(),
		AnchorId:       req.AnchorID,
		RoomId:         req.RoomID,
		GiftId:         req.GiftID,
		Amount:         req.Amount,
	}

	// 调用 gRPC
	pbResp, err := m.giftClient.SendGift(ctx, pbReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	return &SendGiftResp{
		UserID:  pbResp.UserId,
		Balance: pbResp.Balance,
	}, nil
}

// ListGifts 获取礼物列表
func (m *GiftMapper) ListGifts(ctx context.Context) (*ListGiftsResp, error) {
	// 调用 gRPC
	pbGifts, err := m.giftClient.ListGifts(ctx)
	if err != nil {
		return nil, err
	}

	// 转换响应
	gifts := make([]*GiftResp, len(pbGifts))
	for i, pbGift := range pbGifts {
		gifts[i] = pbGiftToResp(pbGift)
	}

	return &ListGiftsResp{
		Gifts: gifts,
	}, nil
}

// ==================== Wallet 业务方法 ====================

// GetWalletBalance 获取钱包余额
func (m *WalletMapper) GetWalletBalance(ctx context.Context, req *GetWalletBalanceReq) (*WalletResp, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	if req.UserID == "" {
		return nil, errors.New("user_id is required")
	}

	// 调用 gRPC
	pbWallet, err := m.giftClient.GetWalletBalance(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	if pbWallet == nil {
		return nil, errors.New("wallet not found")
	}

	// 转换响应
	return &WalletResp{
		UserID:        pbWallet.UserId,
		Balance:       pbWallet.Balance,
		VersionNumber: pbWallet.VersionNumber,
		CreatedAt:     pbWallet.CreatedAt,
		UpdatedAt:     pbWallet.UpdatedAt,
	}, nil
}

// ==================== Leaderboard 业务方法 ====================

// GetLeaderboard 获取排行榜
func (m *LeaderboardMapper) GetLeaderboard(ctx context.Context, req *GetLeaderboardReq) (*LeaderboardResp, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	if req.RoomID == "" {
		return nil, errors.New("room_id is required")
	}

	// 默认返回 Top 100
	topN := req.TopN
	if topN <= 0 {
		topN = 100
	}

	// 调用 gRPC
	pbResp, err := m.giftClient.GetLeaderboard(ctx, req.RoomID, topN)
	if err != nil {
		return nil, err
	}

	if pbResp == nil {
		return nil, errors.New("leaderboard response is nil")
	}

	// 转换响应
	entries := make([]*LeaderboardEntryResp, len(pbResp.Entries))
	for i, pbEntry := range pbResp.Entries {
		entries[i] = &LeaderboardEntryResp{
			Rank:   pbEntry.Rank,
			UserID: pbEntry.UserId,
			Score:  pbEntry.Score,
		}
	}

	return &LeaderboardResp{
		RoomID:  pbResp.RoomId,
		Entries: entries,
	}, nil
}

// ==================== 转换函数 ====================

// pbGiftToResp 将 pb.Gift 转换为 GiftResp
func pbGiftToResp(pbGift *pb.Gift) *GiftResp {
	if pbGift == nil {
		return nil
	}

	return &GiftResp{
		ID:          pbGift.Id,
		Name:        pbGift.Name,
		Description: pbGift.Description,
		IconURL:     pbGift.IconUrl,
		CacheKey:    pbGift.CacheKey,
		Price:       pbGift.Price,
		VIPOnly:     pbGift.VipOnly,
		Status:      pbGift.Status.String(),
		CreatedAt:   pbGift.CreatedAt,
		UpdatedAt:   pbGift.UpdatedAt,
	}
}
