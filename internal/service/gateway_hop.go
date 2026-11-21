package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"paymentcenter/internal/model"
)

// 本文件把 Stripe 支付中转到 B 站壳包。
// 支付中心建单后 POST {B域名}/{运行目录}/check.php，解析壳包返回的
// pay_skip_url / pay_code / pay_html_type，再回给 A 站。
// 由 createGatewayCheckout 调用，不是独立对外接口。

// hopStripePayToSiteB 向 B 站壳包 POST check.php，用返回的 pay_skip_url 作为 A 站跳转地址。
func (a *App) hopStripePayToSiteB(platform *model.Platform, account *model.ChannelAccount, siteB *model.SiteB, order *model.Order, req GatewayPayRequest) (*gatewayCheckoutResult, error) {
	if siteB == nil || strings.TrimSpace(siteB.Domain) == "" {
		return nil, fmt.Errorf("通道账号未绑定可用 B 站")
	}
	if !siteB.Status {
		return nil, fmt.Errorf("B站已停用")
	}
	if strings.TrimSpace(a.gatewayBaseURL) == "" {
		return nil, fmt.Errorf("未配置 GATEWAY_BASE_URL，无法把回调地址传给 B 站")
	}

	paymentCode, err := channelPaymentCode(platform)
	if err != nil {
		return nil, err
	}

	sk, err := requireStripeSecretKey(account)
	if err != nil {
		return nil, err
	}
	wb, err := requireStripeWebhookSecret(account)
	if err != nil {
		return nil, err
	}

	checkURL := buildSiteBCheckURL(siteB.Domain, siteB.RunDirectory)
	sysURL := strings.TrimRight(a.gatewayBaseURL, "/") + "/api/gateway"
	getInfoURL := strings.TrimRight(a.gatewayBaseURL, "/") + "/api/gateway/getInfo"
	accessToken := a.issueSiteBAccessToken(order.ID)

	orderInfo, err := a.buildSiteBOrderInfo(account, order, req, sk, wb)
	if err != nil {
		return nil, err
	}
	orderInfoJSON, err := json.Marshal(orderInfo)
	if err != nil {
		return nil, err
	}

	accountInfo, err := json.Marshal(map[string]interface{}{
		"frame_id":        0,
		"is_return_check": 0,
		"goods_type":      0,
		"order_no_type":   0,
		"ip_whitelist":    "",
		"payment_type_info": map[string]string{
			"payment_code": paymentCode,
		},
	})
	if err != nil {
		return nil, err
	}

	siteMode := strings.TrimSpace(req.SiteMode)
	if siteMode == "" {
		siteMode = "1"
	}

	form := url.Values{}
	form.Set("AccountInfo", string(accountInfo))
	form.Set("getInfoUrl", getInfoURL)
	form.Set("InvoiceNumber", order.ID)
	form.Set("OrderNumber", order.MerchantOrder)
	form.Set("Currency", strings.ToUpper(strings.TrimSpace(order.Currency)))
	form.Set("store_domain", firstNonEmpty(order.MerchantSite, siteB.Domain))
	form.Set("store_name", order.MerchantSite)
	form.Set("site_mode", siteMode)
	form.Set("access_token", accessToken)
	form.Set("shipping_method_fee", "0")
	form.Set("sys_url", sysURL)
	form.Set("pay_process_url", req.ResultURL)
	form.Set("get_order", base64.StdEncoding.EncodeToString(orderInfoJSON))

	body, err := postSiteBForm(checkURL, form)
	if err != nil {
		return nil, err
	}

	result, err := parseSiteBPayResponse(body)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// buildSiteBCheckURL 拼壳包入口地址 https://{域名}/{运行目录}/check.php。
func buildSiteBCheckURL(domain, runDirectory string) string {
	domain = strings.TrimSpace(strings.ToLower(domain))
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.Trim(domain, "/")
	runDirectory = strings.Trim(strings.TrimSpace(runDirectory), "/")
	if runDirectory == "" {
		runDirectory = model.SiteBDefaultRunDirectory
	}
	return "https://" + domain + "/" + runDirectory + "/check.php"
}

// channelPaymentCode 用网关通道所属平台 code 作为壳包目录名（stripe / paypal）。
func channelPaymentCode(platform *model.Platform) (string, error) {
	if platform == nil {
		return "", fmt.Errorf("通道平台无效")
	}
	code := strings.ToLower(strings.TrimSpace(platform.Code))
	if code == "" {
		return "", fmt.Errorf("通道平台编码为空")
	}
	return code, nil
}

// issueSiteBAccessToken 签发 B 回调支付中心时用的 access_token。
func (a *App) issueSiteBAccessToken(orderID string) string {
	mac := hmac.New(sha256.New, []byte(a.authSecret))
	mac.Write([]byte(strings.TrimSpace(orderID)))
	return hex.EncodeToString(mac.Sum(nil))
}

// verifySiteBAccessToken 校验 B 带回的 api-token 是否与该订单签发值一致。
func (a *App) verifySiteBAccessToken(orderID, token string) bool {
	orderID = strings.TrimSpace(orderID)
	token = strings.TrimSpace(token)
	if orderID == "" || token == "" {
		return false
	}
	expected := a.issueSiteBAccessToken(orderID)
	return hmac.Equal([]byte(expected), []byte(token))
}

// buildSiteBOrderInfo 组装壳包 get_order / getInfoUrl 所需的订单、密钥结构。
func (a *App) buildSiteBOrderInfo(account *model.ChannelAccount, order *model.Order, req GatewayPayRequest, skSecret, wbSecret string) (map[string]interface{}, error) {
	paymentSetting, err := json.Marshal([]map[string]string{{
		"sk_secret": skSecret,
		"pk_secret": strings.TrimSpace(account.PublicKey),
		"wb_secret": wbSecret,
	}})
	if err != nil {
		return nil, err
	}

	currencyCode := strings.ToUpper(strings.TrimSpace(order.Currency))
	currencyInfo, err := json.Marshal(map[string]interface{}{
		"code":  currencyCode,
		"value": 1,
	})
	if err != nil {
		return nil, err
	}

	amountMajor := stripeMinorToMajor(order.Amount, order.Currency)
	details := req.OrderDetails
	if strings.TrimSpace(details.BillingEmail) == "" {
		details.BillingEmail = order.CustomerEmail
	}
	if strings.TrimSpace(details.BillingTelephone) == "" {
		details.BillingTelephone = order.CustomerPhone
	}
	if strings.TrimSpace(details.BillingFirstname) == "" && strings.TrimSpace(details.BillingLastname) == "" {
		details.BillingFirstname = order.CustomerName
	}
	if strings.TrimSpace(details.BillingStreetAddress) == "" {
		details.BillingStreetAddress = order.BillAddress
	}
	if strings.TrimSpace(details.BillingCity) == "" {
		details.BillingCity = order.BillCity
	}
	if strings.TrimSpace(details.BillingRegion) == "" {
		details.BillingRegion = order.BillState
	}
	if strings.TrimSpace(details.BillingPostcode) == "" {
		details.BillingPostcode = order.BillZip
	}
	if strings.TrimSpace(details.BillingCountry) == "" {
		details.BillingCountry = order.BillCountry
	}
	if strings.TrimSpace(details.ShippingStreetAddress) == "" {
		details.ShippingStreetAddress = firstNonEmpty(order.ShipAddress, details.BillingStreetAddress)
	}
	if strings.TrimSpace(details.ShippingCity) == "" {
		details.ShippingCity = firstNonEmpty(order.ShipCity, details.BillingCity)
	}
	if strings.TrimSpace(details.ShippingRegion) == "" {
		details.ShippingRegion = firstNonEmpty(order.ShipState, details.BillingRegion)
	}
	if strings.TrimSpace(details.ShippingPostcode) == "" {
		details.ShippingPostcode = firstNonEmpty(order.ShipZip, details.BillingPostcode)
	}
	if strings.TrimSpace(details.ShippingCountry) == "" {
		details.ShippingCountry = firstNonEmpty(order.ShipCountry, details.BillingCountry)
	}
	if strings.TrimSpace(details.ShippingEmail) == "" {
		details.ShippingEmail = firstNonEmpty(details.BillingEmail, order.CustomerEmail)
	}
	if strings.TrimSpace(details.ClientIP) == "" {
		details.ClientIP = order.CustomerIP
	}

	orderDetails := map[string]interface{}{
		"billing_firstname":       details.BillingFirstname,
		"billing_lastname":        details.BillingLastname,
		"billing_street_address":  details.BillingStreetAddress,
		"billing_suburb":          details.BillingSuburb,
		"billing_city":            details.BillingCity,
		"billing_region":          details.BillingRegion,
		"billing_postcode":        details.BillingPostcode,
		"billing_country":         details.BillingCountry,
		"billing_telephone":       details.BillingTelephone,
		"billing_email":           details.BillingEmail,
		"shipping_firstname":      firstNonEmpty(details.ShippingFirstname, details.BillingFirstname),
		"shipping_lastname":       firstNonEmpty(details.ShippingLastname, details.BillingLastname),
		"shipping_street_address": details.ShippingStreetAddress,
		"shipping_suburb":         details.ShippingSuburb,
		"shipping_city":           details.ShippingCity,
		"shipping_region":         details.ShippingRegion,
		"shipping_postcode":       details.ShippingPostcode,
		"shipping_country":        details.ShippingCountry,
		"shipping_telephone":      firstNonEmpty(details.ShippingTelephone, details.BillingTelephone),
		"shipping_email":          details.ShippingEmail,
		"client_ip":               details.ClientIP,
		"currency_code":           currencyCode,
		"currency_info":           string(currencyInfo),
		"client_order_total":      formatShopyyPayAmount(amountMajor),
	}

	orderGoods, err := siteBOrderGoods(order, req.Subject)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"payment_setting": string(paymentSetting),
		"order": map[string]interface{}{
			"id":           order.ID,
			"order_id":     order.ID,
			"order_number": order.MerchantOrder,
			"amount":       amountMajor,
			"status":       siteBOrderStatusPaying,
			"cancel_url":   firstNonEmpty(req.CancelURL, order.ReturnURL),
			"result_url":   firstNonEmpty(req.ResultURL, order.ReturnURL),
			"return_url":   order.ReturnURL,
		},
		"orderGoods":   orderGoods,
		"orderDetails": orderDetails,
	}, nil
}

// siteBOrderGoods 把落库商品快照转成壳包 orderGoods；没有则用订单标题凑一行。
func siteBOrderGoods(order *model.Order, subject string) ([]map[string]interface{}, error) {
	items := make([]orderGoodsSnapshot, 0)
	raw := strings.TrimSpace(order.OrderGoods)
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &items); err != nil {
			return nil, err
		}
	}
	if len(items) == 0 {
		name := strings.TrimSpace(subject)
		if name == "" {
			name = "Order " + order.MerchantOrder
		}
		items = []orderGoodsSnapshot{{Name: name, Qty: 1}}
	}
	out := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			name = "Order " + order.MerchantOrder
		}
		out = append(out, map[string]interface{}{
			"name":  name,
			"sku":   item.SKU,
			"price": item.Price,
			"qty":   item.Qty,
			"total": item.Total,
			"image": item.Image,
			"link":  item.Link,
		})
	}
	return out, nil
}

// postSiteBForm POST application/x-www-form-urlencoded 到壳包 check.php。
func postSiteBForm(target string, form url.Values) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	req.Header.Set("User-Agent", "payment-center")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 B 站 check.php 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("读取 B 站响应失败: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("B 站 check.php HTTP %d: %s", resp.StatusCode, truncateNotifyBody(body))
	}
	return body, nil
}

type siteBPayResponse struct {
	Code        json.RawMessage `json:"code"`
	Msg         string          `json:"msg"`
	PayCode     json.RawMessage `json:"pay_code"`
	PaySkipURL  string          `json:"pay_skip_url"`
	PayHTMLType json.RawMessage `json:"pay_html_type"`
}

// parseSiteBPayResponse 解析壳包回包，取出 pay_skip_url、pay_code、pay_html_type。
func parseSiteBPayResponse(body []byte) (*gatewayCheckoutResult, error) {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return nil, fmt.Errorf("B 站返回空响应")
	}

	var raw siteBPayResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("B 站返回不是 JSON: %s", truncateNotifyBody(body))
	}

	code, _ := parseFlexibleInt(raw.Code)
	if code == 1 {
		msg := strings.TrimSpace(raw.Msg)
		if msg == "" {
			msg = "B 站创建支付跳转失败"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	paySkipURL := strings.TrimSpace(raw.PaySkipURL)
	payCode, _ := parseFlexibleInt(raw.PayCode)
	payHTMLType, _ := parseFlexibleInt(raw.PayHTMLType)
	if paySkipURL == "" && payCode != 1 {
		return nil, fmt.Errorf("B 站未返回 pay_skip_url: %s", truncateNotifyBody(body))
	}
	if payCode == 0 {
		payCode = 2
	}
	if payHTMLType == 0 {
		payHTMLType = 2
	}
	return &gatewayCheckoutResult{
		CheckoutURL: paySkipURL,
		PayCode:     payCode,
		PayHTMLType: payHTMLType,
	}, nil
}

// parseFlexibleInt 把壳包 JSON 里可能是数字或字符串的字段转成 int64。
func parseFlexibleInt(raw json.RawMessage) (int64, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, nil
	}
	var asInt int64
	if err := json.Unmarshal(raw, &asInt); err == nil {
		return asInt, nil
	}
	var asFloat float64
	if err := json.Unmarshal(raw, &asFloat); err == nil {
		return int64(asFloat), nil
	}
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		asString = strings.TrimSpace(asString)
		if asString == "" {
			return 0, nil
		}
		v, err := strconv.ParseInt(asString, 10, 64)
		if err != nil {
			return 0, err
		}
		return v, nil
	}
	return 0, fmt.Errorf("invalid int")
}
