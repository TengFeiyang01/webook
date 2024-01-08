package web

import (
	"errors"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"webook/webook/internal/domain"
	"webook/webook/internal/service"
)

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	userIdKey            = "userId"
)

type UserHandler struct {
	svc            *service.UserService
	emailRegExp    *regexp.Regexp
	passwordRegExp *regexp.Regexp
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		svc:            svc,
		emailRegExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
	}
}

func (c *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")

	ug.POST("/signup", c.SignUp)
	ug.POST("login", c.Login)
	ug.POST("/edit", c.Edit)
	ug.GET("/profile", c.Profile)
}

func (c *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	var req SignUpReq

	if err := ctx.Bind(&req); err != nil {
		return
	}

	isEmail, err := c.emailRegExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "邮箱不正确")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入的密码不一致")
		return
	}

	isPassword, err := c.passwordRegExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	if !isPassword {
		ctx.String(http.StatusOK, "密码必须包含数字、特殊字符，并且长度不能小于 8 位")
		return
	}

	// 调用一些 svc 的方法
	err = c.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if errors.Is(err, service.ErrUserDuplicateEmail) {
		ctx.String(http.StatusOK, "邮箱冲突")
		return
	}

	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}

	ctx.String(http.StatusOK, "hello 注册成功")
}

// Login 用户登录接口
func (c *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	// 当我们调用 Bind 方法的时候，如果有问题，Bind 方法已经直接写响应回去了
	if err := ctx.Bind(&req); err != nil {
		return
	}
	u, err := c.svc.Login(ctx.Request.Context(), req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidUserOrEmail) {
		ctx.String(http.StatusOK, "用户名或者密码不正确，请重试")
		return
	}
	sess := sessions.Default(ctx)
	sess.Set(userIdKey, u.Id)
	sess.Options(sessions.Options{
		// 60 秒过期
		MaxAge: 60,
	})
	err = sess.Save()
	if err != nil {
		ctx.String(http.StatusOK, "服务器异常")
		return
	}
	ctx.String(http.StatusOK, "登录成功")
}

func (c *UserHandler) Edit(ctx *gin.Context) {

}

// Profile 用户详情
func (c *UserHandler) Profile(ctx *gin.Context) {
	type Profile struct {
		Email string
	}
	sess := sessions.Default(ctx)
	id := sess.Get(userIdKey).(int64)
	u, err := c.svc.Profile(ctx, id)
	if err != nil {
		// 按照道理来说，这边 id 对应的数据肯定存在，所以要是没找到，
		// 那就说明是系统出了问题。
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Profile{
		Email: u.Email,
	})
}
