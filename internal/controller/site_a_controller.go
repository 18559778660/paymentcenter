package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// SiteAController 控制层：A 站管理。
type SiteAController struct {
	app *service.App
}

// NewSiteAController 创建 A 站控制器。
func NewSiteAController(app *service.App) *SiteAController {
	return &SiteAController{app: app}
}

// List A 站列表。
func (m *SiteAController) List(c *gin.Context) {
	q := service.SiteAListQuery{
		Domain: c.Query("domain"),
		Status: c.Query("status"),
	}
	if v := c.Query("merchantId"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			merchantID := uint(id)
			q.MerchantID = &merchantID
		}
	}
	list, err := m.app.ListSiteAs(q)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, list)
}

// Create 新增 A 站。
func (m *SiteAController) Create(c *gin.Context) {
	var req service.CreateSiteARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.CreateSiteA(req, operatorName(c))
	if err != nil {
		writeSiteAError(c, err)
		return
	}
	response.SuccessMsg(c, item, "created")
}

// BatchStatus 批量更新状态。
func (m *SiteAController) BatchStatus(c *gin.Context) {
	var req service.BatchUpdateSiteAStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	count, err := m.app.BatchUpdateSiteAStatus(req, operatorName(c))
	if err != nil {
		writeSiteAError(c, err)
		return
	}
	response.SuccessMsg(c, gin.H{"count": count}, "updated")
}

func writeSiteAError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrSiteANotFound):
		response.Fail(c, "A站不存在")
	case errors.Is(err, service.ErrSiteADomainExists):
		response.Fail(c, "域名已存在")
	case errors.Is(err, service.ErrSiteADomainInvalid):
		response.Fail(c, "请输入有效域名")
	case errors.Is(err, service.ErrSiteAFrameworkInvalid):
		response.Fail(c, "请选择有效框架")
	case errors.Is(err, service.ErrSiteAStatusInvalid):
		response.Fail(c, "状态无效")
	case errors.Is(err, service.ErrSiteAMerchantInvalid):
		response.Fail(c, "商户不存在")
	case errors.Is(err, service.ErrSiteABatchEmpty):
		response.Fail(c, "请选择记录")
	default:
		response.Fail(c, err.Error())
	}
}
