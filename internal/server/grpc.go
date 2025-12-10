package server

import (
	"github.com/weetime/agent-matrix/internal/conf"
	"github.com/weetime/agent-matrix/internal/service"

	v1 "github.com/weetime/agent-matrix/protos/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Bootstrap,
	apiKey *service.ApiKeyService,
	config *service.ConfigService,
	sysParams *service.SysParamsService,
	agent *service.AgentService,
	device *service.DeviceService,
	admin *service.AdminService,
	voiceClone *service.VoiceCloneService,
	ota *service.OtaService,
	logger log.Logger,
) *grpc.Server {

	opts := []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			validate.Validator(),
			logging.Server(logger),
		),
	}
	if c.Server.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Server.Grpc.Network))
	}
	if c.Server.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Server.Grpc.Addr))
	}
	if c.Server.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Server.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterApiKeyServiceServer(srv, apiKey)
	v1.RegisterConfigServiceServer(srv, config)
	v1.RegisterSysParamsServiceServer(srv, sysParams)
	v1.RegisterAgentServiceServer(srv, agent)
	v1.RegisterDeviceServiceServer(srv, device)
	v1.RegisterAdminServiceServer(srv, admin)
	v1.RegisterVoiceCloneServiceServer(srv, voiceClone)
	v1.RegisterOtaServiceServer(srv, ota)
	return srv
}
