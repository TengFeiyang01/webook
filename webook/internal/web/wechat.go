package web

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"net/http"
	"time"
	"github.com/TengFeiyang01/webook/webook/internal/service"
	"github.com/TengFeiyang01/webook/webook/internal/service/oauth2/wechat"
	ijwt "github.com/TengFeiyang01/webook/webook/internal/web/jwt"
	"github.com/TengFeiyang01/webook/webook/pkg/ginx"
)

type OAuth2WechatHandler struct {
	svc     wechat.Service
	userSvc service.UserService
	ijwt.Handler
	stateKey []byte
	cfg      WechatHandlerConfig
}

type WechatHandlerConfig struct {
	Secure bool
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserService, cfg WechatHandlerConfig, jwtHdl ijwt.Handler) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:      svc,
		userSvc:  userSvc,
		stateKey: []byte("GpJCNEnLiNblrZj5xdY9aG0cgVdKHCxh"),
		cfg:      cfg,
		Handler:  jwtHdl,
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", h.AuthURL)
	g.Any("/callback", h.Callback)
}

func (h *OAuth2WechatHandler) AuthURL(ctx *gin.Context) {
	state := uuid.New()
	url, err := h.svc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "构造url失败",
		})
		return
	}
	if err := h.setStateCookie(ctx, state); err != nil {
		ctx.JSON(http.StatusInternalServerError, ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统异常",
		})
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Data: url,
	})
}

func (h *OAuth2WechatHandler) setStateCookie(ctx *gin.Context, state string) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, StateClaims{
		state: state,
		RegisteredClaims: jwt.RegisteredClaims{
			// 过期时间，预期中一个用户完成登录的时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	})
	tokenStr, err := token.SignedString(h.stateKey)
	if err != nil {
		return err
	}
	ctx.SetCookie("jwt-state", tokenStr, 600,
		"/oauth2/wechat/callback", "", h.cfg.Secure, true)
	return nil
}

type StateClaims struct {
	jwt.RegisteredClaims
	state string
}

func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	// 验证微信的 code
	code := ctx.Query("code")
	err := h.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "登录失败",
		})
		return
	}

	info, err := h.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		})
		return
	}
	// 从 userService 拿uid
	u, err := h.userSvc.FindOrCreateByWechat(ctx, info)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		})
		return
	}

	if err = h.SetLoginToken(ctx, u.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Code: http.StatusOK,
		Data: info,
	})
}

func (h *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")
	// 校验一下我的 state
	ck, err := ctx.Cookie("jwt-state")
	if err != nil {
		// 做好监控，有人搞你
		ctx.JSON(http.StatusInternalServerError, ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		})
		return fmt.Errorf("拿不到 state 的 cookie, %w", err)
	}
	var sc StateClaims
	token, err := jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return h.stateKey, nil
	})
	if err != nil || !token.Valid {
		ctx.JSON(http.StatusInternalServerError, ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		})
		// 做好监控，有人搞你
		return fmt.Errorf("token 已经过期,  %w", err)
	}
	if sc.state != state {
		return errors.New("state 不相等")
	}
	return nil
}
