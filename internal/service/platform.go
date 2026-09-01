package service

import "paymentcenter/internal/model"

// PlatformOption 平台下拉项。
type PlatformOption struct {
	ID    uint   `json:"id"`
	Code  string `json:"code"`
	Label string `json:"label"`
}

// ListPlatformOptions 返回平台下拉列表。
func (a *App) ListPlatformOptions() ([]PlatformOption, error) {
	list, err := a.store.ListPlatforms()
	if err != nil {
		return nil, err
	}
	out := make([]PlatformOption, 0, len(list))
	for _, item := range list {
		if item.Status != model.PlatformStatusEnabled {
			continue
		}
		label := item.Name
		if label == "" {
			label = item.Code
		}
		out = append(out, PlatformOption{
			ID:    item.ID,
			Code:  item.Code,
			Label: label,
		})
	}
	return out, nil
}

// loadPlatformMetaMap 加载平台元数据映射。
func (a *App) loadPlatformMetaMap() (map[uint]model.Platform, error) {
	list, err := a.store.ListPlatforms()
	if err != nil {
		return nil, err
	}
	result := make(map[uint]model.Platform, len(list))
	for _, item := range list {
		result[item.ID] = item
	}
	return result, nil
}

// getPlatformByID 根据 ID 获取平台。
func (a *App) getPlatformByID(id uint) (*model.Platform, error) {
	platform, err := a.store.GetPlatformByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrChannelPlatformInvalid
		}
		return nil, err
	}
	if platform.Status != model.PlatformStatusEnabled {
		return nil, ErrChannelPlatformInvalid
	}
	return platform, nil
}
