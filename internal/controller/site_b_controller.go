package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// SiteBController 控制层：B 站管理。
type SiteBController struct {
	app *service.App
}

// NewSiteBController 创建 B 站控制器。
func NewSiteBController(app *service.App) *SiteBController {
	return &SiteBController{app: app}
}

// List B 站列表。
func (m *SiteBController) List(c *gin.Context) {
	q := service.SiteBListQuery{
		Domain: c.Query("domain"),
		Remark: c.Query("remark"),
	}
	if v := c.Query("platformId"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			platformID := uint(id)
			q.PlatformID = &platformID
		}
	}
	if v := c.Query("id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			uid := uint(id)
			q.ID = &uid
		}
	}
	if v := c.Query("status"); v != "" {
		status := v == "1" || v == "true"
		q.Status = &status
	}
	list, err := m.app.ListSiteBs(q)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, list)
}

// Create 新增 B 站。
func (m *SiteBController) Create(c *gin.Context) {
	var req service.CreateSiteBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.CreateSiteB(req, operatorName(c))
	if err != nil {
		writeSiteBError(c, err)
		return
	}
	response.SuccessMsg(c, item, "created")
}

// Update 编辑 B 站 FTP 信息。
func (m *SiteBController) Update(c *gin.Context) {
	id, err := parseSiteBID(c)
	if err != nil {
		response.Fail(c, "无效的B站ID")
		return
	}
	var req service.UpdateSiteBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.UpdateSiteB(id, req, operatorName(c))
	if err != nil {
		writeSiteBError(c, err)
		return
	}
	response.SuccessMsg(c, item, "updated")
}

// SetStatus 更新 B 站状态。
func (m *SiteBController) SetStatus(c *gin.Context) {
	id, err := parseSiteBID(c)
	if err != nil {
		response.Fail(c, "无效的B站ID")
		return
	}
	var req service.SetSiteBStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.SetSiteBStatus(id, req, operatorName(c))
	if err != nil {
		writeSiteBError(c, err)
		return
	}
	response.SuccessMsg(c, item, "updated")
}

// Gateways B 站网关列表。
func (m *SiteBController) Gateways(c *gin.Context) {
	id, err := parseSiteBID(c)
	if err != nil {
		response.Fail(c, "无效的B站ID")
		return
	}
	list, err := m.app.ListSiteBGateways(id)
	if err != nil {
		writeSiteBError(c, err)
		return
	}
	response.Success(c, list)
}

func parseSiteBID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, err
	}
	return uint(id), nil
}

func writeSiteBError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrSiteBNotFound):
		response.Fail(c, "B站不存在")
	case errors.Is(err, service.ErrSiteBDomainExists):
		response.Fail(c, "域名已存在")
	case errors.Is(err, service.ErrSiteBDomainInvalid):
		response.Fail(c, "请输入有效域名")
	case errors.Is(err, service.ErrSiteBPlatformInvalid):
		response.Fail(c, "请选择有效通道平台")
	case errors.Is(err, service.ErrSiteBFrameworkInvalid):
		response.Fail(c, "请选择有效框架")
	default:
		response.Fail(c, err.Error())
	}
}
