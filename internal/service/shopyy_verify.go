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

type shopyyVerifyPayload struct {
	OrderNumber  string `json:"orderNumber"`
	PayUSDAmount string `json:"pay_usd_amount"`
	PayCurrency  string `json:"pay_currency"`
	PayAmount    string `json:"pay_amount"`
	OrderTime    string `json:"order_time"`
	Domain       string `json:"domain"`
	Email        string `json:"email"`
	CustomerIP   string `json:"customerIp"`
	SerialNo     string `json:"serial_no"`
	Signature    string `json:"signature"`
}

func formatShopyyPayAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

func encodeShopyyNotifyVerify(snapshot ShopyyNotifyVerifySnapshot, serialNo, secret string) (string, error) {
	serialNo = strings.TrimSpace(serialNo)
	signParams := map[string]string{
		"orderNumber":    strings.TrimSpace(snapshot.OrderNumber),
		"pay_usd_amount": strings.TrimSpace(snapshot.PayUSDAmount),
		"pay_currency":   strings.TrimSpace(snapshot.PayCurrency),
		"pay_amount":     strings.TrimSpace(snapshot.PayAmount),
		"order_time":     strings.TrimSpace(snapshot.OrderTime),
		"domain":         strings.TrimSpace(snapshot.Domain),
		"email":          strings.TrimSpace(snapshot.Email),
		"customerIp":     strings.TrimSpace(snapshot.CustomerIP),
		"serial_no":      serialNo,
	}
	payload := shopyyVerifyPayload{
		OrderNumber:  signParams["orderNumber"],
		PayUSDAmount: signParams["pay_usd_amount"],
		PayCurrency:  signParams["pay_currency"],
		PayAmount:    signParams["pay_amount"],
		OrderTime:    signParams["order_time"],
		Domain:       signParams["domain"],
		Email:        signParams["email"],
		CustomerIP:   signParams["customerIp"],
		SerialNo:     serialNo,
	}
	payload.Signature = shopyyGetSign(signParams, secret)

	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(raw), nil
}

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
