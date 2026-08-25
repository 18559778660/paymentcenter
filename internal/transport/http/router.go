package http

import (
	"os"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/config"
	"paymentcenter/internal/controller"
	"paymentcenter/internal/middleware"
	"paymentcenter/internal/service"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter(cfg config.Config, app *service.App) *Router {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS())
	_ = os.MkdirAll("./uploads", 0o755)
	r.Static("/api/files", "./uploads")

	healthController := controller.NewHealthController(cfg)
	authController := controller.NewAuthController(app)
	menuController := controller.NewMenuController(app)
	merchantController := controller.NewMerchantController(app)
	merchantGroupController := controller.NewMerchantGroupController(app)
	cardTypeController := controller.NewCardTypeController(app)
	currencyController := controller.NewCurrencyController(app)
	countryController := controller.NewCountryController(app)
	channelController := controller.NewChannelController(app)
	gatewayController := controller.NewGatewayController(app)
	orderController := controller.NewOrderController(app)

	api := r.Group("/api")
	{
		// A 站对接网关，无需登录
		api.GET("/gateway", gatewayController.Access)

		api.POST("/auth/login", authController.Login)
		// 退出不校验 token：前端可能已清本地 token，或 token 已过期，仍应能成功退出
		api.POST("/auth/logout", authController.Logout)

		authed := api.Group("")
		authed.Use(middleware.Auth(app))
		{
			// 用户信息
			authed.GET("/user/info", authController.UserInfo)
			// 用户权限码
			authed.GET("/auth/codes", authController.Codes)
			// 用户菜单
			authed.GET("/menu/all", authController.Menus)

			// 菜单管理
			authed.GET("/system/menu/list", menuController.List)
			authed.GET("/system/menu/name-exists", menuController.NameExists)
			authed.GET("/system/menu/path-exists", menuController.PathExists)
			authed.POST("/system/menu", menuController.Create)
			authed.PUT("/system/menu/:id", menuController.Update)
			authed.DELETE("/system/menu/:id", menuController.Delete)

			// 商户管理
			authed.GET("/merchants", merchantController.List)
			authed.GET("/merchants/options", merchantController.Options)
			authed.POST("/merchants", merchantController.Create)
			authed.PUT("/merchants/:id", merchantController.Update)
			authed.PUT("/merchants/:id/star", merchantController.SetStar)
			authed.PUT("/merchants/:id/status", merchantController.SetStatus)
			authed.POST("/upload/avatar", merchantController.UploadAvatar)

			// 商户分组
			authed.GET("/merchant-groups", merchantGroupController.List)
			authed.POST("/merchant-groups", merchantGroupController.Create)
			authed.PUT("/merchant-groups/:id", merchantGroupController.Update)
			authed.DELETE("/merchant-groups/:id", merchantGroupController.Delete)

			// 卡头验证
			authed.GET("/card-types", cardTypeController.List)
			authed.GET("/card-types/brands", cardTypeController.Brands)
			authed.POST("/card-types", cardTypeController.Create)
			authed.PUT("/card-types/:id", cardTypeController.Update)

			// 货币列表
			authed.GET("/currencies", currencyController.List)
			authed.GET("/currencies/options", currencyController.Options)
			authed.POST("/currencies", currencyController.Create)
			authed.PUT("/currencies/:id", currencyController.Update)
			authed.DELETE("/currencies/:id", currencyController.Delete)

			// 国家列表
			authed.GET("/countries", countryController.List)
			authed.GET("/countries/options", countryController.Options)
			authed.POST("/countries", countryController.Create)
			authed.PUT("/countries/:id", countryController.Update)
			authed.DELETE("/countries/:id", countryController.Delete)

			// 通道列表
			authed.GET("/channels", channelController.List)
			authed.POST("/channels", channelController.Create)
			authed.PUT("/channels/:id", channelController.Update)
			authed.PUT("/channels/:id/limits", channelController.UpdateLimits)
			authed.PUT("/channels/:id/status", channelController.SetStatus)
			authed.POST("/channels/:id/package", channelController.UploadPackage)
			authed.GET("/channels/:id/package", channelController.DownloadPackage)

			// 健康检查
			authed.GET("/health", healthController.Health)

			// 订单管理
			authed.POST("/orders", orderController.Create)
			authed.GET("/orders", orderController.List)
			authed.GET("/orders/:id", orderController.Get)
			authed.POST("/orders/:id/paid", orderController.MarkPaid)
			authed.POST("/orders/:id/failed", orderController.MarkFailed)
		}
	}

	return &Router{engine: r}
}

func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
