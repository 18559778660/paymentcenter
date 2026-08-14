package service

// 业务层：登录、用户信息、权限码、token 签发与校验。

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password") // 账号或密码错误
	ErrInvalidToken       = errors.New("invalid token")                // token 无效
	ErrExpiredToken       = errors.New("token expired")                // token 过期
	ErrUserDisabled       = errors.New("user disabled")                // 账号已禁用
)

// LoginRequest 登录入参，对应前端 POST /auth/login 的 JSON。
type LoginRequest struct {
	Username string `json:"username" binding:"required"` // 登录账号
	Password string `json:"password" binding:"required"` // 明文密码
}

// LoginResponse 登录出参。前端只取 data.accessToken。
type LoginResponse struct {
	AccessToken string `json:"accessToken"` // 访问令牌，后续请求放在 Authorization 头
}

// AuthUser 鉴权通过后放在上下文里的当前用户。
type AuthUser struct {
	ID       uint     `json:"id"`       // 用户ID
	Username string   `json:"username"` // 登录账号
	Roles    []string `json:"roles"`    // 角色编码列表，例如 super
}

// UserInfoResponse 对齐 Vben GET /user/info 的 data 字段。
type UserInfoResponse struct {
	UserID   string   `json:"userId"`   // 用户ID，字符串
	Username string   `json:"username"` // 登录账号
	RealName string   `json:"realName"` // 显示名
	Avatar   string   `json:"avatar"`   // 头像
	Desc     string   `json:"desc"`     // 描述，可空
	HomePath string   `json:"homePath"` // 登录后跳转页
	Roles    []string `json:"roles"`    // 角色编码，前端用来过滤菜单
	Token    string   `json:"token"`    // 前端类型里有这个字段，可给空字符串
}

// tokenPayload 签发 token 时写入的内容，只存用户标识和过期时间。
type tokenPayload struct {
	UserID    uint   `json:"user_id"`  // 用户ID
	Username  string `json:"username"` // 登录账号
	ExpiresAt int64  `json:"exp"`      // 过期时间戳
}

// Login 校验账号密码，签发 accessToken，并更新最后登录时间。
func (a *App) Login(req LoginRequest) (LoginResponse, error) {
	user, err := a.store.FindUserByUsername(req.Username)
	if err != nil {
		return LoginResponse{}, ErrInvalidCredentials
	}
	if user.Status != model.UserStatusEnabled {
		return LoginResponse{}, ErrUserDisabled
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return LoginResponse{}, ErrInvalidCredentials
	}

	now := time.Now()
	user.LastLoginAt = &now
	if err := a.store.SaveUser(user); err != nil {
		return LoginResponse{}, err
	}

	token, _, err := a.issueToken(user)
	if err != nil {
		return LoginResponse{}, err
	}
	return LoginResponse{AccessToken: token}, nil
}

// GetUserInfo 按用户ID组装前端需要的用户信息。
func (a *App) GetUserInfo(userID uint) (UserInfoResponse, error) {
	user, err := a.store.GetUserByID(userID)
	if err != nil {
		return UserInfoResponse{}, ErrInvalidToken
	}
	if user.Status != model.UserStatusEnabled {
		return UserInfoResponse{}, ErrUserDisabled
	}
	roles, err := a.roleCodes(user.ID)
	if err != nil {
		return UserInfoResponse{}, err
	}
	homePath := user.HomePath
	if homePath == "" {
		homePath = "/dashboard/analytics"
	}
	realName := user.RealName
	if realName == "" {
		realName = user.Username
	}
	return UserInfoResponse{
		UserID:   strconv.FormatUint(uint64(user.ID), 10),
		Username: user.Username,
		RealName: realName,
		Avatar:   user.Avatar,
		Desc:     "",
		HomePath: homePath,
		Roles:    roles,
		Token:    "",
	}, nil
}

// GetAccessCodes 取出该用户角色下所有菜单权限码，给前端按钮权限用。
func (a *App) GetAccessCodes(userID uint) ([]string, error) {
	menus, err := a.store.ListMenusByUserID(userID)
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0)
	seen := map[string]struct{}{}
	for _, menu := range menus {
		if menu.AuthCode == "" {
			continue
		}
		if _, ok := seen[menu.AuthCode]; ok {
			continue
		}
		seen[menu.AuthCode] = struct{}{}
		codes = append(codes, menu.AuthCode)
	}
	return codes, nil
}

// AuthenticateToken 校验 token，确认用户仍存在且未禁用，供中间件调用。
func (a *App) AuthenticateToken(token string) (AuthUser, error) {
	payload, err := a.parseToken(token)
	if err != nil {
		return AuthUser{}, err
	}
	user, err := a.store.GetUserByID(payload.UserID)
	if err != nil {
		return AuthUser{}, ErrInvalidToken
	}
	if user.Status != model.UserStatusEnabled {
		return AuthUser{}, ErrUserDisabled
	}
	roles, err := a.roleCodes(user.ID)
	if err != nil {
		return AuthUser{}, err
	}
	return AuthUser{ID: user.ID, Username: user.Username, Roles: roles}, nil
}

// roleCodes 查询用户已绑定的角色编码。
func (a *App) roleCodes(userID uint) ([]string, error) {
	roles, err := a.store.ListRolesByUserID(userID)
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(roles))
	for _, role := range roles {
		codes = append(codes, role.Code)
	}
	if len(codes) == 0 {
		return []string{}, nil
	}
	return codes, nil
}

// issueToken 签发访问令牌：内容 + HMAC 签名。
func (a *App) issueToken(user *model.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(a.tokenTTL)
	payload := tokenPayload{
		UserID:    user.ID,
		Username:  user.Username,
		ExpiresAt: expiresAt.Unix(),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", time.Time{}, err
	}
	body := base64.RawURLEncoding.EncodeToString(raw)
	return body + "." + a.sign(body), expiresAt, nil
}

// parseToken 解析并校验令牌签名、过期时间。
func (a *App) parseToken(token string) (tokenPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return tokenPayload{}, ErrInvalidToken
	}

	expected := a.sign(parts[0])
	if subtle.ConstantTimeCompare([]byte(expected), []byte(parts[1])) != 1 {
		return tokenPayload{}, ErrInvalidToken
	}

	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return tokenPayload{}, ErrInvalidToken
	}
	var payload tokenPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return tokenPayload{}, ErrInvalidToken
	}
	if payload.ExpiresAt <= time.Now().Unix() {
		return tokenPayload{}, ErrExpiredToken
	}
	return payload, nil
}

// sign 用 AUTH_SECRET 对 token 内容做 HMAC-SHA256 签名。
func (a *App) sign(body string) string {
	mac := hmac.New(sha256.New, []byte(a.authSecret))
	mac.Write([]byte(body))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// hashPassword 用 bcrypt 哈希明文密码，写入数据库。
func (a *App) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// isNotFound 判断是否为数据库未找到记录。
func isNotFound(err error) bool {
	return errors.Is(err, store.ErrNotFound)
}
