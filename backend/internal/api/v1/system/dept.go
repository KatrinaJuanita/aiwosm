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

// DeptController 部门管理控制器 对应Java后端的SysDeptController
type DeptController struct {
	deptService *system.DeptService
}

// NewDeptController 创建部门管理控制器实例
func NewDeptController() *DeptController {
	return &DeptController{
		deptService: system.NewDeptService(),
	}
}

// List 获取部门列表 对应Java后端的list方法
// @Summary 获取部门列表
// @Description 获取部门列表数据
// @Tags 部门管理
// @Accept json
// @Produce json
// @Param deptName query string false "部门名称"
// @Param status query string false "部门状态"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dept/list [get]
func (c *DeptController) List(ctx *gin.Context) {
	fmt.Printf("DeptController.List: 获取部门列表\n")

	// 构建查询条件
	dept := &model.SysDept{}
	if deptName := ctx.Query("deptName"); deptName != "" {
		dept.DeptName = deptName
	}
	if status := ctx.Query("status"); status != "" {
		dept.Status = status
	}

	// 查询部门列表
	depts, err := c.deptService.SelectDeptList(dept)
	if err != nil {
		fmt.Printf("DeptController.List: 查询部门列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询部门列表失败")
		return
	}

	fmt.Printf("DeptController.List: 查询部门列表成功, 数量=%d\n", len(depts))
	response.SuccessWithData(ctx, depts)
}

// ExcludeChild 查询部门列表（排除节点） 对应Java后端的excludeChild方法
// @Summary 查询部门列表（排除节点）
// @Description 查询部门列表，排除指定节点及其子节点
// @Tags 部门管理
// @Accept json
// @Produce json
// @Param deptId path int true "部门ID"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dept/list/exclude/{deptId} [get]
func (c *DeptController) ExcludeChild(ctx *gin.Context) {
	fmt.Printf("DeptController.ExcludeChild: 查询部门列表（排除节点）\n")

	deptIdStr := ctx.Param("deptId")
	deptId, err := strconv.ParseInt(deptIdStr, 10, 64)
	if err != nil {
		response.ErrorWithMessage(ctx, "部门ID格式错误")
		return
	}

	// 查询所有部门
	depts, err := c.deptService.SelectDeptList(&model.SysDept{})
	if err != nil {
		fmt.Printf("DeptController.ExcludeChild: 查询部门列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询部门列表失败")
		return
	}

	// 过滤掉指定部门及其子部门
	var filteredDepts []model.SysDept
	for _, dept := range depts {
		// 排除自己
		if dept.DeptID == deptId {
			continue
		}
		// 排除子部门（通过ancestors字段判断）
		if dept.Ancestors != "" && contains(dept.Ancestors, strconv.FormatInt(deptId, 10)) {
			continue
		}
		filteredDepts = append(filteredDepts, dept)
	}

	fmt.Printf("DeptController.ExcludeChild: 查询部门列表成功, 排除后数量=%d\n", len(filteredDepts))
	response.SuccessWithData(ctx, filteredDepts)
}

// GetInfo 根据部门编号获取详细信息 对应Java后端的getInfo方法
// @Summary 获取部门详细信息
// @Description 根据部门ID获取部门详细信息
// @Tags 部门管理
// @Accept json
// @Produce json
// @Param deptId path int true "部门ID"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dept/{deptId} [get]
func (c *DeptController) GetInfo(ctx *gin.Context) {
	fmt.Printf("DeptController.GetInfo: 获取部门详细信息\n")

	deptIdStr := ctx.Param("deptId")
	deptId, err := strconv.ParseInt(deptIdStr, 10, 64)
	if err != nil {
		response.ErrorWithMessage(ctx, "部门ID格式错误")
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)

	// 校验数据权限
	if err := c.deptService.CheckDeptDataScope(currentUser.User, deptId); err != nil {
		fmt.Printf("DeptController.GetInfo: 数据权限校验失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 查询部门详情
	dept, err := c.deptService.SelectDeptById(deptId)
	if err != nil {
		fmt.Printf("DeptController.GetInfo: 查询部门详情失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询部门详情失败")
		return
	}

	if dept == nil {
		response.ErrorWithMessage(ctx, "部门不存在")
		return
	}

	fmt.Printf("DeptController.GetInfo: 查询部门详情成功, DeptID=%d\n", deptId)
	response.SuccessWithData(ctx, dept)
}

// Add 新增部门 对应Java后端的add方法
// @Summary 新增部门
// @Description 新增部门信息
// @Tags 部门管理
// @Accept json
// @Produce json
// @Param dept body model.SysDept true "部门信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dept [post]
func (c *DeptController) Add(ctx *gin.Context) {
	fmt.Printf("DeptController.Add: 新增部门\n")

	var dept model.SysDept
	if err := ctx.ShouldBindJSON(&dept); err != nil {
		fmt.Printf("DeptController.Add: 参数绑定失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数格式错误")
		return
	}

	// 校验部门名称唯一性
	if !c.deptService.CheckDeptNameUnique(&dept) {
		response.ErrorWithMessage(ctx, fmt.Sprintf("新增部门'%s'失败，部门名称已存在", dept.DeptName))
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)
	dept.CreateBy = currentUser.User.UserName

	// 新增部门
	if err := c.deptService.InsertDept(&dept); err != nil {
		fmt.Printf("DeptController.Add: 新增部门失败: %v\n", err)
		// 记录操作日志 - 失败
		operlog.RecordOperLog(ctx, "部门管理", "新增", fmt.Sprintf("新增部门失败: %s", err.Error()), false)
		response.ErrorWithMessage(ctx, "新增部门失败")
		return
	}

	// 记录操作日志 - 成功 对应Java后端的@Log(title = "部门管理", businessType = BusinessType.INSERT)
	operlog.RecordOperLog(ctx, "部门管理", "新增", fmt.Sprintf("新增部门成功，部门名称: %s", dept.DeptName), true)
	fmt.Printf("DeptController.Add: 新增部门成功, DeptName=%s\n", dept.DeptName)
	response.SuccessWithMessage(ctx, "新增成功")
}

// Edit 修改部门 对应Java后端的edit方法
// @Summary 修改部门
// @Description 修改部门信息
// @Tags 部门管理
// @Accept json
// @Produce json
// @Param dept body model.SysDept true "部门信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dept [put]
func (c *DeptController) Edit(ctx *gin.Context) {
	fmt.Printf("DeptController.Edit: 修改部门\n")

	var dept model.SysDept
	if err := ctx.ShouldBindJSON(&dept); err != nil {
		fmt.Printf("DeptController.Edit: 参数绑定失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数格式错误")
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)

	// 校验数据权限
	if err := c.deptService.CheckDeptDataScope(currentUser.User, dept.DeptID); err != nil {
		fmt.Printf("DeptController.Edit: 数据权限校验失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 校验部门名称唯一性
	if !c.deptService.CheckDeptNameUnique(&dept) {
		response.ErrorWithMessage(ctx, fmt.Sprintf("修改部门'%s'失败，部门名称已存在", dept.DeptName))
		return
	}

	// 校验上级部门不能是自己
	if dept.ParentID == dept.DeptID {
		response.ErrorWithMessage(ctx, fmt.Sprintf("修改部门'%s'失败，上级部门不能是自己", dept.DeptName))
		return
	}

	// 如果停用部门，检查是否包含未停用的子部门
	if dept.Status == "1" {
		count, err := c.deptService.SelectNormalChildrenDeptById(dept.DeptID)
		if err == nil && count > 0 {
			response.ErrorWithMessage(ctx, "该部门包含未停用的子部门！")
			return
		}
	}

	// 设置更新者
	dept.UpdateBy = currentUser.User.UserName

	// 修改部门
	if err := c.deptService.UpdateDept(&dept); err != nil {
		fmt.Printf("DeptController.Edit: 修改部门失败: %v\n", err)
		// 记录操作日志 - 失败
		operlog.RecordOperLog(ctx, "部门管理", "修改", fmt.Sprintf("修改部门失败: %s", err.Error()), false)
		response.ErrorWithMessage(ctx, "修改部门失败")
		return
	}

	// 记录操作日志 - 成功 对应Java后端的@Log(title = "部门管理", businessType = BusinessType.UPDATE)
	operlog.RecordOperLog(ctx, "部门管理", "修改", fmt.Sprintf("修改部门成功，部门名称: %s", dept.DeptName), true)
	fmt.Printf("DeptController.Edit: 修改部门成功, DeptID=%d\n", dept.DeptID)
	response.SuccessWithMessage(ctx, "修改成功")
}

// Remove 删除部门 对应Java后端的remove方法
// @Summary 删除部门
// @Description 删除部门信息
// @Tags 部门管理
// @Accept json
// @Produce json
// @Param deptId path int true "部门ID"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dept/{deptId} [delete]
func (c *DeptController) Remove(ctx *gin.Context) {
	fmt.Printf("DeptController.Remove: 删除部门\n")

	deptIdStr := ctx.Param("deptId")
	deptId, err := strconv.ParseInt(deptIdStr, 10, 64)
	if err != nil {
		fmt.Printf("DeptController.Remove: 部门ID格式错误: %v\n", err)
		response.ErrorWithMessage(ctx, "部门ID格式错误")
		return
	}

	fmt.Printf("DeptController.Remove: 开始删除部门, DeptID=%d\n", deptId)

	// 检查部门是否存在
	dept, err := c.deptService.SelectDeptById(deptId)
	if err != nil {
		fmt.Printf("DeptController.Remove: 查询部门失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询部门失败")
		return
	}
	if dept == nil {
		fmt.Printf("DeptController.Remove: 部门不存在, DeptID=%d\n", deptId)
		response.ErrorWithMessage(ctx, "部门不存在")
		return
	}

	// 检查是否存在下级部门
	if c.deptService.HasChildByDeptId(deptId) {
		fmt.Printf("DeptController.Remove: 存在下级部门, DeptID=%d\n", deptId)
		response.ErrorWithMessage(ctx, "存在下级部门,不允许删除")
		return
	}

	// 检查部门是否存在用户
	if c.deptService.CheckDeptExistUser(deptId) {
		fmt.Printf("DeptController.Remove: 部门存在用户, DeptID=%d\n", deptId)
		response.ErrorWithMessage(ctx, "部门存在用户,不允许删除")
		return
	}

	// 获取当前登录用户
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		fmt.Printf("DeptController.Remove: 用户未登录\n")
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}
	currentUser := loginUser.(*model.LoginUser)

	// 校验数据权限
	if err := c.deptService.CheckDeptDataScope(currentUser.User, deptId); err != nil {
		fmt.Printf("DeptController.Remove: 数据权限校验失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 删除部门
	if err := c.deptService.DeleteDeptById(deptId); err != nil {
		fmt.Printf("DeptController.Remove: 删除部门失败: %v\n", err)
		// 记录操作日志 - 失败
		operlog.RecordOperLog(ctx, "部门管理", "删除", fmt.Sprintf("删除部门失败: %s", err.Error()), false)
		response.ErrorWithMessage(ctx, "删除部门失败")
		return
	}

	// 记录操作日志 - 成功 对应Java后端的@Log(title = "部门管理", businessType = BusinessType.DELETE)
	operlog.RecordOperLog(ctx, "部门管理", "删除", fmt.Sprintf("删除部门成功，部门ID: %d", deptId), true)
	fmt.Printf("DeptController.Remove: 删除部门成功, DeptID=%d\n", deptId)
	response.SuccessWithMessage(ctx, "删除成功")
}

// TreeSelect 获取部门下拉树列表 对应Java后端的treeselect方法
// @Summary 获取部门下拉树列表
// @Description 获取部门树形结构数据用于下拉选择
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dept/treeselect [get]
func (c *DeptController) TreeSelect(ctx *gin.Context) {
	fmt.Printf("DeptController.TreeSelect: 获取部门下拉树列表\n")

	// 构建查询条件
	dept := &model.SysDept{}
	if deptName := ctx.Query("deptName"); deptName != "" {
		dept.DeptName = deptName
	}
	if status := ctx.Query("status"); status != "" {
		dept.Status = status
	}

	// 查询部门树选择结构
	treeSelect, err := c.deptService.SelectDeptTreeList(dept)
	if err != nil {
		fmt.Printf("DeptController.TreeSelect: 查询部门树失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询部门树失败")
		return
	}

	fmt.Printf("DeptController.TreeSelect: 获取部门下拉树列表成功, 数量=%d\n", len(treeSelect))
	response.SuccessWithData(ctx, treeSelect)
}

// RoleDeptTreeSelect 加载对应角色部门列表树 对应Java后端的roleDeptTreeselect方法
// @Summary 加载对应角色部门列表树
// @Description 根据角色ID获取部门树和已选中的部门
// @Tags 部门管理
// @Accept json
// @Produce json
// @Param roleId path int true "角色ID"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dept/roleDeptTreeselect/{roleId} [get]
func (c *DeptController) RoleDeptTreeSelect(ctx *gin.Context) {
	fmt.Printf("DeptController.RoleDeptTreeSelect: 加载对应角色部门列表树\n")

	roleIdStr := ctx.Param("roleId")
	roleId, err := strconv.ParseInt(roleIdStr, 10, 64)
	if err != nil {
		response.ErrorWithMessage(ctx, "角色ID格式错误")
		return
	}

	// 查询所有部门树
	treeSelect, err := c.deptService.SelectDeptTreeList(&model.SysDept{})
	if err != nil {
		fmt.Printf("DeptController.RoleDeptTreeSelect: 查询部门树失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询部门树失败")
		return
	}

	// 获取角色已选中的部门ID列表
	checkedKeys, err := c.deptService.SelectDeptListByRoleId(roleId)
	if err != nil {
		fmt.Printf("DeptController.RoleDeptTreeSelect: 查询角色部门失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询角色部门失败")
		return
	}

	// 使用AjaxResult格式，与Java后端保持一致
	ajax := response.AjaxSuccess()
	ajax.Put("checkedKeys", checkedKeys)
	ajax.Put("depts", treeSelect)

	fmt.Printf("DeptController.RoleDeptTreeSelect: 加载对应角色部门列表树成功, RoleID=%d\n", roleId)
	response.SendAjaxResult(ctx, ajax)
}

// Export 导出部门数据 对应Java后端的export方法
// @Summary 导出部门数据
// @Description 导出部门数据到Excel文件
// @Tags 部门管理
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param deptName query string false "部门名称"
// @Param status query string false "部门状态"
// @Security ApiKeyAuth
// @Success 200 {file} file "Excel文件"
// @Router /system/dept/export [post]
func (c *DeptController) Export(ctx *gin.Context) {
	fmt.Printf("DeptController.Export: 导出部门数据\n")

	// 权限验证 - 对应Java后端的@PreAuthorize("@ss.hasPermi('system:dept:export')")
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}
	currentUser := loginUser.(*model.LoginUser)

	// 检查导出权限
	hasPermission := false
	for _, permission := range currentUser.Permissions {
		if permission == "system:dept:export" || permission == "*:*:*" {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		response.ErrorWithMessage(ctx, "没有权限执行此操作")
		return
	}

	// 构建查询条件
	dept := &model.SysDept{}
	if deptName := ctx.Query("deptName"); deptName != "" {
		dept.DeptName = deptName
	}
	if status := ctx.Query("status"); status != "" {
		dept.Status = status
	}

	// 获取时间范围参数（部门表没有时间范围字段，这里可以忽略或用于其他逻辑）
	beginTime := ctx.Query("params[beginTime]")
	endTime := ctx.Query("params[endTime]")
	_ = beginTime // 避免未使用变量警告
	_ = endTime   // 避免未使用变量警告

	// 查询部门列表
	depts, err := c.deptService.SelectDeptList(dept)
	if err != nil {
		fmt.Printf("DeptController.Export: 查询部门列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询部门列表失败")
		return
	}

	// 转换为导出格式
	var exportData []model.SysDeptExport
	for _, d := range depts {
		exportData = append(exportData, model.SysDeptExport{
			DeptID:     d.DeptID,
			ParentID:   d.ParentID,
			DeptName:   d.DeptName,
			OrderNum:   d.OrderNum,
			Leader:     d.Leader,
			Phone:      d.Phone,
			Email:      d.Email,
			Status:     getStatusText(d.Status),
			CreateTime: d.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}

	// 生成Excel文件
	excelUtil := excel.NewExcelUtil()
	fileData, err := excelUtil.ExportExcel(exportData, "部门数据", "部门列表")
	if err != nil {
		fmt.Printf("DeptController.Export: 导出Excel失败: %v\n", err)
		response.ErrorWithMessage(ctx, "导出失败: "+err.Error())
		return
	}

	// 生成带日期时间的中文文件名
	filename := fmt.Sprintf("部门数据导出_%s.xlsx", time.Now().Format("20060102_150405"))

	// 设置响应头
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(filename)))
	ctx.Header("Content-Length", strconv.Itoa(len(fileData)))

	// 返回文件数据
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileData)

	// 记录操作日志
	operlog.RecordOperLog(ctx, "部门管理", "导出", fmt.Sprintf("导出部门数据，共%d条记录", len(exportData)), true)

	fmt.Printf("DeptController.Export: 导出部门数据成功, 数量=%d\n", len(exportData))
}

// getStatusText 获取状态文本
func getStatusText(status string) string {
	switch status {
	case "0":
		return "正常"
	case "1":
		return "停用"
	default:
		return "未知"
	}
}

// contains 检查ancestors字符串中是否包含指定的deptId
// ancestors格式: "0,100,101" 或 "0" 等
func contains(ancestors, deptId string) bool {
	if ancestors == "" || deptId == "" {
		return false
	}

	// 将ancestors按逗号分割成数组
	ancestorList := strings.Split(ancestors, ",")

	// 检查deptId是否在祖先列表中
	for _, ancestor := range ancestorList {
		ancestor = strings.TrimSpace(ancestor)
		if ancestor == deptId {
			return true
		}
	}

	return false
}
