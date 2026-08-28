package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// ChannelGroupController 控制层：通道分组。
type ChannelGroupController struct {
	app *service.App
}

// NewChannelGroupController 创建通道分组控制器。
func NewChannelGroupController(app *service.App) *ChannelGroupController {
	return &ChannelGroupController{app: app}
}

// List 通道分组列表。
func (m *ChannelGroupController) List(c *gin.Context) {
	q := service.ChannelGroupListQuery{
		Code: c.Query("code"),
	}
	if v := c.Query("id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			uid := uint(id)
			q.ID = &uid
		}
	}
	list, err := m.app.ListChannelGroups(q)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, list)
}

// Create 新增通道分组。
func (m *ChannelGroupController) Create(c *gin.Context) {
	var req service.CreateChannelGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.CreateChannelGroup(req, operatorName(c))
	if err != nil {
		writeChannelGroupError(c, err)
		return
	}
	response.SuccessMsg(c, item, "created")
}

// Update 编辑通道分组。
func (m *ChannelGroupController) Update(c *gin.Context) {
	id, err := parseChannelGroupID(c)
	if err != nil {
		response.Fail(c, "无效的分组ID")
		return
	}
	var req service.UpdateChannelGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.UpdateChannelGroup(id, req, operatorName(c))
	if err != nil {
		writeChannelGroupError(c, err)
		return
	}
	response.SuccessMsg(c, item, "updated")
}

func parseChannelGroupID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, err
	}
	return uint(id), nil
}

func writeChannelGroupError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrChannelGroupNotFound):
		response.Fail(c, "分组不存在")
	case errors.Is(err, service.ErrChannelGroupCodeExists):
		response.Fail(c, "分组CODE已存在")
	case errors.Is(err, service.ErrChannelGroupNameExists):
		response.Fail(c, "分组名已存在")
	case errors.Is(err, service.ErrChannelGroupCodeInvalid):
		response.Fail(c, "请输入分组CODE")
	case errors.Is(err, service.ErrChannelGroupNameInvalid):
		response.Fail(c, "请输入分组名")
	case errors.Is(err, service.ErrChannelGroupInterceptRangeInvalid):
		response.Fail(c, "限制最小金额不能高于限制最大金额")
	case errors.Is(err, service.ErrChannelGroupSuccessSettingInvalid):
		response.Fail(c, "成功设置需同时配置支付频率，以及成功次数或失败次数")
	default:
		response.Fail(c, err.Error())
	}
}
