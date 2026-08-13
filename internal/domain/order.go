package domain

import "time"

type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "created"
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusFailed    OrderStatus = "failed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID            string      `json:"id"`
	MerchantOrder  string      `json:"merchant_order"`
	MerchantSite   string      `json:"merchant_site"`
	Channel        string      `json:"channel"`
	Provider       string      `json:"provider"`
	Amount         int64       `json:"amount"`
	Currency       string      `json:"currency"`
	ReturnURL      string      `json:"return_url"`
	NotifyURL      string      `json:"notify_url"`
	CheckoutURL    string      `json:"checkout_url,omitempty"`
	ProviderRef    string      `json:"provider_ref,omitempty"`
	Status         OrderStatus `json:"status"`
	ErrorMessage   string      `json:"error_message,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}
