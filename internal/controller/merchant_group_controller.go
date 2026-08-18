package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/middleware"
	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// MerchantGroupController 控制层：商户分组 CRUD。
type MerchantGroupController struct {
	app *service.App
}

// NewMerchantGroupController 创建商户分组控制器。
func NewMerchantGroupController(app *service.App) *MerchantGroupController {
	return &MerchantGroupController{app: app}
}

// List 分组列表，支持按 ID、标题筛选。
func (m *MerchantGroupController) List(c *gin.Context) {
	q := service.MerchantGroupListQuery{
		Name: c.Query("name"),
	}
	if v := c.Query("id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, "无效的分组ID")
			return
		}
		gid := uint(id)
		q.ID = &gid
	}
	list, err := m.app.ListMerchantGroups(q)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, list)
}

// Create 新建分组。
func (m *MerchantGroupController) Create(c *gin.Context) {
	var req service.MerchantGroupSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.CreateMerchantGroup(req, operatorName(c))
	if err != nil {
		writeMerchantGroupError(c, err)
		return
	}
	response.SuccessMsg(c, item, "created")
}

// Update 编辑分组。
func (m *MerchantGroupController) Update(c *gin.Context) {
	id, ok := parseGroupID(c)
	if !ok {
		return
	}
	var req service.MerchantGroupSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.UpdateMerchantGroup(id, req, operatorName(c))
	if err != nil {
		writeMerchantGroupError(c, err)
		return
	}
	response.SuccessMsg(c, item, "updated")
}

// Delete 删除分组。
func (m *MerchantGroupController) Delete(c *gin.Context) {
	id, ok := parseGroupID(c)
	if !ok {
		return
	}
	if err := m.app.DeleteMerchantGroup(id); err != nil {
		writeMerchantGroupError(c, err)
		return
	}
	response.Success(c, nil)
}

func parseGroupID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, "无效的分组ID")
		return 0, false
	}
	return uint(id), true
}

func operatorName(c *gin.Context) string {
	if user, ok := middleware.CurrentUser(c); ok {
		return user.Username
	}
	return "system"
}

func writeMerchantGroupError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrMerchantGroupNotFound):
		response.Fail(c, "分组不存在")
	case errors.Is(err, service.ErrMerchantGroupNameExists):
		response.Fail(c, "分组名已存在")
	case errors.Is(err, service.ErrMerchantGroupNameInvalid):
		response.Fail(c, "请输入分组名，最多 64 个字")
	case errors.Is(err, service.ErrMerchantGroupMerchantInvalid):
		response.Fail(c, "所选商户不存在")
	default:
		response.Fail(c, err.Error())
	}
}
