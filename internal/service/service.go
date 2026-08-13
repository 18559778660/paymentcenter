package service

import (
	"fmt"
	"strconv"
	"time"

	"paymentcenter/internal/domain"
	"paymentcenter/internal/store"
)

type App struct {
	store *store.MemoryStore
}

func NewApp(st *store.MemoryStore) *App {
	return &App{store: st}
}

type CreateOrderRequest struct {
	MerchantOrder string `json:"merchant_order" binding:"required"`
	MerchantSite  string `json:"merchant_site" binding:"required"`
	Amount        int64  `json:"amount" binding:"required"`
	Currency      string `json:"currency" binding:"required"`
	ReturnURL     string `json:"return_url" binding:"required"`
	NotifyURL     string `json:"notify_url" binding:"required"`
}

type CreateOrderResponse struct {
	OrderID     string `json:"order_id"`
	Status      string `json:"status"`
	CheckoutURL string `json:"checkout_url"`
}

func (a *App) CreateOrder(req CreateOrderRequest) CreateOrderResponse {
	now := time.Now().UTC()
	id := "pc_" + strconv.FormatInt(now.UnixNano(), 10)
	order := &domain.Order{
		ID:           id,
		MerchantOrder: req.MerchantOrder,
		MerchantSite:  req.MerchantSite,
		Channel:       "win_stripe",
		Provider:      "stripe",
		Amount:        req.Amount,
		Currency:      req.Currency,
		ReturnURL:     req.ReturnURL,
		NotifyURL:     req.NotifyURL,
		Status:        domain.OrderStatusCreated,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	order.CheckoutURL = fmt.Sprintf("/mock/checkout/%s", order.ID)
	a.store.Save(order)
	return CreateOrderResponse{
		OrderID:     order.ID,
		Status:      string(order.Status),
		CheckoutURL: order.CheckoutURL,
	}
}

func (a *App) GetOrder(id string) (*domain.Order, error) {
	return a.store.Get(id)
}

func (a *App) ListOrders() []*domain.Order {
	return a.store.List()
}

func (a *App) MarkPaid(id, providerRef string) (*domain.Order, error) {
	order, err := a.store.Get(id)
	if err != nil {
		return nil, err
	}
	order.Status = domain.OrderStatusPaid
	order.ProviderRef = providerRef
	order.UpdatedAt = time.Now().UTC()
	a.store.Save(order)
	return order, nil
}

func (a *App) MarkFailed(id, message string) (*domain.Order, error) {
	order, err := a.store.Get(id)
	if err != nil {
		return nil, err
	}
	order.Status = domain.OrderStatusFailed
	order.ErrorMessage = message
	order.UpdatedAt = time.Now().UTC()
	a.store.Save(order)
	return order, nil
}
