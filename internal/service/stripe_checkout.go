package service

import (
	"fmt"
	"strings"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
)

// stripeCheckoutInput Stripe Checkout 输入参数。
type stripeCheckoutInput struct {
	SecretKey     string
	OrderID       string
	MerchantOrder string
	Amount        int64
	Currency      string
	Subject       string
	ReturnURL     string
}

// stripeCheckoutResult Stripe Checkout 输出结果。
type stripeCheckoutResult struct {
	SessionID   string
	CheckoutURL string
}

// createStripeCheckoutSession 创建 Stripe Checkout 会话。
func createStripeCheckoutSession(in stripeCheckoutInput) (*stripeCheckoutResult, error) {
	secretKey := strings.TrimSpace(in.SecretKey)
	if secretKey == "" {
		return nil, fmt.Errorf("stripe secret key missing")
	}
	if in.Amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	currency := strings.ToLower(strings.TrimSpace(in.Currency))
	if currency == "" {
		return nil, fmt.Errorf("currency required")
	}
	subject := strings.TrimSpace(in.Subject)
	if subject == "" {
		subject = "Order " + in.MerchantOrder
	}
	returnURL := strings.TrimSpace(in.ReturnURL)
	if returnURL == "" {
		return nil, fmt.Errorf("return_url required")
	}
	successURL := returnURL
	if strings.Contains(successURL, "?") {
		successURL += "&order_id=" + in.OrderID
	} else {
		successURL += "?order_id=" + in.OrderID
	}

	stripe.Key = secretKey
	params := &stripe.CheckoutSessionParams{
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		ClientReferenceID: stripe.String(in.OrderID),
		SuccessURL:        stripe.String(successURL),
		CancelURL:         stripe.String(returnURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Quantity: stripe.Int64(1),
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:   stripe.String(currency),
					UnitAmount: stripe.Int64(in.Amount),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(subject),
					},
				},
			},
		},
		Metadata: map[string]string{
			"order_id":       in.OrderID,
			"merchant_order": in.MerchantOrder,
		},
	}
	sess, err := session.New(params)
	if err != nil {
		return nil, err
	}
	return &stripeCheckoutResult{
		SessionID:   sess.ID,
		CheckoutURL: sess.URL,
	}, nil
}
