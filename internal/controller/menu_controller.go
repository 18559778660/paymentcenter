package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/service"
	"paymentcenter/internal/store"
	"paymentcenter/internal/util/response"
)

// MenuController 控制层：菜单管理 CRUD。
type MenuController struct {
	app *service.App
}

// NewMenuController 创建菜单管理控制器。
func NewMenuController(app *service.App) *MenuController {
	return &MenuController{app: app}
}

// List 返回全部菜单树。
func (m *MenuController) List(c *gin.Context) {
	menus, err := m.app.ListAdminMenus()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, menus)
}

// Create 新增菜单。
func (m *MenuController) Create(c *gin.Context) {
	var req service.SaveMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	menu, err := m.app.CreateAdminMenu(req)
	if err != nil {
		m.writeSaveError(c, err)
		return
	}
	response.SuccessMsg(c, menu, "created")
}

// Update 更新菜单。
func (m *MenuController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, "无效的菜单ID")
		return
	}
	var req service.SaveMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	menu, err := m.app.UpdateAdminMenu(uint(id), req)
	if err != nil {
		m.writeSaveError(c, err)
		return
	}
	response.SuccessMsg(c, menu, "updated")
}

// Delete 删除菜单。
func (m *MenuController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, "无效的菜单ID")
		return
	}
	if err := m.app.DeleteAdminMenu(uint(id)); err != nil {
		if errors.Is(err, service.ErrMenuHasChildren) {
			response.Fail(c, "请先删除子菜单")
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			response.Fail(c, "菜单不存在")
			return
		}
		response.Fail(c, err.Error())
		return
	}
	response.SuccessMsg(c, nil, "deleted")
}

// NameExists 检查路由 Name 是否已存在。
func (m *MenuController) NameExists(c *gin.Context) {
	name := c.Query("name")
	excludeID, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	exists, err := m.app.IsMenuNameExists(name, uint(excludeID))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, exists)
}

// PathExists 检查路径是否已存在。
func (m *MenuController) PathExists(c *gin.Context) {
	path := c.Query("path")
	excludeID, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	exists, err := m.app.IsMenuPathExists(path, uint(excludeID))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, exists)
}

// writeSaveError 写入保存错误。
func (m *MenuController) writeSaveError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrMenuNameExists):
		response.Fail(c, "路由名称已存在")
	case errors.Is(err, service.ErrMenuPathExists):
		response.Fail(c, "路由路径已存在")
	case errors.Is(err, store.ErrNotFound):
		response.Fail(c, "菜单不存在")
	default:
		response.Fail(c, err.Error())
	}
}
