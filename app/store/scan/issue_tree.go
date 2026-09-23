// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package scan

import (
	"context"

	"github.com/easysoft/gitfox-community/app/constant"
	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
	"github.com/easysoft/gitfox-community/pkg/database"
)

// ListDistinctFilesByRepo 按仓库或任务维度聚合 issue 文件路径
func (s *ScanIssueStore) ListDistinctFilesByRepo(ctx context.Context, repoID int64, scanMethods []constant.ScanMethod, taskFilter *scanModel.IssueTaskFilter) ([]*scanModel.IssueFileRaw, error) {
	if taskFilter != nil {
		return s.listDistinctFilesByTask(ctx, taskFilter, scanMethods)
	}
	return s.listDistinctFilesByRepo(ctx, repoID, scanMethods)
}

func (s *ScanIssueStore) listDistinctFilesByRepo(ctx context.Context, repoID int64, scanMethods []constant.ScanMethod) ([]*scanModel.IssueFileRaw, error) {
	var dst []*scanModel.IssueFileRaw
	stmt := s.db.WithContext(ctx).
		Table("ops_scan_issues").
		Select("repoBranch, path, COUNT(*) AS totalCount").
		Where("repoID = ? AND deleted = ?", repoID, 0).
		Group("repoBranch, path").
		Order("totalCount DESC")
	if len(scanMethods) > 0 {
		stmt = stmt.Where("scanMethod IN ?", scanMethods)
	}
	err := stmt.Find(&dst).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return dst, nil
}

func (s *ScanIssueStore) listDistinctFilesByTask(ctx context.Context, filter *scanModel.IssueTaskFilter, scanMethods []constant.ScanMethod) ([]*scanModel.IssueFileRaw, error) {
	var dst []*scanModel.IssueFileRaw
	stmt := s.db.WithContext(ctx).
		Table("ops_scan_issues AS i").
		Select("i.repoBranch, i.path, COUNT(*) AS totalCount").
		Joins("LEFT JOIN ops_scan_tasks AS t ON i.updatedByTaskID = t.id").
		Joins("LEFT JOIN ops_scan_plans AS p ON t.planID = p.id").
		Where("p.id = ? AND i.repoID = ? AND i.repoBranch = ? AND i.createdByTaskID <= ? AND i.deleted = ?",
			filter.PlanID, filter.RepoID, filter.Branch, filter.TaskID, false).
		Group("i.repoBranch, i.path").
		Order("totalCount DESC")
	if len(scanMethods) > 0 {
		stmt = stmt.Where("i.scanMethod IN ?", scanMethods)
	}
	err := stmt.Find(&dst).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return dst, nil
}

// ListDistinctRulesByRepo 按仓库或任务维度聚合 issue 规则
func (s *ScanIssueStore) ListDistinctRulesByRepo(ctx context.Context, repoID int64, taskFilter *scanModel.IssueTaskFilter) ([]*scanModel.IssueRuleRaw, error) {
	if taskFilter != nil {
		return s.listDistinctRulesByTask(ctx, taskFilter)
	}
	return s.listDistinctRulesByRepo(ctx, repoID)
}

func (s *ScanIssueStore) listDistinctRulesByRepo(ctx context.Context, repoID int64) ([]*scanModel.IssueRuleRaw, error) {
	var dst []*scanModel.IssueRuleRaw
	subQuery := s.db.Table("ops_scan_issues").
		Select("ruleID, COUNT(*) AS totalCount").
		Where("repoID = ? AND deleted = ?", repoID, 0).
		Group("ruleID")

	err := s.db.WithContext(ctx).
		Table("ops_scan_rules AS r").
		Select("r.id, r.name, r.lang, r.plugin, r.type, si.totalCount").
		Joins("JOIN (?) si ON r.id = si.ruleID", subQuery).
		Order("si.totalCount DESC").
		Find(&dst).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return dst, nil
}

func (s *ScanIssueStore) listDistinctRulesByTask(ctx context.Context, filter *scanModel.IssueTaskFilter) ([]*scanModel.IssueRuleRaw, error) {
	var dst []*scanModel.IssueRuleRaw
	subQuery := s.db.Table("ops_scan_issues AS i").
		Select("i.ruleID, COUNT(*) AS totalCount").
		Joins("LEFT JOIN ops_scan_tasks AS t ON i.updatedByTaskID = t.id").
		Joins("LEFT JOIN ops_scan_plans AS p ON t.planID = p.id").
		Where("p.id = ? AND i.repoID = ? AND i.repoBranch = ? AND i.createdByTaskID <= ? AND i.deleted = ?",
			filter.PlanID, filter.RepoID, filter.Branch, filter.TaskID, false).
		Group("i.ruleID")

	err := s.db.WithContext(ctx).
		Table("ops_scan_rules AS r").
		Select("r.id, r.name, r.lang, r.plugin, r.type, si.totalCount").
		Joins("JOIN (?) si ON r.id = si.ruleID", subQuery).
		Order("si.totalCount DESC").
		Find(&dst).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return dst, nil
}
