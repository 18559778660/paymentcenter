package controller

import (
	"errors"
	"io"

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

// Access A 站探活。通过 channel 或 group 参数识别路由，当前仅返回可达状态。
func (g *GatewayController) Access(c *gin.Context) {
	data := map[string]interface{}{
		"channel": c.Query("channel"),
		"group":   c.Query("group"),
		"status":  "ready",
	}
	response.SuccessMsg(c, data, "gateway ready")
}

// Pay A 站发起支付。URL 带 channel 或 group，Body 传订单信息，Header 或 Body 传商户密钥。
func (g *GatewayController) Pay(c *gin.Context) {
	var req service.GatewayPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	secretKey := c.GetHeader("X-Merchant-Secret")
	res, err := g.app.GatewayPay(req, service.GatewayPayQuery{
		Channel: c.Query("channel"),
		Group:   c.Query("group"),
	}, secretKey)
	if err != nil {
		writeGatewayPayError(c, err)
		return
	}
	response.SuccessMsg(c, res, "created")
}

// StripeWebhook Stripe 支付结果回调。
func (g *GatewayController) StripeWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.Fail(c, "读取请求失败")
		return
	}
	signature := c.GetHeader("Stripe-Signature")
	if err := g.app.HandleStripeWebhook(payload, signature); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.SuccessMsg(c, nil, "ok")
}

// writeGatewayPayError 写入网关支付错误。
func writeGatewayPayError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrGatewayMerchantInvalid):
		response.Fail(c, "商户密钥无效")
	case errors.Is(err, service.ErrGatewaySiteInvalid):
		response.Fail(c, "A站未审核或不属于该商户")
	case errors.Is(err, service.ErrGatewayRouteInvalid):
		response.Fail(c, "网关路由无效，请检查 channel 或 group 参数")
	case errors.Is(err, service.ErrGatewayChannelDisabled):
		response.Fail(c, "通道已禁用")
	case errors.Is(err, service.ErrGatewayAccountUnavailable):
		response.Fail(c, "暂无可用通道账号")
	case errors.Is(err, service.ErrGatewayPlatformUnsupported):
		response.Fail(c, "当前仅支持 Stripe 直连")
	case errors.Is(err, service.ErrGatewayStripeFailed):
		response.Fail(c, err.Error())
	default:
		response.Fail(c, err.Error())
	}
}
