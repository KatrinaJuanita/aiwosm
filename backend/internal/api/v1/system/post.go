package system

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"wosm/internal/repository/model"
	"wosm/internal/service/system"
	"wosm/pkg/excel"
	"wosm/pkg/export"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// PostController 岗位管理控制器 对应Java后端的SysPostController
type PostController struct {
	postService    *system.PostService
	operLogService *system.OperLogService
}

// NewPostController 创建岗位管理控制器实例
func NewPostController() *PostController {
	return &PostController{
		postService:    system.NewPostService(),
		operLogService: system.NewOperLogService(),
	}
}

// List 获取岗位列表 对应Java后端的list方法
// @Summary 获取岗位列表
// @Description 获取岗位列表，支持分页和条件查询
// @Tags 岗位管理
// @Accept json
// @Produce json
// @Param postCode query string false "岗位编码"
// @Param postName query string false "岗位名称"
// @Param status query string false "状态"
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.TableDataInfo
// @Router /system/post/list [get]
func (c *PostController) List(ctx *gin.Context) {
	fmt.Printf("PostController.List: 获取岗位列表\n")

	// 绑定查询参数
	var params model.PostQueryParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		response.ErrorWithMessage(ctx, "参数绑定失败: "+err.Error())
		return
	}

	// 构建查询条件
	post := &model.SysPost{
		PostCode: params.PostCode,
		PostName: params.PostName,
		Status:   params.Status,
	}

	// 检查是否需要分页
	if params.PageNum > 0 && params.PageSize > 0 {
		// 分页查询
		posts, total, err := c.postService.SelectPostListWithPage(post, params.PageNum, params.PageSize)
		if err != nil {
			response.ErrorWithMessage(ctx, "查询岗位列表失败: "+err.Error())
			return
		}

		// 使用Java后端兼容的TableDataInfo格式
		tableData := response.GetDataTable(posts, total)
		response.SendTableDataInfo(ctx, tableData)
	} else {
		// 不分页查询（用于下拉框等场景）
		posts, err := c.postService.SelectPostList(post)
		if err != nil {
			response.ErrorWithMessage(ctx, "查询岗位列表失败: "+err.Error())
			return
		}

		// 使用Java后端兼容的TableDataInfo格式
		tableData := response.GetDataTable(posts, int64(len(posts)))
		response.SendTableDataInfo(ctx, tableData)
	}
}

// GetInfo 根据岗位编号获取详细信息 对应Java后端的getInfo方法
// @Summary 获取岗位详情
// @Description 根据岗位ID获取岗位详细信息
// @Tags 岗位管理
// @Accept json
// @Produce json
// @Param postId path int true "岗位ID"
// @Success 200 {object} response.Response{data=model.SysPost}
// @Router /system/post/{postId} [get]
func (c *PostController) GetInfo(ctx *gin.Context) {
	postIdStr := ctx.Param("postId")
	fmt.Printf("PostController.GetInfo: 获取岗位详情, PostId=%s\n", postIdStr)

	postId, err := strconv.ParseInt(postIdStr, 10, 64)
	if err != nil {
		response.ErrorWithMessage(ctx, "岗位ID格式错误")
		return
	}

	post, err := c.postService.SelectPostById(postId)
	if err != nil {
		response.ErrorWithMessage(ctx, "查询岗位详情失败: "+err.Error())
		return
	}

	if post == nil {
		response.ErrorWithMessage(ctx, "岗位不存在")
		return
	}

	response.SuccessWithData(ctx, post)
}

// Add 新增岗位 对应Java后端的add方法
// @Summary 新增岗位
// @Description 新增岗位
// @Tags 岗位管理
// @Accept json
// @Produce json
// @Param post body model.SysPost true "岗位信息"
// @Success 200 {object} response.Response
// @Router /system/post [post]
func (c *PostController) Add(ctx *gin.Context) {
	fmt.Printf("PostController.Add: 新增岗位\n")

	var post model.SysPost
	if err := ctx.ShouldBindJSON(&post); err != nil {
		response.ErrorWithMessage(ctx, "参数绑定失败: "+err.Error())
		return
	}

	// 获取当前用户信息
	username, exists := ctx.Get("username")
	if !exists {
		response.ErrorWithMessage(ctx, "获取用户信息失败")
		return
	}
	post.CreateBy = username.(string)

	// 新增岗位
	err := c.postService.InsertPost(&post)
	if err != nil {
		response.ErrorWithMessage(ctx, err.Error())

		// 记录操作日志
		c.recordOperLog(ctx, "岗位管理", "新增", fmt.Sprintf("新增岗位失败: %s", err.Error()), false)
		return
	}

	// 记录操作日志
	c.recordOperLog(ctx, "岗位管理", "新增", fmt.Sprintf("新增岗位成功，名称: %s", post.PostName), true)

	response.SuccessWithMessage(ctx, "新增成功")
}

// Edit 修改岗位 对应Java后端的edit方法
// @Summary 修改岗位
// @Description 修改岗位
// @Tags 岗位管理
// @Accept json
// @Produce json
// @Param post body model.SysPost true "岗位信息"
// @Success 200 {object} response.Response
// @Router /system/post [put]
func (c *PostController) Edit(ctx *gin.Context) {
	fmt.Printf("PostController.Edit: 修改岗位\n")

	var post model.SysPost
	if err := ctx.ShouldBindJSON(&post); err != nil {
		response.ErrorWithMessage(ctx, "参数绑定失败: "+err.Error())
		return
	}

	// 获取当前用户信息
	username, exists := ctx.Get("username")
	if !exists {
		response.ErrorWithMessage(ctx, "获取用户信息失败")
		return
	}
	post.UpdateBy = username.(string)

	// 修改岗位
	err := c.postService.UpdatePost(&post)
	if err != nil {
		response.ErrorWithMessage(ctx, err.Error())

		// 记录操作日志
		c.recordOperLog(ctx, "岗位管理", "修改", fmt.Sprintf("修改岗位失败: %s", err.Error()), false)
		return
	}

	// 记录操作日志
	c.recordOperLog(ctx, "岗位管理", "修改", fmt.Sprintf("修改岗位成功，ID: %d", post.PostID), true)

	response.SuccessWithMessage(ctx, "修改成功")
}

// Remove 删除岗位 对应Java后端的remove方法
// @Summary 删除岗位
// @Description 删除岗位，支持批量删除
// @Tags 岗位管理
// @Accept json
// @Produce json
// @Param postIds path string true "岗位ID列表，多个ID用逗号分隔"
// @Success 200 {object} response.Response
// @Router /system/post/{postIds} [delete]
func (c *PostController) Remove(ctx *gin.Context) {
	postIdsStr := ctx.Param("postIds")
	fmt.Printf("PostController.Remove: 删除岗位, PostIds=%s\n", postIdsStr)

	if postIdsStr == "" {
		response.ErrorWithMessage(ctx, "岗位ID不能为空")
		return
	}

	// 解析岗位ID列表
	idStrings := strings.Split(postIdsStr, ",")
	var postIds []int64
	for _, idStr := range idStrings {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err != nil {
			response.ErrorWithMessage(ctx, "岗位ID格式错误: "+idStr)
			return
		}
		postIds = append(postIds, id)
	}

	// 删除岗位
	err := c.postService.DeletePostByIds(postIds)
	if err != nil {
		response.ErrorWithMessage(ctx, err.Error())

		// 记录操作日志
		c.recordOperLog(ctx, "岗位管理", "删除", fmt.Sprintf("删除岗位失败: %s", err.Error()), false)
		return
	}

	// 记录操作日志
	c.recordOperLog(ctx, "岗位管理", "删除", fmt.Sprintf("删除岗位成功，ID: %v", postIds), true)

	response.SuccessWithMessage(ctx, "删除成功")
}

// OptionSelect 获取岗位选择框列表 对应Java后端的optionselect方法
// @Summary 获取岗位选择框列表
// @Description 获取岗位选择框列表
// @Tags 岗位管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]model.SysPost}
// @Router /system/post/optionselect [get]
func (c *PostController) OptionSelect(ctx *gin.Context) {
	fmt.Printf("PostController.OptionSelect: 获取岗位选择框列表\n")

	posts, err := c.postService.GetPostOptionSelect()
	if err != nil {
		response.ErrorWithMessage(ctx, "查询岗位选择框列表失败: "+err.Error())
		return
	}

	response.SuccessWithData(ctx, posts)
}

// Export 导出岗位 对应Java后端的export方法
// @Summary 导出岗位
// @Description 导出岗位数据
// @Tags 岗位管理
// @Accept json
// @Produce application/vnd.ms-excel
// @Param postCode query string false "岗位编码"
// @Param postName query string false "岗位名称"
// @Param status query string false "状态"
// @Success 200 {file} file "Excel文件"
// @Router /system/post/export [post]
func (c *PostController) Export(ctx *gin.Context) {
	// 权限验证 - 对应Java后端的@PreAuthorize("@ss.hasPermi('system:post:export')")
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}

	currentUser := loginUser.(*model.LoginUser)
	hasPermission := currentUser.User.IsAdmin()
	if !hasPermission {
		for _, perm := range currentUser.Permissions {
			if perm == "system:post:export" || perm == "system:post:*" {
				hasPermission = true
				break
			}
		}
	}

	if !hasPermission {
		response.ErrorWithMessage(ctx, "权限不足")
		return
	}

	fmt.Printf("PostController.Export: 导出岗位数据开始，用户: %s\n", currentUser.User.UserName)

	// 解析POST请求表单参数 对应Java后端的SysPost post参数绑定
	formParams, err := export.ParseFormParams(ctx)
	if err != nil {
		fmt.Printf("PostController.Export: 解析表单参数失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数解析失败: "+err.Error())
		return
	}

	// 解析岗位查询参数
	queryParams := export.ParsePostQueryParams(formParams)

	// 构建查询条件 对应Java后端的SysPost对象
	post := &model.SysPost{
		PostCode: queryParams.PostCode,
		PostName: queryParams.PostName,
		Status:   queryParams.Status,
	}

	fmt.Printf("PostController.Export: 查询条件 - PostCode=%s, PostName=%s, Status=%s, BeginTime=%v, EndTime=%v\n",
		post.PostCode, post.PostName, post.Status, queryParams.BeginTime, queryParams.EndTime)

	// 查询所有符合条件的岗位
	posts, err := c.postService.SelectPostList(post)
	if err != nil {
		fmt.Printf("PostController.Export: 查询岗位数据失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询失败: "+err.Error())
		return
	}

	// 使用Excel工具类导出 对应Java后端的ExcelUtil<SysPost> util = new ExcelUtil<SysPost>(SysPost.class)
	excelUtil := excel.NewExcelUtil()
	fileData, err := excelUtil.ExportExcel(posts, "岗位数据", "岗位列表")
	if err != nil {
		fmt.Printf("PostController.Export: 导出Excel失败: %v\n", err)
		response.ErrorWithMessage(ctx, "导出失败: "+err.Error())
		return
	}

	// 生成带日期时间的中文文件名
	now := time.Now()
	filename := fmt.Sprintf("岗位数据导出_%s.xlsx", now.Format("20060102_150405"))

	// 设置响应头 对应Java后端的util.exportExcel(response, list, "岗位数据")
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(filename)))
	ctx.Header("Content-Length", strconv.Itoa(len(fileData)))

	// 返回文件数据
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileData)

	fmt.Printf("PostController.Export: 导出岗位数据成功, 数量=%d, 文件大小=%d bytes\n", len(posts), len(fileData))

	// 记录操作日志
	c.recordOperLog(ctx, "岗位管理", "导出", fmt.Sprintf("导出岗位成功，数量: %d", len(posts)), true)
}

// recordOperLog 记录操作日志
func (c *PostController) recordOperLog(ctx *gin.Context, title, businessType, content string, success bool) {
	// 获取用户信息
	username, _ := ctx.Get("username")

	// 确定业务类型
	var businessTypeInt int
	switch businessType {
	case "新增":
		businessTypeInt = model.BusinessTypeInsert
	case "修改":
		businessTypeInt = model.BusinessTypeUpdate
	case "删除":
		businessTypeInt = model.BusinessTypeDelete
	case "导出":
		businessTypeInt = model.BusinessTypeExport
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
