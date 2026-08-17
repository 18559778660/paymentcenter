package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
)

var (
	ErrMerchantNameExists    = errors.New("merchant name exists")
	ErrMerchantNameInvalid   = errors.New("merchant name invalid")
	ErrMerchantParentInvalid = errors.New("merchant parent invalid")
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
	items := make([]MerchantListItem, 0, len(list))
	for _, m := range list {
		items = append(items, toMerchantListItem(m, parentNames))
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

	item := toMerchantListItem(*merchant, map[uint]string{})
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

// toMerchantListItem 转换为商户列表项。
func toMerchantListItem(m model.Merchant, parentNames map[uint]string) MerchantListItem {
	parentName := "-"
	if m.ParentID != nil {
		if name, ok := parentNames[*m.ParentID]; ok && name != "" {
			parentName = name
		} else {
			parentName = fmt.Sprintf("#%d", *m.ParentID)
		}
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
