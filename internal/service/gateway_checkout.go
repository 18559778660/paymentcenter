package service

import (
	"fmt"

	"paymentcenter/internal/model"
)

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
