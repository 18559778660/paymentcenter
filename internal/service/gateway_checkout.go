package service

import (
	"fmt"

	"paymentcenter/internal/model"
)

// gatewayCheckoutResult 通道创建支付跳转的结果。
type gatewayCheckoutResult struct {
	CheckoutURL string
	ProviderRef string
	PayCode     int64
	PayHTMLType int64
}

// createGatewayCheckout 按平台创建支付跳转。Stripe 走 B 站壳包 check.php，不再直连 Checkout。
func (a *App) createGatewayCheckout(platform *model.Platform, account *model.ChannelAccount, siteB *model.SiteB, order *model.Order, req GatewayPayRequest) (*gatewayCheckoutResult, error) {
	switch platform.Code {
	case model.PlatformCodeStripe:
		if _, err := requireStripeSecretKey(account); err != nil {
			return nil, err
		}
		if _, err := requireStripeWebhookSecret(account); err != nil {
			return nil, err
		}
		result, err := a.hopStripePayToSiteB(platform, account, siteB, order, req)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrGatewaySiteBFailed, err)
		}
		return result, nil
	default:
		return nil, ErrGatewayPlatformUnsupported
	}
}
