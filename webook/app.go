package main

import (
	"github.com/gin-gonic/gin"
	"webook/webook/internal/events"
)

type App struct {
	Server    *gin.Engine
	Consumers []events.Consumer
}
