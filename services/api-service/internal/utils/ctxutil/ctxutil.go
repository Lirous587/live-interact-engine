package ctxutil

import (
	"github.com/gin-gonic/gin"
)

// 私有常量，防止外部直接使用硬编码 key
const (
	ctxKeyUserID = "x-user-id"
)

func SetUserID(c *gin.Context, userID string) {
	c.Set(ctxKeyUserID, userID)
}

func GetUserID(c *gin.Context) (string, bool) {
	val, exists := c.Get(ctxKeyUserID)
	if !exists {
		return "", false
	}

	userID, ok := val.(string)
	if !ok {
		return "", false
	}
	return userID, true
}
