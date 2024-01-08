package middleware

import (
	"encoding/gob"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
)

// LoginJWTMiddleBuilder JWT 登录校验
type LoginJWTMiddleBuilder struct {
	paths []string
}

func NewLoginJWTMiddleBuilder() *LoginJWTMiddleBuilder {
	return &LoginJWTMiddleBuilder{}
}

func (l *LoginJWTMiddleBuilder) IgnorePaths(path string) *LoginJWTMiddleBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginJWTMiddleBuilder) Build() gin.HandlerFunc {
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

		// 我现在使用 JWT 来校验
		tokenHeader := ctx.GetHeader("Authorization")
		if tokenHeader == "" {
			// 没登陆 401
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		segs := strings.SplitN(tokenHeader, " ", 2)
		if len(segs) != 2 {
			// 没登录 有人瞎搞
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := segs[1]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte("moyn8y9abnd7q4zkq2m73yw8tu9j5ixm"), nil
		})
		if err != nil {
			// 没登陆 Bearer
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// err 为 nil token 不为 nil
		if token == nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
