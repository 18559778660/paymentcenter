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

// 创建路由
func NewRouter(cfg config.Config, app *service.App) *Router {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api")
	{
		// 健康检查
		api.GET("/health", func(c *gin.Context) {
			response.Success(c, gin.H{"service": "payment-center", "addr": cfg.Addr})
		})
		// 创建订单
		api.POST("/orders", func(c *gin.Context) {
			var req service.CreateOrderRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				response.Fail(c, err.Error())
				return
			}
			response.SuccessMsg(c, app.CreateOrder(req), "created")
		})
		// 获取订单列表
		api.GET("/orders", func(c *gin.Context) {
			response.Success(c, gin.H{"items": app.ListOrders()})
		})
		// 获取订单
		api.GET("/orders/:id", func(c *gin.Context) {
			order, err := app.GetOrder(c.Param("id"))
			if err != nil {
				response.Fail(c, err.Error())
				return
			}
			response.Success(c, order)
		})
		// 标记订单已支付
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
		// 标记订单已失败
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
