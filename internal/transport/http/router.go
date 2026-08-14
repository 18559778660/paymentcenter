package http

import (
	"github.com/gin-gonic/gin"

	"paymentcenter/internal/config"
	"paymentcenter/internal/controller"
	"paymentcenter/internal/middleware"
	"paymentcenter/internal/service"
)

type Router struct {
	engine *gin.Engine
}

// 创建路由
func NewRouter(cfg config.Config, app *service.App) *Router {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	healthController := controller.NewHealthController(cfg)
	authController := controller.NewAuthController(app)
	orderController := controller.NewOrderController(app)

	api := r.Group("/api")
	{
		api.POST("/auth/login", authController.Login)

		authed := api.Group("")
		authed.Use(middleware.Auth(app))
		{
			authed.GET("/health", healthController.Health)

			authed.POST("/orders", orderController.Create)
			authed.GET("/orders", orderController.List)
			authed.GET("/orders/:id", orderController.Get)
			authed.POST("/orders/:id/paid", orderController.MarkPaid)
			authed.POST("/orders/:id/failed", orderController.MarkFailed)
		}
	}

	return &Router{engine: r}
}

// 启动服务器
func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
