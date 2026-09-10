package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

var ErrGatewayParamInvalid = errors.New("gateway param invalid")

// gatewayPayPayload A 站支付入参（当前为 Shopyy 协议，后续其它 A 站可扩展适配）。
type gatewayPayPayload struct {
	OrderNumber     string               `json:"orderNumber"`
	Currency        string               `json:"currency"`
	Amount          json.RawMessage      `json:"amount"`
	ReturnURL       string               `json:"return_url"`
	AsynURL         string               `json:"asyn_url"`
	CancelURL       string               `json:"cancel_url"`
	ResultURL       string               `json:"result_url"`
	Domain          string               `json:"domain"`
	OrderTime       string               `json:"order_time"`
	Email           string               `json:"email"`
	CustomerIp      string               `json:"customerIp"`
	BillingName     string               `json:"billing_name"`
	ShippingName    string               `json:"shipping_name"`
	BillingAddress  string               `json:"billing_address"`
	ShippingAddress string               `json:"shipping_address"`
	FrameType       string               `json:"frame_type"`
	SiteMode        string               `json:"site_mode"`
	OrderDetails    gatewayOrderDetails  `json:"order_details"`
	OrderGoods      []gatewayOrderGoods  `json:"order_goods"`
}

type gatewayOrderDetails struct {
	BillingFirstname     string `json:"billing_firstname"`
	BillingLastname      string `json:"billing_lastname"`
	BillingStreetAddress string `json:"billing_street_address"`
	BillingCity          string `json:"billing_city"`
	BillingRegion        string `json:"billing_region"`
	BillingPostcode      string `json:"billing_postcode"`
	BillingCountry       string `json:"billing_country"`
	BillingTelephone     string `json:"billing_telephone"`
	BillingEmail         string `json:"billing_email"`
	ShippingFirstname    string `json:"shipping_firstname"`
	ShippingLastname     string `json:"shipping_lastname"`
	ShippingStreetAddress string `json:"shipping_street_address"`
	ShippingCity         string `json:"shipping_city"`
	ShippingRegion       string `json:"shipping_region"`
	ShippingPostcode     string `json:"shipping_postcode"`
	ShippingCountry      string `json:"shipping_country"`
	ShippingTelephone    string `json:"shipping_telephone"`
	ShippingEmail        string `json:"shipping_email"`
	ClientIP             string `json:"client_ip"`
}

type gatewayOrderGoods struct {
	SKU   string          `json:"sku"`
	Name  string          `json:"name"`
	Image string          `json:"image"`
	Total json.RawMessage `json:"total"`
	Price json.RawMessage `json:"price"`
	Qty   json.RawMessage `json:"qty"`
	Link  string          `json:"link"`
}

// orderGoodsSnapshot 落库商品快照结构。
type orderGoodsSnapshot struct {
	SKU   string `json:"sku"`
	Name  string `json:"name"`
	Image string `json:"image,omitempty"`
	Total string `json:"total"`
	Price string `json:"price"`
	Qty   int    `json:"qty"`
	Link  string `json:"link,omitempty"`
}

// stripeZeroDecimalCurrencies Stripe 零小数币种，主单位即最小单位。
var stripeZeroDecimalCurrencies = map[string]struct{}{
	"bif": {}, "clp": {}, "djf": {}, "gnf": {}, "jpy": {}, "kmf": {},
	"krw": {}, "mga": {}, "pyg": {}, "rwf": {}, "ugx": {}, "vnd": {},
	"vuv": {}, "xaf": {}, "xof": {}, "xpf": {},
}

// stripeThreeDecimalCurrencies Stripe 三位小数币种。
var stripeThreeDecimalCurrencies = map[string]struct{}{
	"bhd": {}, "jod": {}, "kwd": {}, "omr": {}, "tnd": {},
}

// NormalizeGatewayPayRequest 解析 A 站入参并映射为内部订单字段。
func NormalizeGatewayPayRequest(payload []byte) (GatewayPayRequest, error) {
	if len(payload) == 0 {
		return GatewayPayRequest{}, fmt.Errorf("%w: empty body", ErrGatewayParamInvalid)
	}
	raw, err := parseGatewayPayPayload(payload)
	if err != nil {
		return GatewayPayRequest{}, fmt.Errorf("%w: invalid json", ErrGatewayParamInvalid)
	}

	req := GatewayPayRequest{
		MerchantOrder: strings.TrimSpace(raw.OrderNumber),
		MerchantSite:  strings.TrimSpace(raw.Domain),
		Currency:      raw.Currency,
		ReturnURL:     strings.TrimSpace(raw.ReturnURL),
		NotifyURL:     strings.TrimSpace(raw.AsynURL),
		CancelURL:     strings.TrimSpace(raw.CancelURL),
		ResultURL:     strings.TrimSpace(raw.ResultURL),
		SiteMode:      strings.TrimSpace(raw.SiteMode),
	}
	if len(raw.OrderGoods) > 0 {
		req.Subject = strings.TrimSpace(raw.OrderGoods[0].Name)
	}
	fillGatewayCustomerSnapshot(&req, raw)

	amountMajor, err := parseGatewayAmountMajor(raw.Amount)
	if err != nil {
		return GatewayPayRequest{}, fmt.Errorf("%w: %v", ErrGatewayParamInvalid, err)
	}
	payAmount := formatShopyyPayAmount(amountMajor)
	payCurrency := strings.ToUpper(strings.TrimSpace(raw.Currency))
	req.ShopyyVerify = ShopyyNotifyVerifySnapshot{
		OrderNumber:  req.MerchantOrder,
		PayUSDAmount: payAmount,
		PayCurrency:  payCurrency,
		PayAmount:    payAmount,
		OrderTime:    strings.TrimSpace(raw.OrderTime),
		Domain:       req.MerchantSite,
		Email:        firstNonEmpty(req.CustomerEmail, strings.TrimSpace(raw.Email)),
		CustomerIP:   firstNonEmpty(req.CustomerIP, strings.TrimSpace(raw.CustomerIp)),
	}
	req.Amount, err = majorAmountToStripeMinor(amountMajor, req.Currency)
	if err != nil {
		return GatewayPayRequest{}, fmt.Errorf("%w: %v", ErrGatewayParamInvalid, err)
	}
	if err := validateGatewayPayRequest(req); err != nil {
		return GatewayPayRequest{}, err
	}
	return req, nil
}

// fillGatewayCustomerSnapshot 从 Shopyy 报文提取详情用客户/地址/商品快照。
func fillGatewayCustomerSnapshot(req *GatewayPayRequest, raw gatewayPayPayload) {
	d := raw.OrderDetails
	req.CustomerName = firstNonEmpty(
		strings.TrimSpace(raw.BillingName),
		joinPersonName(d.BillingFirstname, d.BillingLastname),
		strings.TrimSpace(raw.ShippingName),
		joinPersonName(d.ShippingFirstname, d.ShippingLastname),
	)
	req.CustomerEmail = firstNonEmpty(
		strings.TrimSpace(raw.Email),
		strings.TrimSpace(d.BillingEmail),
		strings.TrimSpace(d.ShippingEmail),
	)
	req.CustomerPhone = firstNonEmpty(
		strings.TrimSpace(d.BillingTelephone),
		strings.TrimSpace(d.ShippingTelephone),
	)
	req.CustomerIP = firstNonEmpty(
		strings.TrimSpace(raw.CustomerIp),
		strings.TrimSpace(d.ClientIP),
	)

	req.BillCountry = strings.TrimSpace(d.BillingCountry)
	req.BillState = strings.TrimSpace(d.BillingRegion)
	req.BillCity = strings.TrimSpace(d.BillingCity)
	req.BillZip = strings.TrimSpace(d.BillingPostcode)
	req.BillAddress = firstNonEmpty(
		strings.TrimSpace(d.BillingStreetAddress),
		strings.TrimSpace(raw.BillingAddress),
	)

	req.ShipCountry = strings.TrimSpace(d.ShippingCountry)
	req.ShipState = strings.TrimSpace(d.ShippingRegion)
	req.ShipCity = strings.TrimSpace(d.ShippingCity)
	req.ShipZip = strings.TrimSpace(d.ShippingPostcode)
	req.ShipAddress = firstNonEmpty(
		strings.TrimSpace(d.ShippingStreetAddress),
		strings.TrimSpace(raw.ShippingAddress),
	)

	if len(raw.OrderGoods) == 0 {
		return
	}
	items := make([]orderGoodsSnapshot, 0, len(raw.OrderGoods))
	for _, g := range raw.OrderGoods {
		items = append(items, orderGoodsSnapshot{
			SKU:   strings.TrimSpace(g.SKU),
			Name:  strings.TrimSpace(g.Name),
			Image: strings.TrimSpace(g.Image),
			Total: rawMessageToTrimmedString(g.Total),
			Price: rawMessageToTrimmedString(g.Price),
			Qty:   rawMessageToInt(g.Qty),
			Link:  strings.TrimSpace(g.Link),
		})
	}
	if b, err := json.Marshal(items); err == nil {
		req.OrderGoods = string(b)
	}
}

func joinPersonName(first, last string) string {
	first = strings.TrimSpace(first)
	last = strings.TrimSpace(last)
	switch {
	case first != "" && last != "":
		return first + " " + last
	case first != "":
		return first
	default:
		return last
	}
}

func rawMessageToTrimmedString(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return strings.TrimSpace(asString)
	}
	var asNumber json.Number
	if err := json.Unmarshal(raw, &asNumber); err == nil {
		return strings.TrimSpace(asNumber.String())
	}
	return strings.Trim(string(raw), `"`)
}

func rawMessageToInt(raw json.RawMessage) int {
	s := rawMessageToTrimmedString(raw)
	if s == "" {
		return 0
	}
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return int(math.Round(f))
	}
	return 0
}

// parseGatewayPayPayload 解析 Shopyy 支付 body：{"order":"{...json...}"}。
func parseGatewayPayPayload(payload []byte) (gatewayPayPayload, error) {
	var wrapper struct {
		Order string `json:"order"`
	}
	if err := json.Unmarshal(payload, &wrapper); err != nil {
		return gatewayPayPayload{}, err
	}
	orderJSON := strings.TrimSpace(wrapper.Order)
	if orderJSON == "" {
		return gatewayPayPayload{}, errors.New("missing order")
	}

	var raw gatewayPayPayload
	if err := json.Unmarshal([]byte(orderJSON), &raw); err != nil {
		return gatewayPayPayload{}, err
	}
	return raw, nil
}

// majorAmountToStripeMinor A 站主货币单位转 Stripe 最小单位。
// Shopyy 文档约定 amount 为 decimal、保留两位小数（如 EUR 1.00）；
// 换算倍数按 Stripe 对各 currency 的规则，不是一律 ×100。
func majorAmountToStripeMinor(amount float64, currency string) (int64, error) {
	if amount <= 0 {
		return 0, errors.New("invalid amount")
	}
	currency = strings.ToLower(strings.TrimSpace(currency))
	if currency == "" {
		return 0, errors.New("missing currency")
	}
	if _, ok := stripeZeroDecimalCurrencies[currency]; ok {
		return int64(math.Round(amount)), nil
	}
	multiplier := 100.0
	if _, ok := stripeThreeDecimalCurrencies[currency]; ok {
		multiplier = 1000.0
	}
	minor := math.Round(amount * multiplier)
	if minor <= 0 {
		return 0, errors.New("invalid amount")
	}
	return int64(minor), nil
}

// validateGatewayPayRequest 验证网关支付入参。
func validateGatewayPayRequest(req GatewayPayRequest) error {
	if strings.TrimSpace(req.MerchantOrder) == "" {
		return fmt.Errorf("%w: missing orderNumber", ErrGatewayParamInvalid)
	}
	if strings.TrimSpace(req.MerchantSite) == "" {
		return fmt.Errorf("%w: missing domain", ErrGatewayParamInvalid)
	}
	if req.Amount <= 0 {
		return fmt.Errorf("%w: invalid amount", ErrGatewayParamInvalid)
	}
	if strings.TrimSpace(req.Currency) == "" {
		return fmt.Errorf("%w: missing currency", ErrGatewayParamInvalid)
	}
	if strings.TrimSpace(req.ReturnURL) == "" {
		return fmt.Errorf("%w: missing return_url", ErrGatewayParamInvalid)
	}
	if strings.TrimSpace(req.NotifyURL) == "" {
		return fmt.Errorf("%w: missing asyn_url", ErrGatewayParamInvalid)
	}
	if strings.TrimSpace(req.ResultURL) == "" {
		return fmt.Errorf("%w: missing result_url", ErrGatewayParamInvalid)
	}
	if strings.TrimSpace(req.CancelURL) == "" {
		return fmt.Errorf("%w: missing cancel_url", ErrGatewayParamInvalid)
	}
	return nil
}

// parseGatewayAmountMajor 解析网关支付金额。
func parseGatewayAmountMajor(raw json.RawMessage) (float64, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, errors.New("missing amount")
	}
	var asInt int64
	if err := json.Unmarshal(raw, &asInt); err == nil {
		return float64(asInt), nil
	}
	var asFloat float64
	if err := json.Unmarshal(raw, &asFloat); err == nil {
		return asFloat, nil
	}
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		asString = strings.TrimSpace(asString)
		if asString == "" {
			return 0, errors.New("missing amount")
		}
		v, err := strconv.ParseFloat(asString, 64)
		if err != nil {
			return 0, errors.New("invalid amount")
		}
		return v, nil
	}
	return 0, errors.New("invalid amount")
}
