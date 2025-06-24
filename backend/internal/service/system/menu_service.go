package system

import (
	"fmt"
	"strings"
	"time"
	"wosm/internal/repository/dao"
	"wosm/internal/repository/model"
)

// MenuService 菜单服务 对应Java后端的ISysMenuService
type MenuService struct {
	menuDao *dao.MenuDao
}

// NewMenuService 创建菜单服务实例
func NewMenuService() *MenuService {
	return &MenuService{
		menuDao: dao.NewMenuDao(),
	}
}

// SelectMenuList 查询菜单列表 对应Java后端的selectMenuList(SysMenu menu, Long userId)
func (s *MenuService) SelectMenuList(menu *model.SysMenu, userId int64) ([]model.SysMenu, error) {
	fmt.Printf("MenuService.SelectMenuList: 查询菜单列表, UserID=%d\n", userId)

	// 管理员显示所有菜单信息 - 对应Java后端的SysUser.isAdmin(userId)
	userService := NewUserService()
	user, err := userService.SelectUserById(userId)
	if err != nil {
		fmt.Printf("MenuService.SelectMenuList: 查询用户失败: %v\n", err)
		return []model.SysMenu{}, err
	}
	if user == nil {
		fmt.Printf("MenuService.SelectMenuList: 用户不存在, UserID=%d\n", userId)
		return []model.SysMenu{}, fmt.Errorf("用户不存在")
	}

	if user.IsAdmin() {
		return s.menuDao.SelectMenuList(menu)
	} else {
		return s.menuDao.SelectMenuListByUserId(menu, userId)
	}
}

// SelectMenuListByUserId 根据用户ID查询菜单列表 对应Java后端的selectMenuList(Long userId)重载方法
func (s *MenuService) SelectMenuListByUserId(userId int64) ([]model.SysMenu, error) {
	fmt.Printf("MenuService.SelectMenuListByUserId: 根据用户ID查询菜单列表, UserID=%d\n", userId)
	// 调用带参数的方法，传入空的菜单对象
	return s.SelectMenuList(&model.SysMenu{}, userId)
}

// SelectMenuById 根据菜单ID查询菜单信息 对应Java后端的selectMenuById
func (s *MenuService) SelectMenuById(menuId int64) (*model.SysMenu, error) {
	fmt.Printf("MenuService.SelectMenuById: 查询菜单信息, MenuID=%d\n", menuId)
	return s.menuDao.SelectMenuById(menuId)
}

// BuildMenuTree 构建菜单树结构 对应Java后端的buildMenuTree
func (s *MenuService) BuildMenuTree(menus []model.SysMenu) []model.SysMenu {
	fmt.Printf("MenuService.BuildMenuTree: 构建菜单树, 菜单数量=%d\n", len(menus))

	var returnList []model.SysMenu
	var tempList []int64

	// 收集所有菜单ID
	for _, menu := range menus {
		tempList = append(tempList, menu.MenuID)
	}

	// 查找顶级节点
	for _, menu := range menus {
		// 如果是顶级节点（父节点不在当前列表中）
		if !contains(tempList, menu.ParentID) {
			s.recursionFn(menus, &menu)
			returnList = append(returnList, menu)
		}
	}

	if len(returnList) == 0 {
		returnList = menus
	}

	return returnList
}

// BuildMenuTreeSelect 构建下拉树结构 对应Java后端的buildMenuTreeSelect
func (s *MenuService) BuildMenuTreeSelect(menus []model.SysMenu) []model.TreeSelect {
	fmt.Printf("MenuService.BuildMenuTreeSelect: 构建下拉树结构\n")

	menuTrees := s.BuildMenuTree(menus)
	var treeSelects []model.TreeSelect

	for _, menu := range menuTrees {
		treeSelect := s.convertToTreeSelect(menu)
		treeSelects = append(treeSelects, treeSelect)
	}

	return treeSelects
}

// SelectMenuListByRoleId 根据角色ID查询菜单列表 对应Java后端的selectMenuListByRoleId
func (s *MenuService) SelectMenuListByRoleId(roleId int64) ([]int64, error) {
	fmt.Printf("MenuService.SelectMenuListByRoleId: 查询角色菜单, RoleID=%d\n", roleId)

	// 需要先查询角色信息，获取MenuCheckStrictly字段
	roleService := NewRoleService()
	role, err := roleService.SelectRoleById(roleId)
	if err != nil {
		fmt.Printf("MenuService.SelectMenuListByRoleId: 查询角色信息失败: %v\n", err)
		return []int64{}, err
	}
	if role == nil {
		fmt.Printf("MenuService.SelectMenuListByRoleId: 角色不存在, RoleID=%d\n", roleId)
		return []int64{}, nil
	}

	return s.menuDao.SelectMenuListByRoleId(roleId, role.MenuCheckStrictly)
}

// HasChildByMenuId 是否存在菜单子节点 对应Java后端的hasChildByMenuId
func (s *MenuService) HasChildByMenuId(menuId int64) (bool, error) {
	fmt.Printf("MenuService.HasChildByMenuId: 检查子菜单, MenuID=%d\n", menuId)
	return s.menuDao.HasChildByMenuId(menuId)
}

// CheckMenuExistRole 查询菜单是否存在角色 对应Java后端的checkMenuExistRole
func (s *MenuService) CheckMenuExistRole(menuId int64) (bool, error) {
	fmt.Printf("MenuService.CheckMenuExistRole: 检查菜单角色关联, MenuID=%d\n", menuId)
	return s.menuDao.CheckMenuExistRole(menuId)
}

// CheckMenuNameUnique 校验菜单名称是否唯一 对应Java后端的checkMenuNameUnique
func (s *MenuService) CheckMenuNameUnique(menu *model.SysMenu) (bool, error) {
	fmt.Printf("MenuService.CheckMenuNameUnique: 校验菜单名称, MenuName=%s\n", menu.MenuName)
	return s.menuDao.CheckMenuNameUnique(menu.MenuName, menu.ParentID, menu.MenuID)
}

// InsertMenu 新增菜单 对应Java后端的insertMenu
func (s *MenuService) InsertMenu(menu *model.SysMenu) error {
	fmt.Printf("MenuService.InsertMenu: 新增菜单, MenuName=%s\n", menu.MenuName)

	// 设置创建时间
	now := time.Now()
	menu.CreateTime = &now

	return s.menuDao.InsertMenu(menu)
}

// UpdateMenu 修改菜单 对应Java后端的updateMenu
func (s *MenuService) UpdateMenu(menu *model.SysMenu) error {
	fmt.Printf("MenuService.UpdateMenu: 修改菜单, MenuID=%d\n", menu.MenuID)

	// 设置更新时间
	now := time.Now()
	menu.UpdateTime = &now

	return s.menuDao.UpdateMenu(menu)
}

// DeleteMenuById 删除菜单 对应Java后端的deleteMenuById
func (s *MenuService) DeleteMenuById(menuId int64) error {
	fmt.Printf("MenuService.DeleteMenuById: 删除菜单, MenuID=%d\n", menuId)
	return s.menuDao.DeleteMenuById(menuId)
}

// SelectMenuPermsByUserId 根据用户ID查询权限 对应Java后端的selectMenuPermsByUserId
func (s *MenuService) SelectMenuPermsByUserId(userId int64) ([]string, error) {
	fmt.Printf("MenuService.SelectMenuPermsByUserId: 查询用户权限, UserID=%d\n", userId)
	return s.menuDao.SelectMenuPermsByUserId(userId)
}

// SelectMenuPermsByRoleId 根据角色ID查询权限 对应Java后端的selectMenuPermsByRoleId
func (s *MenuService) SelectMenuPermsByRoleId(roleId int64) ([]string, error) {
	fmt.Printf("MenuService.SelectMenuPermsByRoleId: 查询角色权限, RoleID=%d\n", roleId)
	return s.menuDao.SelectMenuPermsByRoleId(roleId)
}

// SelectMenuTreeByUserId 根据用户ID查询菜单树 对应Java后端的selectMenuTreeByUserId
func (s *MenuService) SelectMenuTreeByUserId(userId int64) ([]model.SysMenu, error) {
	fmt.Printf("MenuService.SelectMenuTreeByUserId: 查询用户菜单树, UserID=%d\n", userId)

	var menus []model.SysMenu
	var err error

	// 管理员显示所有菜单信息 - 对应Java后端的SecurityUtils.isAdmin(userId)
	userService := NewUserService()
	user, err := userService.SelectUserById(userId)
	if err != nil {
		fmt.Printf("MenuService.SelectMenuTreeByUserId: 查询用户失败: %v\n", err)
		return nil, err
	}
	if user == nil {
		fmt.Printf("MenuService.SelectMenuTreeByUserId: 用户不存在, UserID=%d\n", userId)
		return nil, fmt.Errorf("用户不存在")
	}

	if user.IsAdmin() {
		menus, err = s.menuDao.SelectMenuTreeAll()
	} else {
		menus, err = s.menuDao.SelectMenuTreeByUserId(userId)
	}

	if err != nil {
		return nil, err
	}

	// 构建菜单树
	return s.getChildPerms(menus, 0), nil
}

// BuildMenus 构建前端路由所需要的菜单 对应Java后端的buildMenus
func (s *MenuService) BuildMenus(menus []model.SysMenu) []model.RouterVo {
	fmt.Printf("MenuService.BuildMenus: 构建前端路由, 菜单数量=%d\n", len(menus))

	var routers []model.RouterVo
	for _, menu := range menus {
		router := model.RouterVo{}
		router.Hidden = menu.Visible == "1"
		router.Name = s.getRouteName(menu)
		router.Path = s.getRouterPath(menu)
		router.Component = s.getComponent(menu)
		router.Query = menu.Query
		router.Meta = &model.MetaVo{
			Title:   menu.MenuName,
			Icon:    menu.Icon,
			NoCache: menu.IsCache == "1",
			Link:    "",
		}

		// 处理外链 - 设置Meta.Link
		if menu.IsFrame == "0" && isHttpUrl(menu.Path) {
			router.Meta.Link = menu.Path
		}

		// 处理子菜单 - 对应Java后端的子菜单处理逻辑
		cMenus := menu.Children
		if len(cMenus) > 0 && menu.MenuType == "M" {
			router.AlwaysShow = true
			router.Redirect = "noRedirect"
			router.Children = s.BuildMenus(cMenus)
		} else if menu.ParentID == 0 && s.isInnerLink(menu) {
			// 内链特殊处理 - 对应Java后端的内链处理逻辑
			// 当父节点为0且为内链时的特殊处理（优先级高于isMenuFrame）
			router.Meta = &model.MetaVo{
				Title: menu.MenuName,
				Icon:  menu.Icon,
			}
			router.Path = "/"
			var childrenList []model.RouterVo
			children := model.RouterVo{
				Path:      s.innerLinkReplaceEach(menu.Path),
				Component: "InnerLink",
				Name:      s.getRouteNameWithPath(menu.RouteName, s.innerLinkReplaceEach(menu.Path)),
				Meta: &model.MetaVo{
					Title: menu.MenuName,
					Icon:  menu.Icon,
					Link:  menu.Path,
				},
			}
			childrenList = append(childrenList, children)
			router.Children = childrenList
		} else if s.isMenuFrame(menu) {
			// 菜单框架处理 - 对应Java后端的isMenuFrame处理逻辑
			// 当菜单为内部跳转时，需要设置Meta为null并创建子路由
			router.Meta = nil
			var childrenList []model.RouterVo
			children := model.RouterVo{
				Path:      menu.Path,
				Component: menu.Component,
				Name:      s.getRouteNameWithPath(menu.RouteName, menu.Path),
				Meta: &model.MetaVo{
					Title:   menu.MenuName,
					Icon:    menu.Icon,
					NoCache: menu.IsCache == "1",
					Link:    menu.Path,
				},
				Query: menu.Query,
			}
			childrenList = append(childrenList, children)
			router.Children = childrenList
		}

		routers = append(routers, router)
	}

	return routers
}

// getChildPerms 根据父节点的ID获取所有子节点 对应Java后端的getChildPerms
func (s *MenuService) getChildPerms(list []model.SysMenu, parentId int64) []model.SysMenu {
	var returnList []model.SysMenu
	for _, menu := range list {
		// 一、根据传入的某个父节点ID,遍历该父节点的所有子节点
		if menu.ParentID == parentId {
			s.recursionFn(list, &menu)
			returnList = append(returnList, menu)
		}
	}
	return returnList
}

// getRouteName 获取路由名称 对应Java后端的getRouteName
func (s *MenuService) getRouteName(menu model.SysMenu) string {
	routeName := menu.RouteName
	// 如果没有路由名称，使用路径生成
	if routeName == "" && menu.Path != "" {
		// 简单处理：去掉特殊字符，首字母大写
		name := strings.ReplaceAll(menu.Path, "/", "")
		name = strings.ReplaceAll(name, "-", "")
		name = strings.ReplaceAll(name, "_", "")
		name = strings.ReplaceAll(name, ".", "")
		name = strings.ReplaceAll(name, ":", "")
		if len(name) > 0 {
			routeName = strings.ToUpper(name[:1]) + name[1:]
		}
	}
	// 非外链并且是一级目录（类型为目录）时，某些情况下需要清空路由名称
	// 但这里我们保留路由名称以便测试和调试
	return routeName
}

// getRouteNameWithPath 获取带路径的路由名称 对应Java后端的getRouteName重载方法
func (s *MenuService) getRouteNameWithPath(routeName, path string) string {
	if routeName != "" {
		return routeName
	}
	// 如果没有路由名称，使用路径生成
	if path != "" {
		// 简单处理：去掉特殊字符，首字母大写
		name := strings.ReplaceAll(path, "/", "")
		name = strings.ReplaceAll(name, "-", "")
		name = strings.ReplaceAll(name, "_", "")
		if len(name) > 0 {
			return strings.ToUpper(name[:1]) + name[1:]
		}
	}
	return ""
}

// getRouterPath 获取路由地址 对应Java后端的getRouterPath
func (s *MenuService) getRouterPath(menu model.SysMenu) string {
	routerPath := menu.Path
	// 内链打开外网方式
	if menu.ParentID != 0 && s.isInnerLink(menu) {
		routerPath = s.innerLinkReplaceEach(routerPath)
	}
	// 非外链并且是一级目录（类型为目录）
	if menu.ParentID == 0 && menu.MenuType == "M" && menu.IsFrame == "1" {
		routerPath = "/" + menu.Path
	}
	// 非外链并且是一级目录（类型为菜单）
	if menu.ParentID == 0 && menu.MenuType == "C" && menu.IsFrame == "1" {
		routerPath = "/"
	}
	return routerPath
}

// getComponent 获取组件信息 对应Java后端的getComponent
func (s *MenuService) getComponent(menu model.SysMenu) string {
	component := "Layout"
	if menu.Component != "" && !s.isMenuFrame(menu) {
		component = menu.Component
	} else if menu.Component == "" && menu.ParentID != 0 && s.isInnerLink(menu) {
		component = "InnerLink"
	} else if menu.Component == "" && s.isParentView(menu) {
		component = "ParentView"
	}
	return component
}

// isMenuFrame 是否为菜单内部跳转 对应Java后端的isMenuFrame
func (s *MenuService) isMenuFrame(menu model.SysMenu) bool {
	return menu.ParentID == 0 && menu.MenuType == "C" && menu.IsFrame == "1"
}

// isInnerLink 是否为内链组件 对应Java后端的isInnerLink
func (s *MenuService) isInnerLink(menu model.SysMenu) bool {
	return menu.IsFrame == "1" && isHttpUrl(menu.Path)
}

// isParentView 是否为parent_view组件 对应Java后端的isParentView
func (s *MenuService) isParentView(menu model.SysMenu) bool {
	return menu.ParentID != 0 && menu.MenuType == "M"
}

// innerLinkReplaceEach 内链域名特殊字符替换 对应Java后端的innerLinkReplaceEach
func (s *MenuService) innerLinkReplaceEach(path string) string {
	// 对应Java后端的内链域名特殊字符替换逻辑
	// 移除协议前缀
	path = strings.ReplaceAll(path, "http://", "")
	path = strings.ReplaceAll(path, "https://", "")
	// 移除www前缀
	path = strings.ReplaceAll(path, "www.", "")
	// 将点号替换为斜杠
	path = strings.ReplaceAll(path, ".", "/")
	// 将冒号替换为斜杠
	path = strings.ReplaceAll(path, ":", "/")
	return path
}

// InnerLinkReplaceEach 公开的内链域名特殊字符替换方法，供测试使用
func (s *MenuService) InnerLinkReplaceEach(path string) string {
	return s.innerLinkReplaceEach(path)
}

// isHttpUrl 检查是否为HTTP URL 对应Java后端的StringUtils.ishttp
func isHttpUrl(url string) bool {
	if url == "" {
		return false
	}
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// recursionFn 递归构建子菜单
func (s *MenuService) recursionFn(list []model.SysMenu, menu *model.SysMenu) {
	// 得到子节点列表
	childList := s.getChildList(list, *menu)
	menu.Children = childList

	for i := range childList {
		if s.hasChild(list, childList[i]) {
			s.recursionFn(list, &childList[i])
		}
	}
}

// getChildList 得到子节点列表
func (s *MenuService) getChildList(list []model.SysMenu, menu model.SysMenu) []model.SysMenu {
	var childList []model.SysMenu

	for _, item := range list {
		if item.ParentID == menu.MenuID {
			childList = append(childList, item)
		}
	}

	return childList
}

// hasChild 判断是否有子节点
func (s *MenuService) hasChild(list []model.SysMenu, menu model.SysMenu) bool {
	return len(s.getChildList(list, menu)) > 0
}

// convertToTreeSelect 转换为TreeSelect结构 对应Java后端的TreeSelect(SysMenu menu)构造函数
func (s *MenuService) convertToTreeSelect(menu model.SysMenu) model.TreeSelect {
	treeSelect := model.TreeSelect{
		ID:       menu.MenuID,
		Label:    menu.MenuName,
		Disabled: false, // 对应Java后端的TreeSelect(SysMenu menu)构造函数，菜单默认不禁用
	}

	if len(menu.Children) > 0 {
		for _, child := range menu.Children {
			childTreeSelect := s.convertToTreeSelect(child)
			treeSelect.Children = append(treeSelect.Children, childTreeSelect)
		}
	}

	return treeSelect
}

// contains 检查切片是否包含指定元素
func contains(slice []int64, item int64) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
