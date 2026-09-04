package model

import "time"

// OrderStatus 支付订单状态。
type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "created"   // 已创建
	OrderStatusPending   OrderStatus = "pending"   // 待支付
	OrderStatusPaid      OrderStatus = "paid"      // 已支付
	OrderStatusFailed    OrderStatus = "failed"    // 失败
	OrderStatusCancelled OrderStatus = "cancelled" // 已取消
)

// Order 模型层：支付中心订单，对应 payment_orders 表。
type Order struct {
	ID            string      `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`                        // 支付中心订单号
	MerchantOrder string      `gorm:"column:merchant_order;type:varchar(128);not null;index" json:"merchant_order"` // A站订单号
	MerchantSite     string      `gorm:"column:merchant_site;type:varchar(255);not null" json:"merchant_site"`   // A站站点标识
	MerchantID       uint        `gorm:"column:merchant_id;not null;default:0;index" json:"merchant_id"`         // 商户ID
	Channel          string      `gorm:"column:channel;type:varchar(64);not null" json:"channel"`                // 通道代码
	ChannelAccountID uint        `gorm:"column:channel_account_id;not null;default:0;index" json:"channel_account_id"` // 通道账号ID
	SiteBID          uint        `gorm:"column:site_b_id;not null;default:0;index" json:"site_b_id"`             // B站ID（下单时快照）
	SiteB            string      `gorm:"column:site_b;type:varchar(191);not null;default:''" json:"site_b"`      // B站域名（下单时快照）
	Provider         string      `gorm:"column:provider;type:varchar(64);not null" json:"provider"`              // 支付平台，例如 stripe
	Amount        int64       `gorm:"column:amount;not null" json:"amount"`                                   // 金额，最小货币单位
	Currency      string      `gorm:"column:currency;type:varchar(16);not null" json:"currency"`              // 币种
	ReturnURL     string      `gorm:"column:return_url;type:text;not null" json:"return_url"`                 // 同步返回地址
	NotifyURL     string      `gorm:"column:notify_url;type:text;not null" json:"notify_url"`                 // 异步通知地址
	NotifyVerify  string      `gorm:"column:notify_verify;type:text" json:"notify_verify,omitempty"`          // Shopyy 异步 verify 快照 JSON
	CheckoutURL   string      `gorm:"column:checkout_url;type:text" json:"checkout_url,omitempty"`            // 支付跳转地址
	ProviderRef   string      `gorm:"column:provider_ref;type:varchar(128);not null;default:''" json:"provider_ref,omitempty"` // 支付平台交易号
	Status        OrderStatus `gorm:"column:status;type:varchar(32);not null;index" json:"status"`            // 订单状态
	ErrorMessage  string      `gorm:"column:error_message;type:text" json:"error_message,omitempty"`          // 失败原因
	CreatedAt     time.Time   `gorm:"column:created_at;index" json:"created_at"`                              // 创建时间
	UpdatedAt     time.Time   `gorm:"column:updated_at" json:"updated_at"`                                    // 更新时间
}

// TableName 指定表名。
func (Order) TableName() string {
	return "payment_orders"
}
