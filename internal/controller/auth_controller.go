package controller

import (
	"errors"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/middleware"
	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// AuthController 控制层：处理登录、用户信息、权限码、退出。
type AuthController struct {
	app *service.App
}

// NewAuthController 创建认证控制器。
func NewAuthController(app *service.App) *AuthController {
	return &AuthController{app: app}
}

// Login 登录。接收用户名密码，返回 accessToken，给前端存到请求头。
func (a *AuthController) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "用户名和密码不能为空")
		return
	}
	res, err := a.app.Login(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Fail(c, "用户名或密码错误")
			return
		}
		if errors.Is(err, service.ErrUserDisabled) {
			response.Fail(c, "账号已禁用")
			return
		}
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, res)
}

// UserInfo 获取当前登录用户信息。前端登录成功后会立刻调用。
func (a *AuthController) UserInfo(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	info, err := a.app.GetUserInfo(user.ID)
	if err != nil {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	response.Success(c, info)
}

// Codes 获取当前用户权限码列表。前端用来控制按钮级权限，没有则返回空数组。
func (a *AuthController) Codes(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	codes, err := a.app.GetAccessCodes(user.ID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, codes)
}

// Logout 退出登录。当前只返回成功，前端会自行清掉本地 token。
func (a *AuthController) Logout(c *gin.Context) {
	response.Success(c, "")
}

// Menus 返回当前用户菜单树，前端 accessMode=backend 时用来生成侧边栏。
func (a *AuthController) Menus(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	menus, err := a.app.GetUserMenus(user.ID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, menus)
}
