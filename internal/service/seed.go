package service

import (
	"paymentcenter/internal/model"
)

// menuSeed 初始化菜单用的内存结构，启动时写入 menus 表。
type menuSeed struct {
	Name      string     // 路由 Name，需唯一
	Title     string     // 显示标题
	Path      string     // 前端路径
	Component string     // 前端组件路径
	Icon      string     // 图标
	AuthCode  string     // 权限码
	Type      int        // 0目录 1菜单 2按钮
	Sort      int        // 排序
	Children  []menuSeed // 子菜单
}

// SeedRBAC 启动时幂等初始化角色、菜单、默认管理员。已存在的数据不会覆盖。
func (a *App) SeedRBAC(adminUsername, adminPassword string) error {
	superRole, err := a.ensureRole("super", "超级管理员", "拥有全部权限")
	if err != nil {
		return err
	}
	adminRole, err := a.ensureRole("admin", "管理员", "日常运营权限")
	if err != nil {
		return err
	}

	menuIDs, err := a.ensureMenus(rbacMenuTree())
	if err != nil {
		return err
	}
	for _, menuID := range menuIDs {
		if err := a.store.EnsureRoleMenu(superRole.ID, menuID); err != nil {
			return err
		}
		if err := a.store.EnsureRoleMenu(adminRole.ID, menuID); err != nil {
			return err
		}
	}

	if adminUsername == "" || adminPassword == "" {
		return nil
	}
	return a.ensureAdminUser(adminUsername, adminPassword, superRole.ID)
}

// ensureRole 按编码查找角色，没有则创建。
func (a *App) ensureRole(code, name, remark string) (*model.Role, error) {
	role, err := a.store.FindRoleByCode(code)
	if err == nil {
		return role, nil
	}
	if !isNotFound(err) {
		return nil, err
	}
	role = &model.Role{
		Code:   code,
		Name:   name,
		Remark: remark,
		Status: model.RoleStatusEnabled,
	}
	if err := a.store.CreateRole(role); err != nil {
		return nil, err
	}
	return role, nil
}

// ensureMenus 按树写入菜单，已存在的 Name 跳过创建，返回全部菜单 ID。
func (a *App) ensureMenus(seeds []menuSeed) ([]uint, error) {
	ids := make([]uint, 0)
	var walk func(parentID uint, items []menuSeed) error
	walk = func(parentID uint, items []menuSeed) error {
		for _, item := range items {
			menu, err := a.store.FindMenuByName(item.Name)
			if err != nil && !isNotFound(err) {
				return err
			}
			if isNotFound(err) {
				menu = &model.Menu{
					ParentID:  parentID,
					Name:      item.Name,
					Title:     item.Title,
					Path:      item.Path,
					Component: item.Component,
					Icon:      item.Icon,
					AuthCode:  item.AuthCode,
					Type:      item.Type,
					Sort:      item.Sort,
					Status:    model.MenuStatusEnabled,
				}
				if err := a.store.CreateMenu(menu); err != nil {
					return err
				}
			} else {
				menu.ParentID = parentID
				menu.Title = item.Title
				menu.Path = item.Path
				menu.Component = item.Component
				menu.Icon = item.Icon
				menu.AuthCode = item.AuthCode
				menu.Type = item.Type
				menu.Sort = item.Sort
				menu.Status = model.MenuStatusEnabled
				if err := a.store.SaveMenu(menu); err != nil {
					return err
				}
			}
			ids = append(ids, menu.ID)
			if err := walk(menu.ID, item.Children); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(0, seeds); err != nil {
		return nil, err
	}
	return ids, nil
}

// ensureAdminUser 确保默认管理员存在，并绑定超级管理员角色。已有账号不改密码。
func (a *App) ensureAdminUser(username, password string, superRoleID uint) error {
	user, err := a.store.FindUserByUsername(username)
	if err != nil && !isNotFound(err) {
		return err
	}
	if isNotFound(err) {
		hash, hashErr := a.hashPassword(password)
		if hashErr != nil {
			return hashErr
		}
		user = &model.User{
			Username:     username,
			PasswordHash: hash,
			RealName:     "管理员",
			HomePath:     "/dashboard/analytics",
			Status:       model.UserStatusEnabled,
		}
		if err := a.store.CreateUser(user); err != nil {
			return err
		}
	}
	return a.store.EnsureUserRole(user.ID, superRoleID)
}

// rbacMenuTree 内置菜单树：首页 + 权限管理。后续业务菜单再往这里加。
func rbacMenuTree() []menuSeed {
	return []menuSeed{
		{
			Name: "Dashboard", Title: "首页", Path: "/dashboard", Icon: "lucide:home",
			Type: model.MenuTypeDir, Sort: -1,
			Children: []menuSeed{
				{
					Name: "Analytics", Title: "首页", Path: "/dashboard/analytics",
					Component: "/dashboard/analytics/index", Icon: "lucide:home",
					AuthCode: "dashboard:view", Type: model.MenuTypeMenu, Sort: 1,
				},
			},
		},
		{
			Name: "Permission", Title: "权限管理", Path: "/permission", Icon: "lucide:key",
			Type: model.MenuTypeDir, Sort: 110,
			Children: []menuSeed{
				{
					Name: "PermissionUser", Title: "用户管理", Path: "/permission/user",
					Component: "/_shared/placeholder", Icon: "lucide:user",
					AuthCode: "system:user:list", Type: model.MenuTypeMenu, Sort: 1,
				},
				{
					Name: "PermissionRole", Title: "角色管理", Path: "/permission/role",
					Component: "/_shared/placeholder", Icon: "lucide:users-round",
					AuthCode: "system:role:list", Type: model.MenuTypeMenu, Sort: 2,
				},
				{
					Name: "PermissionMenu", Title: "菜单管理", Path: "/permission/menu",
					Component: "/_shared/placeholder", Icon: "lucide:menu",
					AuthCode: "system:menu:list", Type: model.MenuTypeMenu, Sort: 3,
				},
			},
		},
	}
}

// EnsureDefaultAdmin 启动入口调用，内部转到 SeedRBAC。
func (a *App) EnsureDefaultAdmin(username, password string) error {
	return a.SeedRBAC(username, password)
}
