package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// OrderLogController 控制层：订单日志。
type OrderLogController struct {
	app *service.App
}

// NewOrderLogController 创建订单日志控制器。
func NewOrderLogController(app *service.App) *OrderLogController {
	return &OrderLogController{app: app}
}

// List 查询订单日志列表。
func (o *OrderLogController) List(c *gin.Context) {
	var merchantID uint
	if raw := c.Query("merchantId"); raw != "" {
		n, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			response.Fail(c, "无效的商户ID")
			return
		}
		merchantID = uint(n)
	}
	list, err := o.app.ListOrderLogs(service.OrderLogListQuery{
		MerchantID:    merchantID,
		MerchantOrder: c.Query("orderNo"),
		CreatedFrom:   c.Query("createdFrom"),
		CreatedTo:     c.Query("createdTo"),
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, list)
}
