package adapter

import (
	"live-interact-engine/services/danmaku-service/internal/domain"
	pb "live-interact-engine/shared/proto/danmaku"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func DanmakuModelToProto(model *domain.DanmakuModel) *pb.Danmaku {
	return &pb.Danmaku{
		Id:              model.ID,
		RoomId:          model.RoomId,
		UserId:          model.UserId,
		Username:        model.Username,
		Content:         model.Content,
		Type:            pb.DanmakuType(model.Type),
		CreatedAt:       timestamppb.New(model.CreatedAt),
		MentionedUserId: model.MentionedUserId,
	}
}

func ProtoToDanmakuModel(danmaku *pb.Danmaku) *domain.DanmakuModel {
	return &domain.DanmakuModel{
		ID:              danmaku.Id,
		RoomId:          danmaku.RoomId,
		UserId:          danmaku.UserId,
		Username:        danmaku.Username,
		Content:         danmaku.Content,
		Type:            domain.DanmakuType(danmaku.Type),
		CreatedAt:       danmaku.CreatedAt.AsTime(),
		MentionedUserId: danmaku.MentionedUserId,
	}
}

func RoomModelToProto(model *domain.RoomModel) *pb.Room {
	return &pb.Room{
		Id:           model.ID,
		HostId:       model.HostId,
		Title:        model.Title,
		State:        pb.RoomState(model.State),
		ViewerCount:  model.ViewerCount,
		BlockedUsers: model.BlockedUsers,
		CreatedAt:    timestamppb.New(model.CreatedAt),
	}
}

func ProtoToRoomModel(danmaku *pb.Room) *domain.RoomModel {
	return &domain.RoomModel{
		ID:           danmaku.Id,
		HostId:       danmaku.HostId,
		Title:        danmaku.Title,
		State:        domain.RoomState(danmaku.State),
		ViewerCount:  danmaku.ViewerCount,
		BlockedUsers: danmaku.BlockedUsers,
		CreatedAt:    danmaku.CreatedAt.AsTime(),
	}
}
