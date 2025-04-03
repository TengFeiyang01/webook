package ginx

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
)

var l = logger.NewNopLogger()
var vector *prometheus.CounterVec

func InitCounter(opt prometheus.CounterOpts) {
	vector = prometheus.NewCounterVec(opt, []string{"code"})
	prometheus.MustRegister(vector)
}

// WrapBodyAndToken bizFn 就是你的业务逻辑
func WrapBodyAndToken[Req any, C any](bizFn func(ctx *gin.Context, req Req, uc C) (Result, error)) gin.HandlerFunc {
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
		uc, ok := val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		res, err := bizFn(ctx, req, uc)
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		if err != nil {
			l.Error("执行业务逻辑失败",
				logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
			return
		}
		ctx.JSON(http.StatusOK, res)
	}
}

// WrapBodyAndToken bizFn 就是你的业务逻辑
func WrapBodyAndTokenV1[Req any, C any](l logger.LoggerV1, bizFn func(ctx *gin.Context, req Req, uc C) (Result, error)) gin.HandlerFunc {
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
		uc, ok := val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		res, err := bizFn(ctx, req, uc)
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		if err != nil {
			l.Error("执行业务逻辑失败",
				logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
			return
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBodyV1[Req any](bizFn func(ctx *gin.Context, req Req) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			l.Error("输入错误", logger.Error(err))
			return
		}
		l.Debug("输入参数", logger.Field{Key: "req", Value: req})

		res, err := bizFn(ctx, req)
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		if err != nil {
			l.Error("执行业务逻辑失败",
				logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
			return
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBody[Req any](l logger.LoggerV1, bizFn func(ctx *gin.Context, req Req) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			l.Error("输入错误", logger.Error(err))
			return
		}
		l.Debug("输入参数", logger.Field{Key: "req", Value: req})

		res, err := bizFn(ctx, req)
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		if err != nil {
			l.Error("执行业务逻辑失败",
				logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
			return
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapToken[C any](bizFn func(ctx *gin.Context, uc C) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		val, ok := ctx.Get("user")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		uc, ok := val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		res, err := bizFn(ctx, uc)
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		if err != nil {
			l.Error("执行业务逻辑失败",
				logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
			return
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapTokenV1[C any](l logger.LoggerV1, bizFn func(ctx *gin.Context, uc C) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		val, ok := ctx.Get("user")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		uc, ok := val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		res, err := bizFn(ctx, uc)
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		if err != nil {
			l.Error("执行业务逻辑失败",
				logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
			return
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapClaimsV1[Claims any](bizFn func(ctx *gin.Context, uc Claims) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
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
		res, err := bizFn(ctx, uc)
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		if err != nil {
			l.Error("执行业务逻辑失败",
				logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
			return
		}
		ctx.JSON(http.StatusOK, res)
	}
}
