package scan

import (
	"context"

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
		ID:          data.ID,
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
