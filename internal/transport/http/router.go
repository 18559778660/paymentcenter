package http

import (
	"github.com/gin-gonic/gin"

	"paymentcenter/internal/config"
	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter(cfg config.Config, app *service.App) *Router {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			response.Success(c, gin.H{"service": "payment-center", "addr": cfg.Addr})
		})
		api.POST("/orders", func(c *gin.Context) {
			var req service.CreateOrderRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				response.Fail(c, err.Error())
				return
			}
			response.SuccessMsg(c, app.CreateOrder(req), "created")
		})
		api.GET("/orders", func(c *gin.Context) {
			response.Success(c, gin.H{"items": app.ListOrders()})
		})
		api.GET("/orders/:id", func(c *gin.Context) {
			order, err := app.GetOrder(c.Param("id"))
			if err != nil {
				response.Fail(c, err.Error())
				return
			}
			response.Success(c, order)
		})
		api.POST("/orders/:id/paid", func(c *gin.Context) {
			var body struct {
				ProviderRef string `json:"provider_ref"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				response.Fail(c, err.Error())
				return
			}
			order, err := app.MarkPaid(c.Param("id"), body.ProviderRef)
			if err != nil {
				response.Fail(c, err.Error())
				return
			}
			response.SuccessMsg(c, order, "updated")
		})
		api.POST("/orders/:id/failed", func(c *gin.Context) {
			var body struct {
				Message string `json:"message"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				response.Fail(c, err.Error())
				return
			}
			order, err := app.MarkFailed(c.Param("id"), body.Message)
			if err != nil {
				response.Fail(c, err.Error())
				return
			}
			response.SuccessMsg(c, order, "updated")
		})
	}

	return &Router{engine: r}
}

func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
