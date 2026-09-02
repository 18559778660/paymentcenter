package model

import "time"

const (
	ChannelAccountStatusDisabled = 0
	ChannelAccountStatusEnabled  = 1
)

// ChannelAccount 通道账号。
type ChannelAccount struct {
	ID                uint       `gorm:"column:id;primaryKey;autoIncrement;comment:账号ID" json:"id"`
	ChannelID         uint       `gorm:"column:channel_id;not null;index;uniqueIndex:idx_channel_accounts_channel_site_b;comment:通道ID" json:"channel_id"`
	SiteBID           uint       `gorm:"column:site_b_id;not null;index;uniqueIndex:idx_channel_accounts_channel_site_b;comment:B站ID" json:"site_b_id"`
	AccountNo         string     `gorm:"column:account_no;type:varchar(128);not null;comment:账号名称" json:"account_no"`
	Alias             string     `gorm:"column:alias;type:varchar(128);not null;default:'';comment:别名" json:"alias"`
	Status            int        `gorm:"column:status;not null;default:1;index;comment:状态 1启用 0禁用" json:"status"`
	ResetHour         int        `gorm:"column:reset_hour;not null;default:0;comment:重置小时" json:"reset_hour"`
	ResetTimezone     string     `gorm:"column:reset_timezone;type:varchar(32);not null;default:'北京重置时间';comment:重置时区" json:"reset_timezone"`
	DailyOrderLimit   int        `gorm:"column:daily_order_limit;not null;default:0;comment:日限单数" json:"daily_order_limit"`
	DailyAmountLimit  float64    `gorm:"column:daily_amount_limit;type:decimal(18,2);not null;default:0;comment:日限金额USD" json:"daily_amount_limit"`
	DailyRecvCount    int        `gorm:"column:daily_recv_count;not null;default:0;comment:日收笔数" json:"daily_recv_count"`
	DailyRecvAmount   float64    `gorm:"column:daily_recv_amount;type:decimal(18,2);not null;default:0;comment:日收金额USD" json:"daily_recv_amount"`
	InterceptMode     string     `gorm:"column:intercept_mode;type:varchar(16);not null;default:'reset';comment:拦截模式" json:"intercept_mode"`
	InterceptCurrency string     `gorm:"column:intercept_currency;type:varchar(16);not null;default:'USD';comment:拦截货币" json:"intercept_currency"`
	InterceptMax      float64    `gorm:"column:intercept_max;type:decimal(18,2);not null;default:0;comment:拦截最大金额" json:"intercept_max"`
	InterceptMin      float64    `gorm:"column:intercept_min;type:decimal(18,2);not null;default:0;comment:拦截最小金额" json:"intercept_min"`
	AmountLimitMode   string     `gorm:"column:amount_limit_mode;type:varchar(16);not null;default:'reset';comment:金额限制模式" json:"amount_limit_mode"`
	CalcCurrency      string     `gorm:"column:calc_currency;type:varchar(16);not null;default:'USD';comment:计算货币" json:"calc_currency"`
	Currencies        StringList `gorm:"column:currencies;type:json;not null;comment:支付货币" json:"currencies"`
	AllowCountries    StringList `gorm:"column:allow_countries;type:json;not null;comment:仅支持国家" json:"allow_countries"`
	PreferCountries   StringList `gorm:"column:prefer_countries;type:json;not null;comment:优先国家" json:"prefer_countries"`
	DisableCountries  StringList `gorm:"column:disable_countries;type:json;not null;comment:禁用国家" json:"disable_countries"`
	AllowCardTypes    StringList `gorm:"column:allow_card_types;type:json;not null;comment:仅支持卡类型" json:"allow_card_types"`
	DisableCardTypes  StringList `gorm:"column:disable_card_types;type:json;not null;comment:禁用卡类型" json:"disable_card_types"`
	DisableCardBrands StringList `gorm:"column:disable_card_brands;type:json;not null;comment:禁用卡头" json:"disable_card_brands"`
	PayFrequency      int        `gorm:"column:pay_frequency;not null;default:0;comment:支付频率天" json:"pay_frequency"`
	SuccessCountLimit int        `gorm:"column:success_count_limit;not null;default:0;comment:指定时间内限制成功次数" json:"success_count_limit"`
	MaxSuccessCount   int        `gorm:"column:max_success_count;not null;default:0;comment:最多收款笔数" json:"max_success_count"`
	Sort              int        `gorm:"column:sort;not null;default:0;comment:排序" json:"sort"`
	AppID             string     `gorm:"column:app_id;type:varchar(512);not null;default:'';comment:应用ID" json:"app_id"`
	PublicKey         string     `gorm:"column:public_key;type:varchar(512);not null;default:'';comment:公钥" json:"public_key"`
	MerchantID        string     `gorm:"column:merchant_id;type:varchar(128);not null;default:'';comment:商户ID" json:"merchant_id"`
	WebSecret         string     `gorm:"column:web_secret;type:varchar(512);not null;default:'';comment:web秘钥" json:"web_secret"`
	PrivateKey        string     `gorm:"column:private_key;type:text;not null;comment:私钥" json:"private_key"`
	Environment       string     `gorm:"column:environment;type:varchar(32);not null;default:'live';comment:环境" json:"environment"`
	PaymentMethod     string     `gorm:"column:payment_method;type:varchar(32);not null;default:'card';comment:支付方式" json:"payment_method"`
	AssignedUserID    uint       `gorm:"column:assigned_user_id;not null;default:0;index;comment:分配子账号用户ID，0表示未分配" json:"assigned_user_id"`
	AssignedUser      string     `gorm:"-" json:"assigned_user"`
	TotalReceived     float64    `gorm:"column:total_received;type:decimal(18,2);not null;default:0;comment:总收款USD" json:"total_received"`
	UnpaidClosed      bool       `gorm:"column:unpaid_closed;not null;default:0;comment:跳转未付关闭" json:"unpaid_closed"`
	RestrictedClosed  bool       `gorm:"column:restricted_closed;not null;default:0;comment:账号受限关闭" json:"restricted_closed"`
	CannotOpenAt8     bool       `gorm:"column:cannot_open_at8;not null;default:0;comment:B站打不开" json:"cannot_open_at8"`
	Remark            string     `gorm:"column:remark;type:varchar(512);not null;default:'';comment:备注" json:"remark"`
	CreatedBy         string     `gorm:"column:created_by;type:varchar(64);not null;default:'';comment:创建人" json:"created_by"`
	UpdatedBy         string     `gorm:"column:updated_by;type:varchar(64);not null;default:'';comment:更新人" json:"updated_by"`
	CreatedAt         time.Time  `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (ChannelAccount) TableName() string {
	return "channel_accounts"
}
