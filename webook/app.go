package main

import (
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"webook/webook/interactive/events"
)

type App struct {
	Server    *gin.Engine
	Consumers []events.Consumer
	cron      *cron.Cron
}
