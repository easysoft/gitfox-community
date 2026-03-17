package scan

import (
	"context"
	"strings"

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
	return s.enrichTrigger(ctx, trigger), nil
}

func (s *ScanTriggerStore) GetByRepoPlanAndID(ctx context.Context, repoID uint, planID uint, id uint) (*scanModel.ScanTrigger, error) {
	var trigger ScanPlanTrigger
	if err := s.db.WithContext(ctx).Where("repoID = ? AND planID = ? AND id = ? AND deleted = ?", repoID, planID, id, false).First(&trigger).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return s.enrichTrigger(ctx, trigger), nil
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

	// 批量查询关联的 conditions（完整对象 + IDs）
	triggerIDs := make([]uint64, 0, len(pageResult.Data))
	for _, t := range pageResult.Data {
		triggerIDs = append(triggerIDs, t.ID)
	}

	conditionsObjMap := make(map[uint64][]*scanModel.ScanCondition)
	conditionsIDMap := make(map[uint64][]int64)
	if len(triggerIDs) > 0 {
		var conditions []ScanPlanCondition
		if err := s.db.WithContext(ctx).
			Where("triggerID IN ? AND deleted = ?", triggerIDs, false).
			Find(&conditions).Error; err == nil {
			for _, c := range conditions {
				conditionsIDMap[c.TriggerID] = append(conditionsIDMap[c.TriggerID], int64(c.ID))
				conditionsObjMap[c.TriggerID] = append(conditionsObjMap[c.TriggerID], mapScanConditionToModel(c))
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

	// 批量查询关联的 solutions（完整对象 + IDs）
	solutionsIDMap := make(map[uint64][]int64)
	solutionsObjMap := make(map[uint64][]*scanModel.Solution)
	if len(triggerIDs) > 0 {
		// 先查关联关系
		var triggerSolutions []ScanTriggerSolution
		if err := s.db.WithContext(ctx).
			Where("triggerID IN ?", triggerIDs).
			Find(&triggerSolutions).Error; err == nil && len(triggerSolutions) > 0 {
			// 收集所有 solutionID 并按 triggerID 分组
			allSolutionIDs := make([]uint64, 0)
			triggerToSolutionIDs := make(map[uint64][]uint64)
			for _, ts := range triggerSolutions {
				triggerToSolutionIDs[ts.TriggerID] = append(triggerToSolutionIDs[ts.TriggerID], ts.SolutionID)
				solutionsIDMap[ts.TriggerID] = append(solutionsIDMap[ts.TriggerID], int64(ts.SolutionID))
				allSolutionIDs = append(allSolutionIDs, ts.SolutionID)
			}
			// 批量查 solution 完整对象
			var solutions []SolutionRules
			if err := s.db.WithContext(ctx).
				Where("id IN ? AND deleted = ?", allSolutionIDs, false).
				Find(&solutions).Error; err == nil {
				solutionMap := make(map[uint64]*scanModel.Solution, len(solutions))
				for _, sol := range solutions {
					solutionMap[sol.ID] = mapSolutionToModel(sol)
				}
				for tid, sids := range triggerToSolutionIDs {
					for _, sid := range sids {
						if obj, ok := solutionMap[sid]; ok {
							solutionsObjMap[tid] = append(solutionsObjMap[tid], obj)
						}
					}
				}
			}
		}
	}

	modelData := make([]*scanModel.ScanTrigger, 0, len(pageResult.Data))
	for _, item := range pageResult.Data {
		m := mapScanTriggerToModel(*item)
		m.ConditionIDs = conditionsIDMap[item.ID]
		m.Conditions = conditionsObjMap[item.ID]
		m.ConditionCount = len(conditionsIDMap[item.ID])
		m.SolutionIDs = solutionsIDMap[item.ID]
		m.Solutions = solutionsObjMap[item.ID]
		m.SolutionCount = len(solutionsIDMap[item.ID])
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
		ScanType:    constant.ScanType(data.ScanType),
		Cron:        data.Cron,
		CronBranch:  data.CronBranch,
		Keywords:    data.Keywords,
		Actions:     parseActions(data.Actions),
		Disabled:    data.Disabled,
		CreatedBy:   data.CreatedBy,
		CreatedDate: data.CreatedDate,
		EditedBy:    data.EditedBy,
		EditedDate:  data.EditedDate,
		Deleted:     data.Deleted,
	}
}

// parseActions 将逗号分隔的 actions 字符串转换为 []TriggerAction。
func parseActions(s string) []constant.TriggerAction {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	actions := make([]constant.TriggerAction, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			actions = append(actions, constant.TriggerAction(p))
		}
	}
	return actions
}

// JoinActions 将 []TriggerAction 转换为逗号分隔的字符串。
func JoinActions(actions []constant.TriggerAction) string {
	if len(actions) == 0 {
		return ""
	}
	parts := make([]string, len(actions))
	for i, a := range actions {
		parts[i] = string(a)
	}
	return strings.Join(parts, ",")
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

// enrichTrigger 为单个 trigger 填充关联的 conditions 和 solutions 完整对象。
func (s *ScanTriggerStore) enrichTrigger(ctx context.Context, data ScanPlanTrigger) *scanModel.ScanTrigger {
	m := mapScanTriggerToModel(data)

	// 查询关联 conditions
	var conditions []ScanPlanCondition
	if err := s.db.WithContext(ctx).
		Where("triggerID = ? AND deleted = ?", data.ID, false).
		Find(&conditions).Error; err == nil {
		for _, c := range conditions {
			m.ConditionIDs = append(m.ConditionIDs, int64(c.ID))
			m.Conditions = append(m.Conditions, mapScanConditionToModel(c))
		}
		m.ConditionCount = len(conditions)
	}

	// 查询关联 solutions
	var triggerSolutions []ScanTriggerSolution
	if err := s.db.WithContext(ctx).
		Where("triggerID = ?", data.ID).
		Find(&triggerSolutions).Error; err == nil && len(triggerSolutions) > 0 {
		solutionIDs := make([]uint64, 0, len(triggerSolutions))
		for _, ts := range triggerSolutions {
			solutionIDs = append(solutionIDs, ts.SolutionID)
			m.SolutionIDs = append(m.SolutionIDs, int64(ts.SolutionID))
		}
		var solutions []SolutionRules
		if err := s.db.WithContext(ctx).
			Where("id IN ? AND deleted = ?", solutionIDs, false).
			Find(&solutions).Error; err == nil {
			for _, sol := range solutions {
				m.Solutions = append(m.Solutions, mapSolutionToModel(sol))
			}
		}
		m.SolutionCount = len(m.SolutionIDs)
	}

	return m
}
