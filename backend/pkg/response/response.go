package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 响应状态码常量 对应Java后端的HttpStatus
const (
	SUCCESS = 200 // 成功
	ERROR   = 500 // 系统内部错误
	WARN    = 601 // 系统警告消息
)

// Result 统一响应结构 对应Java后端的AjaxResult
type Result struct {
	Code int         `json:"code"`           // 状态码
	Msg  string      `json:"msg"`            // 返回消息
	Data interface{} `json:"data,omitempty"` // 数据对象
}

// PageResult 分页响应结构 对应Java后端的TableDataInfo
type PageResult struct {
	Code  int         `json:"code"`  // 状态码
	Msg   string      `json:"msg"`   // 返回消息
	Total int64       `json:"total"` // 总记录数
	Rows  interface{} `json:"rows"`  // 列表数据
}

// Success 成功响应
func Success(c *gin.Context) {
	c.JSON(http.StatusOK, Result{
		Code: SUCCESS,
		Msg:  "操作成功",
	})
}

// SuccessWithData 成功响应带数据
func SuccessWithData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Result{
		Code: SUCCESS,
		Msg:  "操作成功",
		Data: data,
	})
}

// SuccessWithFields 成功响应带字段（对应Java的AjaxResult.put方式）
func SuccessWithFields(c *gin.Context, fields map[string]interface{}) {
	result := map[string]interface{}{
		"code": SUCCESS,
		"msg":  "操作成功",
	}

	// 将额外字段添加到根级别
	for key, value := range fields {
		result[key] = value
	}

	c.JSON(http.StatusOK, result)
}

// SuccessWithMessage 成功响应带消息
func SuccessWithMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Result{
		Code: SUCCESS,
		Msg:  message,
	})
}

// SuccessWithDetailed 成功响应带数据和消息
func SuccessWithDetailed(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, Result{
		Code: SUCCESS,
		Msg:  message,
		Data: data,
	})
}

// Error 错误响应
func Error(c *gin.Context) {
	c.JSON(http.StatusOK, Result{
		Code: ERROR,
		Msg:  "操作失败",
	})
}

// ErrorWithMessage 错误响应带消息
func ErrorWithMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Result{
		Code: ERROR,
		Msg:  message,
	})
}

// ErrorWithDetailed 错误响应带详细信息
func ErrorWithDetailed(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Result{
		Code: code,
		Msg:  message,
	})
}

// ErrorWithCode 错误响应带HTTP状态码
func ErrorWithCode(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, Result{
		Code: httpCode,
		Msg:  message,
	})
}

// Warn 警告响应
func Warn(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Result{
		Code: WARN,
		Msg:  message,
	})
}

// Page 分页响应
func Page(c *gin.Context, total int64, rows interface{}) {
	c.JSON(http.StatusOK, PageResult{
		Code:  SUCCESS,
		Msg:   "查询成功",
		Total: total,
		Rows:  rows,
	})
}

// PageWithMessage 分页响应带消息
func PageWithMessage(c *gin.Context, total int64, rows interface{}, message string) {
	c.JSON(http.StatusOK, PageResult{
		Code:  SUCCESS,
		Msg:   message,
		Total: total,
		Rows:  rows,
	})
}

// ========== Java后端兼容方法 ==========

// AjaxResult 对应Java后端的AjaxResult，继承HashMap<String, Object>
type AjaxResult map[string]interface{}

// TableDataInfo 对应Java后端的TableDataInfo，表格分页数据对象
type TableDataInfo struct {
	Total int64       `json:"total"` // 总记录数
	Rows  interface{} `json:"rows"`  // 列表数据
	Code  int         `json:"code"`  // 消息状态码
	Msg   string      `json:"msg"`   // 消息内容
}

// NewTableDataInfo 创建TableDataInfo对象 对应Java后端的TableDataInfo构造函数
func NewTableDataInfo() *TableDataInfo {
	return &TableDataInfo{
		Code: SUCCESS,
		Msg:  "查询成功",
	}
}

// NewTableDataInfoWithData 创建带数据的TableDataInfo对象 对应Java后端的TableDataInfo(List<?> list, long total)
func NewTableDataInfoWithData(list interface{}, total int64) *TableDataInfo {
	return &TableDataInfo{
		Code:  SUCCESS,
		Msg:   "查询成功",
		Rows:  list,
		Total: total,
	}
}

// NewAjaxResult 创建新的AjaxResult
func NewAjaxResult(code int, msg string, data interface{}) AjaxResult {
	result := make(AjaxResult)
	result["code"] = code
	result["msg"] = msg
	if data != nil {
		result["data"] = data
	}
	return result
}

// AjaxSuccess 返回成功消息 - 对应Java后端AjaxResult.success()
func AjaxSuccess() AjaxResult {
	return AjaxSuccessWithMessage("操作成功")
}

// AjaxSuccessWithData 返回成功数据 - 对应Java后端AjaxResult.success(Object data)
func AjaxSuccessWithData(data interface{}) AjaxResult {
	return AjaxSuccessWithDetailed("操作成功", data)
}

// AjaxSuccessWithMessage 返回成功消息 - 对应Java后端AjaxResult.success(String msg)
func AjaxSuccessWithMessage(msg string) AjaxResult {
	return AjaxSuccessWithDetailed(msg, nil)
}

// AjaxSuccessWithDetailed 返回成功消息 - 对应Java后端AjaxResult.success(String msg, Object data)
func AjaxSuccessWithDetailed(msg string, data interface{}) AjaxResult {
	return NewAjaxResult(SUCCESS, msg, data)
}

// AjaxWarn 返回警告消息 - 对应Java后端AjaxResult.warn(String msg)
func AjaxWarn(msg string) AjaxResult {
	return AjaxWarnWithData(msg, nil)
}

// AjaxWarnWithData 返回警告消息 - 对应Java后端AjaxResult.warn(String msg, Object data)
func AjaxWarnWithData(msg string, data interface{}) AjaxResult {
	return NewAjaxResult(WARN, msg, data)
}

// AjaxError 返回错误消息 - 对应Java后端AjaxResult.error()
func AjaxError() AjaxResult {
	return AjaxErrorWithMessage("操作失败")
}

// AjaxErrorWithMessage 返回错误消息 - 对应Java后端AjaxResult.error(String msg)
func AjaxErrorWithMessage(msg string) AjaxResult {
	return AjaxErrorWithData(msg, nil)
}

// AjaxErrorWithData 返回错误消息 - 对应Java后端AjaxResult.error(String msg, Object data)
func AjaxErrorWithData(msg string, data interface{}) AjaxResult {
	return NewAjaxResult(ERROR, msg, data)
}

// AjaxErrorWithCode 返回错误消息 - 对应Java后端AjaxResult.error(int code, String msg)
func AjaxErrorWithCode(code int, msg string) AjaxResult {
	return NewAjaxResult(code, msg, nil)
}

// Put 方便链式调用 - 对应Java后端AjaxResult.put(String key, Object value)
func (r AjaxResult) Put(key string, value interface{}) AjaxResult {
	r[key] = value
	return r
}

// IsSuccess 判断是否为成功响应 - 对应Java后端AjaxResult.isSuccess()
func (r AjaxResult) IsSuccess() bool {
	if code, ok := r["code"].(int); ok {
		return code == SUCCESS
	}
	return false
}

// IsError 判断是否为错误响应 - 对应Java后端AjaxResult.isError()
func (r AjaxResult) IsError() bool {
	if code, ok := r["code"].(int); ok {
		return code == ERROR
	}
	return false
}

// IsWarn 判断是否为警告响应 - 对应Java后端AjaxResult.isWarn()
func (r AjaxResult) IsWarn() bool {
	if code, ok := r["code"].(int); ok {
		return code == WARN
	}
	return false
}

// GetDataTable 响应请求分页数据 - 对应Java后端BaseController.getDataTable(List<?> list)
func GetDataTable(list interface{}, total int64) *TableDataInfo {
	return &TableDataInfo{
		Code:  SUCCESS,
		Msg:   "查询成功",
		Rows:  list,
		Total: total,
	}
}

// ToAjax 响应返回结果 - 对应Java后端BaseController.toAjax(int rows)
func ToAjax(rows int) AjaxResult {
	if rows > 0 {
		return AjaxSuccess()
	}
	return AjaxError()
}

// ToAjaxBool 响应返回结果 - 对应Java后端BaseController.toAjax(boolean result)
func ToAjaxBool(result bool) AjaxResult {
	if result {
		return AjaxSuccess()
	}
	return AjaxError()
}

// SendAjaxResult 发送AjaxResult响应
func SendAjaxResult(c *gin.Context, result AjaxResult) {
	c.JSON(http.StatusOK, result)
}

// SendTableDataInfo 发送TableDataInfo响应
func SendTableDataInfo(c *gin.Context, tableData *TableDataInfo) {
	c.JSON(http.StatusOK, tableData)
}

// IsSuccess 判断Result是否为成功响应 - 对应Java后端R.isSuccess(R<T> ret)
func (r *Result) IsSuccess() bool {
	return r.Code == SUCCESS
}

// IsError 判断Result是否为错误响应 - 对应Java后端R.isError(R<T> ret)
func (r *Result) IsError() bool {
	return r.Code == ERROR
}

// IsWarn 判断Result是否为警告响应
func (r *Result) IsWarn() bool {
	return r.Code == WARN
}

// IsSuccessResult 静态方法判断Result是否成功 - 对应Java后端R.isSuccess(R<T> ret)
func IsSuccessResult(result *Result) bool {
	return result != nil && result.Code == SUCCESS
}

// IsErrorResult 静态方法判断Result是否错误 - 对应Java后端R.isError(R<T> ret)
func IsErrorResult(result *Result) bool {
	return result != nil && result.Code == ERROR
}
