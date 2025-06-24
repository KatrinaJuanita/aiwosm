package monitor

import (
	"fmt"
	"time"
	"wosm/internal/repository/dao"
	"wosm/internal/repository/model"
)

// JobLogService 定时任务调度日志服务 对应Java后端的ISysJobLogService
type JobLogService struct {
	jobLogDao *dao.JobLogDao
}

// NewJobLogService 创建定时任务调度日志服务实例
func NewJobLogService() *JobLogService {
	return &JobLogService{
		jobLogDao: dao.NewJobLogDao(),
	}
}

// SelectJobLogList 查询定时任务调度日志列表 对应Java后端的selectJobLogList
func (s *JobLogService) SelectJobLogList(jobLog *model.SysJobLog, pageNum, pageSize int) ([]model.SysJobLog, int64, error) {
	fmt.Printf("JobLogService.SelectJobLogList: 查询定时任务调度日志列表, PageNum=%d, PageSize=%d\n", pageNum, pageSize)
	return s.jobLogDao.SelectJobLogListWithPage(jobLog, pageNum, pageSize)
}

// SelectJobLogListAll 查询所有定时任务调度日志列表（不分页） 对应Java后端的selectJobLogList
func (s *JobLogService) SelectJobLogListAll(jobLog *model.SysJobLog) ([]model.SysJobLog, error) {
	fmt.Printf("JobLogService.SelectJobLogListAll: 查询所有定时任务调度日志列表\n")
	return s.jobLogDao.SelectJobLogList(jobLog)
}

// SelectJobLogById 根据调度日志ID查询调度日志信息 对应Java后端的selectJobLogById
func (s *JobLogService) SelectJobLogById(jobLogId int64) (*model.SysJobLog, error) {
	fmt.Printf("JobLogService.SelectJobLogById: 查询定时任务调度日志信息, JobLogID=%d\n", jobLogId)
	return s.jobLogDao.SelectJobLogById(jobLogId)
}

// AddJobLog 新增定时任务调度日志 对应Java后端的addJobLog
func (s *JobLogService) AddJobLog(jobLog *model.SysJobLog) error {
	fmt.Printf("JobLogService.AddJobLog: 新增定时任务调度日志, JobName=%s\n", jobLog.JobName)

	// 参数验证
	if err := model.ValidateJobLog(jobLog, false); err != nil {
		return err
	}

	// 设置创建时间
	now := time.Now()
	jobLog.CreateTime = &now

	// 设置默认状态
	if jobLog.Status == "" {
		jobLog.Status = model.JobLogStatusNormal
	}

	return s.jobLogDao.InsertJobLog(jobLog)
}

// DeleteJobLogByIds 批量删除定时任务调度日志 对应Java后端的deleteJobLogByIds
func (s *JobLogService) DeleteJobLogByIds(jobLogIds []int64) error {
	fmt.Printf("JobLogService.DeleteJobLogByIds: 批量删除定时任务调度日志, JobLogIDs=%v\n", jobLogIds)

	if len(jobLogIds) == 0 {
		return fmt.Errorf("删除的日志ID不能为空")
	}

	return s.jobLogDao.DeleteJobLogByIds(jobLogIds)
}

// DeleteJobLogById 删除定时任务调度日志 对应Java后端的deleteJobLogById
func (s *JobLogService) DeleteJobLogById(jobLogId int64) error {
	fmt.Printf("JobLogService.DeleteJobLogById: 删除定时任务调度日志, JobLogID=%d\n", jobLogId)

	if jobLogId <= 0 {
		return fmt.Errorf("删除的日志ID不能为空")
	}

	return s.jobLogDao.DeleteJobLogById(jobLogId)
}

// CleanJobLog 清空定时任务调度日志 对应Java后端的cleanJobLog
func (s *JobLogService) CleanJobLog() error {
	fmt.Printf("JobLogService.CleanJobLog: 清空定时任务调度日志\n")
	return s.jobLogDao.CleanJobLog()
}

// AddJobLogByJob 根据任务信息添加调度日志 - 内部使用
func (s *JobLogService) AddJobLogByJob(job *model.SysJob, status string, message string, exceptionInfo string) error {
	jobLog := &model.SysJobLog{
		JobName:       job.JobName,
		JobGroup:      job.JobGroup,
		InvokeTarget:  job.InvokeTarget,
		JobMessage:    message,
		Status:        status,
		ExceptionInfo: exceptionInfo,
	}

	return s.AddJobLog(jobLog)
}

// AddJobLogSuccess 添加成功的调度日志
func (s *JobLogService) AddJobLogSuccess(job *model.SysJob, message string) error {
	return s.AddJobLogByJob(job, model.JobLogStatusNormal, message, "")
}

// AddJobLogError 添加失败的调度日志
func (s *JobLogService) AddJobLogError(job *model.SysJob, message string, exceptionInfo string) error {
	return s.AddJobLogByJob(job, model.JobLogStatusFail, message, exceptionInfo)
}
