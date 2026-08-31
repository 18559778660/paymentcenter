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
	ErrGatewayMerchantInvalid     = errors.New("gateway merchant invalid")
	ErrGatewaySiteInvalid         = errors.New("gateway site invalid")
	ErrGatewayRouteInvalid        = errors.New("gateway route invalid")
	ErrGatewayChannelDisabled     = errors.New("gateway channel disabled")
	ErrGatewayAccountUnavailable  = errors.New("gateway account unavailable")
	ErrGatewayPlatformUnsupported = errors.New("gateway platform unsupported")
	ErrGatewayStripeFailed        = errors.New("gateway stripe failed")
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

// GatewayPay A 站经网关创建 Stripe Checkout 并落库订单。
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
	platform, err := a.getPlatformByID(channel.PlatformID)
	if err != nil || platform.Code != model.PlatformCodeStripe {
		return CreateOrderResponse{}, ErrGatewayPlatformUnsupported
	}
	if channel.Status != model.ChannelStatusEnabled {
		return CreateOrderResponse{}, ErrGatewayChannelDisabled
	}

	stripeKey := resolveStripeSecretKey(account, a.stripeAPIKey)
	now := time.Now().UTC()
	orderID := "pc_" + strconv.FormatInt(now.UnixNano(), 10)
	order := &model.Order{
		ID:            orderID,
		MerchantOrder: strings.TrimSpace(req.MerchantOrder),
		MerchantSite:  siteDomain,
		Channel:       channel.Name,
		Provider:      model.PlatformCodeStripe,
		Amount:        req.Amount,
		Currency:      strings.ToLower(strings.TrimSpace(req.Currency)),
		ReturnURL:     strings.TrimSpace(req.ReturnURL),
		NotifyURL:     strings.TrimSpace(req.NotifyURL),
		Status:        model.OrderStatusCreated,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := a.store.SaveOrder(order); err != nil {
		return CreateOrderResponse{}, err
	}

	checkout, err := createStripeCheckoutSession(stripeCheckoutInput{
		SecretKey:     stripeKey,
		OrderID:       order.ID,
		MerchantOrder: order.MerchantOrder,
		Amount:        order.Amount,
		Currency:      order.Currency,
		Subject:       req.Subject,
		ReturnURL:     order.ReturnURL,
	})
	if err != nil {
		order.Status = model.OrderStatusFailed
		order.ErrorMessage = err.Error()
		order.UpdatedAt = time.Now().UTC()
		_ = a.store.SaveOrder(order)
		return CreateOrderResponse{}, fmt.Errorf("%w: %v", ErrGatewayStripeFailed, err)
	}

	order.Status = model.OrderStatusPending
	order.CheckoutURL = checkout.CheckoutURL
	order.ProviderRef = checkout.SessionID
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
	return &enabled[rand.Intn(len(enabled))], nil
}

// resolveStripeSecretKey 解析 Stripe 密钥。
func resolveStripeSecretKey(account *model.ChannelAccount, fallback string) string {
	if account != nil {
		for _, key := range []string{account.PrivateKey, account.WebSecret, account.AppID} {
			key = strings.TrimSpace(key)
			if strings.HasPrefix(key, "sk_") {
				return key
			}
		}
	}
	return strings.TrimSpace(fallback)
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

// HandleStripeWebhook 处理 Stripe 回调，更新订单并通知 A 站。
func (a *App) HandleStripeWebhook(payload []byte, signature string) error {
	if a.stripeWebhookSecret == "" {
		return fmt.Errorf("stripe webhook secret not configured")
	}
	event, err := webhook.ConstructEvent(payload, signature, a.stripeWebhookSecret)
	if err != nil {
		return err
	}
	switch event.Type {
	case "checkout.session.completed":
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			return err
		}
		orderID := strings.TrimSpace(sess.ClientReferenceID)
		if orderID == "" && sess.Metadata != nil {
			orderID = strings.TrimSpace(sess.Metadata["order_id"])
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
