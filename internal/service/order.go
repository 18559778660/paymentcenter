package service

import (
	"fmt"
	"strings"
	"time"

	"paymentcenter/internal/model"
)

// CreateOrderResponse 创建支付订单出参。
type CreateOrderResponse struct {
	OrderID     string `json:"order_id"`
	Status      string `json:"status"`
	CheckoutURL string `json:"checkout_url"`
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
	UsdDiff          string `json:"usdDiff"`
	Status           string `json:"status"`
	ErrorMessage     string `json:"errorMessage,omitempty"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

// GetOrder 按支付中心订单号查询一笔订单。
func (a *App) GetOrder(id string) (*model.Order, error) {
	return a.store.GetOrder(id)
}

// ListOrders 查询订单列表（含商户名、账号名；B站优先读订单快照字段）。
func (a *App) ListOrders() ([]OrderListItem, error) {
	orders, err := a.store.ListOrders()
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
		Fee:              "-",        // 预留：暂无手续费字段
		UsdDiff:          "-",        // 预留：暂无美元偏差字段
		Status:           string(order.Status),
		ErrorMessage:     order.ErrorMessage,
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
