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
	if secretKey == "" {
		return CreateOrderResponse{}, ErrGatewayMerchantInvalid
	}
	merchant, err := a.validateGatewayMerchant(secretKey)
	if err != nil {
		return CreateOrderResponse{}, err
	}
	siteDomain, err := a.validateGatewaySiteA(merchant, req.MerchantSite)
	if err != nil {
		return CreateOrderResponse{}, err
	}

	channel, account, err := a.resolveGatewayRoute(q)
	if err != nil {
		return CreateOrderResponse{}, err
	}
	platform, err := a.validateGatewayChannel(channel)
	if err != nil {
		return CreateOrderResponse{}, err
	}
	if err := validateGatewayAccount(platform, account); err != nil {
		return CreateOrderResponse{}, err
	}

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
	order := &model.Order{
		ID:               orderID,
		MerchantOrder:    strings.TrimSpace(req.MerchantOrder),
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
		Status:           model.OrderStatusCreated,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := a.store.SaveOrder(order); err != nil {
		return CreateOrderResponse{}, err
	}

	checkoutURL, providerRef, err := a.createGatewayCheckout(platform, account, order, req.Subject, req.ResultURL, req.CancelURL)
	if err != nil {
		order.Status = model.OrderStatusFailed
		order.ErrorMessage = err.Error()
		order.UpdatedAt = time.Now().UTC()
		_ = a.store.SaveOrder(order)
		return CreateOrderResponse{}, err
	}

	order.Status = model.OrderStatusPending
	order.CheckoutURL = checkoutURL
	order.ProviderRef = providerRef
	order.UpdatedAt = time.Now().UTC()
	if err := a.store.SaveOrder(order); err != nil {
		return CreateOrderResponse{}, err
	}
	return CreateOrderResponse{
		OrderID:     order.ID,
		Status:      string(order.Status),
		CheckoutURL: order.CheckoutURL,
	}, nil
}
