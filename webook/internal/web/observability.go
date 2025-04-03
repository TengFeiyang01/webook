package web

import (
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"time"
)

type ObservabilityHandler struct {
}

func (o *ObservabilityHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("test")
	g.GET("/metric", func(ctx *gin.Context) {
		sleep := rand.Int31n(1000)
		time.Sleep(time.Millisecond * time.Duration(sleep))
		ctx.String(http.StatusOK, "OK")
	})
}
