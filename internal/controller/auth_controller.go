package controller

import (
	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	app *service.App
}

func NewAuthController(app *service.App) *AuthController {
	return &AuthController{app: app}
}

// 登录
func (a *AuthController) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err.Error())
		return
	}
	res, err := a.app.Login(req)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, res)
}
