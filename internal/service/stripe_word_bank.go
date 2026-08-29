package service

import (
	"errors"
	"strings"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
)

var (
	ErrStripeWordBankNotFound              = errors.New("stripe word bank not found")
	ErrStripeWordBankNameExists            = errors.New("stripe word bank name exists")
	ErrStripeWordBankNameInvalid           = errors.New("stripe word bank name invalid")
	ErrStripeWordBankNameMustStartWithSlash = errors.New("stripe word bank name must start with slash")
	ErrStripeWordBankNameMustNotStartWithSlash = errors.New("stripe word bank name must not start with slash")
	ErrStripeWordBankConfigItemInvalid     = errors.New("stripe word bank config item invalid")
)

var stripeWordBankConfigItems = map[string]struct{}{
	model.StripeWordBankConfigWebhook:  {},
	model.StripeWordBankConfigCallback: {},
	model.StripeWordBankConfigDirectory: {},
}

// StripeWordBankListItem Stripe 单词库列表行。
type StripeWordBankListItem struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	UsageCount int    `json:"usageCount"`
	ConfigItem string `json:"configItem"`
}

// StripeWordBankListQuery 列表筛选。
type StripeWordBankListQuery struct {
	ConfigItem string
}

// CreateStripeWordBankRequest 新增 Stripe 单词。
type CreateStripeWordBankRequest struct {
	Name       string `json:"name" binding:"required"`
	ConfigItem string `json:"configItem" binding:"required"`
}

// UpdateStripeWordBankRequest 编辑 Stripe 单词（仅配置项）。
type UpdateStripeWordBankRequest struct {
	ConfigItem string `json:"configItem" binding:"required"`
}

// ListStripeWordBanks Stripe 单词库列表。
func (a *App) ListStripeWordBanks(q StripeWordBankListQuery) ([]StripeWordBankListItem, error) {
	list, err := a.store.ListStripeWordBanks(store.StripeWordBankListFilter{
		ConfigItem: q.ConfigItem,
	})
	if err != nil {
		return nil, err
	}
	out := make([]StripeWordBankListItem, 0, len(list))
	for _, item := range list {
		out = append(out, toStripeWordBankListItem(item))
	}
	return out, nil
}

// CreateStripeWordBank 新增 Stripe 单词。
func (a *App) CreateStripeWordBank(req CreateStripeWordBankRequest, operator string) (*StripeWordBankListItem, error) {
	configItem, err := normalizeStripeWordBankConfigItem(req.ConfigItem)
	if err != nil {
		return nil, err
	}
	name, err := normalizeStripeWordBankName(req.Name, configItem)
	if err != nil {
		return nil, err
	}
	if exist, err := a.store.FindStripeWordBankByName(name); err == nil && exist != nil {
		return nil, ErrStripeWordBankNameExists
	} else if err != nil && !isNotFound(err) {
		return nil, err
	}
	item := &model.StripeWordBank{
		Name:       name,
		UsageCount: 0,
		ConfigItem: configItem,
		CreatedBy:  operator,
		UpdatedBy:  operator,
	}
	if err := a.store.CreateStripeWordBank(item); err != nil {
		return nil, err
	}
	out := toStripeWordBankListItem(*item)
	return &out, nil
}

// UpdateStripeWordBank 编辑 Stripe 单词配置项。
func (a *App) UpdateStripeWordBank(id uint, req UpdateStripeWordBankRequest, operator string) (*StripeWordBankListItem, error) {
	item, err := a.store.GetStripeWordBankByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrStripeWordBankNotFound
		}
		return nil, err
	}
	configItem, err := normalizeStripeWordBankConfigItem(req.ConfigItem)
	if err != nil {
		return nil, err
	}
	if _, err := normalizeStripeWordBankName(item.Name, configItem); err != nil {
		return nil, err
	}
	item.ConfigItem = configItem
	item.UpdatedBy = operator
	if err := a.store.SaveStripeWordBank(item); err != nil {
		return nil, err
	}
	out := toStripeWordBankListItem(*item)
	return &out, nil
}

// DeleteStripeWordBank 删除 Stripe 单词。
func (a *App) DeleteStripeWordBank(id uint) error {
	item, err := a.store.GetStripeWordBankByID(id)
	if err != nil {
		if isNotFound(err) {
			return ErrStripeWordBankNotFound
		}
		return err
	}
	return a.store.DeleteStripeWordBank(item.ID)
}

func toStripeWordBankListItem(item model.StripeWordBank) StripeWordBankListItem {
	return StripeWordBankListItem{
		ID:         item.ID,
		Name:       item.Name,
		UsageCount: item.UsageCount,
		ConfigItem: item.ConfigItem,
	}
}

func normalizeStripeWordBankName(name, configItem string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ErrStripeWordBankNameInvalid
	}
	if configItem == model.StripeWordBankConfigDirectory {
		if strings.HasPrefix(name, "/") {
			return "", ErrStripeWordBankNameMustNotStartWithSlash
		}
	} else if !strings.HasPrefix(name, "/") {
		return "", ErrStripeWordBankNameMustStartWithSlash
	}
	if len(name) > 128 {
		return "", ErrStripeWordBankNameInvalid
	}
	return name, nil
}

func normalizeStripeWordBankConfigItem(configItem string) (string, error) {
	configItem = strings.TrimSpace(configItem)
	if _, ok := stripeWordBankConfigItems[configItem]; !ok {
		return "", ErrStripeWordBankConfigItemInvalid
	}
	return configItem, nil
}
