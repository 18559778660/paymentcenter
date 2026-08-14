package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

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

func CurrentUser(c *gin.Context) (service.AuthUser, bool) {
	v, ok := c.Get("auth_user")
	if !ok {
		return service.AuthUser{}, false
	}
	user, ok := v.(service.AuthUser)
	return user, ok
}

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
