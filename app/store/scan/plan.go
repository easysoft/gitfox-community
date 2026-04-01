package scan

import (
	"context"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
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
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted = ?", id, false).First(&plan).Error; err != nil {
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
	tx := s.db.WithContext(ctx).Where("id = ? AND deleted = ?", id, false)
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
	err := s.db.WithContext(ctx).Where("repoID = ? AND name = ? AND deleted = ?", repoID, name, false).First(&plan).Error
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
		if err := tx.Model(&ScanPlan{}).Where("id = ?", id).Update("deleted", true).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		if err := tx.Model(&ScanPlanSolution{}).Where("planID = ?", id).Update("deleted", true).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		return nil
	})
}

// List 查询指定仓库的扫描计划列表
func (s *ScanPlanStore) List(ctx context.Context, repoID uint, filter scanModel.ScanPlanReqFilter) (*database.PageResult[*scanModel.ScanPlan], error) {
	tx := s.db.WithContext(ctx).Model(&ScanPlan{}).Where("ops_scan_plans.deleted = ?", false)

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
		// 按最新任务结果过滤：子查询取每个计划最新任务，匹配 result
		tx = tx.Where("ops_scan_plans.id IN (?)",
			s.db.Model(&ScanTask{}).Select("planID").
				Where("id IN (?)",
					s.db.Model(&ScanTask{}).Select("MAX(id)").
						Where("deleted = ?", false).Group("planID"),
				).Where("result = ?", filter.LatestTaskResult),
		)
	} else {
		tx = tx.Where("ops_scan_plans.id IN (?)",
			s.db.Model(&ScanTask{}).Select("planID").
				Where("id IN (?)",
					s.db.Model(&ScanTask{}).Select("MAX(id)").
						Where("deleted = ?", false).Group("planID"),
				),
		)
	}
	if filter.LatestTaskStatus != "" {
		latestTaskSubQ := s.db.Model(&ScanTask{}).Select("MAX(id)").
			Where("deleted = ?", false).Group("planID")
		statusQuery := s.db.Table("ops_scan_tasks t").Select("t.planID").
			Joins("INNER JOIN ops_pipeline_executions e ON t.executionID = e.id").
			Where("t.id IN (?)", latestTaskSubQ)

		switch constant.ScanTaskStatus(filter.LatestTaskStatus) {
		case constant.ScanTaskStatusFailed:
			statusQuery = statusQuery.Where("e.status IN ?", []string{
				string(constant.CIStatusFailure),
				string(constant.CIStatusKilled),
				string(constant.CIStatusError),
			})
		case constant.ScanTaskStatusInProgress:
			statusQuery = statusQuery.Where("e.status NOT IN ?", []string{
				string(constant.CIStatusSuccess),
				string(constant.CIStatusFailure),
				string(constant.CIStatusKilled),
				string(constant.CIStatusError),
			})
		default:
			statusQuery = statusQuery.Where("e.status = ?", filter.LatestTaskStatus)
		}
		tx = tx.Where("ops_scan_plans.id IN (?)", statusQuery)
	}

	var plans []*ScanPlan
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

	// 批量查询关联的 solution IDs
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

	latestTaskMap, err := loadLatestTasksByPlanIDs(ctx, s.db, planIDs)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	modelData := make([]*scanModel.ScanPlan, 0, len(pageResult.Data))
	for _, item := range pageResult.Data {
		m := mapScanPlanToModel(*item)
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
		if task, ok := latestTaskMap[item.ID]; ok {
			applyLatestTaskToPlan(m, task)
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

func normalizeScanPlanSortField(sort string) string {
	switch strings.TrimSpace(sort) {
	case "", "id":
		return "id"
	case "name":
		return "name"
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

// GetSolutionIDList 获取计划关联的方案ID列表
func (s *ScanPlanStore) GetSolutionIDList(ctx context.Context, planID uint) ([]int64, error) {
	var ids []int64
	err := s.db.WithContext(ctx).Model(&ScanPlanSolution{}).
		Where("planID = ? AND deleted = ?", planID, false).
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
			Where("id IN ? AND deleted = ?", solutionIDs, false).
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
			} else if existing.Deleted {
				if err := tx.Model(&ScanPlanSolution{}).
					Where("planID = ? AND solutionID = ?", planID, solutionID).
					Update("deleted", false).Error; err != nil {
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
		Update("deleted", true).Error; err != nil {
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

	tx := s.db.WithContext(ctx).Where("deleted = ?", false)
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
		Deleted:     data.Deleted,
	}
}
