package kit

import (
	"crypto/aes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
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

