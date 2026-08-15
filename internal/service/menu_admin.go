package service

import (
	"errors"
	"fmt"

	"paymentcenter/internal/model"
)

var (
	ErrMenuHasChildren = errors.New("menu has children")
	ErrMenuNameExists  = errors.New("menu name exists")
	ErrMenuPathExists  = errors.New("menu path exists")
)

// AdminMenuNode 菜单管理页用的树节点。
type AdminMenuNode struct {
	ID        uint                   `json:"id"`
	ParentID  uint                   `json:"parentId"`
	Name      string                 `json:"name"`
	Title     string                 `json:"title"`
	Path      string                 `json:"path"`
	Component string                 `json:"component"`
	Icon      string                 `json:"icon"`
	AuthCode  string                 `json:"authCode"`
	Type      int                    `json:"type"`
	Sort      int                    `json:"sort"`
	Status    int                    `json:"status"`
	Meta      map[string]interface{} `json:"meta"`
	Children  []AdminMenuNode        `json:"children,omitempty"`
}

// SaveMenuRequest 创建/更新菜单入参。
type SaveMenuRequest struct {
	ParentID  uint                   `json:"parentId"`
	Name      string                 `json:"name" binding:"required"`
	Title     string                 `json:"title" binding:"required"`
	Path      string                 `json:"path"`
	Component string                 `json:"component"`
	Icon      string                 `json:"icon"`
	AuthCode  string                 `json:"authCode"`
	Type      int                    `json:"type"`
	Sort      int                    `json:"sort"`
	Status    *int                   `json:"status"`
	Meta      map[string]interface{} `json:"meta"`
}

// ListAdminMenus 返回全部菜单树，给菜单管理页。
func (a *App) ListAdminMenus() ([]AdminMenuNode, error) {
	menus, err := a.store.ListAllMenus()
	if err != nil {
		return nil, err
	}
	byParent := map[uint][]model.Menu{}
	for _, menu := range menus {
		byParent[menu.ParentID] = append(byParent[menu.ParentID], menu)
	}
	return buildAdminMenus(0, byParent), nil
}

// buildAdminMenus 构建菜单树。
func buildAdminMenus(parentID uint, byParent map[uint][]model.Menu) []AdminMenuNode {
	items := byParent[parentID]
	result := make([]AdminMenuNode, 0, len(items))
	for _, menu := range items {
		meta := map[string]interface{}{}
		for k, v := range menu.Meta {
			meta[k] = v
		}
		node := AdminMenuNode{
			ID:        menu.ID,
			ParentID:  menu.ParentID,
			Name:      menu.Name,
			Title:     menu.Title,
			Path:      menu.Path,
			Component: menu.Component,
			Icon:      menu.Icon,
			AuthCode:  menu.AuthCode,
			Type:      menu.Type,
			Sort:      menu.Sort,
			Status:    menu.Status,
			Meta:      meta,
			Children:  buildAdminMenus(menu.ID, byParent),
		}
		result = append(result, node)
	}
	return result
}

// CreateAdminMenu 新增菜单，并默认绑给 super / admin。
func (a *App) CreateAdminMenu(req SaveMenuRequest) (*AdminMenuNode, error) {
	if err := a.validateMenuRequest(req, 0); err != nil {
		return nil, err
	}
	status := model.MenuStatusEnabled
	if req.Status != nil {
		status = *req.Status
	}
	menu := &model.Menu{
		ParentID:  req.ParentID,
		Name:      req.Name,
		Title:     req.Title,
		Path:      req.Path,
		Component: req.Component,
		Icon:      req.Icon,
		AuthCode:  req.AuthCode,
		Type:      req.Type,
		Sort:      req.Sort,
		Status:    status,
		Meta:      model.JSONMap(req.Meta),
	}
	if err := a.store.CreateMenu(menu); err != nil {
		return nil, err
	}
	if err := a.bindMenuToDefaultRoles(menu.ID); err != nil {
		return nil, err
	}
	return toAdminMenuNode(menu), nil
}

// UpdateAdminMenu 更新菜单。
func (a *App) UpdateAdminMenu(id uint, req SaveMenuRequest) (*AdminMenuNode, error) {
	menu, err := a.store.GetMenuByID(id)
	if err != nil {
		return nil, err
	}
	if err := a.validateMenuRequest(req, id); err != nil {
		return nil, err
	}
	if req.ParentID == id {
		return nil, fmt.Errorf("parent cannot be self")
	}
	menu.ParentID = req.ParentID
	menu.Name = req.Name
	menu.Title = req.Title
	menu.Path = req.Path
	menu.Component = req.Component
	menu.Icon = req.Icon
	menu.AuthCode = req.AuthCode
	menu.Type = req.Type
	menu.Sort = req.Sort
	if req.Status != nil {
		menu.Status = *req.Status
	}
	menu.Meta = model.JSONMap(req.Meta)
	if err := a.store.SaveMenu(menu); err != nil {
		return nil, err
	}
	return toAdminMenuNode(menu), nil
}

// DeleteAdminMenu 删除菜单；有子菜单时拒绝。
func (a *App) DeleteAdminMenu(id uint) error {
	if _, err := a.store.GetMenuByID(id); err != nil {
		return err
	}
	count, err := a.store.CountMenusByParentID(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrMenuHasChildren
	}
	if err := a.store.DeleteRoleMenusByMenuID(id); err != nil {
		return err
	}
	return a.store.DeleteMenu(id)
}

// IsMenuNameExists 菜单 Name 是否已存在。
func (a *App) IsMenuNameExists(name string, excludeID uint) (bool, error) {
	return a.store.MenuNameExists(name, excludeID)
}

// IsMenuPathExists 菜单 Path 是否已存在。
func (a *App) IsMenuPathExists(path string, excludeID uint) (bool, error) {
	return a.store.MenuPathExists(path, excludeID)
}

// validateMenuRequest 验证菜单请求。
func (a *App) validateMenuRequest(req SaveMenuRequest, excludeID uint) error {
	if req.Type < model.MenuTypeDir || req.Type > model.MenuTypeButton {
		return fmt.Errorf("invalid menu type")
	}
	if req.ParentID > 0 {
		if _, err := a.store.GetMenuByID(req.ParentID); err != nil {
			return fmt.Errorf("parent menu not found")
		}
	}
	exists, err := a.store.MenuNameExists(req.Name, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return ErrMenuNameExists
	}
	exists, err = a.store.MenuPathExists(req.Path, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return ErrMenuPathExists
	}
	return nil
}

// bindMenuToDefaultRoles 将菜单绑定给 super / admin。
func (a *App) bindMenuToDefaultRoles(menuID uint) error {
	for _, code := range []string{"super", "admin"} {
		role, err := a.store.FindRoleByCode(code)
		if err != nil {
			if isNotFound(err) {
				continue
			}
			return err
		}
		if err := a.store.EnsureRoleMenu(role.ID, menuID); err != nil {
			return err
		}
	}
	return nil
}

// toAdminMenuNode 转换为菜单管理页用的树节点。
func toAdminMenuNode(menu *model.Menu) *AdminMenuNode {
	meta := map[string]interface{}{}
	for k, v := range menu.Meta {
		meta[k] = v
	}
	return &AdminMenuNode{
		ID:        menu.ID,
		ParentID:  menu.ParentID,
		Name:      menu.Name,
		Title:     menu.Title,
		Path:      menu.Path,
		Component: menu.Component,
		Icon:      menu.Icon,
		AuthCode:  menu.AuthCode,
		Type:      menu.Type,
		Sort:      menu.Sort,
		Status:    menu.Status,
		Meta:      meta,
	}
}
