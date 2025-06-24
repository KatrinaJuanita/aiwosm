package dao

import (
	"fmt"
	"wosm/internal/repository/model"
	"wosm/pkg/database"

	"gorm.io/gorm"
)

// JobLogDao 定时任务调度日志数据访问对象 对应Java后端的SysJobLogMapper
type JobLogDao struct {
	db *gorm.DB
}

// NewJobLogDao 创建定时任务调度日志DAO
func NewJobLogDao() *JobLogDao {
	return &JobLogDao{
		db: database.GetDB(),
	}
}

// SelectJobLogList 查询定时任务调度日志列表 对应Java后端的selectJobLogList
func (d *JobLogDao) SelectJobLogList(jobLog *model.SysJobLog) ([]model.SysJobLog, error) {
	var jobLogs []model.SysJobLog
	query := d.db.Model(&model.SysJobLog{})

	// 构建查询条件 - 严格按照Java后端的查询逻辑
	if jobLog.JobName != "" {
		query = query.Where("job_name LIKE ?", "%"+jobLog.JobName+"%")
	}
	if jobLog.JobGroup != "" {
		query = query.Where("job_group = ?", jobLog.JobGroup)
	}
	if jobLog.Status != "" {
		query = query.Where("status = ?", jobLog.Status)
	}
	if jobLog.InvokeTarget != "" {
		query = query.Where("invoke_target LIKE ?", "%"+jobLog.InvokeTarget+"%")
	}

	// 时间范围查询 - 对应Java后端的date_format查询逻辑
	if jobLog.StartTime != nil {
		// 对应Java后端的 date_format(create_time,'%Y%m%d') >= date_format(#{params.beginTime},'%Y%m%d')
		query = query.Where("CONVERT(date, create_time) >= CONVERT(date, ?)", jobLog.StartTime)
	}
	if jobLog.StopTime != nil {
		// 对应Java后端的 date_format(create_time,'%Y%m%d') <= date_format(#{params.endTime},'%Y%m%d')
		query = query.Where("CONVERT(date, create_time) <= CONVERT(date, ?)", jobLog.StopTime)
	}

	// 排序 - 对应Java后端的 order by create_time desc
	query = query.Order("create_time DESC")

	err := query.Find(&jobLogs).Error
	if err != nil {
		fmt.Printf("SelectJobLogList: 查询定时任务调度日志列表失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectJobLogList: 查询到定时任务调度日志数量=%d\n", len(jobLogs))
	return jobLogs, nil
}

// SelectJobLogAll 查询所有定时任务调度日志 对应Java后端的selectJobLogAll
func (d *JobLogDao) SelectJobLogAll() ([]model.SysJobLog, error) {
	var jobLogs []model.SysJobLog

	err := d.db.Model(&model.SysJobLog{}).Order("create_time DESC").Find(&jobLogs).Error
	if err != nil {
		fmt.Printf("SelectJobLogAll: 查询所有定时任务调度日志失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectJobLogAll: 查询到定时任务调度日志数量=%d\n", len(jobLogs))
	return jobLogs, nil
}

// SelectJobLogListWithPage 分页查询定时任务调度日志列表
func (d *JobLogDao) SelectJobLogListWithPage(jobLog *model.SysJobLog, pageNum, pageSize int) ([]model.SysJobLog, int64, error) {
	var jobLogs []model.SysJobLog
	var total int64

	// 构建查询条件
	query := d.db.Model(&model.SysJobLog{})
	if jobLog.JobName != "" {
		query = query.Where("job_name LIKE ?", "%"+jobLog.JobName+"%")
	}
	if jobLog.JobGroup != "" {
		query = query.Where("job_group = ?", jobLog.JobGroup)
	}
	if jobLog.Status != "" {
		query = query.Where("status = ?", jobLog.Status)
	}
	if jobLog.InvokeTarget != "" {
		query = query.Where("invoke_target LIKE ?", "%"+jobLog.InvokeTarget+"%")
	}
	if jobLog.StartTime != nil {
		// 对应Java后端的 date_format(create_time,'%Y%m%d') >= date_format(#{params.beginTime},'%Y%m%d')
		query = query.Where("CONVERT(date, create_time) >= CONVERT(date, ?)", jobLog.StartTime)
	}
	if jobLog.StopTime != nil {
		// 对应Java后端的 date_format(create_time,'%Y%m%d') <= date_format(#{params.endTime},'%Y%m%d')
		query = query.Where("CONVERT(date, create_time) <= CONVERT(date, ?)", jobLog.StopTime)
	}

	// 先查询总数
	err := query.Count(&total).Error
	if err != nil {
		fmt.Printf("SelectJobLogListWithPage: 查询定时任务调度日志总数失败: %v\n", err)
		return nil, 0, err
	}

	// 计算偏移量
	offset := (pageNum - 1) * pageSize

	// 分页查询数据 - 对应Java后端的 order by create_time desc
	err = query.Order("create_time DESC").Offset(offset).Limit(pageSize).Find(&jobLogs).Error
	if err != nil {
		fmt.Printf("SelectJobLogListWithPage: 分页查询定时任务调度日志列表失败: %v\n", err)
		return nil, 0, err
	}

	fmt.Printf("SelectJobLogListWithPage: 查询到定时任务调度日志数量=%d, 总数=%d\n", len(jobLogs), total)
	return jobLogs, total, nil
}

// SelectJobLogById 根据调度日志ID查询调度日志信息 对应Java后端的selectJobLogById
func (d *JobLogDao) SelectJobLogById(jobLogId int64) (*model.SysJobLog, error) {
	var jobLog model.SysJobLog
	err := d.db.Where("job_log_id = ?", jobLogId).First(&jobLog).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		fmt.Printf("SelectJobLogById: 查询定时任务调度日志失败: %v\n", err)
		return nil, err
	}

	return &jobLog, nil
}

// InsertJobLog 新增定时任务调度日志 对应Java后端的insertJobLog
func (d *JobLogDao) InsertJobLog(jobLog *model.SysJobLog) error {
	err := d.db.Create(jobLog).Error
	if err != nil {
		fmt.Printf("InsertJobLog: 新增定时任务调度日志失败: %v\n", err)
		return err
	}

	fmt.Printf("InsertJobLog: 新增定时任务调度日志成功, JobLogID=%d\n", jobLog.JobLogID)
	return nil
}

// DeleteJobLogByIds 批量删除定时任务调度日志 对应Java后端的deleteJobLogByIds
func (d *JobLogDao) DeleteJobLogByIds(jobLogIds []int64) error {
	err := d.db.Where("job_log_id IN ?", jobLogIds).Delete(&model.SysJobLog{}).Error
	if err != nil {
		fmt.Printf("DeleteJobLogByIds: 批量删除定时任务调度日志失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteJobLogByIds: 批量删除定时任务调度日志成功, 数量=%d\n", len(jobLogIds))
	return nil
}

// DeleteJobLogById 删除定时任务调度日志 对应Java后端的deleteJobLogById
func (d *JobLogDao) DeleteJobLogById(jobLogId int64) error {
	err := d.db.Where("job_log_id = ?", jobLogId).Delete(&model.SysJobLog{}).Error
	if err != nil {
		fmt.Printf("DeleteJobLogById: 删除定时任务调度日志失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteJobLogById: 删除定时任务调度日志成功, JobLogID=%d\n", jobLogId)
	return nil
}

// CleanJobLog 清空定时任务调度日志 对应Java后端的cleanJobLog
func (d *JobLogDao) CleanJobLog() error {
	err := d.db.Exec("TRUNCATE TABLE sys_job_log").Error
	if err != nil {
		fmt.Printf("CleanJobLog: 清空定时任务调度日志失败: %v\n", err)
		return err
	}

	fmt.Printf("CleanJobLog: 清空定时任务调度日志成功\n")
	return nil
}
