//go:build wireinject
// +build wireinject

// go:build wireinject
package main

import (
	"context"

	"nova/internal"
	"nova/internal/biz"
	"nova/internal/data"
	"nova/internal/hook"
	"nova/internal/server"
	"nova/internal/service"

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
