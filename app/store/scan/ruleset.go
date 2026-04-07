package scan

import (
	"context"
	"strings"

	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

var _ IRuleSet = (*RuleSetStore)(nil)

type RuleSetStore struct {
	db *gorm.DB
}

func NewRuleSetStore(db *gorm.DB) *RuleSetStore {
	return &RuleSetStore{
		db: db,
	}
}

func (s *RuleSetStore) Create(ctx context.Context, ruleSet *RuleSet) (*scanModel.RuleSet, error) {
	if err := s.db.WithContext(ctx).Create(&ruleSet).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapToModel(*ruleSet), nil
}

func (s *RuleSetStore) GetByID(ctx context.Context, id uint) (*scanModel.RuleSet, error) {
	type rulesetWithCount struct {
		RuleSet
		RulesCount int64 `gorm:"column:rules_count"`
	}

	var ruleSet rulesetWithCount
	if err := s.db.WithContext(ctx).
		Table("ops_scan_rulesets AS r").
		Select("r.*, (SELECT COUNT(*) FROM ops_scan_ruleset_rules WHERE rulesetID = r.id AND deleted = ?) as rules_count", false).
		Where("r.id = ?", id).
		First(&ruleSet).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	m := mapToModel(ruleSet.RuleSet)
	m.RulesCount = ruleSet.RulesCount
	return m, nil
}

func (s *RuleSetStore) Update(ctx context.Context, ruleSet *RuleSet, id uint) (*scanModel.RuleSet, error) {
	ruleSet.ID = uint64(id)
	if err := s.db.WithContext(ctx).Updates(ruleSet).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// 重新查询完整记录,因为 Updates 只更新指定非零值字段,不会回填其他字段
	var updated RuleSet
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&updated).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return mapToModel(updated), nil
}

// Delete 软删除规则集及其关联的规则记录
func (s *RuleSetStore) Delete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 软删除规则集
		if err := tx.Model(&RuleSet{}).Where("id = ?", id).Update("deleted", true).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}

		// 软删除关联表中的记录
		if err := tx.Model(&RuleSetRules{}).Where("ruleSetID = ?", id).Update("deleted", true).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}

		return nil
	})
}

// IsUsedBySolution 检查规则集是否被扫描方案使用
func (s *RuleSetStore) IsUsedBySolution(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&SolutionRuleSets{}).
		Where("ruleSetID = ? AND deleted = ?", id, false).
		Count(&count).Error
	if err != nil {
		return false, database.ProcessSQLErrorf(ctx, err)
	}
	return count > 0, nil
}

// List 查询规则集列表（支持分页和过滤）
func (s *RuleSetStore) List(ctx context.Context, filter scanModel.RuleSetReqFilter) (*database.PageResult[*scanModel.RuleSet], error) {
	type rulesetWithCount struct {
		RuleSet
		RulesCount int64 `gorm:"column:rules_count"`
	}

	tx := s.db.WithContext(ctx).
		Table("ops_scan_rulesets AS r").
		Select("r.*, (SELECT COUNT(*) FROM ops_scan_ruleset_rules WHERE rulesetID = r.id AND deleted = ?) as rules_count", false).
		Where("r.deleted = ?", false)

	// 过滤条件
	if filter.ID != 0 {
		tx = tx.Where("r.id = ?", filter.ID)
	}
	if filter.Name != "" {
		tx = tx.Where("r.name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.Lang != "" {
		tx = tx.Where("r.lang = ?", strings.ToLower(filter.Lang))
	}
	if filter.Plugin != "" {
		tx = tx.Where("r.plugin = ?", filter.Plugin)
	}
	if filter.Tag != "" {
		tx = tx.Where("r.tag = ?", filter.Tag)
	}
	if filter.Type != "" {
		tx = tx.Where("r.type = ?", filter.Type)
	}
	if filter.Status != "" {
		tx = tx.Where("r.status = ?", filter.Status)
	}

	// 分页查询
	var ruleSets []*rulesetWithCount
	pageResult, err := database.Paginate(ctx, tx, database.PageParam{
		Page:      filter.Page,
		PageSize:  filter.PageSize,
		Sort:      normalizeRuleSetSortField(filter.Sort),
		Order:     normalizeSortOrder(filter.Order, "asc"),
		NeedTotal: true,
	}, &ruleSets)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// 转换为 model 层数据
	modelData := make([]*scanModel.RuleSet, 0, len(pageResult.Data))
	for _, item := range pageResult.Data {
		m := mapToModel(item.RuleSet)
		m.RulesCount = item.RulesCount
		modelData = append(modelData, m)
	}

	return &database.PageResult[*scanModel.RuleSet]{
		Data:     modelData,
		Total:    pageResult.Total,
		Page:     pageResult.Page,
		PageSize: pageResult.PageSize,
	}, nil
}

func normalizeRuleSetSortField(sort string) string {
	switch strings.TrimSpace(sort) {
	case "", "id":
		return "id"
	case "name":
		return "name"
	case "lang":
		return "lang"
	case "tag":
		return "tag"
	case "type":
		return "type"
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

// mapToModel 将 RuleSet 转换为 scanModel.RuleSet
func mapToModel(data RuleSet) *scanModel.RuleSet {
	return &scanModel.RuleSet{
		ID:          int64(data.ID),
		Name:        data.Name,
		Desc:        data.Desc,
		DescEn:      data.DescEn,
		Lang:        data.Lang,
		Plugin:      data.Plugin,
		Tag:         data.Tag,
		Status:      data.Status,
		CreatedBy:   data.CreatedBy,
		CreatedDate: data.CreatedDate,
		EditedBy:    data.EditedBy,
		EditedDate:  data.EditedDate,
		IsCustom:    data.IsCustom,
		Deleted:     data.Deleted,
	}
}

// GetRuleIDList 根据规则集ID获取关联的规则ID列表
func (s *RuleSetStore) GetRuleIDList(ctx context.Context, ruleSetID uint) ([]int64, error) {
	var ruleIDs []int64
	err := s.db.WithContext(ctx).Model(&RuleSetRules{}).
		Where("rulesetID = ? AND deleted = ?", ruleSetID, false).
		Pluck("ruleID", &ruleIDs).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return ruleIDs, nil
}

// BindRule 绑定规则到规则集（将 deleted 设置为 false，不存在则新增记录）
func (s *RuleSetStore) BindRule(ctx context.Context, ruleSetID uint, ruleIDs []uint) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, ruleID := range ruleIDs {
			// 查询关联记录是否存在
			var existingRecord RuleSetRules
			err := tx.Where("rulesetID = ? AND ruleID = ?", ruleSetID, ruleID).
				First(&existingRecord).Error

			if err == gorm.ErrRecordNotFound {
				// 记录不存在，插入新记录
				newRecord := RuleSetRules{
					RuleSetID: uint64(ruleSetID),
					RuleID:    uint64(ruleID),
					Deleted:   false,
				}
				if err := tx.Create(&newRecord).Error; err != nil {
					return database.ProcessSQLErrorf(ctx, err)
				}
			} else if err != nil {
				// 查询出错
				return database.ProcessSQLErrorf(ctx, err)
			} else if existingRecord.Deleted {
				// 记录存在但已被删除，恢复记录
				if err := tx.Model(&RuleSetRules{}).
					Where("rulesetID = ? AND ruleID = ?", ruleSetID, ruleID).
					Update("deleted", false).Error; err != nil {
					return database.ProcessSQLErrorf(ctx, err)
				}
			}
			// 如果记录存在且未被删除，不做任何操作
		}
		return nil
	})
	if err != nil {
		return err
	}
	return s.autoUpdatePlugin(ctx, ruleSetID)
}

// UnBindRule 解绑规则与规则集（将 deleted 设置为 true）
func (s *RuleSetStore) UnBindRule(ctx context.Context, ruleSetID uint, ruleIDs []uint) error {
	err := s.db.WithContext(ctx).Model(&RuleSetRules{}).
		Where("rulesetID = ? AND ruleID IN ?", ruleSetID, ruleIDs).
		Update("deleted", true).Error
	if err != nil {
		return err
	}
	return s.autoUpdatePlugin(ctx, ruleSetID)
}

// autoUpdatePlugin 根据规则集当前绑定的规则自动更新 plugin 字段。
// 统计所有已绑定规则的 plugin 分布，去除 custom 类，按数量降序用逗号拼接后写回规则集。
// 若规则集已无任何绑定规则，则将 plugin 重置为 "default"。
func (s *RuleSetStore) autoUpdatePlugin(ctx context.Context, ruleSetID uint) error {
	type pluginCount struct {
		Plugin string `gorm:"column:plugin"`
		Count  int64  `gorm:"column:count"`
	}

	var counts []pluginCount
	err := s.db.WithContext(ctx).
		Table("ops_scan_rules AS r").
		Select("r.plugin, COUNT(*) AS count").
		Joins("JOIN ops_scan_ruleset_rules AS rsr ON r.id = rsr.ruleID").
		Where("rsr.rulesetID = ? AND rsr.deleted = ?", ruleSetID, false).
		Group("r.plugin").
		Order("count DESC").
		Find(&counts).Error
	if err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}

	var plugin string
	if len(counts) == 0 {
		plugin = "default"
	} else {
		var plugins []string
		for _, pc := range counts {
			if strings.Contains(strings.ToLower(pc.Plugin), "custom") {
				continue
			}
			plugins = append(plugins, pc.Plugin)
		}
		plugin = strings.Join(plugins, ",")
	}

	// 与当前值一致时跳过写库
	var current RuleSet
	if err := s.db.WithContext(ctx).Select("plugin").Where("id = ?", ruleSetID).First(&current).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	if current.Plugin == plugin {
		return nil
	}

	return s.db.WithContext(ctx).
		Model(&RuleSet{}).
		Where("id = ?", ruleSetID).
		Update("plugin", plugin).Error
}

