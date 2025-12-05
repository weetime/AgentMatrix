package kit

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strings"
	"time"
)

const (
	// DeviceCaptchaKeyPrefix 设备验证码Redis Key前缀（与Java版本保持一致）
	DeviceCaptchaKeyPrefix = "sys:device:captcha:"
	// DeviceActivationCodeKeyPrefix 设备激活码Redis Key前缀
	DeviceActivationCodeKeyPrefix = "ota:activation:code:"
	// DeviceActivationDataKeyPrefix 设备激活数据Redis Key前缀
	DeviceActivationDataKeyPrefix = "ota:activation:data:"
	// AgentDeviceCountKeyPrefix 智能体设备数量缓存Key前缀
	AgentDeviceCountKeyPrefix = "agent:device:count:"
	// AgentDeviceLastConnectedKeyPrefix 智能体设备最后连接时间缓存Key前缀
	AgentDeviceLastConnectedKeyPrefix = "agent:device:lastConnected:"
)

// GetDeviceCaptchaKey 获取设备验证码Redis Key
func GetDeviceCaptchaKey(code string) string {
	return DeviceCaptchaKeyPrefix + code
}

// GetDeviceActivationCodeKey 获取设备激活码Redis Key
func GetDeviceActivationCodeKey(code string) string {
	return DeviceActivationCodeKeyPrefix + code
}

// GetDeviceActivationDataKey 获取设备激活数据Redis Key
func GetDeviceActivationDataKey(safeDeviceId string) string {
	return DeviceActivationDataKeyPrefix + safeDeviceId
}

// GetAgentDeviceCountKey 获取智能体设备数量缓存Key
func GetAgentDeviceCountKey(agentId string) string {
	return AgentDeviceCountKeyPrefix + agentId
}

// GetAgentDeviceLastConnectedKey 获取智能体设备最后连接时间缓存Key
func GetAgentDeviceLastConnectedKey(agentId string) string {
	return AgentDeviceLastConnectedKeyPrefix + agentId
}

// GenerateBearerToken 生成Bearer Token
// 格式：SHA256(日期字符串 + MQTT签名密钥)
// 日期格式：yyyy-MM-dd（如：2025-01-15）
func GenerateBearerToken(dateStr, signatureKey string) string {
	tokenContent := dateStr + signatureKey
	hash := sha256.Sum256([]byte(tokenContent))
	return hex.EncodeToString(hash[:])
}

// GenerateBearerTokenForToday 生成今天的Bearer Token
func GenerateBearerTokenForToday(signatureKey string) string {
	dateStr := time.Now().Format("2006-01-02")
	return GenerateBearerToken(dateStr, signatureKey)
}

// GenerateDeviceActivationCode 生成6位设备激活码
func GenerateDeviceActivationCode() string {
	chars := "0123456789"
	code := make([]byte, 6)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		code[i] = chars[n.Int64()]
	}
	return string(code)
}

// GenerateDeviceRegisterCode 生成6位设备注册验证码（与Java的Math.random().substring(2, 8)逻辑类似）
func GenerateDeviceRegisterCode() string {
	// 生成6位随机数字
	code := make([]byte, 6)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		code[i] = byte('0' + n.Int64())
	}
	return string(code)
}

// SafeDeviceID 将设备ID转换为安全的格式（替换冒号为下划线并转小写）
func SafeDeviceID(deviceId string) string {
	result := strings.ReplaceAll(deviceId, ":", "_")
	return strings.ToLower(result)
}
