package middleware

import (
	"encoding/gob"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
)

type LoginMiddlewareBuilder struct {
	paths []string
}

// LoginJWTMiddlewareBuilder JWT 登录校验
type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 不需要登录校验
		for _, path := range l.paths {
			if path == ctx.Request.URL.Path {
				return
			}
		}
		// 我现在使用 JWT 来登录校验
		tokenHeader := ctx.GetHeader("Authorization")
		// 没带 jwt
		if tokenHeader == "" {
			// 没带 jwt
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		segs := strings.Split(tokenHeader, " ")
		// 带了 格式不对
		if len(segs) != 2 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := segs[1]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte("GpJCNEnLiNblrZj5xdY9aG5cgVdKHCxh"), nil
		})
		// 格式对了内容不对
		if err != nil {
			// 没登陆 Bearer xxx1234
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if token == nil || !token.Valid {
			// 没登陆
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// 此时 jwt 校验成功
	}
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
