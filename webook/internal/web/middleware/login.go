package middleware

import (
	"encoding/gob"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddleBuilder struct {
	paths []string
}

func NewLoginMiddleBuilder() *LoginMiddleBuilder {
	return &LoginMiddleBuilder{}
}

func (l *LoginMiddleBuilder) IgnorePaths(path string) *LoginMiddleBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginMiddleBuilder) Build() gin.HandlerFunc {
	gob.Register(time.Time{})
	return func(ctx *gin.Context) {
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		// 不需要登录校验的
		if ctx.Request.URL.Path == "/users/login" || ctx.Request.URL.Path == "/users/signup" {
			return
		}

		sess := sessions.Default(ctx)
		id := sess.Get("userId")
		if id == nil {
			// 没有登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 先拿到上一次的的更新时间
		updateTime := sess.Get("update_time")
		sess.Options(sessions.Options{
			MaxAge: 60,
		})
		now := time.Now().UnixMilli()
		sess.Set("userId", id)
		// 还没刷新过，刚登陆
		if updateTime == nil {
			sess.Set("update_time", now)
			sess.Save()
			return
		}
		// updateTime 是有的
		updateTimeVal, _ := updateTime.(int64)

		// 超时未刷新
		// now.Sub(updateTimeVal) > time.Minute 这是采用Time.time
		if now-updateTimeVal > 60*1000 {
			sess.Set("update_time", now)
			sess.Save()
			return
		}

	}
}
