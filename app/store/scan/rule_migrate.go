package scan

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/resources"
	"gorm.io/gorm"
)

// Migrate 从嵌入的 JSON 文件中迁移规则数据
func (s *RuleStore) Migrate(ctx context.Context, rulesFile *resources.RulesFile) error {
	now := datetime.NowLocalTime()

	// 使用事务批量插入或更新
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// 1. 检查该插件的规则是否已迁移
		version, err := s.GetRuleMigrationVersion(ctx, tx, rulesFile.Plugin, rulesFile.Lang)
		if err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}

		// 版本一致时也需要确保默认扫描方案与规则集绑定关系存在
		if version == rulesFile.Version {
			return createMigrateScanRuleSolutions(ctx, tx, rulesFile, now)
		}

		if version == "" {
			// 未迁移过，直接插入
			return insertRules(ctx, tx, rulesFile, true, now)
		}

		//  已迁移过 判断版本
		if compareNumericVersions(version, rulesFile.Version) < 0 {
			return insertRules(ctx, tx, rulesFile, false, now)
		}

		// 当前版本较新时，仅做默认方案与规则集绑定关系兜底
		return createMigrateScanRuleSolutions(ctx, tx, rulesFile, now)
	})
}

// parsePriority 将字符串转换为 ScanRulePriority
func parsePriority(priority string) constant.ScanRulePriority {
	switch strings.ToLower(priority) {
	case "low":
		return constant.ScanRulePriorityLow
	case "medium":
		return constant.ScanRulePriorityMedium
	case "high":
		return constant.ScanRulePriorityHigh
	default:
		return constant.ScanRulePriorityMedium // 默认中等优先级
	}
}

// GetRuleMigrationVersion 获取规则迁移版本
func (s *RuleStore) GetRuleMigrationVersion(ctx context.Context, tx *gorm.DB, plugin string, lang string) (string, error) {
	if plugin == "" || lang == "" {
		return "", errors.New("plugin or lang is empty")
	}
	var migration RuleMigration
	err := tx.Where("name = ? AND deleted = ?", getRuleMigrationName(plugin, lang), false).First(&migration).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", database.ProcessSQLErrorf(ctx, err)
	}
	return migration.Version, nil
}

// getRuleMigrationName 获取规则迁移名称
func getRuleMigrationName(plugin string, lang string) string {
	return fmt.Sprintf("%s-%s", plugin, lang)
}

// 直接插入 规则数据
func insertRules(ctx context.Context, tx *gorm.DB, rulesFile *resources.RulesFile, firstRun bool, now *datetime.Local) error {

	if !firstRun {
		// 删除默认规则
		if err := tx.Where("plugin = ? AND lang = ? AND isCustom = ? AND deleted = ?", rulesFile.Plugin, rulesFile.Lang, false, false).Delete(&Rule{}).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
	}

	// 组装数据 插入规则数据
	for _, ruleJSON := range rulesFile.Rules {
		rule := &Rule{
			RuleKey:     ruleJSON.RuleKey,
			Name:        ruleJSON.Name,
			Desc:        ruleJSON.Description,
			DescEn:      ruleJSON.DescriptionEn,
			Lang:        constant.ScanLangName(ruleJSON.Lang),
			Plugin:      ruleJSON.Plugin,
			Tag:         ruleJSON.Tag,
			Priority:    parsePriority(ruleJSON.Priority),
			Type:        constant.ScanRuleType(ruleJSON.Type),
			Content:     database.NoNull(ruleJSON.Content),
			ContentEn:   database.NoNull(ruleJSON.ContentEn),
			IsCustom:    false,
			Status:      constant.ScanRuleStatusEnabled,
			CreatedBy:   "system",
			CreatedDate: now,
			EditedBy:    "system",
			EditedDate:  now,
			Deleted:     false,
		}
		err := tx.Create(rule).Error
		if err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
	}

	// 创建规则集合
	err := createMigrateScanRuleSet(ctx, tx, rulesFile, now)
	if err != nil {
		return err
	}

	// 创建解决方案
	err = createMigrateScanRuleSolutions(ctx, tx, rulesFile, now)
	if err != nil {
		return err
	}

	// 更新规则版本
	err = updateOrCreateMigrateScanRuleVersion(ctx, tx, rulesFile, now)
	if err != nil {
		return err
	}
	return nil
}

// 创建默认规则集
func createMigrateScanRuleSet(ctx context.Context, tx *gorm.DB, rulesFile *resources.RulesFile, now *datetime.Local) error {
	ruleSetName := fmt.Sprintf("%s:%s默认规则集", rulesFile.Lang, rulesFile.Plugin)
	ruleSetNameEn := fmt.Sprintf("%s:%s default rule set", rulesFile.Lang, rulesFile.Plugin)
	if rulesFile.Lang == string(constant.ScanLangCustomName) {
		ruleSetName = fmt.Sprintf("%s默认规则集", rulesFile.Plugin)
		ruleSetNameEn = fmt.Sprintf("%s default rule set", rulesFile.Plugin)
	}

	var oldRuleSet RuleSet
	err := tx.Where("name = ? AND lang = ? AND plugin = ? AND isCustom = ? AND deleted = ?", ruleSetName, rulesFile.Lang, rulesFile.Plugin, false, false).Find(&oldRuleSet).Error
	if err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}

	if oldRuleSet.ID != 0 {
		if err = tx.Where("id = ? AND isCustom = ? AND deleted = ?", oldRuleSet.ID, false, false).Delete(&RuleSet{}).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}

		// 删除set 对应的 关联关系
		if err = tx.Where("rulesetID = ?", oldRuleSet.ID).Delete(&RuleSetRules{}).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
	}

	// 创建规则集合
	ruleSet := &RuleSet{
		Name:        ruleSetName,
		Desc:        ruleSetName,
		DescEn:      ruleSetNameEn,
		Lang:        constant.ScanLangName(rulesFile.Lang),
		IsCustom:    false,
		Status:      constant.ScanRuleStatusEnabled,
		Plugin:      rulesFile.Plugin,
		CreatedBy:   "system",
		CreatedDate: now,
		EditedBy:    "system",
		EditedDate:  now,
		Deleted:     false,
	}
	err = tx.Create(ruleSet).Error
	if err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}

	err = createMigrateScanRuleSetRules(ctx, tx, ruleSet, now)
	if err != nil {
		return err
	}
	return nil
}

// 创建默认规则关联的规则
func createMigrateScanRuleSetRules(ctx context.Context, tx *gorm.DB, ruleSet *RuleSet, now *datetime.Local) error {
	// 获取所有规则的ID
	var ruleIDs []int64
	if err := tx.Model(&Rule{}).Select("id").Where("lang = ? AND plugin = ? AND isCustom = ? AND deleted = ?", ruleSet.Lang, ruleSet.Plugin, false, false).Scan(&ruleIDs).Error; err != nil {
		return err
	}

	if len(ruleIDs) == 0 {
		return errors.New("ruleIDs not found")
	}

	return rulesSetBindRules(ctx, tx, ruleSet, ruleIDs, now)
}

// 创建默认解决方案（幂等：存在则补齐绑定，不删除重建）
func createMigrateScanRuleSolutions(ctx context.Context, tx *gorm.DB, rulesFile *resources.RulesFile, now *datetime.Local) error {
	solutionName := fmt.Sprintf("%s默认扫描方案", rulesFile.Lang)
	solutionNameEn := fmt.Sprintf("%s default scan solution", rulesFile.Lang)
	if rulesFile.Lang == string(constant.ScanLangCustomName) {
		solutionName = fmt.Sprintf("%s默认扫描方案", rulesFile.Plugin)
		solutionNameEn = fmt.Sprintf("%s default scan solution", rulesFile.Plugin)
	}

	var solution SolutionRules
	err := tx.Where("name = ? AND isCustom = ? AND deleted = ?", solutionName, false, false).First(&solution).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return database.ProcessSQLErrorf(ctx, err)
		}

		solution = SolutionRules{
			Name:        solutionName,
			Desc:        solutionName,
			DescEn:      solutionNameEn,
			IsCustom:    false,
			Status:      constant.ScanRuleStatusEnabled,
			CreatedBy:   "system",
			CreatedDate: now,
			EditedBy:    "system",
			EditedDate:  now,
			Deleted:     false,
		}
		if err = tx.Create(&solution).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
	}

	// 获取默认规则集ID（按语言）
	var ruleSetIDs []int64
	if err := tx.Model(&RuleSet{}).Select("id").Where("lang = ? AND isCustom = ? AND deleted = ?", rulesFile.Lang, false, false).Scan(&ruleSetIDs).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	if len(ruleSetIDs) == 0 {
		return errors.New("ruleSetIDs not found")
	}

	// 查询已有有效绑定
	var boundRuleSetIDs []int64
	if err := tx.Model(&SolutionRuleSets{}).
		Where("solutionID = ? AND deleted = ?", solution.ID, false).
		Pluck("ruleSetID", &boundRuleSetIDs).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	boundSet := make(map[int64]struct{}, len(boundRuleSetIDs))
	for _, id := range boundRuleSetIDs {
		boundSet[id] = struct{}{}
	}

	for _, ruleSetID := range ruleSetIDs {
		if _, exists := boundSet[ruleSetID]; exists {
			continue
		}

		// 尝试恢复被软删的历史绑定
		result := tx.Model(&SolutionRuleSets{}).
			Where("solutionID = ? AND ruleSetID = ? AND deleted = ?", solution.ID, uint64(ruleSetID), true).
			Updates(map[string]interface{}{
				"deleted":    false,
				"editedBy":   "system",
				"editedDate": now,
			})
		if result.Error != nil {
			return database.ProcessSQLErrorf(ctx, result.Error)
		}
		if result.RowsAffected > 0 {
			continue
		}

		// 没有历史记录则新增绑定
		solutionRuleSet := &SolutionRuleSets{
			SolutionID:  uint64(solution.ID),
			RuleSetID:   uint64(ruleSetID),
			CreatedBy:   "system",
			CreatedDate: now,
			EditedBy:    "system",
			EditedDate:  now,
			Deleted:     false,
		}
		if err := tx.Create(solutionRuleSet).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
	}

	return nil
}

// 更新或者创建规则版本
func updateOrCreateMigrateScanRuleVersion(ctx context.Context, tx *gorm.DB, rulesFile *resources.RulesFile, now *datetime.Local) error {
	migrationName := getRuleMigrationName(rulesFile.Plugin, rulesFile.Lang)

	// 先查询是否存在
	var existing RuleMigration
	err := tx.Where("name = ? AND deleted = ?", migrationName, false).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// 不存在，创建新记录
		migration := RuleMigration{
			Name:        migrationName,
			Version:     rulesFile.Version,
			CreatedBy:   "system",
			CreatedDate: now,
			Deleted:     false,
		}
		if err = tx.Create(&migration).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
	} else if err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	} else {
		// 已存在，更新版本
		updates := map[string]interface{}{
			"version":    rulesFile.Version,
			"editedBy":   "system",
			"editedDate": now,
		}
		if err := tx.Model(&existing).Updates(updates).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
	}

	return nil
}

// 关联规则集与规则
func rulesSetBindRules(ctx context.Context, tx *gorm.DB, ruleSet *RuleSet, ruleIDs []int64, now *datetime.Local) error {
	if len(ruleIDs) == 0 {
		return nil
	}

	for _, ruleID := range ruleIDs {
		newRuleSetRule := &RuleSetRules{
			RuleSetID:   ruleSet.ID,
			RuleID:      uint64(ruleID),
			CreatedBy:   "system",
			CreatedDate: now,
			EditedBy:    "system",
			EditedDate:  now,
			Deleted:     false,
		}
		if err := tx.Create(newRuleSetRule).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
	}

	return nil
}

// compareNumericVersions 比较两个版本号的大小
func compareNumericVersions(v1, v2 string) int {
	num1, err1 := strconv.ParseInt(v1, 10, 64)
	num2, err2 := strconv.ParseInt(v2, 10, 64)
	// 如果解析失败，使用字符串比较作为后备
	if err1 != nil || err2 != nil {
		if v1 < v2 {
			return -1
		} else if v1 > v2 {
			return 1
		}
		return 0
	}
	// 数字比较
	if num1 < num2 {
		return -1 // v1 < v2
	} else if num1 > num2 {
		return 1 // v1 > v2
	}
	return 0 // v1 == v2
}
