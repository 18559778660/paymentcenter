package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// CardTypeController 控制层：卡头验证 / 卡类型 CRUD。
type CardTypeController struct {
	app *service.App
}

// NewCardTypeController 创建卡类型控制器。
func NewCardTypeController(app *service.App) *CardTypeController {
	return &CardTypeController{app: app}
}

// List 卡类型列表，支持按缩写、名称筛选。
func (m *CardTypeController) List(c *gin.Context) {
	list, err := m.app.ListCardTypes(service.CardTypeListQuery{
		Field:   c.Query("field"),
		Keyword: c.Query("keyword"),
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, list)
}

// Brands 卡名称下拉，配置来自 card_brands.json。
func (m *CardTypeController) Brands(c *gin.Context) {
	list, err := service.ListCardBrands()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, list)
}

// Create 新建卡类型。
func (m *CardTypeController) Create(c *gin.Context) {
	var req service.CardTypeSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.CreateCardType(req, operatorName(c))
	if err != nil {
		writeCardTypeError(c, err)
		return
	}
	response.SuccessMsg(c, item, "created")
}

// Update 编辑卡类型。
func (m *CardTypeController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, "无效的卡类型ID")
		return
	}
	var req service.CardTypeSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.UpdateCardType(uint(id), req, operatorName(c))
	if err != nil {
		writeCardTypeError(c, err)
		return
	}
	response.SuccessMsg(c, item, "updated")
}

func writeCardTypeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCardTypeNotFound):
		response.Fail(c, "卡类型不存在")
	case errors.Is(err, service.ErrCardTypeCodeExists):
		response.Fail(c, "缩写已存在")
	case errors.Is(err, service.ErrCardTypeCodeInvalid):
		response.Fail(c, "请输入合法缩写，仅字母和数字，最多 32 位")
	case errors.Is(err, service.ErrCardTypeNameInvalid):
		response.Fail(c, "请选择卡名称")
	case errors.Is(err, service.ErrCardTypeLengthInvalid):
		response.Fail(c, "卡号长度只能是 13 到 19")
	case errors.Is(err, service.ErrCardTypePrefixInvalid):
		response.Fail(c, "请填写有效卡头规则")
	default:
		response.Fail(c, err.Error())
	}
}
