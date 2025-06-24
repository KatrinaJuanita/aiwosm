package system

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"wosm/internal/repository/model"
	systemService "wosm/internal/service/system"
	"wosm/pkg/datascope"
	"wosm/pkg/excel"
	"wosm/pkg/export"
	"wosm/pkg/operlog"
	"wosm/pkg/response"
	"wosm/pkg/utils"

	"github.com/gin-gonic/gin"
)

// RoleController 角色管理控制器 对应Java后端的SysRoleController
type RoleController struct {
	roleService    *systemService.RoleService
	deptService    *systemService.DeptService
	userService    *systemService.UserService
	operLogService *systemService.OperLogService
}

// NewRoleController 创建角色管理控制器实例
func NewRoleController() *RoleController {
	return &RoleController{
		roleService:    systemService.NewRoleService(),
		deptService:    systemService.NewDeptService(),
		userService:    systemService.NewUserService(),
		operLogService: systemService.NewOperLogService(),
	}
}

// List 查询角色列表 对应Java后端的list方法
// @PreAuthorize("@ss.hasPermi('system:role:list')")
// @Router /system/role/list [get]
func (c *RoleController) List(ctx *gin.Context) {
	fmt.Printf("RoleController.List: 查询角色列表\n")

	// 设置请求分页数据 - 对应Java后端的startPage()
	pageDomain := utils.StartPage(ctx)

	// 构建查询条件 - 对应Java后端直接绑定SysRole实体
	role := &model.SysRole{}
	if roleName := ctx.Query("roleName"); roleName != "" {
		role.RoleName = roleName
	}
	if status := ctx.Query("status"); status != "" {
		role.Status = status
	}
	if roleKey := ctx.Query("roleKey"); roleKey != "" {
		role.RoleKey = roleKey
	}

	// 查询角色列表 - 对应Java后端的roleService.selectRoleList(role)
	roles, total, err := c.roleService.SelectRoleList(role, pageDomain.PageNum, pageDomain.PageSize)
	if err != nil {
		fmt.Printf("RoleController.List: 查询角色列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询失败: "+err.Error())
		return
	}

	// 使用Java后端兼容的TableDataInfo格式
	fmt.Printf("RoleController.List: 查询角色列表成功, 总数=%d\n", total)
	tableData := response.GetDataTable(roles, total)
	response.SendTableDataInfo(ctx, tableData)
}

// GetInfo 根据角色ID获取详细信息 对应Java后端的getInfo方法
// @Router /system/role/{roleId} [get]
func (c *RoleController) GetInfo(ctx *gin.Context) {
	roleIdStr := ctx.Param("roleId")
	roleId, err := strconv.ParseInt(roleIdStr, 10, 64)
	if err != nil {
		fmt.Printf("RoleController.GetInfo: 角色ID格式错误: %s\n", roleIdStr)
		response.ErrorWithMessage(ctx, "角色ID格式错误")
		return
	}

	fmt.Printf("RoleController.GetInfo: 查询角色详情, RoleID=%d\n", roleId)

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)

	// 校验数据权限
	if err := c.roleService.CheckRoleDataScope(currentUser.User, roleId); err != nil {
		fmt.Printf("RoleController.GetInfo: 数据权限校验失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 查询角色信息
	role, err := c.roleService.SelectRoleById(roleId)
	if err != nil {
		fmt.Printf("RoleController.GetInfo: 查询角色详情失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询角色详情失败")
		return
	}

	if role == nil {
		fmt.Printf("RoleController.GetInfo: 角色不存在, RoleID=%d\n", roleId)
		response.ErrorWithMessage(ctx, "角色不存在")
		return
	}

	fmt.Printf("RoleController.GetInfo: 查询角色详情成功, RoleID=%d\n", roleId)
	response.SendAjaxResult(ctx, response.AjaxSuccessWithData(role))
}

// Add 新增角色 对应Java后端的add方法
// @Router /system/role [post]
func (c *RoleController) Add(ctx *gin.Context) {
	fmt.Printf("RoleController.Add: 新增角色\n")

	var role model.SysRole
	if err := ctx.ShouldBindJSON(&role); err != nil {
		fmt.Printf("RoleController.Add: 参数绑定失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数格式错误")
		return
	}

	// 校验角色名称是否唯一
	if unique, err := c.roleService.CheckRoleNameUnique(&role); err != nil {
		fmt.Printf("RoleController.Add: 校验角色名称失败: %v\n", err)
		response.ErrorWithMessage(ctx, "校验角色名称失败")
		return
	} else if !unique {
		response.ErrorWithMessage(ctx, fmt.Sprintf("新增角色'%s'失败，角色名称已存在", role.RoleName))
		return
	}

	// 校验角色权限字符串是否唯一
	if unique, err := c.roleService.CheckRoleKeyUnique(&role); err != nil {
		fmt.Printf("RoleController.Add: 校验角色权限字符串失败: %v\n", err)
		response.ErrorWithMessage(ctx, "校验角色权限字符串失败")
		return
	} else if !unique {
		response.ErrorWithMessage(ctx, fmt.Sprintf("新增角色'%s'失败，角色权限已存在", role.RoleName))
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)
	role.CreateBy = currentUser.User.UserName

	// 新增角色
	if err := c.roleService.InsertRole(&role); err != nil {
		fmt.Printf("RoleController.Add: 新增角色失败: %v\n", err)
		response.ErrorWithMessage(ctx, "新增角色失败")
		return
	}

	fmt.Printf("RoleController.Add: 新增角色成功, RoleID=%d\n", role.RoleID)
	response.SendAjaxResult(ctx, response.AjaxSuccess())
}

// Edit 修改角色 对应Java后端的edit方法
// @Router /system/role [put]
func (c *RoleController) Edit(ctx *gin.Context) {
	fmt.Printf("RoleController.Edit: 修改角色\n")

	var role model.SysRole
	if err := ctx.ShouldBindJSON(&role); err != nil {
		fmt.Printf("RoleController.Edit: 参数绑定失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数格式错误")
		return
	}

	// 校验角色是否允许操作
	if err := c.roleService.CheckRoleAllowed(&role); err != nil {
		fmt.Printf("RoleController.Edit: 角色操作校验失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)

	// 校验数据权限
	if err := c.roleService.CheckRoleDataScope(currentUser.User, role.RoleID); err != nil {
		fmt.Printf("RoleController.Edit: 数据权限校验失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 校验角色名称是否唯一
	if unique, err := c.roleService.CheckRoleNameUnique(&role); err != nil {
		fmt.Printf("RoleController.Edit: 校验角色名称失败: %v\n", err)
		response.ErrorWithMessage(ctx, "校验角色名称失败")
		return
	} else if !unique {
		response.ErrorWithMessage(ctx, fmt.Sprintf("修改角色'%s'失败，角色名称已存在", role.RoleName))
		return
	}

	// 校验角色权限字符串是否唯一
	if unique, err := c.roleService.CheckRoleKeyUnique(&role); err != nil {
		fmt.Printf("RoleController.Edit: 校验角色权限字符串失败: %v\n", err)
		response.ErrorWithMessage(ctx, "校验角色权限字符串失败")
		return
	} else if !unique {
		response.ErrorWithMessage(ctx, fmt.Sprintf("修改角色'%s'失败，角色权限已存在", role.RoleName))
		return
	}

	// 设置更新者
	role.UpdateBy = currentUser.User.UserName

	// 修改角色
	if err := c.roleService.UpdateRole(&role); err != nil {
		fmt.Printf("RoleController.Edit: 修改角色失败: %v\n", err)
		response.ErrorWithMessage(ctx, "修改角色失败")
		return
	}

	// 更新缓存用户权限 对应Java后端的权限缓存更新逻辑
	// 如果当前用户不是管理员，需要更新其权限缓存
	if !currentUser.User.IsAdmin() {
		fmt.Printf("RoleController.Edit: 更新用户权限缓存, UserID=%d\n", currentUser.User.UserID)

		// 重新查询用户信息
		updatedUser, err := c.userService.SelectUserByLoginName(currentUser.User.UserName)
		if err != nil {
			fmt.Printf("RoleController.Edit: 查询用户信息失败: %v\n", err)
		} else if updatedUser != nil {
			// 创建权限服务实例
			permissionService := systemService.NewPermissionService()

			// 获取更新后的权限
			permissions, err := permissionService.GetMenuPermission(updatedUser)
			if err != nil {
				fmt.Printf("RoleController.Edit: 获取用户权限失败: %v\n", err)
			} else {
				// 更新当前登录用户的权限信息
				currentUser.User = updatedUser
				currentUser.Permissions = permissions

				// TODO: 这里应该更新Redis中的Token信息，但由于我们没有TokenService，暂时跳过
				// 对应Java后端的tokenService.setLoginUser(loginUser)
				fmt.Printf("RoleController.Edit: 用户权限缓存更新成功, PermissionCount=%d\n", len(permissions))
			}
		}
	}

	fmt.Printf("RoleController.Edit: 修改角色成功, RoleID=%d\n", role.RoleID)
	response.Success(ctx)
}

// Remove 删除角色 对应Java后端的remove方法
// @Router /system/role/{roleIds} [delete]
func (c *RoleController) Remove(ctx *gin.Context) {
	idsStr := ctx.Param("roleIds")
	if idsStr == "" {
		response.ErrorWithMessage(ctx, "参数错误")
		return
	}

	fmt.Printf("RoleController.Remove: 删除角色, IDs=%s\n", idsStr)

	// 解析角色ID列表
	var roleIds []int64
	for _, idStr := range strings.Split(idsStr, ",") {
		if id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64); err == nil {
			roleIds = append(roleIds, id)
		}
	}

	if len(roleIds) == 0 {
		response.ErrorWithMessage(ctx, "参数错误")
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)

	// 批量删除角色
	if err := c.roleService.DeleteRoleByIds(currentUser.User, roleIds); err != nil {
		fmt.Printf("RoleController.Remove: 删除角色失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	fmt.Printf("RoleController.Remove: 删除角色成功, 数量=%d\n", len(roleIds))
	response.SuccessWithMessage(ctx, "删除成功")
}

// ChangeStatus 修改角色状态 对应Java后端的changeStatus方法
// @Router /system/role/changeStatus [put]
func (c *RoleController) ChangeStatus(ctx *gin.Context) {
	fmt.Printf("RoleController.ChangeStatus: 修改角色状态\n")

	var role model.SysRole
	if err := ctx.ShouldBindJSON(&role); err != nil {
		fmt.Printf("RoleController.ChangeStatus: 参数绑定失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数格式错误")
		return
	}

	// 校验角色是否允许操作
	if err := c.roleService.CheckRoleAllowed(&role); err != nil {
		fmt.Printf("RoleController.ChangeStatus: 角色操作校验失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)

	// 校验数据权限
	if err := c.roleService.CheckRoleDataScope(currentUser.User, role.RoleID); err != nil {
		fmt.Printf("RoleController.ChangeStatus: 数据权限校验失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 设置更新者
	role.UpdateBy = currentUser.User.UserName

	// 修改角色状态
	if err := c.roleService.UpdateRoleStatus(&role); err != nil {
		fmt.Printf("RoleController.ChangeStatus: 修改角色状态失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	status := "启用"
	if role.Status == "1" {
		status = "停用"
	}
	fmt.Printf("RoleController.ChangeStatus: 修改角色状态成功, RoleID=%d, Status=%s\n", role.RoleID, status)
	response.SuccessWithMessage(ctx, status+"成功")
}

// DataScope 角色数据权限分配 对应Java后端的dataScope方法
// @Router /system/role/dataScope [put]
func (c *RoleController) DataScope(ctx *gin.Context) {
	fmt.Printf("RoleController.DataScope: 角色数据权限分配\n")

	var role model.SysRole
	if err := ctx.ShouldBindJSON(&role); err != nil {
		fmt.Printf("RoleController.DataScope: 参数绑定失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数格式错误")
		return
	}

	// 校验角色是否允许操作
	if err := c.roleService.CheckRoleAllowed(&role); err != nil {
		fmt.Printf("RoleController.DataScope: 角色操作校验失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)

	// 校验数据权限
	if err := c.roleService.CheckRoleDataScope(currentUser.User, role.RoleID); err != nil {
		fmt.Printf("RoleController.DataScope: 数据权限校验失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 设置更新者
	role.UpdateBy = currentUser.User.UserName

	// 修改数据权限
	if err := c.roleService.AuthDataScope(&role); err != nil {
		fmt.Printf("RoleController.DataScope: 修改数据权限失败: %v\n", err)
		response.ErrorWithMessage(ctx, "修改数据权限失败")
		return
	}

	fmt.Printf("RoleController.DataScope: 修改数据权限成功, RoleID=%d\n", role.RoleID)
	response.SuccessWithMessage(ctx, "修改成功")
}

// DeptTree 获取角色部门树列表 对应Java后端的deptTree方法
// @Router /system/role/deptTree/{roleId} [get]
func (c *RoleController) DeptTree(ctx *gin.Context) {
	fmt.Printf("RoleController.DeptTree: 获取角色部门树列表, URL=%s\n", ctx.Request.URL.Path)

	roleIdStr := ctx.Param("roleId")
	roleId, err := strconv.ParseInt(roleIdStr, 10, 64)
	if err != nil {
		response.ErrorWithMessage(ctx, "角色ID格式错误")
		return
	}

	// 获取角色已选中的部门ID列表
	checkedKeys, err := c.deptService.SelectDeptListByRoleId(roleId)
	if err != nil {
		fmt.Printf("RoleController.DeptTree: 查询角色部门失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询角色部门失败")
		return
	}

	// 获取所有部门树
	depts, err := c.deptService.SelectDeptTreeList(&model.SysDept{})
	if err != nil {
		fmt.Printf("RoleController.DeptTree: 查询部门树失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询部门树失败")
		return
	}

	// 按照Java后端的格式返回数据 对应Java后端的AjaxResult.success().put("checkedKeys", ...).put("depts", ...)
	result := response.AjaxSuccess().Put("checkedKeys", checkedKeys).Put("depts", depts)

	fmt.Printf("RoleController.DeptTree: 获取角色部门树列表成功, RoleID=%d, CheckedKeys=%v\n", roleId, checkedKeys)
	response.SendAjaxResult(ctx, result)
}

// OptionSelect 获取角色选择框列表 对应Java后端的optionselect方法
// @Router /system/role/optionselect [get]
func (c *RoleController) OptionSelect(ctx *gin.Context) {
	fmt.Printf("RoleController.OptionSelect: 获取角色选择框列表\n")

	// 查询所有角色
	roles, err := c.roleService.SelectRoleAll()
	if err != nil {
		fmt.Printf("RoleController.OptionSelect: 查询角色列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询角色列表失败")
		return
	}

	fmt.Printf("RoleController.OptionSelect: 查询角色选择框列表成功, 数量=%d\n", len(roles))
	response.SuccessWithData(ctx, roles)
}

// AllocatedList 查询已分配用户角色列表 对应Java后端的allocatedList方法
// @Router /system/role/authUser/allocatedList [get]
func (c *RoleController) AllocatedList(ctx *gin.Context) {
	fmt.Printf("RoleController.AllocatedList: 查询已分配用户角色列表\n")

	// 获取角色ID
	roleId, _ := strconv.ParseInt(ctx.Query("roleId"), 10, 64)
	if roleId <= 0 {
		response.ErrorWithMessage(ctx, "角色ID不能为空")
		return
	}

	// 获取分页参数
	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))

	// 构建查询条件 对应Java后端的SysUser user参数
	user := &model.SysUser{
		UserName:    ctx.Query("userName"),
		Phonenumber: ctx.Query("phonenumber"),
		RoleID:      &roleId, // 设置角色ID用于查询
	}

	// 获取当前登录用户信息（用于数据权限）
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "获取用户信息失败")
		return
	}
	currentUser := loginUser.(*model.LoginUser)

	// 应用数据权限 对应Java后端的@DataScope(deptAlias = "d", userAlias = "u")
	params := make(map[string]interface{})
	err := datascope.ApplyDataScope(currentUser.User, "d", "u", "system:role:list", params)
	if err != nil {
		fmt.Printf("RoleController.AllocatedList: 应用数据权限失败: %v\n", err)
		response.ErrorWithMessage(ctx, "数据权限校验失败")
		return
	}

	// 设置数据权限参数
	if user.Params == nil {
		user.Params = make(map[string]interface{})
	}
	for key, value := range params {
		user.Params[key] = value
	}

	// 查询已分配用户角色列表 对应Java后端的userService.selectAllocatedList(user)
	users, total, err := c.userService.SelectAllocatedList(currentUser.User, user, pageNum, pageSize)
	if err != nil {
		fmt.Printf("RoleController.AllocatedList: 查询已分配用户角色列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询失败: "+err.Error())
		return
	}

	// 使用Java后端兼容的TableDataInfo格式 对应Java后端的getDataTable(list)
	tableData := response.GetDataTable(users, total)
	fmt.Printf("RoleController.AllocatedList: 查询已分配用户角色列表成功, 总数=%d\n", total)
	response.SendTableDataInfo(ctx, tableData)
}

// UnallocatedList 查询未分配用户角色列表 对应Java后端的unallocatedList方法
// @Router /system/role/authUser/unallocatedList [get]
func (c *RoleController) UnallocatedList(ctx *gin.Context) {
	fmt.Printf("RoleController.UnallocatedList: 查询未分配用户角色列表\n")

	// 获取角色ID
	roleId, _ := strconv.ParseInt(ctx.Query("roleId"), 10, 64)
	if roleId <= 0 {
		response.ErrorWithMessage(ctx, "角色ID不能为空")
		return
	}

	// 获取分页参数
	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))

	// 构建查询条件 对应Java后端的SysUser user参数
	user := &model.SysUser{
		UserName:    ctx.Query("userName"),
		Phonenumber: ctx.Query("phonenumber"),
		RoleID:      &roleId, // 设置角色ID用于查询
	}

	// 获取当前登录用户信息（用于数据权限）
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "获取用户信息失败")
		return
	}
	currentUser := loginUser.(*model.LoginUser)

	// 应用数据权限 对应Java后端的@DataScope(deptAlias = "d", userAlias = "u")
	params := make(map[string]interface{})
	err := datascope.ApplyDataScope(currentUser.User, "d", "u", "system:role:list", params)
	if err != nil {
		fmt.Printf("RoleController.UnallocatedList: 应用数据权限失败: %v\n", err)
		response.ErrorWithMessage(ctx, "数据权限校验失败")
		return
	}

	// 设置数据权限参数
	if user.Params == nil {
		user.Params = make(map[string]interface{})
	}
	for key, value := range params {
		user.Params[key] = value
	}

	// 查询未分配用户角色列表 对应Java后端的userService.selectUnallocatedList(user)
	users, total, err := c.userService.SelectUnallocatedList(currentUser.User, user, pageNum, pageSize)
	if err != nil {
		fmt.Printf("RoleController.UnallocatedList: 查询未分配用户角色列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询失败: "+err.Error())
		return
	}

	// 使用Java后端兼容的TableDataInfo格式 对应Java后端的getDataTable(list)
	tableData := response.GetDataTable(users, total)
	fmt.Printf("RoleController.UnallocatedList: 查询未分配用户角色列表成功, 总数=%d\n", total)
	response.SendTableDataInfo(ctx, tableData)
}

// CancelAuthUser 取消授权用户 对应Java后端的cancelAuthUser方法
// @Router /system/role/authUser/cancel [put]
func (c *RoleController) CancelAuthUser(ctx *gin.Context) {
	fmt.Printf("RoleController.CancelAuthUser: 取消授权用户\n")

	// 使用临时结构体来处理字符串到int的转换
	var requestData struct {
		UserID interface{} `json:"userId"`
		RoleID interface{} `json:"roleId"`
	}

	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		fmt.Printf("RoleController.CancelAuthUser: 参数绑定失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数格式错误: "+err.Error())
		return
	}

	fmt.Printf("RoleController.CancelAuthUser: 接收到原始参数 UserID=%v, RoleID=%v\n", requestData.UserID, requestData.RoleID)

	// 转换UserID
	var userId int64
	switch v := requestData.UserID.(type) {
	case float64:
		userId = int64(v)
	case string:
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			userId = id
		} else {
			response.ErrorWithMessage(ctx, "用户ID格式错误")
			return
		}
	default:
		response.ErrorWithMessage(ctx, "用户ID类型错误")
		return
	}

	// 转换RoleID
	var roleId int64
	switch v := requestData.RoleID.(type) {
	case float64:
		roleId = int64(v)
	case string:
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			roleId = id
		} else {
			response.ErrorWithMessage(ctx, "角色ID格式错误")
			return
		}
	default:
		response.ErrorWithMessage(ctx, "角色ID类型错误")
		return
	}

	fmt.Printf("RoleController.CancelAuthUser: 转换后参数 UserID=%d, RoleID=%d\n", userId, roleId)

	// 校验参数
	if userId <= 0 || roleId <= 0 {
		fmt.Printf("RoleController.CancelAuthUser: 参数无效 UserID=%d, RoleID=%d\n", userId, roleId)
		response.ErrorWithMessage(ctx, "用户ID和角色ID不能为空")
		return
	}

	// 构建用户角色对象
	userRole := &model.SysUserRole{
		UserID: userId,
		RoleID: roleId,
	}

	// 取消授权用户 对应Java后端的roleService.deleteAuthUser(userRole)
	if err := c.roleService.DeleteAuthUser(userRole); err != nil {
		fmt.Printf("RoleController.CancelAuthUser: 取消授权失败: %v\n", err)
		response.ErrorWithMessage(ctx, "取消授权失败: "+err.Error())
		return
	}

	fmt.Printf("RoleController.CancelAuthUser: 取消授权成功, UserID=%d, RoleID=%d\n", userId, roleId)
	response.SuccessWithMessage(ctx, "取消授权成功")
}

// CancelAuthUserAll 批量取消授权用户 对应Java后端的cancelAuthUserAll方法
// @Router /system/role/authUser/cancelAll [put]
func (c *RoleController) CancelAuthUserAll(ctx *gin.Context) {
	fmt.Printf("RoleController.CancelAuthUserAll: 批量取消授权用户\n")

	// 获取角色ID
	roleId, _ := strconv.ParseInt(ctx.Query("roleId"), 10, 64)
	if roleId <= 0 {
		response.ErrorWithMessage(ctx, "角色ID不能为空")
		return
	}

	// 获取用户ID列表
	userIdsStr := ctx.Query("userIds")
	if userIdsStr == "" {
		response.ErrorWithMessage(ctx, "用户ID不能为空")
		return
	}

	var userIds []int64
	for _, idStr := range strings.Split(userIdsStr, ",") {
		if id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64); err == nil && id > 0 {
			userIds = append(userIds, id)
		}
	}

	if len(userIds) == 0 {
		response.ErrorWithMessage(ctx, "用户ID格式错误")
		return
	}

	// 批量取消授权
	if err := c.roleService.DeleteAuthUsers(roleId, userIds); err != nil {
		fmt.Printf("RoleController.CancelAuthUserAll: 批量取消授权失败: %v\n", err)
		response.ErrorWithMessage(ctx, "取消授权失败")
		return
	}

	fmt.Printf("RoleController.CancelAuthUserAll: 批量取消授权成功, RoleID=%d, UserIDs=%v\n", roleId, userIds)
	response.SuccessWithMessage(ctx, "取消授权成功")
}

// SelectAuthUserAll 批量选择授权用户 对应Java后端的selectAuthUserAll方法
// @Router /system/role/authUser/selectAll [put]
func (c *RoleController) SelectAuthUserAll(ctx *gin.Context) {
	fmt.Printf("RoleController.SelectAuthUserAll: 批量选择授权用户\n")

	// 获取角色ID
	roleId, _ := strconv.ParseInt(ctx.Query("roleId"), 10, 64)
	if roleId <= 0 {
		response.ErrorWithMessage(ctx, "角色ID不能为空")
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)

	// 校验数据权限
	if err := c.roleService.CheckRoleDataScope(currentUser.User, roleId); err != nil {
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 获取用户ID列表
	userIdsStr := ctx.Query("userIds")
	if userIdsStr == "" {
		response.ErrorWithMessage(ctx, "用户ID不能为空")
		return
	}

	var userIds []int64
	for _, idStr := range strings.Split(userIdsStr, ",") {
		if id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64); err == nil && id > 0 {
			userIds = append(userIds, id)
		}
	}

	if len(userIds) == 0 {
		response.ErrorWithMessage(ctx, "用户ID格式错误")
		return
	}

	// 批量选择授权
	if err := c.roleService.InsertAuthUsers(roleId, userIds); err != nil {
		fmt.Printf("RoleController.SelectAuthUserAll: 批量选择授权失败: %v\n", err)
		response.ErrorWithMessage(ctx, "授权失败")
		return
	}

	fmt.Printf("RoleController.SelectAuthUserAll: 批量选择授权成功, RoleID=%d, UserIDs=%v\n", roleId, userIds)
	response.SuccessWithMessage(ctx, "授权成功")
}

// Export 导出角色数据 对应Java后端的export方法
// @Router /system/role/export [post]
func (c *RoleController) Export(ctx *gin.Context) {
	// 权限验证 - 对应Java后端的@PreAuthorize("@ss.hasPermi('system:role:export')")
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}

	user := loginUser.(*model.LoginUser)
	hasPermission := user.User.IsAdmin()
	if !hasPermission {
		for _, perm := range user.Permissions {
			if perm == "system:role:export" || perm == "system:role:*" {
				hasPermission = true
				break
			}
		}
	}

	if !hasPermission {
		response.ErrorWithMessage(ctx, "权限不足")
		return
	}

	fmt.Printf("RoleController.Export: 导出角色数据开始，用户: %s\n", user.User.UserName)

	// 解析POST请求表单参数 对应Java后端的SysRole role参数绑定
	formParams, err := export.ParseFormParams(ctx)
	if err != nil {
		fmt.Printf("RoleController.Export: 解析表单参数失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数解析失败: "+err.Error())
		return
	}

	// 解析角色查询参数
	queryParams := export.ParseRoleQueryParams(formParams)

	// 构建查询条件 对应Java后端的SysRole对象
	role := &model.SysRole{
		RoleName: queryParams.RoleName,
		Status:   queryParams.Status,
		RoleKey:  queryParams.RoleKey,
	}

	fmt.Printf("RoleController.Export: 查询条件 - RoleName=%s, Status=%s, RoleKey=%s, BeginTime=%v, EndTime=%v\n",
		role.RoleName, role.Status, role.RoleKey, queryParams.BeginTime, queryParams.EndTime)

	// 查询所有符合条件的角色（不分页）
	roles, err := c.roleService.SelectRoleListAll(role)
	if err != nil {
		fmt.Printf("RoleController.Export: 查询角色数据失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询失败: "+err.Error())
		return
	}

	// 使用Excel工具类导出 对应Java后端的ExcelUtil<SysRole> util = new ExcelUtil<SysRole>(SysRole.class)
	fmt.Printf("RoleController.Export: 开始创建Excel工具类\n")
	excelUtil := excel.NewExcelUtil()
	fmt.Printf("RoleController.Export: Excel工具类创建成功，开始导出角色数据\n")

	fileData, err := excelUtil.ExportExcel(roles, "角色数据", "角色列表")
	if err != nil {
		fmt.Printf("RoleController.Export: 导出Excel失败: %v\n", err)
		response.ErrorWithMessage(ctx, "导出失败: "+err.Error())
		return
	}
	fmt.Printf("RoleController.Export: Excel导出成功，文件大小: %d bytes\n", len(fileData))

	// 生成带日期时间的中文文件名
	now := time.Now()
	filename := fmt.Sprintf("角色数据导出_%s.xlsx", now.Format("20060102_150405"))

	// 设置响应头 对应Java后端的util.exportExcel(response, list, "角色数据")
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(filename)))
	ctx.Header("Content-Length", strconv.Itoa(len(fileData)))

	// 返回文件数据
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileData)

	fmt.Printf("RoleController.Export: 导出角色数据成功, 数量=%d, 文件大小=%d bytes\n", len(roles), len(fileData))

	// 记录操作日志 - 对应Java后端的@Log(title = "角色管理", businessType = BusinessType.EXPORT)
	operlog.RecordOperLog(ctx, "角色管理", "导出", fmt.Sprintf("导出角色成功，数量: %d", len(roles)), true)
}
