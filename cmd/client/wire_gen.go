// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"NakedVPN/internal/biz"
	"NakedVPN/internal/conf"
	"NakedVPN/internal/server"
	"NakedVPN/internal/service"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
)

// Injectors from wire.go:

// wireApp init kratos application.
func wireApp(client *conf.Client, logger log.Logger) (*kratos.App, func(), error) {
	handleClientUseCase := biz.NewHandleClientUseCase(client, logger)
	clientStreamProcessing := service.NewClientStreamProcessing(handleClientUseCase, client, logger)
	netClient := server.NewNetClient(client, clientStreamProcessing, logger)
	app := newApp(logger, netClient)
	return app, func() {
	}, nil
}
