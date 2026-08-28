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
	PlatformID     uint   `json:"platformId"`
	Platform       string `json:"platform"`
	PlatformName   string `json:"platformName"`
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
	ID         *uint
	Domain     string
	Remark     string
	Status     *bool
	PlatformID *uint
}

// CreateSiteBRequest 新增 B 站。
type CreateSiteBRequest struct {
	Domain     string `json:"domain" binding:"required"`
	PlatformID uint   `json:"platformId" binding:"required"`
	Framework  string `json:"framework"`
	IsFtp      *bool  `json:"isFtp"`
	Host         string `json:"host"`
	Account      string `json:"account"`
	Password     string `json:"password"`
	RunDirectory string `json:"runDirectory"`
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
		ID:         q.ID,
		Domain:     q.Domain,
		Remark:     q.Remark,
		Status:     q.Status,
		PlatformID: q.PlatformID,
	})
	if err != nil {
		return nil, err
	}
	platformMap, err := a.loadPlatformMetaMap()
	if err != nil {
		return nil, err
	}
	platformStatus, err := a.loadPlatformChannelStatusMap()
	if err != nil {
		return nil, err
	}
	out := make([]SiteBListItem, 0, len(list))
	for _, item := range list {
		out = append(out, toSiteBListItem(item, platformMap, platformStatus))
	}
	return out, nil
}

// CreateSiteB 新增 B 站。
func (a *App) CreateSiteB(req CreateSiteBRequest, operator string) (*SiteBListItem, error) {
	domain, err := normalizeSiteBDomain(req.Domain)
	if err != nil {
		return nil, err
	}
	platform, err := a.getPlatformByID(req.PlatformID)
	if err != nil {
		return nil, ErrSiteBPlatformInvalid
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
	platformStatus, err := a.loadPlatformChannelStatusMap()
	if err != nil {
		return nil, err
	}
	isFtp := true
	if req.IsFtp != nil {
		isFtp = *req.IsFtp
	}
	runDirectory := strings.TrimSpace(req.RunDirectory)
	if runDirectory == "" && platform.Code != model.PlatformCodeStripe {
		runDirectory = model.SiteBDefaultRunDirectory
	}
	item := &model.SiteB{
		Domain:         domain,
		PlatformID:     platform.ID,
		Framework:      framework,
		Status:         true,
		ChannelEnabled: platformStatus[platform.ID],
		IsFtp:          isFtp,
		Host:           strings.TrimSpace(req.Host),
		Account:        strings.TrimSpace(req.Account),
		Password:       req.Password,
		RunDirectory:   runDirectory,
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

func (a *App) loadPlatformChannelStatusMap() (map[uint]bool, error) {
	channels, err := a.store.ListChannels(store.ChannelListFilter{})
	if err != nil {
		return nil, err
	}
	result := map[uint]bool{}
	for _, channel := range channels {
		if channel.PlatformID == 0 {
			continue
		}
		if channel.Status == model.ChannelStatusEnabled {
			result[channel.PlatformID] = true
		} else if _, ok := result[channel.PlatformID]; !ok {
			result[channel.PlatformID] = false
		}
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
	platformMap, err := a.loadPlatformMetaMap()
	if err != nil {
		return nil, err
	}
	platformStatus, err := a.loadPlatformChannelStatusMap()
	if err != nil {
		return nil, err
	}
	out := toSiteBListItem(*item, platformMap, platformStatus)
	return &out, nil
}

func toSiteBListItem(item model.SiteB, platformMap map[uint]model.Platform, platformStatus map[uint]bool) SiteBListItem {
	platformCode := ""
	platformName := ""
	if platform, ok := platformMap[item.PlatformID]; ok {
		platformCode = platform.Code
		platformName = platform.Name
	}
	channelEnabled := platformStatus[item.PlatformID]
	return SiteBListItem{
		ID:             item.ID,
		Domain:         item.Domain,
		Channel:        platformCode,
		ChannelEnabled: channelEnabled,
		PlatformID:     item.PlatformID,
		Platform:       platformCode,
		PlatformName:   platformName,
		Framework:      item.Framework,
		Status:         item.Status,
		IsFtp:          item.IsFtp,
		Host:           item.Host,
		Account:        item.Account,
		Password:       item.Password,
		LinkAddress:    buildSiteBLinkAddress(item.Host, item.Account, item.Password),
		RunDirectory:   item.RunDirectory,
		Remark:         item.Remark,
		CreatedBy:      item.CreatedBy,
		CreatedAt:      item.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedBy:      item.UpdatedBy,
		UpdatedAt:      item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func buildSiteBLinkAddress(host, account, password string) string {
	host = strings.TrimSpace(host)
	account = strings.TrimSpace(account)
	password = strings.TrimSpace(password)
	if host == "" && account == "" && password == "" {
		return "ftp://:@"
	}
	return formatSiteBLinkHost(host)
}

func formatSiteBLinkHost(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}
	lower := strings.ToLower(host)
	if strings.HasPrefix(lower, "https://") {
		return host
	}
	if strings.HasPrefix(lower, "http://") {
		return "https://" + host[len("http://"):]
	}
	return "https://" + host
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
