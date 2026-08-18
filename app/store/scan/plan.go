package scan

import (
	"context"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

var _ IScanPlan = (*ScanPlanStore)(nil)

type ScanPlanStore struct {
	db *gorm.DB
}

func NewScanPlanStore(db *gorm.DB) *ScanPlanStore {
	return &ScanPlanStore{db: db}
}

func (s *ScanPlanStore) Create(ctx context.Context, plan *ScanPlan) (*scanModel.ScanPlan, error) {
	if err := s.db.WithContext(ctx).Create(plan).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapScanPlanToModel(*plan), nil
}

func (s *ScanPlanStore) GetByID(ctx context.Context, id uint) (*scanModel.ScanPlan, error) {
	var plan ScanPlan
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted = ?", id, 0).First(&plan).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	m := mapScanPlanToModel(plan)
	repoMetaMap, err := loadRepoMetaByRepoIDs(ctx, s.db, []uint64{plan.RepoID})
	if err == nil {
		if meta, ok := repoMetaMap[plan.RepoID]; ok {
			m.RepoName = meta.Name
			m.SpaceID = meta.SpaceID
			spaceNameMap, e := loadSpaceNamesBySpaceIDs(ctx, s.db, []uint64{uint64(meta.SpaceID)})
			if e == nil {
				m.SpaceName = spaceNameMap[uint64(meta.SpaceID)]
			}
		}
	}
	latestTaskMap, err := loadLatestTasksByPlanIDs(ctx, s.db, []uint64{plan.ID})
	if err == nil {
		applyLatestTaskToPlan(m, latestTaskMap[plan.ID])
	}
	return m, nil
}

func (s *ScanPlanStore) GetByRepoIDAndID(ctx context.Context, repoID uint, id uint) (*scanModel.ScanPlan, error) {
	var plan ScanPlan
	tx := s.db.WithContext(ctx).Where("id = ? AND deleted = ?", id, 0)
	if repoID != 0 {
		tx = tx.Where("repoID = ?", repoID)
	}
	if err := tx.First(&plan).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	m := mapScanPlanToModel(plan)
	repoMetaMap, err := loadRepoMetaByRepoIDs(ctx, s.db, []uint64{plan.RepoID})
	if err == nil {
		if meta, ok := repoMetaMap[plan.RepoID]; ok {
			m.RepoName = meta.Name
			m.SpaceID = meta.SpaceID
			spaceNameMap, e := loadSpaceNamesBySpaceIDs(ctx, s.db, []uint64{uint64(meta.SpaceID)})
			if e == nil {
				m.SpaceName = spaceNameMap[uint64(meta.SpaceID)]
			}
		}
	}
	latestTaskMap, err := loadLatestTasksByPlanIDs(ctx, s.db, []uint64{plan.ID})
	if err == nil {
		applyLatestTaskToPlan(m, latestTaskMap[plan.ID])
	}
	return m, nil
}

func (s *ScanPlanStore) FindByName(ctx context.Context, repoID uint, name string) (*scanModel.ScanPlan, error) {
	var plan ScanPlan
	err := s.db.WithContext(ctx).Where("repoID = ? AND name = ? AND deleted = ?", repoID, name, 0).First(&plan).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapScanPlanToModel(plan), nil
}

func (s *ScanPlanStore) Update(ctx context.Context, plan *ScanPlan) error {
	if err := s.db.WithContext(ctx).Updates(plan).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// Delete 软删除扫描计划及其关联的方案绑定关系
func (s *ScanPlanStore) Delete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&ScanPlan{}).Where("id = ?", id).Update("deleted", 1).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		if err := tx.Model(&ScanPlanSolution{}).Where("planID = ?", id).Update("deleted", 1).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		return nil
	})
}

// scanPlanWithTask 扫描计划联查结果（含最新任务信息）
type scanPlanWithTask struct {
	ScanPlan
	LatestTaskID          *uint64                  `gorm:"column:latest_task_id"`
	LatestTaskCreated     *datetime.Local          `gorm:"column:latest_task_created"`
	LatestTaskResult      *constant.ScanTaskResult `gorm:"column:latest_task_result"`
	LatestExecutionStatus *constant.CIStatus       `gorm:"column:latest_execution_status"`
}

// List 查询指定仓库的扫描计划列表
// 通过 LEFT JOIN 子查询将最新任务信息内联到主查询中，
// 使所有排序字段（包括 latestTaskResult/latestTaskStatus/latestTaskCreated）都可以在 SQL 层排序，
// 避免内存排序导致分页数据遗漏。
func (s *ScanPlanStore) List(ctx context.Context, repoID uint, filter scanModel.ScanPlanReqFilter) (*database.PageResult[*scanModel.ScanPlan], error) {
	// 构建主查询：LEFT JOIN 最新任务子查询
	tx := s.db.WithContext(ctx).
		Table("ops_scan_plans").
		Select(`ops_scan_plans.*,
			latest_task.id AS latest_task_id,
			latest_task.createdDate AS latest_task_created,
			latest_task.result AS latest_task_result,
			e.status AS latest_execution_status`).
		Joins(`LEFT JOIN (
			SELECT t1.*
			FROM ops_scan_tasks t1
			INNER JOIN (
				SELECT planID, MAX(id) AS max_id
				FROM ops_scan_tasks
				WHERE deleted = ?
				GROUP BY planID
			) t2 ON t1.id = t2.max_id
		) latest_task ON ops_scan_plans.id = latest_task.planID`, 0).
		Joins("LEFT JOIN ops_pipeline_executions e ON latest_task.executionID = e.id").
		Where("ops_scan_plans.deleted = ?", 0)

	if filter.Name != "" {
		tx = tx.Where("ops_scan_plans.name LIKE ?", "%"+filter.Name+"%")
	}

	if repoID != 0 {
		tx = tx.Where("ops_scan_plans.repoID = ?", repoID)
	}

	if filter.SpaceID != 0 {
		tx = tx.Joins("JOIN ops_repo ON ops_repo.id = ops_scan_plans.repoID AND ops_repo.deleted = ?", 0).
			Where("ops_repo.spaceID = ?", filter.SpaceID)
	}

	if filter.LatestTaskResult != "" {
		tx = tx.Where("latest_task.result = ?", filter.LatestTaskResult)
	}

	if filter.LatestTaskStatus != "" {
		switch constant.ScanTaskStatus(filter.LatestTaskStatus) {
		case constant.ScanTaskStatusFailed:
			tx = tx.Where("e.status IN ?", []string{
				string(constant.CIStatusFailure),
				string(constant.CIStatusKilled),
				string(constant.CIStatusError),
			})
		case constant.ScanTaskStatusInProgress:
			tx = tx.Where("e.status NOT IN ?", []string{
				string(constant.CIStatusSuccess),
				string(constant.CIStatusFailure),
				string(constant.CIStatusKilled),
				string(constant.CIStatusError),
			})
		default:
			tx = tx.Where("e.status = ?", filter.LatestTaskStatus)
		}
	}

	var plans []*scanPlanWithTask
	pageResult, err := database.Paginate(ctx, tx, database.PageParam{
		Page:      filter.Page,
		PageSize:  filter.PageSize,
		Sort:      normalizeScanPlanSortField(filter.Sort),
		Order:     normalizeSortOrder(filter.Order, "desc"),
		NeedTotal: true,
	}, &plans)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// 批量查询关联数据
	planIDs := make([]uint64, 0, len(pageResult.Data))
	repoIDs := make([]uint64, 0, len(pageResult.Data))
	for _, p := range pageResult.Data {
		planIDs = append(planIDs, p.ID)
		repoIDs = append(repoIDs, p.RepoID)
	}

	repoMetaMap, err := loadRepoMetaByRepoIDs(ctx, s.db, repoIDs)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	spaceIDs := make([]uint64, 0, len(repoMetaMap))
	for _, meta := range repoMetaMap {
		spaceIDs = append(spaceIDs, uint64(meta.SpaceID))
	}
	spaceNameMap, err := loadSpaceNamesBySpaceIDs(ctx, s.db, spaceIDs)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	solutionsMap, err := loadSolutionIDsByPlanIDs(ctx, s.db, planIDs)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	triggerIDsMap, err := loadTriggerIDsByPlanIDs(ctx, s.db, planIDs)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	conditionIDsMap, err := loadPlanConditionsByPlanIDs(ctx, s.db, planIDs)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	modelData := make([]*scanModel.ScanPlan, 0, len(pageResult.Data))
	for _, item := range pageResult.Data {
		m := mapScanPlanToModel(item.ScanPlan)
		if meta, ok := repoMetaMap[item.RepoID]; ok {
			m.RepoName = meta.Name
			m.SpaceID = meta.SpaceID
			m.SpaceName = spaceNameMap[uint64(meta.SpaceID)]
		}
		m.SolutionIDs = solutionsMap[item.ID]
		m.SolutionCount = len(solutionsMap[item.ID])
		m.TriggerIDs = triggerIDsMap[item.ID]
		m.TriggerCount = len(triggerIDsMap[item.ID])
		m.ConditionIDs = conditionIDsMap[item.ID]
		m.ConditionCount = len(conditionIDsMap[item.ID])
		// 从 JOIN 结果直接设置最新任务信息
		if item.LatestTaskID != nil {
			taskID := int64(*item.LatestTaskID)
			m.LatestTaskID = &taskID
		}
		m.LatestTaskResult = item.LatestTaskResult
		m.LatestTaskCreated = item.LatestTaskCreated
		if item.LatestExecutionStatus != nil {
			status := convertExecutionTaskStatus(*item.LatestExecutionStatus)
			m.LatestTaskStatus = &status
		}
		modelData = append(modelData, m)
	}

	return &database.PageResult[*scanModel.ScanPlan]{
		Data:     modelData,
		Total:    pageResult.Total,
		Page:     pageResult.Page,
		PageSize: pageResult.PageSize,
	}, nil
}

// normalizeScanPlanSortField 将前端传入的排序字段映射为安全的数据库列名
func normalizeScanPlanSortField(sort string) string {
	switch strings.TrimSpace(sort) {
	case "", "id":
		return "ops_scan_plans.id"
	case "name":
		return "ops_scan_plans.name"
	case "scanType":
		return "ops_scan_plans.scanType"
	case "createdBy":
		return "ops_scan_plans.createdBy"
	case "createdDate":
		return "ops_scan_plans.createdDate"
	case "editedBy":
		return "ops_scan_plans.editedBy"
	case "editedDate":
		return "ops_scan_plans.editedDate"
	case "latestTaskResult":
		return "latest_task_result"
	case "latestTaskStatus":
		return "latest_execution_status"
	case "latestTaskCreated":
		return "latest_task_created"
	default:
		return "ops_scan_plans.id"
	}
}

// GetSolutionIDList 获取计划关联的方案ID列表
func (s *ScanPlanStore) GetSolutionIDList(ctx context.Context, planID uint) ([]int64, error) {
	var ids []int64
	err := s.db.WithContext(ctx).Model(&ScanPlanSolution{}).
		Where("planID = ? AND deleted = ?", planID, 0).
		Pluck("solutionID", &ids).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return ids, nil
}

// BindSolutions 为扫描计划绑定扫描方案
func (s *ScanPlanStore) BindSolutions(ctx context.Context, planID uint, solutionIDs []uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var validSolutionIDs []uint
		if err := tx.Model(&SolutionRules{}).
			Where("id IN ? AND deleted = ?", solutionIDs, 0).
			Pluck("id", &validSolutionIDs).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		if len(validSolutionIDs) != len(solutionIDs) {
			return database.ProcessSQLErrorf(ctx, gorm.ErrRecordNotFound)
		}

		for _, solutionID := range solutionIDs {
			var existing ScanPlanSolution
			err := tx.Where("planID = ? AND solutionID = ?", planID, solutionID).First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				newRecord := ScanPlanSolution{PlanID: uint64(planID), SolutionID: uint64(solutionID)}
				if err := tx.Create(&newRecord).Error; err != nil {
					return database.ProcessSQLErrorf(ctx, err)
				}
			} else if err != nil {
				return database.ProcessSQLErrorf(ctx, err)
			} else if existing.Deleted != 0 {
				if err := tx.Model(&ScanPlanSolution{}).
					Where("planID = ? AND solutionID = ?", planID, solutionID).
					Update("deleted", 0).Error; err != nil {
					return database.ProcessSQLErrorf(ctx, err)
				}
			}
		}
		return nil
	})
}

// UnbindSolutions 解除扫描计划与方案的绑定
func (s *ScanPlanStore) UnbindSolutions(ctx context.Context, planID uint, solutionIDs []uint) error {
	if err := s.db.WithContext(ctx).Model(&ScanPlanSolution{}).
		Where("planID = ? AND solutionID IN ?", planID, solutionIDs).
		Update("deleted", 1).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// ListUnbindSolutions 获取扫描计划未绑定的方案列表
func (s *ScanPlanStore) ListUnbindSolutions(ctx context.Context, planID uint) ([]*scanModel.Solution, error) {
	boundIDs, err := s.GetSolutionIDList(ctx, planID)
	if err != nil {
		return nil, err
	}

	tx := s.db.WithContext(ctx).Where("deleted = ?", 0)
	if len(boundIDs) > 0 {
		tx = tx.Where("id NOT IN ?", boundIDs)
	}

	var solutions []*SolutionRules
	if err := tx.Find(&solutions).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	result := make([]*scanModel.Solution, 0, len(solutions))
	for _, s := range solutions {
		result = append(result, mapSolutionToModel(*s))
	}
	return result, nil
}

func mapScanPlanToModel(data ScanPlan) *scanModel.ScanPlan {
	branches, err := scanModel.ParsePattern(data.Branches)
	if err != nil {
		branches = &scanModel.Pattern{}
	}
	files, err := scanModel.ParsePattern(data.Files)
	if err != nil {
		files = &scanModel.Pattern{}
	}

	return &scanModel.ScanPlan{
		ID:          int64(data.ID),
		RepoID:      int64(data.RepoID),
		Name:        data.Name,
		Desc:        data.Desc,
		ScanType:    data.ScanType,
		Branches:    branches,
		Files:       files,
		CreatedBy:   data.CreatedBy,
		CreatedDate: data.CreatedDate,
		EditedBy:    data.EditedBy,
		EditedDate:  data.EditedDate,
		Deleted:     data.Deleted != 0,
	}
}
