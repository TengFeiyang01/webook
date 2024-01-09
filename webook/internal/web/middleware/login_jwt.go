package middleware

import (
	"encoding/gob"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"strings"
	"time"
	"webook/webook/internal/web"
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
		claims := &web.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("moyn8y9abnd7q4zkq2m73yw8tu9j5ixm"), nil
		})
		if err != nil {
			// 没登陆 Bearer
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		//claims.ExpiresAt.Time.Before(time.Now()) {
		//	// 过期了
		//}

		// err 为 nil token 不为 nil
		if token == nil || !token.Valid || claims.Uid == 0 {
			// 没登陆
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// 每 十秒钟 刷新一次
		now := time.Now()
		if claims.ExpiresAt.Sub(now) < time.Second*50 {
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute))
			tokenStr, err = token.SignedString([]byte("moyn8y9abnd7q4zkq2m73yw8tu9j5ixm"))
			if err != nil {
				// 记录日志
				log.Println("jwt 续约失败", err)
			}
			// 放进 header 中
			ctx.Header("x-jwt-token", tokenStr)
		}
		ctx.Set("claims", claims)
		//ctx.Set("userId", claims.Uid)
	}
}
