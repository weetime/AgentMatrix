package kit

import (
	"context"
	"fmt"
	"time"

	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
)

const (
	// CaptchaKeyPrefix 验证码Redis Key前缀（与Java版本保持一致）
	CaptchaKeyPrefix = "sys:captcha:"
	// CaptchaExpiration 验证码过期时间（5分钟）
	CaptchaExpiration = 5 * time.Minute
	// CaptchaLength 验证码长度（5位数字）
	CaptchaLength = 5
)

// GenerateCaptcha 生成图形验证码
// 返回: (base64Image, uuid, error)
func GenerateCaptcha(ctx context.Context, redisClient *redis.Client, uuid string) (string, string, error) {
	// 生成5位数字验证码
	driver := base64Captcha.NewDriverDigit(60, 200, CaptchaLength, 0.3, 20)
	captcha := base64Captcha.NewCaptcha(driver, base64Captcha.DefaultMemStore)

	// 生成验证码
	id, b64s, answer, err := captcha.Generate()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate captcha: %w", err)
	}

	// 使用传入的uuid作为key，如果没有则使用生成的id
	captchaKey := uuid
	if captchaKey == "" {
		captchaKey = id
	}

	// 存储验证码到Redis
	key := CaptchaKeyPrefix + captchaKey
	if err := redisClient.Set(ctx, key, answer, CaptchaExpiration).Err(); err != nil {
		return "", "", fmt.Errorf("failed to save captcha to redis: %w", err)
	}

	return b64s, captchaKey, nil
}

// ValidateCaptcha 验证图形验证码
// uuid: 验证码ID
// code: 用户输入的验证码
// delete: 验证成功后是否删除验证码
func ValidateCaptcha(ctx context.Context, redisClient *redis.Client, uuid, code string, delete bool) bool {
	if code == "" {
		return false
	}

	// 从Redis获取验证码
	key := CaptchaKeyPrefix + uuid
	storedCode, err := redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return false
	}
	if err != nil {
		return false
	}

	// 验证验证码（不区分大小写）
	valid := storedCode != "" && (storedCode == code || storedCode == code)
	if !valid {
		return false
	}

	// 如果验证成功且需要删除，则删除验证码
	if delete {
		_ = redisClient.Del(ctx, key)
	}

	return true
}
