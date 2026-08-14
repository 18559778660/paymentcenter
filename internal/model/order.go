package model

import "time"

type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "created"
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusFailed    OrderStatus = "failed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// Order 是支付中心订单模型，对应 payment_orders 表。
type Order struct {
	ID            string      `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	MerchantOrder string      `gorm:"column:merchant_order;type:varchar(128);not null;index" json:"merchant_order"`
	MerchantSite  string      `gorm:"column:merchant_site;type:varchar(255);not null" json:"merchant_site"`
	Channel       string      `gorm:"column:channel;type:varchar(64);not null" json:"channel"`
	Provider      string      `gorm:"column:provider;type:varchar(64);not null" json:"provider"`
	Amount        int64       `gorm:"column:amount;not null" json:"amount"`
	Currency      string      `gorm:"column:currency;type:varchar(16);not null" json:"currency"`
	ReturnURL     string      `gorm:"column:return_url;type:text;not null" json:"return_url"`
	NotifyURL     string      `gorm:"column:notify_url;type:text;not null" json:"notify_url"`
	CheckoutURL   string      `gorm:"column:checkout_url;type:text" json:"checkout_url,omitempty"`
	ProviderRef   string      `gorm:"column:provider_ref;type:varchar(128);not null;default:''" json:"provider_ref,omitempty"`
	Status        OrderStatus `gorm:"column:status;type:varchar(32);not null;index" json:"status"`
	ErrorMessage  string      `gorm:"column:error_message;type:text" json:"error_message,omitempty"`
	CreatedAt     time.Time   `gorm:"column:created_at;index" json:"created_at"`
	UpdatedAt     time.Time   `gorm:"column:updated_at" json:"updated_at"`
}

func (Order) TableName() string {
	return "payment_orders"
}
