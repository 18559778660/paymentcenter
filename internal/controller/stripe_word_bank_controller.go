package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// StripeWordBankController 控制层：Stripe 单词库。
type StripeWordBankController struct {
	app *service.App
}

// NewStripeWordBankController 创建 Stripe 单词库控制器。
func NewStripeWordBankController(app *service.App) *StripeWordBankController {
	return &StripeWordBankController{app: app}
}

// List Stripe 单词库列表。
func (m *StripeWordBankController) List(c *gin.Context) {
	list, err := m.app.ListStripeWordBanks(service.StripeWordBankListQuery{
		ConfigItem: c.Query("configItem"),
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, list)
}

// Create 新增 Stripe 单词。
func (m *StripeWordBankController) Create(c *gin.Context) {
	var req service.CreateStripeWordBankRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.CreateStripeWordBank(req, operatorName(c))
	if err != nil {
		writeStripeWordBankError(c, err)
		return
	}
	response.SuccessMsg(c, item, "created")
}

// Update 编辑 Stripe 单词。
func (m *StripeWordBankController) Update(c *gin.Context) {
	id, err := parseStripeWordBankID(c)
	if err != nil {
		response.Fail(c, "无效的ID")
		return
	}
	var req service.UpdateStripeWordBankRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.UpdateStripeWordBank(id, req, operatorName(c))
	if err != nil {
		writeStripeWordBankError(c, err)
		return
	}
	response.SuccessMsg(c, item, "updated")
}

// Delete 删除 Stripe 单词。
func (m *StripeWordBankController) Delete(c *gin.Context) {
	id, err := parseStripeWordBankID(c)
	if err != nil {
		response.Fail(c, "无效的ID")
		return
	}
	if err := m.app.DeleteStripeWordBank(id); err != nil {
		writeStripeWordBankError(c, err)
		return
	}
	response.SuccessMsg(c, nil, "deleted")
}

func parseStripeWordBankID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, err
	}
	return uint(id), nil
}

func writeStripeWordBankError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrStripeWordBankNotFound):
		response.Fail(c, "路径不存在")
	case errors.Is(err, service.ErrStripeWordBankNameExists):
		response.Fail(c, "路径名称已存在")
	case errors.Is(err, service.ErrStripeWordBankNameInvalid):
		response.Fail(c, "名称无效")
	case errors.Is(err, service.ErrStripeWordBankNameMustStartWithSlash):
		response.Fail(c, "路径类名称需以 / 开头")
	case errors.Is(err, service.ErrStripeWordBankNameMustNotStartWithSlash):
		response.Fail(c, "目录类名称不能以 / 开头")
	case errors.Is(err, service.ErrStripeWordBankConfigItemInvalid):
		response.Fail(c, "配置项无效")
	default:
		response.Fail(c, err.Error())
	}
}
