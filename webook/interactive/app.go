package main

import (
	"webook/webook/pkg/grpcx"
	"webook/webook/pkg/saramax"
)

type App struct {
	server    *grpcx.Server
	consumers []saramax.Consumer
}
