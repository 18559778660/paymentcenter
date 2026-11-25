package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
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
	response.ShopyyPaySuccess(c, res.CheckoutURL, res.PayCode, res.PayHTMLType)
}

// OrderStatusUpdate B 站壳包回写支付状态。路径必须是 /api/gateway/orderStatusUpdate。
func (g *GatewayController) OrderStatusUpdate(c *gin.Context) {
	token := firstNonEmptyHeader(c, "Api-Token", "api-token")
	orderJSON := strings.TrimSpace(c.PostForm("order"))
	if orderJSON == "" {
		payload, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "读取请求失败"})
			return
		}
		orderJSON = parseSiteBOrderJSON(payload)
	}
	res, err := g.app.SiteBOrderStatusUpdate(token, orderJSON)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":          1,
		"msg":           "ok",
		"redirect_next": res.RedirectNext,
		"status":        res.Status,
	})
}

// StripeWebhook 路1遗留：接收 Stripe 打到支付中心的 webhook。
// 路2 付款结果由 B 站调 OrderStatusUpdate，不会再走这里。
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
		response.ShopyyPayFail(c, "当前仅支持 Stripe 直连") // 路1遗留文案，现实际为 Stripe 经 B 站中转
	case errors.Is(err, service.ErrGatewayStripeFailed):
		response.ShopyyPayFail(c, err.Error())
	case errors.Is(err, service.ErrGatewaySiteBFailed):
		response.ShopyyPayFail(c, err.Error())
	default:
		response.ShopyyPayFail(c, err.Error())
	}
}

// firstNonEmptyHeader 获取第一个非空头信息。
func firstNonEmptyHeader(c *gin.Context, keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(c.GetHeader(key)); v != "" {
			return v
		}
	}
	return ""
}

// parseSiteBOrderJSON 解析 SiteB 订单 JSON。
func parseSiteBOrderJSON(payload []byte) string {
	trimmed := strings.TrimSpace(string(payload))
	if trimmed == "" {
		return ""
	}
	var wrapper struct {
		Order json.RawMessage `json:"order"`
	}
	if err := json.Unmarshal(payload, &wrapper); err == nil && len(wrapper.Order) > 0 {
		if len(wrapper.Order) > 0 && wrapper.Order[0] == '"' {
			var asString string
			if err := json.Unmarshal(wrapper.Order, &asString); err == nil {
				return strings.TrimSpace(asString)
			}
		}
		return strings.TrimSpace(string(wrapper.Order))
	}
	return trimmed
}
