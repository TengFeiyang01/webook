package web

import (
	"errors"
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"time"
	"github.com/TengFeiyang01/webook/webook/internal/domain"
	"github.com/TengFeiyang01/webook/webook/internal/errs"
	"github.com/TengFeiyang01/webook/webook/internal/service"
	ijwt "github.com/TengFeiyang01/webook/webook/internal/web/jwt"
	"github.com/TengFeiyang01/webook/webook/pkg/ginx"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
)

const biz = "login"

// 确保 UserHandler 实现了 handler 的接口
// var _ handler = &UserHandler{}
// 这个更优雅
var _ handler = (*UserHandler)(nil)

// UserHandler 和用户有关的路由
type UserHandler struct {
	svc            service.UserService
	codeSvc        service.CodeService
	emailRegExp    *regexp.Regexp
	passwordRegExp *regexp.Regexp
	phoneRegExp    *regexp.Regexp
	cmd            redis.Cmdable
	ijwt.Handler
	l logger.LoggerV1
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService, cmd redis.Cmdable, jwtHdl ijwt.Handler, l logger.LoggerV1) *UserHandler {
	const (
		emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		// 和上面比起来，用 ` 看起来就比较清爽
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
		phoneRegexPattern    = `^(?:(?:\+|00)86)?1(?:(?:3[\d])|(?:4[5-79])|(?:5[0-35-9])|(?:6[5-7])|(?:7[0-8])|(?:8[\d])|(?:9[01256789]))\d{8}$`
	)
	return &UserHandler{
		svc:            svc,
		codeSvc:        codeSvc,
		emailRegExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		phoneRegExp:    regexp.MustCompile(phoneRegexPattern, regexp.None),
		cmd:            cmd,
		Handler:        jwtHdl,
		l:              l,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")

	//ug.POST("/login", u.Login)
	ug.POST("/login", ginx.WrapBody[LoginReq](u.l.With(logger.String("method", "loginJWT")), u.LoginJWT))
	ug.POST("/signup", ginx.WrapBody[SignupReq](u.l.With(logger.String("method", "signup")), u.SignUp))
	ug.POST("/edit", ginx.WrapBodyAndTokenV1[EditReq, ijwt.UserClaims](u.l.With(logger.String("method", "edit")), u.Edit))
	ug.GET("/profile", ginx.WrapTokenV1[ijwt.UserClaims](u.l.With(logger.String("method", "profileJWT")), u.ProfileJWT))
	ug.POST("/login_sms/code/send", ginx.WrapBody[SendLoginSMSReq](u.l.With(logger.String("method", "sendLoginSmsCode")), u.SendLoginSMSCode))
	ug.POST("/login_sms", ginx.WrapBody[LoginSMSReq](u.l.With(logger.String("method", "loginSms")), u.LoginSMS))
	//ug.POST("/logout", u.Logout)
	ug.POST("/logout", u.LogoutJWT)
	ug.POST("/refresh_token", u.RefreshToken)
}

func (u *UserHandler) LogoutJWT(ctx *gin.Context) {
	err := u.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ginx.Result{
			Code: 5,
			Msg:  "failed to logout",
		})
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Code: 0,
		Msg:  "logout success",
	})
}

// RefreshToken 可以同时刷新长短 token, 用 Redis 来记录是否有效, 即 Refresh_token 是一次性的
// 参考登录校验 比较 User-Agent 来增强安全性
func (u *UserHandler) RefreshToken(ctx *gin.Context) {
	// 只有这个接口拿出来的 才是 refresh_token 其他地方都是 access_token
	refreshToken := u.ExtractToken(ctx)
	var rc ijwt.RefreshClaims
	token, err := jwt.ParseWithClaims(refreshToken, &rc, func(*jwt.Token) (interface{}, error) {
		return ijwt.RtKey, nil
	})
	if err != nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err = u.CheckSession(ctx, rc.Ssid)
	if err != nil {
		// 要么 redis 的问题, 要么已经退出登录
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if err := u.SetJWTToken(ctx, rc.Uid, rc.Ssid); err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "ok",
	})
}

func (u *UserHandler) SignUp(ctx *gin.Context, req SignupReq) (ginx.Result, error) {
	isEmail, err := u.emailRegExp.MatchString(req.Email)
	if err != nil {
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, err
	}

	if !isEmail {
		return ginx.Result{
			Code: http.StatusUnauthorized,
			Msg:  "你的邮箱格式不对",
		}, fmt.Errorf("你的邮箱格式不对")
	}
	fmt.Println("==================")
	if req.Password != req.ConfirmPassword {
		return ginx.Result{
			Code: http.StatusUnauthorized,
			Msg:  "两次输入的密码不一致",
		}, fmt.Errorf("两次输入的密码不一致")
	}

	isPassword, err := u.passwordRegExp.MatchString(req.Password)
	if err != nil {
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, fmt.Errorf("系统错误")
	}

	if !isPassword {
		return ginx.Result{
			Code: http.StatusBadRequest,
			Msg:  "密码必须包含数字、特殊字符，并且长度不能小于 8 位",
		}, fmt.Errorf("密码必须包含数字、特殊字符，并且长度不能小于 8 位")
	}
	err = u.svc.SignUp(ctx.Request.Context(), domain.User{Email: req.Email, Password: req.Password})
	if errors.Is(err, service.ErrUserDuplicate) {
		// 这是复用
		span := trace.SpanFromContext(ctx.Request.Context())
		span.AddEvent("邮箱冲突")
		return ginx.Result{
			Code: http.StatusOK,
			Msg:  "邮箱冲突",
		}, err
	}
	if err != nil {
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, fmt.Errorf("系统错误")
	}

	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "hello 注册成功",
	}, nil
}

func (u *UserHandler) LoginJWT(ctx *gin.Context, req LoginReq) (ginx.Result, error) {
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		return ginx.Result{
			Code: errs.UserInvalidOrPassword,
			Msg:  "用户名或密码不对",
		}, err
	}
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, fmt.Errorf("系统错误")
	}

	if err = u.SetLoginToken(ctx, user.ID); err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, fmt.Errorf("系统错误")
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "登陆成功",
	}, nil
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var req LoginReq
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		})
		return
	}
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		ctx.JSON(http.StatusUnauthorized, ginx.Result{
			Code: http.StatusUnauthorized,
			Msg:  "用户名或密码不对",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		})
		return
	}

	sess := sessions.Default(ctx)
	sess.Set("user_id", user.ID)
	sess.Options(sessions.Options{
		MaxAge: 60,
	})
	_ = sess.Save()
	ctx.JSON(http.StatusOK, ginx.Result{
		Code: http.StatusOK,
		Msg:  "登陆成功",
	})
}

func (u *UserHandler) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sess.Options(sessions.Options{
		MaxAge: -1,
	})
	_ = sess.Save()

	ctx.JSON(http.StatusOK, ginx.Result{
		Code: http.StatusOK,
		Msg:  "退出登录成功",
	})
}

func (u *UserHandler) Edit(ctx *gin.Context, req EditReq, uc ijwt.UserClaims) (ginx.Result, error) {
	if req.Nickname == "" {
		return ginx.Result{Code: 4, Msg: "昵称不能为空"}, fmt.Errorf("昵称为空")
	}

	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		return ginx.Result{Code: http.StatusOK, Msg: "日期格式不对"}, err
	}
	if len(req.Nickname) > 1024 {
		return ginx.Result{Code: 4, Msg: "关于我太长"}, fmt.Errorf("关于我太长")
	}

	err = u.svc.UpdateNonSensitiveInfo(ctx.Request.Context(), domain.User{
		ID:       uc.Uid,
		NickName: req.Nickname,
		BirthDay: birthday,
		AboutMe:  req.AboutMe,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}
	return ginx.Result{Code: http.StatusOK, Msg: "OK"}, nil
}

func (u *UserHandler) ProfileJWT(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	user, err := u.svc.Profile(ctx.Request.Context(), uc.Uid)
	if err != nil {
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, err
	}

	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "OK",
		Data: ProfileReq{
			Email:    user.Email,
			Nickname: user.NickName,
			Birthday: user.BirthDay.Format(time.DateOnly),
			AboutMe:  user.AboutMe,
		},
	}, nil
}

func (u *UserHandler) Profile(ctx *gin.Context) (ginx.Result, error) {
	sess := sessions.Default(ctx)
	id := sess.Get("user_id").(int64)
	user, err := u.svc.Profile(ctx.Request.Context(), id)
	if err != nil {
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, err
	}

	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "OK",
		Data: ProfileReq{
			Email:    user.Email,
			Nickname: user.NickName,
			Birthday: user.BirthDay.Format(time.DateOnly),
			AboutMe:  user.AboutMe,
		},
	}, nil
}

func (u *UserHandler) SendLoginSMSCode(ctx *gin.Context, req SendLoginSMSReq) (ginx.Result, error) {
	ok, err := u.phoneRegExp.MatchString(req.Phone)
	if err != nil {
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	if !ok {
		return ginx.Result{
			Code: http.StatusUnauthorized,
			Msg:  "手机号格式错误",
		}, nil
	}

	err = u.codeSvc.Send(ctx.Request.Context(), biz, req.Phone)
	switch {
	case err == nil:
		return ginx.Result{
			Code: http.StatusOK,
			Msg:  "发送成功",
		}, nil
	case errors.Is(err, service.ErrCodeSendTooMany):
		return ginx.Result{
			Code: http.StatusOK,
			Msg:  "发送太频繁, 请稍后再试",
		}, nil
	default:
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, err
	}
}

func (u *UserHandler) LoginSMS(ctx *gin.Context, req LoginSMSReq) (ginx.Result, error) {
	ok, err := u.codeSvc.Verify(ctx.Request.Context(), biz, req.Phone, req.Code)
	if err != nil {
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, err
	}

	if !ok {
		return ginx.Result{
			Code: http.StatusUnauthorized,
			Msg:  "验证码不正确",
		}, nil
	}

	user, err := u.svc.FindOrCreate(ctx.Request.Context(), req.Phone)
	if err != nil {
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, err
	}

	if err := u.SetLoginToken(ctx, user.ID); err != nil {
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "登陆成功",
	}, nil
}
