package crypto

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordManager 密码处理管理器
type PasswordManager struct {
	cost int
}

// NewPasswordManager 创建密码管理器（cost 决定加密强度，通常 10-12）
func NewPasswordManager(cost int) *PasswordManager {
	if cost < bcrypt.MinCost {
		cost = bcrypt.DefaultCost
	}
	return &PasswordManager{cost: cost}
}

// HashPassword 对密码进行哈希处理
func (pm *PasswordManager) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), pm.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword 验证密码是否匹配哈希值
func (pm *PasswordManager) VerifyPassword(passwordHash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) == nil
}
