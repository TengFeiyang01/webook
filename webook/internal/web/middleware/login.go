package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
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
		id := sess.Get("UserId")
		if id != nil {
			// 没有登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
