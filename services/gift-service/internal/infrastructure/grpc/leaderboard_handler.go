package grpc

import (
	"context"

	pb "live-interact-engine/shared/proto/gift"
	"live-interact-engine/shared/svcerr"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

// ==================== LeaderboardService Handler ====================

type LeaderboardHandler struct {
	pb.UnimplementedLeaderboardServiceServer
	// TODO: 注入 LeaderboardService 当实现时
}

// NewLeaderboardHandler 创建 LeaderboardHandler 实例
func NewLeaderboardHandler() *LeaderboardHandler {
	return &LeaderboardHandler{}
}

// GetLeaderboard 获取排行榜（按送礼金额排序）
func (h *LeaderboardHandler) GetLeaderboard(ctx context.Context, req *pb.GetLeaderboardRequest) (*pb.GetLeaderboardResponse, error) {
	span := trace.SpanFromContext(ctx)

	_, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	topN := req.TopN
	if topN <= 0 || topN > 1000 {
		topN = 100 // 默认返回 Top 100
	}

	// TODO: 从 Redis ZSET 获取排行榜数据
	// entries, err := h.leaderboardService.GetLeaderboard(ctx, roomID, int(topN))

	return &pb.GetLeaderboardResponse{
		RoomId:  req.RoomId,
		Entries: []*pb.LeaderboardEntry{}, // 临时返回空列表
	}, nil
}
