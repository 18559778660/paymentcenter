package service

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
)

// CreateOrderResponse 创建支付订单出参。
type CreateOrderResponse struct {
	OrderID     string `json:"order_id"`
	Status      string `json:"status"`
	CheckoutURL string `json:"checkout_url"`
	PayCode     int64  `json:"pay_code,omitempty"`
	PayHTMLType int64  `json:"pay_html_type,omitempty"`
}

// OrderListItem 管理端订单列表项（兼容预留列 + 当前 payment_orders 字段）。
type OrderListItem struct {
	ID               string `json:"id"`
	MerchantID       uint   `json:"merchantId"`
	MerchantName     string `json:"merchantName"`
	MerchantOrder    string `json:"merchantOrder"`
	MerchantSite     string `json:"merchantSite"`
	ProviderRef      string `json:"providerRef"`
	SiteBID          uint   `json:"siteBId"`
	SiteB            string `json:"siteB"`
	Channel          string `json:"channel"`
	ChannelAccountID uint   `json:"channelAccountId"`
	AccountName      string `json:"accountName"`
	Provider         string `json:"provider"`
	Amount           int64  `json:"amount"`
	Currency         string `json:"currency"`
	SiteAmount       string `json:"siteAmount"`
	TradeAmount      string `json:"tradeAmount"`
	Fee              string `json:"fee"`
	FeeAmount        int64  `json:"feeAmount"`
	UsdDiff          string `json:"usdDiff"`
	Status           string `json:"status"`
	ErrorMessage     string `json:"errorMessage,omitempty"`
	CustomerName     string `json:"customerName"`
	CustomerEmail    string `json:"customerEmail"`
	CustomerPhone    string `json:"customerPhone"`
	CustomerIP       string `json:"customerIp"`
	BillCountry      string `json:"billCountry"`
	BillState        string `json:"billState"`
	BillCity         string `json:"billCity"`
	BillZip          string `json:"billZip"`
	BillAddress      string `json:"billAddress"`
	ShipCountry      string `json:"shipCountry"`
	ShipState        string `json:"shipState"`
	ShipCity         string `json:"shipCity"`
	ShipZip          string `json:"shipZip"`
	ShipAddress      string `json:"shipAddress"`
	OrderGoods       string `json:"orderGoods,omitempty"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

// GetOrder 按支付中心订单号查询一笔订单。
func (a *App) GetOrder(id string) (*model.Order, error) {
	return a.store.GetOrder(id)
}

// OrderSummary 管理端订单汇总栏。
type OrderSummary struct {
	TotalCount   int64  `json:"totalCount"`
	UnpaidCount  int64  `json:"unpaidCount"`
	FailedCount  int64  `json:"failedCount"`
	SuccessCount int64  `json:"successCount"`
	PayRate      string `json:"payRate"`
	SuccessRate  string `json:"successRate"`
	TotalRate    string `json:"totalRate"`
	AmountUsd    string `json:"amountUsd"`
}

// OrderListQuery 管理端订单列表筛选。
type OrderListQuery struct {
	MerchantID       uint
	CustomerEmail    string
	CustomerPhone    string
	CustomerIP       string
	MerchantOrder    string
	MerchantSite     string
	Status           string
	Provider         string
	ChannelAccountID uint
	SiteB            string
	CreatedFrom      string
	CreatedTo        string
	Country          string
	Currency         string
	MinAmountMajor   string
	MaxAmountMajor   string
	ProviderRef      string
}

// ListOrders 查询订单列表（含商户名、账号名；B站优先读订单快照字段）。
func (a *App) ListOrders(q OrderListQuery) ([]OrderListItem, error) {
	minAmount, maxAmount, err := parseOrderAmountRange(q.MinAmountMajor, q.MaxAmountMajor, q.Currency)
	if err != nil {
		return nil, err
	}
	orders, err := a.store.ListOrders(store.OrderListFilter{
		MerchantID:       q.MerchantID,
		CustomerEmail:    strings.TrimSpace(q.CustomerEmail),
		CustomerPhone:    strings.TrimSpace(q.CustomerPhone),
		CustomerIP:       strings.TrimSpace(q.CustomerIP),
		MerchantOrder:    strings.TrimSpace(q.MerchantOrder),
		MerchantSite:     strings.TrimSpace(q.MerchantSite),
		Status:           strings.TrimSpace(q.Status),
		Provider:         strings.TrimSpace(q.Provider),
		ChannelAccountID: q.ChannelAccountID,
		SiteB:            strings.TrimSpace(q.SiteB),
		CreatedFrom:      strings.TrimSpace(q.CreatedFrom),
		CreatedTo:        strings.TrimSpace(q.CreatedTo),
		Country:          strings.TrimSpace(q.Country),
		Currency:         strings.TrimSpace(q.Currency),
		MinAmount:        minAmount,
		MaxAmount:        maxAmount,
		ProviderRef:      strings.TrimSpace(q.ProviderRef),
	})
	if err != nil {
		return nil, err
	}
	merchantNames, err := a.loadOrderMerchantNameMap(orders)
	if err != nil {
		return nil, err
	}
	accountMap, err := a.loadOrderAccountMap(orders)
	if err != nil {
		return nil, err
	}
	out := make([]OrderListItem, 0, len(orders))
	for _, order := range orders {
		if order == nil {
			continue
		}
		out = append(out, toOrderListItem(order, merchantNames, accountMap))
	}
	return out, nil
}

// parseOrderAmountRange 将前端主单位金额转为最小货币单位；未填币种时按 ×100。
func parseOrderAmountRange(minRaw, maxRaw, currency string) (minAmount, maxAmount *int64, err error) {
	currency = strings.TrimSpace(currency)
	if strings.TrimSpace(minRaw) != "" {
		major, parseErr := strconv.ParseFloat(strings.TrimSpace(minRaw), 64)
		if parseErr != nil || major < 0 {
			return nil, nil, fmt.Errorf("无效的最小金额")
		}
		minor, convErr := majorAmountToStripeMinorOrDefault(major, currency)
		if convErr != nil {
			return nil, nil, convErr
		}
		minAmount = &minor
	}
	if strings.TrimSpace(maxRaw) != "" {
		major, parseErr := strconv.ParseFloat(strings.TrimSpace(maxRaw), 64)
		if parseErr != nil || major < 0 {
			return nil, nil, fmt.Errorf("无效的最大金额")
		}
		minor, convErr := majorAmountToStripeMinorOrDefault(major, currency)
		if convErr != nil {
			return nil, nil, convErr
		}
		maxAmount = &minor
	}
	if minAmount != nil && maxAmount != nil && *minAmount > *maxAmount {
		return nil, nil, fmt.Errorf("最小金额不能大于最大金额")
	}
	return minAmount, maxAmount, nil
}

func majorAmountToStripeMinorOrDefault(amount float64, currency string) (int64, error) {
	if amount < 0 {
		return 0, fmt.Errorf("无效金额")
	}
	if amount == 0 {
		return 0, nil
	}
	currency = strings.TrimSpace(currency)
	if currency == "" {
		return int64(math.Round(amount * 100)), nil
	}
	return majorAmountToStripeMinor(amount, currency)
}

// GetOrderSummary 统计全部订单汇总（不受列表条数限制）。
func (a *App) GetOrderSummary() (OrderSummary, error) {
	rows, err := a.store.CountOrdersByStatus()
	if err != nil {
		return OrderSummary{}, err
	}

	var created, pending, paid, failed, cancelled int64
	for _, row := range rows {
		switch model.OrderStatus(row.Status) {
		case model.OrderStatusCreated:
			created = row.Count
		case model.OrderStatusPending:
			pending = row.Count
		case model.OrderStatusPaid:
			paid = row.Count
		case model.OrderStatusFailed:
			failed = row.Count
		case model.OrderStatusCancelled:
			cancelled = row.Count
		}
	}

	total := created + pending + paid + failed + cancelled
	unpaid := created + pending
	failTotal := failed + cancelled
	finished := paid + failTotal

	amountUsd, err := a.sumPaidAmountUSD()
	if err != nil {
		return OrderSummary{}, err
	}

	return OrderSummary{
		TotalCount:   total,
		UnpaidCount:  unpaid,
		FailedCount:  failTotal,
		SuccessCount: paid,
		PayRate:      formatPercent(paid, total),
		SuccessRate:  formatPercent(paid, finished),
		TotalRate:    formatPercent(finished, total),
		AmountUsd:    fmt.Sprintf("%.2f", amountUsd),
	}, nil
}

// sumPaidAmountUSD 将已支付金额按货币汇率折算为 USD。
// 汇率约定：rate = 1 USD 可兑换的该币种数量（USD=1），USD = major / rate。
func (a *App) sumPaidAmountUSD() (float64, error) {
	paidRows, err := a.store.SumPaidOrderAmountsByCurrency()
	if err != nil {
		return 0, err
	}
	if len(paidRows) == 0 {
		return 0, nil
	}
	currencies, err := a.store.ListCurrencies(store.CurrencyListFilter{})
	if err != nil {
		return 0, err
	}
	rateByCode := make(map[string]float64, len(currencies))
	for _, item := range currencies {
		code := strings.ToUpper(strings.TrimSpace(item.Code))
		if code == "" {
			continue
		}
		rateByCode[code] = item.Rate
	}

	var totalUSD float64
	for _, row := range paidRows {
		code := strings.ToUpper(strings.TrimSpace(row.Currency))
		if code == "" || row.Amount == 0 {
			continue
		}
		major := float64(row.Amount) / 100
		if code == "USD" {
			totalUSD += major
			continue
		}
		rate := rateByCode[code]
		if rate <= 0 || math.IsNaN(rate) || math.IsInf(rate, 0) {
			// 无有效汇率时按 1:1 计入，避免整栏为空。
			totalUSD += major
			continue
		}
		totalUSD += major / rate
	}
	return totalUSD, nil
}

func formatPercent(num, den int64) string {
	if den <= 0 {
		return "0%"
	}
	pct := float64(num) * 100 / float64(den)
	return fmt.Sprintf("%.2f%%", pct)
}

// MarkPaid 把订单标记为已支付。
func (a *App) MarkPaid(id, providerRef string) (*model.Order, error) {
	order, err := a.store.GetOrder(id)
	if err != nil {
		return nil, err
	}
	order.Status = model.OrderStatusPaid
	order.ProviderRef = providerRef
	order.UpdatedAt = time.Now().UTC()
	if err := a.store.SaveOrder(order); err != nil {
		return nil, err
	}
	a.appendOrderLogFromOrder(order, model.OrderLogTypeRecord, "人工标记订单已支付")
	return order, nil
}

// MarkFailed 把订单标记为失败。
func (a *App) MarkFailed(id, message string) (*model.Order, error) {
	order, err := a.store.GetOrder(id)
	if err != nil {
		return nil, err
	}
	order.Status = model.OrderStatusFailed
	order.ErrorMessage = message
	order.UpdatedAt = time.Now().UTC()
	if err := a.store.SaveOrder(order); err != nil {
		return nil, err
	}
	remark := "人工标记订单失败"
	if strings.TrimSpace(message) != "" {
		remark = remark + ": " + strings.TrimSpace(message)
	}
	a.appendOrderLogFromOrder(order, model.OrderLogTypeFailRecord, remark)
	return order, nil
}

func (a *App) loadOrderMerchantNameMap(orders []*model.Order) (map[uint]string, error) {
	ids := make([]uint, 0, len(orders))
	seen := map[uint]struct{}{}
	for _, order := range orders {
		if order == nil || order.MerchantID == 0 {
			continue
		}
		if _, ok := seen[order.MerchantID]; ok {
			continue
		}
		seen[order.MerchantID] = struct{}{}
		ids = append(ids, order.MerchantID)
	}
	if len(ids) == 0 {
		return map[uint]string{}, nil
	}
	merchants, err := a.store.GetMerchantsByIDs(ids)
	if err != nil {
		return nil, err
	}
	names := make(map[uint]string, len(merchants))
	for _, merchant := range merchants {
		names[merchant.ID] = merchant.Name
	}
	return names, nil
}

func (a *App) loadOrderAccountMap(orders []*model.Order) (map[uint]model.ChannelAccount, error) {
	accountIDs := make([]uint, 0, len(orders))
	seenAccount := map[uint]struct{}{}
	for _, order := range orders {
		if order == nil || order.ChannelAccountID == 0 {
			continue
		}
		if _, ok := seenAccount[order.ChannelAccountID]; ok {
			continue
		}
		seenAccount[order.ChannelAccountID] = struct{}{}
		accountIDs = append(accountIDs, order.ChannelAccountID)
	}
	accountMap := map[uint]model.ChannelAccount{}
	if len(accountIDs) == 0 {
		return accountMap, nil
	}
	accounts, err := a.store.GetChannelAccountsByIDs(accountIDs)
	if err != nil {
		return nil, err
	}
	for _, account := range accounts {
		accountMap[account.ID] = account
	}
	return accountMap, nil
}

func toOrderListItem(
	order *model.Order,
	merchantNames map[uint]string,
	accountMap map[uint]model.ChannelAccount,
) OrderListItem {
	amountText := formatOrderAmount(order.Amount, order.Currency)
	accountName := ""
	if account, ok := accountMap[order.ChannelAccountID]; ok {
		accountName = strings.TrimSpace(account.AccountNo)
		if alias := strings.TrimSpace(account.Alias); alias != "" {
			if accountName == "" {
				accountName = alias
			} else {
				accountName = accountName + " " + alias
			}
		}
	}
	merchantName := merchantNames[order.MerchantID]
	if merchantName == "" && order.MerchantID > 0 {
		merchantName = fmt.Sprintf("#%d", order.MerchantID)
	}
	return OrderListItem{
		ID:               order.ID,
		MerchantID:       order.MerchantID,
		MerchantName:     merchantName,
		MerchantOrder:    order.MerchantOrder,
		MerchantSite:     order.MerchantSite,
		ProviderRef:      order.ProviderRef,
		SiteBID:          order.SiteBID,
		SiteB:            strings.TrimSpace(order.SiteB),
		Channel:          order.Channel,
		ChannelAccountID: order.ChannelAccountID,
		AccountName:      accountName,
		Provider:         order.Provider,
		Amount:           order.Amount,
		Currency:         order.Currency,
		SiteAmount:       amountText,
		TradeAmount:      amountText, // 当前无独立交易金额字段，与网站金额一致
		Fee:              formatOrderFee(order.Fee, order.Currency),
		FeeAmount:        order.Fee,
		UsdDiff:          formatOrderUsdDiff(order.UsdDiff),
		Status:           string(order.Status),
		ErrorMessage:     order.ErrorMessage,
		CustomerName:     order.CustomerName,
		CustomerEmail:    order.CustomerEmail,
		CustomerPhone:    order.CustomerPhone,
		CustomerIP:       order.CustomerIP,
		BillCountry:      order.BillCountry,
		BillState:        order.BillState,
		BillCity:         order.BillCity,
		BillZip:          order.BillZip,
		BillAddress:      order.BillAddress,
		ShipCountry:      order.ShipCountry,
		ShipState:        order.ShipState,
		ShipCity:         order.ShipCity,
		ShipZip:          order.ShipZip,
		ShipAddress:      order.ShipAddress,
		OrderGoods:       order.OrderGoods,
		CreatedAt:        order.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        order.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// formatOrderAmount 将最小货币单位金额格式化为展示文案，例如 "EUR 1.00"。
func formatOrderAmount(amount int64, currency string) string {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	major := float64(amount) / 100
	if currency == "" {
		return fmt.Sprintf("%.2f", major)
	}
	return fmt.Sprintf("%s %.2f", currency, major)
}

// formatOrderFee 格式化 Stripe 手续费；尚未回填时显示 "-"。
func formatOrderFee(fee int64, currency string) string {
	if fee <= 0 {
		return "-"
	}
	return formatOrderAmount(fee, currency)
}

// formatOrderUsdDiff 格式化按商户汇率偏差折算后的美元金额。
func formatOrderUsdDiff(v *float64) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("USD %.2f", *v)
}

// calcMerchantUsdDiff 按商户汇率偏差计算订单美元金额。
// 货币表汇率约定：rate = 1 USD 可兑换的该币种数量；非美元时：
//
//	美元金额 = 网站金额 / 汇率 × (100 - rateDiff) / 100
//
// 网站币为 USD 时不套用偏差，直接取网站金额。
func (a *App) calcMerchantUsdDiff(amountMinor int64, currency string, rateDiff int) float64 {
	if rateDiff < 0 {
		rateDiff = 0
	}
	if rateDiff > 100 {
		rateDiff = 100
	}
	major := stripeMinorToMajor(amountMinor, currency)
	code := strings.ToUpper(strings.TrimSpace(currency))
	if code == "" || code == "USD" {
		return roundUSDAmount(major)
	}

	rate := 0.0
	if item, err := a.store.FindCurrencyByCode(code); err == nil && item != nil {
		rate = item.Rate
	}
	usd := major
	if rate > 0 && !math.IsNaN(rate) && !math.IsInf(rate, 0) {
		usd = major / rate
	}
	usd = usd * float64(100-rateDiff) / 100
	return roundUSDAmount(usd)
}

// stripeMinorToMajor Stripe 最小货币单位转主单位。
func stripeMinorToMajor(amount int64, currency string) float64 {
	currency = strings.ToLower(strings.TrimSpace(currency))
	if _, ok := stripeZeroDecimalCurrencies[currency]; ok {
		return float64(amount)
	}
	divisor := 100.0
	if _, ok := stripeThreeDecimalCurrencies[currency]; ok {
		divisor = 1000.0
	}
	return float64(amount) / divisor
}

func roundUSDAmount(v float64) float64 {
	return math.Round(v*10000) / 10000
}
