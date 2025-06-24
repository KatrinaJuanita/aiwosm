package system

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"wosm/internal/repository/model"
	"wosm/internal/service/system"
	"wosm/internal/utils"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// ProfileController 个人信息控制器 对应Java后端的SysProfileController
type ProfileController struct {
	userService    *system.UserService
	operLogService *system.OperLogService
}

// NewProfileController 创建个人信息控制器实例
func NewProfileController() *ProfileController {
	return &ProfileController{
		userService:    system.NewUserService(),
		operLogService: system.NewOperLogService(),
	}
}

// Profile 个人信息 对应Java后端的profile方法
// @Summary 获取个人信息
// @Description 获取当前登录用户的个人信息
// @Tags 个人中心
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=model.SysUser}
// @Router /system/user/profile [get]
func (c *ProfileController) Profile(ctx *gin.Context) {
	fmt.Printf("ProfileController.Profile: 获取个人信息\n")

	// 获取当前登录用户
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "获取用户信息失败")
		return
	}

	currentUser := loginUser.(*model.LoginUser)
	user := currentUser.User

	// 查询用户角色组
	roleGroup, err := c.userService.SelectUserRoleGroup(user.UserName)
	if err != nil {
		fmt.Printf("ProfileController.Profile: 查询用户角色组失败: %v\n", err)
		roleGroup = ""
	}

	// 查询用户岗位组
	postGroup, err := c.userService.SelectUserPostGroup(user.UserName)
	if err != nil {
		fmt.Printf("ProfileController.Profile: 查询用户岗位组失败: %v\n", err)
		postGroup = ""
	}

	// 构建响应数据
	data := map[string]interface{}{
		"user":      user,
		"roleGroup": roleGroup,
		"postGroup": postGroup,
	}

	response.SuccessWithData(ctx, data)
}

// UpdateProfile 修改用户 对应Java后端的updateProfile方法
// @Summary 修改个人信息
// @Description 修改当前登录用户的个人信息
// @Tags 个人中心
// @Accept json
// @Produce json
// @Param user body model.SysUser true "用户信息"
// @Success 200 {object} response.Response
// @Router /system/user/profile [put]
func (c *ProfileController) UpdateProfile(ctx *gin.Context) {
	fmt.Printf("ProfileController.UpdateProfile: 修改个人信息\n")

	// 获取当前登录用户
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "获取用户信息失败")
		return
	}

	currentUser := loginUser.(*model.LoginUser)
	user := currentUser.User

	// 绑定请求参数
	var updateUser model.SysUser
	if err := ctx.ShouldBindJSON(&updateUser); err != nil {
		response.ErrorWithMessage(ctx, "参数绑定失败: "+err.Error())
		return
	}

	// 更新允许修改的字段
	user.NickName = updateUser.NickName
	user.Email = updateUser.Email
	user.Phonenumber = updateUser.Phonenumber
	user.Sex = updateUser.Sex

	// 检查手机号唯一性
	if user.Phonenumber != "" {
		isUnique := c.userService.CheckPhoneUnique(user)
		if !isUnique {
			response.ErrorWithMessage(ctx, fmt.Sprintf("修改用户'%s'失败，手机号码已存在", user.UserName))
			c.recordOperLog(ctx, "个人信息", "修改", "修改个人信息失败: 手机号码已存在", false)
			return
		}
	}

	// 检查邮箱唯一性
	if user.Email != "" && !c.userService.CheckEmailUnique(user) {
		response.ErrorWithMessage(ctx, fmt.Sprintf("修改用户'%s'失败，邮箱账号已存在", user.UserName))
		c.recordOperLog(ctx, "个人信息", "修改", "修改个人信息失败: 邮箱账号已存在", false)
		return
	}

	// 更新用户信息
	err := c.userService.UpdateUserProfile(user)
	if err != nil {
		response.ErrorWithMessage(ctx, "修改个人信息异常，请联系管理员: "+err.Error())
		c.recordOperLog(ctx, "个人信息", "修改", "修改个人信息失败: "+err.Error(), false)
		return
	}

	// 更新缓存中的用户信息
	ctx.Set("loginUser", currentUser)

	// 记录操作日志
	c.recordOperLog(ctx, "个人信息", "修改", "修改个人信息成功", true)

	response.SuccessWithMessage(ctx, "修改成功")
}

// UpdatePwd 重置密码 对应Java后端的updatePwd方法
// @Summary 重置密码
// @Description 修改当前登录用户的密码
// @Tags 个人中心
// @Accept json
// @Produce json
// @Param params body map[string]string true "密码参数"
// @Success 200 {object} response.Response
// @Router /system/user/profile/updatePwd [put]
func (c *ProfileController) UpdatePwd(ctx *gin.Context) {
	fmt.Printf("ProfileController.UpdatePwd: 重置密码\n")

	// 获取当前登录用户
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "获取用户信息失败")
		return
	}

	currentUser := loginUser.(*model.LoginUser)
	user := currentUser.User

	// 绑定请求参数
	var params map[string]string
	if err := ctx.ShouldBindJSON(&params); err != nil {
		response.ErrorWithMessage(ctx, "参数绑定失败: "+err.Error())
		return
	}

	oldPassword := params["oldPassword"]
	newPassword := params["newPassword"]

	if oldPassword == "" {
		response.ErrorWithMessage(ctx, "旧密码不能为空")
		return
	}

	if newPassword == "" {
		response.ErrorWithMessage(ctx, "新密码不能为空")
		return
	}

	// 验证旧密码
	if !utils.CheckBcryptPassword(oldPassword, user.Password) {
		response.ErrorWithMessage(ctx, "修改密码失败，旧密码错误")
		c.recordOperLog(ctx, "个人信息", "修改", "修改密码失败: 旧密码错误", false)
		return
	}

	// 检查新密码是否与旧密码相同
	if utils.CheckBcryptPassword(newPassword, user.Password) {
		response.ErrorWithMessage(ctx, "新密码不能与旧密码相同")
		c.recordOperLog(ctx, "个人信息", "修改", "修改密码失败: 新密码不能与旧密码相同", false)
		return
	}

	// 加密新密码
	hashedPassword, err := utils.BcryptPassword(newPassword)
	if err != nil {
		response.ErrorWithMessage(ctx, "密码加密失败: "+err.Error())
		c.recordOperLog(ctx, "个人信息", "修改", "修改密码失败: 密码加密失败", false)
		return
	}

	// 更新密码
	err = c.userService.ResetUserPwd(user.UserID, hashedPassword)
	if err != nil {
		response.ErrorWithMessage(ctx, "修改密码异常，请联系管理员: "+err.Error())
		c.recordOperLog(ctx, "个人信息", "修改", "修改密码失败: "+err.Error(), false)
		return
	}

	// 更新缓存中的用户密码
	user.Password = hashedPassword
	ctx.Set("loginUser", currentUser)

	// 记录操作日志
	c.recordOperLog(ctx, "个人信息", "修改", "修改密码成功", true)

	response.SuccessWithMessage(ctx, "修改成功")
}

// Avatar 头像上传 对应Java后端的avatar方法
// @Summary 头像上传
// @Description 上传用户头像
// @Tags 个人中心
// @Accept multipart/form-data
// @Produce json
// @Param avatarfile formData file true "头像文件"
// @Success 200 {object} response.Response{data=map[string]string}
// @Router /system/user/profile/avatar [post]
func (c *ProfileController) Avatar(ctx *gin.Context) {
	fmt.Printf("ProfileController.Avatar: 头像上传\n")

	// 获取当前登录用户
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "获取用户信息失败")
		return
	}

	currentUser := loginUser.(*model.LoginUser)
	user := currentUser.User

	// 获取上传的文件
	file, err := ctx.FormFile("avatarfile")
	if err != nil {
		response.ErrorWithMessage(ctx, "获取上传文件失败: "+err.Error())
		c.recordOperLog(ctx, "用户头像", "修改", "头像上传失败: 获取上传文件失败", false)
		return
	}

	// 检查文件大小（限制为2MB）
	if file.Size > 2*1024*1024 {
		response.ErrorWithMessage(ctx, "上传文件大小不能超过2MB")
		c.recordOperLog(ctx, "用户头像", "修改", "头像上传失败: 文件大小超过限制", false)
		return
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp"}
	isValidExt := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			isValidExt = true
			break
		}
	}
	if !isValidExt {
		response.ErrorWithMessage(ctx, "只允许上传jpg、jpeg、png、gif、bmp格式的图片")
		c.recordOperLog(ctx, "用户头像", "修改", "头像上传失败: 文件格式不支持", false)
		return
	}

	// 生成文件名
	filename := fmt.Sprintf("avatar_%d_%d%s", user.UserID, time.Now().Unix(), ext)

	// 保存文件
	uploadPath := "uploads/avatar/" + filename
	if err := ctx.SaveUploadedFile(file, uploadPath); err != nil {
		response.ErrorWithMessage(ctx, "保存文件失败: "+err.Error())
		c.recordOperLog(ctx, "用户头像", "修改", "头像上传失败: 保存文件失败", false)
		return
	}

	// 构建访问URL
	avatarURL := "/profile/avatar/" + filename

	// 更新用户头像
	err = c.userService.UpdateUserAvatar(user.UserID, avatarURL)
	if err != nil {
		response.ErrorWithMessage(ctx, "更新用户头像失败: "+err.Error())
		c.recordOperLog(ctx, "用户头像", "修改", "头像上传失败: 更新用户头像失败", false)
		return
	}

	// 删除旧头像文件（如果存在）
	if user.Avatar != "" && user.Avatar != avatarURL {
		oldAvatarPath := "uploads" + strings.TrimPrefix(user.Avatar, "/profile")
		// 这里可以添加删除旧文件的逻辑
		fmt.Printf("ProfileController.Avatar: 删除旧头像文件: %s\n", oldAvatarPath)
	}

	// 更新缓存中的用户头像
	user.Avatar = avatarURL
	ctx.Set("loginUser", currentUser)

	// 记录操作日志
	c.recordOperLog(ctx, "用户头像", "修改", "头像上传成功", true)

	// 返回结果
	data := map[string]string{
		"imgUrl": avatarURL,
	}
	response.SuccessWithData(ctx, data)
}

// recordOperLog 记录操作日志
func (c *ProfileController) recordOperLog(ctx *gin.Context, title, businessType, content string, success bool) {
	// 获取用户信息
	username, _ := ctx.Get("username")

	// 确定业务类型
	var businessTypeInt int
	switch businessType {
	case "修改":
		businessTypeInt = model.BusinessTypeUpdate
	default:
		businessTypeInt = model.BusinessTypeOther
	}

	// 构建操作日志
	now := time.Now()
	operLog := &model.SysOperLog{
		Title:         title,
		BusinessType:  businessTypeInt,
		Method:        ctx.Request.Method,
		RequestMethod: ctx.Request.Method,
		OperatorType:  model.OperatorTypeManage, // 后台用户
		OperName:      "",
		DeptName:      "",
		OperURL:       ctx.Request.URL.Path,
		OperIP:        ctx.ClientIP(),
		OperLocation:  "", // 可以通过IP获取地理位置
		OperParam:     content,
		JSONResult:    "",
		Status:        model.OperStatusFail, // 失败
		ErrorMsg:      "",
		OperTime:      &now,
	}

	if username != nil {
		operLog.OperName = username.(string)
	}

	if success {
		operLog.Status = model.OperStatusSuccess // 成功
	}

	// 异步记录日志，不影响主流程
	go func() {
		if err := c.operLogService.InsertOperLog(operLog); err != nil {
			fmt.Printf("记录操作日志失败: %v\n", err)
		}
	}()
}
