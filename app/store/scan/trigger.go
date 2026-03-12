package scan

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

var _ IScanTrigger = (*ScanTriggerStore)(nil)

type ScanTriggerStore struct {
	db *gorm.DB
}

func NewScanTriggerStore(db *gorm.DB) *ScanTriggerStore {
	return &ScanTriggerStore{db: db}
}

func (s *ScanTriggerStore) Create(ctx context.Context, trigger *ScanPlanTrigger) (*scanModel.ScanTrigger, error) {
	if err := s.db.WithContext(ctx).Create(trigger).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapScanTriggerToModel(*trigger), nil
}

func (s *ScanTriggerStore) GetByID(ctx context.Context, id uint) (*scanModel.ScanTrigger, error) {
	var trigger ScanPlanTrigger
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted = ?", id, false).First(&trigger).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapScanTriggerToModel(trigger), nil
}

func (s *ScanTriggerStore) GetByRepoPlanAndID(ctx context.Context, repoID uint, planID uint, id uint) (*scanModel.ScanTrigger, error) {
	var trigger ScanPlanTrigger
	if err := s.db.WithContext(ctx).Where("repoID = ? AND planID = ? AND id = ? AND deleted = ?", repoID, planID, id, false).First(&trigger).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapScanTriggerToModel(trigger), nil
}

func (s *ScanTriggerStore) Update(ctx context.Context, trigger *ScanPlanTrigger) error {
	if err := s.db.WithContext(ctx).Updates(trigger).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (s *ScanTriggerStore) Delete(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Model(&ScanPlanTrigger{}).
		Where("id = ?", id).Update("deleted", true).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (s *ScanTriggerStore) List(ctx context.Context, repoID uint, planID uint, filter scanModel.ScanTriggerReqFilter) (*database.PageResult[*scanModel.ScanTrigger], error) {
	tx := s.db.WithContext(ctx).Model(&ScanPlanTrigger{}).
		Where("repoID = ? AND planID = ? AND deleted = ?", repoID, planID, false)

	var triggers []*ScanPlanTrigger
	pageResult, err := database.Paginate(ctx, tx, database.PageParam{
		Page:      filter.Page,
		PageSize:  filter.PageSize,
		Sort:      normalizeScanTriggerSortField(filter.Sort),
		Order:     normalizeSortOrder(filter.Order, "desc"),
		NeedTotal: true,
	}, &triggers)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// 批量查询关联的 condition IDs
	triggerIDs := make([]uint64, 0, len(pageResult.Data))
	for _, t := range pageResult.Data {
		triggerIDs = append(triggerIDs, t.ID)
	}

	conditionsMap := make(map[uint64][]int64)
	if len(triggerIDs) > 0 {
		var conditions []ScanPlanCondition
		if err := s.db.WithContext(ctx).
			Where("triggerID IN ? AND deleted = ?", triggerIDs, false).
			Find(&conditions).Error; err == nil {
			for _, c := range conditions {
				conditionsMap[c.TriggerID] = append(conditionsMap[c.TriggerID], int64(c.ID))
			}
		}
	}

	// 批量查询每个触发器的最新任务（关联 execution 获取真实状态）
	type triggerTaskJoin struct {
		ScanTask
		ExecutionStatus constant.CIStatus `gorm:"column:executionStatus"`
	}
	latestTaskMap := make(map[uint64]*triggerTaskJoin)
	if len(triggerIDs) > 0 {
		subQuery := s.db.Model(&ScanTask{}).
			Select("MAX(id) as id").
			Where("triggerID IN ? AND deleted = ?", triggerIDs, false).
			Group("triggerID")
		var tasks []triggerTaskJoin
		if err := s.db.WithContext(ctx).
			Table("ops_scan_tasks").
			Select("ops_scan_tasks.*, e.status as executionStatus").
			Joins("INNER JOIN ops_pipeline_executions e ON ops_scan_tasks.executionID = e.id").
			Where("ops_scan_tasks.id IN (?)", subQuery).
			Find(&tasks).Error; err == nil {
			for i := range tasks {
				if tasks[i].TriggerID != nil {
					latestTaskMap[*tasks[i].TriggerID] = &tasks[i]
				}
			}
		}
	}

	modelData := make([]*scanModel.ScanTrigger, 0, len(pageResult.Data))
	for _, item := range pageResult.Data {
		m := mapScanTriggerToModel(*item)
		m.ConditionIDs = conditionsMap[item.ID]
		m.ConditionCount = len(conditionsMap[item.ID])
		if task, ok := latestTaskMap[item.ID]; ok {
			taskID := int64(task.ID)
			m.LatestTaskID = &taskID
			if task.CreatedDate != nil {
				m.LatestTaskCreated = task.CreatedDate
			}
			m.LatestTaskResult = task.Result
			status := convertExecutionTaskStatus(task.ExecutionStatus)
			m.LatestTaskStatus = &status
		}
		modelData = append(modelData, m)
	}

	return &database.PageResult[*scanModel.ScanTrigger]{
		Data:     modelData,
		Total:    pageResult.Total,
		Page:     pageResult.Page,
		PageSize: pageResult.PageSize,
	}, nil
}

func (s *ScanTriggerStore) Count(ctx context.Context, repoID uint, planID uint) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&ScanPlanTrigger{}).
		Where("repoID = ? AND planID = ? AND deleted = ?", repoID, planID, false).
		Count(&count).Error; err != nil {
		return 0, database.ProcessSQLErrorf(ctx, err)
	}
	return count, nil
}

func mapScanTriggerToModel(data ScanPlanTrigger) *scanModel.ScanTrigger {
	return &scanModel.ScanTrigger{
		ID:          int64(data.ID),
		PlanID:      int64(data.PlanID),
		RepoID:      int64(data.RepoID),
		Name:        data.Name,
		Desc:        data.Desc,
		Type:        scanModel.ScanTriggerType(data.Type),
		Cron:        data.Cron,
		CronBranch:  data.CronBranch,
		Keywords:    data.Keywords,
		Actions:     data.Actions,
		Disabled:    data.Disabled,
		CreatedBy:   data.CreatedBy,
		CreatedDate: data.CreatedDate,
		EditedBy:    data.EditedBy,
		EditedDate:  data.EditedDate,
		Deleted:     data.Deleted,
	}
}

func normalizeScanTriggerSortField(sort string) string {
	switch sort {
	case "", "id":
		return "id"
	case "name":
		return "name"
	case "type":
		return "type"
	case "createdBy":
		return "createdBy"
	case "createdDate":
		return "createdDate"
	case "editedBy":
		return "editedBy"
	case "editedDate":
		return "editedDate"
	default:
		return "id"
	}
}
