package model

import "time"

const (
	ChannelGroupStatusDisabled = 0
	ChannelGroupStatusEnabled  = 1
)

// ChannelGroup 通道分组配置表。
type ChannelGroup struct {
	// ID 分组主键。
	ID uint `gorm:"column:id;primaryKey;autoIncrement;comment:分组ID" json:"id"`
	// Code 分组 CODE，全局唯一，对接网关使用。
	Code string `gorm:"column:code;type:varchar(64);not null;uniqueIndex;comment:分组CODE" json:"code"`
	// Name 分组名称，全局唯一。
	Name string `gorm:"column:name;type:varchar(64);not null;uniqueIndex;comment:分组名称" json:"name"`
	// TotalAmount 累计收款金额，单位 USD。
	TotalAmount float64 `gorm:"column:total_amount;type:decimal(18,2);not null;default:0;comment:累计收款USD" json:"total_amount"`
	// Balance 可用余额，单位 USD。
	Balance float64 `gorm:"column:balance;type:decimal(18,2);not null;default:0;comment:余额USD" json:"balance"`
	// DailyOrderLimit 日限单数，0 表示不限制。
	DailyOrderLimit int `gorm:"column:daily_order_limit;not null;default:0;comment:日限单数" json:"daily_order_limit"`
	// DailyAmountLimit 日限金额，单位 USD，0 表示不限制。
	DailyAmountLimit float64 `gorm:"column:daily_amount_limit;type:decimal(18,2);not null;default:0;comment:日限金额USD" json:"daily_amount_limit"`
	// DailyRecvCount 日收笔数。
	DailyRecvCount int `gorm:"column:daily_recv_count;not null;default:0;comment:日收笔数" json:"daily_recv_count"`
	// DailyRecvAmount 日收金额，单位 USD。
	DailyRecvAmount float64 `gorm:"column:daily_recv_amount;type:decimal(18,2);not null;default:0;comment:日收金额USD" json:"daily_recv_amount"`
	// PayFrequency 支付频率限制，单位天。
	PayFrequency int `gorm:"column:pay_frequency;not null;default:0;comment:支付频率天" json:"pay_frequency"`
	// FailCount 失败次数限制，0 表示不限制。
	FailCount int `gorm:"column:fail_count;not null;default:0;comment:失败次数限制" json:"fail_count"`
	// SuccessCount 成功次数限制，0 表示不限制。
	SuccessCount int `gorm:"column:success_count;not null;default:0;comment:成功次数限制" json:"success_count"`
	// AmountLimitMode 金额限制模式：reset 重置、intercept 拦截。
	AmountLimitMode string `gorm:"column:amount_limit_mode;type:varchar(16);not null;default:'reset';comment:金额限制模式" json:"amount_limit_mode"`
	// CalcCurrency 金额计算使用的货币代码。
	CalcCurrency string `gorm:"column:calc_currency;type:varchar(16);not null;default:'USD';comment:计算货币" json:"calc_currency"`
	// InterceptMax 拦截最大金额。
	InterceptMax float64 `gorm:"column:intercept_max;type:decimal(18,2);not null;default:0;comment:拦截最大金额" json:"intercept_max"`
	// InterceptMin 拦截最小金额。
	InterceptMin float64 `gorm:"column:intercept_min;type:decimal(18,2);not null;default:0;comment:拦截最小金额" json:"intercept_min"`
	// OldCustomerDays 老客户判断天数。
	OldCustomerDays int `gorm:"column:old_customer_days;not null;default:30;comment:老客户判断天数" json:"old_customer_days"`
	// AllowCountries 仅支持的国家代码列表。
	AllowCountries StringList `gorm:"column:allow_countries;type:json;not null;comment:仅支持国家" json:"allow_countries"`
	// PreferCountries 优先使用的国家代码列表。
	PreferCountries StringList `gorm:"column:prefer_countries;type:json;not null;comment:优先国家" json:"prefer_countries"`
	// DisableCountries 禁用的国家代码列表。
	DisableCountries StringList `gorm:"column:disable_countries;type:json;not null;comment:禁用国家" json:"disable_countries"`
	// AllowCardTypes 仅支持的卡类型列表。
	AllowCardTypes StringList `gorm:"column:allow_card_types;type:json;not null;comment:仅支持卡类型" json:"allow_card_types"`
	// DisableCardTypes 禁用的卡类型列表。
	DisableCardTypes StringList `gorm:"column:disable_card_types;type:json;not null;comment:禁用卡类型" json:"disable_card_types"`
	// DisableCardBrands 禁用的卡品牌/卡头列表。
	DisableCardBrands StringList `gorm:"column:disable_card_brands;type:json;not null;comment:禁用卡头" json:"disable_card_brands"`
	// CollectRule 收款规则：random 随机、round 轮询。
	CollectRule string `gorm:"column:collect_rule;type:varchar(16);not null;default:'random';comment:收款规则" json:"collect_rule"`
	// AutoShip 是否自动发货。
	AutoShip bool `gorm:"column:auto_ship;not null;comment:自动发货" json:"auto_ship"`
	// Status 分组状态：1 启用、0 禁用。
	Status int `gorm:"column:status;not null;default:1;index;comment:状态 1启用 0禁用" json:"status"`
	// CreatedBy 创建人账号。
	CreatedBy string `gorm:"column:created_by;type:varchar(64);not null;default:'';comment:创建人" json:"created_by"`
	// UpdatedBy 最后更新人账号。
	UpdatedBy string `gorm:"column:updated_by;type:varchar(64);not null;default:'';comment:更新人" json:"updated_by"`
	// CreatedAt 创建时间。
	CreatedAt time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	// UpdatedAt 最后更新时间。
	UpdatedAt time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (ChannelGroup) TableName() string {
	return "channel_groups"
}
