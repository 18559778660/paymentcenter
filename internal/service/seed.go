package service

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"paymentcenter/internal/model"
)

// menusJSON 内置菜单种子，改菜单编辑 menus.json 即可。
//
// 顶层结构字段：name / title / path / component / icon / authCode / type / sort
// type：0目录 1菜单 2按钮
//
// meta 为 Vben 前端 RouteMeta 扩展，原样透传，常用键：
//
//	hideInMenu          不在侧边栏显示
//	hideChildrenInMenu  侧边栏不展开子菜单
//	affixTab            标签栏钉住，关不掉
//	keepAlive           页面缓存（需要时再加）
//
//go:embed menus.json
var menusJSON []byte

// menuSeed 初始化菜单用的内存结构，启动时写入 menus 表。
type menuSeed struct {
	Name      string                 `json:"name"`      // 路由 Name，需唯一
	Title     string                 `json:"title"`     // 显示标题
	Path      string                 `json:"path"`      // 前端路径
	Component string                 `json:"component"` // 前端组件路径
	Icon      string                 `json:"icon"`      // 图标
	AuthCode  string                 `json:"authCode"`  // 权限码
	Type      int                    `json:"type"`      // 0目录 1菜单 2按钮
	Sort      int                    `json:"sort"`      // 排序
	Meta      map[string]interface{} `json:"meta"`      // Vben meta，见上方常用键说明
	Children  []menuSeed             `json:"children"`  // 子菜单
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
	merchantRole, err := a.ensureRole("merchant", "商户", "商户后台账号")
	if err != nil {
		return err
	}
	distributionRole, err := a.ensureRole("distribution", "分配子账号", "可登录并查看通道账号")
	if err != nil {
		return err
	}

	seeds, err := loadMenuSeeds()
	if err != nil {
		return err
	}
	menuIDs, err := a.ensureMenus(seeds)
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
	// 商户角色先开放首页，后续可按业务再扩
	if analytics, err := a.store.FindMenuByName("Analytics"); err == nil {
		if err := a.ensureRoleMenuWithAncestors(merchantRole.ID, analytics.ID); err != nil {
			return err
		}
		if err := a.ensureRoleMenuWithAncestors(distributionRole.ID, analytics.ID); err != nil {
			return err
		}
	}
	if channelAccount, err := a.store.FindMenuByName("ChannelAccount"); err == nil {
		if err := a.ensureRoleMenuWithAncestors(distributionRole.ID, channelAccount.ID); err != nil {
			return err
		}
	}

	if err := a.ensureCardTypes(); err != nil {
		return err
	}
	if err := a.ensureCurrencies(); err != nil {
		return err
	}
	if err := a.ensureCountries(); err != nil {
		return err
	}
	if err := a.ensurePlatforms(); err != nil {
		return err
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

// ensureRoleMenuWithAncestors 给角色绑定菜单，并递归绑定所有父级目录（否则子菜单无法出现在侧边栏）。
func (a *App) ensureRoleMenuWithAncestors(roleID, menuID uint) error {
	menu, err := a.store.GetMenuByID(menuID)
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return err
	}
	if err := a.store.EnsureRoleMenu(roleID, menu.ID); err != nil {
		return err
	}
	if menu.ParentID > 0 {
		return a.ensureRoleMenuWithAncestors(roleID, menu.ParentID)
	}
	return nil
}

// ensureMenus 按树写入菜单，已存在的 Name 不覆盖，返回全部菜单 ID。
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
					Meta:      model.JSONMap(item.Meta),
				}
				if err := a.store.CreateMenu(menu); err != nil {
					return err
				}
			} else if len(menu.Meta) == 0 && len(item.Meta) > 0 {
				// 旧数据没有 meta：只补 meta，不覆盖标题路径等
				menu.Meta = model.JSONMap(item.Meta)
				if err := a.store.SaveMenu(menu); err != nil {
					return err
				}
			} else if menu.Component == "/_shared/placeholder" &&
				item.Component != "" &&
				item.Component != menu.Component {
				// 占位页升级为真实页面时，只更新 component
				menu.Component = item.Component
				if err := a.store.SaveMenu(menu); err != nil {
					return err
				}
			} else if item.Type == model.MenuTypeDir && menu.Type != model.MenuTypeDir {
				// 种子是目录但库里被标成菜单时，只纠正 type
				menu.Type = model.MenuTypeDir
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
			Type:         model.UserTypeAdmin,
			Status:       model.UserStatusEnabled,
		}
		if err := a.store.CreateUser(user); err != nil {
			return err
		}
	}
	return a.store.EnsureUserRole(user.ID, superRoleID)
}

// ensureCardTypes 首次启动按 card_brands.json 写入默认卡类型，已有数据不覆盖。
func (a *App) ensureCardTypes() error {
	n, err := a.store.CountCardTypes()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	seeds, err := cardTypeSeedRecords()
	if err != nil {
		return err
	}
	for i := range seeds {
		if err := a.store.CreateCardType(&seeds[i]); err != nil {
			return err
		}
	}
	return nil
}

// ensureCurrencies 首次启动按 currencies.json 写入默认货币，已有数据不覆盖。
func (a *App) ensureCurrencies() error {
	n, err := a.store.CountCurrencies()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	seeds, err := currencySeedRecords()
	if err != nil {
		return err
	}
	for i := range seeds {
		if err := a.store.CreateCurrency(&seeds[i]); err != nil {
			return err
		}
	}
	return nil
}

// ensureCountries 首次启动按 countries.json 写入默认国家，已有数据不覆盖。
func (a *App) ensureCountries() error {
	n, err := a.store.CountCountries()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	seeds, err := countrySeedRecords()
	if err != nil {
		return err
	}
	for i := range seeds {
		if err := a.store.CreateCountry(&seeds[i]); err != nil {
			return err
		}
	}
	return nil
}

// ensurePlatforms 幂等写入默认通道平台。
func (a *App) ensurePlatforms() error {
	if _, err := a.ensurePlatform(model.PlatformCodeStripe, "Stripe", 1); err != nil {
		return err
	}
	_, err := a.ensurePlatform(model.PlatformCodePaypal, "Paypal", 2)
	return err
}

func (a *App) ensurePlatform(code, name string, sort int) (*model.Platform, error) {
	platform, err := a.store.FindPlatformByCode(code)
	if err == nil {
		return platform, nil
	}
	if !isNotFound(err) {
		return nil, err
	}
	platform = &model.Platform{
		Code:   code,
		Name:   name,
		Sort:   sort,
		Status: model.PlatformStatusEnabled,
	}
	if err := a.store.CreatePlatform(platform); err != nil {
		return nil, err
	}
	return platform, nil
}

// loadMenuSeeds 读取 menus.json。只给库里还不存在的菜单做插入。
func loadMenuSeeds() ([]menuSeed, error) {
	var seeds []menuSeed
	if err := json.Unmarshal(menusJSON, &seeds); err != nil {
		return nil, fmt.Errorf("parse menus.json: %w", err)
	}
	return seeds, nil
}

// EnsureDefaultAdmin 启动入口调用，内部转到 SeedRBAC。
func (a *App) EnsureDefaultAdmin(username, password string) error {
	return a.SeedRBAC(username, password)
}
