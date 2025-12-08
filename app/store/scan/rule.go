package scan

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

var _ IRule = (*RuleStore)(nil)

type RuleStore struct {
	db *gorm.DB
}

func NewRuleStore(db *gorm.DB) *RuleStore {
	return &RuleStore{
		db: db,
	}
}

// ListByRuleSetID 根据规则集ID查询关联的规则列表
func (s *RuleStore) ListByRuleSetID(
	ctx context.Context,
	ruleIDs []int64,
	filter scanModel.RuleReqFilter,
) (*database.PageResult[*scanModel.Rule], error) {

	tx := s.db.WithContext(ctx).Model(&Rule{}).
		Where("id IN ? AND deleted = ?", ruleIDs, false)

	// 分页查询
	var rules []*Rule
	pageResult, err := database.PaginateDefault(ctx, tx, filter.Page, filter.Size, &rules)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// 转换为 model 层数据
	modelData := make([]*scanModel.Rule, 0, len(pageResult.Data))
	for _, item := range pageResult.Data {
		modelData = append(modelData, mapRuleToModel(item))
	}

	return &database.PageResult[*scanModel.Rule]{
		Data:     modelData,
		Total:    pageResult.Total,
		Page:     pageResult.Page,
		PageSize: pageResult.PageSize,
	}, nil
}

// mapRuleToModel 将 store.Rule 转换为 scanModel.Rule
func mapRuleToModel(data *Rule) *scanModel.Rule {
	if data == nil {
		return nil
	}
	return &scanModel.Rule{
		ID:          int64(data.ID),
		RuleKey:     data.RuleKey,
		Name:        data.Name,
		Desc:        data.Desc,
		DescEn:      data.DescEn,
		Lang:        data.Lang,
		Plugin:      data.Plugin,
		Tag:         data.Tag,
		Priority:    data.Priority,
		Type:        data.Type,
		Content:     data.Content,
		ContentEn:   data.ContentEn,
		IsCustom:    data.IsCustom,
		Status:      data.Status,
		CreatedBy:   data.CreatedBy,
		CreatedDate: data.CreatedDate,
		EditedBy:    data.EditedBy,
		EditedDate:  data.EditedDate,
		Deleted:     data.Deleted,
	}
}

// ExcludeByRuleSetID 查询排除指定ID的规则列表
func (s *RuleStore) ExcludeByRuleSetID(
	ctx context.Context,
	excludeIDs []int64,
	filter scanModel.RuleReqFilter,
) (*database.PageResult[*scanModel.Rule], error) {
	tx := s.db.WithContext(ctx).Model(&Rule{}).Where("deleted = ?", false)

	// 排除已绑定的规则ID
	if len(excludeIDs) > 0 {
		tx = tx.Where("id NOT IN ?", excludeIDs)
	}

	// 分页查询
	var rules []*Rule
	pageResult, err := database.PaginateDefault(ctx, tx, filter.Page, filter.Size, &rules)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// 转换为 model 层数据
	modelData := make([]*scanModel.Rule, 0, len(pageResult.Data))
	for _, item := range pageResult.Data {
		modelData = append(modelData, mapRuleToModel(item))
	}

	return &database.PageResult[*scanModel.Rule]{
		Data:     modelData,
		Total:    pageResult.Total,
		Page:     pageResult.Page,
		PageSize: pageResult.PageSize,
	}, nil
}

// GetByID 根据ID查询规则
func (s *RuleStore) GetByID(ctx context.Context, id uint) (*scanModel.Rule, error) {
	var rule Rule
	err := s.db.WithContext(ctx).Where("deleted = ?", false).First(&rule, id).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapRuleToModel(&rule), nil
}

// List 查询规则列表
func (s *RuleStore) List(ctx context.Context, filter scanModel.RuleListReqFilter) (*database.PageResult[*scanModel.Rule], error) {
	tx := s.db.WithContext(ctx).Model(&Rule{}).Where("ops_scan_rules.deleted = ?", false)

	if filter.Bind == scanModel.ScanRuleBindBind {
		if filter.RuleSetID > 0 {
			subQuery := s.db.Model(&RuleSetRules{}).
				Select("ruleID").
				Where("rulesetID = ? AND deleted = ?", filter.RuleSetID, false)
			tx = tx.Where("ops_scan_rules.id IN (?)", subQuery)
		} else {
			subQuery := s.db.Model(&RuleSetRules{}).
				Distinct("ruleID").
				Where("deleted = ?", false)
			tx = tx.Where("ops_scan_rules.id IN (?)", subQuery)
		}
	}

	// 过滤查询参数
	if filter.Name != "" {
		tx = tx.Where("ops_scan_rules.name LIKE ?", "%"+filter.Name+"%")
	}

	if filter.Lang != "" {
		tx = tx.Where("ops_scan_rules.lang = ?", filter.Lang)
	}

	if filter.Plugin != "" {
		tx = tx.Where("ops_scan_rules.plugin = ?", filter.Plugin)
	}

	if filter.Tag != "" {
		tx = tx.Where("ops_scan_rules.tag = ?", filter.Tag)
	}

	if filter.Priority != "" {
		tx = tx.Where("ops_scan_rules.priority = ?", filter.Priority)
	}

	if filter.Status != "" {
		tx = tx.Where("ops_scan_rules.status = ?", filter.Status)
	}

	if filter.Type != "" {
		tx = tx.Where("ops_scan_rules.type = ?", filter.Type)
	}

	// 排序：组合 sort 和 order
	if filter.Sort != "" && filter.Order != "" {
		tx = tx.Order("ops_scan_rules." + string(filter.Sort) + " " + string(filter.Order))
	}

	// 分页查询
	var rules []*Rule
	pageResult, err := database.PaginateDefault(ctx, tx, filter.Page, filter.PageSize, &rules)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// 转换为 model 层数据
	modelData := make([]*scanModel.Rule, 0, len(pageResult.Data))
	for _, item := range pageResult.Data {
		modelData = append(modelData, mapRuleToModel(item))
	}

	return &database.PageResult[*scanModel.Rule]{
		Data:     modelData,
		Total:    pageResult.Total,
		Page:     pageResult.Page,
		PageSize: pageResult.PageSize,
	}, nil
}

// UpdateStatus 切换规则状态（enabled ↔ disabled）
func (s *RuleStore) UpdateStatus(ctx context.Context, id uint) error {
	var rule Rule
	// 查询规则
	err := s.db.WithContext(ctx).Where("deleted = ?", false).First(&rule, id).Error
	if err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}

	// 使用 UpdateStatus 方法切换状态
	rule.Status = rule.Status.UpdateStatus()

	// 保存更新
	return database.ProcessSQLErrorf(ctx, s.db.WithContext(ctx).Save(&rule).Error)
}

func (s *RuleStore) AllScanBaseByField(ctx context.Context, field string) ([]scanModel.ScanBaseResp, error) {
	var rules []scanModel.ScanBaseResp
	err := s.db.WithContext(ctx).
		Model(&Rule{}).
		Select(fmt.Sprintf("%s, COUNT(*) as count", field)).
		Where("deleted = ?", false).
		Group(field).
		Order("count DESC, " + field + " ASC").
		Find(&rules).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return rules, nil
}
