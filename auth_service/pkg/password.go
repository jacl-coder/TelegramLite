package pkg

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordManager 密码管理器
type PasswordManager struct {
	cost int
}

// NewPasswordManager 创建密码管理器
func NewPasswordManager() *PasswordManager {
	return &PasswordManager{
		cost: bcrypt.DefaultCost,
	}
}

// HashPassword 加密密码
func (pm *PasswordManager) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), pm.cost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword 验证密码
func (pm *PasswordManager) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// IsValidPassword 检查密码强度
func (pm *PasswordManager) IsValidPassword(password string) bool {
	// 简单的密码验证规则
	if len(password) < 6 {
		return false
	}
	if len(password) > 128 {
		return false
	}
	return true
}
