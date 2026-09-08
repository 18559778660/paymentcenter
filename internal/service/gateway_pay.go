package service

import (
	"errors"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"paymentcenter/internal/model"
)

// 网关支付错误。
var (
	ErrGatewayMerchantInvalid             = errors.New("gateway merchant invalid")
	ErrGatewaySiteInvalid                 = errors.New("gateway site invalid")
	ErrGatewayRouteInvalid                = errors.New("gateway route invalid")
	ErrGatewayChannelDisabled             = errors.New("gateway channel disabled")
	ErrGatewayAccountUnavailable          = errors.New("gateway account unavailable")
	ErrGatewayPlatformUnsupported         = errors.New("gateway platform unsupported")
	ErrGatewayStripeFailed                = errors.New("gateway stripe failed")
	ErrGatewayAccountStripeKeyMissing     = errors.New("gateway account stripe key missing")
	ErrGatewayAccountWebhookSecretMissing = errors.New("gateway account webhook secret missing")
)

// GatewayPayRequest 网关支付内部订单字段（由 A 站入参映射而来）。
type GatewayPayRequest struct {
	MerchantOrder string
	MerchantSite  string
	Amount        int64
	Currency      string
	ReturnURL     string
	NotifyURL     string
	CancelURL     string
	ResultURL     string
	Subject       string
	SiteMode      string
	ShopyyVerify  ShopyyNotifyVerifySnapshot
}

// GatewayPayQuery 网关路由参数。
type GatewayPayQuery struct {
	Channel string
	Group   string
}

// GatewayPay A 站经网关创建支付并落库订单，按通道所属平台分发。
func (a *App) GatewayPay(req GatewayPayRequest, q GatewayPayQuery, secretKey string) (CreateOrderResponse, error) {
	secretKey = strings.TrimSpace(secretKey)
	merchantOrder := strings.TrimSpace(req.MerchantOrder)
	if secretKey == "" {
		a.appendOrderLog("", 0, merchantOrder, model.OrderLogTypeFailLog, "商户密钥无效")
		return CreateOrderResponse{}, ErrGatewayMerchantInvalid
	}
	merchant, err := a.validateGatewayMerchant(secretKey)
	if err != nil {
		a.appendOrderLog("", 0, merchantOrder, model.OrderLogTypeFailLog, "商户密钥无效")
		return CreateOrderResponse{}, err
	}
	siteDomain, err := a.validateGatewaySiteA(merchant, req.MerchantSite)
	if err != nil {
		a.appendOrderLog("", merchant.ID, merchantOrder, model.OrderLogTypeFailLog, "A站未审核或不属于该商户")
		return CreateOrderResponse{}, err
	}

	a.appendOrderLog("", merchant.ID, merchantOrder, model.OrderLogTypeRecord, "开始解析网关路由")
	channel, account, err := a.resolveGatewayRoute(q)
	if err != nil {
		a.appendOrderLog("", merchant.ID, merchantOrder, model.OrderLogTypeFailLog, "网关路由无效，请检查 channel 或 group 参数")
		return CreateOrderResponse{}, err
	}
	platform, err := a.validateGatewayChannel(channel)
	if err != nil {
		a.appendOrderLog("", merchant.ID, merchantOrder, model.OrderLogTypeFailLog, "通道已禁用")
		return CreateOrderResponse{}, err
	}
	if err := validateGatewayAccount(platform, account); err != nil {
		msg := "暂无可用通道账号"
		if errors.Is(err, ErrGatewayAccountStripeKeyMissing) {
			msg = "通道账号未配置 Stripe 私钥"
		} else if errors.Is(err, ErrGatewayAccountWebhookSecretMissing) {
			msg = "通道账号未配置 Webhook 密钥"
		} else if errors.Is(err, ErrGatewayPlatformUnsupported) {
			msg = "当前仅支持 Stripe 直连"
		}
		a.appendOrderLog("", merchant.ID, merchantOrder, model.OrderLogTypeFailLog, msg)
		return CreateOrderResponse{}, err
	}
	a.appendOrderLog("", merchant.ID, merchantOrder, model.OrderLogTypeRecord,
		"通道账号 ["+strconv.FormatUint(uint64(account.ID), 10)+"] 验证通过")

	now := time.Now().UTC()
	orderID := "pc_" + strconv.FormatInt(now.UnixNano(), 10)
	verifyJSON, err := json.Marshal(req.ShopyyVerify)
	if err != nil {
		return CreateOrderResponse{}, err
	}
	siteBID := account.SiteBID
	siteBDomain := ""
	if siteBID > 0 {
		if siteB, siteErr := a.store.GetSiteBByID(siteBID); siteErr == nil && siteB != nil {
			siteBDomain = strings.TrimSpace(siteB.Domain)
		} else if siteErr != nil && !isNotFound(siteErr) {
			return CreateOrderResponse{}, siteErr
		}
	}
	usdDiff := a.calcMerchantUsdDiff(req.Amount, req.Currency, merchant.RateDiff)
	order := &model.Order{
		ID:               orderID,
		MerchantOrder:    merchantOrder,
		MerchantSite:     siteDomain,
		MerchantID:       merchant.ID,
		Channel:          channel.Name,
		ChannelAccountID: account.ID,
		SiteBID:          siteBID,
		SiteB:            siteBDomain,
		Provider:         platform.Code,
		Amount:           req.Amount,
		Currency:         strings.ToLower(strings.TrimSpace(req.Currency)),
		ReturnURL:        strings.TrimSpace(req.ReturnURL),
		NotifyURL:        strings.TrimSpace(req.NotifyURL),
		NotifyVerify:     string(verifyJSON),
		UsdDiff:          &usdDiff,
		Status:           model.OrderStatusCreated,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := a.store.SaveOrder(order); err != nil {
		a.appendOrderLogFromOrder(order, model.OrderLogTypeFailLog, "订单落库失败: "+err.Error())
		return CreateOrderResponse{}, err
	}
	a.appendOrderLogFromOrder(order, model.OrderLogTypeRecord, "订单已创建，开始请求通道创建支付会话")

	checkoutURL, providerRef, err := a.createGatewayCheckout(platform, account, order, req.Subject, req.ResultURL, req.CancelURL)
	if err != nil {
		order.Status = model.OrderStatusFailed
		order.ErrorMessage = err.Error()
		order.UpdatedAt = time.Now().UTC()
		_ = a.store.SaveOrder(order)
		a.appendOrderLogFromOrder(order, model.OrderLogTypeFailLog, "创建支付会话失败: "+err.Error())
		a.appendOrderLogFromOrder(order, model.OrderLogTypeFailRecord, "更新订单状态为 failed")
		return CreateOrderResponse{}, err
	}

	order.Status = model.OrderStatusPending
	order.CheckoutURL = checkoutURL
	order.ProviderRef = providerRef
	order.UpdatedAt = time.Now().UTC()
	if err := a.store.SaveOrder(order); err != nil {
		a.appendOrderLogFromOrder(order, model.OrderLogTypeFailLog, "更新待支付状态失败: "+err.Error())
		return CreateOrderResponse{}, err
	}
	a.appendOrderLogFromOrder(order, model.OrderLogTypeRecord,
		`开始跳转链接:{"code":2,"pay_code":2,"pay_skip_url":"`+checkoutURL+`","pay_html_type":1}`)
	a.appendOrderLogFromOrder(order, model.OrderLogTypeRecord, "更新订单状态为 pending")
	return CreateOrderResponse{
		OrderID:     order.ID,
		Status:      string(order.Status),
		CheckoutURL: order.CheckoutURL,
	}, nil
}
