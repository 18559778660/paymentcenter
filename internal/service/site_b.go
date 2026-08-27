package service

import (
	"errors"
	"strings"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
)

var (
	ErrSiteBNotFound         = errors.New("site b not found")
	ErrSiteBDomainExists     = errors.New("site b domain exists")
	ErrSiteBDomainInvalid    = errors.New("site b domain invalid")
	ErrSiteBPlatformInvalid  = errors.New("site b platform invalid")
	ErrSiteBFrameworkInvalid = errors.New("site b framework invalid")
)

// SiteBListItem B 站列表行。
type SiteBListItem struct {
	ID             uint   `json:"id"`
	Domain         string `json:"domain"`
	Channel        string `json:"channel"`
	ChannelEnabled bool   `json:"channelEnabled"`
	Platform       string `json:"platform"`
	Framework      string `json:"framework"`
	Status         bool   `json:"status"`
	IsFtp          bool   `json:"isFtp"`
	Host           string `json:"host"`
	Account        string `json:"account"`
	Password       string `json:"password"`
	LinkAddress    string `json:"linkAddress"`
	RunDirectory   string `json:"runDirectory"`
	Remark         string `json:"remark"`
	CreatedBy      string `json:"createdBy"`
	CreatedAt      string `json:"createdAt"`
	UpdatedBy      string `json:"updatedBy"`
	UpdatedAt      string `json:"updatedAt"`
}

// SiteBListQuery 列表筛选。
type SiteBListQuery struct {
	ID       *uint
	Domain   string
	Remark   string
	Status   *bool
	Platform string
}

// CreateSiteBRequest 新增 B 站。
type CreateSiteBRequest struct {
	Domain    string `json:"domain" binding:"required"`
	Platform  string `json:"platform"`
	Framework string `json:"framework"`
	IsFtp     *bool  `json:"isFtp"`
	Host      string `json:"host"`
	Account   string `json:"account"`
	Password  string `json:"password"`
}

// UpdateSiteBRequest 编辑 B 站 FTP 信息。
type UpdateSiteBRequest struct {
	Host     string `json:"host"`
	Account  string `json:"account"`
	Password string `json:"password"`
}

// SetSiteBStatusRequest 更新 B 站状态。
type SetSiteBStatusRequest struct {
	Status bool `json:"status"`
}

// ListSiteBs 查询 B 站列表。
func (a *App) ListSiteBs(q SiteBListQuery) ([]SiteBListItem, error) {
	list, err := a.store.ListSiteBs(store.SiteBListFilter{
		ID:       q.ID,
		Domain:   q.Domain,
		Remark:   q.Remark,
		Status:   q.Status,
		Platform: q.Platform,
	})
	if err != nil {
		return nil, err
	}
	channelStatus, err := a.loadChannelStatusMap()
	if err != nil {
		return nil, err
	}
	out := make([]SiteBListItem, 0, len(list))
	for _, item := range list {
		out = append(out, toSiteBListItem(item, channelStatus))
	}
	return out, nil
}

// CreateSiteB 新增 B 站。
func (a *App) CreateSiteB(req CreateSiteBRequest, operator string) (*SiteBListItem, error) {
	domain, err := normalizeSiteBDomain(req.Domain)
	if err != nil {
		return nil, err
	}
	platform, channelEnabled, err := a.normalizeSiteBPlatform(req.Platform)
	if err != nil {
		return nil, err
	}
	framework, err := normalizeSiteBFramework(req.Framework)
	if err != nil {
		return nil, err
	}
	if exist, err := a.store.FindSiteBByDomain(domain); err == nil && exist != nil {
		return nil, ErrSiteBDomainExists
	} else if err != nil && !isNotFound(err) {
		return nil, err
	}
	isFtp := true
	if req.IsFtp != nil {
		isFtp = *req.IsFtp
	}
	item := &model.SiteB{
		Domain:         domain,
		Platform:       platform,
		Framework:      framework,
		Status:         true,
		ChannelEnabled: channelEnabled,
		IsFtp:          isFtp,
		Host:           strings.TrimSpace(req.Host),
		Account:        strings.TrimSpace(req.Account),
		Password:       req.Password,
		RunDirectory:   model.SiteBDefaultRunDirectory,
		CreatedBy:      operator,
		UpdatedBy:      operator,
	}
	if err := a.store.CreateSiteB(item); err != nil {
		return nil, err
	}
	return a.getSiteBItem(item.ID)
}

// UpdateSiteB 编辑 B 站 FTP 信息。
func (a *App) UpdateSiteB(id uint, req UpdateSiteBRequest, operator string) (*SiteBListItem, error) {
	item, err := a.store.GetSiteBByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrSiteBNotFound
		}
		return nil, err
	}
	item.Host = strings.TrimSpace(req.Host)
	item.Account = strings.TrimSpace(req.Account)
	item.Password = req.Password
	item.UpdatedBy = operator
	if err := a.store.SaveSiteB(item); err != nil {
		return nil, err
	}
	return a.getSiteBItem(item.ID)
}

// SetSiteBStatus 更新 B 站启用状态。
func (a *App) SetSiteBStatus(id uint, req SetSiteBStatusRequest, operator string) (*SiteBListItem, error) {
	item, err := a.store.GetSiteBByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrSiteBNotFound
		}
		return nil, err
	}
	item.Status = req.Status
	item.UpdatedBy = operator
	if err := a.store.SaveSiteB(item); err != nil {
		return nil, err
	}
	return a.getSiteBItem(item.ID)
}

func (a *App) loadChannelStatusMap() (map[string]bool, error) {
	channels, err := a.store.ListChannels(store.ChannelListFilter{})
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool, len(channels))
	for _, channel := range channels {
		result[channel.Name] = channel.Status == model.ChannelStatusEnabled
	}
	return result, nil
}

func (a *App) getSiteBItem(id uint) (*SiteBListItem, error) {
	item, err := a.store.GetSiteBByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrSiteBNotFound
		}
		return nil, err
	}
	channelStatus, err := a.loadChannelStatusMap()
	if err != nil {
		return nil, err
	}
	out := toSiteBListItem(*item, channelStatus)
	return &out, nil
}

func toSiteBListItem(item model.SiteB, channelStatus map[string]bool) SiteBListItem {
	channelEnabled := item.ChannelEnabled
	if enabled, ok := channelStatus[item.Platform]; ok {
		channelEnabled = enabled
	}
	return SiteBListItem{
		ID:             item.ID,
		Domain:         item.Domain,
		Channel:        item.Platform,
		ChannelEnabled: channelEnabled,
		Platform:       item.Platform,
		Framework:      item.Framework,
		Status:         item.Status,
		IsFtp:          item.IsFtp,
		Host:           item.Host,
		Account:        item.Account,
		Password:       item.Password,
		LinkAddress:    buildSiteBLinkAddress(item.IsFtp, item.Host, item.Account),
		RunDirectory:   item.RunDirectory,
		Remark:         item.Remark,
		CreatedBy:      item.CreatedBy,
		CreatedAt:      item.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedBy:      item.UpdatedBy,
		UpdatedAt:      item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func buildSiteBLinkAddress(isFtp bool, host, account string) string {
	if !isFtp {
		return ""
	}
	accountText := strings.TrimSpace(account)
	if accountText == "" {
		accountText = "-"
	}
	hostText := strings.TrimSpace(host)
	return "ftp://" + accountText + "@" + hostText
}

func normalizeSiteBDomain(domain string) (string, error) {
	domain = strings.TrimSpace(strings.ToLower(domain))
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimSuffix(domain, "/")
	if domain == "" {
		return "", ErrSiteBDomainInvalid
	}
	if len(domain) > 191 {
		return "", ErrSiteBDomainInvalid
	}
	return domain, nil
}

func normalizeSiteBPlatform(platform string, lookup func(name string) (*model.Channel, error)) (string, bool, error) {
	platform = strings.TrimSpace(platform)
	if platform == "" {
		return "", false, ErrSiteBPlatformInvalid
	}
	channel, err := lookup(platform)
	if err != nil {
		if isNotFound(err) {
			return "", false, ErrSiteBPlatformInvalid
		}
		return "", false, err
	}
	return channel.Name, channel.Status == model.ChannelStatusEnabled, nil
}

func (a *App) normalizeSiteBPlatform(platform string) (string, bool, error) {
	return normalizeSiteBPlatform(platform, a.store.FindChannelByName)
}

func normalizeSiteBFramework(framework string) (string, error) {
	switch strings.TrimSpace(framework) {
	case "", model.SiteBFrameworkOther:
		return model.SiteBFrameworkOther, nil
	case model.SiteBFrameworkShopyy, model.SiteBFrameworkShopify:
		return strings.TrimSpace(framework), nil
	default:
		return "", ErrSiteBFrameworkInvalid
	}
}
