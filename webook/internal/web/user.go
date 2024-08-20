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

// UserHandler 和用户有关的路由
type UserHandler struct {
	svc            *service.UserService
	emailRegExp    *regexp.Regexp
	passwordRegExp *regexp.Regexp
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	const (
		emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		// 和上面比起来，用 ` 看起来就比较清爽
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	)
	return &UserHandler{
		svc:            svc,
		emailRegExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
	}
}

func (u *UserHandler) RegisterRoute(server *gin.Engine) {
	ug := server.Group("/users")

	//ug.POST("/login", u.Login)
	ug.POST("/login", u.LoginJWT)
	ug.POST("/signup", u.SignUp)
	ug.POST("/edit", u.Edit)
	ug.GET("/profile", u.ProfileJWT)
	ug.POST("/logout", u.Logout)
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
	fmt.Println(req)

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
		ctx.String(http.StatusOK, "两次输入的密码不一致")
		return
	}

	isPassword, err := u.passwordRegExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	if !isPassword {
		ctx.String(http.StatusOK, "密码必须包含数字、特殊字符，并且长度不能小于 8 位")
		return
	}
	err = u.svc.SignUp(ctx, domain.User{Email: req.Email, Password: req.Password})
	if errors.As(err, &service.ErrUserDuplicateEmail) {
		ctx.String(http.StatusOK, "邮箱冲突")
		return
	}
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
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
	if errors.As(err, &service.ErrInvalidUserOrPassword) {
		ctx.String(http.StatusUnauthorized, "用户名或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	// 用 JWT 设置登录态
	// 生成一个 JWT token

	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			// 过期时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
		Uid:       user.ID,
		UserAgent: ctx.Request.UserAgent(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("GpJCNEnLiNblrZj5xdY9aG5cgVdKHCxh"))
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	ctx.Header("x-jwt-token", tokenStr)
	fmt.Println(tokenStr, user)
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
	if errors.As(err, &service.ErrInvalidUserOrPassword) {
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
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"about_me"`
	}
	var req EditReq
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	_, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		ctx.String(http.StatusOK, "生日格式不对")
		return
	}
}

func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
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
	println(claims.Uid)
	ctx.String(http.StatusOK, "你的 profile")
}

func (u *UserHandler) Profile(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sess.Get("user_id")
	ctx.String(200, "123")
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}
