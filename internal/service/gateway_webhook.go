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
