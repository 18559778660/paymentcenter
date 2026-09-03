package service

import (
	"errors"
	"log"
	"strings"

	"paymentcenter/internal/model"
)

// validateGatewayMerchant 校验商户密钥与状态。
func (a *App) validateGatewayMerchant(secretKey string) (*model.Merchant, error) {
	secretKey = strings.TrimSpace(secretKey)
	if secretKey == "" {
		log.Printf("gateway merchant auth failed: empty Api-Token")
		return nil, ErrGatewayMerchantInvalid
	}
	merchant, err := a.store.FindMerchantBySecretKey(secretKey)
	if err != nil {
		if isNotFound(err) {
			log.Printf("gateway merchant auth failed: secret_key not found token_len=%d", len(secretKey))
			return nil, ErrGatewayMerchantInvalid
		}
		return nil, err
	}
	if merchant.Status != model.MerchantStatusEnabled {
		log.Printf("gateway merchant auth failed: merchant_id=%d status=%d", merchant.ID, merchant.Status)
		return nil, ErrGatewayMerchantInvalid
	}
	return merchant, nil
}

// validateGatewaySiteA 校验 A 站归属与审核状态，返回规范化域名。
func (a *App) validateGatewaySiteA(merchant *model.Merchant, merchantSite string) (string, error) {
	siteDomain, err := normalizeSiteADomain(merchantSite)
	if err != nil {
		log.Printf("gateway site auth failed: invalid domain raw=%q", merchantSite)
		return "", ErrGatewaySiteInvalid
	}
	siteA, err := a.store.FindSiteAByDomain(siteDomain)
	if err != nil {
		if isNotFound(err) {
			log.Printf("gateway site auth failed: site not found domain=%s merchant_id=%d", siteDomain, merchant.ID)
			return "", ErrGatewaySiteInvalid
		}
		return "", err
	}
	if siteA.MerchantID != merchant.ID {
		log.Printf(
			"gateway site auth failed: merchant mismatch domain=%s merchant_id=%d site_merchant_id=%d",
			siteDomain, merchant.ID, siteA.MerchantID,
		)
		return "", ErrGatewaySiteInvalid
	}
	if siteA.Status != model.SiteAStatusAudited {
		log.Printf(
			"gateway site auth failed: site not audited domain=%s merchant_id=%d status=%s",
			siteDomain, merchant.ID, siteA.Status,
		)
		return "", ErrGatewaySiteInvalid
	}
	return siteDomain, nil
}

// validateGatewayChannel 校验通道与平台是否可用。
func (a *App) validateGatewayChannel(channel *model.Channel) (*model.Platform, error) {
	if channel.Status != model.ChannelStatusEnabled {
		return nil, ErrGatewayChannelDisabled
	}
	platform, err := a.getPlatformByID(channel.PlatformID)
	if err != nil {
		if errors.Is(err, ErrChannelPlatformInvalid) {
			return nil, ErrGatewayPlatformUnsupported
		}
		return nil, err
	}
	return platform, nil
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
