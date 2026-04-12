package adapter

import (
	"live-interact-engine/services/user-service/internal/domain"
	pb "live-interact-engine/shared/proto/user"

	"github.com/google/uuid"
)

func DomainUserIdentityToProto(identity *domain.UserIdentity) *pb.UserIdentity {
	if identity == nil {
		return nil
	}
	return &pb.UserIdentity{
		UserId: identity.UserID.String(),
	}
}

func ProtoUserIdentityToDomain(identity *pb.UserIdentity) *domain.UserIdentity {
	if identity == nil {
		return nil
	}
	return &domain.UserIdentity{
		UserID: uuid.MustParse(identity.UserId),
		UserIdentityMetadata: domain.UserIdentityMetadata{
			DeviceID: identity.Metadata.DeviceId,
		},
	}
}
