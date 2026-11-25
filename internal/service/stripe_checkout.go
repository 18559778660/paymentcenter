package service

import (
	"fmt"
	"strings"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/balancetransaction"
	"github.com/stripe/stripe-go/v82/charge"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/paymentintent"
)

// 本文件是路1遗留：支付中心用 SK 直连创建 Stripe Checkout、再查手续费。
// createStripeCheckoutSession 已无调用方（建单改走 B 站 hop）。
// fetchStripeCheckoutFee 只被 gateway_webhook.go 的路1回调用。

// stripeCheckoutInput Stripe Checkout 输入参数。
type stripeCheckoutInput struct {
	SecretKey     string
	OrderID       string
	MerchantOrder string
	Amount        int64
	Currency      string
	Subject       string
	ResultURL     string
	CancelURL     string
}

// stripeCheckoutResult Stripe Checkout 输出结果。
type stripeCheckoutResult struct {
	SessionID   string
	CheckoutURL string
}

// createStripeCheckoutSession 路1遗留：支付中心直连创建 Checkout 会话。现已无调用。
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
	resultURL := strings.TrimSpace(in.ResultURL)
	if resultURL == "" {
		return nil, fmt.Errorf("result_url required")
	}
	cancelURL := strings.TrimSpace(in.CancelURL)
	if cancelURL == "" {
		return nil, fmt.Errorf("cancel_url required")
	}

	stripe.Key = secretKey
	params := &stripe.CheckoutSessionParams{
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		ClientReferenceID: stripe.String(in.OrderID),
		SuccessURL:        stripe.String(resultURL),
		CancelURL:         stripe.String(cancelURL),
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

// fetchStripeCheckoutFee 路1遗留：Checkout Session → PaymentIntent → Charge 读手续费。
func fetchStripeCheckoutFee(secretKey, sessionID string) (int64, error) {
	secretKey = strings.TrimSpace(secretKey)
	sessionID = strings.TrimSpace(sessionID)
	if secretKey == "" {
		return 0, fmt.Errorf("stripe secret key missing")
	}
	if sessionID == "" {
		return 0, fmt.Errorf("checkout session id missing")
	}

	stripe.Key = secretKey

	sessParams := &stripe.CheckoutSessionParams{}
	sessParams.AddExpand("payment_intent")
	sess, err := session.Get(sessionID, sessParams)
	if err != nil {
		return 0, err
	}
	if sess.PaymentIntent == nil || strings.TrimSpace(sess.PaymentIntent.ID) == "" {
		return 0, fmt.Errorf("checkout session missing payment_intent")
	}

	piParams := &stripe.PaymentIntentParams{}
	piParams.AddExpand("latest_charge.balance_transaction")
	pi, err := paymentintent.Get(sess.PaymentIntent.ID, piParams)
	if err != nil {
		return 0, err
	}
	if pi.LatestCharge == nil || strings.TrimSpace(pi.LatestCharge.ID) == "" {
		return 0, fmt.Errorf("payment_intent missing latest_charge")
	}

	fee := feeFromBalanceTransaction(pi.LatestCharge.BalanceTransaction)
	if fee > 0 {
		return fee, nil
	}

	// Charge 上可能只有 balance_transaction ID，再单独取一次。
	chParams := &stripe.ChargeParams{}
	chParams.AddExpand("balance_transaction")
	ch, err := charge.Get(pi.LatestCharge.ID, chParams)
	if err != nil {
		return 0, err
	}
	fee = feeFromBalanceTransaction(ch.BalanceTransaction)
	if fee > 0 {
		return fee, nil
	}
	if ch.BalanceTransaction == nil || strings.TrimSpace(ch.BalanceTransaction.ID) == "" {
		return 0, fmt.Errorf("stripe fee not available yet")
	}

	bt, err := balancetransaction.Get(ch.BalanceTransaction.ID, nil)
	if err != nil {
		return 0, err
	}
	fee = feeFromBalanceTransaction(bt)
	if fee <= 0 {
		return 0, fmt.Errorf("stripe fee not available yet")
	}
	return fee, nil
}

// feeFromBalanceTransaction 读取已展开的 balance_transaction.fee。
func feeFromBalanceTransaction(bt *stripe.BalanceTransaction) int64 {
	if bt == nil {
		return 0
	}
	return bt.Fee
}
