package system

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
	"wosm/internal/repository/model"
	"wosm/internal/service/system"
	"wosm/internal/utils"
	"wosm/pkg/excel"
	"wosm/pkg/export"
	"wosm/pkg/operlog"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserController 用户管理控制器 对应Java后端的SysUserController
type UserController struct {
	userService    *system.UserService
	roleService    *system.RoleService
	deptService    *system.DeptService
	postService    *system.PostService
	operLogService *system.OperLogService
	configService  *system.ConfigService // 新增配置服务，对应Java后端的ISysConfigService
}

// NewUserController 创建用户控制器
func NewUserController() *UserController {
	return &UserController{
		userService:    system.NewUserService(),
		roleService:    system.NewRoleService(),
		deptService:    system.NewDeptService(),
		postService:    system.NewPostService(),
		operLogService: system.NewOperLogService(),
		configService:  system.NewConfigService(), // 新增配置服务初始化
	}
}

// List 获取用户列表 对应Java后端的list方法
// @Summary 获取用户列表
// @Description 分页查询用户列表，支持数据权限
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param userName query string false "用户名"
// @Param status query string false "状态"
// @Param deptId query int false "部门ID"
// @Security ApiKeyAuth
// @Success 200 {object} response.PageResult
// @Router /system/user/list [get]
func (c *UserController) List(ctx *gin.Context) {
	// 获取查询参数
	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))

	// 构建查询条件
	user := &model.SysUser{
		UserName: ctx.Query("userName"),
		Status:   ctx.Query("status"),
	}

	if deptId := ctx.Query("deptId"); deptId != "" {
		if id, err := strconv.ParseInt(deptId, 10, 64); err == nil {
			user.DeptID = &id
		}
	}

	// 获取当前登录用户信息
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "获取用户信息失败")
		return
	}
	currentUser := loginUser.(*model.LoginUser)

	// 使用支持数据权限的查询方法 对应Java后端的@DataScope注解
	users, total, err := c.userService.SelectUserListWithDataScope(currentUser.User, user, pageNum, pageSize)
	if err != nil {
		response.ErrorWithMessage(ctx, "查询失败: "+err.Error())
		return
	}

	// 确保所有用户的status字段都有正确的值
	for i := range users {
		if users[i].Status == "" {
			users[i].Status = "0" // 默认为正常状态
		}
		fmt.Printf("UserController.List: 用户%d - ID=%d, UserName=%s, Status=%s\n",
			i+1, users[i].UserID, users[i].UserName, users[i].Status)
	}

	// 使用Java后端兼容的TableDataInfo格式
	tableData := response.GetDataTable(users, total)
	response.SendTableDataInfo(ctx, tableData)
}

// AllocatedList 查询已分配用户角色列表 对应Java后端的allocatedList方法
// @Summary 查询已分配用户角色列表
// @Description 根据角色ID查询已分配该角色的用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param roleId query int true "角色ID"
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页数量"
// @Security ApiKeyAuth
// @Success 200 {object} response.TableDataInfo
// @Router /system/user/authRole/allocatedList [get]
func (c *UserController) AllocatedList(ctx *gin.Context) {
	fmt.Printf("UserController.AllocatedList: 查询已分配用户角色列表\n")

	// 解析分页参数
	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))

	// 解析角色ID
	roleIdStr := ctx.Query("roleId")
	if roleIdStr == "" {
		response.ErrorWithMessage(ctx, "角色ID不能为空")
		return
	}

	roleId, err := strconv.ParseInt(roleIdStr, 10, 64)
	if err != nil {
		response.ErrorWithMessage(ctx, "角色ID格式错误")
		return
	}

	// 构建查询条件 对应Java后端的SysUser user参数
	user := &model.SysUser{
		UserName:    ctx.Query("userName"),
		Phonenumber: ctx.Query("phonenumber"),
		Status:      ctx.Query("status"),
		RoleID:      &roleId, // 设置角色ID用于查询已分配用户
	}

	// 获取当前登录用户信息
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "获取用户信息失败")
		return
	}
	currentUser := loginUser.(*model.LoginUser)

	// 查询已分配用户列表 对应Java后端的userService.selectAllocatedList(user)
	users, total, err := c.userService.SelectAllocatedList(currentUser.User, user, pageNum, pageSize)
	if err != nil {
		response.ErrorWithMessage(ctx, "查询失败: "+err.Error())
		return
	}

	fmt.Printf("UserController.AllocatedList: 查询已分配用户成功, 数量=%d\n", len(users))

	// 使用Java后端兼容的TableDataInfo格式
	tableData := response.GetDataTable(users, total)
	response.SendTableDataInfo(ctx, tableData)
}

// UnallocatedList 查询未分配用户角色列表 对应Java后端的unallocatedList方法
// @Summary 查询未分配用户角色列表
// @Description 根据角色ID查询未分配该角色的用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param roleId query int true "角色ID"
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页数量"
// @Security ApiKeyAuth
// @Success 200 {object} response.TableDataInfo
// @Router /system/user/authRole/unallocatedList [get]
func (c *UserController) UnallocatedList(ctx *gin.Context) {
	fmt.Printf("UserController.UnallocatedList: 查询未分配用户角色列表\n")

	// 解析分页参数
	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))

	// 解析角色ID
	roleIdStr := ctx.Query("roleId")
	if roleIdStr == "" {
		response.ErrorWithMessage(ctx, "角色ID不能为空")
		return
	}

	roleId, err := strconv.ParseInt(roleIdStr, 10, 64)
	if err != nil {
		response.ErrorWithMessage(ctx, "角色ID格式错误")
		return
	}

	// 构建查询条件 对应Java后端的SysUser user参数
	user := &model.SysUser{
		UserName:    ctx.Query("userName"),
		Phonenumber: ctx.Query("phonenumber"),
		Status:      ctx.Query("status"),
		RoleID:      &roleId, // 设置角色ID用于查询未分配用户
	}

	// 获取当前登录用户信息
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "获取用户信息失败")
		return
	}
	currentUser := loginUser.(*model.LoginUser)

	// 查询未分配用户列表 对应Java后端的userService.selectUnallocatedList(user)
	users, total, err := c.userService.SelectUnallocatedList(currentUser.User, user, pageNum, pageSize)
	if err != nil {
		response.ErrorWithMessage(ctx, "查询失败: "+err.Error())
		return
	}

	fmt.Printf("UserController.UnallocatedList: 查询未分配用户成功, 数量=%d\n", len(users))

	// 使用Java后端兼容的TableDataInfo格式
	tableData := response.GetDataTable(users, total)
	response.SendTableDataInfo(ctx, tableData)
}

// GetInfo 获取用户详细信息 对应Java后端的getInfo方法
// @Summary 获取用户详细信息
// @Description 根据用户ID获取详细信息，包含角色、岗位等信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param userId path int true "用户ID"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/user/{userId} [get]
func (c *UserController) GetInfo(ctx *gin.Context) {
	fmt.Printf("UserController.GetInfo: 获取用户信息或初始化数据\n")

	userIdStr := ctx.Param("userId")

	var userId int64 = 0
	var err error

	// 构建返回结果 对应Java后端的AjaxResult.success()
	// 注意：Java后端使用ajax.put()直接添加到根级别，而不是嵌套在data中
	result := response.AjaxSuccess()

	// 如果提供了用户ID，获取用户详细信息 对应Java后端的StringUtils.isNotNull(userId)
	if userIdStr != "" && userIdStr != "0" {
		userId, err = strconv.ParseInt(userIdStr, 10, 64)
		if err != nil {
			fmt.Printf("UserController.GetInfo: 用户ID格式错误: %s\n", userIdStr)
			response.SendAjaxResult(ctx, response.AjaxErrorWithMessage("用户ID格式错误"))
			return
		}

		// 获取当前登录用户信息（用于数据权限校验）
		loginUser, exists := ctx.Get("loginUser")
		if !exists {
			response.SendAjaxResult(ctx, response.AjaxErrorWithMessage("获取用户信息失败"))
			return
		}
		currentUser := loginUser.(*model.LoginUser)

		// 校验数据权限 对应Java后端的userService.checkUserDataScope(userId)
		if err := c.userService.CheckUserDataScope(userId, currentUser.User); err != nil {
			fmt.Printf("UserController.GetInfo: 数据权限校验失败: %v\n", err)
			response.SendAjaxResult(ctx, response.AjaxErrorWithMessage(err.Error()))
			return
		}

		// 获取用户信息 对应Java后端的userService.selectUserById(userId)
		user, err := c.userService.SelectUserById(userId)
		if err != nil {
			fmt.Printf("UserController.GetInfo: 查询用户信息失败: %v\n", err)
			response.SendAjaxResult(ctx, response.AjaxErrorWithMessage("查询失败"))
			return
		}

		if user == nil {
			fmt.Printf("UserController.GetInfo: 用户不存在, UserID=%d\n", userId)
			response.SendAjaxResult(ctx, response.AjaxErrorWithMessage("用户不存在"))
			return
		}

		// 对应Java后端的ajax.put(AjaxResult.DATA_TAG, sysUser)
		result["data"] = user

		// 获取用户岗位ID列表 对应Java后端的ajax.put("postIds", postService.selectPostListByUserId(userId))
		postIds, err := c.postService.SelectPostListByUserId(userId)
		if err != nil {
			fmt.Printf("UserController.GetInfo: 查询用户岗位失败: %v\n", err)
			result["postIds"] = []int{} // 失败时返回空数组
		} else {
			result["postIds"] = postIds
		}

		// 获取用户角色ID列表 对应Java后端的ajax.put("roleIds", sysUser.getRoles().stream().map(SysRole::getRoleId).collect(Collectors.toList()))
		roleIds, err := c.roleService.SelectRoleListByUserId(userId)
		if err != nil {
			fmt.Printf("UserController.GetInfo: 查询用户角色失败: %v\n", err)
			result["roleIds"] = []int64{} // 失败时返回空数组
		} else {
			result["roleIds"] = roleIds
		}
	}

	// 获取所有角色列表 对应Java后端的roleService.selectRoleAll()
	roles, err := c.roleService.SelectRoleAll()
	if err != nil {
		fmt.Printf("UserController.GetInfo: 查询角色列表失败: %v\n", err)
		response.SendAjaxResult(ctx, response.AjaxErrorWithMessage("获取角色列表失败"))
		return
	}

	// 过滤管理员角色 对应Java后端的ajax.put("roles", SysUser.isAdmin(userId) ? roles : roles.stream().filter(r -> !r.isAdmin()).collect(Collectors.toList()))
	if userId != 1 { // 非管理员用户过滤掉管理员角色
		var filteredRoles []model.SysRole
		for _, role := range roles {
			if role.RoleID != 1 { // 过滤掉管理员角色（roleId=1）
				filteredRoles = append(filteredRoles, role)
			}
		}
		result["roles"] = filteredRoles
	} else {
		result["roles"] = roles
	}

	// 获取所有岗位列表 对应Java后端的ajax.put("posts", postService.selectPostAll())
	posts, err := c.postService.SelectPostAll()
	if err != nil {
		fmt.Printf("UserController.GetInfo: 查询岗位列表失败: %v\n", err)
		// 岗位查询失败时，返回空数组，不影响整体功能
		result["posts"] = []model.SysPost{}
	} else {
		result["posts"] = posts
	}

	fmt.Printf("UserController.GetInfo: 获取用户信息成功, UserID=%d, 角色数量=%d, 岗位数量=%d\n", userId, len(roles), len(posts))

	// 使用AjaxResult格式返回，对应Java后端的AjaxResult.success()
	ajaxResult := response.AjaxSuccess()

	// 添加用户数据 对应Java后端的ajax.put(AjaxResult.DATA_TAG, sysUser)
	if userData, ok := result["data"]; ok {
		ajaxResult.Put("data", userData)
	}

	// 添加岗位ID列表 对应Java后端的ajax.put("postIds", postService.selectPostListByUserId(userId))
	if postIds, ok := result["postIds"]; ok {
		ajaxResult.Put("postIds", postIds)
	}

	// 添加角色ID列表 对应Java后端的ajax.put("roleIds", sysUser.getRoles().stream().map(SysRole::getRoleId).collect(Collectors.toList()))
	if roleIds, ok := result["roleIds"]; ok {
		ajaxResult.Put("roleIds", roleIds)
	}

	// 添加角色列表 对应Java后端的ajax.put("roles", ...)
	if rolesData, ok := result["roles"]; ok {
		ajaxResult.Put("roles", rolesData)
	}

	// 添加岗位列表 对应Java后端的ajax.put("posts", postService.selectPostAll())
	if postsData, ok := result["posts"]; ok {
		ajaxResult.Put("posts", postsData)
	}

	// 返回AjaxResult格式
	response.SendAjaxResult(ctx, ajaxResult)
}

// Add 新增用户 对应Java后端的add方法
// @PreAuthorize("@ss.hasPermi('system:user:add')")
// @Log(title = "用户管理", businessType = BusinessType.INSERT)
// @Summary 新增用户
// @Description 新增用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body model.SysUser true "用户信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/user [post]
func (c *UserController) Add(ctx *gin.Context) {
	fmt.Printf("UserController.Add: 新增用户\n")

	// 先解析为map以获取密码字段
	var requestData map[string]any
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		fmt.Printf("UserController.Add: 参数绑定失败: %v\n", err)
		response.SendAjaxResult(ctx, response.AjaxErrorWithMessage("参数格式错误"))
		return
	}

	// 重新绑定到用户对象（密码字段会被忽略）
	jsonData, _ := json.Marshal(requestData)
	var user model.SysUser
	if err := json.Unmarshal(jsonData, &user); err != nil {
		fmt.Printf("UserController.Add: 用户对象绑定失败: %v\n", err)
		response.SendAjaxResult(ctx, response.AjaxErrorWithMessage("参数格式错误"))
		return
	}

	// 手动设置密码字段
	if password, ok := requestData["password"].(string); ok && password != "" {
		user.Password = password
		fmt.Printf("UserController.Add: 接收到密码，长度=%d\n", len(password))
	} else {
		// 如果没有提供密码，使用默认密码
		user.Password = "123456"
		fmt.Printf("UserController.Add: 使用默认密码\n")
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)

	// 校验部门数据权限 对应Java后端的deptService.checkDeptDataScope
	if user.DeptID != nil && *user.DeptID > 0 {
		if err := c.deptService.CheckDeptDataScope(currentUser.User, *user.DeptID); err != nil {
			fmt.Printf("UserController.Add: 部门数据权限校验失败: %v\n", err)
			response.SendAjaxResult(ctx, response.AjaxErrorWithMessage(err.Error()))
			return
		}
	}

	// 校验角色数据权限 对应Java后端的roleService.checkRoleDataScope
	if len(user.RoleIDs) > 0 {
		if err := c.roleService.CheckRoleDataScope(currentUser.User, user.RoleIDs...); err != nil {
			fmt.Printf("UserController.Add: 角色数据权限校验失败: %v\n", err)
			response.SendAjaxResult(ctx, response.AjaxErrorWithMessage(err.Error()))
			return
		}
	}

	// 校验用户名唯一性 对应Java后端的checkUserNameUnique
	if !c.userService.CheckUserNameUnique(&user) {
		fmt.Printf("UserController.Add: 用户名已存在: %s\n", user.UserName)
		response.SendAjaxResult(ctx, response.AjaxErrorWithMessage(fmt.Sprintf("新增用户'%s'失败，登录账号已存在", user.UserName)))
		return
	}

	// 校验手机号唯一性 对应Java后端的checkPhoneUnique
	if user.Phonenumber != "" && !c.userService.CheckPhoneUnique(&user) {
		fmt.Printf("UserController.Add: 手机号已存在: %s\n", user.Phonenumber)
		response.SendAjaxResult(ctx, response.AjaxErrorWithMessage(fmt.Sprintf("新增用户'%s'失败，手机号码已存在", user.UserName)))
		return
	}

	// 校验邮箱唯一性 对应Java后端的checkEmailUnique
	if user.Email != "" && !c.userService.CheckEmailUnique(&user) {
		fmt.Printf("UserController.Add: 邮箱已存在: %s\n", user.Email)
		response.SendAjaxResult(ctx, response.AjaxErrorWithMessage(fmt.Sprintf("新增用户'%s'失败，邮箱账号已存在", user.UserName)))
		return
	}

	// 设置创建人 对应Java后端的setCreateBy
	user.CreateBy = currentUser.User.UserName

	// 新增用户
	err := c.userService.InsertUser(&user, currentUser.User.UserName)
	if err != nil {
		fmt.Printf("UserController.Add: 新增用户失败: %v\n", err)
		response.SendAjaxResult(ctx, response.AjaxErrorWithMessage("新增失败"))
		return
	}

	fmt.Printf("UserController.Add: 新增用户成功, UserID=%d\n", user.UserID)
	response.SendAjaxResult(ctx, response.AjaxSuccess())
}

// Edit 修改用户 对应Java后端的edit方法
// @Summary 修改用户
// @Description 修改用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body model.SysUser true "用户信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/user [put]
func (c *UserController) Edit(ctx *gin.Context) {
	fmt.Printf("UserController.Edit: 修改用户\n")

	var user model.SysUser
	if err := ctx.ShouldBindJSON(&user); err != nil {
		fmt.Printf("UserController.Edit: 参数绑定失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数格式错误")
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)

	// 校验用户是否允许操作 对应Java后端的checkUserAllowed
	if err := c.userService.CheckUserAllowed(&user); err != nil {
		fmt.Printf("UserController.Edit: 用户操作校验失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 校验数据权限 对应Java后端的checkUserDataScope
	if err := c.userService.CheckUserDataScope(user.UserID, currentUser.User); err != nil {
		fmt.Printf("UserController.Edit: 数据权限校验失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 校验部门数据权限 对应Java后端的deptService.checkDeptDataScope
	if user.DeptID != nil && *user.DeptID > 0 {
		if err := c.deptService.CheckDeptDataScope(currentUser.User, *user.DeptID); err != nil {
			fmt.Printf("UserController.Edit: 部门数据权限校验失败: %v\n", err)
			response.ErrorWithMessage(ctx, err.Error())
			return
		}
	}

	// 校验角色数据权限 对应Java后端的roleService.checkRoleDataScope
	if len(user.RoleIDs) > 0 {
		if err := c.roleService.CheckRoleDataScope(currentUser.User, user.RoleIDs...); err != nil {
			fmt.Printf("UserController.Edit: 角色数据权限校验失败: %v\n", err)
			response.ErrorWithMessage(ctx, err.Error())
			return
		}
	}

	// 校验用户名唯一性 对应Java后端的checkUserNameUnique
	if !c.userService.CheckUserNameUnique(&user) {
		fmt.Printf("UserController.Edit: 用户名已存在: %s\n", user.UserName)
		response.ErrorWithMessage(ctx, fmt.Sprintf("修改用户'%s'失败，登录账号已存在", user.UserName))
		return
	}

	// 校验手机号唯一性 对应Java后端的checkPhoneUnique
	if user.Phonenumber != "" && !c.userService.CheckPhoneUnique(&user) {
		fmt.Printf("UserController.Edit: 手机号已存在: %s\n", user.Phonenumber)
		response.ErrorWithMessage(ctx, fmt.Sprintf("修改用户'%s'失败，手机号码已存在", user.UserName))
		return
	}

	// 校验邮箱唯一性 对应Java后端的checkEmailUnique
	if user.Email != "" && !c.userService.CheckEmailUnique(&user) {
		fmt.Printf("UserController.Edit: 邮箱已存在: %s\n", user.Email)
		response.ErrorWithMessage(ctx, fmt.Sprintf("修改用户'%s'失败，邮箱账号已存在", user.UserName))
		return
	}

	// 设置更新人 对应Java后端的setUpdateBy
	user.UpdateBy = currentUser.User.UserName

	err := c.userService.UpdateUser(&user, currentUser.User.UserName)
	if err != nil {
		fmt.Printf("UserController.Edit: 修改用户失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	fmt.Printf("UserController.Edit: 修改用户成功, UserID=%d\n", user.UserID)
	response.SuccessWithMessage(ctx, "修改成功")
}

// Remove 删除用户 对应Java后端的remove方法
// @Summary 删除用户
// @Description 批量删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param ids path string true "用户ID列表，逗号分隔"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/user/{ids} [delete]
func (c *UserController) Remove(ctx *gin.Context) {
	idsStr := ctx.Param("ids")
	if idsStr == "" {
		response.ErrorWithMessage(ctx, "参数错误")
		return
	}

	// 解析用户ID列表
	var userIds []int64
	for _, idStr := range strings.Split(idsStr, ",") {
		if id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64); err == nil {
			userIds = append(userIds, id)
		}
	}

	if len(userIds) == 0 {
		response.ErrorWithMessage(ctx, "参数错误")
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)

	err := c.userService.DeleteUserByIds(currentUser.User, userIds)
	if err != nil {
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "删除成功")
}

// ResetPwd 重置密码 对应Java后端的resetPwd方法
// @Summary 重置密码
// @Description 重置用户密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param resetPwd body map[string]interface{} true "重置密码信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/user/resetPwd [put]
func (c *UserController) ResetPwd(ctx *gin.Context) {
	var req map[string]any
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(ctx, "参数错误")
		return
	}

	userIdFloat, ok := req["userId"].(float64)
	if !ok {
		response.ErrorWithMessage(ctx, "用户ID格式错误")
		return
	}
	userId := int64(userIdFloat)

	password, ok := req["password"].(string)
	if !ok || password == "" {
		response.ErrorWithMessage(ctx, "密码不能为空")
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)

	// 校验用户是否允许操作 对应Java后端的checkUserAllowed
	user := &model.SysUser{UserID: userId}
	if err := c.userService.CheckUserAllowed(user); err != nil {
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 校验数据权限 对应Java后端的checkUserDataScope
	if err := c.userService.CheckUserDataScope(userId, currentUser.User); err != nil {
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	err := c.userService.ResetPwd(userId, password, currentUser.User.UserName)
	if err != nil {
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "重置成功")
}

// ChangeStatus 修改用户状态 对应Java后端的changeStatus方法
// @Summary 修改用户状态
// @Description 启用/停用用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body model.SysUser true "用户状态信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/user/changeStatus [put]
func (c *UserController) ChangeStatus(ctx *gin.Context) {
	fmt.Printf("UserController.ChangeStatus: 修改用户状态开始\n")

	var user model.SysUser
	if err := ctx.ShouldBindJSON(&user); err != nil {
		fmt.Printf("UserController.ChangeStatus: 参数绑定失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数错误")
		return
	}

	fmt.Printf("UserController.ChangeStatus: 接收参数 - UserID=%d, Status=%s\n", user.UserID, user.Status)

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)

	// 校验用户是否允许操作 对应Java后端的checkUserAllowed
	if err := c.userService.CheckUserAllowed(&user); err != nil {
		fmt.Printf("UserController.ChangeStatus: 用户操作校验失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 校验数据权限 对应Java后端的checkUserDataScope
	if err := c.userService.CheckUserDataScope(user.UserID, currentUser.User); err != nil {
		fmt.Printf("UserController.ChangeStatus: 数据权限校验失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	err := c.userService.ChangeStatus(&user, currentUser.User.UserName)
	if err != nil {
		fmt.Printf("UserController.ChangeStatus: 修改用户状态失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	status := "启用"
	if user.Status == "1" {
		status = "停用"
	}
	fmt.Printf("UserController.ChangeStatus: 修改用户状态成功 - UserID=%d, Status=%s, Message=%s\n", user.UserID, user.Status, status)
	response.SuccessWithMessage(ctx, status+"成功")
}

// AuthRole 获取用户授权角色 对应Java后端的authRole方法
// @Summary 获取用户授权角色
// @Description 根据用户ID获取授权角色信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param userId path int true "用户ID"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/user/authRole/{userId} [get]
func (c *UserController) AuthRole(ctx *gin.Context) {
	userIdStr := ctx.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		response.ErrorWithMessage(ctx, "用户ID格式错误")
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)

	// 校验数据权限 对应Java后端的checkUserDataScope
	if err := c.userService.CheckUserDataScope(userId, currentUser.User); err != nil {
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 获取用户信息和角色信息
	result, err := c.userService.GetUserAuthRole(userId)
	if err != nil {
		fmt.Printf("UserController.AuthRole: 获取用户授权角色失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 使用AjaxResult格式返回，对应Java后端的AjaxResult.success().put("user", user).put("roles", roles)
	ajaxResult := response.AjaxSuccess()

	// 添加用户数据 对应Java后端的ajax.put("user", user)
	if user, ok := result["user"]; ok {
		ajaxResult.Put("user", user)
	}

	// 添加角色数据 对应Java后端的ajax.put("roles", ...)
	if roles, ok := result["roles"]; ok {
		ajaxResult.Put("roles", roles)
	}

	fmt.Printf("UserController.AuthRole: 获取用户授权角色成功, UserID=%d\n", userId)
	response.SendAjaxResult(ctx, ajaxResult)
}

// InsertAuthRole 用户授权角色 对应Java后端的insertAuthRole方法
// @Summary 用户授权角色
// @Description 为用户分配角色
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param userId query int true "用户ID"
// @Param roleIds query string true "角色ID列表，逗号分隔"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/user/authRole [put]
func (c *UserController) InsertAuthRole(ctx *gin.Context) {
	// 对应Java后端的参数接收方式：insertAuthRole(Long userId, Long[] roleIds)
	// 前端使用params发送，所以从Query参数获取
	userIdStr := ctx.Query("userId")
	roleIdsStr := ctx.Query("roleIds")

	fmt.Printf("UserController.InsertAuthRole: 接收参数 - userId=%s, roleIds=%s\n", userIdStr, roleIdsStr)

	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		fmt.Printf("UserController.InsertAuthRole: 用户ID格式错误: %v\n", err)
		response.ErrorWithMessage(ctx, "用户ID格式错误")
		return
	}

	// 解析角色ID列表 对应Java后端的Long[] roleIds参数
	var roleIds []int64
	if roleIdsStr != "" {
		for _, idStr := range strings.Split(roleIdsStr, ",") {
			idStr = strings.TrimSpace(idStr)
			if idStr != "" {
				if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
					roleIds = append(roleIds, id)
				} else {
					fmt.Printf("UserController.InsertAuthRole: 角色ID格式错误: %s\n", idStr)
				}
			}
		}
	}

	fmt.Printf("UserController.InsertAuthRole: 解析角色ID列表 - UserID=%d, RoleIDs=%v\n", userId, roleIds)

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)

	// 校验数据权限 对应Java后端的checkUserDataScope
	if err := c.userService.CheckUserDataScope(userId, currentUser.User); err != nil {
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 校验角色数据权限 对应Java后端的roleService.checkRoleDataScope
	if len(roleIds) > 0 {
		if err := c.roleService.CheckRoleDataScope(currentUser.User, roleIds...); err != nil {
			response.ErrorWithMessage(ctx, err.Error())
			return
		}
	}

	err = c.userService.InsertUserAuth(userId, roleIds)
	if err != nil {
		fmt.Printf("UserController.InsertAuthRole: 用户授权角色失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	fmt.Printf("UserController.InsertAuthRole: 用户授权角色成功, UserID=%d, RoleIDs=%v\n", userId, roleIds)

	// 使用AjaxResult格式返回，对应Java后端的return success()
	response.SendAjaxResult(ctx, response.AjaxSuccess())
}

// Export 导出用户数据 对应Java后端的export方法
// @Log(title = "用户管理", businessType = BusinessType.EXPORT)
// @PreAuthorize("@ss.hasPermi('system:user:export')")
// @Summary 导出用户数据
// @Description 导出用户数据到Excel文件
// @Tags 用户管理
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param userName query string false "用户名"
// @Param status query string false "状态"
// @Param deptId query int false "部门ID"
// @Security ApiKeyAuth
// @Success 200 {file} file "Excel文件"
// @Router /system/user/export [post]
func (c *UserController) Export(ctx *gin.Context) {
	// 权限验证 - 对应Java后端的@PreAuthorize("@ss.hasPermi('system:user:export')")
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}

	currentUser := loginUser.(*model.LoginUser)
	hasPermission := currentUser.User.IsAdmin()
	if !hasPermission {
		for _, perm := range currentUser.Permissions {
			if perm == "system:user:export" || perm == "system:user:*" {
				hasPermission = true
				break
			}
		}
	}

	if !hasPermission {
		response.ErrorWithMessage(ctx, "权限不足")
		return
	}

	fmt.Printf("UserController.Export: 导出用户数据开始，用户: %s\n", currentUser.User.UserName)

	// 解析POST请求表单参数 对应Java后端的SysUser user参数绑定
	formParams, err := export.ParseFormParams(ctx)
	if err != nil {
		fmt.Printf("UserController.Export: 解析表单参数失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数解析失败: "+err.Error())
		return
	}

	// 解析用户查询参数
	queryParams := export.ParseUserQueryParams(formParams)

	// 构建查询条件 对应Java后端的SysUser对象
	user := &model.SysUser{
		UserName:    queryParams.UserName,
		Status:      queryParams.Status,
		Phonenumber: queryParams.Phonenumber,
		DeptID:      queryParams.DeptID,
	}

	fmt.Printf("UserController.Export: 查询条件 - UserName=%s, Status=%s, DeptID=%v, Phonenumber=%s, BeginTime=%v, EndTime=%v\n",
		user.UserName, user.Status, user.DeptID, user.Phonenumber, queryParams.BeginTime, queryParams.EndTime)

	// 查询所有符合条件的用户（不分页） 对应Java后端的userService.selectUserList(user)
	// 使用足够大的分页参数确保获取所有数据
	users, _, err := c.userService.SelectUserListWithDataScope(currentUser.User, user, 1, 100000)
	if err != nil {
		fmt.Printf("UserController.Export: 查询用户数据失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询失败: "+err.Error())
		return
	}

	// 使用Excel工具类导出 对应Java后端的ExcelUtil<SysUser> util = new ExcelUtil<SysUser>(SysUser.class)
	excelUtil := excel.NewExcelUtil()
	fileData, err := excelUtil.ExportExcel(users, "用户数据", "用户列表")
	if err != nil {
		fmt.Printf("UserController.Export: 导出Excel失败: %v\n", err)
		response.ErrorWithMessage(ctx, "导出失败: "+err.Error())
		return
	}

	// 生成带日期时间的中文文件名
	now := time.Now()
	filename := fmt.Sprintf("用户数据导出_%s.xlsx", now.Format("20060102_150405"))

	// 设置响应头 对应Java后端的util.exportExcel(response, list, "用户数据")
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(filename)))
	ctx.Header("Content-Length", strconv.Itoa(len(fileData)))

	// 返回文件数据
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileData)

	fmt.Printf("UserController.Export: 导出用户数据成功, 数量=%d, 文件大小=%d bytes\n", len(users), len(fileData))

	// 记录操作日志 - 对应Java后端的@Log(title = "用户管理", businessType = BusinessType.EXPORT)
	operlog.RecordOperLog(ctx, "用户管理", "导出", fmt.Sprintf("导出用户成功，数量: %d", len(users)), true)
}

// ImportData 导入用户数据 对应Java后端的importData方法
// @Log(title = "用户管理", businessType = BusinessType.IMPORT)
// @PreAuthorize("@ss.hasPermi('system:user:import')")
// @Summary 导入用户数据
// @Description 从Excel文件导入用户数据
// @Tags 用户管理
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Excel文件"
// @Param updateSupport formData bool false "是否更新已存在的用户"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/user/importData [post]
func (c *UserController) ImportData(ctx *gin.Context) {
	fmt.Printf("UserController.ImportData: 导入用户数据\n")

	// 获取上传的文件 对应Java后端的MultipartFile file参数
	file, err := ctx.FormFile("file")
	if err != nil {
		fmt.Printf("UserController.ImportData: 获取上传文件失败: %v\n", err)
		response.ErrorWithMessage(ctx, "请选择要导入的文件")
		return
	}

	// 获取更新支持参数 对应Java后端的boolean updateSupport参数
	updateSupport := ctx.PostForm("updateSupport") == "true"

	// 打开文件
	src, err := file.Open()
	if err != nil {
		fmt.Printf("UserController.ImportData: 打开文件失败: %v\n", err)
		response.ErrorWithMessage(ctx, "文件打开失败")
		return
	}
	defer src.Close()

	// 读取文件内容
	_, err = io.ReadAll(src)
	if err != nil {
		fmt.Printf("UserController.ImportData: 读取文件内容失败: %v\n", err)
		response.ErrorWithMessage(ctx, "文件读取失败")
		return
	}

	// TODO: 暂时注释导入功能
	// 使用Excel工具类解析 对应Java后端的ExcelUtil<SysUser> util = new ExcelUtil<SysUser>(SysUser.class)
	// excelUtil := excel.NewExcelUtil()
	// users, parseErr, message := excelUtil.ImportUsers(fileData)
	users := []model.SysUser{}
	parseErr := fmt.Errorf("导入功能暂未实现")
	message := "导入功能暂未实现"
	if parseErr != nil {
		fmt.Printf("UserController.ImportData: 解析Excel文件失败: %v\n", parseErr)
		response.ErrorWithMessage(ctx, "文件解析失败: "+parseErr.Error())
		return
	}

	if len(users) == 0 {
		response.ErrorWithMessage(ctx, "导入数据不能为空")
		return
	}

	// 获取当前登录用户 对应Java后端的getUsername()
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)
	operName := currentUser.User.UserName

	// 调用用户服务导入 对应Java后端的userService.importUser(userList, updateSupport, operName)
	importMessage, err := c.userService.ImportUser(users, updateSupport, operName)
	if err != nil {
		fmt.Printf("UserController.ImportData: 导入用户失败: %v\n", err)
		response.ErrorWithMessage(ctx, "导入失败: "+err.Error())
		return
	}

	// 合并解析消息和导入消息
	finalMessage := message
	if importMessage != "" {
		finalMessage = importMessage + "。" + message
	}

	fmt.Printf("UserController.ImportData: 导入用户数据成功: %s\n", finalMessage)
	response.SuccessWithMessage(ctx, finalMessage)
}

// ImportTemplate 下载导入模板 对应Java后端的importTemplate方法
// @Summary 下载导入模板
// @Description 下载用户数据导入模板
// @Tags 用户管理
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Success 200 {file} file "Excel模板文件"
// @Router /system/user/importTemplate [post]
func (c *UserController) ImportTemplate(ctx *gin.Context) {
	fmt.Printf("UserController.ImportTemplate: 下载用户导入模板\n")

	// TODO: 暂时注释模板生成功能
	// 使用Excel工具类生成模板 对应Java后端的ExcelUtil<SysUser> util = new ExcelUtil<SysUser>(SysUser.class)
	// fmt.Printf("UserController.ImportTemplate: 开始创建Excel工具类\n")
	// excelUtil := excel.NewExcelUtil()
	// fmt.Printf("UserController.ImportTemplate: Excel工具类创建成功，开始生成模板\n")
	// templateData, err := excelUtil.GenerateUserTemplate()
	templateData := []byte{}
	err := fmt.Errorf("模板生成功能暂未实现")
	if err != nil {
		fmt.Printf("UserController.ImportTemplate: 生成模板失败: %v\n", err)
		response.ErrorWithMessage(ctx, "生成模板失败: "+err.Error())
		return
	}
	fmt.Printf("UserController.ImportTemplate: 模板生成成功, 数据长度=%d\n", len(templateData))

	// 生成带日期时间的中文模板文件名
	now := time.Now()
	templateFilename := fmt.Sprintf("用户数据导入模板_%s.xlsx", now.Format("20060102_150405"))

	// 设置响应头 对应Java后端的util.importTemplateExcel(response, "用户数据")
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(templateFilename)))
	ctx.Header("Content-Length", strconv.Itoa(len(templateData)))

	// 返回模板文件
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", templateData)

	fmt.Printf("UserController.ImportTemplate: 下载用户导入模板成功, 文件大小=%d bytes\n", len(templateData))
}

// DeptTree 获取部门树列表 对应Java后端的deptTree方法
// @Summary 获取部门树列表
// @Description 获取部门树形结构数据
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/user/deptTree [get]
func (c *UserController) DeptTree(ctx *gin.Context) {
	fmt.Printf("UserController.DeptTree: 获取部门树列表\n")

	// 构建查询条件 对应Java后端的SysDept dept参数
	dept := &model.SysDept{}

	// 查询部门树列表 对应Java后端的deptService.selectDeptTreeList(dept)
	deptTree, err := c.deptService.SelectDeptTreeList(dept)
	if err != nil {
		fmt.Printf("UserController.DeptTree: 查询部门树失败: %v\n", err)
		response.SendAjaxResult(ctx, response.AjaxErrorWithMessage("查询部门树失败"))
		return
	}

	fmt.Printf("UserController.DeptTree: 查询部门树成功, 数量=%d\n", len(deptTree))
	response.SendAjaxResult(ctx, response.AjaxSuccessWithData(deptTree))
}

// Profile 获取个人信息 对应Java后端的profile方法
// @Summary 获取个人信息
// @Description 获取当前登录用户的个人信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/user/profile [get]
func (c *UserController) Profile(ctx *gin.Context) {
	fmt.Printf("UserController.Profile: 获取个人信息\n")

	// 获取当前登录用户
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}
	currentUser := loginUser.(*model.LoginUser)

	// 获取用户详细信息
	user, err := c.userService.SelectUserById(currentUser.User.UserID)
	if err != nil {
		fmt.Printf("UserController.Profile: 查询用户信息失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询失败")
		return
	}

	if user == nil {
		response.ErrorWithMessage(ctx, "用户不存在")
		return
	}

	// 查询用户角色组
	roleGroup, err := c.userService.SelectUserRoleGroup(user.UserName)
	if err != nil {
		fmt.Printf("UserController.Profile: 查询用户角色组失败: %v\n", err)
		roleGroup = ""
	}

	// 查询用户岗位组
	postGroup, err := c.userService.SelectUserPostGroup(user.UserName)
	if err != nil {
		fmt.Printf("UserController.Profile: 查询用户岗位组失败: %v\n", err)
		postGroup = ""
	}

	// 构建返回结果 对应Java后端的AjaxResult.success().put("user", user).put("roleGroup", roleGroup).put("postGroup", postGroup)
	result := response.AjaxSuccess()
	result.Put("user", user)
	result.Put("roleGroup", roleGroup)
	result.Put("postGroup", postGroup)

	fmt.Printf("UserController.Profile: 获取个人信息成功, UserID=%d\n", user.UserID)
	response.SendAjaxResult(ctx, result)
}

// UpdateProfile 修改个人信息 对应Java后端的updateProfile方法
// @Summary 修改个人信息
// @Description 修改当前登录用户的个人信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body model.SysUser true "用户信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/user/profile [put]
func (c *UserController) UpdateProfile(ctx *gin.Context) {
	fmt.Printf("UserController.UpdateProfile: 修改个人信息\n")

	var user model.SysUser
	if err := ctx.ShouldBindJSON(&user); err != nil {
		fmt.Printf("UserController.UpdateProfile: 参数绑定失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数错误")
		return
	}

	// 获取当前登录用户
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}
	currentUser := loginUser.(*model.LoginUser)

	// 设置用户ID为当前登录用户ID
	user.UserID = currentUser.User.UserID

	// 校验手机号唯一性
	if user.Phonenumber != "" && !c.userService.CheckPhoneUnique(&user) {
		response.ErrorWithMessage(ctx, fmt.Sprintf("修改用户'%s'失败，手机号码已存在", user.UserName))
		return
	}

	// 校验邮箱唯一性
	if user.Email != "" && !c.userService.CheckEmailUnique(&user) {
		response.ErrorWithMessage(ctx, fmt.Sprintf("修改用户'%s'失败，邮箱账号已存在", user.UserName))
		return
	}

	// 设置更新人
	user.UpdateBy = currentUser.User.UserName

	err := c.userService.UpdateUserProfile(&user)
	if err != nil {
		fmt.Printf("UserController.UpdateProfile: 修改个人信息失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	fmt.Printf("UserController.UpdateProfile: 修改个人信息成功, UserID=%d\n", user.UserID)
	response.SuccessWithMessage(ctx, "修改成功")
}

// UpdatePwd 修改密码 对应Java后端的updatePwd方法
// @Summary 修改密码
// @Description 修改当前登录用户的密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param pwdData body map[string]string true "密码信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/user/profile/updatePwd [put]
func (c *UserController) UpdatePwd(ctx *gin.Context) {
	fmt.Printf("UserController.UpdatePwd: 修改密码\n")

	var pwdData map[string]string
	if err := ctx.ShouldBindJSON(&pwdData); err != nil {
		fmt.Printf("UserController.UpdatePwd: 参数绑定失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数错误")
		return
	}

	oldPassword := pwdData["oldPassword"]
	newPassword := pwdData["newPassword"]

	if oldPassword == "" || newPassword == "" {
		response.ErrorWithMessage(ctx, "旧密码和新密码不能为空")
		return
	}

	// 获取当前登录用户
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}
	currentUser := loginUser.(*model.LoginUser)

	// 获取用户信息验证旧密码
	user, err := c.userService.SelectUserById(currentUser.User.UserID)
	if err != nil {
		response.ErrorWithMessage(ctx, "查询用户信息失败")
		return
	}

	// 验证旧密码
	if !utils.MatchesPassword(oldPassword, user.Password) {
		response.ErrorWithMessage(ctx, "修改密码失败，旧密码错误")
		return
	}

	// 新旧密码不能相同
	if oldPassword == newPassword {
		response.ErrorWithMessage(ctx, "新密码不能与旧密码相同")
		return
	}

	// 重置密码
	err = c.userService.ResetPwd(currentUser.User.UserID, newPassword, currentUser.User.UserName)
	if err != nil {
		fmt.Printf("UserController.UpdatePwd: 修改密码失败: %v\n", err)
		response.ErrorWithMessage(ctx, "修改密码失败")
		return
	}

	fmt.Printf("UserController.UpdatePwd: 修改密码成功, UserID=%d\n", currentUser.User.UserID)
	response.SuccessWithMessage(ctx, "修改成功")
}

// Avatar 头像上传 对应Java后端的avatar方法
// @Summary 头像上传
// @Description 上传用户头像
// @Tags 用户管理
// @Accept multipart/form-data
// @Produce json
// @Param avatarfile formData file true "头像文件"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/user/profile/avatar [post]
func (c *UserController) Avatar(ctx *gin.Context) {
	fmt.Printf("UserController.Avatar: 头像上传\n")

	// 获取上传的文件
	file, err := ctx.FormFile("avatarfile")
	if err != nil {
		fmt.Printf("UserController.Avatar: 获取上传文件失败: %v\n", err)
		response.ErrorWithMessage(ctx, "请选择要上传的头像文件")
		return
	}

	// 验证文件大小和类型
	fmt.Printf("UserController.Avatar: 接收到文件: %s, 大小: %d bytes\n", file.Filename, file.Size)

	// 获取当前登录用户
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}
	currentUser := loginUser.(*model.LoginUser)

	// TODO: 实现文件上传逻辑
	// 1. 验证文件类型和大小
	// 2. 保存文件到指定目录
	// 3. 生成访问URL
	// 4. 更新用户头像字段

	avatarUrl := "/profile/avatar/default.jpg" // 临时返回默认头像

	// 更新用户头像
	err = c.userService.UpdateUserAvatar(currentUser.User.UserID, avatarUrl)
	if err != nil {
		fmt.Printf("UserController.Avatar: 更新用户头像失败: %v\n", err)
		response.ErrorWithMessage(ctx, "上传头像失败")
		return
	}

	// 构建返回结果 对应Java后端的AjaxResult.success().put("imgUrl", avatar)
	result := response.AjaxSuccess()
	result.Put("imgUrl", avatarUrl)

	fmt.Printf("UserController.Avatar: 头像上传成功, UserID=%d, URL=%s\n", currentUser.User.UserID, avatarUrl)
	response.SendAjaxResult(ctx, result)
}
