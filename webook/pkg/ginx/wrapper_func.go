package ginx

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"webook/webook/pkg/logger"
)

var l logger.LoggerV1 = logger.NewNopLogger()

// WrapBodyAndClaims bizFn 就是你的业务逻辑
func WrapBodyAndClaims[Req any, Claims any](bizFn func(ctx *gin.Context, req Req, uc Claims) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			l.Error("输入错误", logger.Error(err))
			return
		}
		l.Debug("输入参数", logger.Field{Key: "req", Value: req})
		val, ok := ctx.Get("user")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		uc, ok := val.(Claims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		res, err := bizFn(ctx, req, uc)
		if err != nil {
			l.Error("执行业务逻辑失败", logger.Error(err))
			return
		}
		ctx.JSON(http.StatusOK, res)
	}
}
