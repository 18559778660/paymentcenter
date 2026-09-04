package controller

import (
	"github.com/gin-gonic/gin"

	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// OrderController 控制层：支付订单相关接口。
type OrderController struct {
	app *service.App
}

// NewOrderController 创建订单控制器。
func NewOrderController(app *service.App) *OrderController {
	return &OrderController{app: app}
}

// List 查询订单列表。
func (o *OrderController) List(c *gin.Context) {
	orders, err := o.app.ListOrders()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, orders)
}

// Get 按订单号查询单笔订单。
func (o *OrderController) Get(c *gin.Context) {
	order, err := o.app.GetOrder(c.Param("id"))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, order)
}

// MarkPaid 人工标记订单已支付。
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

// MarkFailed 人工标记订单失败。
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
