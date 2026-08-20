package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// CurrencyController 控制层：货币列表 CRUD。
type CurrencyController struct {
	app *service.App
}

// NewCurrencyController 创建货币控制器。
func NewCurrencyController(app *service.App) *CurrencyController {
	return &CurrencyController{app: app}
}

// List 货币列表，支持按名称、CODE 筛选。
func (m *CurrencyController) List(c *gin.Context) {
	list, err := m.app.ListCurrencies(service.CurrencyListQuery{
		Field:   c.Query("field"),
		Keyword: c.Query("keyword"),
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, list)
}

// Options 货币下拉，配置来自 currencies.json。
func (m *CurrencyController) Options(c *gin.Context) {
	list, err := service.ListCurrencyOptions()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, list)
}

// Create 新建货币。
func (m *CurrencyController) Create(c *gin.Context) {
	var req service.CurrencySaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.CreateCurrency(req, operatorName(c))
	if err != nil {
		writeCurrencyError(c, err)
		return
	}
	response.SuccessMsg(c, item, "created")
}

// Update 编辑货币汇率。
func (m *CurrencyController) Update(c *gin.Context) {
	id, err := parseCurrencyID(c)
	if err != nil {
		response.Fail(c, "无效的货币ID")
		return
	}
	var req service.CurrencySaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.UpdateCurrency(id, req, operatorName(c))
	if err != nil {
		writeCurrencyError(c, err)
		return
	}
	response.SuccessMsg(c, item, "updated")
}

// Delete 删除货币。
func (m *CurrencyController) Delete(c *gin.Context) {
	id, err := parseCurrencyID(c)
	if err != nil {
		response.Fail(c, "无效的货币ID")
		return
	}
	if err := m.app.DeleteCurrency(id); err != nil {
		writeCurrencyError(c, err)
		return
	}
	response.SuccessMsg(c, nil, "deleted")
}

func parseCurrencyID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid currency id")
	}
	return uint(id), nil
}

func writeCurrencyError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCurrencyNotFound):
		response.Fail(c, "货币不存在")
	case errors.Is(err, service.ErrCurrencyCodeExists):
		response.Fail(c, "该货币已存在")
	case errors.Is(err, service.ErrCurrencyCodeInvalid):
		response.Fail(c, "请选择标准货币")
	case errors.Is(err, service.ErrCurrencyRateInvalid):
		response.Fail(c, "请输入有效汇率")
	default:
		response.Fail(c, err.Error())
	}
}
