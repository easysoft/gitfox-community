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
	repoIDs := make([]uint64, 0, len(pageResult.Data))
	for _, t := range pageResult.Data {
		triggerIDs = append(triggerIDs, t.ID)
		repoIDs = append(repoIDs, t.RepoID)
	}

	repoSpaceMap, err := loadRepoSpaceIDsByRepoIDs(ctx, s.db, repoIDs)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	conditionsIDMap, conditionsObjMap, err := loadTriggerConditionsByTriggerIDs(ctx, s.db, triggerIDs)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	solutionsIDMap, solutionsObjMap, err := loadTriggerSolutionsByTriggerIDs(ctx, s.db, triggerIDs)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	latestTaskMap, err := loadLatestTasksByTriggerIDs(ctx, s.db, triggerIDs)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	modelData := make([]*scanModel.ScanTrigger, 0, len(pageResult.Data))
	for _, item := range pageResult.Data {
		m := mapScanTriggerToModel(*item)
		m.SpaceID = repoSpaceMap[item.RepoID]
		m.ConditionIDs = conditionsIDMap[item.ID]
		m.Conditions = conditionsObjMap[item.ID]
		m.ConditionCount = len(conditionsIDMap[item.ID])
		m.SolutionIDs = solutionsIDMap[item.ID]
		m.Solutions = solutionsObjMap[item.ID]
		m.SolutionCount = len(solutionsIDMap[item.ID])
		if task, ok := latestTaskMap[item.ID]; ok {
			applyLatestTaskToTrigger(m, task)
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

// FindByRepoIDAndActions 查询指定仓库下、actions 有交集的所有未禁用触发器。
func (s *ScanTriggerStore) FindByRepoIDAndActions(ctx context.Context, repoID uint, actions []constant.TriggerAction) ([]*scanModel.ScanTrigger, error) {
	var rows []ScanPlanTrigger
	if err := s.db.WithContext(ctx).
		Where("repoID = ? AND deleted = ? AND disabled = ?", repoID, false, false).
		Find(&rows).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	actionSet := make(map[constant.TriggerAction]struct{}, len(actions))
	for _, a := range actions {
		actionSet[a] = struct{}{}
	}

	result := make([]*scanModel.ScanTrigger, 0, len(rows))
	for _, row := range rows {
		for _, a := range parseActions(row.Actions) {
			if _, ok := actionSet[a]; ok {
				result = append(result, mapScanTriggerToModel(row))
				break
			}
		}
	}
	return result, nil
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
	repoSpaceMap, err := loadRepoSpaceIDsByRepoIDs(ctx, s.db, []uint64{data.RepoID})
	if err == nil {
		m.SpaceID = repoSpaceMap[data.RepoID]
	}

	// 查询关联 conditions
	conditionsIDMap, conditionsObjMap, err := loadTriggerConditionsByTriggerIDs(ctx, s.db, []uint64{data.ID})
	if err == nil {
		m.ConditionIDs = conditionsIDMap[data.ID]
		m.Conditions = conditionsObjMap[data.ID]
		m.ConditionCount = len(conditionsIDMap[data.ID])
	}

	// 查询关联 solutions
	solutionsIDMap, solutionsObjMap, err := loadTriggerSolutionsByTriggerIDs(ctx, s.db, []uint64{data.ID})
	if err == nil {
		m.SolutionIDs = solutionsIDMap[data.ID]
		m.Solutions = solutionsObjMap[data.ID]
		m.SolutionCount = len(solutionsIDMap[data.ID])
	}

	return m
}
