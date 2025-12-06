package biz

import (
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(
	NewApiKeyUsecase,
	NewConfigUsecase,
	NewAgentUsecase,
	NewUserUsecase,
	NewUserTokenService,
	NewParamsServiceAdapter,
	NewCaptchaServiceAdapter,
	NewDictTypeUsecase,
	NewDictDataUsecase,
	NewDeviceUsecase,
	NewModelUsecase,
	NewModelProviderUsecase,
	NewTtsVoiceUsecase,
	NewAgentVoicePrintUsecase,
)
