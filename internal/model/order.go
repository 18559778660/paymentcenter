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
	Fee           int64       `gorm:"column:fee;not null;default:0" json:"fee"`                                              // Stripe 手续费，最小货币单位
	UsdDiff       *float64    `gorm:"column:usd_diff;type:decimal(18,4)" json:"usd_diff,omitempty"`                         // 按商户汇率偏差折算后的美元金额
	Status        OrderStatus `gorm:"column:status;type:varchar(32);not null;index" json:"status"`            // 订单状态
	ErrorMessage  string      `gorm:"column:error_message;type:text" json:"error_message,omitempty"`          // 失败原因
	CustomerName  string      `gorm:"column:customer_name;type:varchar(128);not null;default:''" json:"customer_name"` // 客户姓名快照
	CustomerEmail string      `gorm:"column:customer_email;type:varchar(191);not null;default:''" json:"customer_email"` // 客户邮箱快照
	CustomerPhone string      `gorm:"column:customer_phone;type:varchar(64);not null;default:''" json:"customer_phone"` // 客户电话快照
	CustomerIP    string      `gorm:"column:customer_ip;type:varchar(64);not null;default:''" json:"customer_ip"`       // 客户 IP 快照
	BillCountry   string      `gorm:"column:bill_country;type:varchar(64);not null;default:''" json:"bill_country"`     // 账单国家
	BillState     string      `gorm:"column:bill_state;type:varchar(128);not null;default:''" json:"bill_state"`       // 账单省/州
	BillCity      string      `gorm:"column:bill_city;type:varchar(128);not null;default:''" json:"bill_city"`         // 账单城市
	BillZip       string      `gorm:"column:bill_zip;type:varchar(32);not null;default:''" json:"bill_zip"`             // 账单邮编
	BillAddress   string      `gorm:"column:bill_address;type:varchar(512);not null;default:''" json:"bill_address"`   // 账单地址
	ShipCountry   string      `gorm:"column:ship_country;type:varchar(64);not null;default:''" json:"ship_country"`     // 配送国家
	ShipState     string      `gorm:"column:ship_state;type:varchar(128);not null;default:''" json:"ship_state"`       // 配送省/州
	ShipCity      string      `gorm:"column:ship_city;type:varchar(128);not null;default:''" json:"ship_city"`         // 配送城市
	ShipZip       string      `gorm:"column:ship_zip;type:varchar(32);not null;default:''" json:"ship_zip"`             // 配送邮编
	ShipAddress   string      `gorm:"column:ship_address;type:varchar(512);not null;default:''" json:"ship_address"`   // 配送地址
	OrderGoods    string      `gorm:"column:order_goods;type:text" json:"order_goods,omitempty"`                            // 商品快照 JSON
	CreatedAt     time.Time   `gorm:"column:created_at;index" json:"created_at"`                              // 创建时间
	UpdatedAt     time.Time   `gorm:"column:updated_at" json:"updated_at"`                                    // 更新时间
}

// TableName 指定表名。
func (Order) TableName() string {
	return "payment_orders"
}
