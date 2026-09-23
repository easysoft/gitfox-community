// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package jobs

import (
	"context"
	"time"

	jobModel "github.com/easysoft/gitfox-community/app/model/jobs"
	"github.com/easysoft/gitfox-community/pkg/database"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

type Store struct {
	db *gorm.DB
}

// NewScheduleJob 创建一个新的 ScheduleJob 实例
func NewScheduleJob(db *gorm.DB) *Store {
	return &Store{
		db: db,
	}
}

// GetByName fetches a job by its name (unique identifier).
func (s *Store) GetByName(ctx context.Context, name string) (*jobModel.ScheduleJob, error) {
	var job ScheduleJob
	if err := s.db.WithContext(ctx).
		Where("name = ?", name).
		First(&job).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapJobToModel(job), nil
}

// ListByGroupID fetches all jobs for a group id.
func (s *Store) ListByGroupID(ctx context.Context, groupID string) ([]*jobModel.ScheduleJob, error) {
	var jobs []ScheduleJob
	if err := s.db.WithContext(ctx).
		Where("groupID = ?", groupID).
		Find(&jobs).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	result := make([]*jobModel.ScheduleJob, len(jobs))
	for i, job := range jobs {
		result[i] = mapJobToModel(job)
	}
	return result, nil
}

// DeleteByGroupID deletes all jobs for a group id (physical delete).
func (s *Store) DeleteByGroupID(ctx context.Context, groupID string) (int64, error) {
	result := s.db.WithContext(ctx).
		Where("groupID = ?", groupID).
		Delete(&ScheduleJob{})

	if result.Error != nil {
		return 0, database.ProcessSQLErrorf(ctx, result.Error)
	}
	return result.RowsAffected, nil
}

// Create is used to create a new job.
func (s *Store) Create(ctx context.Context, job *jobModel.ScheduleJob) error {
	storeJob := mapJobToStore(job)
	if err := s.db.WithContext(ctx).Create(storeJob).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	job.ID = uint(storeJob.ID)
	return nil
}

// Upsert will insert the job in the database if the job didn't already exist,
// or it will update the existing one but only if its definition has changed.
//
// 采用"先查询、再插入或更新"的两段式实现，避免依赖 clause.OnConflict：
// gorm.io/driver/gaussdb 将 OnConflict 生成为 MySQL 的 ON DUPLICATE KEY UPDATE，
// 在 GaussDB 不同兼容格式（PG/B）下支持情况不一致，容易报错。改用标准
// SELECT + INSERT/UPDATE 可同时兼容两种格式。
func (s *Store) Upsert(ctx context.Context, job *jobModel.ScheduleJob) error {
	storeJob := mapJobToStore(job)
	storeJob.EditedDate = datetime.NowLocalTime()

	// 改为事务内“先查后写”，并兼容并发插入（创建失败时重查转更新）。
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing ScheduleJob
		err := tx.Where("name = ?", storeJob.Name).First(&existing).Error
		switch {
		case err == gorm.ErrRecordNotFound:
			if createErr := tx.Create(storeJob).Error; createErr != nil {
				// 并发下另一事务已插入同名记录：重查命中则转更新，否则返回原错误
				if retryErr := tx.Where("name = ?", storeJob.Name).First(&existing).Error; retryErr != nil {
					return createErr
				}
			} else {
				job.ID = uint(storeJob.ID)
				return nil
			}
		case err != nil:
			return err
		}

		if err := tx.Model(&ScheduleJob{}).
			Where("name = ?", storeJob.Name).
			Select(upsertUpdateColumns).
			Updates(storeJob).Error; err != nil {
			return err
		}
		job.ID = uint(existing.ID)
		return nil
	})
	if err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// upsertUpdateColumns 与旧 ON DUPLICATE KEY UPDATE / MERGE 的更新列保持一致。
var upsertUpdateColumns = []string{
	"editedDate", "type", "priority", "data", "result",
	"maxDurationSeconds", "maxRetries", "state",
	"scheduled", "isRecurring", "recurringCron",
}

// UpdateDefinition is used to update a job definition.
func (s *Store) UpdateDefinition(ctx context.Context, job *jobModel.ScheduleJob) error {
	updates := map[string]any{
		"editedDate":         datetime.NowLocalTime(),
		"type":               job.Type,
		"priority":           job.Priority,
		"data":               database.NoNull(job.Data),
		"result":             database.NoNull(job.Result),
		"maxDurationSeconds": job.MaxDurationSeconds,
		"maxRetries":         job.MaxRetries,
		"state":              job.State,
		"scheduled":          job.Scheduled,
		"isRecurring":        job.IsRecurring,
		"recurringCron":      job.RecurringCron,
		"groupID":            job.GroupID,
	}

	result := s.db.WithContext(ctx).
		Model(&ScheduleJob{}).
		Where("name = ?", job.Name).
		Updates(updates)

	if result.Error != nil {
		return database.ProcessSQLErrorf(ctx, result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.ErrResourceNotFound
	}

	return nil
}

// UpdateExecution is used to update a job before and after execution.
func (s *Store) UpdateExecution(ctx context.Context, job *jobModel.ScheduleJob) error {
	updates := map[string]any{
		"editedDate":          datetime.NowLocalTime(),
		"result":              job.Result,
		"state":               job.State,
		"scheduled":           job.Scheduled,
		"totalExecutions":     job.TotalExecutions,
		"runBy":               job.RunBy,
		"runDeadline":         job.RunDeadline,
		"lastExecuted":        job.LastExecuted,
		"consecutiveFailures": job.ConsecutiveFailures,
		"lastFailureError":    database.NoNull(job.LastFailureError),
	}

	result := s.db.WithContext(ctx).
		Model(&ScheduleJob{}).
		Where("name = ?", job.Name).
		Updates(updates)

	if result.Error != nil {
		return database.ProcessSQLErrorf(ctx, result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.ErrResourceNotFound
	}

	return nil
}

// UpdateProgress is used to update a job progress data.
// Only updates if the job is in running state.
func (s *Store) UpdateProgress(ctx context.Context, job *jobModel.ScheduleJob) error {
	updates := map[string]any{
		"editedDate":  datetime.NowLocalTime(),
		"result":      job.Result,
		"runProgress": job.RunProgress,
	}

	result := s.db.WithContext(ctx).
		Model(&ScheduleJob{}).
		Where("name = ? AND state = ?", job.Name, jobModel.StateRunning).
		Updates(updates)

	if result.Error != nil {
		return database.ProcessSQLErrorf(ctx, result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.ErrResourceNotFound
	}

	return nil
}

// CountRunning returns number of jobs that are currently being run.
func (s *Store) CountRunning(ctx context.Context) (int, error) {
	var count int64
	if err := s.db.WithContext(ctx).
		Model(&ScheduleJob{}).
		Where("state = ?", jobModel.StateRunning).
		Count(&count).Error; err != nil {
		return 0, database.ProcessSQLErrorf(ctx, err)
	}
	return int(count), nil
}

// ListReady returns a list of jobs that are ready for execution:
// The jobs with state="scheduled" and scheduled time in the past.
func (s *Store) ListReady(ctx context.Context, now time.Time, limit int) ([]*jobModel.ScheduleJob, error) {
	var jobs []ScheduleJob

	nowLocal := toLocalTime(now)

	if err := s.db.WithContext(ctx).
		Where("state = ? AND scheduled <= ?", jobModel.StateScheduled, nowLocal).
		Order("priority DESC, scheduled ASC, name ASC").
		Limit(limit).
		Find(&jobs).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	result := make([]*jobModel.ScheduleJob, len(jobs))
	for i, job := range jobs {
		result[i] = mapJobToModel(job)
	}
	return result, nil
}

// ListDeadlineExceeded returns a list of jobs that have exceeded their execution deadline.
func (s *Store) ListDeadlineExceeded(ctx context.Context, now time.Time) ([]*jobModel.ScheduleJob, error) {
	var jobs []ScheduleJob

	nowLocal := toLocalTime(now)

	if err := s.db.WithContext(ctx).
		Where("state = ? AND runDeadline < ?", jobModel.StateRunning, nowLocal).
		Order("runDeadline ASC").
		Find(&jobs).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	result := make([]*jobModel.ScheduleJob, len(jobs))
	for i, job := range jobs {
		result[i] = mapJobToModel(job)
	}
	return result, nil
}

// NextScheduledTime returns a scheduled time of the next ready job or zero time if no such job exists.
func (s *Store) NextScheduledTime(ctx context.Context, now time.Time) (time.Time, error) {
	var job ScheduleJob

	nowLocal := toLocalTime(now)

	if err := s.db.WithContext(ctx).
		Select("scheduled").
		Where("state = ? AND scheduled > ?", jobModel.StateScheduled, nowLocal).
		Order("scheduled ASC").
		Limit(1).
		First(&job).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return time.Time{}, nil
		}
		return time.Time{}, database.ProcessSQLErrorf(ctx, err)
	}

	if job.Scheduled == nil {
		return time.Time{}, nil
	}
	return time.Time(*job.Scheduled), nil
}

// DeleteOld removes non-recurring jobs that have finished execution or have failed or been canceled.
func (s *Store) DeleteOld(ctx context.Context, olderThan time.Time) (int64, error) {
	olderThanLocal := toLocalTime(olderThan)

	result := s.db.WithContext(ctx).
		Where("isRecurring = ? AND state IN (?, ?, ?) AND lastExecuted < ?",
			0, jobModel.StateFinished, jobModel.StateFailed, jobModel.StateCanceled, olderThanLocal).
		Delete(&ScheduleJob{})

	if result.Error != nil {
		return 0, database.ProcessSQLErrorf(ctx, result.Error)
	}
	return result.RowsAffected, nil
}

// DeleteByName deletes a job by its name (physical delete).
func (s *Store) DeleteByName(ctx context.Context, name string) error {
	result := s.db.WithContext(ctx).
		Where("name = ?", name).
		Delete(&ScheduleJob{})

	if result.Error != nil {
		return database.ProcessSQLErrorf(ctx, result.Error)
	}
	return nil
}

// mapJobToModel 将 Store 层结构转换为 Model 层结构
func mapJobToModel(j ScheduleJob) *jobModel.ScheduleJob {
	return &jobModel.ScheduleJob{
		ID:                  uint(j.ID),
		Name:                j.Name,
		Type:                j.Type,
		Priority:            jobModel.Priority(j.Priority),
		Data:                j.Data.String,
		Result:              j.Result.String,
		MaxDurationSeconds:  j.MaxDurationSeconds,
		MaxRetries:          j.MaxRetries,
		State:               jobModel.State(j.State),
		Scheduled:           j.Scheduled,
		TotalExecutions:     j.TotalExecutions,
		RunBy:               j.RunBy,
		RunDeadline:         j.RunDeadline,
		RunProgress:         j.RunProgress,
		LastExecuted:        j.LastExecuted,
		IsRecurring:         j.IsRecurring != 0,
		RecurringCron:       j.RecurringCron,
		ConsecutiveFailures: j.ConsecutiveFailures,
		LastFailureError:    j.LastFailureError.String,
		GroupID:             j.GroupID,
		CreatedBy:           j.CreatedBy,
		CreatedDate:         j.CreatedDate,
		EditedBy:            j.EditedBy,
		EditedDate:          j.EditedDate,
	}
}

// mapJobToStore 将 Model 层结构转换为 Store 层结构
func mapJobToStore(j *jobModel.ScheduleJob) *ScheduleJob {
	return &ScheduleJob{
		ID:                  int64(j.ID),
		Name:                j.Name,
		Type:                j.Type,
		Priority:            int(j.Priority),
		Data:                database.NoNull(j.Data),
		Result:              database.NoNull(j.Result),
		MaxDurationSeconds:  j.MaxDurationSeconds,
		MaxRetries:          j.MaxRetries,
		State:               string(j.State),
		Scheduled:           j.Scheduled,
		TotalExecutions:     j.TotalExecutions,
		RunBy:               j.RunBy,
		RunDeadline:         j.RunDeadline,
		RunProgress:         j.RunProgress,
		LastExecuted:        j.LastExecuted,
		IsRecurring:         database.BoolToUint8(j.IsRecurring),
		RecurringCron:       j.RecurringCron,
		ConsecutiveFailures: j.ConsecutiveFailures,
		LastFailureError:    database.NoNull(j.LastFailureError),
		GroupID:             j.GroupID,
		CreatedBy:           j.CreatedBy,
		CreatedDate:         j.CreatedDate,
		EditedBy:            j.EditedBy,
		EditedDate:          j.EditedDate,
	}
}

// toLocalTime 将 time.Time 转换为 *datetime.Local
func toLocalTime(t time.Time) *datetime.Local {
	localTime := datetime.Local(t.In(datetime.CST))
	return &localTime
}
