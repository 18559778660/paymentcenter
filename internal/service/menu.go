package service

import (
	"strings"

	"paymentcenter/internal/model"
)

// VbenMenu 前端动态路由格式，对应 GET /menu/all 的 data。
type VbenMenu struct {
	Name      string                 `json:"name"`
	Path      string                 `json:"path"`
	Component string                 `json:"component,omitempty"`
	Redirect  string                 `json:"redirect,omitempty"`
	Meta      map[string]interface{} `json:"meta"`
	Children  []VbenMenu             `json:"children,omitempty"`
}

// GetUserMenus 把当前用户有权限的菜单转成 Vben 路由树。按钮类型不进侧边栏。
func (a *App) GetUserMenus(userID uint) ([]VbenMenu, error) {
	menus, err := a.store.ListMenusByUserID(userID)
	if err != nil {
		return nil, err
	}

	byParent := map[uint][]model.Menu{}
	for _, menu := range menus {
		if menu.Type == model.MenuTypeButton {
			continue
		}
		byParent[menu.ParentID] = append(byParent[menu.ParentID], menu)
	}
	return buildVbenMenus(0, "", byParent), nil
}

func buildVbenMenus(parentID uint, parentPath string, byParent map[uint][]model.Menu) []VbenMenu {
	items := byParent[parentID]
	result := make([]VbenMenu, 0, len(items))
	for _, menu := range items {
		node := VbenMenu{
			Name: menu.Name,
			Path: menuRoutePath(parentPath, menu.Path),
			Meta: map[string]interface{}{
				"icon":  menu.Icon,
				"order": menu.Sort,
				"title": menu.Title,
			},
		}
		if menu.Name == "Analytics" {
			node.Meta["affixTab"] = true
			node.Meta["hideInMenu"] = true
		}
		if menu.Name == "Dashboard" {
			node.Meta["hideChildrenInMenu"] = true
		}

		children := buildVbenMenus(menu.ID, menu.Path, byParent)
		if menu.Type == model.MenuTypeDir {
			node.Children = children
			if len(children) > 0 {
				node.Redirect = firstLeafPath(menu.Path, children)
			}
		} else {
			node.Component = menu.Component
			if len(children) > 0 {
				node.Children = children
			}
		}
		result = append(result, node)
	}
	return result
}

func menuRoutePath(parentPath, fullPath string) string {
	if parentPath == "" {
		return fullPath
	}
	prefix := strings.TrimSuffix(parentPath, "/") + "/"
	if strings.HasPrefix(fullPath, prefix) {
		return strings.TrimPrefix(fullPath, prefix)
	}
	return fullPath
}

func firstLeafPath(parentFullPath string, children []VbenMenu) string {
	if len(children) == 0 {
		return parentFullPath
	}
	child := children[0]
	joined := strings.TrimSuffix(parentFullPath, "/") + "/" + strings.TrimPrefix(child.Path, "/")
	if len(child.Children) > 0 {
		return firstLeafPath(joined, child.Children)
	}
	if strings.HasPrefix(child.Path, "/") {
		return child.Path
	}
	return joined
}
