package kit

import (
	"math/rand"
	"time"
)

const (
	// PasswordCharacters 密码字符集
	PasswordCharacters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()"
	// PasswordDigits 数字字符集
	PasswordDigits = "0123456789"
	// PasswordLowercase 小写字母字符集
	PasswordLowercase = "abcdefghijklmnopqrstuvwxyz"
	// PasswordUppercase 大写字母字符集
	PasswordUppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// PasswordSpecial 特殊字符集
	PasswordSpecial = "!@#$%^&*()"
	// PasswordLength 密码长度
	PasswordLength = 12
)

var (
	passwordRand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// GenerateRandomPassword 生成12位随机密码
// 确保包含至少1个数字、1个小写字母、1个大写字母、1个特殊字符
// 剩余字符从字符集随机选择，最后打乱字符顺序
func GenerateRandomPassword() string {
	password := make([]byte, PasswordLength)

	// 确保包含至少一个数字
	password[0] = PasswordDigits[passwordRand.Intn(len(PasswordDigits))]
	// 确保包含至少一个小写字母
	password[1] = PasswordLowercase[passwordRand.Intn(len(PasswordLowercase))]
	// 确保包含至少一个大写字母
	password[2] = PasswordUppercase[passwordRand.Intn(len(PasswordUppercase))]
	// 确保包含至少一个特殊符号
	password[3] = PasswordSpecial[passwordRand.Intn(len(PasswordSpecial))]

	// 生成剩余的8个字符
	for i := 4; i < PasswordLength; i++ {
		password[i] = PasswordCharacters[passwordRand.Intn(len(PasswordCharacters))]
	}

	// 打乱密码中字符的顺序
	for i := len(password) - 1; i > 0; i-- {
		j := passwordRand.Intn(i + 1)
		password[i], password[j] = password[j], password[i]
	}

	return string(password)
}

