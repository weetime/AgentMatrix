package kit

import (
	"crypto/md5"
	"fmt"

	"github.com/google/uuid"
)

const (
	// TokenExpireSeconds Token过期时间（12小时）
	TokenExpireSeconds = 3600 * 12
)

// GenerateToken 生成Token（MD5(UUID)）
// 返回32位十六进制字符串
func GenerateToken() string {
	// 生成UUID
	id := uuid.New().String()
	// 计算MD5
	hash := md5.Sum([]byte(id))
	// 转换为十六进制字符串
	return fmt.Sprintf("%x", hash)
}

// toHexString 将字节数组转换为十六进制字符串
func toHexString(data []byte) string {
	if data == nil {
		return ""
	}
	const hexCode = "0123456789abcdef"
	result := make([]byte, len(data)*2)
	for i, b := range data {
		result[i*2] = hexCode[(b>>4)&0xF]
		result[i*2+1] = hexCode[b&0xF]
	}
	return string(result)
}
