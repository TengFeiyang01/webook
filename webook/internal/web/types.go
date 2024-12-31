package web

import "github.com/gin-gonic/gin"

type handler interface {
	RegisterRoutes(*gin.Engine)
}

type LoginSMSReq struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}
type SendLoginSMSReq struct {
	Phone string `json:"phone"`
}

type ProfileReq struct {
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Birthday string `json:"birthday"`
	AboutMe  string `json:"about_me"`
}

type LoginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SignupReq struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type EditReq struct {
	Nickname string `json:"nickname"`
	Birthday string `json:"birthday"`
	AboutMe  string `json:"about_me"`
}
