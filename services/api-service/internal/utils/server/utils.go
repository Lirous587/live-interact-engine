package server

import (
	"live-interact-engine/services/api-service/internal/utils/reskit/codes"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"

func GetUserID(ctx *gin.Context) (int64, error) {
	uidStr, exist := ctx.Get(UserIDKey)
	if !exist {
		return 0, codes.ErrUserNotFound
	}

	userID, ok := uidStr.(int64)
	if !ok {
		return 0, codes.ErrUserIDInvalid
	}

	return userID, nil
}
