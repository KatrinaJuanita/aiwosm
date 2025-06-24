package export

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ParseFormParams 解析POST请求表单参数 对应Java后端的参数绑定
// 前端使用application/x-www-form-urlencoded格式发送数据
func ParseFormParams(ctx *gin.Context) (url.Values, error) {
	// 解析请求体
	if err := ctx.Request.ParseForm(); err != nil {
		return nil, err
	}

	return ctx.Request.PostForm, nil
}

// GetStringParam 获取字符串参数
func GetStringParam(params url.Values, key string) string {
	if values, exists := params[key]; exists && len(values) > 0 {
		return strings.TrimSpace(values[0])
	}
	return ""
}

// GetIntParam 获取整数参数
func GetIntParam(params url.Values, key string) *int {
	if str := GetStringParam(params, key); str != "" {
		if val, err := strconv.Atoi(str); err == nil {
			return &val
		}
	}
	return nil
}

// GetInt64Param 获取int64参数
func GetInt64Param(params url.Values, key string) *int64 {
	if str := GetStringParam(params, key); str != "" {
		if val, err := strconv.ParseInt(str, 10, 64); err == nil {
			return &val
		}
	}
	return nil
}

// GetTimeParam 解析时间参数 支持多种格式
func GetTimeParam(params url.Values, key string) *time.Time {
	if str := GetStringParam(params, key); str != "" {
		// 尝试多种时间格式
		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05.000Z",
			"2006-01-02",
		}

		for _, format := range formats {
			if t, err := time.Parse(format, str); err == nil {
				return &t
			}
		}
	}
	return nil
}

// GetTimeRange 获取时间范围参数
func GetTimeRange(params url.Values) (*time.Time, *time.Time) {
	beginTime := GetTimeParam(params, "beginTime")
	endTime := GetTimeParam(params, "endTime")

	// 如果有结束时间，设置为当天的23:59:59
	if endTime != nil {
		endOfDay := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 999999999, endTime.Location())
		endTime = &endOfDay
	}

	return beginTime, endTime
}

// UserQueryParams 用户查询参数 对应Java后端的SysUser
type UserQueryParams struct {
	UserName    string     `form:"userName"`
	Phonenumber string     `form:"phonenumber"`
	Status      string     `form:"status"`
	DeptID      *int64     `form:"deptId"`
	BeginTime   *time.Time `form:"beginTime"`
	EndTime     *time.Time `form:"endTime"`
}

// ParseUserQueryParams 解析用户查询参数
func ParseUserQueryParams(params url.Values) *UserQueryParams {
	beginTime, endTime := GetTimeRange(params)

	return &UserQueryParams{
		UserName:    GetStringParam(params, "userName"),
		Phonenumber: GetStringParam(params, "phonenumber"),
		Status:      GetStringParam(params, "status"),
		DeptID:      GetInt64Param(params, "deptId"),
		BeginTime:   beginTime,
		EndTime:     endTime,
	}
}

// RoleQueryParams 角色查询参数 对应Java后端的SysRole
type RoleQueryParams struct {
	RoleName  string     `form:"roleName"`
	RoleKey   string     `form:"roleKey"`
	Status    string     `form:"status"`
	BeginTime *time.Time `form:"beginTime"`
	EndTime   *time.Time `form:"endTime"`
}

// ParseRoleQueryParams 解析角色查询参数
func ParseRoleQueryParams(params url.Values) *RoleQueryParams {
	beginTime, endTime := GetTimeRange(params)

	return &RoleQueryParams{
		RoleName:  GetStringParam(params, "roleName"),
		RoleKey:   GetStringParam(params, "roleKey"),
		Status:    GetStringParam(params, "status"),
		BeginTime: beginTime,
		EndTime:   endTime,
	}
}

// PostQueryParams 岗位查询参数 对应Java后端的SysPost
type PostQueryParams struct {
	PostCode  string     `form:"postCode"`
	PostName  string     `form:"postName"`
	Status    string     `form:"status"`
	BeginTime *time.Time `form:"beginTime"`
	EndTime   *time.Time `form:"endTime"`
}

// ParsePostQueryParams 解析岗位查询参数
func ParsePostQueryParams(params url.Values) *PostQueryParams {
	beginTime, endTime := GetTimeRange(params)

	return &PostQueryParams{
		PostCode:  GetStringParam(params, "postCode"),
		PostName:  GetStringParam(params, "postName"),
		Status:    GetStringParam(params, "status"),
		BeginTime: beginTime,
		EndTime:   endTime,
	}
}

// DictTypeQueryParams 字典类型查询参数 对应Java后端的SysDictType
type DictTypeQueryParams struct {
	DictName  string     `form:"dictName"`
	DictType  string     `form:"dictType"`
	Status    string     `form:"status"`
	BeginTime *time.Time `form:"beginTime"`
	EndTime   *time.Time `form:"endTime"`
}

// ParseDictTypeQueryParams 解析字典类型查询参数
func ParseDictTypeQueryParams(params url.Values) *DictTypeQueryParams {
	beginTime, endTime := GetTimeRange(params)

	return &DictTypeQueryParams{
		DictName:  GetStringParam(params, "dictName"),
		DictType:  GetStringParam(params, "dictType"),
		Status:    GetStringParam(params, "status"),
		BeginTime: beginTime,
		EndTime:   endTime,
	}
}

// DictDataQueryParams 字典数据查询参数 对应Java后端的SysDictData
type DictDataQueryParams struct {
	DictLabel string     `form:"dictLabel"`
	DictValue string     `form:"dictValue"`
	DictType  string     `form:"dictType"`
	Status    string     `form:"status"`
	BeginTime *time.Time `form:"beginTime"`
	EndTime   *time.Time `form:"endTime"`
}

// ParseDictDataQueryParams 解析字典数据查询参数
func ParseDictDataQueryParams(params url.Values) *DictDataQueryParams {
	beginTime, endTime := GetTimeRange(params)

	return &DictDataQueryParams{
		DictLabel: GetStringParam(params, "dictLabel"),
		DictValue: GetStringParam(params, "dictValue"),
		DictType:  GetStringParam(params, "dictType"),
		Status:    GetStringParam(params, "status"),
		BeginTime: beginTime,
		EndTime:   endTime,
	}
}

// ConfigQueryParams 参数配置查询参数 对应Java后端的SysConfig
type ConfigQueryParams struct {
	ConfigName string     `form:"configName"`
	ConfigKey  string     `form:"configKey"`
	ConfigType string     `form:"configType"`
	BeginTime  *time.Time `form:"beginTime"`
	EndTime    *time.Time `form:"endTime"`
}

// ParseConfigQueryParams 解析参数配置查询参数
func ParseConfigQueryParams(params url.Values) *ConfigQueryParams {
	beginTime, endTime := GetTimeRange(params)

	return &ConfigQueryParams{
		ConfigName: GetStringParam(params, "configName"),
		ConfigKey:  GetStringParam(params, "configKey"),
		ConfigType: GetStringParam(params, "configType"),
		BeginTime:  beginTime,
		EndTime:    endTime,
	}
}

// OperLogQueryParams 操作日志查询参数 对应Java后端的SysOperLog
type OperLogQueryParams struct {
	Title        string     `form:"title"`
	OperName     string     `form:"operName"`
	BusinessType *int       `form:"businessType"`
	Status       *int       `form:"status"`
	BeginTime    *time.Time `form:"beginTime"`
	EndTime      *time.Time `form:"endTime"`
}

// ParseOperLogQueryParams 解析操作日志查询参数
func ParseOperLogQueryParams(params url.Values) *OperLogQueryParams {
	beginTime, endTime := GetTimeRange(params)

	return &OperLogQueryParams{
		Title:        GetStringParam(params, "title"),
		OperName:     GetStringParam(params, "operName"),
		BusinessType: GetIntParam(params, "businessType"),
		Status:       GetIntParam(params, "status"),
		BeginTime:    beginTime,
		EndTime:      endTime,
	}
}

// LoginLogQueryParams 登录日志查询参数 对应Java后端的SysLogininfor
type LoginLogQueryParams struct {
	Ipaddr    string     `form:"ipaddr"`
	UserName  string     `form:"userName"`
	Status    string     `form:"status"`
	BeginTime *time.Time `form:"beginTime"`
	EndTime   *time.Time `form:"endTime"`
}

// ParseLoginLogQueryParams 解析登录日志查询参数
func ParseLoginLogQueryParams(params url.Values) *LoginLogQueryParams {
	beginTime, endTime := GetTimeRange(params)

	return &LoginLogQueryParams{
		Ipaddr:    GetStringParam(params, "ipaddr"),
		UserName:  GetStringParam(params, "userName"),
		Status:    GetStringParam(params, "status"),
		BeginTime: beginTime,
		EndTime:   endTime,
	}
}
