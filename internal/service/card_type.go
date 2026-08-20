package service

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"unicode"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
)

var (
	ErrCardTypeNotFound      = errors.New("card type not found")
	ErrCardTypeCodeExists    = errors.New("card type code exists")
	ErrCardTypeCodeInvalid   = errors.New("card type code invalid")
	ErrCardTypeNameInvalid   = errors.New("card type name invalid")
	ErrCardTypeLengthInvalid = errors.New("card type length invalid")
	ErrCardTypePrefixInvalid = errors.New("card type prefix invalid")
)

//go:embed card_brands.json
var cardBrandsJSON []byte

// cardTypeConfig 卡品牌配置，每条字段必须齐全。
// 下拉只取 value/label；空库种子插入全部条目。
type cardTypeConfig struct {
	Value    string                `json:"value"`
	Label    string                `json:"label"`
	Code     string                `json:"code"`
	Lengths  []int                 `json:"lengths"`
	Prefixes []model.CardBinPrefix `json:"prefixes"`
}

// CardBrandOption 卡名称下拉项，只暴露名称。
type CardBrandOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

var (
	cardBrandOnce    sync.Once
	cardBrandLoadErr error
	cardBrandOptions []CardBrandOption
	cardBrandLabels  map[string]string
	cardTypeSeeds    []model.CardType
)

// loadCardBrands 加载卡品牌配置。
func loadCardBrands() error {
	cardBrandOnce.Do(func() {
		var configs []cardTypeConfig
		if err := json.Unmarshal(cardBrandsJSON, &configs); err != nil {
			cardBrandLoadErr = fmt.Errorf("parse card_brands.json: %w", err)
			return
		}
		labels := make(map[string]string, len(configs))
		opts := make([]CardBrandOption, 0, len(configs))
		seeds := make([]model.CardType, 0)
		for i, item := range configs {
			value := strings.TrimSpace(item.Value)
			label := strings.TrimSpace(item.Label)
			code := strings.TrimSpace(item.Code)
			if value == "" || label == "" || code == "" || len(item.Lengths) == 0 || len(item.Prefixes) == 0 {
				cardBrandLoadErr = fmt.Errorf("card_brands.json: index %d 字段不完整", i)
				return
			}
			labels[value] = label
			opts = append(opts, CardBrandOption{Value: value, Label: label})
			seeds = append(seeds, model.CardType{
				Code:      code,
				Name:      value,
				Lengths:   model.CardBinLengths(item.Lengths),
				Prefixes:  model.CardBinPrefixes(item.Prefixes),
				CreatedBy: "system",
				UpdatedBy: "system",
			})
		}
		cardBrandOptions = opts
		cardBrandLabels = labels
		cardTypeSeeds = seeds
	})
	return cardBrandLoadErr
}

// cardTypeSeedRecords 空库首次启动要插入的默认卡类型，来自同一份 JSON。
func cardTypeSeedRecords() ([]model.CardType, error) {
	if err := loadCardBrands(); err != nil {
		return nil, err
	}
	out := make([]model.CardType, len(cardTypeSeeds))
	copy(out, cardTypeSeeds)
	return out, nil
}

func cardBrandLabel(code string) string {
	_ = loadCardBrands()
	if label, ok := cardBrandLabels[code]; ok {
		return label
	}
	return code
}

// ListCardBrands 返回卡名称下拉，只取 JSON 里的 value/label。
func ListCardBrands() ([]CardBrandOption, error) {
	if err := loadCardBrands(); err != nil {
		return nil, err
	}
	out := make([]CardBrandOption, len(cardBrandOptions))
	copy(out, cardBrandOptions)
	return out, nil
}

var allowedCardLengths = map[int]struct{}{
	13: {}, 14: {}, 15: {}, 16: {}, 17: {}, 18: {}, 19: {},
}

// CardTypeListItem 卡类型列表行。
type CardTypeListItem struct {
	ID        uint                  `json:"id"`
	Code      string                `json:"code"`
	Name      string                `json:"name"`
	NameLabel string                `json:"nameLabel"`
	Lengths   []int                 `json:"lengths"`
	Prefixes  []model.CardBinPrefix `json:"prefixes"`
	CreatedBy string                `json:"createdBy"`
	CreatedAt string                `json:"createdAt"`
	UpdatedBy string                `json:"updatedBy"`
	UpdatedAt string                `json:"updatedAt"`
}

// CardTypeSaveRequest 新增/编辑卡类型入参。
type CardTypeSaveRequest struct {
	Code     string                `json:"code"`
	Name     string                `json:"name"`
	Lengths  []int                 `json:"lengths"`
	Prefixes []model.CardBinPrefix `json:"prefixes"`
}

// CardTypeListQuery 卡类型列表筛选。
type CardTypeListQuery struct {
	Field   string
	Keyword string
}

// ListCardTypes 查询卡类型列表。
func (a *App) ListCardTypes(q CardTypeListQuery) ([]CardTypeListItem, error) {
	keyword := strings.TrimSpace(q.Keyword)
	filter := store.CardTypeListFilter{
		Field:   strings.TrimSpace(q.Field),
		Keyword: keyword,
	}
	if keyword != "" && filter.Field != "code" {
		filter.Names = matchCardBrandCodes(keyword)
	}
	list, err := a.store.ListCardTypes(filter)
	if err != nil {
		return nil, err
	}
	out := make([]CardTypeListItem, 0, len(list))
	for _, item := range list {
		out = append(out, toCardTypeListItem(item))
	}
	return out, nil
}

// CreateCardType 新建卡类型。
func (a *App) CreateCardType(req CardTypeSaveRequest, operator string) (*CardTypeListItem, error) {
	code, name, lengths, prefixes, err := normalizeCardType(req)
	if err != nil {
		return nil, err
	}
	if exist, err := a.store.FindCardTypeByCode(code); err == nil && exist != nil {
		return nil, ErrCardTypeCodeExists
	} else if err != nil && !isNotFound(err) {
		return nil, err
	}
	item := &model.CardType{
		Code:      code,
		Name:      name,
		Lengths:   lengths,
		Prefixes:  prefixes,
		CreatedBy: operator,
		UpdatedBy: operator,
	}
	if err := a.store.CreateCardType(item); err != nil {
		return nil, err
	}
	return a.getCardTypeItem(item.ID)
}

// UpdateCardType 编辑卡类型。
func (a *App) UpdateCardType(id uint, req CardTypeSaveRequest, operator string) (*CardTypeListItem, error) {
	item, err := a.store.GetCardTypeByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrCardTypeNotFound
		}
		return nil, err
	}
	code, name, lengths, prefixes, err := normalizeCardType(req)
	if err != nil {
		return nil, err
	}
	if exist, err := a.store.FindCardTypeByCode(code); err == nil && exist.ID != item.ID {
		return nil, ErrCardTypeCodeExists
	} else if err != nil && !isNotFound(err) {
		return nil, err
	}
	item.Code = code
	item.Name = name
	item.Lengths = lengths
	item.Prefixes = prefixes
	item.UpdatedBy = operator
	if err := a.store.SaveCardType(item); err != nil {
		return nil, err
	}
	return a.getCardTypeItem(item.ID)
}

func (a *App) getCardTypeItem(id uint) (*CardTypeListItem, error) {
	item, err := a.store.GetCardTypeByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrCardTypeNotFound
		}
		return nil, err
	}
	out := toCardTypeListItem(*item)
	return &out, nil
}

func normalizeCardType(req CardTypeSaveRequest) (string, string, model.CardBinLengths, model.CardBinPrefixes, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if code == "" || len([]rune(code)) > 32 {
		return "", "", nil, nil, ErrCardTypeCodeInvalid
	}
	for _, r := range code {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		return "", "", nil, nil, ErrCardTypeCodeInvalid
	}
	name := strings.TrimSpace(req.Name)
	if err := loadCardBrands(); err != nil {
		return "", "", nil, nil, err
	}
	if _, ok := cardBrandLabels[name]; !ok {
		return "", "", nil, nil, ErrCardTypeNameInvalid
	}
	lengths, err := normalizeCardLengths(req.Lengths)
	if err != nil {
		return "", "", nil, nil, err
	}
	prefixes, err := normalizeCardPrefixes(req.Prefixes)
	if err != nil {
		return "", "", nil, nil, err
	}
	return code, name, lengths, prefixes, nil
}

func normalizeCardLengths(raw []int) (model.CardBinLengths, error) {
	seen := map[int]struct{}{}
	out := make(model.CardBinLengths, 0, len(raw))
	for _, n := range raw {
		if _, ok := allowedCardLengths[n]; !ok {
			return nil, ErrCardTypeLengthInvalid
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	sort.Ints(out)
	return out, nil
}

func normalizeCardPrefixes(raw []model.CardBinPrefix) (model.CardBinPrefixes, error) {
	out := make(model.CardBinPrefixes, 0, len(raw))
	for _, item := range raw {
		start := strings.TrimSpace(item.Start)
		end := strings.TrimSpace(item.End)
		if start == "" && end == "" {
			continue
		}
		if !isDigits(start) {
			return nil, ErrCardTypePrefixInvalid
		}
		if end != "" {
			if !isDigits(end) || len(end) != len(start) {
				return nil, ErrCardTypePrefixInvalid
			}
			if end < start {
				return nil, ErrCardTypePrefixInvalid
			}
		}
		out = append(out, model.CardBinPrefix{Start: start, End: end})
	}
	if len(out) == 0 {
		return nil, ErrCardTypePrefixInvalid
	}
	return out, nil
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func matchCardBrandCodes(keyword string) []string {
	_ = loadCardBrands()
	kw := strings.ToLower(keyword)
	out := make([]string, 0)
	for code, label := range cardBrandLabels {
		if strings.Contains(strings.ToLower(label), kw) || strings.Contains(code, kw) {
			out = append(out, code)
		}
	}
	return out
}

func toCardTypeListItem(item model.CardType) CardTypeListItem {
	lengths := []int(item.Lengths)
	if lengths == nil {
		lengths = []int{}
	}
	prefixes := []model.CardBinPrefix(item.Prefixes)
	if prefixes == nil {
		prefixes = []model.CardBinPrefix{}
	}
	return CardTypeListItem{
		ID:        item.ID,
		Code:      item.Code,
		Name:      item.Name,
		NameLabel: cardBrandLabel(item.Name),
		Lengths:   lengths,
		Prefixes:  prefixes,
		CreatedBy: item.CreatedBy,
		CreatedAt: item.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedBy: item.UpdatedBy,
		UpdatedAt: item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
