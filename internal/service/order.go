package service

import (
	"time"

	"paymentcenter/internal/model"
)

// CreateOrderResponse 创建支付订单出参。
type CreateOrderResponse struct {
	OrderID     string `json:"order_id"`
	Status      string `json:"status"`
	CheckoutURL string `json:"checkout_url"`
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
