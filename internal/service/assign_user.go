package service

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
)

var (
	ErrAssignUserNotFound              = errors.New("assign user not found")
	ErrAssignUserUsernameExists        = errors.New("assign user username exists")
	ErrAssignUserUsernameInvalid       = errors.New("assign user username invalid")
	ErrAssignUserPasswordInvalid       = errors.New("assign user password invalid")
	ErrChannelAccountAlreadyAssigned   = errors.New("channel account already assigned")
)

var assignUsernamePattern = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

// AssignUserListItem 账号分配列表行。
type AssignUserListItem struct {
	ID            uint   `json:"id"`
	Username      string `json:"username"`
	Nickname      string `json:"nickname"`
	AssignedCount int    `json:"assignedCount"`
}

// AssignUserListQuery 列表筛选。
type AssignUserListQuery struct {
	Field   string
	Keyword string
}

// CreateAssignUserRequest 新增分配子账号。
type CreateAssignUserRequest struct {
	Username string `json:"username" binding:"required"`
	Nickname string `json:"nickname"`
	Password string `json:"password" binding:"required"`
}

// AssignUserAccountItem 子账号下的通道账号行。
type AssignUserAccountItem struct {
	ID            uint   `json:"id"`
	ChannelName   string `json:"channelName"`
	Assigned      bool   `json:"assigned"`
	AccountStatus bool   `json:"accountStatus"`
	SiteB         string `json:"siteB"`
	Channel       string `json:"channel"`
	Remark        string `json:"remark"`
	PaymentMethod string `json:"paymentMethod"`
}

// AssignUserAccountListQuery 子账号通道账号列表筛选。
type AssignUserAccountListQuery struct {
	ChannelID *uint
}

// ListAssignUsers 分配子账号列表。
func (a *App) ListAssignUsers(q AssignUserListQuery) ([]AssignUserListItem, error) {
	list, err := a.store.ListUsersByType(model.UserTypeDistribution, store.AssignUserListFilter{
		Field:   q.Field,
		Keyword: q.Keyword,
	})
	if err != nil {
		return nil, err
	}
	userIDs := make([]uint, 0, len(list))
	for _, item := range list {
		userIDs = append(userIDs, item.ID)
	}
	countMap, err := a.store.CountChannelAccountsByAssignedUserIDs(userIDs)
	if err != nil {
		return nil, err
	}
	out := make([]AssignUserListItem, 0, len(list))
	for _, item := range list {
		nickname := strings.TrimSpace(item.RealName)
		if nickname == "" {
			nickname = item.Username
		}
		out = append(out, AssignUserListItem{
			ID:            item.ID,
			Username:      item.Username,
			Nickname:      nickname,
			AssignedCount: int(countMap[item.ID]),
		})
	}
	return out, nil
}

// CreateAssignUser 新增可登录的分配子账号。
func (a *App) CreateAssignUser(req CreateAssignUserRequest) (*AssignUserListItem, error) {
	username := strings.TrimSpace(req.Username)
	if !assignUsernamePattern.MatchString(username) {
		return nil, ErrAssignUserUsernameInvalid
	}
	password := req.Password
	if utf8.RuneCountInString(password) < 6 || utf8.RuneCountInString(password) > 20 {
		return nil, ErrAssignUserPasswordInvalid
	}
	if _, err := a.store.FindUserByUsername(username); err == nil {
		return nil, ErrAssignUserUsernameExists
	} else if !isNotFound(err) {
		return nil, err
	}
	hash, err := a.hashPassword(password)
	if err != nil {
		return nil, err
	}
	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		nickname = username
	}
	user := &model.User{
		Username:     username,
		PasswordHash: hash,
		RealName:     nickname,
		HomePath:     "/dashboard/analytics",
		Type:         model.UserTypeDistribution,
		Status:       model.UserStatusEnabled,
	}
	if err := a.store.CreateUser(user); err != nil {
		return nil, err
	}
	role, err := a.store.FindRoleByCode("distribution")
	if err != nil {
		return nil, err
	}
	if err := a.store.EnsureUserRole(user.ID, role.ID); err != nil {
		return nil, err
	}
	item := AssignUserListItem{
		ID:            user.ID,
		Username:      user.Username,
		Nickname:      nickname,
		AssignedCount: 0,
	}
	return &item, nil
}

// ListAssignUserAccounts 列出全部通道账号，并标记是否分配给当前子账号。
func (a *App) ListAssignUserAccounts(userID uint, q AssignUserAccountListQuery) ([]AssignUserAccountItem, error) {
	user, err := a.store.GetUserByID(userID)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrAssignUserNotFound
		}
		return nil, err
	}
	if user.Type != model.UserTypeDistribution {
		return nil, ErrAssignUserNotFound
	}
	accounts, err := a.store.ListChannelAccounts(store.ChannelAccountListFilter{})
	if err != nil {
		return nil, err
	}
	channelMap, err := a.loadChannelNameMap()
	if err != nil {
		return nil, err
	}
	siteBMap, err := a.loadSiteBDomainMap()
	if err != nil {
		return nil, err
	}
	out := make([]AssignUserAccountItem, 0, len(accounts))
	for _, item := range accounts {
		if q.ChannelID != nil && *q.ChannelID > 0 && item.ChannelID != *q.ChannelID {
			continue
		}
		channelName := strings.TrimSpace(item.Alias)
		if channelName == "" {
			channelName = item.AccountNo
		}
		out = append(out, AssignUserAccountItem{
			ID:            item.ID,
			ChannelName:   channelName,
			Assigned:      item.AssignedUserID == user.ID,
			AccountStatus: item.Status == model.ChannelAccountStatusEnabled,
			SiteB:         siteBMap[item.SiteBID],
			Channel:       channelMap[item.ChannelID],
			Remark:        item.Remark,
			PaymentMethod: item.PaymentMethod,
		})
	}
	return out, nil
}

// SetAssignUserAccountAssignment 设置通道账号是否分配给子账号（一对一：账号只能分给一个子账号）。
func (a *App) SetAssignUserAccountAssignment(userID, accountID uint, assigned bool, operator string) error {
	user, err := a.store.GetUserByID(userID)
	if err != nil {
		if isNotFound(err) {
			return ErrAssignUserNotFound
		}
		return err
	}
	if user.Type != model.UserTypeDistribution {
		return ErrAssignUserNotFound
	}
	account, err := a.store.GetChannelAccountByID(accountID)
	if err != nil {
		if isNotFound(err) {
			return ErrChannelAccountNotFound
		}
		return err
	}
	if assigned {
		if account.AssignedUserID != 0 && account.AssignedUserID != user.ID {
			return ErrChannelAccountAlreadyAssigned
		}
		account.AssignedUserID = user.ID
	} else {
		if account.AssignedUserID != user.ID {
			return nil
		}
		account.AssignedUserID = 0
	}
	account.UpdatedBy = operator
	return a.store.SaveChannelAccount(account)
}

// GetUserForScope 查询用户完整信息，供列表数据范围控制。
func (a *App) GetUserForScope(userID uint) (*model.User, error) {
	return a.store.GetUserByID(userID)
}

func (a *App) loadAssignedUserNameMap(userIDs []uint) (map[uint]string, error) {
	result := make(map[uint]string, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}
	users, err := a.store.GetUsersByIDs(userIDs)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if user.Type != model.UserTypeDistribution {
			continue
		}
		name := strings.TrimSpace(user.RealName)
		if name == "" {
			name = user.Username
		}
		result[user.ID] = name
	}
	return result, nil
}
