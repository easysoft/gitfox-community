package scan

import (
	"context"
	"fmt"
	"strings"

	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

var _ ISolution = (*SolutionStore)(nil)

type SolutionStore struct {
	db *gorm.DB
}

func NewSolutionStore(db *gorm.DB) *SolutionStore {
	return &SolutionStore{db: db}
}

func (s *SolutionStore) Create(ctx context.Context, solution *SolutionRules) (*scanModel.Solution, error) {
	if err := s.db.WithContext(ctx).Create(solution).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapSolutionToModel(*solution), nil
}

func (s *SolutionStore) GetByID(ctx context.Context, id uint) (*scanModel.Solution, error) {
	var solution SolutionRules
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted = ?", id, false).First(&solution).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapSolutionToModel(solution), nil
}

func (s *SolutionStore) Find(ctx context.Context, name string) (*scanModel.Solution, error) {
	var solution SolutionRules
	err := s.db.WithContext(ctx).Where("name = ? AND deleted = ?", name, false).First(&solution).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapSolutionToModel(solution), nil
}

func (s *SolutionStore) Update(ctx context.Context, solution *SolutionRules) error {
	if err := s.db.WithContext(ctx).Updates(solution).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// Delete 软删除扫描方案及其关联的规则集记录
func (s *SolutionStore) Delete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&SolutionRules{}).Where("id = ?", id).Update("deleted", true).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		if err := tx.Model(&SolutionRuleSets{}).Where("solutionID = ?", id).Update("deleted", true).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		return nil
	})
}

// List 查询扫描方案列表（支持分页和过滤）
func (s *SolutionStore) List(ctx context.Context, filter scanModel.SolutionReqFilter) (*database.PageResult[*scanModel.Solution], error) {
	tx := s.db.WithContext(ctx).Model(&SolutionRules{}).Where("deleted = ?", false)

	if filter.ID > 0 {
		tx = tx.Where("id = ?", filter.ID)
	}
	if filter.Name != "" {
		tx = tx.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.Tag != "" {
		tx = tx.Where("tag = ?", filter.Tag)
	}
	if filter.Status != "" {
		tx = tx.Where("status = ?", filter.Status)
	}

	var solutions []*SolutionRules
	pageResult, err := database.Paginate(ctx, tx, database.PageParam{
		Page:      filter.Page,
		PageSize:  filter.PageSize,
		Sort:      normalizeSolutionSortField(filter.Sort),
		Order:     normalizeSortOrder(filter.Order, "asc"),
		NeedTotal: true,
	}, &solutions)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// 批量查询聚合信息
	solutionIDs := make([]uint64, 0, len(pageResult.Data))
	for _, item := range pageResult.Data {
		solutionIDs = append(solutionIDs, item.ID)
	}

	// 每个方案的规则集数量
	ruleSetsCountMap := make(map[uint64]int64)
	// 每个方案的规则数量
	rulesCountMap := make(map[uint64]int64)

	if len(solutionIDs) > 0 {
		// 查询规则集数量
		type countRow struct {
			SolutionID uint64 `gorm:"column:solutionID"`
			Count      int64  `gorm:"column:cnt"`
		}
		var rsCounts []countRow
		var solutionRuleSetsModel SolutionRuleSets
		if err := s.db.WithContext(ctx).Table(solutionRuleSetsModel.TableName()).
			Select("solutionID, COUNT(*) as cnt").
			Where("solutionID IN ? AND deleted = ?", solutionIDs, false).
			Group("solutionID").Find(&rsCounts).Error; err == nil {
			for _, row := range rsCounts {
				ruleSetsCountMap[row.SolutionID] = row.Count
			}
		}

		// 查询规则数量（通过 solution_rulesets JOIN ruleset_rules）
		var ruleCounts []countRow
		if err := s.db.WithContext(ctx).Table("ops_scan_solution_rulesets AS ssr").
			Select("ssr.solutionID, COUNT(srr.id) as cnt").
			Joins("JOIN ops_scan_ruleset_rules AS srr ON srr.rulesetID = ssr.ruleSetID AND srr.deleted = ?", false).
			Where("ssr.solutionID IN ? AND ssr.deleted = ?", solutionIDs, false).
			Group("ssr.solutionID").Find(&ruleCounts).Error; err == nil {
			for _, row := range ruleCounts {
				rulesCountMap[row.SolutionID] = row.Count
			}
		}
	}

	modelData := make([]*scanModel.Solution, 0, len(pageResult.Data))
	for _, item := range pageResult.Data {
		m := mapSolutionToModel(*item)
		m.RuleSetsCount = ruleSetsCountMap[item.ID]
		m.RulesCount = rulesCountMap[item.ID]
		modelData = append(modelData, m)
	}

	return &database.PageResult[*scanModel.Solution]{
		Data:     modelData,
		Total:    pageResult.Total,
		Page:     pageResult.Page,
		PageSize: pageResult.PageSize,
	}, nil
}

func normalizeSolutionSortField(sort string) string {
	switch strings.TrimSpace(sort) {
	case "", "id":
		return "id"
	case "name":
		return "name"
	case "status":
		return "status"
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

// GetRulesetIDList 获取方案关联的规则集ID列表
func (s *SolutionStore) GetRulesetIDList(ctx context.Context, solutionID uint) ([]int64, error) {
	var rulesetIDs []int64
	err := s.db.WithContext(ctx).Model(&SolutionRuleSets{}).
		Where("solutionID = ? AND deleted = ?", solutionID, false).
		Pluck("ruleSetID", &rulesetIDs).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return rulesetIDs, nil
}

// ListRulesets 获取扫描方案已绑定的规则集列表（含每个规则集的规则数量）
func (s *SolutionStore) ListRulesets(ctx context.Context, solutionID uint) ([]*scanModel.RuleSet, error) {
	type rulesetWithCount struct {
		RuleSet
		RulesCount int64 `gorm:"column:rules_count"`
	}

	var rulesetsWithCount []rulesetWithCount
	err := s.db.WithContext(ctx).
		Table("ops_scan_solution_rulesets AS ssr").
		Select("r.*, (SELECT COUNT(*) FROM ops_scan_ruleset_rules WHERE rulesetID = r.id AND deleted = ?) as rules_count", false).
		Joins("LEFT JOIN ops_scan_rulesets AS r ON ssr.ruleSetID = r.id").
		Where("ssr.solutionID = ? AND ssr.deleted = ?", solutionID, false).
		Find(&rulesetsWithCount).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	result := make([]*scanModel.RuleSet, 0, len(rulesetsWithCount))
	for _, r := range rulesetsWithCount {
		m := mapToModel(r.RuleSet)
		m.RulesCount = r.RulesCount
		result = append(result, m)
	}
	return result, nil
}

// ListUnbindRulesets 获取扫描方案未绑定的规则集列表
func (s *SolutionStore) ListUnbindRulesets(ctx context.Context, solutionID uint) ([]*scanModel.RuleSet, error) {
	rulesetIDs, err := s.GetRulesetIDList(ctx, solutionID)
	if err != nil {
		return nil, err
	}

	tx := s.db.WithContext(ctx).Where("deleted = ?", false)
	if len(rulesetIDs) > 0 {
		tx = tx.Where("id NOT IN ?", rulesetIDs)
	}

	var rulesets []*RuleSet
	if err := tx.Find(&rulesets).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	result := make([]*scanModel.RuleSet, 0, len(rulesets))
	for _, r := range rulesets {
		result = append(result, mapToModel(*r))
	}
	return result, nil
}

// BindRulesets 为扫描方案绑定规则集
func (s *SolutionStore) BindRulesets(ctx context.Context, solutionID uint, rulesetIDs []uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 验证规则集是否存在
		var existingIDs []uint
		var rulesetModel RuleSet
		if err := tx.Table(rulesetModel.TableName()).
			Where("id IN ? AND deleted = ?", rulesetIDs, false).
			Pluck("id", &existingIDs).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		if len(existingIDs) != len(rulesetIDs) {
			return fmt.Errorf("some rulesets not found")
		}

		for _, rulesetID := range rulesetIDs {
			var existing SolutionRuleSets
			err := tx.Where("solutionID = ? AND ruleSetID = ?", solutionID, rulesetID).First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				newRecord := SolutionRuleSets{
					SolutionID: uint64(solutionID),
					RuleSetID:  uint64(rulesetID),
					Deleted:    false,
				}
				if err := tx.Create(&newRecord).Error; err != nil {
					return database.ProcessSQLErrorf(ctx, err)
				}
			} else if err != nil {
				return database.ProcessSQLErrorf(ctx, err)
			} else if existing.Deleted {
				if err := tx.Model(&SolutionRuleSets{}).
					Where("solutionID = ? AND ruleSetID = ?", solutionID, rulesetID).
					Update("deleted", false).Error; err != nil {
					return database.ProcessSQLErrorf(ctx, err)
				}
			}
		}
		return nil
	})
}

// UnbindRulesets 解除扫描方案与规则集的绑定
func (s *SolutionStore) UnbindRulesets(ctx context.Context, solutionID uint, rulesetIDs []uint) error {
	if err := s.db.WithContext(ctx).Model(&SolutionRuleSets{}).
		Where("solutionID = ? AND ruleSetID IN ?", solutionID, rulesetIDs).
		Update("deleted", true).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// GetSolutionsByPlanID 获取计划关联的方案ID列表
func (s *SolutionStore) GetSolutionsByPlanID(ctx context.Context, planID uint) ([]*scanModel.Solution, []int64, error) {
	var ids []int64
	var solutions []*scanModel.Solution
	err := s.db.WithContext(ctx).Model(&ScanPlanSolution{}).
		Where("planID = ? AND deleted = ?", planID, false).
		Pluck("solutionID", &ids).Error
	if err != nil {
		return nil, nil, database.ProcessSQLErrorf(ctx, err)
	}

	for _, id := range ids {
		solution, err := s.GetByID(ctx, uint(id))
		if err != nil {
			return nil, nil, err
		}
		solutions = append(solutions, solution)
	}

	return solutions, ids, nil
}

// IsUsedByPlan 检查扫描方案是否被扫描计划使用
func (s *SolutionStore) IsUsedByPlan(ctx context.Context, solutionID uint) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&ScanPlanSolution{}).
		Where("solutionID = ? AND deleted = ?", solutionID, false).
		Count(&count).Error
	if err != nil {
		return false, database.ProcessSQLErrorf(ctx, err)
	}
	return count > 0, nil
}

func mapSolutionToModel(data SolutionRules) *scanModel.Solution {
	return &scanModel.Solution{
		ID:          int64(data.ID),
		Name:        data.Name,
		Desc:        data.Desc,
		DescEn:      data.DescEn,
		Tag:         data.Tag,
		IsCustom:    data.IsCustom,
		Status:      data.Status,
		CreatedBy:   data.CreatedBy,
		CreatedDate: data.CreatedDate,
		EditedBy:    data.EditedBy,
		EditedDate:  data.EditedDate,
		Deleted:     data.Deleted,
	}
}
