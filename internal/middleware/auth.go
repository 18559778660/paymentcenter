package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// 认证中间件
func Auth(app *service.App) gin.HandlerFunc {
	// 认证
	return func(c *gin.Context) {
		// 获取 Authorization 头
		header := c.GetHeader("Authorization")
		// 如果 Authorization 头为空，则返回错误
		if header == "" {
			// 返回错误
			response.Fail(c, "missing authorization header")
			c.Abort()
			return
		}

		// 定义前缀
		const prefix = "Bearer "
		// 如果 Authorization 头不以 Bearer 开头，则返回错误
		if !strings.HasPrefix(header, prefix) {
			response.Fail(c, "invalid authorization header")
			c.Abort()
			return
		}
		// 认证 Token
		user, err := app.AuthenticateToken(strings.TrimSpace(strings.TrimPrefix(header, prefix)))
		if err != nil {
			response.Fail(c, err.Error())
			c.Abort()
			return
		}

		// 设置用户
		c.Set("auth_user", user)
		// 继续执行
		c.Next()
	}
}
