package service

import (
	"errors"
	"strings"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
)

var (
	ErrChannelAccountNotFound           = errors.New("channel account not found")
	ErrChannelAccountNoInvalid          = errors.New("channel account no invalid")
	ErrChannelAccountChannelInvalid    = errors.New("channel account channel invalid")
	ErrChannelAccountSiteBInvalid      = errors.New("channel account site b invalid")
	ErrChannelAccountChannelSiteBExists = errors.New("channel account channel site b exists")
	ErrChannelAccountSuccessSettingInvalid = errors.New("channel account success setting invalid")
	ErrChannelAccountGroupBound            = errors.New("channel account group bound")
)

// ChannelAccountListItem 通道账号列表行。
type ChannelAccountListItem struct {
	ID                uint     `json:"id"`
	ChannelID         uint     `json:"channelId"`
	Channel           string   `json:"channel"`
	SiteBID           uint     `json:"siteBId"`
	SiteB             string   `json:"siteB"`
	AccountNo         string   `json:"accountNo"`
	Alias             string   `json:"alias"`
	Remark            string   `json:"remark"`
	PaymentMethod     string   `json:"paymentMethod"`
	GroupNames          []string `json:"groupNames"`
	AssignedUser      string   `json:"assignedUser"`
	TotalReceived     float64  `json:"totalReceived"`
	Status            bool     `json:"status"`
	ResetTimezone     string   `json:"resetTimezone"`
	ResetHour         int      `json:"resetHour"`
	DailyOrderLimit   int      `json:"dailyOrderLimit"`
	DailyAmountLimit  float64  `json:"dailyAmountLimit"`
	DailyRecvCount    int      `json:"dailyRecvCount"`
	DailyRecvAmount   float64  `json:"dailyRecvAmount"`
	InterceptMode     string   `json:"interceptMode"`
	InterceptCurrency string   `json:"interceptCurrency"`
	InterceptMax      float64  `json:"interceptMax"`
	InterceptMin      float64  `json:"interceptMin"`
	AmountLimitMode   string   `json:"amountLimitMode"`
	CalcCurrency      string   `json:"calcCurrency"`
	Currencies        []string `json:"currencies"`
	AllowCountries    []string `json:"allowCountries"`
	PreferCountries   []string `json:"preferCountries"`
	DisableCountries  []string `json:"disableCountries"`
	AllowCardTypes    []string `json:"allowCardTypes"`
	DisableCardTypes  []string `json:"disableCardTypes"`
	DisableCardBrands []string `json:"disableCardBrands"`
	PayFrequency      int      `json:"payFrequency"`
	SuccessCountLimit int      `json:"successCountLimit"`
	MaxSuccessCount   int      `json:"maxSuccessCount"`
	SuccessMode       string   `json:"successMode"`
	Sort              int      `json:"sort"`
	AppID             string   `json:"appId"`
	MerchantID        string   `json:"merchantId"`
	WebSecret         string   `json:"webSecret"`
	PrivateKey        string   `json:"privateKey"`
	Environment       string   `json:"environment"`
	UnpaidClosed      bool     `json:"unpaidClosed"`
	RestrictedClosed  bool     `json:"restrictedClosed"`
	CannotOpenAt8     bool     `json:"cannotOpenAt8"`
	CreatedBy         string   `json:"createdBy"`
	CreatedAt         string   `json:"createdAt"`
	UpdatedBy         string   `json:"updatedBy"`
	UpdatedAt         string   `json:"updatedAt"`
}

// ChannelAccountListQuery 列表筛选。
type ChannelAccountListQuery struct {
	ID           *uint
	ChannelID    *uint
	ChannelName  string
	Alias        string
	Remark       string
	CreatedFrom  string
	CreatedTo    string
	GroupID          *uint
	AssignedUserID   *uint
	ListFilter       string
	ScopeUserID      uint
	ScopeUserType    string
}

// CreateChannelAccountRequest 新增通道账号。
type CreateChannelAccountRequest struct {
	ChannelID        uint     `json:"channelId" binding:"required"`
	SiteBID          uint     `json:"siteBId"`
	AccountNo        string   `json:"accountNo" binding:"required"`
	Alias            string   `json:"alias"`
	Status           *bool    `json:"status"`
	ResetHour        *int     `json:"resetHour"`
	DailyOrderLimit  int      `json:"dailyOrderLimit"`
	DailyAmountLimit float64  `json:"dailyAmountLimit"`
	CalcCurrency     string   `json:"calcCurrency"`
	Currencies       []string `json:"currencies"`
	PreferCountries  []string `json:"preferCountries"`
	DisableCountries []string `json:"disableCountries"`
	Sort             int      `json:"sort"`
	AppID            string   `json:"appId"`
	WebSecret        string   `json:"webSecret"`
	PrivateKey       string   `json:"privateKey"`
	Remark           string   `json:"remark"`
}

// UpdateChannelAccountRequest 编辑通道账号。
type UpdateChannelAccountRequest struct {
	AccountNo   string `json:"accountNo" binding:"required"`
	Alias       string `json:"alias"`
	Status      *bool  `json:"status"`
	SiteBID     uint   `json:"siteBId"`
	Sort        int    `json:"sort"`
	AppID       string `json:"appId"`
	MerchantID  string `json:"merchantId"`
	WebSecret   string `json:"webSecret"`
	PrivateKey  string `json:"privateKey"`
	Environment string `json:"environment"`
	Remark      string `json:"remark"`
}

// UpdateChannelAccountLimitsRequest 限制配置。
type UpdateChannelAccountLimitsRequest struct {
	ResetHour         int      `json:"resetHour"`
	DailyAmountLimit  float64  `json:"dailyAmountLimit"`
	Currencies        []string `json:"currencies"`
	SuccessCountLimit int      `json:"successCountLimit"`
	AllowCountries    []string `json:"allowCountries"`
	AmountLimitMode   string   `json:"amountLimitMode"`
	InterceptMax      float64  `json:"interceptMax"`
	AllowCardTypes    []string `json:"allowCardTypes"`
	DisableCardBrands []string `json:"disableCardBrands"`
	DailyOrderLimit   int      `json:"dailyOrderLimit"`
	MaxSuccessCount   int      `json:"maxSuccessCount"`
	PayFrequency      int      `json:"payFrequency"`
	PreferCountries   []string `json:"preferCountries"`
	DisableCountries  []string `json:"disableCountries"`
	CalcCurrency      string   `json:"calcCurrency"`
	InterceptMin      float64  `json:"interceptMin"`
	DisableCardTypes  []string `json:"disableCardTypes"`
}

// ListChannelAccounts 通道账号列表。
func (a *App) ListChannelAccounts(q ChannelAccountListQuery) ([]ChannelAccountListItem, error) {
	filter := store.ChannelAccountListFilter{
		ID:             q.ID,
		ChannelID:      q.ChannelID,
		ChannelName:    q.ChannelName,
		Alias:          q.Alias,
		Remark:         q.Remark,
		CreatedFrom:    q.CreatedFrom,
		CreatedTo:      q.CreatedTo,
		GroupID:        q.GroupID,
		AssignedUserID: q.AssignedUserID,
		ListFilter:     q.ListFilter,
	}
	if q.ScopeUserType == model.UserTypeDistribution && q.ScopeUserID > 0 {
		filter.AssignedUserID = &q.ScopeUserID
	}
	list, err := a.store.ListChannelAccounts(filter)
	if err != nil {
		return nil, err
	}
	channelMap, err := a.loadChannelNameMap()
	if err != nil {
		return nil, err
	}
	siteBMap, err := a.loadSiteBDomainMap()
	if err != nil {
		return nil, err
	}
	groupMap, err := a.loadChannelAccountGroupNamesMap(list)
	if err != nil {
		return nil, err
	}
	assignedUserIDs := make([]uint, 0)
	seenAssigned := map[uint]struct{}{}
	for _, item := range list {
		if item.AssignedUserID == 0 {
			continue
		}
		if _, ok := seenAssigned[item.AssignedUserID]; ok {
			continue
		}
		seenAssigned[item.AssignedUserID] = struct{}{}
		assignedUserIDs = append(assignedUserIDs, item.AssignedUserID)
	}
	assignedUserMap, err := a.loadAssignedUserNameMap(assignedUserIDs)
	if err != nil {
		return nil, err
	}
	out := make([]ChannelAccountListItem, 0, len(list))
	for _, item := range list {
		groupNames := groupMap[item.ID]
		if groupNames == nil {
			groupNames = []string{}
		}
		out = append(out, toChannelAccountListItem(item, channelMap, siteBMap, groupNames, assignedUserMap))
	}
	return out, nil
}

// CreateChannelAccount 新增通道账号。
func (a *App) CreateChannelAccount(req CreateChannelAccountRequest, operator string) (*ChannelAccountListItem, error) {
	accountNo := strings.TrimSpace(req.AccountNo)
	if accountNo == "" {
		return nil, ErrChannelAccountNoInvalid
	}
	if _, err := a.store.GetChannelByID(req.ChannelID); err != nil {
		if isNotFound(err) {
			return nil, ErrChannelAccountChannelInvalid
		}
		return nil, err
	}
	if req.SiteBID == 0 {
		return nil, ErrChannelAccountSiteBInvalid
	}
	if _, err := a.store.GetSiteBByID(req.SiteBID); err != nil {
		if isNotFound(err) {
			return nil, ErrChannelAccountSiteBInvalid
		}
		return nil, err
	}
	if exist, err := a.store.FindChannelAccountByChannelAndSiteB(req.ChannelID, req.SiteBID); err == nil && exist != nil {
		return nil, ErrChannelAccountChannelSiteBExists
	} else if err != nil && !isNotFound(err) {
		return nil, err
	}
	status := model.ChannelAccountStatusEnabled
	if req.Status != nil && !*req.Status {
		status = model.ChannelAccountStatusDisabled
	}
	resetHour := 0
	if req.ResetHour != nil {
		resetHour = *req.ResetHour
	}
	calcCurrency := defaultString(req.CalcCurrency, "USD")
	currencies := model.StringList(req.Currencies)
	if len(currencies) == 0 && calcCurrency != "" {
		currencies = model.StringList{calcCurrency}
	}
	item := &model.ChannelAccount{
		ChannelID:         req.ChannelID,
		SiteBID:           req.SiteBID,
		AccountNo:         accountNo,
		Alias:             strings.TrimSpace(req.Alias),
		Status:            status,
		ResetHour:         resetHour,
		ResetTimezone:     "北京重置时间",
		DailyOrderLimit:   req.DailyOrderLimit,
		DailyAmountLimit:  req.DailyAmountLimit,
		InterceptMode:     "reset",
		InterceptCurrency: calcCurrency,
		AmountLimitMode:   "reset",
		CalcCurrency:      calcCurrency,
		Currencies:        currencies,
		PreferCountries:   model.StringList(req.PreferCountries),
		DisableCountries:  model.StringList(req.DisableCountries),
		Sort:              req.Sort,
		AppID:             strings.TrimSpace(req.AppID),
		WebSecret:         strings.TrimSpace(req.WebSecret),
		PrivateKey:        req.PrivateKey,
		PaymentMethod:     "card",
		Remark:            strings.TrimSpace(req.Remark),
		CreatedBy:         operator,
		UpdatedBy:         operator,
	}
	if err := a.store.CreateChannelAccount(item); err != nil {
		return nil, err
	}
	return a.getChannelAccountItem(item.ID)
}

// UpdateChannelAccount 编辑通道账号。
func (a *App) UpdateChannelAccount(id uint, req UpdateChannelAccountRequest, operator string) (*ChannelAccountListItem, error) {
	item, err := a.store.GetChannelAccountByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrChannelAccountNotFound
		}
		return nil, err
	}
	accountNo := strings.TrimSpace(req.AccountNo)
	if accountNo == "" {
		return nil, ErrChannelAccountNoInvalid
	}
	if req.SiteBID == 0 {
		return nil, ErrChannelAccountSiteBInvalid
	}
	if _, err := a.store.GetSiteBByID(req.SiteBID); err != nil {
		if isNotFound(err) {
			return nil, ErrChannelAccountSiteBInvalid
		}
		return nil, err
	}
	if req.SiteBID != item.SiteBID {
		if exist, err := a.store.FindChannelAccountByChannelAndSiteB(item.ChannelID, req.SiteBID); err == nil && exist != nil && exist.ID != item.ID {
			return nil, ErrChannelAccountChannelSiteBExists
		} else if err != nil && !isNotFound(err) {
			return nil, err
		}
	}
	item.AccountNo = accountNo
	item.Alias = strings.TrimSpace(req.Alias)
	if req.Status != nil {
		if *req.Status {
			item.Status = model.ChannelAccountStatusEnabled
		} else {
			item.Status = model.ChannelAccountStatusDisabled
		}
	}
	item.SiteBID = req.SiteBID
	item.Sort = req.Sort
	item.AppID = strings.TrimSpace(req.AppID)
	item.MerchantID = strings.TrimSpace(req.MerchantID)
	item.WebSecret = strings.TrimSpace(req.WebSecret)
	item.PrivateKey = req.PrivateKey
	env := strings.TrimSpace(req.Environment)
	if env == "" {
		env = "live"
	}
	item.Environment = env
	item.Remark = strings.TrimSpace(req.Remark)
	item.UpdatedBy = operator
	if err := a.store.SaveChannelAccount(item); err != nil {
		return nil, err
	}
	return a.getChannelAccountItem(item.ID)
}

// UpdateChannelAccountLimits 更新限制配置。
func (a *App) UpdateChannelAccountLimits(id uint, req UpdateChannelAccountLimitsRequest, operator string) (*ChannelAccountListItem, error) {
	item, err := a.store.GetChannelAccountByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrChannelAccountNotFound
		}
		return nil, err
	}
	if err := validateInterceptRange(req.InterceptMin, req.InterceptMax); err != nil {
		return nil, err
	}
	if err := validateAccountSuccessSetting(req.PayFrequency, req.SuccessCountLimit); err != nil {
		return nil, err
	}
	item.ResetHour = req.ResetHour
	item.DailyAmountLimit = req.DailyAmountLimit
	item.Currencies = model.StringList(req.Currencies)
	item.SuccessCountLimit = req.SuccessCountLimit
	item.AllowCountries = model.StringList(req.AllowCountries)
	amountLimitMode := defaultString(req.AmountLimitMode, "reset")
	item.AmountLimitMode = amountLimitMode
	item.InterceptMode = amountLimitMode
	item.InterceptMax = req.InterceptMax
	item.AllowCardTypes = model.StringList(req.AllowCardTypes)
	item.DisableCardBrands = model.StringList(req.DisableCardBrands)
	item.DailyOrderLimit = req.DailyOrderLimit
	item.MaxSuccessCount = req.MaxSuccessCount
	item.PayFrequency = req.PayFrequency
	item.PreferCountries = model.StringList(req.PreferCountries)
	item.DisableCountries = model.StringList(req.DisableCountries)
	calcCurrency := defaultString(req.CalcCurrency, "USD")
	item.CalcCurrency = calcCurrency
	item.InterceptCurrency = calcCurrency
	item.InterceptMin = req.InterceptMin
	item.DisableCardTypes = model.StringList(req.DisableCardTypes)
	item.UpdatedBy = operator
	if err := a.store.SaveChannelAccount(item); err != nil {
		return nil, err
	}
	return a.getChannelAccountItem(item.ID)
}

// SetChannelAccountStatus 启用/禁用通道账号。
func (a *App) SetChannelAccountStatus(id uint, enabled bool, operator string) (*ChannelAccountListItem, error) {
	item, err := a.store.GetChannelAccountByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrChannelAccountNotFound
		}
		return nil, err
	}
	if enabled {
		item.Status = model.ChannelAccountStatusEnabled
	} else {
		item.Status = model.ChannelAccountStatusDisabled
	}
	item.UpdatedBy = operator
	if err := a.store.SaveChannelAccount(item); err != nil {
		return nil, err
	}
	return a.getChannelAccountItem(item.ID)
}

// DeleteChannelAccount 删除通道账号，已绑定分组时不允许删除。
func (a *App) DeleteChannelAccount(id uint) error {
	if _, err := a.store.GetChannelAccountByID(id); err != nil {
		if isNotFound(err) {
			return ErrChannelAccountNotFound
		}
		return err
	}
	count, err := a.store.CountChannelGroupMembersByAccountID(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrChannelAccountGroupBound
	}
	return a.store.DeleteChannelAccount(id)
}

func (a *App) getChannelAccountItem(id uint) (*ChannelAccountListItem, error) {
	item, err := a.store.GetChannelAccountByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrChannelAccountNotFound
		}
		return nil, err
	}
	channelMap, err := a.loadChannelNameMap()
	if err != nil {
		return nil, err
	}
	siteBMap, err := a.loadSiteBDomainMap()
	if err != nil {
		return nil, err
	}
	groupMap, err := a.loadChannelAccountGroupNamesMap([]model.ChannelAccount{*item})
	if err != nil {
		return nil, err
	}
	groupNames := groupMap[item.ID]
	if groupNames == nil {
		groupNames = []string{}
	}
	assignedUserMap, err := a.loadAssignedUserNameMap([]uint{item.AssignedUserID})
	if err != nil {
		return nil, err
	}
	out := toChannelAccountListItem(*item, channelMap, siteBMap, groupNames, assignedUserMap)
	return &out, nil
}

func (a *App) loadChannelNameMap() (map[uint]string, error) {
	list, err := a.store.ListChannels(store.ChannelListFilter{})
	if err != nil {
		return nil, err
	}
	result := make(map[uint]string, len(list))
	for _, item := range list {
		result[item.ID] = item.Name
	}
	return result, nil
}

func (a *App) loadSiteBDomainMap() (map[uint]string, error) {
	list, err := a.store.ListSiteBs(store.SiteBListFilter{})
	if err != nil {
		return nil, err
	}
	result := make(map[uint]string, len(list))
	for _, item := range list {
		result[item.ID] = item.Domain
	}
	return result, nil
}

func (a *App) loadChannelGroupNameMap() (map[uint]string, error) {
	list, err := a.store.ListChannelGroups(store.ChannelGroupListFilter{})
	if err != nil {
		return nil, err
	}
	result := make(map[uint]string, len(list))
	for _, item := range list {
		result[item.ID] = item.Name
	}
	return result, nil
}

func (a *App) loadChannelAccountGroupNamesMap(accounts []model.ChannelAccount) (map[uint][]string, error) {
	if len(accounts) == 0 {
		return map[uint][]string{}, nil
	}
	accountIDs := make([]uint, 0, len(accounts))
	for _, item := range accounts {
		accountIDs = append(accountIDs, item.ID)
	}
	members, err := a.store.ListChannelGroupMembersByAccountIDs(accountIDs)
	if err != nil {
		return nil, err
	}
	groupNameMap, err := a.loadChannelGroupNameMap()
	if err != nil {
		return nil, err
	}
	result := make(map[uint][]string, len(accounts))
	for _, member := range members {
		name := groupNameMap[member.GroupID]
		if name == "" {
			continue
		}
		result[member.ChannelAccountID] = append(result[member.ChannelAccountID], name)
	}
	return result, nil
}

func toChannelAccountListItem(item model.ChannelAccount, channelMap map[uint]string, siteBMap map[uint]string, groupNames []string, assignedUserMap map[uint]string) ChannelAccountListItem {
	channelName := channelMap[item.ChannelID]
	siteBDomain := siteBMap[item.SiteBID]
	assignedUser := assignedUserMap[item.AssignedUserID]
	return ChannelAccountListItem{
		ID:                item.ID,
		ChannelID:         item.ChannelID,
		Channel:           channelName,
		SiteBID:           item.SiteBID,
		SiteB:             siteBDomain,
		AccountNo:         item.AccountNo,
		Alias:             item.Alias,
		Remark:            item.Remark,
		PaymentMethod:     item.PaymentMethod,
		GroupNames:        groupNames,
		AssignedUser:      assignedUser,
		TotalReceived:     item.TotalReceived,
		Status:            item.Status == model.ChannelAccountStatusEnabled,
		ResetTimezone:     item.ResetTimezone,
		ResetHour:         item.ResetHour,
		DailyOrderLimit:   item.DailyOrderLimit,
		DailyAmountLimit:  item.DailyAmountLimit,
		DailyRecvCount:    item.DailyRecvCount,
		DailyRecvAmount:   item.DailyRecvAmount,
		InterceptMode:     item.InterceptMode,
		InterceptCurrency: item.InterceptCurrency,
		InterceptMax:      item.InterceptMax,
		InterceptMin:      item.InterceptMin,
		AmountLimitMode:   item.AmountLimitMode,
		CalcCurrency:      item.CalcCurrency,
		Currencies:        []string(item.Currencies),
		AllowCountries:    []string(item.AllowCountries),
		PreferCountries:   []string(item.PreferCountries),
		DisableCountries:  []string(item.DisableCountries),
		AllowCardTypes:    []string(item.AllowCardTypes),
		DisableCardTypes:  []string(item.DisableCardTypes),
		DisableCardBrands: []string(item.DisableCardBrands),
		PayFrequency:      item.PayFrequency,
		SuccessCountLimit: item.SuccessCountLimit,
		MaxSuccessCount:   item.MaxSuccessCount,
		SuccessMode:       resolveAccountSuccessMode(item.PayFrequency, item.SuccessCountLimit),
		Sort:              item.Sort,
		AppID:             item.AppID,
		MerchantID:        item.MerchantID,
		WebSecret:         item.WebSecret,
		PrivateKey:        item.PrivateKey,
		Environment:       item.Environment,
		UnpaidClosed:      item.UnpaidClosed,
		RestrictedClosed:  item.RestrictedClosed,
		CannotOpenAt8:     item.CannotOpenAt8,
		CreatedBy:         item.CreatedBy,
		CreatedAt:         item.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedBy:         item.UpdatedBy,
		UpdatedAt:         item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func resolveAccountSuccessMode(payFrequency, successCountLimit int) string {
	if payFrequency > 0 && successCountLimit > 0 {
		return "limited"
	}
	return "unlimited"
}

func validateAccountSuccessSetting(payFrequency, successCountLimit int) error {
	hasFrequency := payFrequency > 0
	hasCount := successCountLimit > 0
	if hasFrequency && !hasCount {
		return ErrChannelAccountSuccessSettingInvalid
	}
	if hasCount && !hasFrequency {
		return ErrChannelAccountSuccessSettingInvalid
	}
	return nil
}
