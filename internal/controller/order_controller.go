package controller

import (
	"fmt"
	"strconv"

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
	q, err := parseOrderListQuery(c)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	orders, err := o.app.ListOrders(q)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, orders)
}

func parseOrderListQuery(c *gin.Context) (service.OrderListQuery, error) {
	var q service.OrderListQuery
	if raw := c.Query("merchantId"); raw != "" {
		n, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return q, fmt.Errorf("无效的商户ID")
		}
		q.MerchantID = uint(n)
	}
	if raw := c.Query("accountId"); raw != "" {
		n, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return q, fmt.Errorf("无效的账号ID")
		}
		q.ChannelAccountID = uint(n)
	}
	q.CustomerEmail = c.Query("email")
	q.CustomerPhone = c.Query("phone")
	q.CustomerIP = c.Query("ip")
	q.MerchantOrder = c.Query("orderNo")
	q.MerchantSite = c.Query("siteA")
	q.Status = c.Query("status")
	q.Provider = c.Query("payMethod")
	q.SiteB = c.Query("siteB")
	q.CreatedFrom = c.Query("createdFrom")
	q.CreatedTo = c.Query("createdTo")
	q.Country = c.Query("country")
	q.Currency = c.Query("currency")
	q.MinAmountMajor = c.Query("minAmount")
	q.MaxAmountMajor = c.Query("maxAmount")
	q.ProviderRef = c.Query("transactionNo")
	return q, nil
}

// Summary 查询订单汇总栏。
func (o *OrderController) Summary(c *gin.Context) {
	summary, err := o.app.GetOrderSummary()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, summary)
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
