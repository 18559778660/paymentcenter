package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// ShopyyNotifyVerifySnapshot 异步回调 verify 所需字段（下单时快照，通知时补 serial_no 与 signature）。
type ShopyyNotifyVerifySnapshot struct {
	OrderNumber  string `json:"orderNumber"`
	PayUSDAmount string `json:"pay_usd_amount"`
	PayCurrency  string `json:"pay_currency"`
	PayAmount    string `json:"pay_amount"`
	OrderTime    string `json:"order_time"`
	Domain       string `json:"domain"`
	Email        string `json:"email"`
	CustomerIP   string `json:"customerIp"`
}

func formatShopyyPayAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

// encodeShopyyNotifyVerify 编码 Shopyy 通知 verify 字段。
func encodeShopyyNotifyVerify(snapshot ShopyyNotifyVerifySnapshot, serialNo, secret string) (string, error) {
	verify := map[string]string{
		"orderNumber":    strings.TrimSpace(snapshot.OrderNumber),
		"pay_usd_amount": strings.TrimSpace(snapshot.PayUSDAmount),
		"pay_currency":   strings.TrimSpace(snapshot.PayCurrency),
		"pay_amount":     strings.TrimSpace(snapshot.PayAmount),
		"order_time":     strings.TrimSpace(snapshot.OrderTime),
		"domain":         strings.TrimSpace(snapshot.Domain),
		"email":          strings.TrimSpace(snapshot.Email),
		"customerIp":     strings.TrimSpace(snapshot.CustomerIP),
		"serial_no":      strings.TrimSpace(serialNo),
	}
	verify["signature"] = shopyyGetSign(verify, secret)

	raw, err := json.Marshal(verify)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(raw), nil
}

// decodeShopyyNotifyVerifySnapshot 解码 Shopyy 通知 verify 快照。
func decodeShopyyNotifyVerifySnapshot(raw string) (ShopyyNotifyVerifySnapshot, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ShopyyNotifyVerifySnapshot{}, fmt.Errorf("missing notify verify snapshot")
	}
	var snapshot ShopyyNotifyVerifySnapshot
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return ShopyyNotifyVerifySnapshot{}, err
	}
	return snapshot, nil
}
