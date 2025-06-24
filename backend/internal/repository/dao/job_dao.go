package dao

import (
	"fmt"
	"wosm/internal/repository/model"
	"wosm/pkg/database"

	"gorm.io/gorm"
)

// JobDao 定时任务数据访问层 对应Java后端的SysJobMapper
type JobDao struct {
	db *gorm.DB
}

// NewJobDao 创建定时任务数据访问层实例
func NewJobDao() *JobDao {
	return &JobDao{
		db: database.GetDB(),
	}
}

// SelectJobList 查询调度任务日志集合 对应Java后端的selectJobList
func (d *JobDao) SelectJobList(job *model.SysJob) ([]model.SysJob, error) {
	var jobList []model.SysJob
	query := d.db.Model(&model.SysJob{})

	// 构建查询条件
	if job.JobName != "" {
		query = query.Where("job_name LIKE ?", "%"+job.JobName+"%")
	}
	if job.JobGroup != "" {
		query = query.Where("job_group = ?", job.JobGroup)
	}
	if job.Status != "" {
		query = query.Where("status = ?", job.Status)
	}
	if job.InvokeTarget != "" {
		query = query.Where("invoke_target LIKE ?", "%"+job.InvokeTarget+"%")
	}
	if job.BeginTime != "" {
		query = query.Where("create_time >= ?", job.BeginTime)
	}
	if job.EndTime != "" {
		query = query.Where("create_time <= ?", job.EndTime)
	}

	err := query.Order("job_id DESC").Find(&jobList).Error
	if err != nil {
		fmt.Printf("SelectJobList: 查询定时任务列表失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectJobList: 查询到定时任务数量=%d\n", len(jobList))
	return jobList, nil
}

// SelectJobAll 查询所有调度任务 对应Java后端的selectJobAll
func (d *JobDao) SelectJobAll() ([]model.SysJob, error) {
	var jobList []model.SysJob
	err := d.db.Find(&jobList).Error
	if err != nil {
		fmt.Printf("SelectJobAll: 查询所有定时任务失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectJobAll: 查询到定时任务数量=%d\n", len(jobList))
	return jobList, nil
}

// SelectJobById 通过调度ID查询调度任务信息 对应Java后端的selectJobById
func (d *JobDao) SelectJobById(jobId int64) (*model.SysJob, error) {
	var job model.SysJob
	err := d.db.Where("job_id = ?", jobId).First(&job).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Printf("SelectJobById: 定时任务不存在, JobID=%d\n", jobId)
			return nil, nil
		}
		fmt.Printf("SelectJobById: 查询定时任务失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectJobById: 查询定时任务成功, JobID=%d, JobName=%s\n", job.JobID, job.JobName)
	return &job, nil
}

// InsertJob 新增调度任务信息 对应Java后端的insertJob
func (d *JobDao) InsertJob(job *model.SysJob) error {
	err := d.db.Create(job).Error
	if err != nil {
		fmt.Printf("InsertJob: 新增定时任务失败: %v\n", err)
		return err
	}

	fmt.Printf("InsertJob: 新增定时任务成功, JobID=%d, JobName=%s\n", job.JobID, job.JobName)
	return nil
}

// UpdateJob 修改调度任务信息 对应Java后端的updateJob
func (d *JobDao) UpdateJob(job *model.SysJob) error {
	err := d.db.Where("job_id = ?", job.JobID).Updates(job).Error
	if err != nil {
		fmt.Printf("UpdateJob: 修改定时任务失败: %v\n", err)
		return err
	}

	fmt.Printf("UpdateJob: 修改定时任务成功, JobID=%d, JobName=%s\n", job.JobID, job.JobName)
	return nil
}

// DeleteJobById 通过调度ID删除调度任务信息 对应Java后端的deleteJobById
func (d *JobDao) DeleteJobById(jobId int64) error {
	err := d.db.Where("job_id = ?", jobId).Delete(&model.SysJob{}).Error
	if err != nil {
		fmt.Printf("DeleteJobById: 删除定时任务失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteJobById: 删除定时任务成功, JobID=%d\n", jobId)
	return nil
}

// DeleteJobByIds 批量删除调度任务信息 对应Java后端的deleteJobByIds
func (d *JobDao) DeleteJobByIds(jobIds []int64) error {
	err := d.db.Where("job_id IN ?", jobIds).Delete(&model.SysJob{}).Error
	if err != nil {
		fmt.Printf("DeleteJobByIds: 批量删除定时任务失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteJobByIds: 批量删除定时任务成功, 数量=%d\n", len(jobIds))
	return nil
}

// CheckJobNameUnique 检查任务名称是否唯一
func (d *JobDao) CheckJobNameUnique(jobName, jobGroup string, jobId int64) (bool, error) {
	var count int64
	query := d.db.Model(&model.SysJob{}).Where("job_name = ? AND job_group = ?", jobName, jobGroup)

	// 如果是更新操作，排除当前记录
	if jobId > 0 {
		query = query.Where("job_id != ?", jobId)
	}

	err := query.Count(&count).Error
	if err != nil {
		fmt.Printf("CheckJobNameUnique: 检查任务名称唯一性失败: %v\n", err)
		return false, err
	}

	isUnique := count == 0
	fmt.Printf("CheckJobNameUnique: 任务名称唯一性检查, JobName=%s, JobGroup=%s, IsUnique=%t\n", jobName, jobGroup, isUnique)
	return isUnique, nil
}
