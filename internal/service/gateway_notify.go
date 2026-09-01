package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"paymentcenter/internal/model"
)

// notifyMerchantSite 通知 A 站支付结果。
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

// firstNonEmpty 返回第一个非空字符串。
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
