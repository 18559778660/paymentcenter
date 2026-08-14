package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// 认证
// 获取 Authorization 头
// 如果 Authorization 头为空，则返回错误
// 如果 Authorization 头不以 Bearer 开头，则返回错误
// 认证 Token
// 获取用户
// 如果用户不存在，则返回错误
// 如果用户状态不启用，则返回错误
// 设置用户
// 继续执行
func Auth(app *service.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Unauthorized(c, "Unauthorized")
			c.Abort()
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			response.Unauthorized(c, "Unauthorized")
			c.Abort()
			return
		}

		user, err := app.AuthenticateToken(strings.TrimSpace(strings.TrimPrefix(header, prefix)))
		if err != nil {
			response.Unauthorized(c, "Unauthorized")
			c.Abort()
			return
		}

		c.Set("auth_user", user)
		c.Next()
	}
}

// 当前用户
// 获取用户
// 如果用户不存在，则返回 false
// 如果用户存在，则返回用户
func CurrentUser(c *gin.Context) (service.AuthUser, bool) {
	// 获取用户
	v, ok := c.Get("auth_user")
	if !ok {
		return service.AuthUser{}, false
	}
	user, ok := v.(service.AuthUser)
	return user, ok
}

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept-Language")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
