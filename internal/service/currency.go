package service

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"unicode"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
)

var (
	ErrCurrencyNotFound    = errors.New("currency not found")
	ErrCurrencyCodeExists  = errors.New("currency code exists")
	ErrCurrencyCodeInvalid = errors.New("currency code invalid")
	ErrCurrencyRateInvalid = errors.New("currency rate invalid")
)

//go:embed currencies.json
var currenciesJSON []byte

type currencyConfig struct {
	Code string  `json:"code"`
	Name string  `json:"name"`
	Rate float64 `json:"rate"`
}

// CurrencyOption 货币下拉项，来自 currencies.json。
type CurrencyOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

var (
	currencyOnce     sync.Once
	currencyLoadErr  error
	currencyOptions  []CurrencyOption
	currencyNames    map[string]string
	currencySeedRows []model.Currency
)

func loadCurrencies() error {
	currencyOnce.Do(func() {
		var configs []currencyConfig
		if err := json.Unmarshal(currenciesJSON, &configs); err != nil {
			currencyLoadErr = fmt.Errorf("parse currencies.json: %w", err)
			return
		}
		names := make(map[string]string, len(configs))
		opts := make([]CurrencyOption, 0, len(configs))
		seeds := make([]model.Currency, 0, len(configs))
		for i, item := range configs {
			code := strings.ToUpper(strings.TrimSpace(item.Code))
			name := strings.TrimSpace(item.Name)
			if code == "" || name == "" || !isCurrencyCode(code) || !isValidRate(item.Rate) {
				currencyLoadErr = fmt.Errorf("currencies.json: index %d 字段不完整", i)
				return
			}
			names[code] = name
			opts = append(opts, CurrencyOption{
				Value: code,
				Label: fmt.Sprintf("%s（%s）", name, code),
			})
			seeds = append(seeds, model.Currency{
				Code:      code,
				Name:      name,
				Rate:      item.Rate,
				CreatedBy: "system",
				UpdatedBy: "system",
			})
		}
		currencyOptions = opts
		currencyNames = names
		currencySeedRows = seeds
	})
	return currencyLoadErr
}

func isCurrencyCode(code string) bool {
	if len(code) < 3 || len(code) > 16 {
		return false
	}
	for _, r := range code {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		return false
	}
	return true
}

func isValidRate(rate float64) bool {
	return !math.IsNaN(rate) && !math.IsInf(rate, 0) && rate >= 0
}

func currencyNameByCode(code string) (string, bool) {
	_ = loadCurrencies()
	name, ok := currencyNames[code]
	return name, ok
}

// ListCurrencyOptions 返回货币下拉，只取 JSON 里的 code/name。
func ListCurrencyOptions() ([]CurrencyOption, error) {
	if err := loadCurrencies(); err != nil {
		return nil, err
	}
	out := make([]CurrencyOption, len(currencyOptions))
	copy(out, currencyOptions)
	return out, nil
}

func currencySeedRecords() ([]model.Currency, error) {
	if err := loadCurrencies(); err != nil {
		return nil, err
	}
	out := make([]model.Currency, len(currencySeedRows))
	copy(out, currencySeedRows)
	return out, nil
}

// CurrencyListItem 货币列表行。
type CurrencyListItem struct {
	ID        uint    `json:"id"`
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Rate      float64 `json:"rate"`
	CreatedBy string  `json:"createdBy"`
	CreatedAt string  `json:"createdAt"`
	UpdatedBy string  `json:"updatedBy"`
	UpdatedAt string  `json:"updatedAt"`
}

// CurrencySaveRequest 新增/编辑货币入参。
type CurrencySaveRequest struct {
	Code string  `json:"code"`
	Rate float64 `json:"rate"`
}

// CurrencyListQuery 货币列表筛选。
type CurrencyListQuery struct {
	Field   string
	Keyword string
}

// ListCurrencies 查询货币列表。
func (a *App) ListCurrencies(q CurrencyListQuery) ([]CurrencyListItem, error) {
	list, err := a.store.ListCurrencies(store.CurrencyListFilter{
		Field:   strings.TrimSpace(q.Field),
		Keyword: strings.TrimSpace(q.Keyword),
	})
	if err != nil {
		return nil, err
	}
	out := make([]CurrencyListItem, 0, len(list))
	for _, item := range list {
		out = append(out, toCurrencyListItem(item))
	}
	return out, nil
}

// CreateCurrency 新建货币。名称从 currencies.json 带出。
func (a *App) CreateCurrency(req CurrencySaveRequest, operator string) (*CurrencyListItem, error) {
	code, name, rate, err := normalizeCurrency(req)
	if err != nil {
		return nil, err
	}
	if exist, err := a.store.FindCurrencyByCode(code); err == nil && exist != nil {
		return nil, ErrCurrencyCodeExists
	} else if err != nil && !isNotFound(err) {
		return nil, err
	}
	item := &model.Currency{
		Code:      code,
		Name:      name,
		Rate:      rate,
		CreatedBy: operator,
		UpdatedBy: operator,
	}
	if err := a.store.CreateCurrency(item); err != nil {
		return nil, err
	}
	return a.getCurrencyItem(item.ID)
}

// UpdateCurrency 编辑货币汇率，编码不允许改。
func (a *App) UpdateCurrency(id uint, req CurrencySaveRequest, operator string) (*CurrencyListItem, error) {
	item, err := a.store.GetCurrencyByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrCurrencyNotFound
		}
		return nil, err
	}
	_, _, rate, err := normalizeCurrency(CurrencySaveRequest{
		Code: item.Code,
		Rate: req.Rate,
	})
	if err != nil {
		return nil, err
	}
	item.Rate = rate
	item.UpdatedBy = operator
	if err := a.store.SaveCurrency(item); err != nil {
		return nil, err
	}
	return a.getCurrencyItem(item.ID)
}

// DeleteCurrency 删除货币。
func (a *App) DeleteCurrency(id uint) error {
	if _, err := a.store.GetCurrencyByID(id); err != nil {
		if isNotFound(err) {
			return ErrCurrencyNotFound
		}
		return err
	}
	return a.store.DeleteCurrency(id)
}

func (a *App) getCurrencyItem(id uint) (*CurrencyListItem, error) {
	item, err := a.store.GetCurrencyByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrCurrencyNotFound
		}
		return nil, err
	}
	out := toCurrencyListItem(*item)
	return &out, nil
}

func normalizeCurrency(req CurrencySaveRequest) (string, string, float64, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if !isCurrencyCode(code) {
		return "", "", 0, ErrCurrencyCodeInvalid
	}
	if err := loadCurrencies(); err != nil {
		return "", "", 0, err
	}
	name, ok := currencyNameByCode(code)
	if !ok {
		return "", "", 0, ErrCurrencyCodeInvalid
	}
	if !isValidRate(req.Rate) {
		return "", "", 0, ErrCurrencyRateInvalid
	}
	return code, name, req.Rate, nil
}

func toCurrencyListItem(item model.Currency) CurrencyListItem {
	name := item.Name
	if label, ok := currencyNameByCode(item.Code); ok {
		name = label
	}
	return CurrencyListItem{
		ID:        item.ID,
		Code:      item.Code,
		Name:      name,
		Rate:      item.Rate,
		CreatedBy: item.CreatedBy,
		CreatedAt: item.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedBy: item.UpdatedBy,
		UpdatedAt: item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
