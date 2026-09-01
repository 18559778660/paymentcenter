package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
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

// GatewayPayRequest A 站经网关发起支付。
type GatewayPayRequest struct {
	MerchantOrder string `json:"merchant_order" binding:"required"`
	MerchantSite  string `json:"merchant_site" binding:"required"`
	Amount        int64  `json:"amount" binding:"required"`
	Currency      string `json:"currency" binding:"required"`
	ReturnURL     string `json:"return_url" binding:"required"`
	NotifyURL     string `json:"notify_url" binding:"required"`
	Subject       string `json:"subject"`
	SecretKey     string `json:"secret_key"`
}

// GatewayPayQuery 网关路由参数。
type GatewayPayQuery struct {
	Channel string
	Group   string
}

// GatewayPay A 站经网关创建支付并落库订单，按通道所属平台分发。
func (a *App) GatewayPay(req GatewayPayRequest, q GatewayPayQuery, secretKey string) (CreateOrderResponse, error) {
	secretKey = firstNonEmpty(strings.TrimSpace(secretKey), strings.TrimSpace(req.SecretKey))
	if secretKey == "" {
		return CreateOrderResponse{}, ErrGatewayMerchantInvalid
	}
	merchant, err := a.store.FindMerchantBySecretKey(secretKey)
	if err != nil {
		if isNotFound(err) {
			return CreateOrderResponse{}, ErrGatewayMerchantInvalid
		}
		return CreateOrderResponse{}, err
	}
	if merchant.Status != model.MerchantStatusEnabled {
		return CreateOrderResponse{}, ErrGatewayMerchantInvalid
	}
	siteDomain, err := normalizeSiteADomain(req.MerchantSite)
	if err != nil {
		return CreateOrderResponse{}, ErrGatewaySiteInvalid
	}
	siteA, err := a.store.FindSiteAByDomain(siteDomain)
	if err != nil {
		if isNotFound(err) {
			return CreateOrderResponse{}, ErrGatewaySiteInvalid
		}
		return CreateOrderResponse{}, err
	}
	if siteA.MerchantID != merchant.ID {
		return CreateOrderResponse{}, ErrGatewaySiteInvalid
	}
	if siteA.Status != model.SiteAStatusAudited {
		return CreateOrderResponse{}, ErrGatewaySiteInvalid
	}

	channel, account, err := a.resolveGatewayRoute(q)
	if err != nil {
		return CreateOrderResponse{}, err
	}
	if channel.Status != model.ChannelStatusEnabled {
		return CreateOrderResponse{}, ErrGatewayChannelDisabled
	}
	platform, err := a.getPlatformByID(channel.PlatformID)
	if err != nil {
		if errors.Is(err, ErrChannelPlatformInvalid) {
			return CreateOrderResponse{}, ErrGatewayPlatformUnsupported
		}
		return CreateOrderResponse{}, err
	}
	if err := validateGatewayAccount(platform, account); err != nil {
		return CreateOrderResponse{}, err
	}

	now := time.Now().UTC()
	orderID := "pc_" + strconv.FormatInt(now.UnixNano(), 10)
	order := &model.Order{
		ID:               orderID,
		MerchantOrder:    strings.TrimSpace(req.MerchantOrder),
		MerchantSite:     siteDomain,
		Channel:          channel.Name,
		ChannelAccountID: account.ID,
		Provider:         platform.Code,
		Amount:           req.Amount,
		Currency:         strings.ToLower(strings.TrimSpace(req.Currency)),
		ReturnURL:        strings.TrimSpace(req.ReturnURL),
		NotifyURL:        strings.TrimSpace(req.NotifyURL),
		Status:           model.OrderStatusCreated,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := a.store.SaveOrder(order); err != nil {
		return CreateOrderResponse{}, err
	}

	checkoutURL, providerRef, err := a.createGatewayCheckout(platform, account, order, req.Subject)
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

// validateGatewayAccount 按平台校验通道账号配置是否完整。
func validateGatewayAccount(platform *model.Platform, account *model.ChannelAccount) error {
	switch platform.Code {
	case model.PlatformCodeStripe:
		if _, err := requireStripeSecretKey(account); err != nil {
			return err
		}
		if _, err := requireStripeWebhookSecret(account); err != nil {
			return err
		}
		return nil
	default:
		return ErrGatewayPlatformUnsupported
	}
}

// createGatewayCheckout 按平台创建支付会话。
func (a *App) createGatewayCheckout(platform *model.Platform, account *model.ChannelAccount, order *model.Order, subject string) (checkoutURL, providerRef string, err error) {
	switch platform.Code {
	case model.PlatformCodeStripe:
		stripeKey, err := requireStripeSecretKey(account)
		if err != nil {
			return "", "", err
		}
		checkout, err := createStripeCheckoutSession(stripeCheckoutInput{
			SecretKey:     stripeKey,
			OrderID:       order.ID,
			MerchantOrder: order.MerchantOrder,
			Amount:        order.Amount,
			Currency:      order.Currency,
			Subject:       subject,
			ReturnURL:     order.ReturnURL,
		})
		if err != nil {
			return "", "", fmt.Errorf("%w: %v", ErrGatewayStripeFailed, err)
		}
		return checkout.CheckoutURL, checkout.SessionID, nil
	default:
		return "", "", ErrGatewayPlatformUnsupported
	}
}

// resolveGatewayRoute 解析网关路由。
func (a *App) resolveGatewayRoute(q GatewayPayQuery) (*model.Channel, *model.ChannelAccount, error) {
	channelName := strings.TrimSpace(q.Channel)
	groupCode := strings.TrimSpace(q.Group)
	if channelName == "" && groupCode == "" {
		return nil, nil, ErrGatewayRouteInvalid
	}
	if channelName != "" && groupCode != "" {
		return nil, nil, ErrGatewayRouteInvalid
	}
	if channelName != "" {
		channel, err := a.store.FindChannelByName(channelName)
		if err != nil {
			if isNotFound(err) {
				return nil, nil, ErrGatewayRouteInvalid
			}
			return nil, nil, err
		}
		account, err := a.pickChannelAccount(channel.ID, nil)
		if err != nil {
			return nil, nil, err
		}
		return channel, account, nil
	}
	group, err := a.store.FindChannelGroupByCode(groupCode)
	if err != nil {
		if isNotFound(err) {
			return nil, nil, ErrGatewayRouteInvalid
		}
		return nil, nil, err
	}
	accountIDs, err := a.store.ListChannelGroupMemberAccountIDs(group.ID)
	if err != nil {
		return nil, nil, err
	}
	account, err := a.pickChannelAccount(0, accountIDs)
	if err != nil {
		return nil, nil, err
	}
	channel, err := a.store.GetChannelByID(account.ChannelID)
	if err != nil {
		if isNotFound(err) {
			return nil, nil, ErrGatewayRouteInvalid
		}
		return nil, nil, err
	}
	// 返回通道和账号。
	return channel, account, nil
}

// pickChannelAccount 选择通道账号。
func (a *App) pickChannelAccount(channelID uint, accountIDs []uint) (*model.ChannelAccount, error) {
	var accounts []model.ChannelAccount
	var err error
	if channelID > 0 {
		cid := channelID
		accounts, err = a.store.ListChannelAccounts(store.ChannelAccountListFilter{
			ChannelID: &cid,
		})
	} else {
		all, listErr := a.store.ListChannelAccounts(store.ChannelAccountListFilter{})
		if listErr != nil {
			return nil, listErr
		}
		idSet := make(map[uint]struct{}, len(accountIDs))
		for _, id := range accountIDs {
			idSet[id] = struct{}{}
		}
		for _, item := range all {
			if _, ok := idSet[item.ID]; ok {
				accounts = append(accounts, item)
			}
		}
	}
	if err != nil {
		return nil, err
	}
	enabled := make([]model.ChannelAccount, 0, len(accounts))
	for _, item := range accounts {
		if item.Status == model.ChannelAccountStatusEnabled {
			enabled = append(enabled, item)
		}
	}
	if len(enabled) == 0 {
		return nil, ErrGatewayAccountUnavailable
	}
	// 随机选择一个通道账号。
	return &enabled[rand.Intn(len(enabled))], nil
}

// requireStripeSecretKey 取通道账号 Stripe Secret Key（private_key），未配置则报错。
func requireStripeSecretKey(account *model.ChannelAccount) (string, error) {
	if account == nil {
		return "", ErrGatewayAccountStripeKeyMissing
	}
	key := strings.TrimSpace(account.PrivateKey)
	if key == "" {
		return "", ErrGatewayAccountStripeKeyMissing
	}
	return key, nil
}

// requireStripeWebhookSecret 取通道账号 Stripe Webhook Secret（web_secret），未配置则报错。
func requireStripeWebhookSecret(account *model.ChannelAccount) (string, error) {
	if account == nil {
		return "", ErrGatewayAccountWebhookSecretMissing
	}
	secret := strings.TrimSpace(account.WebSecret)
	if secret == "" {
		return "", ErrGatewayAccountWebhookSecretMissing
	}
	return secret, nil
}

// stripeCheckoutSessionEvent Stripe Checkout 会话事件。
type stripeCheckoutSessionEvent struct {
	ClientReferenceID string            `json:"client_reference_id"`
	Metadata          map[string]string `json:"metadata"`
}

// stripeWebhookEventThin Stripe Webhook 事件精简版。
type stripeWebhookEventThin struct {
	Type string `json:"type"`
	Data struct {
		Object json.RawMessage `json:"object"`
	} `json:"data"`
}

// firstNonEmpty 返回第一个非空字符串。
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// peekStripeWebhookOrderID 从 Stripe Webhook 事件中提取订单 ID。
func peekStripeWebhookOrderID(payload []byte) (string, error) {
	var thin stripeWebhookEventThin
	if err := json.Unmarshal(payload, &thin); err != nil {
		return "", err
	}
	if thin.Type != "checkout.session.completed" {
		return "", nil
	}
	var sess stripeCheckoutSessionEvent
	if err := json.Unmarshal(thin.Data.Object, &sess); err != nil {
		return "", err
	}
	orderID := strings.TrimSpace(sess.ClientReferenceID)
	if orderID == "" && sess.Metadata != nil {
		orderID = strings.TrimSpace(sess.Metadata["order_id"])
	}
	return orderID, nil
}

// HandleStripeWebhook 处理 Stripe 回调，按订单关联的通道账号 web_secret 验签。
func (a *App) HandleStripeWebhook(payload []byte, signature string) error {
	orderID, err := peekStripeWebhookOrderID(payload)
	if err != nil {
		return err
	}
	if orderID == "" {
		return nil
	}
	order, err := a.store.GetOrder(orderID)
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return err
	}
	if order.ChannelAccountID == 0 {
		return ErrGatewayAccountWebhookSecretMissing
	}
	account, err := a.store.GetChannelAccountByID(order.ChannelAccountID)
	if err != nil {
		if isNotFound(err) {
			return ErrGatewayAccountWebhookSecretMissing
		}
		return err
	}
	webhookSecret, err := requireStripeWebhookSecret(account)
	if err != nil {
		return err
	}
	event, err := webhook.ConstructEvent(payload, signature, webhookSecret)
	if err != nil {
		return err
	}
	switch event.Type {
	case "checkout.session.completed":
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			return err
		}
		if order.Status == model.OrderStatusPaid {
			return nil
		}
		order.Status = model.OrderStatusPaid
		order.ProviderRef = sess.ID
		order.UpdatedAt = time.Now().UTC()
		if err := a.store.SaveOrder(order); err != nil {
			return err
		}
		return a.notifyMerchantSite(order)
	default:
		return nil
	}
}

// notifyMerchantSite 通知商户站点。
func (a *App) notifyMerchantSite(order *model.Order) error {
	if strings.TrimSpace(order.NotifyURL) == "" {
		return nil
	}
	body, err := json.Marshal(map[string]interface{}{
		"order_id":       order.ID,
		"merchant_order": order.MerchantOrder,
		"merchant_site":  order.MerchantSite,
		"status":         order.Status,
		"provider":       order.Provider,
		"provider_ref":   order.ProviderRef,
		"amount":         order.Amount,
		"currency":       order.Currency,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, order.NotifyURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("merchant notify failed: status %d", resp.StatusCode)
	}
	return nil
}
