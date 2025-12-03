package server

import (
	"nova/internal/conf"
	"nova/internal/service"
	v1 "nova/protos/nova/v1"

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
	logger log.Logger,
) *http.Server {

	opts := []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			validate.Validator(),
			logging.Server(logger),
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
	v1.RegisterApiKeyServiceHTTPServer(srv, apiKey)
	v1.RegisterConfigServiceHTTPServer(srv, config)
	srv.HandlePrefix("/q/", openapiv2.NewHandler())
	srv.HandleFunc("/ws", service.WebSocketHandler)
	return srv
}
