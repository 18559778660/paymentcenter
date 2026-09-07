package service

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"

	"paymentcenter/internal/model"
)

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

// stripeCheckoutSessionEventTypes 需要处理的 Checkout Session 事件。
var stripeCheckoutSessionEventTypes = map[string]struct{}{
	"checkout.session.completed":           {},
	"checkout.session.expired":             {},
	"checkout.session.async_payment_failed": {},
}

// peekStripeWebhookOrderID 从 Stripe Webhook 事件中提取订单 ID。
func peekStripeWebhookOrderID(payload []byte) (string, error) {
	var thin stripeWebhookEventThin
	if err := json.Unmarshal(payload, &thin); err != nil {
		return "", err
	}
	if _, ok := stripeCheckoutSessionEventTypes[thin.Type]; !ok {
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

	// Webhook endpoint 的 API 版本可能与 stripe-go 不一致，验签后自行解析所需字段即可。
	event, err := webhook.ConstructEventWithOptions(payload, signature, webhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		return err
	}
	switch event.Type {
	case "checkout.session.completed":
		return a.applyStripeCheckoutPaid(order, account, event.Data.Raw)
	case "checkout.session.expired":
		return a.applyStripeCheckoutTerminal(order, event.Data.Raw, model.OrderStatusCancelled, "Checkout session expired")
	case "checkout.session.async_payment_failed":
		return a.applyStripeCheckoutTerminal(order, event.Data.Raw, model.OrderStatusFailed, "Async payment failed")
	default:
		return nil
	}
}

// applyStripeCheckoutPaid 将会话标记为已支付、回填 Stripe 手续费并通知商户站。
func (a *App) applyStripeCheckoutPaid(order *model.Order, account *model.ChannelAccount, raw json.RawMessage) error {
	var sess stripe.CheckoutSession
	if err := json.Unmarshal(raw, &sess); err != nil {
		return err
	}
	sessionID := firstNonEmpty(sess.ID, order.ProviderRef)
	fee := a.lookupStripeCheckoutFee(account, sessionID)

	// 已支付：仅在尚未写入手续费时补 fee，避免重复通知 A 站。
	if order.Status == model.OrderStatusPaid {
		if order.Fee > 0 {
			return nil
		}
		if fee > 0 {
			order.Fee = fee
			order.ProviderRef = firstNonEmpty(order.ProviderRef, sessionID)
			order.UpdatedAt = time.Now().UTC()
			return a.store.SaveOrder(order)
		}
		a.scheduleStripeFeeBackfill(order.ID, order.ChannelAccountID, sessionID)
		return nil
	}

	order.Status = model.OrderStatusPaid
	order.ProviderRef = sessionID
	if fee > 0 {
		order.Fee = fee
	}
	order.ErrorMessage = ""
	order.UpdatedAt = time.Now().UTC()
	if err := a.store.SaveOrder(order); err != nil {
		return err
	}
	if order.Fee <= 0 {
		a.scheduleStripeFeeBackfill(order.ID, order.ChannelAccountID, sessionID)
	}
	return a.notifyMerchantSite(order)
}

// lookupStripeCheckoutFee 同步快速查询一次手续费；查不到返回 0，由异步补查承接。
func (a *App) lookupStripeCheckoutFee(account *model.ChannelAccount, sessionID string) int64 {
	secretKey, err := requireStripeSecretKey(account)
	if err != nil {
		return 0
	}
	fee, err := fetchStripeCheckoutFee(secretKey, sessionID)
	if err != nil || fee <= 0 {
		return 0
	}
	return fee
}

// scheduleStripeFeeBackfill 支付成功后若手续费尚未就绪，后台延迟二次/三次补查并落库。
func (a *App) scheduleStripeFeeBackfill(orderID string, accountID uint, sessionID string) {
	orderID = strings.TrimSpace(orderID)
	sessionID = strings.TrimSpace(sessionID)
	if orderID == "" || accountID == 0 || sessionID == "" {
		return
	}
	go func() {
		delays := []time.Duration{3 * time.Second, 8 * time.Second, 20 * time.Second}
		for _, wait := range delays {
			time.Sleep(wait)
			order, err := a.store.GetOrder(orderID)
			if err != nil {
				return
			}
			if order.Fee > 0 {
				return
			}
			account, err := a.store.GetChannelAccountByID(accountID)
			if err != nil {
				continue
			}
			fee := a.lookupStripeCheckoutFee(account, firstNonEmpty(sessionID, order.ProviderRef))
			if fee <= 0 {
				continue
			}
			order.Fee = fee
			order.ProviderRef = firstNonEmpty(order.ProviderRef, sessionID)
			order.UpdatedAt = time.Now().UTC()
			_ = a.store.SaveOrder(order)
			return
		}
	}()
}

// applyStripeCheckoutTerminal 将会话标记为终态（取消/失败）并通知商户站。
func (a *App) applyStripeCheckoutTerminal(order *model.Order, raw json.RawMessage, status model.OrderStatus, message string) error {
	var sess stripe.CheckoutSession
	if err := json.Unmarshal(raw, &sess); err != nil {
		return err
	}
	// 已支付不降级；已是失败/取消则幂等跳过。
	if order.Status == model.OrderStatusPaid {
		return nil
	}
	if order.Status == model.OrderStatusFailed || order.Status == model.OrderStatusCancelled {
		return nil
	}
	order.Status = status
	order.ProviderRef = firstNonEmpty(order.ProviderRef, sess.ID)
	order.ErrorMessage = message
	order.UpdatedAt = time.Now().UTC()
	if err := a.store.SaveOrder(order); err != nil {
		return err
	}
	return a.notifyMerchantSite(order)
}
