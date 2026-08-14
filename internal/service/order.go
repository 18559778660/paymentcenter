package service

import (
	"fmt"
	"strconv"
	"time"

	"paymentcenter/internal/model"
)

// CreateOrderRequest 创建订单入参。
type CreateOrderRequest struct {
	MerchantOrder string `json:"merchant_order" binding:"required"`
	MerchantSite  string `json:"merchant_site" binding:"required"`
	Amount        int64  `json:"amount" binding:"required"`
	Currency      string `json:"currency" binding:"required"`
	ReturnURL     string `json:"return_url" binding:"required"`
	NotifyURL     string `json:"notify_url" binding:"required"`
}

// CreateOrderResponse 创建订单出参。
type CreateOrderResponse struct {
	OrderID     string `json:"order_id"`
	Status      string `json:"status"`
	CheckoutURL string `json:"checkout_url"`
}

// CreateOrder 创建一笔支付订单，生成中心订单号和收银台地址。
func (a *App) CreateOrder(req CreateOrderRequest) (CreateOrderResponse, error) {
	now := time.Now().UTC()
	id := "pc_" + strconv.FormatInt(now.UnixNano(), 10)
	order := &model.Order{
		ID:            id,
		MerchantOrder: req.MerchantOrder,
		MerchantSite:  req.MerchantSite,
		Channel:       "win_stripe",
		Provider:      "stripe",
		Amount:        req.Amount,
		Currency:      req.Currency,
		ReturnURL:     req.ReturnURL,
		NotifyURL:     req.NotifyURL,
		Status:        model.OrderStatusCreated,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	order.CheckoutURL = fmt.Sprintf("/mock/checkout/%s", order.ID)
	if err := a.store.SaveOrder(order); err != nil {
		return CreateOrderResponse{}, err
	}
	return CreateOrderResponse{
		OrderID:     order.ID,
		Status:      string(order.Status),
		CheckoutURL: order.CheckoutURL,
	}, nil
}

// GetOrder 按支付中心订单号查询一笔订单。
func (a *App) GetOrder(id string) (*model.Order, error) {
	return a.store.GetOrder(id)
}

// ListOrders 查询全部订单。
func (a *App) ListOrders() ([]*model.Order, error) {
	return a.store.ListOrders()
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
