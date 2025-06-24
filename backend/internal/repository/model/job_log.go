package model

import (
	"fmt"
	"time"
)

// SysJobLog 定时任务调度日志表 对应Java后端的SysJobLog实体
// 严格按照Java后端真实数据库表结构定义（基于SqlServer_ry_20250522_COMPLETE.sql）：
// job_log_id, job_name, job_group, invoke_target, job_message, status, exception_info, create_time
type SysJobLog struct {
	JobLogID      int64      `gorm:"column:job_log_id;primaryKey;autoIncrement" json:"jobLogId"`      // 任务日志ID
	JobName       string     `gorm:"column:job_name;size:64;not null" json:"jobName"`                 // 任务名称
	JobGroup      string     `gorm:"column:job_group;size:64;not null" json:"jobGroup"`               // 任务组名
	InvokeTarget  string     `gorm:"column:invoke_target;size:500;not null" json:"invokeTarget"`      // 调用目标字符串
	JobMessage    string     `gorm:"column:job_message;size:500" json:"jobMessage"`                   // 日志信息
	Status        string     `gorm:"column:status;size:1;default:0" json:"status"`                    // 执行状态（0正常 1失败）
	ExceptionInfo string     `gorm:"column:exception_info;size:2000;default:''" json:"exceptionInfo"` // 异常信息
	CreateTime    *time.Time `gorm:"column:create_time" json:"createTime"`                            // 创建时间

	// 扩展字段（用于前端显示和查询）
	StartTime *time.Time `gorm:"-" json:"startTime"` // 开始时间（用于查询）
	StopTime  *time.Time `gorm:"-" json:"stopTime"`  // 停止时间（用于查询）
}

// TableName 指定表名
func (SysJobLog) TableName() string {
	return "sys_job_log"
}

// 状态常量 - 对应Java后端的ScheduleConstants
const (
	JobLogStatusNormal = "0" // 正常
	JobLogStatusFail   = "1" // 失败
)

// IsNormal 判断是否正常状态
func (j *SysJobLog) IsNormal() bool {
	return j.Status == JobLogStatusNormal
}

// IsFail 判断是否失败状态
func (j *SysJobLog) IsFail() bool {
	return j.Status == JobLogStatusFail
}

// GetStatusText 获取状态文本
func (j *SysJobLog) GetStatusText() string {
	switch j.Status {
	case JobLogStatusNormal:
		return "正常"
	case JobLogStatusFail:
		return "失败"
	default:
		return "未知"
	}
}

// ValidateJobLog 验证任务日志参数
func ValidateJobLog(jobLog *SysJobLog, isUpdate bool) error {
	if jobLog.JobName == "" {
		return fmt.Errorf("任务名称不能为空")
	}
	if len(jobLog.JobName) > 64 {
		return fmt.Errorf("任务名称长度不能超过64个字符")
	}

	if jobLog.JobGroup == "" {
		return fmt.Errorf("任务组名不能为空")
	}
	if len(jobLog.JobGroup) > 64 {
		return fmt.Errorf("任务组名长度不能超过64个字符")
	}

	if jobLog.InvokeTarget == "" {
		return fmt.Errorf("调用目标字符串不能为空")
	}
	if len(jobLog.InvokeTarget) > 500 {
		return fmt.Errorf("调用目标字符串长度不能超过500个字符")
	}

	if len(jobLog.JobMessage) > 500 {
		return fmt.Errorf("日志信息长度不能超过500个字符")
	}

	if len(jobLog.ExceptionInfo) > 2000 {
		return fmt.Errorf("异常信息长度不能超过2000个字符")
	}

	// 验证状态
	if jobLog.Status != "" && jobLog.Status != JobLogStatusNormal && jobLog.Status != JobLogStatusFail {
		return fmt.Errorf("执行状态值无效")
	}

	return nil
}

// String 字符串表示 对应Java后端的toString方法
func (j *SysJobLog) String() string {
	return fmt.Sprintf("SysJobLog{JobLogID=%d, JobName='%s', JobGroup='%s', JobMessage='%s', Status='%s', ExceptionInfo='%s', StartTime=%v, StopTime=%v}",
		j.JobLogID, j.JobName, j.JobGroup, j.JobMessage, j.Status, j.ExceptionInfo, j.StartTime, j.StopTime)
}

// SysJobLogExport 定时任务调度日志导出结构体 对应Java后端的Excel导出格式
type SysJobLogExport struct {
	JobLogID      int64  `excel:"name:日志编号;sort:1"`
	JobName       string `excel:"name:任务名称;sort:2"`
	JobGroup      string `excel:"name:任务组名;sort:3"`
	InvokeTarget  string `excel:"name:调用目标;sort:4"`
	JobMessage    string `excel:"name:日志信息;sort:5"`
	Status        string `excel:"name:执行状态;sort:6;readConverterExp:0=成功,1=失败"`
	ExceptionInfo string `excel:"name:异常信息;sort:7"`
	CreateTime    string `excel:"name:执行时间;sort:8"`
}
