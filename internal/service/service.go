package service

import (
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(
	NewApiKeyService,
	NewConfigService,
	NewSysParamsService,
	NewAgentService,
	NewUserService,
	NewSysDictTypeService,
	NewSysDictDataService,
	NewDeviceService,
	NewModelService,
	NewTtsVoiceService,
	NewAdminService,
	NewDatasetService,
	NewVoiceCloneService,
	NewOtaService,
)
