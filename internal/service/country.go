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
	ErrCountryNotFound       = errors.New("country not found")
	ErrCountryCodeExists     = errors.New("country code exists")
	ErrCountryCodeInvalid    = errors.New("country code invalid")
	ErrCountryCardBinInvalid = errors.New("country card bin ratio invalid")
)

//go:embed countries.json
var countriesJSON []byte

type countryConfig struct {
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	CardBinRatio float64 `json:"cardBinRatio"`
}

// CountryOption 国家下拉项，来自 countries.json。
type CountryOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

var (
	countryOnce     sync.Once
	countryLoadErr  error
	countryOptions  []CountryOption
	countryNames    map[string]string
	countrySeedRows []model.Country
)

func loadCountries() error {
	countryOnce.Do(func() {
		var configs []countryConfig
		if err := json.Unmarshal(countriesJSON, &configs); err != nil {
			countryLoadErr = fmt.Errorf("parse countries.json: %w", err)
			return
		}
		names := make(map[string]string, len(configs))
		opts := make([]CountryOption, 0, len(configs))
		seeds := make([]model.Country, 0, len(configs))
		for i, item := range configs {
			code := strings.ToUpper(strings.TrimSpace(item.Code))
			name := strings.TrimSpace(item.Name)
			if code == "" || name == "" || !isCountryCode(code) || !isValidCardBinRatio(item.CardBinRatio) {
				countryLoadErr = fmt.Errorf("countries.json: index %d 字段不完整", i)
				return
			}
			names[code] = name
			opts = append(opts, CountryOption{
				Value: code,
				Label: fmt.Sprintf("%s（%s）", name, code),
			})
			seeds = append(seeds, model.Country{
				Code:         code,
				Name:         name,
				CardBinRatio: item.CardBinRatio,
				CreatedBy:    "system",
				UpdatedBy:    "system",
			})
		}
		countryOptions = opts
		countryNames = names
		countrySeedRows = seeds
	})
	return countryLoadErr
}

func isCountryCode(code string) bool {
	if len(code) != 2 {
		return false
	}
	for _, r := range code {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func isValidCardBinRatio(ratio float64) bool {
	return !math.IsNaN(ratio) && !math.IsInf(ratio, 0) && ratio >= 0 && ratio <= 100
}

func countryNameByCode(code string) (string, bool) {
	_ = loadCountries()
	name, ok := countryNames[code]
	return name, ok
}

// ListCountryOptions 返回国家下拉，只取 JSON 里的 code/name。
func ListCountryOptions() ([]CountryOption, error) {
	if err := loadCountries(); err != nil {
		return nil, err
	}
	out := make([]CountryOption, len(countryOptions))
	copy(out, countryOptions)
	return out, nil
}

func countrySeedRecords() ([]model.Country, error) {
	if err := loadCountries(); err != nil {
		return nil, err
	}
	out := make([]model.Country, len(countrySeedRows))
	copy(out, countrySeedRows)
	return out, nil
}

// CountryListItem 国家列表行。
type CountryListItem struct {
	ID           uint    `json:"id"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	CardBinRatio float64 `json:"cardBinRatio"`
	CreatedBy    string  `json:"createdBy"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedBy    string  `json:"updatedBy"`
	UpdatedAt    string  `json:"updatedAt"`
}

// CountrySaveRequest 新增/编辑国家入参。
type CountrySaveRequest struct {
	Code         string  `json:"code"`
	CardBinRatio float64 `json:"cardBinRatio"`
}

// CountryListQuery 国家列表筛选。
type CountryListQuery struct {
	Field   string
	Keyword string
}

// ListCountries 查询国家列表。
func (a *App) ListCountries(q CountryListQuery) ([]CountryListItem, error) {
	list, err := a.store.ListCountries(store.CountryListFilter{
		Field:   strings.TrimSpace(q.Field),
		Keyword: strings.TrimSpace(q.Keyword),
	})
	if err != nil {
		return nil, err
	}
	out := make([]CountryListItem, 0, len(list))
	for _, item := range list {
		out = append(out, toCountryListItem(item))
	}
	return out, nil
}

// CreateCountry 新建国家。名称从 countries.json 带出。
func (a *App) CreateCountry(req CountrySaveRequest, operator string) (*CountryListItem, error) {
	code, name, ratio, err := normalizeCountry(req)
	if err != nil {
		return nil, err
	}
	if exist, err := a.store.FindCountryByCode(code); err == nil && exist != nil {
		return nil, ErrCountryCodeExists
	} else if err != nil && !isNotFound(err) {
		return nil, err
	}
	item := &model.Country{
		Code:         code,
		Name:         name,
		CardBinRatio: ratio,
		CreatedBy:    operator,
		UpdatedBy:    operator,
	}
	if err := a.store.CreateCountry(item); err != nil {
		return nil, err
	}
	return a.getCountryItem(item.ID)
}

// UpdateCountry 编辑大卡头占比，编码不允许改。
func (a *App) UpdateCountry(id uint, req CountrySaveRequest, operator string) (*CountryListItem, error) {
	item, err := a.store.GetCountryByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrCountryNotFound
		}
		return nil, err
	}
	_, _, ratio, err := normalizeCountry(CountrySaveRequest{
		Code:         item.Code,
		CardBinRatio: req.CardBinRatio,
	})
	if err != nil {
		return nil, err
	}
	item.CardBinRatio = ratio
	item.UpdatedBy = operator
	if err := a.store.SaveCountry(item); err != nil {
		return nil, err
	}
	return a.getCountryItem(item.ID)
}

// DeleteCountry 删除国家。
func (a *App) DeleteCountry(id uint) error {
	if _, err := a.store.GetCountryByID(id); err != nil {
		if isNotFound(err) {
			return ErrCountryNotFound
		}
		return err
	}
	return a.store.DeleteCountry(id)
}

func (a *App) getCountryItem(id uint) (*CountryListItem, error) {
	item, err := a.store.GetCountryByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrCountryNotFound
		}
		return nil, err
	}
	out := toCountryListItem(*item)
	return &out, nil
}

func normalizeCountry(req CountrySaveRequest) (string, string, float64, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if !isCountryCode(code) {
		return "", "", 0, ErrCountryCodeInvalid
	}
	if err := loadCountries(); err != nil {
		return "", "", 0, err
	}
	name, ok := countryNameByCode(code)
	if !ok {
		return "", "", 0, ErrCountryCodeInvalid
	}
	if !isValidCardBinRatio(req.CardBinRatio) {
		return "", "", 0, ErrCountryCardBinInvalid
	}
	return code, name, req.CardBinRatio, nil
}

func toCountryListItem(item model.Country) CountryListItem {
	name := item.Name
	if label, ok := countryNameByCode(item.Code); ok {
		name = label
	}
	return CountryListItem{
		ID:           item.ID,
		Code:         item.Code,
		Name:         name,
		CardBinRatio: item.CardBinRatio,
		CreatedBy:    item.CreatedBy,
		CreatedAt:    item.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedBy:    item.UpdatedBy,
		UpdatedAt:    item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
