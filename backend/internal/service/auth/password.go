package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 BCrypt 哈希密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码哈希失败: %w", err)
	}
	return string(bytes), nil
}

// VerifyPassword 校验密码与哈希是否匹配
func VerifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
