package model

import "time"

const (
	ChannelStatusDisabled = 0
	ChannelStatusEnabled  = 1
)

// Channel 支付通道配置表。
type Channel struct {
	// ID 通道主键。
	ID uint `gorm:"column:id;primaryKey;autoIncrement;comment:通道ID" json:"id"`
	// Name 通道名称，全局唯一。
	Name string `gorm:"column:name;type:varchar(64);not null;uniqueIndex;comment:通道名称" json:"name"`
	// PackageName 关联压缩包文件名。
	PackageName string `gorm:"column:package_name;type:varchar(128);not null;default:'';comment:压缩包文件名" json:"package_name"`
	// DailyOrderLimit 日限单数，0 表示不限制。
	DailyOrderLimit int `gorm:"column:daily_order_limit;not null;default:0;comment:日限单数" json:"daily_order_limit"`
	// DailyAmountLimit 日限金额，单位 USD，0 表示不限制。
	DailyAmountLimit float64 `gorm:"column:daily_amount_limit;type:decimal(18,2);not null;default:0;comment:日限金额USD" json:"daily_amount_limit"`
	// InterceptMode 拦截模式：reset 重置、keep 保持。
	InterceptMode string `gorm:"column:intercept_mode;type:varchar(16);not null;default:'reset';comment:拦截模式" json:"intercept_mode"`
	// InterceptCurrency 拦截金额使用的货币代码。
	InterceptCurrency string `gorm:"column:intercept_currency;type:varchar(16);not null;default:'USD';comment:拦截货币" json:"intercept_currency"`
	// InterceptMax 拦截最大金额。
	InterceptMax float64 `gorm:"column:intercept_max;type:decimal(18,2);not null;default:0;comment:拦截最大金额" json:"intercept_max"`
	// InterceptMin 拦截最小金额。
	InterceptMin float64 `gorm:"column:intercept_min;type:decimal(18,2);not null;default:0;comment:拦截最小金额" json:"intercept_min"`
	// Status 通道状态：1 启用、0 禁用。
	Status int `gorm:"column:status;not null;default:1;index;comment:状态 1启用 0禁用" json:"status"`
	// PaymentMode 支付模式：local 本地支付、checkout 收银台、embedded 系统内嵌。
	PaymentMode string `gorm:"column:payment_mode;type:varchar(32);not null;default:'local';comment:支付模式" json:"payment_mode"`
	// Remark 备注说明。
	Remark string `gorm:"column:remark;type:varchar(512);not null;default:'';comment:备注" json:"remark"`
	// SettleRate 结算比例，单位百分比。
	SettleRate int `gorm:"column:settle_rate;not null;default:0;comment:结算比例" json:"settle_rate"`
	// SiteBGroup B 站分组标识。
	SiteBGroup string `gorm:"column:site_b_group;type:varchar(64);not null;default:'';comment:B站分组" json:"site_b_group"`
	// ChannelCode 通道 CODE，对接支付平台使用。
	ChannelCode string `gorm:"column:channel_code;type:varchar(64);not null;default:'';comment:通道CODE" json:"channel_code"`
	// PayCode 支付 CODE，对接支付平台使用。
	PayCode string `gorm:"column:pay_code;type:varchar(64);not null;default:'';comment:支付CODE" json:"pay_code"`
	// OrderNoMode 订单号设置：site 网站订单号、strip_suffix 去后缀订单号。
	OrderNoMode string `gorm:"column:order_no_mode;type:varchar(32);not null;default:'site';comment:订单号设置" json:"order_no_mode"`
	// ProductInfo 商品信息来源：kezhan 壳站信息、product 产品数据、rebuild 重组数据。
	ProductInfo string `gorm:"column:product_info;type:varchar(32);not null;default:'kezhan';comment:商品信息" json:"product_info"`
	// ReturnVerify 返回页验证：verify 验证、skip 不验证。
	ReturnVerify string `gorm:"column:return_verify;type:varchar(16);not null;default:'verify';comment:返回页验证" json:"return_verify"`
	// OldCustomerDays 老客户判断天数。
	OldCustomerDays int `gorm:"column:old_customer_days;not null;default:30;comment:老客户判断天数" json:"old_customer_days"`
	// PayParams 支付参数配置。
	PayParams string `gorm:"column:pay_params;type:text;not null;comment:支付参数" json:"pay_params"`
	// ReturnIPWhitelist 返回验证 IP 白名单，多行或逗号分隔。
	ReturnIPWhitelist string `gorm:"column:return_ip_whitelist;type:text;not null;comment:返回验证IP白名单" json:"return_ip_whitelist"`
	// PayFrequency 支付频率限制，单位天；与成功/失败次数配合构成成功设置。
	PayFrequency int `gorm:"column:pay_frequency;not null;default:0;comment:支付频率天" json:"pay_frequency"`
	// FailCount 失败次数限制，0 表示不限制；需配合支付频率使用。
	FailCount int `gorm:"column:fail_count;not null;default:0;comment:失败次数限制" json:"fail_count"`
	// SuccessCount 成功次数限制，0 表示不限制；需配合支付频率使用。
	SuccessCount int `gorm:"column:success_count;not null;default:0;comment:成功次数限制" json:"success_count"`
	// FailAutoClose 失败自动关闭阈值，0 表示不自动关闭。
	FailAutoClose int `gorm:"column:fail_auto_close;not null;default:0;comment:失败自动关闭" json:"fail_auto_close"`
	// MutualHoldAmount 互抛限制金额，单位 USD。
	MutualHoldAmount float64 `gorm:"column:mutual_hold_amount;type:decimal(18,2);not null;default:0;comment:互抛限制金额USD" json:"mutual_hold_amount"`
	// AmountLimitMode 金额限制模式：single 单笔、intercept 拦截。
	AmountLimitMode string `gorm:"column:amount_limit_mode;type:varchar(16);not null;default:'single';comment:金额限制模式" json:"amount_limit_mode"`
	// CalcCurrency 金额计算使用的货币代码。
	CalcCurrency string `gorm:"column:calc_currency;type:varchar(16);not null;default:'USD';comment:计算货币" json:"calc_currency"`
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
	// Countries 限制国家列表，列表页展示用。
	Countries StringList `gorm:"column:countries;type:json;not null;comment:限制国家" json:"countries"`
	// Currencies 支持的支付货币代码列表。
	Currencies StringList `gorm:"column:currencies;type:json;not null;comment:支付货币" json:"currencies"`
	// Mixers 一抛混流器/账号池列表。
	Mixers StringList `gorm:"column:mixers;type:json;not null;comment:一抛混流器" json:"mixers"`
	// CollectRule 收款规则：random 随机、round 轮询。
	CollectRule string `gorm:"column:collect_rule;type:varchar(16);not null;default:'random';comment:收款规则" json:"collect_rule"`
	// ShipRange 发货范围，如 40-50。
	ShipRange string `gorm:"column:ship_range;type:varchar(32);not null;default:'40-50';comment:发货范围" json:"ship_range"`
	// Sort 排序值，越小越靠前。
	Sort int `gorm:"column:sort;not null;default:1;comment:排序" json:"sort"`
	// AutoShip 是否自动发货。
	AutoShip bool `gorm:"column:auto_ship;not null;comment:自动发货" json:"auto_ship"`
	// ReturnKeywords 返回页拦截关键词。
	ReturnKeywords string `gorm:"column:return_keywords;type:text;not null;comment:返回关键词" json:"return_keywords"`
	// DisableBrandWords 禁用品牌词。
	DisableBrandWords string `gorm:"column:disable_brand_words;type:text;not null;comment:禁用品牌词" json:"disable_brand_words"`
	// CreatedBy 创建人账号。
	CreatedBy string `gorm:"column:created_by;type:varchar(64);not null;default:'';comment:创建人" json:"created_by"`
	// UpdatedBy 最后更新人账号。
	UpdatedBy string `gorm:"column:updated_by;type:varchar(64);not null;default:'';comment:更新人" json:"updated_by"`
	// CreatedAt 创建时间。
	CreatedAt time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	// UpdatedAt 最后更新时间。
	UpdatedAt time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (Channel) TableName() string {
	return "channels"
}
