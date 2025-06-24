package system

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"wosm/internal/repository/model"
	systemService "wosm/internal/service/system"
	"wosm/pkg/excel"
	"wosm/pkg/export"
	"wosm/pkg/operlog"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// DictTypeController 字典类型管理控制器 对应Java后端的SysDictTypeController
type DictTypeController struct {
	dictTypeService *systemService.DictTypeService
	dictDataService *systemService.DictDataService
}

// NewDictTypeController 创建字典类型管理控制器实例
func NewDictTypeController() *DictTypeController {
	return &DictTypeController{
		dictTypeService: systemService.NewDictTypeService(),
		dictDataService: systemService.NewDictDataService(),
	}
}

// List 获取字典类型列表 对应Java后端的list方法
// @Summary 获取字典类型列表
// @Description 获取字典类型列表数据
// @Tags 字典类型管理
// @Accept json
// @Produce json
// @Param dictName query string false "字典名称"
// @Param dictType query string false "字典类型"
// @Param status query string false "状态"
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页数量"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dict/type/list [get]
func (c *DictTypeController) List(ctx *gin.Context) {
	fmt.Printf("DictTypeController.List: 获取字典类型列表\n")

	// 获取分页参数
	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))

	// 构建查询条件
	dictType := &model.SysDictType{}
	if dictName := ctx.Query("dictName"); dictName != "" {
		dictType.DictName = dictName
	}
	if dictTypeParam := ctx.Query("dictType"); dictTypeParam != "" {
		dictType.DictType = dictTypeParam
	}
	if status := ctx.Query("status"); status != "" {
		dictType.Status = status
	}

	// 检查是否需要分页
	if pageNum > 0 && pageSize > 0 {
		// 分页查询
		dictTypes, total, err := c.dictTypeService.SelectDictTypeListWithPage(dictType, pageNum, pageSize)
		if err != nil {
			fmt.Printf("DictTypeController.List: 分页查询字典类型列表失败: %v\n", err)
			response.ErrorWithMessage(ctx, "查询字典类型列表失败")
			return
		}

		// 使用Java后端兼容的TableDataInfo格式
		fmt.Printf("DictTypeController.List: 分页查询字典类型列表成功, 数量=%d, 总数=%d\n", len(dictTypes), total)
		tableData := response.GetDataTable(dictTypes, total)
		response.SendTableDataInfo(ctx, tableData)
	} else {
		// 不分页查询（用于下拉框等场景）
		dictTypes, err := c.dictTypeService.SelectDictTypeList(dictType)
		if err != nil {
			fmt.Printf("DictTypeController.List: 查询字典类型列表失败: %v\n", err)
			response.ErrorWithMessage(ctx, "查询字典类型列表失败")
			return
		}

		// 使用Java后端兼容的TableDataInfo格式
		fmt.Printf("DictTypeController.List: 查询字典类型列表成功, 数量=%d\n", len(dictTypes))
		tableData := response.GetDataTable(dictTypes, int64(len(dictTypes)))
		response.SendTableDataInfo(ctx, tableData)
	}
}

// GetInfo 根据字典类型编号获取详细信息 对应Java后端的getInfo方法
// @Summary 获取字典类型详细信息
// @Description 根据字典类型ID获取字典类型详细信息
// @Tags 字典类型管理
// @Accept json
// @Produce json
// @Param dictId path int true "字典类型ID"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dict/type/{dictId} [get]
func (c *DictTypeController) GetInfo(ctx *gin.Context) {
	fmt.Printf("DictTypeController.GetInfo: 获取字典类型详细信息\n")

	dictIdStr := ctx.Param("dictId")
	dictId, err := strconv.ParseInt(dictIdStr, 10, 64)
	if err != nil {
		response.ErrorWithMessage(ctx, "字典类型ID格式错误")
		return
	}

	// 查询字典类型详情
	dictType, err := c.dictTypeService.SelectDictTypeById(dictId)
	if err != nil {
		fmt.Printf("DictTypeController.GetInfo: 查询字典类型详情失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询字典类型详情失败")
		return
	}

	if dictType == nil {
		response.ErrorWithMessage(ctx, "字典类型不存在")
		return
	}

	fmt.Printf("DictTypeController.GetInfo: 查询字典类型详情成功, DictID=%d\n", dictId)
	response.SuccessWithData(ctx, dictType)
}

// Add 新增字典类型 对应Java后端的add方法
// @Summary 新增字典类型
// @Description 新增字典类型信息
// @Tags 字典类型管理
// @Accept json
// @Produce json
// @Param dictType body model.SysDictType true "字典类型信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dict/type [post]
func (c *DictTypeController) Add(ctx *gin.Context) {
	fmt.Printf("DictTypeController.Add: 新增字典类型\n")

	var dictType model.SysDictType
	if err := ctx.ShouldBindJSON(&dictType); err != nil {
		fmt.Printf("DictTypeController.Add: 参数绑定失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数格式错误")
		return
	}

	// 校验字典类型唯一性
	if !c.dictTypeService.CheckDictTypeUnique(&dictType) {
		response.ErrorWithMessage(ctx, fmt.Sprintf("新增字典'%s'失败，字典类型已存在", dictType.DictType))
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)
	dictType.CreateBy = currentUser.User.UserName

	// 新增字典类型
	if err := c.dictTypeService.InsertDictType(&dictType); err != nil {
		fmt.Printf("DictTypeController.Add: 新增字典类型失败: %v\n", err)
		response.ErrorWithMessage(ctx, "新增字典类型失败")
		return
	}

	fmt.Printf("DictTypeController.Add: 新增字典类型成功, DictType=%s\n", dictType.DictType)
	response.SuccessWithMessage(ctx, "新增成功")
}

// Edit 修改字典类型 对应Java后端的edit方法
// @Summary 修改字典类型
// @Description 修改字典类型信息
// @Tags 字典类型管理
// @Accept json
// @Produce json
// @Param dictType body model.SysDictType true "字典类型信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dict/type [put]
func (c *DictTypeController) Edit(ctx *gin.Context) {
	fmt.Printf("DictTypeController.Edit: 修改字典类型\n")

	var dictType model.SysDictType
	if err := ctx.ShouldBindJSON(&dictType); err != nil {
		fmt.Printf("DictTypeController.Edit: 参数绑定失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数格式错误")
		return
	}

	// 校验字典类型唯一性
	if !c.dictTypeService.CheckDictTypeUnique(&dictType) {
		response.ErrorWithMessage(ctx, fmt.Sprintf("修改字典'%s'失败，字典类型已存在", dictType.DictType))
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)
	dictType.UpdateBy = currentUser.User.UserName

	// 修改字典类型
	if err := c.dictTypeService.UpdateDictType(&dictType); err != nil {
		fmt.Printf("DictTypeController.Edit: 修改字典类型失败: %v\n", err)
		response.ErrorWithMessage(ctx, "修改字典类型失败")
		return
	}

	fmt.Printf("DictTypeController.Edit: 修改字典类型成功, DictID=%d\n", dictType.DictID)
	response.SuccessWithMessage(ctx, "修改成功")
}

// Remove 删除字典类型 对应Java后端的remove方法
// @Summary 删除字典类型
// @Description 删除字典类型信息
// @Tags 字典类型管理
// @Accept json
// @Produce json
// @Param dictIds path string true "字典类型ID列表，多个用逗号分隔"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dict/type/{dictIds} [delete]
func (c *DictTypeController) Remove(ctx *gin.Context) {
	fmt.Printf("DictTypeController.Remove: 删除字典类型\n")

	dictIdsStr := ctx.Param("dictIds")
	if dictIdsStr == "" {
		response.ErrorWithMessage(ctx, "字典类型ID不能为空")
		return
	}

	// 解析字典类型ID列表
	dictIdStrs := strings.Split(dictIdsStr, ",")
	var dictIds []int64
	for _, dictIdStr := range dictIdStrs {
		dictId, err := strconv.ParseInt(strings.TrimSpace(dictIdStr), 10, 64)
		if err != nil {
			response.ErrorWithMessage(ctx, "字典类型ID格式错误")
			return
		}
		dictIds = append(dictIds, dictId)
	}

	// 检查字典类型是否被使用
	for _, dictId := range dictIds {
		dictType, err := c.dictTypeService.SelectDictTypeById(dictId)
		if err != nil {
			response.ErrorWithMessage(ctx, "查询字典类型失败")
			return
		}
		if dictType != nil {
			count, err := c.dictDataService.CountDictDataByType(dictType.DictType)
			if err == nil && count > 0 {
				response.ErrorWithMessage(ctx, fmt.Sprintf("%s已分配,不能删除", dictType.DictName))
				return
			}
		}
	}

	// 删除字典类型
	if err := c.dictTypeService.DeleteDictTypeByIds(dictIds); err != nil {
		fmt.Printf("DictTypeController.Remove: 删除字典类型失败: %v\n", err)
		response.ErrorWithMessage(ctx, "删除字典类型失败")
		return
	}

	fmt.Printf("DictTypeController.Remove: 删除字典类型成功, DictIDs=%v\n", dictIds)
	response.SuccessWithMessage(ctx, "删除成功")
}

// RefreshCache 刷新字典缓存 对应Java后端的refreshCache方法
// @Summary 刷新字典缓存
// @Description 刷新字典缓存数据
// @Tags 字典类型管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dict/type/refreshCache [delete]
func (c *DictTypeController) RefreshCache(ctx *gin.Context) {
	fmt.Printf("DictTypeController.RefreshCache: 刷新字典缓存\n")

	// 重置字典缓存
	if err := c.dictTypeService.ResetDictCache(); err != nil {
		fmt.Printf("DictTypeController.RefreshCache: 刷新字典缓存失败: %v\n", err)
		response.ErrorWithMessage(ctx, "刷新字典缓存失败")
		return
	}

	fmt.Printf("DictTypeController.RefreshCache: 刷新字典缓存成功\n")
	response.SuccessWithMessage(ctx, "刷新成功")
}

// OptionSelect 获取字典选择框列表 对应Java后端的optionselect方法
// @Summary 获取字典选择框列表
// @Description 获取字典类型选择框列表数据
// @Tags 字典类型管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dict/type/optionselect [get]
func (c *DictTypeController) OptionSelect(ctx *gin.Context) {
	fmt.Printf("DictTypeController.OptionSelect: 获取字典选择框列表\n")

	// 查询所有字典类型
	dictTypes, err := c.dictTypeService.SelectDictTypeAll()
	if err != nil {
		fmt.Printf("DictTypeController.OptionSelect: 查询字典类型列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询字典类型列表失败")
		return
	}

	fmt.Printf("DictTypeController.OptionSelect: 查询字典选择框列表成功, 数量=%d\n", len(dictTypes))
	response.SuccessWithData(ctx, dictTypes)
}

// Export 导出字典类型数据 对应Java后端的export方法
// @Summary 导出字典类型数据
// @Description 导出字典类型数据
// @Tags 字典类型管理
// @Accept json
// @Produce json
// @Param dictName query string false "字典名称"
// @Param dictType query string false "字典类型"
// @Param status query string false "状态"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dict/type/export [post]
func (c *DictTypeController) Export(ctx *gin.Context) {
	// 权限验证 - 对应Java后端的@PreAuthorize("@ss.hasPermi('system:dict:export')")
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}

	currentUser := loginUser.(*model.LoginUser)
	hasPermission := currentUser.User.IsAdmin()
	if !hasPermission {
		for _, perm := range currentUser.Permissions {
			if perm == "system:dict:export" || perm == "system:dict:*" {
				hasPermission = true
				break
			}
		}
	}

	if !hasPermission {
		response.ErrorWithMessage(ctx, "权限不足")
		return
	}

	fmt.Printf("DictTypeController.Export: 导出字典类型数据，用户: %s\n", currentUser.User.UserName)

	// 解析POST请求表单参数 对应Java后端的SysDictType dictType参数绑定
	formParams, err := export.ParseFormParams(ctx)
	if err != nil {
		fmt.Printf("DictTypeController.Export: 解析表单参数失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数解析失败: "+err.Error())
		return
	}

	// 解析字典类型查询参数
	queryParams := export.ParseDictTypeQueryParams(formParams)

	// 构建查询条件 对应Java后端的SysDictType对象
	dictType := &model.SysDictType{
		DictName: queryParams.DictName,
		DictType: queryParams.DictType,
		Status:   queryParams.Status,
	}

	fmt.Printf("DictTypeController.Export: 查询条件 - DictName=%s, DictType=%s, Status=%s, BeginTime=%v, EndTime=%v\n",
		dictType.DictName, dictType.DictType, dictType.Status, queryParams.BeginTime, queryParams.EndTime)

	// 查询所有符合条件的字典类型（不分页）
	dictTypes, err := c.dictTypeService.SelectDictTypeList(dictType)
	if err != nil {
		response.ErrorWithMessage(ctx, "查询失败")
		return
	}

	// 使用Excel工具类导出 对应Java后端的ExcelUtil<SysDictType> util = new ExcelUtil<SysDictType>(SysDictType.class)
	excelUtil := excel.NewExcelUtil()
	fileData, err := excelUtil.ExportExcel(dictTypes, "字典类型数据", "字典类型列表")
	if err != nil {
		response.ErrorWithMessage(ctx, "导出失败: "+err.Error())
		return
	}

	// 生成带日期时间的中文文件名
	now := time.Now()
	filename := fmt.Sprintf("字典类型数据导出_%s.xlsx", now.Format("20060102_150405"))

	// 设置响应头 对应Java后端的util.exportExcel(response, list, "字典类型数据")
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(filename)))
	ctx.Header("Content-Length", strconv.Itoa(len(fileData)))

	// 返回文件数据
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileData)

	fmt.Printf("DictTypeController.Export: 导出字典类型数据成功, 数量=%d, 文件大小=%d bytes\n", len(dictTypes), len(fileData))

	// 记录操作日志 - 对应Java后端的@Log(title = "字典类型", businessType = BusinessType.EXPORT)
	operlog.RecordOperLog(ctx, "字典类型", "导出", fmt.Sprintf("导出字典类型成功，数量: %d", len(dictTypes)), true)
}
