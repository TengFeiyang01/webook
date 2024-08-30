package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	server := InitWebUser()

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(200, "hello world")
	})
	_ = server.Run(":8080")
}
