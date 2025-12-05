package biz

import (
	"context"

	"github.com/weetime/agent-matrix/internal/kit"
)

// paramsServiceAdapter 参数服务适配器
type paramsServiceAdapter struct {
	configUsecase *ConfigUsecase
}

// NewParamsServiceAdapter 创建参数服务适配器
func NewParamsServiceAdapter(configUsecase *ConfigUsecase) ParamsService {
	return &paramsServiceAdapter{configUsecase: configUsecase}
}

func (p *paramsServiceAdapter) GetValue(paramCode string, isCache bool) (string, error) {
	return p.configUsecase.GetValue(context.Background(), paramCode, isCache)
}

// captchaServiceAdapter 验证码服务适配器
type captchaServiceAdapter struct {
	redisClient *kit.RedisClient
}

// NewCaptchaServiceAdapter 创建验证码服务适配器
func NewCaptchaServiceAdapter(redisClient *kit.RedisClient) CaptchaService {
	return &captchaServiceAdapter{redisClient: redisClient}
}

func (c *captchaServiceAdapter) Validate(ctx context.Context, uuid, code string, delete bool) bool {
	return kit.ValidateCaptcha(ctx, c.redisClient.GetClient(), uuid, code, delete)
}
