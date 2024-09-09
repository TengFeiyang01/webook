package web

import (
	"errors"
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
	"webook/webook/internal/domain"
	"webook/webook/internal/service"
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
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService) *UserHandler {
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
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")

	//ug.POST("/login", u.Login)
	ug.POST("/login", u.LoginJWT)
	ug.POST("/signup", u.SignUp)
	ug.POST("/edit", u.Edit)
	ug.GET("/profile", u.ProfileJWT)
	ug.POST("/logout", u.Logout)
	ug.POST("/login_sms/code/send", u.SendLoginSMSCode)
	ug.POST("/login_sms", u.LoginSMS)
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

	if err = u.SetJWTToken(ctx, user.ID); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg: "系统错误",
		})
		return
	}
	fmt.Println(user)
	ctx.String(http.StatusOK, "登陆成功")
}

func (u *UserHandler) SetJWTToken(ctx *gin.Context, uid int64) error {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			// 过期时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid:       uid,
		UserAgent: ctx.Request.UserAgent(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("GpJCNEnLiNblrZj5xdY9aG5cgVdKHCxh"))
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
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

	uc := ctx.MustGet("user").(UserClaims)
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
	claims, ok := c.(*UserClaims)
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
	if err := ctx.ShouldBind(&req); err != nil {
		return
	}
	// 是不是一个合法的手机号码
	// 考虑正则表达式
	ok, err := u.phoneRegExp.MatchString(req.Phone)
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
			Msg:  "手机号格式错误",
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
		ctx.JSON(http.StatusOK, Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		})
	}
}

func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.ShouldBind(&req); err != nil {
		return
	}

	ok, err := u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		})
		return
	}

	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: http.StatusUnauthorized,
			Msg:  "验证码不正确",
		})
		return
	}

	// 我这个手机号，会不会是一个新用户呢?
	// 这样子
	user, err := u.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		})
		return
	}

	// 这边要怎么样呢
	if err := u.SetJWTToken(ctx, user.ID); err != nil {
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

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}
