package jobs

import (
	"context"
	"time"

	jobModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/jobs"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
func (s *Store) Upsert(ctx context.Context, job *jobModel.ScheduleJob) error {
	storeJob := mapJobToStore(job)
	storeJob.EditedDate = datetime.NowLocalTime()

	// 使用 ON DUPLICATE KEY UPDATE 进行 upsert
	// 只有当定义字段有变化时才更新
	if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"editedDate", "type", "priority", "data", "result",
			"maxDurationSeconds", "maxRetries", "state",
			"scheduled", "isRecurring", "recurringCron",
		}),
	}).Create(storeJob).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	job.ID = uint(storeJob.ID)
	return nil
}

// UpdateDefinition is used to update a job definition.
func (s *Store) UpdateDefinition(ctx context.Context, job *jobModel.ScheduleJob) error {
	updates := map[string]any{
		"editedDate":         datetime.NowLocalTime(),
		"type":               job.Type,
		"priority":           job.Priority,
		"data":               job.Data,
		"result":             job.Result,
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
		"lastFailureError":    job.LastFailureError,
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
			false, jobModel.StateFinished, jobModel.StateFailed, jobModel.StateCanceled, olderThanLocal).
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
		Data:                j.Data,
		Result:              j.Result,
		MaxDurationSeconds:  j.MaxDurationSeconds,
		MaxRetries:          j.MaxRetries,
		State:               jobModel.State(j.State),
		Scheduled:           j.Scheduled,
		TotalExecutions:     j.TotalExecutions,
		RunBy:               j.RunBy,
		RunDeadline:         j.RunDeadline,
		RunProgress:         j.RunProgress,
		LastExecuted:        j.LastExecuted,
		IsRecurring:         j.IsRecurring,
		RecurringCron:       j.RecurringCron,
		ConsecutiveFailures: j.ConsecutiveFailures,
		LastFailureError:    j.LastFailureError,
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
		Data:                j.Data,
		Result:              j.Result,
		MaxDurationSeconds:  j.MaxDurationSeconds,
		MaxRetries:          j.MaxRetries,
		State:               string(j.State),
		Scheduled:           j.Scheduled,
		TotalExecutions:     j.TotalExecutions,
		RunBy:               j.RunBy,
		RunDeadline:         j.RunDeadline,
		RunProgress:         j.RunProgress,
		LastExecuted:        j.LastExecuted,
		IsRecurring:         j.IsRecurring,
		RecurringCron:       j.RecurringCron,
		ConsecutiveFailures: j.ConsecutiveFailures,
		LastFailureError:    j.LastFailureError,
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
