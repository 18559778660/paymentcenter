package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// ChannelController 控制层：通道列表。
type ChannelController struct {
	app *service.App
}

// NewChannelController 创建通道控制器。
func NewChannelController(app *service.App) *ChannelController {
	return &ChannelController{app: app}
}

// List 通道列表。
func (m *ChannelController) List(c *gin.Context) {
	q := service.ChannelListQuery{
		Name: c.Query("name"),
	}
	if v := c.Query("id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			uid := uint(id)
			q.ID = &uid
		}
	}
	list, err := m.app.ListChannels(q)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, list)
}

// Create 新增通道。
func (m *ChannelController) Create(c *gin.Context) {
	var req service.CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.CreateChannel(req, operatorName(c))
	if err != nil {
		writeChannelError(c, err)
		return
	}
	response.SuccessMsg(c, item, "created")
}

// Update 编辑通道信息。
func (m *ChannelController) Update(c *gin.Context) {
	id, err := parseChannelID(c)
	if err != nil {
		response.Fail(c, "无效的通道ID")
		return
	}
	var req service.UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.UpdateChannel(id, req, operatorName(c))
	if err != nil {
		writeChannelError(c, err)
		return
	}
	response.SuccessMsg(c, item, "updated")
}

// UpdateLimits 更新限制配置。
func (m *ChannelController) UpdateLimits(c *gin.Context) {
	id, err := parseChannelID(c)
	if err != nil {
		response.Fail(c, "无效的通道ID")
		return
	}
	var req service.UpdateChannelLimitsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.UpdateChannelLimits(id, req, operatorName(c))
	if err != nil {
		writeChannelError(c, err)
		return
	}
	response.SuccessMsg(c, item, "updated")
}

// SetStatus 启用/禁用通道。
func (m *ChannelController) SetStatus(c *gin.Context) {
	id, err := parseChannelID(c)
	if err != nil {
		response.Fail(c, "无效的通道ID")
		return
	}
	var req struct {
		Status bool `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.SetChannelStatus(id, req.Status, operatorName(c))
	if err != nil {
		writeChannelError(c, err)
		return
	}
	response.Success(c, item)
}

func parseChannelID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, err
	}
	return uint(id), nil
}

func writeChannelError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrChannelNotFound):
		response.Fail(c, "通道不存在")
	case errors.Is(err, service.ErrChannelNameExists):
		response.Fail(c, "通道名已存在")
	case errors.Is(err, service.ErrChannelNameInvalid):
		response.Fail(c, "请输入通道名")
	case errors.Is(err, service.ErrChannelInterceptRangeInvalid):
		response.Fail(c, "限制最小金额不能高于限制最大金额")
	case errors.Is(err, service.ErrChannelSuccessSettingInvalid):
		response.Fail(c, "成功设置需同时配置支付频率，以及成功次数或失败次数")
	default:
		response.Fail(c, err.Error())
	}
}
