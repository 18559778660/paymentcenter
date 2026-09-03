package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"paymentcenter/internal/model"
)

const (
	shopyyOrderStatusPaid   = 3
	shopyyOrderStatusFailed = 4
	shopyyNotifyCodeOK      = 1
)

// shopyyAsyncNotifyRequest Shopyy 异步回调请求体。
type shopyyAsyncNotifyRequest struct {
	OrderNumber string   `json:"order_number"`
	OrderStatus string   `json:"order_status"`
	PayType     *string  `json:"pay_type,omitempty"`
	Remark      string   `json:"remark"`
	Verify      []string `json:"verify,omitempty"`
}

// shopyyNotifyResponse Shopyy 通知响应。
type shopyyNotifyResponse struct {
	Code json.RawMessage `json:"code"`
	Msg  string          `json:"msg"`
}

// notifyMerchantSite 异步通知 Shopyy asyn_url 更新订单状态。
func (a *App) notifyMerchantSite(order *model.Order) error {
	notifyURL := strings.TrimSpace(order.NotifyURL)
	if notifyURL == "" {
		return nil
	}

	secret, err := a.merchantSecretForOrder(order)
	if err != nil {
		return err
	}
	// 解码 Shopyy 通知 verify 快照。
	snapshot, err := decodeShopyyNotifyVerifySnapshot(order.NotifyVerify)
	if err != nil {
		return err
	}
	serialNo := strings.TrimSpace(order.ProviderRef)
	if serialNo == "" {
		serialNo = order.ID
	}
	verify, err := encodeShopyyNotifyVerify(snapshot, serialNo, secret)
	if err != nil {
		return err
	}

	orderStatus, remark := shopyyNotifyPayload(order)
	body, err := json.Marshal(shopyyAsyncNotifyRequest{
		OrderNumber: strings.TrimSpace(order.MerchantOrder),
		OrderStatus: orderStatus,
		Remark:      remark,
		Verify:      []string{verify},
	})
	if err != nil {
		return err
	}
	log.Printf(
		"shopyy notify request order_id=%s merchant_order=%s url=%s body=%s",
		order.ID,
		order.MerchantOrder,
		notifyURL,
		truncateNotifyBody(body),
	)

	req, err := http.NewRequest(http.MethodPost, notifyURL, bytes.NewReader(body))
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Printf(
		"shopyy notify response order_id=%s merchant_order=%s status=%d body=%s",
		order.ID,
		order.MerchantOrder,
		resp.StatusCode,
		truncateNotifyBody(respBody),
	)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("shopyy notify failed: status %d body %s", resp.StatusCode, truncateNotifyBody(respBody))
	}
	if err := parseShopyyNotifyResponse(respBody); err != nil {
		return err
	}
	return nil
}

// merchantSecretForOrder 获取订单所属商户密钥。
func (a *App) merchantSecretForOrder(order *model.Order) (string, error) {
	if order.MerchantID > 0 {
		merchant, err := a.store.GetMerchantByID(order.MerchantID)
		if err != nil {
			return "", err
		}
		secret := strings.TrimSpace(merchant.SecretKey)
		if secret == "" {
			return "", fmt.Errorf("merchant secret missing")
		}
		return secret, nil
	}
	site, err := a.store.FindSiteAByDomain(order.MerchantSite)
	if err != nil {
		return "", fmt.Errorf("merchant secret missing")
	}
	merchant, err := a.store.GetMerchantByID(site.MerchantID)
	if err != nil {
		return "", err
	}
	secret := strings.TrimSpace(merchant.SecretKey)
	if secret == "" {
		return "", fmt.Errorf("merchant secret missing")
	}
	return secret, nil
}

// shopyyNotifyPayload 转换 Shopyy 异步通知状态与备注。
func shopyyNotifyPayload(order *model.Order) (orderStatus, remark string) {
	switch order.Status {
	case model.OrderStatusPaid:
		return strconv.Itoa(shopyyOrderStatusPaid), ""
	default:
		msg := strings.TrimSpace(order.ErrorMessage)
		if msg == "" {
			msg = "Payment failed"
		}
		return strconv.Itoa(shopyyOrderStatusFailed), msg
	}
}

// parseShopyyNotifyResponse 解析 Shopyy 通知响应。
func parseShopyyNotifyResponse(body []byte) error {
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return fmt.Errorf("shopyy notify empty response")
	}

	var resp shopyyNotifyResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("shopyy notify invalid response: %s", truncateNotifyBody(body))
	}
	code, err := parseShopyyNotifyCode(resp.Code)
	if err != nil {
		return fmt.Errorf("shopyy notify invalid code: %s", truncateNotifyBody(body))
	}
	if code != shopyyNotifyCodeOK {
		msg := strings.TrimSpace(resp.Msg)
		if msg == "" {
			msg = "unknown error"
		}
		return fmt.Errorf("shopyy notify rejected: %s", msg)
	}
	return nil
}

// parseShopyyNotifyCode 解析 Shopyy 通知响应码。
func parseShopyyNotifyCode(raw json.RawMessage) (int, error) {
	if len(raw) == 0 {
		return 0, fmt.Errorf("missing code")
	}
	var asInt int
	if err := json.Unmarshal(raw, &asInt); err == nil {
		return asInt, nil
	}
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		asString = strings.TrimSpace(asString)
		if asString == "" {
			return 0, fmt.Errorf("empty code")
		}
		var n int
		if _, err := fmt.Sscanf(asString, "%d", &n); err != nil {
			return 0, err
		}
		return n, nil
	}
	return 0, fmt.Errorf("unsupported code type")
}

// truncateNotifyBody 截断通知响应体。
func truncateNotifyBody(body []byte) string {
	const maxLen = 256
	s := strings.TrimSpace(string(body))
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
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
