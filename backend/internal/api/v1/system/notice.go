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
	"wosm/pkg/operlog"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// NoticeController 通知公告控制器 对应Java后端的SysNoticeController
type NoticeController struct {
	noticeService  *system.NoticeService
	operLogService *system.OperLogService
}

// NewNoticeController 创建通知公告控制器实例
func NewNoticeController() *NoticeController {
	return &NoticeController{
		noticeService:  system.NewNoticeService(),
		operLogService: system.NewOperLogService(),
	}
}

// List 获取通知公告列表 对应Java后端的list方法
// @Summary 获取通知公告列表
// @Description 分页查询通知公告列表，支持权限控制和数据过滤
// @Tags 通知公告管理
// @Accept json
// @Produce json
// @Param noticeTitle query string false "公告标题"
// @Param noticeType query string false "公告类型"
// @Param createBy query string false "创建者"
// @Param status query string false "状态"
// @Param beginTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.TableDataInfo
// @Router /system/notice/list [get]
func (c *NoticeController) List(ctx *gin.Context) {
	fmt.Printf("NoticeController.List: 获取通知公告列表\n")

	// 权限验证：检查用户是否有查询权限 对应Java后端的@PreAuthorize("@ss.hasPermi('system:notice:list')")
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}
	currentUser := loginUser.(*model.LoginUser)

	// 检查权限
	hasPermission := false
	for _, permission := range currentUser.Permissions {
		if permission == "system:notice:list" || permission == "*:*:*" {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		response.ErrorWithMessage(ctx, "没有权限执行此操作")
		return
	}

	// 绑定查询参数
	var params model.NoticeQueryParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		response.ErrorWithMessage(ctx, "参数绑定失败: "+err.Error())
		return
	}

	// 参数验证和清理
	params.NoticeTitle = strings.TrimSpace(params.NoticeTitle)
	params.CreateBy = strings.TrimSpace(params.CreateBy)

	// 验证公告类型
	if params.NoticeType != "" && params.NoticeType != "1" && params.NoticeType != "2" {
		response.ErrorWithMessage(ctx, "公告类型参数无效")
		return
	}

	// 验证状态
	if params.Status != "" && params.Status != "0" && params.Status != "1" {
		response.ErrorWithMessage(ctx, "状态参数无效")
		return
	}

	// 设置默认分页参数并验证
	if params.PageNum <= 0 {
		params.PageNum = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}
	if params.PageSize > 100 {
		params.PageSize = 100 // 限制最大页面大小
	}

	// 数据权限控制：根据用户角色限制查询范围 对应Java后端的@DataScope注解
	if !currentUser.User.IsAdmin() {
		// 非管理员用户的数据权限控制
		// 1. 只能查看自己创建的公告
		// 2. 或者查看状态为正常的公告
		params.DataScope = "user"
		params.CurrentUserId = currentUser.User.UserID
		params.CurrentUserName = currentUser.User.UserName
	}

	// 查询公告列表
	notices, err := c.noticeService.SelectNoticeList(&params)
	if err != nil {
		response.ErrorWithMessage(ctx, "查询公告列表失败: "+err.Error())
		return
	}

	// 查询总数
	total, err := c.noticeService.CountNoticeList(&params)
	if err != nil {
		response.ErrorWithMessage(ctx, "查询公告总数失败: "+err.Error())
		return
	}

	// 数据后处理：为前端提供额外的显示信息
	for i := range notices {
		// 添加类型和状态的文本描述
		notices[i].NoticeTypeText = c.noticeService.GetNoticeTypeText(notices[i].NoticeType)
		notices[i].StatusText = c.noticeService.GetNoticeStatusText(notices[i].Status)

		// 安全处理：截断过长的内容用于列表显示
		if len(notices[i].NoticeContent) > 100 {
			notices[i].NoticeContentPreview = notices[i].NoticeContent[:100] + "..."
		} else {
			notices[i].NoticeContentPreview = notices[i].NoticeContent
		}
	}

	// 使用Java后端兼容的TableDataInfo格式
	tableData := response.GetDataTable(notices, total)
	response.SendTableDataInfo(ctx, tableData)
}

// GetInfo 根据通知公告编号获取详细信息 对应Java后端的getInfo方法
// @Summary 获取通知公告详情
// @Description 根据公告ID获取公告详细信息，支持权限验证
// @Tags 通知公告管理
// @Accept json
// @Produce json
// @Param noticeId path int true "公告ID"
// @Success 200 {object} response.Response{data=model.SysNotice}
// @Router /system/notice/{noticeId} [get]
func (c *NoticeController) GetInfo(ctx *gin.Context) {
	noticeIdStr := ctx.Param("noticeId")
	fmt.Printf("NoticeController.GetInfo: 获取公告详情, NoticeId=%s\n", noticeIdStr)

	// 权限验证：检查用户是否有查询权限 对应Java后端的@PreAuthorize("@ss.hasPermi('system:notice:query')")
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}
	currentUser := loginUser.(*model.LoginUser)

	// 检查权限
	hasPermission := false
	for _, permission := range currentUser.Permissions {
		if permission == "system:notice:query" || permission == "*:*:*" {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		response.ErrorWithMessage(ctx, "没有权限执行此操作")
		return
	}

	// 参数验证
	noticeId, err := strconv.ParseInt(noticeIdStr, 10, 64)
	if err != nil || noticeId <= 0 {
		response.ErrorWithMessage(ctx, "公告ID格式错误")
		return
	}

	// 查询公告详情
	notice, err := c.noticeService.SelectNoticeById(noticeId)
	if err != nil {
		response.ErrorWithMessage(ctx, "查询公告详情失败: "+err.Error())
		return
	}

	if notice == nil {
		response.ErrorWithMessage(ctx, "公告不存在")
		return
	}

	// 数据权限验证：非管理员只能查看自己创建的公告或已发布的公告
	if !currentUser.User.IsAdmin() {
		// 如果不是自己创建的公告，且公告状态为关闭，则不允许查看
		if notice.CreateBy != currentUser.User.UserName && notice.Status == model.NoticeStatusClosed {
			response.ErrorWithMessage(ctx, "无权限查看此公告")
			return
		}
	}

	// 添加扩展信息
	notice.NoticeTypeText = c.noticeService.GetNoticeTypeText(notice.NoticeType)
	notice.StatusText = c.noticeService.GetNoticeStatusText(notice.Status)

	response.SuccessWithData(ctx, notice)
}

// Add 新增通知公告 对应Java后端的add方法
// @Summary 新增通知公告
// @Description 新增通知公告，支持权限验证和数据校验
// @Tags 通知公告管理
// @Accept json
// @Produce json
// @Param notice body model.SysNotice true "公告信息"
// @Success 200 {object} response.Response
// @Router /system/notice [post]
func (c *NoticeController) Add(ctx *gin.Context) {
	fmt.Printf("NoticeController.Add: 新增通知公告\n")

	// 权限验证：检查用户是否有新增权限 对应Java后端的@PreAuthorize("@ss.hasPermi('system:notice:add')")
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}
	currentUser := loginUser.(*model.LoginUser)

	// 检查权限
	hasPermission := false
	for _, permission := range currentUser.Permissions {
		if permission == "system:notice:add" || permission == "*:*:*" {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		response.ErrorWithMessage(ctx, "没有权限执行此操作")
		return
	}

	// 参数绑定和验证
	var notice model.SysNotice
	if err := ctx.ShouldBindJSON(&notice); err != nil {
		response.ErrorWithMessage(ctx, "参数绑定失败: "+err.Error())
		return
	}

	// 数据验证
	if err := c.validateNoticeData(&notice, false); err != nil {
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 设置创建者信息
	notice.CreateBy = currentUser.User.UserName

	// 安全处理：清理和验证内容
	notice.NoticeTitle = strings.TrimSpace(notice.NoticeTitle)
	notice.NoticeContent = strings.TrimSpace(notice.NoticeContent)
	notice.Remark = strings.TrimSpace(notice.Remark)

	// XSS防护：检查标题和内容
	if c.containsXSSContent(notice.NoticeTitle) || c.containsXSSContent(notice.NoticeContent) {
		response.ErrorWithMessage(ctx, "公告内容包含不安全字符，请检查后重新提交")
		return
	}

	// 新增公告
	err := c.noticeService.InsertNotice(&notice)
	if err != nil {
		response.ErrorWithMessage(ctx, "新增公告失败: "+err.Error())

		// 记录操作日志
		c.recordOperLog(ctx, "通知公告", "新增", fmt.Sprintf("新增公告失败: %s", err.Error()), false)
		return
	}

	// 记录操作日志
	c.recordOperLog(ctx, "通知公告", "新增", fmt.Sprintf("新增公告成功，标题: %s", notice.NoticeTitle), true)

	response.SuccessWithMessage(ctx, "新增成功")
}

// Edit 修改通知公告 对应Java后端的edit方法
// @Summary 修改通知公告
// @Description 修改通知公告，支持权限验证和数据校验
// @Tags 通知公告管理
// @Accept json
// @Produce json
// @Param notice body model.SysNotice true "公告信息"
// @Success 200 {object} response.Response
// @Router /system/notice [put]
func (c *NoticeController) Edit(ctx *gin.Context) {
	fmt.Printf("NoticeController.Edit: 修改通知公告\n")

	// 权限验证：检查用户是否有修改权限 对应Java后端的@PreAuthorize("@ss.hasPermi('system:notice:edit')")
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}
	currentUser := loginUser.(*model.LoginUser)

	// 检查权限
	hasPermission := false
	for _, permission := range currentUser.Permissions {
		if permission == "system:notice:edit" || permission == "*:*:*" {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		response.ErrorWithMessage(ctx, "没有权限执行此操作")
		return
	}

	// 参数绑定和验证
	var notice model.SysNotice
	if err := ctx.ShouldBindJSON(&notice); err != nil {
		response.ErrorWithMessage(ctx, "参数绑定失败: "+err.Error())
		return
	}

	// 数据验证
	if err := c.validateNoticeData(&notice, true); err != nil {
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 检查公告是否存在
	existingNotice, err := c.noticeService.SelectNoticeById(notice.NoticeID)
	if err != nil {
		response.ErrorWithMessage(ctx, "查询公告失败: "+err.Error())
		return
	}
	if existingNotice == nil {
		response.ErrorWithMessage(ctx, "公告不存在")
		return
	}

	// 数据权限验证：非管理员只能修改自己创建的公告
	if !currentUser.User.IsAdmin() && existingNotice.CreateBy != currentUser.User.UserName {
		response.ErrorWithMessage(ctx, "无权限修改此公告")
		return
	}

	// 设置更新者信息
	notice.UpdateBy = currentUser.User.UserName

	// 安全处理：清理和验证内容
	notice.NoticeTitle = strings.TrimSpace(notice.NoticeTitle)
	notice.NoticeContent = strings.TrimSpace(notice.NoticeContent)
	notice.Remark = strings.TrimSpace(notice.Remark)

	// XSS防护：检查标题和内容
	if c.containsXSSContent(notice.NoticeTitle) || c.containsXSSContent(notice.NoticeContent) {
		response.ErrorWithMessage(ctx, "公告内容包含不安全字符，请检查后重新提交")
		return
	}

	// 修改公告
	err = c.noticeService.UpdateNotice(&notice)
	if err != nil {
		response.ErrorWithMessage(ctx, "修改公告失败: "+err.Error())

		// 记录操作日志
		c.recordOperLog(ctx, "通知公告", "修改", fmt.Sprintf("修改公告失败: %s", err.Error()), false)
		return
	}

	// 记录操作日志
	c.recordOperLog(ctx, "通知公告", "修改", fmt.Sprintf("修改公告成功，ID: %d", notice.NoticeID), true)

	response.SuccessWithMessage(ctx, "修改成功")
}

// Remove 删除通知公告 对应Java后端的remove方法
// @Summary 删除通知公告
// @Description 删除通知公告，支持批量删除和权限验证
// @Tags 通知公告管理
// @Accept json
// @Produce json
// @Param noticeIds path string true "公告ID列表，多个ID用逗号分隔"
// @Success 200 {object} response.Response
// @Router /system/notice/{noticeIds} [delete]
func (c *NoticeController) Remove(ctx *gin.Context) {
	noticeIdsStr := ctx.Param("noticeIds")
	fmt.Printf("NoticeController.Remove: 删除通知公告, NoticeIds=%s\n", noticeIdsStr)

	// 权限验证：检查用户是否有删除权限 对应Java后端的@PreAuthorize("@ss.hasPermi('system:notice:remove')")
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}
	currentUser := loginUser.(*model.LoginUser)

	// 检查权限
	hasPermission := false
	for _, permission := range currentUser.Permissions {
		if permission == "system:notice:remove" || permission == "*:*:*" {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		response.ErrorWithMessage(ctx, "没有权限执行此操作")
		return
	}

	// 参数验证
	if noticeIdsStr == "" {
		response.ErrorWithMessage(ctx, "公告ID不能为空")
		return
	}

	// 解析公告ID列表
	idStrings := strings.Split(noticeIdsStr, ",")
	var noticeIds []int64
	for _, idStr := range idStrings {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err != nil || id <= 0 {
			response.ErrorWithMessage(ctx, "公告ID格式错误: "+idStr)
			return
		}
		noticeIds = append(noticeIds, id)
	}

	// 限制批量删除数量
	if len(noticeIds) > 100 {
		response.ErrorWithMessage(ctx, "批量删除数量不能超过100条")
		return
	}

	// 数据权限验证：非管理员只能删除自己创建的公告
	if !currentUser.User.IsAdmin() {
		for _, noticeId := range noticeIds {
			notice, err := c.noticeService.SelectNoticeById(noticeId)
			if err != nil {
				response.ErrorWithMessage(ctx, fmt.Sprintf("查询公告失败(ID:%d): %s", noticeId, err.Error()))
				return
			}
			if notice == nil {
				response.ErrorWithMessage(ctx, fmt.Sprintf("公告不存在(ID:%d)", noticeId))
				return
			}
			if notice.CreateBy != currentUser.User.UserName {
				response.ErrorWithMessage(ctx, fmt.Sprintf("无权限删除公告(ID:%d)", noticeId))
				return
			}
		}
	}

	// 删除公告
	var err error
	if len(noticeIds) == 1 {
		err = c.noticeService.DeleteNoticeById(noticeIds[0])
	} else {
		err = c.noticeService.DeleteNoticeByIds(noticeIds)
	}

	if err != nil {
		response.ErrorWithMessage(ctx, "删除公告失败: "+err.Error())

		// 记录操作日志
		c.recordOperLog(ctx, "通知公告", "删除", fmt.Sprintf("删除公告失败: %s", err.Error()), false)
		return
	}

	// 记录操作日志
	c.recordOperLog(ctx, "通知公告", "删除", fmt.Sprintf("删除公告成功，ID: %v", noticeIds), true)

	response.SuccessWithMessage(ctx, "删除成功")
}

// recordOperLog 记录操作日志
func (c *NoticeController) recordOperLog(ctx *gin.Context, title, businessType, content string, success bool) {
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

// validateNoticeData 验证公告数据 对应Java后端的@Validated注解验证
func (c *NoticeController) validateNoticeData(notice *model.SysNotice, isUpdate bool) error {
	// 公告标题验证
	if notice.NoticeTitle == "" {
		return fmt.Errorf("公告标题不能为空")
	}
	if len(notice.NoticeTitle) > 50 {
		return fmt.Errorf("公告标题不能超过50个字符")
	}

	// 公告类型验证
	if notice.NoticeType == "" {
		return fmt.Errorf("公告类型不能为空")
	}
	if notice.NoticeType != model.NoticeTypeNotification && notice.NoticeType != model.NoticeTypeAnnouncement {
		return fmt.Errorf("公告类型无效，必须为1（通知）或2（公告）")
	}

	// 公告内容验证
	if notice.NoticeContent == "" {
		return fmt.Errorf("公告内容不能为空")
	}
	if len(notice.NoticeContent) > 10000 {
		return fmt.Errorf("公告内容不能超过10000个字符")
	}

	// 公告状态验证
	if notice.Status != "" {
		if notice.Status != model.NoticeStatusNormal && notice.Status != model.NoticeStatusClosed {
			return fmt.Errorf("公告状态无效，必须为0（正常）或1（关闭）")
		}
	}

	// 备注验证
	if len(notice.Remark) > 255 {
		return fmt.Errorf("备注不能超过255个字符")
	}

	// 更新操作时验证ID
	if isUpdate && notice.NoticeID <= 0 {
		return fmt.Errorf("公告ID不能为空")
	}

	return nil
}

// containsXSSContent 检查是否包含XSS攻击内容 对应Java后端的@Xss注解
func (c *NoticeController) containsXSSContent(content string) bool {
	// 检查常见的XSS攻击模式
	xssPatterns := []string{
		"<script", "</script>", "javascript:", "vbscript:",
		"onload=", "onerror=", "onclick=", "onmouseover=", "onmouseout=",
		"onfocus=", "onblur=", "onchange=", "onsubmit=", "onreset=",
		"<iframe", "</iframe>", "<object", "</object>", "<embed", "</embed>",
		"<form", "</form>", "<input", "<textarea", "<select",
		"expression(", "url(javascript:", "url(vbscript:",
		"<meta", "<link", "<style", "</style>", "<base",
		"document.cookie", "document.write", "window.location",
		"eval(", "setTimeout(", "setInterval(",
	}

	lowerContent := strings.ToLower(content)
	for _, pattern := range xssPatterns {
		if strings.Contains(lowerContent, pattern) {
			return true
		}
	}

	return false
}

// Export 导出通知公告数据 对应Java后端的export方法
// @Summary 导出通知公告数据
// @Description 导出通知公告数据到Excel文件
// @Tags 通知公告管理
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param noticeTitle query string false "公告标题"
// @Param noticeType query string false "公告类型"
// @Param status query string false "状态"
// @Security ApiKeyAuth
// @Success 200 {file} file "Excel文件"
// @Router /system/notice/export [post]
func (c *NoticeController) Export(ctx *gin.Context) {
	fmt.Printf("NoticeController.Export: 导出通知公告数据\n")

	// 权限验证 - 对应Java后端的@PreAuthorize("@ss.hasPermi('system:notice:export')")
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}
	currentUser := loginUser.(*model.LoginUser)

	// 检查导出权限
	hasPermission := false
	for _, permission := range currentUser.Permissions {
		if permission == "system:notice:export" || permission == "*:*:*" {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		response.ErrorWithMessage(ctx, "没有权限执行此操作")
		return
	}

	// 构建查询条件
	var params model.NoticeQueryParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		response.ErrorWithMessage(ctx, "参数绑定失败: "+err.Error())
		return
	}

	// 数据权限控制：根据用户角色限制查询范围
	if !currentUser.User.IsAdmin() {
		params.DataScope = "user"
		params.CurrentUserId = currentUser.User.UserID
		params.CurrentUserName = currentUser.User.UserName
	}

	// 查询通知公告列表（不分页，导出所有符合条件的数据）
	params.PageNum = 0
	params.PageSize = 0
	notices, err := c.noticeService.SelectNoticeList(&params)
	if err != nil {
		fmt.Printf("NoticeController.Export: 查询通知公告列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询通知公告列表失败")
		return
	}

	// 转换为导出格式
	var exportData []model.SysNoticeExport
	for _, n := range notices {
		exportData = append(exportData, model.SysNoticeExport{
			NoticeID:      n.NoticeID,
			NoticeTitle:   n.NoticeTitle,
			NoticeType:    getNoticeTypeText(n.NoticeType),
			NoticeContent: n.NoticeContent,
			Status:        getNoticeStatusText(n.Status),
			CreateBy:      n.CreateBy,
			CreateTime:    n.CreateTime.Format("2006-01-02 15:04:05"),
			UpdateBy:      n.UpdateBy,
			UpdateTime:    formatUpdateTime(n.UpdateTime),
			Remark:        n.Remark,
		})
	}

	// 生成Excel文件
	excelUtil := excel.NewExcelUtil()
	fileData, err := excelUtil.ExportExcel(exportData, "通知公告数据", "通知公告列表")
	if err != nil {
		fmt.Printf("NoticeController.Export: 导出Excel失败: %v\n", err)
		response.ErrorWithMessage(ctx, "导出失败: "+err.Error())
		return
	}

	// 生成带日期时间的中文文件名
	filename := fmt.Sprintf("通知公告数据导出_%s.xlsx", time.Now().Format("20060102_150405"))

	// 设置响应头
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(filename)))
	ctx.Header("Content-Length", strconv.Itoa(len(fileData)))

	// 返回文件数据
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileData)

	// 记录操作日志
	operlog.RecordOperLog(ctx, "通知公告", "导出", fmt.Sprintf("导出通知公告数据，共%d条记录", len(exportData)), true)

	fmt.Printf("NoticeController.Export: 导出通知公告数据成功, 数量=%d\n", len(exportData))
}

// getNoticeTypeText 获取公告类型文本
func getNoticeTypeText(noticeType string) string {
	switch noticeType {
	case "1":
		return "通知"
	case "2":
		return "公告"
	default:
		return "未知"
	}
}

// getNoticeStatusText 获取公告状态文本
func getNoticeStatusText(status string) string {
	switch status {
	case "0":
		return "正常"
	case "1":
		return "关闭"
	default:
		return "未知"
	}
}

// formatUpdateTime 格式化更新时间
func formatUpdateTime(updateTime *time.Time) string {
	if updateTime == nil {
		return ""
	}
	return updateTime.Format("2006-01-02 15:04:05")
}
