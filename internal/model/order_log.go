package model

import "time"

// 订单日志类型（与管理端展示文案对应）。
const (
	OrderLogTypeRecord     = "log_record"  // 日志记录
	OrderLogTypeFailLog    = "fail_log"    // 失败日志
	OrderLogTypeFailRecord = "fail_record" // 失败记录（终态失败/取消等）
)

// OrderLog 支付订单过程日志，对应 payment_order_logs 表。
type OrderLog struct {
	ID            uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrderID       string    `gorm:"column:order_id;type:varchar(64);not null;default:'';index" json:"order_id"`             // 支付中心订单号
	MerchantID    uint      `gorm:"column:merchant_id;not null;default:0;index" json:"merchant_id"`                         // 商户ID
	MerchantOrder string    `gorm:"column:merchant_order;type:varchar(128);not null;default:'';index" json:"merchant_order"` // A站订单号
	Type          string    `gorm:"column:type;type:varchar(32);not null;index" json:"type"`                                // 日志类型
	Remark        string    `gorm:"column:remark;type:text;not null" json:"remark"`                                         // 备注
	CreatedAt     time.Time `gorm:"column:created_at;index" json:"created_at"`                                              // 创建时间
}

// TableName 指定表名。
func (OrderLog) TableName() string {
	return "payment_order_logs"
}
