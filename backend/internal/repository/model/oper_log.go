package model

import (
	"time"
)

// SysOperLog 操作日志记录表 对应Java后端的SysOperLog实体
// 严格按照真实数据库表结构定义（基于SqlServer_ry_20250522_COMPLETE.sql）：
// 数据库表只有17个字段：oper_id, title, business_type, method, request_method, operator_type, oper_name, dept_name,
// oper_url, oper_ip, oper_location, oper_param, json_result, status, error_msg, oper_time, cost_time
// 注意：虽然Java后端继承BaseEntity，但数据库表实际没有BaseEntity的字段
type SysOperLog struct {
	OperID        int64      `gorm:"column:oper_id;primaryKey;autoIncrement" json:"operId" excel:"name:操作序号;sort:1;cellType:numeric"`                                                    // 日志主键
	Title         string     `gorm:"column:title;size:50" json:"title" excel:"name:系统模块;sort:2"`                                                                                         // 模块标题
	BusinessType  int        `gorm:"column:business_type;default:0" json:"businessType" excel:"name:操作类型;sort:3;readConverterExp:0=其它,1=新增,2=修改,3=删除,4=授权,5=导出,6=导入,7=强退,8=生成代码,9=清空数据"` // 业务类型（0其它 1新增 2修改 3删除）
	Method        string     `gorm:"column:method;size:200" json:"method" excel:"name:请求方法;sort:4"`                                                                                      // 方法名称
	RequestMethod string     `gorm:"column:request_method;size:10" json:"requestMethod" excel:"name:请求方式;sort:5"`                                                                        // 请求方式
	OperatorType  int        `gorm:"column:operator_type;default:0" json:"operatorType" excel:"name:操作类别;sort:6;readConverterExp:0=其它,1=后台用户,2=手机端用户"`                                   // 操作类别（0其它 1后台用户 2手机端用户）
	OperName      string     `gorm:"column:oper_name;size:50" json:"operName" excel:"name:操作人员;sort:7"`                                                                                  // 操作人员
	DeptName      string     `gorm:"column:dept_name;size:50" json:"deptName" excel:"name:部门名称;sort:8"`                                                                                  // 部门名称
	OperURL       string     `gorm:"column:oper_url;size:255" json:"operUrl" excel:"name:请求地址;sort:9;width:30"`                                                                          // 请求URL
	OperIP        string     `gorm:"column:oper_ip;size:128" json:"operIp" excel:"name:主机地址;sort:10"`                                                                                    // 主机地址
	OperLocation  string     `gorm:"column:oper_location;size:255" json:"operLocation" excel:"name:操作地点;sort:11"`                                                                        // 操作地点
	OperParam     string     `gorm:"column:oper_param;type:text" json:"operParam" excel:"name:请求参数;sort:12;width:50;type:export"`                                                        // 请求参数
	JSONResult    string     `gorm:"column:json_result;type:text" json:"jsonResult" excel:"name:返回参数;sort:13;width:50;type:export"`                                                      // 返回参数
	Status        int        `gorm:"column:status;default:0" json:"status" excel:"name:操作状态;sort:14;readConverterExp:0=正常,1=异常"`                                                         // 操作状态（0正常 1异常）
	ErrorMsg      string     `gorm:"column:error_msg;type:text" json:"errorMsg" excel:"name:错误消息;sort:15;width:50;type:export"`                                                          // 错误消息
	OperTime      *time.Time `gorm:"column:oper_time" json:"operTime" excel:"name:操作时间;sort:16;width:30;dateFormat:yyyy-MM-dd HH:mm:ss"`                                                 // 操作时间
	CostTime      int64      `gorm:"column:cost_time;default:0" json:"costTime" excel:"name:消耗时间;sort:17;suffix:毫秒"`                                                                     // 消耗时间

	// 查询条件字段（不映射到数据库）
	BusinessTypes []int  `gorm:"-" json:"businessTypes"` // 业务类型数组
	BeginTime     string `gorm:"-" json:"beginTime"`     // 开始时间
	EndTime       string `gorm:"-" json:"endTime"`       // 结束时间
}

// TableName 指定表名
func (SysOperLog) TableName() string {
	return "sys_oper_log"
}

// 业务类型常量 对应Java后端的BusinessType枚举
const (
	BusinessTypeOther   = 0 // 其它
	BusinessTypeInsert  = 1 // 新增
	BusinessTypeUpdate  = 2 // 修改
	BusinessTypeDelete  = 3 // 删除
	BusinessTypeGrant   = 4 // 授权
	BusinessTypeExport  = 5 // 导出
	BusinessTypeImport  = 6 // 导入
	BusinessTypeForce   = 7 // 强退
	BusinessTypeGencode = 8 // 生成代码
	BusinessTypeClean   = 9 // 清空数据
)

// 操作类别常量 对应Java后端的OperatorType枚举
const (
	OperatorTypeOther  = 0 // 其它
	OperatorTypeManage = 1 // 后台用户
	OperatorTypeMobile = 2 // 手机端用户
)

// 操作状态常量
const (
	OperStatusSuccess = 0 // 正常
	OperStatusFail    = 1 // 异常
)

// GetBusinessTypeName 获取业务类型名称
func GetBusinessTypeName(businessType int) string {
	switch businessType {
	case BusinessTypeOther:
		return "其它"
	case BusinessTypeInsert:
		return "新增"
	case BusinessTypeUpdate:
		return "修改"
	case BusinessTypeDelete:
		return "删除"
	case BusinessTypeGrant:
		return "授权"
	case BusinessTypeExport:
		return "导出"
	case BusinessTypeImport:
		return "导入"
	case BusinessTypeForce:
		return "强退"
	case BusinessTypeGencode:
		return "生成代码"
	case BusinessTypeClean:
		return "清空数据"
	default:
		return "未知"
	}
}

// GetOperatorTypeName 获取操作类别名称
func GetOperatorTypeName(operatorType int) string {
	switch operatorType {
	case OperatorTypeOther:
		return "其它"
	case OperatorTypeManage:
		return "后台用户"
	case OperatorTypeMobile:
		return "手机端用户"
	default:
		return "未知"
	}
}

// GetStatusName 获取状态名称
func GetStatusName(status int) string {
	switch status {
	case OperStatusSuccess:
		return "正常"
	case OperStatusFail:
		return "异常"
	default:
		return "未知"
	}
}
