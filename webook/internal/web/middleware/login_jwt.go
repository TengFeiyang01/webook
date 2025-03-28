package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	ijwt "github.com/TengFeiyang01/webook/webook/internal/web/jwt"
)

// LoginJWTMiddlewareBuilder JWT 登录校验
type LoginJWTMiddlewareBuilder struct {
	paths []string
	ijwt.Handler
}

func NewLoginJWTMiddlewareBuilder(jwtHdl ijwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		Handler: jwtHdl,
	}
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

		tokenStr := l.ExtractToken(ctx)
		claims := ijwt.UserClaims{}
		// 一定要传入指针
		token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("GpJCNEnLiNblrZj5xdY9aG5cgVdKHCxh"), nil
		})
		// 格式对了内容不对
		if err != nil {
			// 没登陆 Bearer xxx1234
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 短 token 过期了，长 token 还在
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
		err = l.CheckSession(ctx, claims.Ssid)
		if err != nil {
			// 要么 redis 的问题, 要么已经退出登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ctx.Set("user", claims)
		//ctx.Set("user_id", claims.Uid)
	}
}
