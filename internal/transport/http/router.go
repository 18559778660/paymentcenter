package http

import (
	"github.com/gin-gonic/gin"

	"paymentcenter/internal/config"
	"paymentcenter/internal/controller"
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
	orderController := controller.NewOrderController(app)

	api := r.Group("/api")
	{
		api.GET("/health", healthController.Health)

		api.POST("/orders", orderController.Create)
		api.GET("/orders", orderController.List)
		api.GET("/orders/:id", orderController.Get)
		api.POST("/orders/:id/paid", orderController.MarkPaid)
		api.POST("/orders/:id/failed", orderController.MarkFailed)
	}

	return &Router{engine: r}
}

func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
