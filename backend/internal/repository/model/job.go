package model

import (
	"fmt"
	"time"
)

// SysJob 定时任务调度表 对应Java后端的SysJob实体
// 严格按照Java后端真实数据库表结构定义（基于SqlServer_ry_20250522_COMPLETE.sql）：
// job_id, job_name, job_group, invoke_target, cron_expression, misfire_policy, concurrent, status, create_by, create_time, update_by, update_time, remark
type SysJob struct {
	JobID          int64      `gorm:"column:job_id;primaryKey;autoIncrement" json:"jobId"`                                     // 任务ID
	JobName        string     `gorm:"column:job_name;size:64;uniqueIndex:idx_job_name_group" json:"jobName"`                   // 任务名称
	JobGroup       string     `gorm:"column:job_group;size:64;default:DEFAULT;uniqueIndex:idx_job_name_group" json:"jobGroup"` // 任务组名
	InvokeTarget   string     `gorm:"column:invoke_target;size:500;not null" json:"invokeTarget"`                              // 调用目标字符串
	CronExpression string     `gorm:"column:cron_expression;size:255;default:''" json:"cronExpression"`                        // cron执行表达式
	MisfirePolicy  string     `gorm:"column:misfire_policy;size:20;default:'3'" json:"misfirePolicy"`                          // 计划执行错误策略（1立即执行 2执行一次 3放弃执行）
	Concurrent     string     `gorm:"column:concurrent;size:1;default:1" json:"concurrent"`                                    // 是否并发执行（0允许 1禁止）
	Status         string     `gorm:"column:status;size:1;default:0" json:"status"`                                            // 状态（0正常 1暂停）
	CreateBy       string     `gorm:"column:create_by;size:64;default:''" json:"createBy"`                                     // 创建者
	CreateTime     *time.Time `gorm:"column:create_time" json:"createTime"`                                                    // 创建时间
	UpdateBy       string     `gorm:"column:update_by;size:64;default:''" json:"updateBy"`                                     // 更新者
	UpdateTime     *time.Time `gorm:"column:update_time" json:"updateTime"`                                                    // 更新时间
	Remark         string     `gorm:"column:remark;size:500;default:''" json:"remark"`                                         // 备注信息

	// 查询条件字段（不映射到数据库）
	BeginTime string `gorm:"-" json:"beginTime"` // 开始时间
	EndTime   string `gorm:"-" json:"endTime"`   // 结束时间

	// 扩展字段（不映射到数据库）
	NextValidTime *time.Time `gorm:"-" json:"nextValidTime"` // 下次执行时间
}

// TableName 指定表名
func (SysJob) TableName() string {
	return "sys_job"
}

// 任务状态常量
const (
	JobStatusNormal = "0" // 正常
	JobStatusPause  = "1" // 暂停
)

// GetStatusName 获取状态名称
func GetJobStatusName(status string) string {
	switch status {
	case JobStatusNormal:
		return "正常"
	case JobStatusPause:
		return "暂停"
	default:
		return "未知"
	}
}

// 并发执行常量
const (
	ConcurrentAllow  = "0" // 允许
	ConcurrentForbid = "1" // 禁止
)

// GetConcurrentName 获取并发执行名称
func GetConcurrentName(concurrent string) string {
	switch concurrent {
	case ConcurrentAllow:
		return "允许"
	case ConcurrentForbid:
		return "禁止"
	default:
		return "未知"
	}
}

// 计划执行错误策略常量 对应Java后端的ScheduleConstants
const (
	MisfirePolicyImmediate = "1" // 立即触发执行 对应Java后端的MISFIRE_IGNORE_MISFIRES
	MisfirePolicyOnce      = "2" // 触发一次执行 对应Java后端的MISFIRE_FIRE_AND_PROCEED
	MisfirePolicyDefault   = "3" // 默认（不触发立即执行）对应Java后端的MISFIRE_DO_NOTHING
)

// GetMisfirePolicyName 获取计划执行错误策略名称
func GetMisfirePolicyName(policy string) string {
	switch policy {
	case MisfirePolicyImmediate:
		return "立即触发执行"
	case MisfirePolicyOnce:
		return "触发一次执行"
	case MisfirePolicyDefault:
		return "不触发立即执行"
	default:
		return "未知"
	}
}

// 任务组名常量
const (
	JobGroupDefault = "DEFAULT" // 默认组
	JobGroupSystem  = "SYSTEM"  // 系统组
)

// IsValidStatus 检查状态是否有效
func (j *SysJob) IsValidStatus() bool {
	return j.Status == JobStatusNormal || j.Status == JobStatusPause
}

// IsValidConcurrent 检查并发执行设置是否有效
func (j *SysJob) IsValidConcurrent() bool {
	return j.Concurrent == ConcurrentAllow || j.Concurrent == ConcurrentForbid
}

// IsValidMisfirePolicy 检查计划执行错误策略是否有效
func (j *SysJob) IsValidMisfirePolicy() bool {
	return j.MisfirePolicy == MisfirePolicyDefault ||
		j.MisfirePolicy == MisfirePolicyImmediate ||
		j.MisfirePolicy == MisfirePolicyOnce
}

// IsNormal 检查任务是否为正常状态
func (j *SysJob) IsNormal() bool {
	return j.Status == JobStatusNormal
}

// IsPaused 检查任务是否为暂停状态
func (j *SysJob) IsPaused() bool {
	return j.Status == JobStatusPause
}

// IsConcurrentAllowed 检查是否允许并发执行
func (j *SysJob) IsConcurrentAllowed() bool {
	return j.Concurrent == ConcurrentAllow
}

// ValidateJob 验证任务参数 对应Java后端的@NotBlank和@Size注解验证
func ValidateJob(job *SysJob, isUpdate bool) error {
	if job.JobName == "" {
		return fmt.Errorf("任务名称不能为空")
	}
	if len(job.JobName) > 64 {
		return fmt.Errorf("任务名称不能超过64个字符")
	}

	if job.InvokeTarget == "" {
		return fmt.Errorf("调用目标字符串不能为空")
	}
	if len(job.InvokeTarget) > 500 {
		return fmt.Errorf("调用目标字符串长度不能超过500个字符")
	}

	if job.CronExpression == "" {
		return fmt.Errorf("Cron执行表达式不能为空")
	}
	if len(job.CronExpression) > 255 {
		return fmt.Errorf("Cron执行表达式不能超过255个字符")
	}

	// 验证状态
	if job.Status != "" && !job.IsValidStatus() {
		return fmt.Errorf("任务状态值无效")
	}

	// 验证并发设置
	if job.Concurrent != "" && !job.IsValidConcurrent() {
		return fmt.Errorf("并发执行设置无效")
	}

	// 验证计划执行错误策略
	if job.MisfirePolicy != "" && !job.IsValidMisfirePolicy() {
		return fmt.Errorf("计划执行错误策略无效")
	}

	return nil
}

// GetNextValidTime 获取下次执行时间 对应Java后端的getNextValidTime方法
func (j *SysJob) GetNextValidTime() *time.Time {
	if j.CronExpression == "" {
		return nil
	}

	// 这里需要使用cron工具包来计算下次执行时间
	// 由于导入问题，暂时返回nil，在实际使用时需要导入cronUtils包
	// nextTime, err := cronUtils.GetNextExecution(j.CronExpression)
	// if err != nil {
	//     return nil
	// }
	// return nextTime
	return nil
}

// String 字符串表示 对应Java后端的toString方法
func (j *SysJob) String() string {
	return fmt.Sprintf("SysJob{JobID=%d, JobName='%s', JobGroup='%s', CronExpression='%s', NextValidTime=%v, MisfirePolicy='%s', Concurrent='%s', Status='%s', CreateBy='%s', CreateTime=%v, UpdateBy='%s', UpdateTime=%v, Remark='%s'}",
		j.JobID, j.JobName, j.JobGroup, j.CronExpression, j.GetNextValidTime(), j.MisfirePolicy, j.Concurrent, j.Status, j.CreateBy, j.CreateTime, j.UpdateBy, j.UpdateTime, j.Remark)
}

// SysJobExport 定时任务导出结构体 对应Java后端的Excel导出格式
type SysJobExport struct {
	JobID          int64  `excel:"name:任务编号;sort:1"`
	JobName        string `excel:"name:任务名称;sort:2"`
	JobGroup       string `excel:"name:任务组名;sort:3"`
	InvokeTarget   string `excel:"name:调用目标字符串;sort:4"`
	CronExpression string `excel:"name:cron执行表达式;sort:5"`
	MisfirePolicy  string `excel:"name:计划执行错误策略;sort:6;readConverterExp:0=默认,1=立即触发执行,2=触发一次执行,3=不触发立即执行"`
	Concurrent     string `excel:"name:是否并发;sort:7;readConverterExp:0=允许,1=禁止"`
	Status         string `excel:"name:任务状态;sort:8;readConverterExp:0=正常,1=暂停"`
	CreateBy       string `excel:"name:创建者;sort:9"`
	CreateTime     string `excel:"name:创建时间;sort:10"`
	UpdateBy       string `excel:"name:更新者;sort:11"`
	UpdateTime     string `excel:"name:更新时间;sort:12"`
	Remark         string `excel:"name:备注;sort:13"`
}
