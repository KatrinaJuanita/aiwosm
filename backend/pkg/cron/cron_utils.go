package cron

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// CronUtils Cron表达式工具类 对应Java后端的CronUtils
type CronUtils struct{}

// NewCronUtils 创建Cron工具实例
func NewCronUtils() *CronUtils {
	return &CronUtils{}
}

// IsValid 校验cron表达式是否有效 对应Java后端的isValid方法
func IsValid(cronExpression string) bool {
	if cronExpression == "" {
		return false
	}

	// 创建cron解析器
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	
	// 尝试解析cron表达式
	_, err := parser.Parse(cronExpression)
	if err != nil {
		fmt.Printf("IsValid: Cron表达式无效: %s, 错误: %v\n", cronExpression, err)
		return false
	}

	fmt.Printf("IsValid: Cron表达式有效: %s\n", cronExpression)
	return true
}

// GetNextExecution 获取下次执行时间 对应Java后端的getNextExecution方法
func GetNextExecution(cronExpression string) (*time.Time, error) {
	if cronExpression == "" {
		return nil, fmt.Errorf("cron表达式不能为空")
	}

	// 创建cron解析器
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	
	// 解析cron表达式
	schedule, err := parser.Parse(cronExpression)
	if err != nil {
		return nil, fmt.Errorf("解析cron表达式失败: %v", err)
	}

	// 获取下次执行时间
	nextTime := schedule.Next(time.Now())
	return &nextTime, nil
}

// GetNextExecutions 获取未来N次执行时间
func GetNextExecutions(cronExpression string, count int) ([]time.Time, error) {
	if cronExpression == "" {
		return nil, fmt.Errorf("cron表达式不能为空")
	}

	if count <= 0 {
		return nil, fmt.Errorf("执行次数必须大于0")
	}

	// 创建cron解析器
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	
	// 解析cron表达式
	schedule, err := parser.Parse(cronExpression)
	if err != nil {
		return nil, fmt.Errorf("解析cron表达式失败: %v", err)
	}

	// 获取未来N次执行时间
	var executions []time.Time
	currentTime := time.Now()
	
	for i := 0; i < count; i++ {
		nextTime := schedule.Next(currentTime)
		executions = append(executions, nextTime)
		currentTime = nextTime
	}

	return executions, nil
}

// ValidateAndGetNext 验证cron表达式并获取下次执行时间
func ValidateAndGetNext(cronExpression string) (*time.Time, error) {
	if !IsValid(cronExpression) {
		return nil, fmt.Errorf("cron表达式无效: %s", cronExpression)
	}

	return GetNextExecution(cronExpression)
}

// GetCronDescription 获取cron表达式的描述（简化版本）
func GetCronDescription(cronExpression string) string {
	if cronExpression == "" {
		return "无效的cron表达式"
	}

	// 简化的描述逻辑，实际项目中可以使用更复杂的解析
	switch cronExpression {
	case "0 0 0 * * ?":
		return "每天0点执行"
	case "0 0 12 * * ?":
		return "每天12点执行"
	case "0 0/5 * * * ?":
		return "每5分钟执行"
	case "0 0 0 1 * ?":
		return "每月1号0点执行"
	case "0 0 0 ? * MON":
		return "每周一0点执行"
	default:
		return "自定义时间执行"
	}
}

// CommonCronExpressions 常用的cron表达式
var CommonCronExpressions = map[string]string{
	"每秒执行":     "* * * * * ?",
	"每分钟执行":    "0 * * * * ?",
	"每5分钟执行":   "0 0/5 * * * ?",
	"每10分钟执行":  "0 0/10 * * * ?",
	"每30分钟执行":  "0 0/30 * * * ?",
	"每小时执行":    "0 0 * * * ?",
	"每天0点执行":   "0 0 0 * * ?",
	"每天12点执行":  "0 0 12 * * ?",
	"每周一0点执行":  "0 0 0 ? * MON",
	"每月1号0点执行": "0 0 0 1 * ?",
	"每年1月1号0点执行": "0 0 0 1 1 ?",
}

// GetCommonCronExpressions 获取常用cron表达式列表
func GetCommonCronExpressions() map[string]string {
	return CommonCronExpressions
}
