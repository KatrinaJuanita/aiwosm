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

// DictDataController 字典数据管理控制器 对应Java后端的SysDictDataController
type DictDataController struct {
	dictDataService *systemService.DictDataService
}

// NewDictDataController 创建字典数据管理控制器实例
func NewDictDataController() *DictDataController {
	return &DictDataController{
		dictDataService: systemService.NewDictDataService(),
	}
}

// List 获取字典数据列表 对应Java后端的list方法
// @Summary 获取字典数据列表
// @Description 获取字典数据列表数据
// @Tags 字典数据管理
// @Accept json
// @Produce json
// @Param dictType query string false "字典类型"
// @Param dictLabel query string false "字典标签"
// @Param status query string false "状态"
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页数量"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dict/data/list [get]
func (c *DictDataController) List(ctx *gin.Context) {
	fmt.Printf("DictDataController.List: 获取字典数据列表\n")

	// 获取分页参数
	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))

	// 构建查询条件
	dictData := &model.SysDictData{}
	if dictType := ctx.Query("dictType"); dictType != "" {
		dictData.DictType = dictType
	}
	if dictLabel := ctx.Query("dictLabel"); dictLabel != "" {
		dictData.DictLabel = dictLabel
	}
	if status := ctx.Query("status"); status != "" {
		dictData.Status = status
	}

	// 检查是否需要分页
	if pageNum > 0 && pageSize > 0 {
		// 分页查询
		dictDatas, total, err := c.dictDataService.SelectDictDataListWithPage(dictData, pageNum, pageSize)
		if err != nil {
			fmt.Printf("DictDataController.List: 分页查询字典数据列表失败: %v\n", err)
			response.ErrorWithMessage(ctx, "查询字典数据列表失败")
			return
		}

		// 使用Java后端兼容的TableDataInfo格式
		fmt.Printf("DictDataController.List: 分页查询字典数据列表成功, 数量=%d, 总数=%d\n", len(dictDatas), total)
		tableData := response.GetDataTable(dictDatas, total)
		response.SendTableDataInfo(ctx, tableData)
	} else {
		// 不分页查询（用于下拉框等场景）
		dictDatas, err := c.dictDataService.SelectDictDataList(dictData)
		if err != nil {
			fmt.Printf("DictDataController.List: 查询字典数据列表失败: %v\n", err)
			response.ErrorWithMessage(ctx, "查询字典数据列表失败")
			return
		}

		// 使用Java后端兼容的TableDataInfo格式
		fmt.Printf("DictDataController.List: 查询字典数据列表成功, 数量=%d\n", len(dictDatas))
		tableData := response.GetDataTable(dictDatas, int64(len(dictDatas)))
		response.SendTableDataInfo(ctx, tableData)
	}
}

// GetInfo 根据字典数据编号获取详细信息 对应Java后端的getInfo方法
// @Summary 获取字典数据详细信息
// @Description 根据字典数据ID获取字典数据详细信息
// @Tags 字典数据管理
// @Accept json
// @Produce json
// @Param dictCode path int true "字典数据ID"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dict/data/{dictCode} [get]
func (c *DictDataController) GetInfo(ctx *gin.Context) {
	fmt.Printf("DictDataController.GetInfo: 获取字典数据详细信息\n")

	dictCodeStr := ctx.Param("dictCode")
	dictCode, err := strconv.ParseInt(dictCodeStr, 10, 64)
	if err != nil {
		response.ErrorWithMessage(ctx, "字典数据ID格式错误")
		return
	}

	// 查询字典数据详情
	dictData, err := c.dictDataService.SelectDictDataById(dictCode)
	if err != nil {
		fmt.Printf("DictDataController.GetInfo: 查询字典数据详情失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询字典数据详情失败")
		return
	}

	if dictData == nil {
		response.ErrorWithMessage(ctx, "字典数据不存在")
		return
	}

	fmt.Printf("DictDataController.GetInfo: 查询字典数据详情成功, DictCode=%d\n", dictCode)
	response.SuccessWithData(ctx, dictData)
}

// DictType 根据字典类型查询字典数据信息 对应Java后端的dictType方法
// @Summary 根据字典类型查询字典数据信息
// @Description 根据字典类型查询字典数据信息
// @Tags 字典数据管理
// @Accept json
// @Produce json
// @Param dictType path string true "字典类型"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dict/data/type/{dictType} [get]
func (c *DictDataController) DictType(ctx *gin.Context) {
	fmt.Printf("DictDataController.DictType: 根据字典类型查询字典数据\n")

	dictType := ctx.Param("dictType")
	if dictType == "" {
		response.ErrorWithMessage(ctx, "字典类型不能为空")
		return
	}

	// 查询字典数据
	dictDatas, err := c.dictDataService.SelectDictDataByType(dictType)
	if err != nil {
		fmt.Printf("DictDataController.DictType: 查询字典数据失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询字典数据失败")
		return
	}

	fmt.Printf("DictDataController.DictType: 查询字典数据成功, DictType=%s, 数量=%d\n", dictType, len(dictDatas))
	response.SuccessWithData(ctx, dictDatas)
}

// Add 新增字典数据 对应Java后端的add方法
// @Summary 新增字典数据
// @Description 新增字典数据信息
// @Tags 字典数据管理
// @Accept json
// @Produce json
// @Param dictData body model.SysDictData true "字典数据信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dict/data [post]
func (c *DictDataController) Add(ctx *gin.Context) {
	fmt.Printf("DictDataController.Add: 新增字典数据\n")

	var dictData model.SysDictData
	if err := ctx.ShouldBindJSON(&dictData); err != nil {
		fmt.Printf("DictDataController.Add: 参数绑定失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数格式错误")
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)
	dictData.CreateBy = currentUser.User.UserName

	// 新增字典数据
	if err := c.dictDataService.InsertDictData(&dictData); err != nil {
		fmt.Printf("DictDataController.Add: 新增字典数据失败: %v\n", err)
		response.ErrorWithMessage(ctx, "新增字典数据失败")
		return
	}

	fmt.Printf("DictDataController.Add: 新增字典数据成功, DictType=%s\n", dictData.DictType)
	response.SuccessWithMessage(ctx, "新增成功")
}

// Edit 修改字典数据 对应Java后端的edit方法
// @Summary 修改字典数据
// @Description 修改字典数据信息
// @Tags 字典数据管理
// @Accept json
// @Produce json
// @Param dictData body model.SysDictData true "字典数据信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dict/data [put]
func (c *DictDataController) Edit(ctx *gin.Context) {
	fmt.Printf("DictDataController.Edit: 修改字典数据\n")

	var dictData model.SysDictData
	if err := ctx.ShouldBindJSON(&dictData); err != nil {
		fmt.Printf("DictDataController.Edit: 参数绑定失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数格式错误")
		return
	}

	// 获取当前登录用户
	loginUser, _ := ctx.Get("loginUser")
	currentUser := loginUser.(*model.LoginUser)
	dictData.UpdateBy = currentUser.User.UserName

	// 修改字典数据
	if err := c.dictDataService.UpdateDictData(&dictData); err != nil {
		fmt.Printf("DictDataController.Edit: 修改字典数据失败: %v\n", err)
		response.ErrorWithMessage(ctx, "修改字典数据失败")
		return
	}

	fmt.Printf("DictDataController.Edit: 修改字典数据成功, DictCode=%d\n", dictData.DictCode)
	response.SuccessWithMessage(ctx, "修改成功")
}

// Remove 删除字典数据 对应Java后端的remove方法
// @Summary 删除字典数据
// @Description 删除字典数据信息
// @Tags 字典数据管理
// @Accept json
// @Produce json
// @Param dictCodes path string true "字典数据ID，多个用逗号分隔"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dict/data/{dictCodes} [delete]
func (c *DictDataController) Remove(ctx *gin.Context) {
	fmt.Printf("DictDataController.Remove: 删除字典数据\n")

	dictCodesStr := ctx.Param("dictCodes")
	if dictCodesStr == "" {
		response.ErrorWithMessage(ctx, "字典数据ID不能为空")
		return
	}

	// 解析字典数据ID列表
	dictCodeStrs := strings.Split(dictCodesStr, ",")
	var dictCodes []int64
	for _, dictCodeStr := range dictCodeStrs {
		dictCode, err := strconv.ParseInt(strings.TrimSpace(dictCodeStr), 10, 64)
		if err != nil {
			response.ErrorWithMessage(ctx, "字典数据ID格式错误")
			return
		}
		dictCodes = append(dictCodes, dictCode)
	}

	// 删除字典数据
	if err := c.dictDataService.DeleteDictDataByIds(dictCodes); err != nil {
		fmt.Printf("DictDataController.Remove: 删除字典数据失败: %v\n", err)
		response.ErrorWithMessage(ctx, "删除字典数据失败")
		return
	}

	// 记录操作日志
	operlog.RecordOperLog(ctx, "字典数据", "删除", fmt.Sprintf("删除字典数据: %s", dictCodesStr), true)

	fmt.Printf("DictDataController.Remove: 删除字典数据成功, DictCodes=%s\n", dictCodesStr)
	response.SuccessWithMessage(ctx, "删除成功")
}

// Export 导出字典数据 对应Java后端的export方法
// @Summary 导出字典数据
// @Description 导出字典数据
// @Tags 字典数据管理
// @Accept json
// @Produce json
// @Param dictType query string false "字典类型"
// @Param dictLabel query string false "字典标签"
// @Param status query string false "状态"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /system/dict/data/export [post]
func (c *DictDataController) Export(ctx *gin.Context) {
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

	fmt.Printf("DictDataController.Export: 导出字典数据，用户: %s\n", currentUser.User.UserName)

	// 解析POST请求表单参数 对应Java后端的SysDictData dictData参数绑定
	formParams, err := export.ParseFormParams(ctx)
	if err != nil {
		fmt.Printf("DictDataController.Export: 解析表单参数失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数解析失败: "+err.Error())
		return
	}

	// 解析字典数据查询参数
	queryParams := export.ParseDictDataQueryParams(formParams)

	// 构建查询条件 对应Java后端的SysDictData对象
	dictData := &model.SysDictData{
		DictLabel: queryParams.DictLabel,
		DictValue: queryParams.DictValue,
		DictType:  queryParams.DictType,
		Status:    queryParams.Status,
	}

	fmt.Printf("DictDataController.Export: 查询条件 - DictLabel=%s, DictValue=%s, DictType=%s, Status=%s, BeginTime=%v, EndTime=%v\n",
		dictData.DictLabel, dictData.DictValue, dictData.DictType, dictData.Status, queryParams.BeginTime, queryParams.EndTime)

	// 查询所有符合条件的字典数据（不分页）
	dictDatas, err := c.dictDataService.SelectDictDataList(dictData)
	if err != nil {
		response.ErrorWithMessage(ctx, "查询失败")
		return
	}

	// 使用Excel工具类导出 对应Java后端的ExcelUtil<SysDictData> util = new ExcelUtil<SysDictData>(SysDictData.class)
	excelUtil := excel.NewExcelUtil()
	fileData, err := excelUtil.ExportExcel(dictDatas, "字典数据", "字典数据列表")
	if err != nil {
		response.ErrorWithMessage(ctx, "导出失败: "+err.Error())
		return
	}

	// 生成带日期时间的中文文件名
	now := time.Now()
	filename := fmt.Sprintf("字典数据导出_%s.xlsx", now.Format("20060102_150405"))

	// 设置响应头 对应Java后端的util.exportExcel(response, list, "字典数据")
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(filename)))
	ctx.Header("Content-Length", strconv.Itoa(len(fileData)))

	// 返回文件数据
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileData)

	fmt.Printf("DictDataController.Export: 导出字典数据成功, 数量=%d, 文件大小=%d bytes\n", len(dictDatas), len(fileData))

	// 记录操作日志 - 对应Java后端的@Log(title = "字典数据", businessType = BusinessType.EXPORT)
	operlog.RecordOperLog(ctx, "字典数据", "导出", fmt.Sprintf("导出字典数据成功，数量: %d", len(dictDatas)), true)
}
