//go:build wireinject
// +build wireinject

// go:build wireinject
package main

import (
	"context"

	"github.com/weetime/agent-matrix/internal"
	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/data"
	"github.com/weetime/agent-matrix/internal/hook"
	"github.com/weetime/agent-matrix/internal/server"
	"github.com/weetime/agent-matrix/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
)

func initApp(configPath string, ctx context.Context) (*kratos.App, func(), error) {
	panic(wire.Build(
		internal.ProviderSet,
		biz.ProviderSet,
		hook.ProviderSet,
		data.ProviderSet,
		service.ProviderSet,
		server.ProviderSet,
		newApp,
	))
}
