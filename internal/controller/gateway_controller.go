package controller

import (
	"errors"
	"io"
	"log"
	"strings"

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

// Pay A 站发起支付。URL 带 channel 或 group，Body 传 Shopyy 订单参数，Header 传 Api-Token。
func (g *GatewayController) Pay(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.ShopyyPayFail(c, "读取请求失败")
		return
	}
	log.Printf(
		"shopyy pay request channel=%s group=%s api_token_set=%t body=%s",
		c.Query("channel"),
		c.Query("group"),
		strings.TrimSpace(c.GetHeader("Api-Token")) != "",
		string(payload),
	)
	req, err := service.NormalizeGatewayPayRequest(payload)
	if err != nil {
		if errors.Is(err, service.ErrGatewayParamInvalid) {
			response.ShopyyPayFail(c, "参数错误")
			return
		}
		response.ShopyyPayFail(c, err.Error())
		return
	}
	secretKey := c.GetHeader("Api-Token")
	res, err := g.app.GatewayPay(req, service.GatewayPayQuery{
		Channel: c.Query("channel"),
		Group:   c.Query("group"),
	}, secretKey)
	if err != nil {
		writeShopyyPayError(c, err)
		return
	}
	response.ShopyyPaySuccess(c, res.CheckoutURL, req.SiteMode)
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

func writeShopyyPayError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrGatewayMerchantInvalid):
		response.ShopyyPayFail(c, "商户密钥无效")
	case errors.Is(err, service.ErrGatewaySiteInvalid):
		response.ShopyyPayFail(c, "A站未审核或不属于该商户")
	case errors.Is(err, service.ErrGatewayRouteInvalid):
		response.ShopyyPayFail(c, "网关路由无效，请检查 channel 或 group 参数")
	case errors.Is(err, service.ErrGatewayChannelDisabled):
		response.ShopyyPayFail(c, "通道已禁用")
	case errors.Is(err, service.ErrGatewayAccountUnavailable):
		response.ShopyyPayFail(c, "暂无可用通道账号")
	case errors.Is(err, service.ErrGatewayAccountStripeKeyMissing):
		response.ShopyyPayFail(c, "通道账号未配置 Stripe 私钥")
	case errors.Is(err, service.ErrGatewayAccountWebhookSecretMissing):
		response.ShopyyPayFail(c, "通道账号未配置 Webhook 密钥")
	case errors.Is(err, service.ErrGatewayPlatformUnsupported):
		response.ShopyyPayFail(c, "当前仅支持 Stripe 直连")
	case errors.Is(err, service.ErrGatewayStripeFailed):
		response.ShopyyPayFail(c, err.Error())
	default:
		response.ShopyyPayFail(c, err.Error())
	}
}
