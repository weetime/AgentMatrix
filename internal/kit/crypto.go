package kit

import (
	"crypto/aes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

// AESEncrypt AES加密 (AES/ECB/PKCS5Padding)
// key: 密钥（16位、24位或32位）
// plainText: 待加密字符串
// 返回: 加密后的Base64字符串
func AESEncrypt(key, plainText string) (string, error) {
	// 确保密钥长度为16、24或32位
	keyBytes := padKey([]byte(key))

	// 创建AES cipher
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("AES加密失败: %w", err)
	}

	// PKCS5Padding填充
	plainBytes := []byte(plainText)
	plainBytes = pkcs5Padding(plainBytes, block.BlockSize())

	// ECB模式加密
	cipherBytes := make([]byte, len(plainBytes))
	for i := 0; i < len(plainBytes); i += block.BlockSize() {
		block.Encrypt(cipherBytes[i:i+block.BlockSize()], plainBytes[i:i+block.BlockSize()])
	}

	// Base64编码
	return base64.StdEncoding.EncodeToString(cipherBytes), nil
}

// padKey 填充密钥到指定长度（16、24或32位）
func padKey(keyBytes []byte) []byte {
	keyLength := len(keyBytes)
	if keyLength == 16 || keyLength == 24 || keyLength == 32 {
		return keyBytes
	}

	// 如果密钥长度不足，用0填充；如果超过，截取前32位
	paddedKey := make([]byte, 32)
	copy(paddedKey, keyBytes)
	return paddedKey
}

// pkcs5Padding PKCS5填充
func pkcs5Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

// MD5HexDigest 使用MD5进行加密，返回十六进制字符串
func MD5HexDigest(text string) string {
	hash := md5.Sum([]byte(text))
	return fmt.Sprintf("%x", hash)
}

// GenerateWebSocketToken 生成WebSocket认证token
// 遵循Python端AuthManager的实现逻辑：token = signature.timestamp
// clientId: 客户端ID
// username: 用户名 (通常为deviceId/macAddress)
// secretKey: 密钥
// 返回: 认证token字符串 (格式: signature.timestamp)
func GenerateWebSocketToken(clientId, username, secretKey string) (string, error) {
	if secretKey == "" {
		return "", fmt.Errorf("WebSocket认证密钥未配置(server.secret)")
	}

	// 获取当前时间戳(秒)
	timestamp := time.Now().Unix()

	// 构建签名内容: clientId|username|timestamp
	content := fmt.Sprintf("%s|%s|%d", clientId, username, timestamp)

	// 生成HMAC-SHA256签名
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(content))
	signature := h.Sum(nil)

	// Base64 URL-safe编码签名(去除填充符=)
	signatureBase64 := base64.RawURLEncoding.EncodeToString(signature)

	// 返回格式: signature.timestamp
	return fmt.Sprintf("%s.%d", signatureBase64, timestamp), nil
}
