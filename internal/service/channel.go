package service

import (
	"errors"
	"strings"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
)

var (
	ErrChannelNotFound             = errors.New("channel not found")
	ErrChannelNameExists           = errors.New("channel name exists")
	ErrChannelNameInvalid          = errors.New("channel name invalid")
	ErrChannelInterceptRangeInvalid = errors.New("channel intercept range invalid")
	ErrChannelSuccessSettingInvalid   = errors.New("channel success setting invalid")
)

// ChannelListItem 通道列表行，字段与前端 ChannelRow 对齐。
type ChannelListItem struct {
	ID                uint     `json:"id"`
	Name              string   `json:"name"`
	PackageName       string   `json:"packageName"`
	TotalAmount       float64  `json:"totalAmount"`
	Balance           float64  `json:"balance"`
	DailyOrderLimit   int      `json:"dailyOrderLimit"`
	DailyAmountLimit  float64  `json:"dailyAmountLimit"`
	DailyRecvCount    int      `json:"dailyRecvCount"`
	DailyRecvAmount   float64  `json:"dailyRecvAmount"`
	InterceptMode     string   `json:"interceptMode"`
	InterceptCurrency string   `json:"interceptCurrency"`
	InterceptMax      float64  `json:"interceptMax"`
	InterceptMin      float64  `json:"interceptMin"`
	Status            bool     `json:"status"`
	PaymentMode       string   `json:"paymentMode"`
	GatewayURL        string   `json:"gatewayUrl"`
	SuccessMode       string   `json:"successMode"`
	Countries         []string `json:"countries"`
	Currencies        []string `json:"currencies"`
	Remark            string   `json:"remark"`
	SettleRate        int      `json:"settleRate"`
	SiteBGroup        string   `json:"siteBGroup"`
	ChannelCode       string   `json:"channelCode"`
	PayCode           string   `json:"payCode"`
	OrderNoMode       string   `json:"orderNoMode"`
	ProductInfo       string   `json:"productInfo"`
	ReturnVerify      string   `json:"returnVerify"`
	OldCustomerDays   int      `json:"oldCustomerDays"`
	PayParams         string   `json:"payParams"`
	ReturnIPWhitelist string   `json:"returnIpWhitelist"`
	PayFrequency      int      `json:"payFrequency"`
	FailCount         int      `json:"failCount"`
	SuccessCount      int      `json:"successCount"`
	FailAutoClose     int      `json:"failAutoClose"`
	MutualHoldAmount  float64  `json:"mutualHoldAmount"`
	AmountLimitMode   string   `json:"amountLimitMode"`
	CalcCurrency      string   `json:"calcCurrency"`
	AllowCountries    []string `json:"allowCountries"`
	PreferCountries   []string `json:"preferCountries"`
	DisableCountries  []string `json:"disableCountries"`
	AllowCardTypes    []string `json:"allowCardTypes"`
	DisableCardTypes  []string `json:"disableCardTypes"`
	DisableCardBrands []string `json:"disableCardBrands"`
	Mixers            []string `json:"mixers"`
	CollectRule       string   `json:"collectRule"`
	ShipRange         string   `json:"shipRange"`
	Sort              int      `json:"sort"`
	AutoShip          bool     `json:"autoShip"`
	ReturnKeywords    string   `json:"returnKeywords"`
	DisableBrandWords string   `json:"disableBrandWords"`
	CreatedBy         string   `json:"createdBy"`
	CreatedAt         string   `json:"createdAt"`
	UpdatedBy         string   `json:"updatedBy"`
	UpdatedAt         string   `json:"updatedAt"`
}

// ChannelListQuery 列表筛选。
type ChannelListQuery struct {
	ID   *uint
	Name string
}

// CreateChannelRequest 新增通道。
type CreateChannelRequest struct {
	Name              string `json:"name" binding:"required"`
	PaymentMode       string `json:"paymentMode"`
	SiteBGroup        string `json:"siteBGroup"`
	OrderNoMode       string `json:"orderNoMode"`
	SettleRate        int    `json:"settleRate"`
	PayParams         string `json:"payParams"`
	ProductInfo       string `json:"productInfo"`
	ChannelCode       string `json:"channelCode"`
	PayCode           string `json:"payCode"`
	ReturnVerify      string `json:"returnVerify"`
	OldCustomerDays   int    `json:"oldCustomerDays"`
	ReturnIPWhitelist string `json:"returnIpWhitelist"`
}

// UpdateChannelRequest 编辑通道信息。
type UpdateChannelRequest struct {
	Name              string   `json:"name" binding:"required"`
	PayCode           string   `json:"payCode"`
	PaymentMode       string   `json:"paymentMode"`
	Mixers            []string `json:"mixers"`
	SettleRate        int      `json:"settleRate"`
	Remark            string   `json:"remark"`
	ReturnIPWhitelist string   `json:"returnIpWhitelist"`
	DisableBrandWords string   `json:"disableBrandWords"`
	CollectRule       string   `json:"collectRule"`
	ShipRange         string   `json:"shipRange"`
	OrderNoMode       string   `json:"orderNoMode"`
	Sort              int      `json:"sort"`
	ProductInfo       string   `json:"productInfo"`
	ReturnVerify      string   `json:"returnVerify"`
	OldCustomerDays   int      `json:"oldCustomerDays"`
	AutoShip          bool     `json:"autoShip"`
	ReturnKeywords    string   `json:"returnKeywords"`
}

// UpdateChannelLimitsRequest 限制配置。
type UpdateChannelLimitsRequest struct {
	ChannelCode       string   `json:"channelCode"`
	DailyAmountLimit  float64  `json:"dailyAmountLimit"`
	PayFrequency      int      `json:"payFrequency"`
	FailCount         int      `json:"failCount"`
	AmountLimitMode   string   `json:"amountLimitMode"`
	InterceptMax      float64  `json:"interceptMax"`
	Currencies        []string `json:"currencies"`
	AllowCountries    []string `json:"allowCountries"`
	AllowCardTypes    []string `json:"allowCardTypes"`
	DisableCardBrands []string `json:"disableCardBrands"`
	DailyOrderLimit   int      `json:"dailyOrderLimit"`
	MutualHoldAmount  float64  `json:"mutualHoldAmount"`
	SuccessCount      int      `json:"successCount"`
	FailAutoClose     int      `json:"failAutoClose"`
	CalcCurrency      string   `json:"calcCurrency"`
	InterceptMin      float64  `json:"interceptMin"`
	PreferCountries   []string `json:"preferCountries"`
	DisableCountries  []string `json:"disableCountries"`
	DisableCardTypes  []string `json:"disableCardTypes"`
}

// ListChannels 通道列表。
func (a *App) ListChannels(q ChannelListQuery) ([]ChannelListItem, error) {
	list, err := a.store.ListChannels(store.ChannelListFilter{
		ID:   q.ID,
		Name: q.Name,
	})
	if err != nil {
		return nil, err
	}
	out := make([]ChannelListItem, 0, len(list))
	for _, item := range list {
		out = append(out, a.toChannelListItem(item))
	}
	return out, nil
}

// CreateChannel 新增通道。
func (a *App) CreateChannel(req CreateChannelRequest, operator string) (*ChannelListItem, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrChannelNameInvalid
	}
	if exist, err := a.store.FindChannelByName(name); err == nil && exist != nil {
		return nil, ErrChannelNameExists
	} else if err != nil && !isNotFound(err) {
		return nil, err
	}
	item := &model.Channel{
		Name:              name,
		PackageName:       "",
		DailyOrderLimit:   0,
		DailyAmountLimit:  0,
		InterceptMode:     "reset",
		InterceptCurrency: "USD",
		InterceptMax:      0,
		InterceptMin:      0,
		Status:            model.ChannelStatusEnabled,
		PaymentMode:       defaultString(req.PaymentMode, "local"),
		Remark:            "",
		SettleRate:        req.SettleRate,
		SiteBGroup:        strings.TrimSpace(req.SiteBGroup),
		ChannelCode:       strings.TrimSpace(req.ChannelCode),
		PayCode:           strings.TrimSpace(req.PayCode),
		OrderNoMode:       defaultString(req.OrderNoMode, "site"),
		ProductInfo:       defaultString(req.ProductInfo, "kezhan"),
		ReturnVerify:      defaultString(req.ReturnVerify, "verify"),
		OldCustomerDays:   defaultInt(req.OldCustomerDays, 30),
		PayParams:         req.PayParams,
		ReturnIPWhitelist: req.ReturnIPWhitelist,
		PayFrequency:      0,
		FailCount:         0,
		SuccessCount:      0,
		FailAutoClose:     0,
		MutualHoldAmount:  0,
		AmountLimitMode:   "single",
		CalcCurrency:      "USD",
		AllowCountries:    model.StringList{},
		PreferCountries:   model.StringList{},
		DisableCountries:  model.StringList{},
		AllowCardTypes:    model.StringList{},
		DisableCardTypes:  model.StringList{},
		DisableCardBrands: model.StringList{},
		Countries:         model.StringList{},
		Currencies:        model.StringList{},
		Mixers:            model.StringList{},
		CollectRule:       "random",
		ShipRange:         "40-50",
		Sort:              1,
		AutoShip:          true,
		ReturnKeywords:    "",
		DisableBrandWords: "",
		CreatedBy:         operator,
		UpdatedBy:         operator,
	}
	if err := a.store.CreateChannel(item); err != nil {
		return nil, err
	}
	out := a.toChannelListItem(*item)
	return &out, nil
}

// UpdateChannel 编辑通道信息。
func (a *App) UpdateChannel(id uint, req UpdateChannelRequest, operator string) (*ChannelListItem, error) {
	item, err := a.store.GetChannelByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrChannelNotFound
		}
		return nil, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrChannelNameInvalid
	}
	if name != item.Name {
		if exist, err := a.store.FindChannelByName(name); err == nil && exist != nil && exist.ID != item.ID {
			return nil, ErrChannelNameExists
		} else if err != nil && !isNotFound(err) {
			return nil, err
		}
	}
	item.Name = name
	item.PayCode = strings.TrimSpace(req.PayCode)
	item.PaymentMode = defaultString(req.PaymentMode, item.PaymentMode)
	item.Mixers = model.StringList(req.Mixers)
	item.SettleRate = req.SettleRate
	item.Remark = strings.TrimSpace(req.Remark)
	item.ReturnIPWhitelist = req.ReturnIPWhitelist
	item.DisableBrandWords = req.DisableBrandWords
	item.CollectRule = defaultString(req.CollectRule, "random")
	shipRange := strings.TrimSpace(req.ShipRange)
	if shipRange == "" {
		shipRange = "40-50"
	}
	item.ShipRange = shipRange
	item.OrderNoMode = defaultString(req.OrderNoMode, "site")
	item.Sort = req.Sort
	if item.Sort <= 0 {
		item.Sort = 1
	}
	item.ProductInfo = defaultString(req.ProductInfo, "kezhan")
	item.ReturnVerify = defaultString(req.ReturnVerify, "verify")
	item.OldCustomerDays = defaultInt(req.OldCustomerDays, 30)
	item.AutoShip = req.AutoShip
	item.ReturnKeywords = req.ReturnKeywords
	item.UpdatedBy = operator
	if err := a.store.SaveChannel(item); err != nil {
		return nil, err
	}
	out := a.toChannelListItem(*item)
	return &out, nil
}

// UpdateChannelLimits 更新限制配置。
func (a *App) UpdateChannelLimits(id uint, req UpdateChannelLimitsRequest, operator string) (*ChannelListItem, error) {
	item, err := a.store.GetChannelByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrChannelNotFound
		}
		return nil, err
	}
	if err := validateInterceptRange(req.InterceptMin, req.InterceptMax); err != nil {
		return nil, err
	}
	if err := validateSuccessSetting(req.PayFrequency, req.SuccessCount, req.FailCount); err != nil {
		return nil, err
	}
	item.ChannelCode = strings.TrimSpace(req.ChannelCode)
	item.DailyAmountLimit = req.DailyAmountLimit
	item.PayFrequency = req.PayFrequency
	item.FailCount = req.FailCount
	item.AmountLimitMode = defaultString(req.AmountLimitMode, "single")
	item.InterceptMax = req.InterceptMax
	item.Currencies = model.StringList(req.Currencies)
	item.AllowCountries = model.StringList(req.AllowCountries)
	item.Countries = model.StringList(req.AllowCountries)
	item.AllowCardTypes = model.StringList(req.AllowCardTypes)
	item.DisableCardBrands = model.StringList(req.DisableCardBrands)
	item.DailyOrderLimit = req.DailyOrderLimit
	item.MutualHoldAmount = req.MutualHoldAmount
	item.SuccessCount = req.SuccessCount
	item.FailAutoClose = req.FailAutoClose
	calcCurrency := defaultString(req.CalcCurrency, "USD")
	item.CalcCurrency = calcCurrency
	item.InterceptCurrency = calcCurrency
	item.InterceptMin = req.InterceptMin
	item.PreferCountries = model.StringList(req.PreferCountries)
	item.DisableCountries = model.StringList(req.DisableCountries)
	item.DisableCardTypes = model.StringList(req.DisableCardTypes)
	item.UpdatedBy = operator
	if err := a.store.SaveChannel(item); err != nil {
		return nil, err
	}
	out := a.toChannelListItem(*item)
	return &out, nil
}

// SetChannelStatus 启用/禁用通道。
func (a *App) SetChannelStatus(id uint, enabled bool, operator string) (*ChannelListItem, error) {
	item, err := a.store.GetChannelByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrChannelNotFound
		}
		return nil, err
	}
	status := model.ChannelStatusDisabled
	if enabled {
		status = model.ChannelStatusEnabled
	}
	item.Status = status
	item.UpdatedBy = operator
	if err := a.store.SaveChannel(item); err != nil {
		return nil, err
	}
	out := a.toChannelListItem(*item)
	return &out, nil
}

func (a *App) toChannelListItem(item model.Channel) ChannelListItem {
	countries := []string(item.Countries)
	if len(countries) == 0 {
		countries = []string(item.AllowCountries)
	}
	return ChannelListItem{
		ID:                item.ID,
		Name:              item.Name,
		PackageName:       item.PackageName,
		TotalAmount:       0,
		Balance:           0,
		DailyOrderLimit:   item.DailyOrderLimit,
		DailyAmountLimit:  item.DailyAmountLimit,
		DailyRecvCount:    0,
		DailyRecvAmount:   0,
		InterceptMode:     item.InterceptMode,
		InterceptCurrency: item.InterceptCurrency,
		InterceptMax:      item.InterceptMax,
		InterceptMin:      item.InterceptMin,
		Status:            item.Status == model.ChannelStatusEnabled,
		PaymentMode:       item.PaymentMode,
		GatewayURL:        a.BuildGatewayURL(item.Name),
		SuccessMode:       resolveSuccessMode(item.PayFrequency, item.SuccessCount, item.FailCount),
		Countries:         countries,
		Currencies:        []string(item.Currencies),
		Remark:            item.Remark,
		SettleRate:        item.SettleRate,
		SiteBGroup:        item.SiteBGroup,
		ChannelCode:       item.ChannelCode,
		PayCode:           item.PayCode,
		OrderNoMode:       item.OrderNoMode,
		ProductInfo:       item.ProductInfo,
		ReturnVerify:      item.ReturnVerify,
		OldCustomerDays:   item.OldCustomerDays,
		PayParams:         item.PayParams,
		ReturnIPWhitelist: item.ReturnIPWhitelist,
		PayFrequency:      item.PayFrequency,
		FailCount:         item.FailCount,
		SuccessCount:      item.SuccessCount,
		FailAutoClose:     item.FailAutoClose,
		MutualHoldAmount:  item.MutualHoldAmount,
		AmountLimitMode:   item.AmountLimitMode,
		CalcCurrency:      item.CalcCurrency,
		AllowCountries:    []string(item.AllowCountries),
		PreferCountries:   []string(item.PreferCountries),
		DisableCountries:  []string(item.DisableCountries),
		AllowCardTypes:    []string(item.AllowCardTypes),
		DisableCardTypes:  []string(item.DisableCardTypes),
		DisableCardBrands: []string(item.DisableCardBrands),
		Mixers:            []string(item.Mixers),
		CollectRule:       item.CollectRule,
		ShipRange:         item.ShipRange,
		Sort:              item.Sort,
		AutoShip:          item.AutoShip,
		ReturnKeywords:    item.ReturnKeywords,
		DisableBrandWords: item.DisableBrandWords,
		CreatedBy:         item.CreatedBy,
		CreatedAt:         item.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedBy:         item.UpdatedBy,
		UpdatedAt:         item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func validateInterceptRange(min, max float64) error {
	if min > max {
		return ErrChannelInterceptRangeInvalid
	}
	return nil
}

func resolveSuccessMode(payFrequency, successCount, failCount int) string {
	if payFrequency > 0 && (successCount > 0 || failCount > 0) {
		return "limited"
	}
	return "unlimited"
}

func validateSuccessSetting(payFrequency, successCount, failCount int) error {
	hasFrequency := payFrequency > 0
	hasCount := successCount > 0 || failCount > 0
	if hasFrequency && !hasCount {
		return ErrChannelSuccessSettingInvalid
	}
	if hasCount && !hasFrequency {
		return ErrChannelSuccessSettingInvalid
	}
	return nil
}

func defaultString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func defaultInt(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
