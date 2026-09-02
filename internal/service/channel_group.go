package service

import (
	"errors"
	"strings"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
)

var (
	ErrChannelGroupNotFound              = errors.New("channel group not found")
	ErrChannelGroupCodeExists            = errors.New("channel group code exists")
	ErrChannelGroupNameExists            = errors.New("channel group name exists")
	ErrChannelGroupCodeInvalid           = errors.New("channel group code invalid")
	ErrChannelGroupNameInvalid           = errors.New("channel group name invalid")
	ErrChannelGroupInterceptRangeInvalid = errors.New("channel group intercept range invalid")
	ErrChannelGroupSuccessSettingInvalid = errors.New("channel group success setting invalid")
	ErrChannelGroupMemberBound           = errors.New("channel group member bound")
)

// ChannelGroupListItem 通道分组列表行，字段与前端 ChannelGroupRow 对齐。
type ChannelGroupListItem struct {
	ID                    uint     `json:"id"`
	Code                  string   `json:"code"`
	Name                  string   `json:"name"`
	TotalAmount           float64  `json:"totalAmount"`
	Balance               float64  `json:"balance"`
	DailyOrderLimit       int      `json:"dailyOrderLimit"`
	DailyAmountLimit      float64  `json:"dailyAmountLimit"`
	DailyRecvCount        int      `json:"dailyRecvCount"`
	DailyRecvAmount       float64  `json:"dailyRecvAmount"`
	PayFrequencyDays      int      `json:"payFrequencyDays"`
	FailLimitCount        int      `json:"failLimitCount"`
	SuccessLimitCount     int      `json:"successLimitCount"`
	InterceptMode         string   `json:"interceptMode"`
	InterceptCurrency     string   `json:"interceptCurrency"`
	InterceptMax          float64  `json:"interceptMax"`
	InterceptMin          float64  `json:"interceptMin"`
	OldCustomerDays       int      `json:"oldCustomerDays"`
	PreferCountries       []string `json:"preferCountries"`
	DisableCountries      []string `json:"disableCountries"`
	AllowCountries        []string `json:"allowCountries"`
	AllowCardTypes        []string `json:"allowCardTypes"`
	DisableCardTypes      []string `json:"disableCardTypes"`
	DisableCardBrands     []string `json:"disableCardBrands"`
	CollectRule           string   `json:"collectRule"`
	AutoShip              bool     `json:"autoShip"`
	AvailableAccountCount int      `json:"availableAccountCount"`
	MemberCount           int      `json:"memberCount"`
	GatewayURL            string   `json:"gatewayUrl"`
	CreatedBy             string   `json:"createdBy"`
	CreatedAt             string   `json:"createdAt"`
	UpdatedBy             string   `json:"updatedBy"`
	UpdatedAt             string   `json:"updatedAt"`
}

// ChannelGroupListQuery 列表筛选。
type ChannelGroupListQuery struct {
	ID   *uint
	Code string
}

// CreateChannelGroupRequest 新增通道分组。
type CreateChannelGroupRequest struct {
	Name              string   `json:"name" binding:"required"`
	Code              string   `json:"code" binding:"required"`
	OldCustomerDays   int      `json:"oldCustomerDays"`
	PayFrequencyDays  int      `json:"payFrequencyDays"`
	FailLimitCount    int      `json:"failLimitCount"`
	SuccessLimitCount int      `json:"successLimitCount"`
	InterceptMode     string   `json:"interceptMode"`
	InterceptCurrency string   `json:"interceptCurrency"`
	InterceptMax      float64  `json:"interceptMax"`
	InterceptMin      float64  `json:"interceptMin"`
	DailyOrderLimit   int      `json:"dailyOrderLimit"`
	DailyAmountLimit  float64  `json:"dailyAmountLimit"`
	PreferCountries   []string `json:"preferCountries"`
	DisableCountries  []string `json:"disableCountries"`
	DisableCardBrands []string `json:"disableCardBrands"`
	DisableCardTypes  []string `json:"disableCardTypes"`
	AllowCountries    []string `json:"allowCountries"`
	AllowCardTypes    []string `json:"allowCardTypes"`
	CollectRule       string   `json:"collectRule"`
	AutoShip          bool     `json:"autoShip"`
}

// UpdateChannelGroupRequest 编辑通道分组（不含 code/name）。
type UpdateChannelGroupRequest struct {
	OldCustomerDays   int      `json:"oldCustomerDays"`
	PayFrequencyDays  int      `json:"payFrequencyDays"`
	FailLimitCount    int      `json:"failLimitCount"`
	SuccessLimitCount int      `json:"successLimitCount"`
	InterceptMode     string   `json:"interceptMode"`
	InterceptCurrency string   `json:"interceptCurrency"`
	InterceptMax      float64  `json:"interceptMax"`
	InterceptMin      float64  `json:"interceptMin"`
	DailyOrderLimit   int      `json:"dailyOrderLimit"`
	DailyAmountLimit  float64  `json:"dailyAmountLimit"`
	PreferCountries   []string `json:"preferCountries"`
	DisableCountries  []string `json:"disableCountries"`
	DisableCardBrands []string `json:"disableCardBrands"`
	DisableCardTypes  []string `json:"disableCardTypes"`
	AllowCountries    []string `json:"allowCountries"`
	AllowCardTypes    []string `json:"allowCardTypes"`
	CollectRule       string   `json:"collectRule"`
	AutoShip          bool     `json:"autoShip"`
}

// ListChannelGroups 通道分组列表。
func (a *App) ListChannelGroups(q ChannelGroupListQuery) ([]ChannelGroupListItem, error) {
	list, err := a.store.ListChannelGroups(store.ChannelGroupListFilter{
		ID:   q.ID,
		Code: q.Code,
	})
	if err != nil {
		return nil, err
	}
	out := make([]ChannelGroupListItem, 0, len(list))
	for _, item := range list {
		row, err := a.toChannelGroupListItem(item)
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, nil
}

// CreateChannelGroup 新增通道分组。
func (a *App) CreateChannelGroup(req CreateChannelGroupRequest, operator string) (*ChannelGroupListItem, error) {
	code, name, err := normalizeChannelGroupIdentity(req.Code, req.Name)
	if err != nil {
		return nil, err
	}
	if err := validateChannelGroupLimits(req.InterceptMin, req.InterceptMax, req.PayFrequencyDays, req.SuccessLimitCount, req.FailLimitCount); err != nil {
		return nil, err
	}
	if exist, err := a.store.FindChannelGroupByCode(code); err == nil && exist != nil {
		return nil, ErrChannelGroupCodeExists
	} else if err != nil && !isNotFound(err) {
		return nil, err
	}
	if exist, err := a.store.FindChannelGroupByName(name); err == nil && exist != nil {
		return nil, ErrChannelGroupNameExists
	} else if err != nil && !isNotFound(err) {
		return nil, err
	}
	item := &model.ChannelGroup{
		Code:            code,
		Name:            name,
		TotalAmount:     0,
		Balance:         0,
		DailyRecvCount:  0,
		DailyRecvAmount: 0,
		Status:          model.ChannelGroupStatusEnabled,
		CollectRule:     defaultString(req.CollectRule, "random"),
		AutoShip:        req.AutoShip,
		CreatedBy:       operator,
		UpdatedBy:       operator,
	}
	applyChannelGroupLimits(item, channelGroupLimitsPayload{
		OldCustomerDays:   req.OldCustomerDays,
		PayFrequencyDays:  req.PayFrequencyDays,
		FailLimitCount:    req.FailLimitCount,
		SuccessLimitCount: req.SuccessLimitCount,
		InterceptMode:     req.InterceptMode,
		InterceptCurrency: req.InterceptCurrency,
		InterceptMax:      req.InterceptMax,
		InterceptMin:      req.InterceptMin,
		DailyOrderLimit:   req.DailyOrderLimit,
		DailyAmountLimit:  req.DailyAmountLimit,
		PreferCountries:   req.PreferCountries,
		DisableCountries:  req.DisableCountries,
		DisableCardBrands: req.DisableCardBrands,
		DisableCardTypes:  req.DisableCardTypes,
		AllowCountries:    req.AllowCountries,
		AllowCardTypes:    req.AllowCardTypes,
	})
	if err := a.store.CreateChannelGroup(item); err != nil {
		return nil, err
	}
	return a.getChannelGroupItem(item.ID)
}

// UpdateChannelGroup 编辑通道分组。
func (a *App) UpdateChannelGroup(id uint, req UpdateChannelGroupRequest, operator string) (*ChannelGroupListItem, error) {
	item, err := a.store.GetChannelGroupByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrChannelGroupNotFound
		}
		return nil, err
	}
	if err := validateChannelGroupLimits(req.InterceptMin, req.InterceptMax, req.PayFrequencyDays, req.SuccessLimitCount, req.FailLimitCount); err != nil {
		return nil, err
	}
	applyChannelGroupLimits(item, channelGroupLimitsPayload{
		OldCustomerDays:   req.OldCustomerDays,
		PayFrequencyDays:  req.PayFrequencyDays,
		FailLimitCount:    req.FailLimitCount,
		SuccessLimitCount: req.SuccessLimitCount,
		InterceptMode:     req.InterceptMode,
		InterceptCurrency: req.InterceptCurrency,
		InterceptMax:      req.InterceptMax,
		InterceptMin:      req.InterceptMin,
		DailyOrderLimit:   req.DailyOrderLimit,
		DailyAmountLimit:  req.DailyAmountLimit,
		PreferCountries:   req.PreferCountries,
		DisableCountries:  req.DisableCountries,
		DisableCardBrands: req.DisableCardBrands,
		DisableCardTypes:  req.DisableCardTypes,
		AllowCountries:    req.AllowCountries,
		AllowCardTypes:    req.AllowCardTypes,
	})
	item.CollectRule = defaultString(req.CollectRule, "random")
	item.AutoShip = req.AutoShip
	item.UpdatedBy = operator
	if err := a.store.SaveChannelGroup(item); err != nil {
		return nil, err
	}
	return a.getChannelGroupItem(item.ID)
}

// DeleteChannelGroup 删除通道分组，仍有账号绑定时不允许删除。
func (a *App) DeleteChannelGroup(id uint) error {
	if _, err := a.store.GetChannelGroupByID(id); err != nil {
		if isNotFound(err) {
			return ErrChannelGroupNotFound
		}
		return err
	}
	count, err := a.store.CountChannelGroupMembersByGroupID(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrChannelGroupMemberBound
	}
	return a.store.DeleteChannelGroup(id)
}

// ChannelGroupAccountItem 分组账号列表行。
type ChannelGroupAccountItem struct {
	ID            uint   `json:"id"`
	ChannelName   string `json:"channelName"`
	InGroup       bool   `json:"inGroup"`
	AccountStatus bool   `json:"accountStatus"`
	SiteB         string `json:"siteB"`
	Channel       string `json:"channel"`
	Remark        string `json:"remark"`
	PaymentMethod string `json:"paymentMethod"`
}

// ChannelGroupAccountListQuery 分组账号列表筛选。
type ChannelGroupAccountListQuery struct {
	ChannelID *uint
}

// ListChannelGroupAccounts 列出全部通道账号，并标记是否归属当前分组。
func (a *App) ListChannelGroupAccounts(groupID uint, q ChannelGroupAccountListQuery) ([]ChannelGroupAccountItem, error) {
	group, err := a.store.GetChannelGroupByID(groupID)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrChannelGroupNotFound
		}
		return nil, err
	}
	accounts, err := a.store.ListChannelAccounts(store.ChannelAccountListFilter{})
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
	memberAccountIDs, err := a.store.ListChannelGroupMemberAccountIDs(group.ID)
	if err != nil {
		return nil, err
	}
	memberSet := make(map[uint]struct{}, len(memberAccountIDs))
	for _, accountID := range memberAccountIDs {
		memberSet[accountID] = struct{}{}
	}
	out := make([]ChannelGroupAccountItem, 0, len(accounts))
	for _, item := range accounts {
		if q.ChannelID != nil && *q.ChannelID > 0 && item.ChannelID != *q.ChannelID {
			continue
		}
		channelName := strings.TrimSpace(item.Alias)
		if channelName == "" {
			channelName = item.AccountNo
		}
		_, inGroup := memberSet[item.ID]
		out = append(out, ChannelGroupAccountItem{
			ID:            item.ID,
			ChannelName:   channelName,
			InGroup:       inGroup,
			AccountStatus: item.Status == model.ChannelAccountStatusEnabled,
			SiteB:         siteBMap[item.SiteBID],
			Channel:       channelMap[item.ChannelID],
			Remark:        item.Remark,
			PaymentMethod: item.PaymentMethod,
		})
	}
	return out, nil
}

// SetChannelGroupAccountMembership 设置账号是否归属分组。
func (a *App) SetChannelGroupAccountMembership(groupID, accountID uint, inGroup bool, operator string) error {
	group, err := a.store.GetChannelGroupByID(groupID)
	if err != nil {
		if isNotFound(err) {
			return ErrChannelGroupNotFound
		}
		return err
	}
	account, err := a.store.GetChannelAccountByID(accountID)
	if err != nil {
		if isNotFound(err) {
			return ErrChannelAccountNotFound
		}
		return err
	}
	if inGroup {
		return a.store.AddChannelGroupMember(group.ID, account.ID)
	}
	return a.store.RemoveChannelGroupMember(group.ID, account.ID)
}

func (a *App) getChannelGroupItem(id uint) (*ChannelGroupListItem, error) {
	item, err := a.store.GetChannelGroupByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrChannelGroupNotFound
		}
		return nil, err
	}
	row, err := a.toChannelGroupListItem(*item)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// toChannelGroupListItem 转换通道分组列表项。
func (a *App) toChannelGroupListItem(item model.ChannelGroup) (ChannelGroupListItem, error) {
	availableCount, err := a.store.CountEnabledChannelAccountsByGroupID(item.ID)
	if err != nil {
		return ChannelGroupListItem{}, err
	}
	memberCount, err := a.store.CountChannelGroupMembersByGroupID(item.ID)
	if err != nil {
		return ChannelGroupListItem{}, err
	}
	return ChannelGroupListItem{
		ID:                    item.ID,
		Code:                  item.Code,
		Name:                  item.Name,
		TotalAmount:           item.TotalAmount,
		Balance:               item.Balance,
		DailyOrderLimit:       item.DailyOrderLimit,
		DailyAmountLimit:      item.DailyAmountLimit,
		DailyRecvCount:        item.DailyRecvCount,
		DailyRecvAmount:       item.DailyRecvAmount,
		PayFrequencyDays:      item.PayFrequency,
		FailLimitCount:        item.FailCount,
		SuccessLimitCount:     item.SuccessCount,
		InterceptMode:         item.AmountLimitMode,
		InterceptCurrency:     item.CalcCurrency,
		InterceptMax:          item.InterceptMax,
		InterceptMin:          item.InterceptMin,
		OldCustomerDays:       item.OldCustomerDays,
		PreferCountries:       []string(item.PreferCountries),
		DisableCountries:      []string(item.DisableCountries),
		AllowCountries:        []string(item.AllowCountries),
		AllowCardTypes:        []string(item.AllowCardTypes),
		DisableCardTypes:      []string(item.DisableCardTypes),
		DisableCardBrands:     []string(item.DisableCardBrands),
		CollectRule:           item.CollectRule,
		AutoShip:              item.AutoShip,
		AvailableAccountCount: int(availableCount),
		MemberCount:           int(memberCount),
		GatewayURL:            a.BuildGroupGatewayURL(item.Code),
		CreatedBy:             item.CreatedBy,
		CreatedAt:             item.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedBy:             item.UpdatedBy,
		UpdatedAt:             item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

type channelGroupLimitsPayload struct {
	OldCustomerDays   int
	PayFrequencyDays  int
	FailLimitCount    int
	SuccessLimitCount int
	InterceptMode     string
	InterceptCurrency string
	InterceptMax      float64
	InterceptMin      float64
	DailyOrderLimit   int
	DailyAmountLimit  float64
	PreferCountries   []string
	DisableCountries  []string
	DisableCardBrands []string
	DisableCardTypes  []string
	AllowCountries    []string
	AllowCardTypes    []string
}

func applyChannelGroupLimits(item *model.ChannelGroup, payload channelGroupLimitsPayload) {
	item.OldCustomerDays = defaultInt(payload.OldCustomerDays, 30)
	item.PayFrequency = payload.PayFrequencyDays
	item.FailCount = payload.FailLimitCount
	item.SuccessCount = payload.SuccessLimitCount
	item.AmountLimitMode = defaultString(payload.InterceptMode, "reset")
	item.CalcCurrency = defaultString(payload.InterceptCurrency, "USD")
	item.InterceptMax = payload.InterceptMax
	item.InterceptMin = payload.InterceptMin
	item.DailyOrderLimit = payload.DailyOrderLimit
	item.DailyAmountLimit = payload.DailyAmountLimit
	item.PreferCountries = model.StringList(payload.PreferCountries)
	item.DisableCountries = model.StringList(payload.DisableCountries)
	item.DisableCardBrands = model.StringList(payload.DisableCardBrands)
	item.DisableCardTypes = model.StringList(payload.DisableCardTypes)
	item.AllowCountries = model.StringList(payload.AllowCountries)
	item.AllowCardTypes = model.StringList(payload.AllowCardTypes)
}

func normalizeChannelGroupIdentity(code, name string) (string, string, error) {
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)
	if code == "" {
		return "", "", ErrChannelGroupCodeInvalid
	}
	if name == "" {
		return "", "", ErrChannelGroupNameInvalid
	}
	return code, name, nil
}

func validateChannelGroupLimits(interceptMin, interceptMax float64, payFrequency, successCount, failCount int) error {
	if interceptMin > interceptMax {
		return ErrChannelGroupInterceptRangeInvalid
	}
	if err := validateSuccessSetting(payFrequency, successCount, failCount); err != nil {
		return ErrChannelGroupSuccessSettingInvalid
	}
	return nil
}
