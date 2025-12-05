package kit

import (
	"fmt"
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost BCrypt加密成本
	BcryptCost = bcrypt.DefaultCost
)

// HashPassword 加密密码
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// IsStrongPassword 检查密码强度
// 要求：至少8位，包含字母和数字
func IsStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasLetter := false
	hasDigit := false

	for _, char := range password {
		if unicode.IsLetter(char) {
			hasLetter = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
		if hasLetter && hasDigit {
			return true
		}
	}

	return hasLetter && hasDigit
}

// IsValidPhone 验证手机号格式
// 支持中国大陆手机号格式：1[3-9]\d{9}
func IsValidPhone(phone string) bool {
	// 移除可能的区号前缀（如+86）
	phoneRegex := regexp.MustCompile(`^(\+86)?1[3-9]\d{9}$`)
	return phoneRegex.MatchString(phone)
}
