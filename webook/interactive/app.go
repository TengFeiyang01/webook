package main

import (
	"github.com/TengFeiyang01/webook/webook/pkg/grpcx"
	"github.com/TengFeiyang01/webook/webook/pkg/saramax"
)

type App struct {
	server    *grpcx.Server
	consumers []saramax.Consumer
}
