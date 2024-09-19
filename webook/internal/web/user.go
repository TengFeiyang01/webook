package web

import (
	"errors"
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
	"webook/webook/internal/domain"
	"webook/webook/internal/service"
	ijwt "webook/webook/internal/web/jwt"
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
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService, cmd redis.Cmdable, jwtHdl ijwt.Handler) *UserHandler {
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
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")

	//ug.POST("/login", u.Login)
	ug.POST("/login", u.LoginJWT)
	ug.POST("/signup", u.SignUp)
	ug.POST("/edit", u.Edit)
	ug.GET("/profile", u.ProfileJWT)
	//ug.POST("/logout", u.Logout)
	ug.POST("/logout", u.LogoutJWT)
	ug.POST("/login_sms/code/send", u.SendLoginSMSCode)
	ug.POST("/login_sms", u.LoginSMS)
	ug.POST("/refresh_token", u.RefreshToken)
}

func (u *UserHandler) LogoutJWT(ctx *gin.Context) {
	err := u.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "failed to logout",
		})
	}
	ctx.JSON(http.StatusOK, Result{
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
	ctx.JSON(http.StatusOK, Result{
		Msg: "ok",
	})
}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	type SignupReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	var req SignupReq
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	isEmail, err := u.emailRegExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	if !isEmail {
		ctx.String(http.StatusUnauthorized, "你的邮箱格式不对")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusUnauthorized, "两次输入的密码不一致")
		return
	}

	isPassword, err := u.passwordRegExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	if !isPassword {
		ctx.String(http.StatusBadRequest, "密码必须包含数字、特殊字符，并且长度不能小于 8 位")
		return
	}
	err = u.svc.SignUp(ctx, domain.User{Email: req.Email, Password: req.Password})
	if errors.Is(err, service.ErrUserDuplicate) {
		ctx.String(http.StatusOK, "邮箱冲突")
		return
	}
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统异常")
		return
	}

	ctx.String(http.StatusOK, "hello 注册成功")
}

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var req LoginReq
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		ctx.String(http.StatusUnauthorized, "用户名或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	// 用 JWT 设置登录态
	// 生成一个 JWT token

	if err = u.SetLoginToken(ctx, user.ID); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg: "系统错误",
		})
		return
	}
	fmt.Println(user)
	ctx.String(http.StatusOK, "登陆成功")
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var req LoginReq
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		ctx.String(http.StatusUnauthorized, "用户名或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	// 登录成功
	sess := sessions.Default(ctx)
	sess.Set("user_id", user.ID)
	sess.Options(sessions.Options{
		MaxAge: 60,
	})
	_ = sess.Save()
	ctx.String(http.StatusOK, "登陆成功")
}

func (u *UserHandler) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sess.Options(sessions.Options{
		//Secure: true,
		//HttpOnly: true,
		MaxAge: -1,
	})
	_ = sess.Save()

	ctx.String(http.StatusOK, "退出登录成功")
}

func (u *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		Email    string `json:"Email,omitempty"`
		Nickname string `json:"nickname" json:"Nickname,omitempty"`
		Birthday string `json:"birthday" json:"Birthday,omitempty"`
		AboutMe  string `json:"about_me" json:"AboutMe,omitempty"`
	}

	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	if req.Nickname == "" {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "昵称不能为空",
		})
		return
	}
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "日期格式不对",
		})
		return
	}
	if len(req.Nickname) > 1024 {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "关于我太长",
		})
	}

	uc := ctx.MustGet("user").(ijwt.UserClaims)
	err = u.svc.UpdateNonSensitiveInfo(ctx, domain.User{
		ID:       uc.Uid,
		NickName: req.Nickname,
		BirthDay: birthday,
		AboutMe:  req.AboutMe,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	type Profile struct {
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"about_me"`
	}
	c, ok := ctx.Get("claims")
	// 必然有 claims
	if !ok {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	user, err := u.svc.Profile(ctx, claims.Uid)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Profile{
		Email:    user.Email,
		Phone:    user.Phone,
		Nickname: user.NickName,
		Birthday: user.BirthDay.Format(time.DateOnly),
		AboutMe:  user.AboutMe,
	})
}

func (u *UserHandler) Profile(ctx *gin.Context) {
	type Profile struct {
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"about_me"`
	}
	sess := sessions.Default(ctx)
	id := sess.Get("user_id").(int64)
	user, err := u.svc.Profile(ctx, id)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Profile{
		Email:    user.Email,
		Phone:    user.Phone,
		Nickname: user.NickName,
		Birthday: user.BirthDay.Format(time.DateOnly),
		AboutMe:  user.AboutMe,
	})
}

func (u *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 是不是一个合法的手机号码
	// 考虑正则表达式
	ok, err := u.phoneRegExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Msg: "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusUnauthorized, Result{
			Msg: "手机号格式错误",
		})
		return
	}

	err = u.codeSvc.Send(ctx, biz, req.Phone)
	switch {
	case err == nil:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送成功",
		})
	case errors.Is(err, service.ErrCodeSendTooMany):
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送太频繁, 请稍后再试",
		})
	default:
		ctx.JSON(http.StatusInternalServerError, Result{
			Msg: "系统错误",
		})
	}
}

func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: http.StatusBadRequest,
			Msg:  "bind失败",
		})
		return
	}

	ok, err := u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		})
		return
	}

	if !ok {
		ctx.JSON(http.StatusUnauthorized, Result{
			Code: http.StatusUnauthorized,
			Msg:  "验证码不正确",
		})
		return
	}

	// 我这个手机号，会不会是一个新用户呢?
	// 这样子
	user, err := u.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		})
		return
	}

	// 这边要怎么样呢
	if err := u.SetLoginToken(ctx, user.ID); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg: "系统错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: http.StatusOK,
		Msg:  "验证码校验成功",
	})
}
