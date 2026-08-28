package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// AssignUserController 控制层：账号分配（分配子账号）。
type AssignUserController struct {
	app *service.App
}

// NewAssignUserController 创建账号分配控制器。
func NewAssignUserController(app *service.App) *AssignUserController {
	return &AssignUserController{app: app}
}

// List 分配子账号列表。
func (m *AssignUserController) List(c *gin.Context) {
	list, err := m.app.ListAssignUsers(service.AssignUserListQuery{
		Field:   c.Query("field"),
		Keyword: c.Query("keyword"),
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, list)
}

// Create 新增分配子账号。
func (m *AssignUserController) Create(c *gin.Context) {
	var req service.CreateAssignUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.CreateAssignUser(req)
	if err != nil {
		writeAssignUserError(c, err)
		return
	}
	response.SuccessMsg(c, item, "created")
}

// ListAccounts 子账号下的通道账号列表（默认全部账号，标记分配状态）。
func (m *AssignUserController) ListAccounts(c *gin.Context) {
	id, err := parseAssignUserID(c)
	if err != nil {
		response.Fail(c, "无效的用户ID")
		return
	}
	list, err := m.app.ListAssignUserAccounts(id, service.AssignUserAccountListQuery{
		ChannelID: parseOptionalUintQuery(c, "channelId"),
	})
	if err != nil {
		writeAssignUserAccountError(c, err)
		return
	}
	response.Success(c, list)
}

// SetAccountAssignment 设置通道账号是否分配给子账号。
func (m *AssignUserController) SetAccountAssignment(c *gin.Context) {
	userID, err := parseAssignUserID(c)
	if err != nil {
		response.Fail(c, "无效的用户ID")
		return
	}
	accountID, err := strconv.ParseUint(c.Param("accountId"), 10, 64)
	if err != nil || accountID == 0 {
		response.Fail(c, "无效的账号ID")
		return
	}
	var req struct {
		Assigned bool `json:"assigned"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	if err := m.app.SetAssignUserAccountAssignment(userID, uint(accountID), req.Assigned, operatorName(c)); err != nil {
		writeAssignUserAccountError(c, err)
		return
	}
	response.SuccessMsg(c, nil, "updated")
}

func parseAssignUserID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, err
	}
	return uint(id), nil
}

func writeAssignUserError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrAssignUserUsernameExists):
		response.Fail(c, "用户名已存在")
	case errors.Is(err, service.ErrAssignUserUsernameInvalid):
		response.Fail(c, "用户名可由英文字母、数字、连字符组成")
	case errors.Is(err, service.ErrAssignUserPasswordInvalid):
		response.Fail(c, "密码需为 6-20 位")
	default:
		response.Fail(c, err.Error())
	}
}

func writeAssignUserAccountError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrAssignUserNotFound):
		response.Fail(c, "分配账号不存在")
	case errors.Is(err, service.ErrChannelAccountNotFound):
		response.Fail(c, "通道账号不存在")
	case errors.Is(err, service.ErrChannelAccountAlreadyAssigned):
		response.Fail(c, "该通道账号已分配给其他子账号")
	default:
		response.Fail(c, err.Error())
	}
}
