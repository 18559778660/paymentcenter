package service

import (
	"errors"
	"strings"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
)

var (
	ErrSiteANotFound         = errors.New("site a not found")
	ErrSiteADomainExists     = errors.New("site a domain exists")
	ErrSiteADomainInvalid    = errors.New("site a domain invalid")
	ErrSiteAFrameworkInvalid = errors.New("site a framework invalid")
	ErrSiteAStatusInvalid    = errors.New("site a status invalid")
	ErrSiteAMerchantInvalid  = errors.New("site a merchant invalid")
	ErrSiteABatchEmpty       = errors.New("site a batch empty")
)

// SiteAListItem A 站列表行。
type SiteAListItem struct {
	ID           uint   `json:"id"`
	MerchantID   uint   `json:"merchantId"`
	MerchantName string `json:"merchantName"`
	Domain       string `json:"domain"`
	Framework    string `json:"framework"`
	Status       string `json:"status"`
	CreatedBy    string `json:"createdBy"`
	CreatedAt    string `json:"createdAt"`
	UpdatedBy    string `json:"updatedBy"`
	UpdatedAt    string `json:"updatedAt"`
}

// SiteAListQuery 列表筛选。
type SiteAListQuery struct {
	MerchantID *uint
	Domain     string
	Status     string
}

// CreateSiteARequest 新增 A 站。
type CreateSiteARequest struct {
	MerchantID uint   `json:"merchantId" binding:"required"`
	Domain     string `json:"domain" binding:"required"`
	Framework  string `json:"framework" binding:"required"`
}

// BatchUpdateSiteAStatusRequest 批量更新状态。
type BatchUpdateSiteAStatusRequest struct {
	IDs    []uint `json:"ids" binding:"required"`
	Status string `json:"status" binding:"required"`
}

// ListSiteAs 查询 A 站列表。
func (a *App) ListSiteAs(q SiteAListQuery) ([]SiteAListItem, error) {
	list, err := a.store.ListSiteAs(store.SiteAListFilter{
		MerchantID: q.MerchantID,
		Domain:     q.Domain,
		Status:     q.Status,
	})
	if err != nil {
		return nil, err
	}
	merchantNames, err := a.loadMerchantNameMap(list)
	if err != nil {
		return nil, err
	}
	out := make([]SiteAListItem, 0, len(list))
	for _, item := range list {
		out = append(out, toSiteAListItem(item, merchantNames))
	}
	return out, nil
}

// CreateSiteA 新增 A 站。
func (a *App) CreateSiteA(req CreateSiteARequest, operator string) (*SiteAListItem, error) {
	domain, err := normalizeSiteADomain(req.Domain)
	if err != nil {
		return nil, err
	}
	framework, err := normalizeSiteAFramework(req.Framework)
	if err != nil {
		return nil, err
	}
	merchant, err := a.store.GetMerchantByID(req.MerchantID)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrSiteAMerchantInvalid
		}
		return nil, err
	}
	if exist, err := a.store.FindSiteAByDomain(domain); err == nil && exist != nil {
		return nil, ErrSiteADomainExists
	} else if err != nil && !isNotFound(err) {
		return nil, err
	}
	status := model.SiteAStatusPending
	if merchant.AuditSiteA == "auto" {
		status = model.SiteAStatusAudited
	}
	item := &model.SiteA{
		MerchantID: merchant.ID,
		Domain:     domain,
		Framework:  framework,
		Status:     status,
		CreatedBy:  operator,
		UpdatedBy:  operator,
	}
	if err := a.store.CreateSiteA(item); err != nil {
		return nil, err
	}
	return a.getSiteAItem(item.ID)
}

// BatchUpdateSiteAStatus 批量更新 A 站状态。
func (a *App) BatchUpdateSiteAStatus(req BatchUpdateSiteAStatusRequest, operator string) (int, error) {
	if len(req.IDs) == 0 {
		return 0, ErrSiteABatchEmpty
	}
	status, err := normalizeSiteAStatus(req.Status)
	if err != nil {
		return 0, err
	}
	if err := a.store.BatchUpdateSiteAStatus(req.IDs, status, operator); err != nil {
		return 0, err
	}
	return len(req.IDs), nil
}

func (a *App) getSiteAItem(id uint) (*SiteAListItem, error) {
	item, err := a.store.GetSiteAByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrSiteANotFound
		}
		return nil, err
	}
	merchantNames, err := a.loadMerchantNameMap([]model.SiteA{*item})
	if err != nil {
		return nil, err
	}
	out := toSiteAListItem(*item, merchantNames)
	return &out, nil
}

func (a *App) loadMerchantNameMap(list []model.SiteA) (map[uint]string, error) {
	ids := make([]uint, 0, len(list))
	seen := map[uint]struct{}{}
	for _, item := range list {
		if item.MerchantID == 0 {
			continue
		}
		if _, ok := seen[item.MerchantID]; ok {
			continue
		}
		seen[item.MerchantID] = struct{}{}
		ids = append(ids, item.MerchantID)
	}
	if len(ids) == 0 {
		return map[uint]string{}, nil
	}
	merchants, err := a.store.GetMerchantsByIDs(ids)
	if err != nil {
		return nil, err
	}
	names := make(map[uint]string, len(merchants))
	for _, merchant := range merchants {
		names[merchant.ID] = merchant.Name
	}
	return names, nil
}

func toSiteAListItem(item model.SiteA, merchantNames map[uint]string) SiteAListItem {
	return SiteAListItem{
		ID:           item.ID,
		MerchantID:   item.MerchantID,
		MerchantName: merchantNames[item.MerchantID],
		Domain:       item.Domain,
		Framework:    item.Framework,
		Status:       item.Status,
		CreatedBy:    item.CreatedBy,
		CreatedAt:    item.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedBy:    item.UpdatedBy,
		UpdatedAt:    item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// normalizeSiteADomain 标准化 A 站域名。
func normalizeSiteADomain(domain string) (string, error) {
	domain = strings.TrimSpace(strings.ToLower(domain))
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimSuffix(domain, "/")
	if domain == "" {
		return "", ErrSiteADomainInvalid
	}
	if len(domain) > 191 {
		return "", ErrSiteADomainInvalid
	}
	return domain, nil
}

func normalizeSiteAFramework(framework string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(framework)) {
	case model.SiteAFrameworkWooCommerce:
		return model.SiteAFrameworkWooCommerce, nil
	case model.SiteAFrameworkShopyy:
		return model.SiteAFrameworkShopyy, nil
	default:
		return "", ErrSiteAFrameworkInvalid
	}
}

func normalizeSiteAStatus(status string) (string, error) {
	switch strings.TrimSpace(status) {
	case model.SiteAStatusPending, model.SiteAStatusAudited, model.SiteAStatusDisabled:
		return strings.TrimSpace(status), nil
	default:
		return "", ErrSiteAStatusInvalid
	}
}
