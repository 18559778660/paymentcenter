package service

import (
	"errors"
	"strings"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
)

var (
	ErrMerchantGroupNotFound        = errors.New("merchant group not found")
	ErrMerchantGroupNameExists      = errors.New("merchant group name exists")
	ErrMerchantGroupNameInvalid     = errors.New("merchant group name invalid")
	ErrMerchantGroupMerchantInvalid = errors.New("merchant group merchant invalid")
)

// MerchantGroupMerchant 分组内的商户简要信息。
type MerchantGroupMerchant struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Account string `json:"account"`
}

// MerchantGroupListItem 商户分组列表行。
type MerchantGroupListItem struct {
	ID        uint                    `json:"id"`
	Name      string                  `json:"name"`
	Merchants []MerchantGroupMerchant `json:"merchants"`
	CreatedBy string                  `json:"createdBy"`
	CreatedAt string                  `json:"createdAt"`
	UpdatedBy string                  `json:"updatedBy"`
	UpdatedAt string                  `json:"updatedAt"`
}

// MerchantGroupSaveRequest 新增/编辑分组入参。
type MerchantGroupSaveRequest struct {
	Name        string `json:"name"`
	MerchantIDs []uint `json:"merchantIds"`
}

// MerchantGroupListQuery 分组列表筛选。
type MerchantGroupListQuery struct {
	ID   *uint
	Name string
}

// ListMerchantGroups 查询分组列表。
func (a *App) ListMerchantGroups(q MerchantGroupListQuery) ([]MerchantGroupListItem, error) {
	groups, err := a.store.ListMerchantGroups(store.MerchantGroupListFilter{
		ID:   q.ID,
		Name: strings.TrimSpace(q.Name),
	})
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return []MerchantGroupListItem{}, nil
	}
	groupIDs := make([]uint, 0, len(groups))
	for _, g := range groups {
		groupIDs = append(groupIDs, g.ID)
	}
	members, err := a.store.ListMerchantGroupMembers(groupIDs)
	if err != nil {
		return nil, err
	}
	merchantIDs := make([]uint, 0, len(members))
	seen := map[uint]struct{}{}
	membersByGroup := map[uint][]uint{}
	for _, m := range members {
		membersByGroup[m.GroupID] = append(membersByGroup[m.GroupID], m.MerchantID)
		if _, ok := seen[m.MerchantID]; ok {
			continue
		}
		seen[m.MerchantID] = struct{}{}
		merchantIDs = append(merchantIDs, m.MerchantID)
	}
	merchantByID := map[uint]model.Merchant{}
	if len(merchantIDs) > 0 {
		list, err := a.store.GetMerchantsByIDs(merchantIDs)
		if err != nil {
			return nil, err
		}
		for _, m := range list {
			merchantByID[m.ID] = m
		}
	}
	out := make([]MerchantGroupListItem, 0, len(groups))
	for _, g := range groups {
		out = append(out, toMerchantGroupListItem(g, membersByGroup[g.ID], merchantByID))
	}
	return out, nil
}

// CreateMerchantGroup 新建分组并绑定商户。
func (a *App) CreateMerchantGroup(req MerchantGroupSaveRequest, operator string) (*MerchantGroupListItem, error) {
	name, err := normalizeMerchantGroupName(req.Name)
	if err != nil {
		return nil, err
	}
	if exist, err := a.store.FindMerchantGroupByName(name); err == nil && exist != nil {
		return nil, ErrMerchantGroupNameExists
	} else if err != nil && !isNotFound(err) {
		return nil, err
	}
	merchantIDs, err := a.normalizeGroupMerchantIDs(req.MerchantIDs)
	if err != nil {
		return nil, err
	}
	group := &model.MerchantGroup{
		Name:      name,
		CreatedBy: operator,
		UpdatedBy: operator,
	}
	if err := a.store.CreateMerchantGroup(group); err != nil {
		return nil, err
	}
	if err := a.store.ReplaceMerchantGroupMembers(group.ID, merchantIDs); err != nil {
		return nil, err
	}
	return a.getMerchantGroupItem(group.ID)
}

// UpdateMerchantGroup 更新分组名和商户列表。
func (a *App) UpdateMerchantGroup(id uint, req MerchantGroupSaveRequest, operator string) (*MerchantGroupListItem, error) {
	group, err := a.store.GetMerchantGroupByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrMerchantGroupNotFound
		}
		return nil, err
	}
	name, err := normalizeMerchantGroupName(req.Name)
	if err != nil {
		return nil, err
	}
	if exist, err := a.store.FindMerchantGroupByName(name); err == nil && exist.ID != group.ID {
		return nil, ErrMerchantGroupNameExists
	} else if err != nil && !isNotFound(err) {
		return nil, err
	}
	merchantIDs, err := a.normalizeGroupMerchantIDs(req.MerchantIDs)
	if err != nil {
		return nil, err
	}
	group.Name = name
	group.UpdatedBy = operator
	if err := a.store.SaveMerchantGroup(group); err != nil {
		return nil, err
	}
	if err := a.store.ReplaceMerchantGroupMembers(group.ID, merchantIDs); err != nil {
		return nil, err
	}
	return a.getMerchantGroupItem(group.ID)
}

// DeleteMerchantGroup 删除分组及其成员关系。
func (a *App) DeleteMerchantGroup(id uint) error {
	if _, err := a.store.GetMerchantGroupByID(id); err != nil {
		if isNotFound(err) {
			return ErrMerchantGroupNotFound
		}
		return err
	}
	return a.store.DeleteMerchantGroup(id)
}

// getMerchantGroupItem 获取商户分组列表项。
func (a *App) getMerchantGroupItem(id uint) (*MerchantGroupListItem, error) {
	list, err := a.ListMerchantGroups(MerchantGroupListQuery{ID: &id})
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, ErrMerchantGroupNotFound
	}
	item := list[0]
	return &item, nil
}

// normalizeGroupMerchantIDs 规范化商户分组成员 ID 列表。
func (a *App) normalizeGroupMerchantIDs(ids []uint) ([]uint, error) {
	uniq := make([]uint, 0, len(ids))
	seen := map[uint]struct{}{}
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}
	if len(uniq) == 0 {
		return []uint{}, nil
	}
	list, err := a.store.GetMerchantsByIDs(uniq)
	if err != nil {
		return nil, err
	}
	if len(list) != len(uniq) {
		return nil, ErrMerchantGroupMerchantInvalid
	}
	return uniq, nil
}

// normalizeMerchantGroupName 规范化商户分组名。
func normalizeMerchantGroupName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" || len([]rune(name)) > 64 {
		return "", ErrMerchantGroupNameInvalid
	}
	return name, nil
}

// toMerchantGroupListItem 转换为商户分组列表项。
func toMerchantGroupListItem(g model.MerchantGroup, merchantIDs []uint, merchantByID map[uint]model.Merchant) MerchantGroupListItem {
	merchants := make([]MerchantGroupMerchant, 0, len(merchantIDs))
	for _, id := range merchantIDs {
		m, ok := merchantByID[id]
		if !ok {
			continue
		}
		merchants = append(merchants, MerchantGroupMerchant{
			ID:      m.ID,
			Name:    m.Name,
			Account: m.Account,
		})
	}
	return MerchantGroupListItem{
		ID:        g.ID,
		Name:      g.Name,
		Merchants: merchants,
		CreatedBy: g.CreatedBy,
		CreatedAt: g.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedBy: g.UpdatedBy,
		UpdatedAt: g.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
