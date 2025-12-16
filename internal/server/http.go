package server

import (
	"github.com/weetime/agent-matrix/internal/conf"
	"github.com/weetime/agent-matrix/internal/middleware"
	"github.com/weetime/agent-matrix/internal/service"
	v1 "github.com/weetime/agent-matrix/protos/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-kratos/swagger-api/openapiv2"
)

// NewHTTPServer new a HTTP server.
func NewHTTPServer(c *conf.Bootstrap,
	apiKey *service.ApiKeyService,
	config *service.ConfigService,
	sysParams *service.SysParamsService,
	agent *service.AgentService,
	user *service.UserService,
	sysDictType *service.SysDictTypeService,
	sysDictData *service.SysDictDataService,
	device *service.DeviceService,
	model *service.ModelService,
	ttsVoice *service.TtsVoiceService,
	tokenService middleware.TokenService,
	admin *service.AdminService,
	dataset *service.DatasetService,
	voiceClone *service.VoiceCloneService,
	ota *service.OtaService,
	logger log.Logger,
) *http.Server {

	// 创建 ServerSecretService 适配器
	serverSecretService := service.NewServerSecretServiceAdapter(config.GetConfigUsecase())

	opts := []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			validate.Validator(),
			logging.Server(logger),
			middleware.AuthMiddleware(tokenService, serverSecretService), // 添加认证中间件
		),
	}
	if c.Server.Http.Network != "" {
		opts = append(opts, http.Network(c.Server.Http.Network))
	}
	if c.Server.Http.Addr != "" {
		opts = append(opts, http.Address(c.Server.Http.Addr))
	}
	if c.Server.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Server.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	// 注册自定义HTTP handlers（必须在protobuf路由注册之前，确保优先匹配）
	service.RegisterVoiceCloneHTTPHandlers(srv, voiceClone)
	service.RegisterOtaHTTPHandlers(srv, ota)
	service.RegisterAgentChatHistoryHTTPHandlers(srv, agent)

	v1.RegisterApiKeyServiceHTTPServer(srv, apiKey)
	v1.RegisterConfigServiceHTTPServer(srv, config)
	v1.RegisterSysParamsServiceHTTPServer(srv, sysParams)
	v1.RegisterAgentServiceHTTPServer(srv, agent)
	v1.RegisterUserServiceHTTPServer(srv, user)
	v1.RegisterSysDictTypeServiceHTTPServer(srv, sysDictType)
	v1.RegisterSysDictDataServiceHTTPServer(srv, sysDictData)
	v1.RegisterDeviceServiceHTTPServer(srv, device)
	v1.RegisterModelServiceHTTPServer(srv, model)
	v1.RegisterTtsVoiceServiceHTTPServer(srv, ttsVoice)
	v1.RegisterAdminServiceHTTPServer(srv, admin)
	v1.RegisterDatasetServiceHTTPServer(srv, dataset)
	v1.RegisterVoiceCloneServiceHTTPServer(srv, voiceClone)
	v1.RegisterOtaServiceHTTPServer(srv, ota)
	srv.HandlePrefix("/q/", openapiv2.NewHandler())
	srv.HandleFunc("/ws", service.WebSocketHandler)
	return srv
}
