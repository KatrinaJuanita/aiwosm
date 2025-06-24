package tool

import (
	"fmt"
	"strconv"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// TestController 测试控制器 对应Java后端的TestController
type TestController struct {
	users map[int]*UserEntity
}

// NewTestController 创建测试控制器实例
func NewTestController() *TestController {
	// 初始化测试数据
	users := make(map[int]*UserEntity)
	users[1] = &UserEntity{
		UserID:   1,
		Username: "admin",
		Password: "admin123",
		Mobile:   "15888888888",
	}
	users[2] = &UserEntity{
		UserID:   2,
		Username: "ry",
		Password: "admin123",
		Mobile:   "15666666666",
	}

	return &TestController{
		users: users,
	}
}

// UserEntity 用户实体 对应Java后端的UserEntity
type UserEntity struct {
	UserID   int    `json:"userId" binding:"required"`   // 用户ID
	Username string `json:"username" binding:"required"` // 用户名称
	Password string `json:"password" binding:"required"` // 用户密码
	Mobile   string `json:"mobile"`                      // 用户手机
}

// UserList 获取用户列表 对应Java后端的userList方法
// @Summary 获取用户列表
// @Description 获取测试用户列表
// @Tags 测试接口
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]UserEntity}
// @Router /test/user/list [get]
func (c *TestController) UserList(ctx *gin.Context) {
	fmt.Printf("TestController.UserList: 获取用户列表\n")

	var userList []UserEntity
	for _, user := range c.users {
		userList = append(userList, *user)
	}

	response.SuccessWithData(ctx, userList)
}

// GetUser 获取用户详细 对应Java后端的getUser方法
// @Summary 获取用户详细
// @Description 根据用户ID获取用户详细信息
// @Tags 测试接口
// @Accept json
// @Produce json
// @Param userId path int true "用户ID"
// @Success 200 {object} response.Response{data=UserEntity}
// @Router /test/user/{userId} [get]
func (c *TestController) GetUser(ctx *gin.Context) {
	userIdStr := ctx.Param("userId")
	fmt.Printf("TestController.GetUser: 获取用户详细, UserId=%s\n", userIdStr)

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		response.ErrorWithMessage(ctx, "用户ID格式错误")
		return
	}

	user, exists := c.users[userId]
	if !exists {
		response.ErrorWithMessage(ctx, "用户不存在")
		return
	}

	response.SuccessWithData(ctx, user)
}

// SaveUser 新增用户 对应Java后端的save方法
// @Summary 新增用户
// @Description 新增测试用户
// @Tags 测试接口
// @Accept json
// @Produce json
// @Param user body UserEntity true "用户信息"
// @Success 200 {object} response.Response
// @Router /test/user/save [post]
func (c *TestController) SaveUser(ctx *gin.Context) {
	fmt.Printf("TestController.SaveUser: 新增用户\n")

	var user UserEntity
	if err := ctx.ShouldBindJSON(&user); err != nil {
		response.ErrorWithMessage(ctx, "参数绑定失败: "+err.Error())
		return
	}

	if user.UserID == 0 {
		response.ErrorWithMessage(ctx, "用户ID不能为空")
		return
	}

	// 检查用户是否已存在
	if _, exists := c.users[user.UserID]; exists {
		response.ErrorWithMessage(ctx, "用户ID已存在")
		return
	}

	c.users[user.UserID] = &user
	response.SuccessWithMessage(ctx, "新增成功")
}

// UpdateUser 更新用户 对应Java后端的update方法
// @Summary 更新用户
// @Description 更新测试用户信息
// @Tags 测试接口
// @Accept json
// @Produce json
// @Param user body UserEntity true "用户信息"
// @Success 200 {object} response.Response
// @Router /test/user/update [put]
func (c *TestController) UpdateUser(ctx *gin.Context) {
	fmt.Printf("TestController.UpdateUser: 更新用户\n")

	var user UserEntity
	if err := ctx.ShouldBindJSON(&user); err != nil {
		response.ErrorWithMessage(ctx, "参数绑定失败: "+err.Error())
		return
	}

	if user.UserID == 0 {
		response.ErrorWithMessage(ctx, "用户ID不能为空")
		return
	}

	// 检查用户是否存在
	if _, exists := c.users[user.UserID]; !exists {
		response.ErrorWithMessage(ctx, "用户不存在")
		return
	}

	c.users[user.UserID] = &user
	response.SuccessWithMessage(ctx, "更新成功")
}

// DeleteUser 删除用户信息 对应Java后端的delete方法
// @Summary 删除用户信息
// @Description 根据用户ID删除用户
// @Tags 测试接口
// @Accept json
// @Produce json
// @Param userId path int true "用户ID"
// @Success 200 {object} response.Response
// @Router /test/user/{userId} [delete]
func (c *TestController) DeleteUser(ctx *gin.Context) {
	userIdStr := ctx.Param("userId")
	fmt.Printf("TestController.DeleteUser: 删除用户, UserId=%s\n", userIdStr)

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		response.ErrorWithMessage(ctx, "用户ID格式错误")
		return
	}

	// 检查用户是否存在
	if _, exists := c.users[userId]; !exists {
		response.ErrorWithMessage(ctx, "用户不存在")
		return
	}

	delete(c.users, userId)
	response.SuccessWithMessage(ctx, "删除成功")
}

// GetTestInfo 获取测试信息
// @Summary 获取测试信息
// @Description 获取API测试相关信息
// @Tags 测试接口
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Router /test/info [get]
func (c *TestController) GetTestInfo(ctx *gin.Context) {
	fmt.Printf("TestController.GetTestInfo: 获取测试信息\n")

	info := map[string]interface{}{
		"name":        "WOSM API测试接口",
		"description": "用于测试API功能的模拟接口",
		"version":     "1.0.0",
		"userCount":   len(c.users),
		"endpoints": []string{
			"GET /test/user/list - 获取用户列表",
			"GET /test/user/{userId} - 获取用户详细",
			"POST /test/user/save - 新增用户",
			"PUT /test/user/update - 更新用户",
			"DELETE /test/user/{userId} - 删除用户",
		},
	}

	response.SuccessWithData(ctx, info)
}

// ResetTestData 重置测试数据
// @Summary 重置测试数据
// @Description 重置测试用户数据到初始状态
// @Tags 测试接口
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /test/reset [post]
func (c *TestController) ResetTestData(ctx *gin.Context) {
	fmt.Printf("TestController.ResetTestData: 重置测试数据\n")

	// 重置为初始数据
	c.users = make(map[int]*UserEntity)
	c.users[1] = &UserEntity{
		UserID:   1,
		Username: "admin",
		Password: "admin123",
		Mobile:   "15888888888",
	}
	c.users[2] = &UserEntity{
		UserID:   2,
		Username: "ry",
		Password: "admin123",
		Mobile:   "15666666666",
	}

	response.SuccessWithMessage(ctx, "测试数据重置成功")
}
