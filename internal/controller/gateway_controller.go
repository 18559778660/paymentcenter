package controller

import (
	"github.com/gin-gonic/gin"

	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// GatewayController A 站对接网关入口（无需登录）。
type GatewayController struct {
	app *service.App
}

// NewGatewayController 创建网关控制器。
func NewGatewayController(app *service.App) *GatewayController {
	return &GatewayController{app: app}
}

// Access A 站访问入口。通过 channel 参数识别通道，当前仅返回可达状态。
func (g *GatewayController) Access(c *gin.Context) {
	channel := c.Query("channel")
	if channel == "" {
		channel = c.Query("Channel")
	}
	data := g.app.GatewayAccess(channel)
	response.SuccessMsg(c, data, "gateway ready")
}
