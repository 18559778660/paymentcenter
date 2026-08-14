package service

import (
	"fmt"
	"strconv"
	"time"

	"paymentcenter/internal/model"
)

type Store interface {
	SaveOrder(order *model.Order) error
	GetOrder(id string) (*model.Order, error)
	ListOrders() ([]*model.Order, error)
	CreateUser(user *model.User) error
	SaveUser(user *model.User) error
	GetUserByID(id uint) (*model.User, error)
	FindUserByUsername(username string) (*model.User, error)
	CreateRole(role *model.Role) error
	FindRoleByCode(code string) (*model.Role, error)
	ListRolesByUserID(userID uint) ([]model.Role, error)
	EnsureUserRole(userID, roleID uint) error
	CreateMenu(menu *model.Menu) error
	FindMenuByName(name string) (*model.Menu, error)
	ListMenus() ([]model.Menu, error)
	SaveMenu(menu *model.Menu) error
	EnsureRoleMenu(roleID, menuID uint) error
	ListMenusByUserID(userID uint) ([]model.Menu, error)
}

type App struct {
	store      Store
	authSecret string
	tokenTTL   time.Duration
}

// 创建应用
func NewApp(st Store, authSecret string, tokenTTL time.Duration) *App {
	return &App{store: st, authSecret: authSecret, tokenTTL: tokenTTL}
}

// 创建订单请求
type CreateOrderRequest struct {
	MerchantOrder string `json:"merchant_order" binding:"required"`
	MerchantSite  string `json:"merchant_site" binding:"required"`
	Amount        int64  `json:"amount" binding:"required"`
	Currency      string `json:"currency" binding:"required"`
	ReturnURL     string `json:"return_url" binding:"required"`
	NotifyURL     string `json:"notify_url" binding:"required"`
}

// 创建订单响应
type CreateOrderResponse struct {
	OrderID     string `json:"order_id"`
	Status      string `json:"status"`
	CheckoutURL string `json:"checkout_url"`
}

// 创建订单
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

// 获取订单
func (a *App) GetOrder(id string) (*model.Order, error) {
	return a.store.GetOrder(id)
}

// 获取订单列表
func (a *App) ListOrders() ([]*model.Order, error) {
	return a.store.ListOrders()
}

// 标记订单已支付
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

// 标记订单已失败
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
