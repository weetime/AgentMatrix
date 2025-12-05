package kit

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	"github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/redis/go-redis/v9"
)

const (
	// SMSValidateCodeKeyPrefix 短信验证码Redis Key前缀
	SMSValidateCodeKeyPrefix = "sms:Validate:Code:"
	// SMSLastSendTimeKeySuffix 最后发送时间Key后缀
	SMSLastSendTimeKeySuffix = ":last_send_time"
	// SMSTodayCountKeySuffix 今日发送次数Key后缀
	SMSTodayCountKeySuffix = ":today_count"
	// SMSExpiration 短信验证码过期时间（5分钟）
	SMSExpiration = 5 * time.Minute
	// SMSMinInterval 短信发送最小间隔（60秒）
	SMSMinInterval = 60 * time.Second
	// SMSDefaultMaxSendCount 默认每日最大发送次数
	SMSDefaultMaxSendCount = 5
	// SMSValidateCodeLength 短信验证码长度（6位数字）
	SMSValidateCodeLength = 6
)

// SMSConfig 短信配置参数代码
const (
	SMSParamAccessKeyID          = "aliyun.sms.access_key_id"
	SMSParamAccessKeySecret      = "aliyun.sms.access_key_secret"
	SMSParamSignName             = "aliyun.sms.sign_name"
	SMSParamTemplateCode         = "aliyun.sms.sms_code_template_code"
	SMSParamMaxSendCount         = "server.sms_max_send_count"
	SMSParamEnableMobileRegister = "server.enable_mobile_register"
)

// SendSMSVerificationCode 发送短信验证码
func SendSMSVerificationCode(
	ctx context.Context,
	phone string,
	redisClient *redis.Client,
	paramsService ParamsService,
) error {
	// 检查发送间隔
	lastSendTimeKey := SMSValidateCodeKeyPrefix + phone + SMSLastSendTimeKeySuffix
	lastSendTimeStr, err := redisClient.Get(ctx, lastSendTimeKey).Result()
	if err == nil && lastSendTimeStr != "" {
		lastSendTime, err := strconv.ParseInt(lastSendTimeStr, 10, 64)
		if err == nil {
			currentTime := time.Now().UnixMilli()
			timeDiff := currentTime - lastSendTime
			if timeDiff < SMSMinInterval.Milliseconds() {
				remainingSeconds := (SMSMinInterval.Milliseconds() - timeDiff) / 1000
				return fmt.Errorf("短信发送过于频繁，请%d秒后再试", remainingSeconds)
			}
		}
	}

	// 检查今日发送次数
	todayCountKey := SMSValidateCodeKeyPrefix + phone + SMSTodayCountKeySuffix
	todayCountStr, _ := redisClient.Get(ctx, todayCountKey).Result()
	todayCount := 0
	if todayCountStr != "" {
		todayCount, _ = strconv.Atoi(todayCountStr)
	}

	// 获取最大发送次数限制
	maxSendCount := SMSDefaultMaxSendCount
	maxSendCountStr, err := paramsService.GetValue(SMSParamMaxSendCount, true)
	if err == nil && maxSendCountStr != "" {
		if count, err := strconv.Atoi(maxSendCountStr); err == nil {
			maxSendCount = count
		}
	}

	if todayCount >= maxSendCount {
		return fmt.Errorf("今日短信发送次数已达上限")
	}

	// 生成6位数字验证码
	validateCode := generateValidateCode(SMSValidateCodeLength)

	// 存储验证码到Redis
	codeKey := SMSValidateCodeKeyPrefix + phone
	if err := redisClient.Set(ctx, codeKey, validateCode, SMSExpiration).Err(); err != nil {
		return fmt.Errorf("failed to save SMS code to redis: %w", err)
	}

	// 更新最后发送时间（60秒过期）
	if err := redisClient.Set(ctx, lastSendTimeKey, time.Now().UnixMilli(), SMSMinInterval).Err(); err != nil {
		return fmt.Errorf("failed to save last send time: %w", err)
	}

	// 更新今日发送次数
	if todayCount == 0 {
		// 设置过期时间为当天剩余时间
		now := time.Now()
		midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		expiration := midnight.Sub(now)
		if err := redisClient.Set(ctx, todayCountKey, "1", expiration).Err(); err != nil {
			return fmt.Errorf("failed to save today count: %w", err)
		}
	} else {
		if err := redisClient.Incr(ctx, todayCountKey).Err(); err != nil {
			return fmt.Errorf("failed to increment today count: %w", err)
		}
	}

	// 发送短信
	if err := sendSMS(ctx, phone, validateCode, paramsService); err != nil {
		// 如果发送失败，回退今日发送次数
		_ = redisClient.Decr(ctx, todayCountKey)
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	return nil
}

// ValidateSMSVerificationCode 验证短信验证码
func ValidateSMSVerificationCode(
	ctx context.Context,
	redisClient *redis.Client,
	phone, code string,
	delete bool,
) bool {
	if code == "" {
		return false
	}

	// 从Redis获取验证码
	key := SMSValidateCodeKeyPrefix + phone
	storedCode, err := redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return false
	}
	if err != nil {
		return false
	}

	// 验证验证码
	valid := storedCode != "" && storedCode == code
	if !valid {
		return false
	}

	// 如果验证成功且需要删除，则删除验证码
	if delete {
		_ = redisClient.Del(ctx, key)
	}

	return true
}

// sendSMS 发送短信
func sendSMS(ctx context.Context, phone, code string, paramsService ParamsService) error {
	// 获取配置
	accessKeyID, err := paramsService.GetValue(SMSParamAccessKeyID, true)
	if err != nil || accessKeyID == "" {
		return fmt.Errorf("SMS access key ID not configured")
	}

	accessKeySecret, err := paramsService.GetValue(SMSParamAccessKeySecret, true)
	if err != nil || accessKeySecret == "" {
		return fmt.Errorf("SMS access key secret not configured")
	}

	signName, err := paramsService.GetValue(SMSParamSignName, true)
	if err != nil || signName == "" {
		return fmt.Errorf("SMS sign name not configured")
	}

	templateCode, err := paramsService.GetValue(SMSParamTemplateCode, true)
	if err != nil || templateCode == "" {
		return fmt.Errorf("SMS template code not configured")
	}

	// 创建客户端配置
	config := &openapi.Config{
		AccessKeyId:     tea.String(accessKeyID),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String("dysmsapi.aliyuncs.com"),
	}
	smsClient, err := dysmsapi.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create SMS client: %w", err)
	}

	// 构建请求
	templateParam := fmt.Sprintf(`{"code":"%s"}`, code)
	sendSmsRequest := &dysmsapi.SendSmsRequest{
		PhoneNumbers:  tea.String(phone),
		SignName:      tea.String(signName),
		TemplateCode:  tea.String(templateCode),
		TemplateParam: tea.String(templateParam),
	}

	// 发送短信
	runtime := &service.RuntimeOptions{}
	response, err := smsClient.SendSmsWithOptions(sendSmsRequest, runtime)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	// 检查响应
	if response.Body.Code != nil && *response.Body.Code != "OK" {
		return fmt.Errorf("SMS send failed: %s", tea.StringValue(response.Body.Message))
	}

	return nil
}

// generateValidateCode 生成指定长度的数字验证码
func generateValidateCode(length int) string {
	chars := "0123456789"
	code := make([]byte, length)
	for i := range code {
		code[i] = chars[rand.Intn(len(chars))]
	}
	return string(code)
}
