package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
)

var (
	ErrMerchantNameExists    = errors.New("merchant name exists")
	ErrMerchantNameInvalid   = errors.New("merchant name invalid")
	ErrMerchantParentInvalid = errors.New("merchant parent invalid")
	ErrMerchantNotFound      = errors.New("merchant not found")
	ErrMerchantPasswordInvalid = errors.New("merchant password invalid")
)

var merchantNamePattern = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

// MerchantListItem 商户列表行，给前端表格。
type MerchantListItem struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	Account          string `json:"account"`
	Password         string `json:"password"`
	Contact          string `json:"contact"`
	ParentID         *uint  `json:"parentId"`
	ParentName       string `json:"parentName"`
	AutoShip         bool   `json:"autoShip"`
	ConfirmEmail     bool   `json:"confirmEmail"`
	Status           bool   `json:"status"`
	LimitMode        string `json:"limitMode"`
	RateDiff         int    `json:"rateDiff"`
	HoldRate         int    `json:"holdRate"`
	MutualHoldRate   int    `json:"mutualHoldRate"`
	HoldStatus       int    `json:"holdStatus"`
	MutualHoldStatus int    `json:"mutualHoldStatus"`
	SecretKey        string `json:"secretKey"`
	AuditSiteA       string `json:"auditSiteA"`
	Starred          bool   `json:"starred"`
	Nickname         string `json:"nickname"`
	Avatar           string `json:"avatar"`
	CreatedBy        string `json:"createdBy"`
	CreatedAt        string `json:"createdAt"`
	UpdatedBy        string `json:"updatedBy"`
	UpdatedAt        string `json:"updatedAt"`
}

// MerchantOption 上级下拉选项。
type MerchantOption struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Account string `json:"account"`
}

// CreateMerchantRequest 新增商户入参。
type CreateMerchantRequest struct {
	Name           string `json:"name" binding:"required"`
	Contact        string `json:"contact"`
	RateDiff       int    `json:"rateDiff"`
	HoldRate       int    `json:"holdRate"`
	MutualHoldRate int    `json:"mutualHoldRate"`
	ConfirmEmail   *int   `json:"confirmEmail"` // 1发送 0不发送
	AuditSiteA     string `json:"auditSiteA"`
	AutoShip       *bool  `json:"autoShip"`
}

// UpdateMerchantRequest 编辑商户 / 用户信息共用入参，未传的字段不改。
type UpdateMerchantRequest struct {
	Name           *string `json:"name"`
	Contact        *string `json:"contact"`
	ParentID       *uint   `json:"parentId"`
	RateDiff       *int    `json:"rateDiff"`
	HoldRate       *int    `json:"holdRate"`
	MutualHoldRate *int    `json:"mutualHoldRate"`
	ConfirmEmail   *int    `json:"confirmEmail"`
	AuditSiteA     *string `json:"auditSiteA"`
	AutoShip       *bool   `json:"autoShip"`
	Nickname       *string `json:"nickname"`
	Password       *string `json:"password"`
	Avatar         *string `json:"avatar"`
}

// MerchantListQuery 列表查询参数。
type MerchantListQuery struct {
	Name             string
	ParentID         *uint
	Status           *int
	HoldStatus       *int
	MutualHoldStatus *int
}

// ListMerchants 查询商户列表。
func (a *App) ListMerchants(q MerchantListQuery) ([]MerchantListItem, error) {
	list, err := a.store.ListMerchants(store.MerchantListFilter{
		Name:             q.Name,
		ParentID:         q.ParentID,
		Status:           q.Status,
		HoldStatus:       q.HoldStatus,
		MutualHoldStatus: q.MutualHoldStatus,
	})
	if err != nil {
		return nil, err
	}
	parentNames := map[uint]string{}
	for _, m := range list {
		if m.ParentID != nil {
			parentNames[*m.ParentID] = ""
		}
	}
	if len(parentNames) > 0 {
		opts, err := a.store.ListMerchantOptions()
		if err != nil {
			return nil, err
		}
		for _, o := range opts {
			parentNames[o.ID] = o.Name
		}
	}
	userIDs := make([]uint, 0, len(list))
	for _, m := range list {
		if m.UserID > 0 {
			userIDs = append(userIDs, m.UserID)
		}
	}
	userByID := map[uint]model.User{}
	if len(userIDs) > 0 {
		users, err := a.store.GetUsersByIDs(userIDs)
		if err != nil {
			return nil, err
		}
		for _, u := range users {
			userByID[u.ID] = u
		}
	}
	items := make([]MerchantListItem, 0, len(list))
	for _, m := range list {
		var user *model.User
		if u, ok := userByID[m.UserID]; ok {
			uu := u
			user = &uu
		}
		items = append(items, toMerchantListItem(m, parentNames, user))
	}
	return items, nil
}

// ListMerchantOptions 上级下拉。
func (a *App) ListMerchantOptions() ([]MerchantOption, error) {
	list, err := a.store.ListMerchantOptions()
	if err != nil {
		return nil, err
	}
	opts := make([]MerchantOption, 0, len(list))
	for _, m := range list {
		opts = append(opts, MerchantOption{ID: m.ID, Name: m.Name, Account: m.Account})
	}
	return opts, nil
}

// CreateMerchant 新建商户：同时创建可登录的 users 账号（type=merchant）。
func (a *App) CreateMerchant(req CreateMerchantRequest, operator string) (*MerchantListItem, error) {
	name := req.Name
	if !merchantNamePattern.MatchString(name) {
		return nil, ErrMerchantNameInvalid
	}
	if req.RateDiff < 0 || req.RateDiff > 100 {
		return nil, fmt.Errorf("rateDiff must be 0~100")
	}
	if _, err := a.store.FindMerchantByName(name); err == nil {
		return nil, ErrMerchantNameExists
	} else if !isNotFound(err) {
		return nil, err
	}

	account, err := a.nextMerchantAccount()
	if err != nil {
		return nil, err
	}

	plainPassword := randomPlainPassword(12)
	hash, err := a.hashPassword(plainPassword)
	if err != nil {
		return nil, err
	}
	confirmEmail := true
	if req.ConfirmEmail != nil {
		confirmEmail = *req.ConfirmEmail == 1
	}
	autoShip := true
	if req.AutoShip != nil {
		autoShip = *req.AutoShip
	}
	auditSiteA := req.AuditSiteA
	if auditSiteA != "auto" {
		auditSiteA = "manual"
	}
	holdStatus := 0
	if req.HoldRate > 0 {
		holdStatus = 1
	}
	mutualHoldStatus := 0
	if req.MutualHoldRate > 0 {
		mutualHoldStatus = 1
	}

	user := &model.User{
		Username:     account,
		PasswordHash: hash,
		RealName:     name,
		HomePath:     "/dashboard/analytics",
		Type:         model.UserTypeMerchant,
		Status:       model.UserStatusEnabled,
	}
	if err := a.store.CreateUser(user); err != nil {
		return nil, err
	}

	merchantRole, err := a.store.FindRoleByCode("merchant")
	if err != nil {
		return nil, fmt.Errorf("merchant role missing: %w", err)
	}
	if err := a.store.EnsureUserRole(user.ID, merchantRole.ID); err != nil {
		return nil, err
	}

	merchant := &model.Merchant{
		UserID:           user.ID,
		Name:             name,
		Account:          account,
		PasswordPlain:    plainPassword,
		Contact:          req.Contact,
		AutoShip:         autoShip,
		ConfirmEmail:     confirmEmail,
		Status:           model.MerchantStatusEnabled,
		LimitMode:        model.MerchantLimitModeUnified,
		RateDiff:         req.RateDiff,
		HoldRate:         req.HoldRate,
		MutualHoldRate:   req.MutualHoldRate,
		HoldStatus:       holdStatus,
		MutualHoldStatus: mutualHoldStatus,
		SecretKey:        randomSecretKey(),
		AuditSiteA:       auditSiteA,
		Starred:          false,
		CreatedBy:        operator,
		UpdatedBy:        operator,
	}
	if err := a.store.CreateMerchant(merchant); err != nil {
		return nil, err
	}

	item := toMerchantListItem(*merchant, map[uint]string{}, user)
	return &item, nil
}

// SetMerchantStarred 设置商户星标。
func (a *App) SetMerchantStarred(id uint, starred bool, operator string) (*MerchantListItem, error) {
	merchant, err := a.store.GetMerchantByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrMerchantNotFound
		}
		return nil, err
	}
	merchant.Starred = starred
	merchant.UpdatedBy = operator
	if err := a.store.SaveMerchant(merchant); err != nil {
		return nil, err
	}
	item := toMerchantListItem(*merchant, map[uint]string{}, nil)
	return &item, nil
}

// SetMerchantStatus 启用/禁用商户，同时同步登录账号状态。
func (a *App) SetMerchantStatus(id uint, enabled bool, operator string) (*MerchantListItem, error) {
	merchant, err := a.store.GetMerchantByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrMerchantNotFound
		}
		return nil, err
	}
	status := model.MerchantStatusDisabled
	userStatus := model.UserStatusDisabled
	if enabled {
		status = model.MerchantStatusEnabled
		userStatus = model.UserStatusEnabled
	}
	merchant.Status = status
	merchant.UpdatedBy = operator
	if err := a.store.SaveMerchant(merchant); err != nil {
		return nil, err
	}
	if merchant.UserID > 0 {
		user, err := a.store.GetUserByID(merchant.UserID)
		if err != nil {
			if !isNotFound(err) {
				return nil, err
			}
		} else {
			user.Status = userStatus
			if err := a.store.SaveUser(user); err != nil {
				return nil, err
			}
		}
	}
	item := toMerchantListItem(*merchant, map[uint]string{}, nil)
	return &item, nil
}

// UpdateMerchant 编辑商户资料和登录用户信息，未传字段保持原值。
func (a *App) UpdateMerchant(id uint, req UpdateMerchantRequest, operator string) (*MerchantListItem, error) {
	merchant, err := a.store.GetMerchantByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrMerchantNotFound
		}
		return nil, err
	}

	var user *model.User
	if merchant.UserID > 0 {
		user, err = a.store.GetUserByID(merchant.UserID)
		if err != nil && !isNotFound(err) {
			return nil, err
		}
	}

	needSaveUser := false
	displayName := req.Name
	if req.Nickname != nil {
		displayName = req.Nickname
	}
	if displayName != nil {
		if err := a.applyMerchantDisplayName(merchant, user, *displayName); err != nil {
			return nil, err
		}
		if user != nil {
			needSaveUser = true
		}
	}
	if req.Contact != nil {
		merchant.Contact = strings.TrimSpace(*req.Contact)
	}
	if req.ParentID != nil {
		if *req.ParentID == 0 {
			merchant.ParentID = nil
		} else {
			if *req.ParentID == merchant.ID {
				return nil, ErrMerchantParentInvalid
			}
			parent, err := a.store.GetMerchantByID(*req.ParentID)
			if err != nil {
				if isNotFound(err) {
					return nil, ErrMerchantParentInvalid
				}
				return nil, err
			}
			merchant.ParentID = &parent.ID
		}
	}
	if req.RateDiff != nil {
		if *req.RateDiff < 0 || *req.RateDiff > 100 {
			return nil, fmt.Errorf("rateDiff must be 0~100")
		}
		merchant.RateDiff = *req.RateDiff
	}
	if req.HoldRate != nil {
		merchant.HoldRate = *req.HoldRate
		if merchant.HoldRate > 0 {
			merchant.HoldStatus = 1
		} else {
			merchant.HoldStatus = 0
		}
	}
	if req.MutualHoldRate != nil {
		merchant.MutualHoldRate = *req.MutualHoldRate
		if merchant.MutualHoldRate > 0 {
			merchant.MutualHoldStatus = 1
		} else {
			merchant.MutualHoldStatus = 0
		}
	}
	if req.ConfirmEmail != nil {
		merchant.ConfirmEmail = *req.ConfirmEmail == 1
	}
	if req.AuditSiteA != nil {
		auditSiteA := *req.AuditSiteA
		if auditSiteA != "auto" {
			auditSiteA = "manual"
		}
		merchant.AuditSiteA = auditSiteA
	}
	if req.AutoShip != nil {
		merchant.AutoShip = *req.AutoShip
	}

	if user != nil {
		if req.Avatar != nil {
			user.Avatar = strings.TrimSpace(*req.Avatar)
			needSaveUser = true
		}
		if req.Password != nil {
			plain := *req.Password
			if len(plain) < 6 || len(plain) > 20 {
				return nil, ErrMerchantPasswordInvalid
			}
			hash, err := a.hashPassword(plain)
			if err != nil {
				return nil, err
			}
			user.PasswordHash = hash
			merchant.PasswordPlain = plain
			needSaveUser = true
		}
	}

	merchant.UpdatedBy = operator
	if err := a.store.SaveMerchant(merchant); err != nil {
		return nil, err
	}
	if needSaveUser && user != nil {
		if err := a.store.SaveUser(user); err != nil {
			return nil, err
		}
	}

	item := toMerchantListItem(*merchant, map[uint]string{}, user)
	return &item, nil
}

// nextMerchantAccount 生成登录账号：WIN00000、WIN00001…按已有最大值 +1。
func (a *App) nextMerchantAccount() (string, error) {
	seq, err := a.store.MaxWINMerchantAccountSeq()
	if err != nil {
		return "", err
	}
	//因为数据库有唯一值 直接+1就好;循环判断有没撞号没必要
	seq++
	if seq > 99999 {
		return "", fmt.Errorf("merchant account sequence overflow")
	}
	return fmt.Sprintf("WIN%05d", seq), nil
}

// applyMerchantDisplayName 商户名和登录显示名同步更新。
func (a *App) applyMerchantDisplayName(merchant *model.Merchant, user *model.User, raw string) error {
	name := strings.TrimSpace(raw)
	if !merchantNamePattern.MatchString(name) {
		return ErrMerchantNameInvalid
	}
	if exist, err := a.store.FindMerchantByName(name); err == nil && exist.ID != merchant.ID {
		return ErrMerchantNameExists
	} else if err != nil && !isNotFound(err) {
		return err
	}
	merchant.Name = name
	if user != nil {
		user.RealName = name
	}
	return nil
}

// toMerchantListItem 转换为商户列表项。
func toMerchantListItem(m model.Merchant, parentNames map[uint]string, user *model.User) MerchantListItem {
	parentName := "-"
	if m.ParentID != nil {
		if name, ok := parentNames[*m.ParentID]; ok && name != "" {
			parentName = name
		} else {
			parentName = fmt.Sprintf("#%d", *m.ParentID)
		}
	}
	nickname := m.Name
	avatar := ""
	if user != nil {
		if user.RealName != "" {
			nickname = user.RealName
		}
		avatar = user.Avatar
	}
	return MerchantListItem{
		ID:               m.ID,
		Name:             m.Name,
		Account:          m.Account,
		Password:         m.PasswordPlain,
		Contact:          m.Contact,
		ParentID:         m.ParentID,
		ParentName:       parentName,
		AutoShip:         m.AutoShip,
		ConfirmEmail:     m.ConfirmEmail,
		Status:           m.Status == model.MerchantStatusEnabled,
		LimitMode:        m.LimitMode,
		RateDiff:         m.RateDiff,
		HoldRate:         m.HoldRate,
		MutualHoldRate:   m.MutualHoldRate,
		HoldStatus:       m.HoldStatus,
		MutualHoldStatus: m.MutualHoldStatus,
		SecretKey:        m.SecretKey,
		AuditSiteA:       m.AuditSiteA,
		Starred:          m.Starred,
		Nickname:         nickname,
		Avatar:           avatar,
		CreatedBy:        m.CreatedBy,
		CreatedAt:        m.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedBy:        m.UpdatedBy,
		UpdatedAt:        m.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// randomPlainPassword 生成随机密码。
func randomPlainPassword(n int) string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789"
	out := make([]byte, n)
	for i := range out {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			out[i] = chars[time.Now().UnixNano()%int64(len(chars))]
			continue
		}
		out[i] = chars[idx.Int64()]
	}
	return string(out)
}

// randomSecretKey 生成随机密钥。
func randomSecretKey() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}
