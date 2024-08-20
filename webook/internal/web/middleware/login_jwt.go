package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"strings"
	"time"
	"webook/webook/internal/web"
)

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
		claims := &web.UserClaims{}
		// 一定要传入指针
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("GpJCNEnLiNblrZj5xdY9aG5cgVdKHCxh"), nil
		})
		// 格式对了内容不对
		if err != nil {
			// 没登陆 Bearer xxx1234
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// !token.Valid 自动检测过期
		if token == nil || !token.Valid || claims.Uid == 0 {
			// 没登陆
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if claims.UserAgent != ctx.Request.UserAgent() {
			// 严重的安全问题
			// 你要加监控
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 此时 jwt 校验成功
		// 每十秒刷新一次 x-jwt-token
		now := time.Now()
		if claims.ExpiresAt.Sub(now) < time.Second*50 {
			claims.ExpiresAt = jwt.NewNumericDate(now.Add(time.Minute))
			tokenStr, err = token.SignedString([]byte("GpJCNEnLiNblrZj5xdY9aG5cgVdKHCxh"))

			if err != nil {
				// 记录日志
				log.Println("jwt 续约失败", err)
			}
			ctx.Header("x-jwt-token", tokenStr)
		}
		ctx.Set("claims", claims)
		//ctx.Set("user_id", claims.Uid)
	}
}
