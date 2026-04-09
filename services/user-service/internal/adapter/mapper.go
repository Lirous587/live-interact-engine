package adapter

import (
	"live-interact-engine/services/user-service/internal/domain"
	pb "live-interact-engine/shared/proto/user"
)

func DomainUserIdentityToProto(identity *domain.UserIdentity) *pb.UserIdentity {
	if identity == nil {
		return nil
	}
	return &pb.UserIdentity{
		UserId: identity.UserID,
	}
}

func ProtoUserIdentityToDomain(identity *pb.UserIdentity) *domain.UserIdentity {
	if identity == nil {
		return nil
	}
	return &domain.UserIdentity{
		UserID:   identity.UserId,
		DeviceID: "", // proto 中没有 DeviceID 字段，暂时设为空
	}
}
