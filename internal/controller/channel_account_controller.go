package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// ChannelAccountController 控制层：通道账号。
type ChannelAccountController struct {
	app *service.App
}

// NewChannelAccountController 创建通道账号控制器。
func NewChannelAccountController(app *service.App) *ChannelAccountController {
	return &ChannelAccountController{app: app}
}

// List 通道账号列表。
func (m *ChannelAccountController) List(c *gin.Context) {
	q := service.ChannelAccountListQuery{
		ChannelName:  c.Query("channelName"),
		Alias:        c.Query("alias"),
		Remark:       c.Query("remark"),
		CreatedFrom:  c.Query("createdFrom"),
		CreatedTo:    c.Query("createdTo"),
		GroupName:    c.Query("groupName"),
		AssignedUser: c.Query("assignedUser"),
		ListFilter:   c.Query("listFilter"),
	}
	if v := c.Query("channelId"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			channelID := uint(id)
			q.ChannelID = &channelID
		}
	}
	if v := c.Query("id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			uid := uint(id)
			q.ID = &uid
		}
	}
	list, err := m.app.ListChannelAccounts(q)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, list)
}

// Create 新增通道账号。
func (m *ChannelAccountController) Create(c *gin.Context) {
	var req service.CreateChannelAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.CreateChannelAccount(req, operatorName(c))
	if err != nil {
		writeChannelAccountError(c, err)
		return
	}
	response.SuccessMsg(c, item, "created")
}

// Update 编辑通道账号。
func (m *ChannelAccountController) Update(c *gin.Context) {
	id, err := parseChannelAccountID(c)
	if err != nil {
		response.Fail(c, "无效的账号ID")
		return
	}
	var req service.UpdateChannelAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.UpdateChannelAccount(id, req, operatorName(c))
	if err != nil {
		writeChannelAccountError(c, err)
		return
	}
	response.SuccessMsg(c, item, "updated")
}

// UpdateLimits 更新限制配置。
func (m *ChannelAccountController) UpdateLimits(c *gin.Context) {
	id, err := parseChannelAccountID(c)
	if err != nil {
		response.Fail(c, "无效的账号ID")
		return
	}
	var req service.UpdateChannelAccountLimitsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.UpdateChannelAccountLimits(id, req, operatorName(c))
	if err != nil {
		writeChannelAccountError(c, err)
		return
	}
	response.SuccessMsg(c, item, "updated")
}

// SetStatus 启用/禁用通道账号。
func (m *ChannelAccountController) SetStatus(c *gin.Context) {
	id, err := parseChannelAccountID(c)
	if err != nil {
		response.Fail(c, "无效的账号ID")
		return
	}
	var req struct {
		Status bool `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.SetChannelAccountStatus(id, req.Status, operatorName(c))
	if err != nil {
		writeChannelAccountError(c, err)
		return
	}
	response.Success(c, item)
}

func parseChannelAccountID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, err
	}
	return uint(id), nil
}

func writeChannelAccountError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrChannelAccountNotFound):
		response.Fail(c, "通道账号不存在")
	case errors.Is(err, service.ErrChannelAccountNoInvalid):
		response.Fail(c, "请输入通道账号")
	case errors.Is(err, service.ErrChannelAccountChannelInvalid):
		response.Fail(c, "请选择有效通道")
	case errors.Is(err, service.ErrChannelAccountSiteBInvalid):
		response.Fail(c, "请选择有效B站")
	case errors.Is(err, service.ErrChannelAccountChannelSiteBExists):
		response.Fail(c, "该通道与B站已存在绑定账号")
	case errors.Is(err, service.ErrChannelAccountSuccessSettingInvalid):
		response.Fail(c, "成功设置需同时配置支付频率和指定时间内限制成功次数")
	case errors.Is(err, service.ErrChannelInterceptRangeInvalid):
		response.Fail(c, "限制最小金额不能高于限制最大金额")
	default:
		response.Fail(c, err.Error())
	}
}
