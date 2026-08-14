package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
)

// 错误
var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("token expired")
	ErrUserDisabled       = errors.New("user disabled")
)

// 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 登录响应
type LoginResponse struct {
	Token     string    `json:"token"`
	TokenType string    `json:"token_type"`
	ExpiresAt time.Time `json:"expires_at"`
	User      AuthUser  `json:"user"`
}

// 认证用户
type AuthUser struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// Token 有效载荷
type tokenPayload struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"exp"`
}

// 确保默认管理员
func (a *App) EnsureDefaultAdmin(username, password string) error {
	// 如果用户名或密码为空，则返回 nil
	if username == "" || password == "" {
		return nil
	}
	if _, err := a.store.FindUserByUsername(username); err == nil {
		return nil
	} else if !errors.Is(err, store.ErrNotFound) {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return a.store.CreateUser(&model.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         "admin",
		Status:       model.UserStatusEnabled,
	})
}

// 登录
func (a *App) Login(req LoginRequest) (LoginResponse, error) {
	// 获取用户
	user, err := a.store.FindUserByUsername(req.Username)
	if err != nil {
		// 如果用户不存在，则返回错误
		return LoginResponse{}, ErrInvalidCredentials
	}
	// 如果用户状态不启用，则返回错误
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

	token, expiresAt, err := a.issueToken(user)
	if err != nil {
		return LoginResponse{}, err
	}
	return LoginResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresAt: expiresAt,
		User: AuthUser{
			ID:       user.ID,
			Username: user.Username,
			Role:     user.Role,
		},
	}, nil
}

// 认证 Token
func (a *App) AuthenticateToken(token string) (AuthUser, error) {
	// 解析 Token
	payload, err := a.parseToken(token)
	if err != nil {
		return AuthUser{}, err
	}
	// 获取用户
	user, err := a.store.GetUserByID(payload.UserID)
	if err != nil {
		return AuthUser{}, ErrInvalidToken
	}
	// 如果用户状态不启用，则返回错误
	if user.Status != model.UserStatusEnabled {
		return AuthUser{}, ErrUserDisabled
	}
	// 返回用户
	return AuthUser{ID: user.ID, Username: user.Username, Role: user.Role}, nil
}

// 颁发 Token
func (a *App) issueToken(user *model.User) (string, time.Time, error) {
	// 设置过期时间
	expiresAt := time.Now().Add(a.tokenTTL)
	// 创建有效载荷
	payload := tokenPayload{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		ExpiresAt: expiresAt.Unix(),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", time.Time{}, err
	}
	body := base64.RawURLEncoding.EncodeToString(raw)
	return body + "." + a.sign(body), expiresAt, nil
}

// 解析 Token
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

// 签名
func (a *App) sign(body string) string {
	// 创建 HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(a.authSecret))
	// 写入 body
	mac.Write([]byte(body))
	// 返回 base64 编码
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
