package service

import (
	pb "nova/protos/nova/v1"

	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(
	NewApiKeyService,
	NewConfigService,
)

var pbErrorInvalidUUID = pb.ErrorInvalidArgument("uuid is invalid")
