package jobs

import (
	"context"
	"time"

	jobModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/jobs"
)

//go:generate mockgen -source=$GOFILE -destination=mock/scheduleJob_store.go -package=mock

type Interface interface {
	// GetByName fetches a job by its name.
	// 通过名称获取工作。
	GetByName(ctx context.Context, name string) (*jobModel.ScheduleJob, error)

	// ListByGroupID fetches all jobs for a group id
	// 通过组ID获取所有作业。
	ListByGroupID(ctx context.Context, groupID string) ([]*jobModel.ScheduleJob, error)

	// DeleteByGroupID deletes all jobs for a group id
	// 通过组ID删除所有作业。
	DeleteByGroupID(ctx context.Context, groupID string) (int64, error)

	// Create is used to create a new job.
	// 创建新作业。
	Create(ctx context.Context, job *jobModel.ScheduleJob) error

	// Upsert will insert the job in the database if the job didn't already exist,
	// or it will update the existing one but only if its definition has changed.
	// 如果该作业尚不存在，Upsert将在数据库中插入该作业，
	// 或者它会更新现有的，但前提是其定义已更改。
	Upsert(ctx context.Context, job *jobModel.ScheduleJob) error

	// UpdateDefinition is used to update a job definition.
	// 更新作业定义。
	UpdateDefinition(ctx context.Context, job *jobModel.ScheduleJob) error

	// UpdateExecution is used to update a job before and after execution.
	// 更新作业执行。
	UpdateExecution(ctx context.Context, job *jobModel.ScheduleJob) error

	// UpdateProgress is used to update a job progress data.
	// 更新作业进度数据。
	UpdateProgress(ctx context.Context, job *jobModel.ScheduleJob) error

	// CountRunning returns number of jobs that are currently being run.
	// 返回当前正在运行的作业数量。
	CountRunning(ctx context.Context) (int, error)

	// ListReady returns a list of jobs that are ready for execution.
	// 返回准备执行的作业列表。
	ListReady(ctx context.Context, now time.Time, limit int) ([]*jobModel.ScheduleJob, error)

	// ListDeadlineExceeded returns a list of jobs that have exceeded their execution deadline.
	// 返回已超过执行截止时间的作业列表。
	ListDeadlineExceeded(ctx context.Context, now time.Time) ([]*jobModel.ScheduleJob, error)

	// NextScheduledTime returns a scheduled time of the next ready job.
	// 返回下一个准备执行的作业的计划时间。
	NextScheduledTime(ctx context.Context, now time.Time) (time.Time, error)

	// DeleteOld removes non-recurring jobs that have finished execution or have failed.
	// 删除已完成执行或失败的非重复作业。
	DeleteOld(ctx context.Context, olderThan time.Time) (int64, error)

	// DeleteByName deletes a job by its name.
	// 通过名称删除作业。
	DeleteByName(ctx context.Context, name string) error
}
