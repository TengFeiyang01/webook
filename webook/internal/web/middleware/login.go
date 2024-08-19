package middleware

import (
	"encoding/gob"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

func (l *LoginMiddlewareBuilder) IgnorePaths(path string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		// 不需要登录校验
		for _, path := range l.paths {
			if path == ctx.Request.URL.Path {
				return
			}
		}
		sess := sessions.Default(ctx)
		if sess == nil {
			// 未登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		id := sess.Get("user_id")
		if id == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		updateTime := sess.Get("update_time")
		sess.Set("user_id", id)
		sess.Options(sessions.Options{
			MaxAge: 60,
		})
		now := time.Now()
		if updateTime == nil {
			sess.Set("update_time", now)
			if err := sess.Save(); err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
			return
		}
		if updateTimeVal, ok := updateTime.(time.Time); ok {
			if now.Sub(updateTimeVal) > 10*time.Second {
				sess.Set("update_time", now)
				_ = sess.Save()
				return
			}
		} else {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}
}
