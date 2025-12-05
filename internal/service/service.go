package service

import (
	pb "github.com/weetime/agent-matrix/protos/v1"

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
)

var pbErrorInvalidUUID = pb.ErrorInvalidArgument("uuid is invalid")
