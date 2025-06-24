package system

import (
	"fmt"
	"strconv"
	"strings"
	"wosm/internal/constants"
	"wosm/internal/repository/model"
	systemService "wosm/internal/service/system"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// MenuController 菜单管理控制器 对应Java后端的SysMenuController
type MenuController struct {
	menuService *systemService.MenuService
}

// NewMenuController 创建菜单管理控制器实例
func NewMenuController() *MenuController {
	return &MenuController{
		menuService: systemService.NewMenuService(),
	}
}

// List 获取菜单列表 对应Java后端的list方法
// @Router /system/menu/list [get]
func (c *MenuController) List(ctx *gin.Context) {
	fmt.Printf("MenuController.List: 查询菜单列表\n")

	// 构建查询条件
	menu := &model.SysMenu{}

	// 获取查询参数
	if menuName := ctx.Query("menuName"); menuName != "" {
		menu.MenuName = menuName
	}
	if status := ctx.Query("status"); status != "" {
		menu.Status = status
	}

	// 获取当前用户ID
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)

	// 查询菜单列表
	menus, err := c.menuService.SelectMenuList(menu, currentUser.UserID)
	if err != nil {
		fmt.Printf("MenuController.List: 查询菜单列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询菜单列表失败")
		return
	}

	fmt.Printf("MenuController.List: 查询菜单列表成功, 数量=%d\n", len(menus))
	response.SuccessWithData(ctx, menus)
}

// GetInfo 根据菜单ID获取详细信息 对应Java后端的getInfo方法
// @Router /system/menu/{menuId} [get]
func (c *MenuController) GetInfo(ctx *gin.Context) {
	menuIdStr := ctx.Param("menuId")
	menuId, err := strconv.ParseInt(menuIdStr, 10, 64)
	if err != nil {
		fmt.Printf("MenuController.GetInfo: 菜单ID格式错误: %s\n", menuIdStr)
		response.ErrorWithMessage(ctx, "菜单ID格式错误")
		return
	}

	fmt.Printf("MenuController.GetInfo: 查询菜单详情, MenuID=%d\n", menuId)

	// 查询菜单信息
	menu, err := c.menuService.SelectMenuById(menuId)
	if err != nil {
		fmt.Printf("MenuController.GetInfo: 查询菜单详情失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询菜单详情失败")
		return
	}

	if menu == nil {
		fmt.Printf("MenuController.GetInfo: 菜单不存在, MenuID=%d\n", menuId)
		response.ErrorWithMessage(ctx, "菜单不存在")
		return
	}

	fmt.Printf("MenuController.GetInfo: 查询菜单详情成功, MenuID=%d\n", menuId)
	response.SuccessWithData(ctx, menu)
}

// TreeSelect 获取菜单下拉树列表 对应Java后端的treeselect方法
// @Summary 获取菜单下拉树列表
// @Description 获取菜单树形结构数据用于下拉选择
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/menu/treeselect [get]
func (c *MenuController) TreeSelect(ctx *gin.Context) {
	fmt.Printf("MenuController.TreeSelect: 获取菜单下拉树列表\n")

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)

	// 构建查询条件
	menu := &model.SysMenu{}
	if menuName := ctx.Query("menuName"); menuName != "" {
		menu.MenuName = menuName
	}
	if status := ctx.Query("status"); status != "" {
		menu.Status = status
	}

	// 查询菜单列表
	menus, err := c.menuService.SelectMenuList(menu, currentUser.UserID)
	if err != nil {
		fmt.Printf("MenuController.TreeSelect: 查询菜单列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询菜单列表失败")
		return
	}

	// 构建菜单树选择结构
	treeSelect := c.menuService.BuildMenuTreeSelect(menus)

	fmt.Printf("MenuController.TreeSelect: 获取菜单下拉树列表成功, 数量=%d\n", len(treeSelect))
	response.SuccessWithData(ctx, treeSelect)
}

// RoleMenuTreeSelect 加载对应角色菜单列表树 对应Java后端的roleMenuTreeselect方法
// @Summary 加载对应角色菜单列表树
// @Description 根据角色ID获取菜单树和已选中的菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param roleId path int true "角色ID"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/menu/roleMenuTreeselect/{roleId} [get]
func (c *MenuController) RoleMenuTreeSelect(ctx *gin.Context) {
	fmt.Printf("MenuController.RoleMenuTreeSelect: 加载对应角色菜单列表树\n")

	roleIdStr := ctx.Param("roleId")
	roleId, err := strconv.ParseInt(roleIdStr, 10, 64)
	if err != nil {
		response.ErrorWithMessage(ctx, "角色ID格式错误")
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)

	// 查询所有菜单列表 - 对应Java后端的menuService.selectMenuList(getUserId())
	menus, err := c.menuService.SelectMenuListByUserId(currentUser.UserID)
	if err != nil {
		fmt.Printf("MenuController.RoleMenuTreeSelect: 查询菜单列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询菜单列表失败")
		return
	}

	// 构建菜单树选择结构
	treeSelect := c.menuService.BuildMenuTreeSelect(menus)

	// 获取角色已选中的菜单ID列表
	checkedKeys, err := c.menuService.SelectMenuListByRoleId(roleId)
	if err != nil {
		fmt.Printf("MenuController.RoleMenuTreeSelect: 查询角色菜单失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询角色菜单失败")
		return
	}

	// 确保checkedKeys不为nil，即使没有菜单也返回空数组
	if checkedKeys == nil {
		checkedKeys = []int64{}
	}

	// 使用AjaxResult格式，与Java后端保持一致
	ajax := response.AjaxSuccess()
	ajax.Put("checkedKeys", checkedKeys)
	ajax.Put("menus", treeSelect)

	fmt.Printf("MenuController.RoleMenuTreeSelect: 加载对应角色菜单列表树成功, RoleID=%d\n", roleId)
	response.SendAjaxResult(ctx, ajax)
}

// Add 新增菜单 对应Java后端的add方法
// @Summary 新增菜单
// @Description 新增菜单信息
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param menu body model.SysMenu true "菜单信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/menu [post]
func (c *MenuController) Add(ctx *gin.Context) {
	fmt.Printf("MenuController.Add: 新增菜单\n")

	var menu model.SysMenu
	if err := ctx.ShouldBindJSON(&menu); err != nil {
		fmt.Printf("MenuController.Add: 参数绑定失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数格式错误")
		return
	}

	// 校验菜单名称唯一性
	unique, err := c.menuService.CheckMenuNameUnique(&menu)
	if err != nil {
		fmt.Printf("MenuController.Add: 校验菜单名称失败: %v\n", err)
		response.ErrorWithMessage(ctx, "校验菜单名称失败")
		return
	}
	if !unique {
		response.ErrorWithMessage(ctx, fmt.Sprintf("新增菜单'%s'失败，菜单名称已存在", menu.MenuName))
		return
	}

	// 校验外链地址 - 对应Java后端的UserConstants.YES_FRAME.equals(menu.getIsFrame())
	if menu.IsFrame == constants.YES_FRAME && menu.Path != "" && !isHttpUrl(menu.Path) {
		response.ErrorWithMessage(ctx, fmt.Sprintf("新增菜单'%s'失败，地址必须以http(s)://开头", menu.MenuName))
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)
	menu.CreateBy = currentUser.User.UserName

	// 新增菜单
	if err := c.menuService.InsertMenu(&menu); err != nil {
		fmt.Printf("MenuController.Add: 新增菜单失败: %v\n", err)
		response.ErrorWithMessage(ctx, "新增菜单失败")
		return
	}

	fmt.Printf("MenuController.Add: 新增菜单成功, MenuName=%s\n", menu.MenuName)
	response.SuccessWithMessage(ctx, "新增成功")
}

// Edit 修改菜单 对应Java后端的edit方法
// @Summary 修改菜单
// @Description 修改菜单信息
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param menu body model.SysMenu true "菜单信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/menu [put]
func (c *MenuController) Edit(ctx *gin.Context) {
	fmt.Printf("MenuController.Edit: 修改菜单\n")

	var menu model.SysMenu
	if err := ctx.ShouldBindJSON(&menu); err != nil {
		fmt.Printf("MenuController.Edit: 参数绑定失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数格式错误")
		return
	}

	// 校验菜单名称唯一性
	unique, err := c.menuService.CheckMenuNameUnique(&menu)
	if err != nil {
		fmt.Printf("MenuController.Edit: 校验菜单名称失败: %v\n", err)
		response.ErrorWithMessage(ctx, "校验菜单名称失败")
		return
	}
	if !unique {
		response.ErrorWithMessage(ctx, fmt.Sprintf("修改菜单'%s'失败，菜单名称已存在", menu.MenuName))
		return
	}

	// 校验外链地址 - 对应Java后端的UserConstants.YES_FRAME.equals(menu.getIsFrame())
	if menu.IsFrame == constants.YES_FRAME && menu.Path != "" && !isHttpUrl(menu.Path) {
		response.ErrorWithMessage(ctx, fmt.Sprintf("修改菜单'%s'失败，地址必须以http(s)://开头", menu.MenuName))
		return
	}

	// 校验上级菜单不能是自己
	if menu.MenuID == menu.ParentID {
		response.ErrorWithMessage(ctx, fmt.Sprintf("修改菜单'%s'失败，上级菜单不能选择自己", menu.MenuName))
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)
	menu.UpdateBy = currentUser.User.UserName

	// 修改菜单
	if err := c.menuService.UpdateMenu(&menu); err != nil {
		fmt.Printf("MenuController.Edit: 修改菜单失败: %v\n", err)
		response.ErrorWithMessage(ctx, "修改菜单失败")
		return
	}

	fmt.Printf("MenuController.Edit: 修改菜单成功, MenuID=%d\n", menu.MenuID)
	response.SuccessWithMessage(ctx, "修改成功")
}

// Remove 删除菜单 对应Java后端的remove方法
// @Summary 删除菜单
// @Description 删除菜单信息
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param menuId path int true "菜单ID"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/menu/{menuId} [delete]
func (c *MenuController) Remove(ctx *gin.Context) {
	fmt.Printf("MenuController.Remove: 删除菜单\n")

	menuIdStr := ctx.Param("menuId")
	menuId, err := strconv.ParseInt(menuIdStr, 10, 64)
	if err != nil {
		response.ErrorWithMessage(ctx, "菜单ID格式错误")
		return
	}

	// 检查是否存在子菜单
	hasChild, err := c.menuService.HasChildByMenuId(menuId)
	if err != nil {
		fmt.Printf("MenuController.Remove: 检查子菜单失败: %v\n", err)
		response.ErrorWithMessage(ctx, "检查子菜单失败")
		return
	}
	if hasChild {
		response.ErrorWithMessage(ctx, "存在子菜单,不允许删除")
		return
	}

	// 检查菜单是否已分配给角色
	assigned, err := c.menuService.CheckMenuExistRole(menuId)
	if err != nil {
		fmt.Printf("MenuController.Remove: 检查菜单角色分配失败: %v\n", err)
		response.ErrorWithMessage(ctx, "检查菜单角色分配失败")
		return
	}
	if assigned {
		response.ErrorWithMessage(ctx, "菜单已分配,不允许删除")
		return
	}

	// 删除菜单
	if err := c.menuService.DeleteMenuById(menuId); err != nil {
		fmt.Printf("MenuController.Remove: 删除菜单失败: %v\n", err)
		response.ErrorWithMessage(ctx, "删除菜单失败")
		return
	}

	fmt.Printf("MenuController.Remove: 删除菜单成功, MenuID=%d\n", menuId)
	response.SuccessWithMessage(ctx, "删除成功")
}

// isHttpUrl 检查是否为HTTP URL 对应Java后端的StringUtils.ishttp
func isHttpUrl(url string) bool {
	if url == "" {
		return false
	}
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}
