package main

import (
	"github.com/TengFeiyang01/webook/webook/interactive/events"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

type App struct {
	Server    *gin.Engine
	Consumers []events.Consumer
	cron      *cron.Cron
}
