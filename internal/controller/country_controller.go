package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// CountryController 控制层：国家列表 CRUD。
type CountryController struct {
	app *service.App
}

// NewCountryController 创建国家控制器。
func NewCountryController(app *service.App) *CountryController {
	return &CountryController{app: app}
}

// List 国家列表，支持按名称、CODE 筛选。
func (m *CountryController) List(c *gin.Context) {
	list, err := m.app.ListCountries(service.CountryListQuery{
		Field:   c.Query("field"),
		Keyword: c.Query("keyword"),
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, list)
}

// Options 国家下拉，配置来自 countries.json。
func (m *CountryController) Options(c *gin.Context) {
	list, err := service.ListCountryOptions()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, list)
}

// Create 新建国家。
func (m *CountryController) Create(c *gin.Context) {
	var req service.CountrySaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.CreateCountry(req, operatorName(c))
	if err != nil {
		writeCountryError(c, err)
		return
	}
	response.SuccessMsg(c, item, "created")
}

// Update 编辑大卡头占比。
func (m *CountryController) Update(c *gin.Context) {
	id, err := parseCountryID(c)
	if err != nil {
		response.Fail(c, "无效的国家ID")
		return
	}
	var req service.CountrySaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.UpdateCountry(id, req, operatorName(c))
	if err != nil {
		writeCountryError(c, err)
		return
	}
	response.SuccessMsg(c, item, "updated")
}

// Delete 删除国家。
func (m *CountryController) Delete(c *gin.Context) {
	id, err := parseCountryID(c)
	if err != nil {
		response.Fail(c, "无效的国家ID")
		return
	}
	if err := m.app.DeleteCountry(id); err != nil {
		writeCountryError(c, err)
		return
	}
	response.SuccessMsg(c, nil, "deleted")
}

func parseCountryID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid country id")
	}
	return uint(id), nil
}

func writeCountryError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCountryNotFound):
		response.Fail(c, "国家不存在")
	case errors.Is(err, service.ErrCountryCodeExists):
		response.Fail(c, "该国家已存在")
	case errors.Is(err, service.ErrCountryCodeInvalid):
		response.Fail(c, "请选择标准国家")
	case errors.Is(err, service.ErrCountryCardBinInvalid):
		response.Fail(c, "请输入有效的大卡头占比")
	default:
		response.Fail(c, err.Error())
	}
}
