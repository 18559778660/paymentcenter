package controller

import (
	"github.com/gin-gonic/gin"

	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// 订单控制器
type OrderController struct {
	app *service.App
}

// 创建订单控制器
func NewOrderController(app *service.App) *OrderController {
	return &OrderController{app: app}
}

// 创建订单
func (o *OrderController) Create(c *gin.Context) {
	var req service.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err.Error())
		return
	}
	res, err := o.app.CreateOrder(req)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.SuccessMsg(c, res, "created")
}

// 获取订单列表
func (o *OrderController) List(c *gin.Context) {
	orders, err := o.app.ListOrders()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": orders})
}

// 获取订单
func (o *OrderController) Get(c *gin.Context) {
	order, err := o.app.GetOrder(c.Param("id"))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, order)
}

// 标记订单已支付
func (o *OrderController) MarkPaid(c *gin.Context) {
	var body struct {
		ProviderRef string `json:"provider_ref"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, err.Error())
		return
	}
	order, err := o.app.MarkPaid(c.Param("id"), body.ProviderRef)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.SuccessMsg(c, order, "updated")
}

// 标记订单已失败
func (o *OrderController) MarkFailed(c *gin.Context) {
	var body struct {
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, err.Error())
		return
	}
	order, err := o.app.MarkFailed(c.Param("id"), body.Message)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.SuccessMsg(c, order, "updated")
}
