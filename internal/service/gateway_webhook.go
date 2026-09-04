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
		return a.applyStripeCheckoutPaid(order, event.Data.Raw)
	case "checkout.session.expired":
		return a.applyStripeCheckoutTerminal(order, event.Data.Raw, model.OrderStatusCancelled, "Checkout session expired")
	case "checkout.session.async_payment_failed":
		return a.applyStripeCheckoutTerminal(order, event.Data.Raw, model.OrderStatusFailed, "Async payment failed")
	default:
		return nil
	}
}

// applyStripeCheckoutPaid 将会话标记为已支付并通知商户站。
func (a *App) applyStripeCheckoutPaid(order *model.Order, raw json.RawMessage) error {
	var sess stripe.CheckoutSession
	if err := json.Unmarshal(raw, &sess); err != nil {
		return err
	}
	if order.Status == model.OrderStatusPaid {
		return nil
	}
	order.Status = model.OrderStatusPaid
	order.ProviderRef = sess.ID
	order.ErrorMessage = ""
	order.UpdatedAt = time.Now().UTC()
	if err := a.store.SaveOrder(order); err != nil {
		return err
	}
	return a.notifyMerchantSite(order)
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
