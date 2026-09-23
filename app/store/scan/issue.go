// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package scan

import (
	"context"
	"strings"

	"github.com/easysoft/gitfox-community/app/constant"
	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
	"github.com/easysoft/gitfox-community/pkg/database"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

func normalizeIssueSortField(sort string) string {
	switch strings.TrimSpace(sort) {
	case "", "id":
		return "i.id"
	case "rulePriority", "priority":
		return "r.priority"
	case "ruleType":
		return "r.type"
	case "status":
		return "i.status"
	case "createDate":
		return "i.createdDate"
	default:
		return "i.id"
	}
}

func normalizeIssueSortOrder(order string) string {
	switch strings.ToLower(strings.TrimSpace(order)) {
	case "asc":
		return "asc"
	case "desc":
		return "desc"
	default:
		return "desc"
	}
}

var _ IScanIssue = (*ScanIssueStore)(nil)

type ScanIssueStore struct {
	db *gorm.DB
}

func NewScanIssueStore(db *gorm.DB) *ScanIssueStore {
	return &ScanIssueStore{db: db}
}

// Create 创建扫描问题
func (s *ScanIssueStore) Create(ctx context.Context, issue *ScanIssue) error {
	if err := s.db.WithContext(ctx).Create(issue).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// BatchCreate 批量创建扫描问题
func (s *ScanIssueStore) BatchCreate(ctx context.Context, issues []*ScanIssue) error {
	if len(issues) == 0 {
		return nil
	}
	if err := s.db.WithContext(ctx).Create(&issues).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// Get 查询单个扫描问题详情（含规则信息）
func (s *ScanIssueStore) Get(ctx context.Context, id uint) (*scanModel.ScanIssueInfo, error) {
	var item issueJoin
	if err := s.db.WithContext(ctx).
		Table("ops_scan_issues AS i").
		Select("i.id, i.issueKey, i.ruleID, r.name AS ruleName, r.ruleKey AS ruleKey, "+
			"r.priority AS rulePriority, r.type AS ruleType, "+
			"i.title, i.content, i.path, i.startLine, i.endLine, i.oldCode, i.newCode, i.category, i.severity, "+
			"i.ppmID, i.pipelineExecutionID, i.repoID, i.repoBranch, "+
			"i.status, i.scanMethod, i.resolution, i.payload, i.createdByTaskID, i.updatedByTaskID, "+
			"i.createdBy, i.editedBy, pp.title AS ppmTitle, "+
			"t.planID, p.name AS planName, i.createdDate, i.resolvedDate, i.closedDate, i.ignoredDate, "+
			"t.triggerID, t.triggerType").
		Joins("LEFT JOIN ops_scan_rules AS r ON i.ruleID = r.id").
		Joins("LEFT JOIN ops_scan_tasks AS t ON i.updatedByTaskID = t.id").
		Joins("LEFT JOIN ops_scan_plans AS p ON t.planID = p.id").
		Joins("LEFT JOIN ops_ppm AS pp ON i.ppmID = pp.id").
		Where("i.id = ? AND i.deleted = ?", id, 0).
		First(&item).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapIssueJoinToModel(&item), nil
}

// List 查询扫描问题列表（含规则信息，支持分页）
func (s *ScanIssueStore) List(ctx context.Context, filter scanModel.ScanIssueReqFilter) (*database.PageResult[*scanModel.ScanIssueInfo], error) {
	stmt := s.db.WithContext(ctx).
		Table("ops_scan_issues AS i").
		Select("i.id, i.issueKey, i.ruleID, r.name AS ruleName, r.ruleKey AS ruleKey, "+
			"r.priority AS rulePriority, r.type AS ruleType, "+
			"i.title, i.content, i.path, i.startLine, i.endLine, i.oldCode, i.newCode, i.category, i.severity, "+
			"i.ppmID, i.pipelineExecutionID, i.repoID, i.repoBranch, "+
			"i.status, i.scanMethod, i.resolution, i.payload, i.createdByTaskID, i.updatedByTaskID, "+
			"i.createdBy, i.editedBy, pp.title AS ppmTitle, "+
			"t.planID, p.name AS planName, i.createdDate, i.resolvedDate, i.closedDate, i.ignoredDate, "+
			"t.triggerID, t.triggerType").
		Joins("LEFT JOIN ops_scan_rules AS r ON i.ruleID = r.id").
		Joins("LEFT JOIN ops_scan_tasks AS t ON i.updatedByTaskID = t.id").
		Joins("LEFT JOIN ops_scan_plans AS p ON t.planID = p.id").
		Joins("LEFT JOIN ops_ppm AS pp ON i.ppmID = pp.id").
		Where("i.deleted = ?", 0)

	if filter.ID != 0 {
		stmt = stmt.Where("i.id = ?", filter.ID)
	}
	if filter.RepoID != 0 {
		stmt = stmt.Where("i.repoID = ?", filter.RepoID)
	}
	if filter.PpmID != 0 {
		stmt = stmt.Where("i.ppmID = ?", filter.PpmID)
	}
	if filter.Branch != "" {
		stmt = stmt.Where("i.repoBranch = ?", filter.Branch)
	}
	if filter.RuleID != 0 {
		stmt = stmt.Where("i.ruleID = ?", filter.RuleID)
	}
	if filter.Status != "" {
		stmt = stmt.Where("i.status = ?", filter.Status)
	}
	if filter.ScanMethods != nil {
		stmt = stmt.Where("i.scanMethod IN ?", filter.ScanMethods)
	}
	if filter.Path != "" {
		stmt = stmt.Where("i.path = ?", filter.Path)
	}
	if filter.Title != "" {
		stmt = stmt.Where("i.title LIKE ?", "%"+filter.Title+"%")
	}
	if filter.Content != "" {
		stmt = stmt.Where("i.content LIKE ?", "%"+filter.Content+"%")
	}
	if filter.RulePriority != "" {
		stmt = stmt.Where("r.priority IN ?", strings.Split(filter.RulePriority, ","))
	}
	if filter.RuleType != "" {
		stmt = stmt.Where("r.type = ?", filter.RuleType)
	}
	if filter.Category != "" {
		stmt = stmt.Where("i.category = ?", filter.Category)
	}
	if filter.Severity != "" {
		stmt = stmt.Where("i.severity = ?", filter.Severity)
	}
	if filter.Tool != "" {
		stmt = stmt.Where("r.plugin = ?", filter.Tool)
	}
	if filter.PlanID != 0 {
		stmt = stmt.Where("t.planID = ?", filter.PlanID)
	}
	if filter.CreateDate != "" {
		stmt = stmt.Where("DATE(i.createdDate) = ?", filter.CreateDate[:10])
	}
	if filter.CreatedGt != "" {
		stmt = stmt.Where("i.createdDate >= ?", filter.CreatedGt)
	}
	if filter.CreatedLt != "" {
		stmt = stmt.Where("i.createdDate < ?", filter.CreatedLt)
	}
	if filter.UpdatedGt != "" {
		stmt = stmt.Where("i.editedDate >= ?", filter.UpdatedGt)
	}
	if filter.UpdatedLt != "" {
		stmt = stmt.Where("i.editedDate < ?", filter.UpdatedLt)
	}
	if filter.ResolvedGt != "" {
		stmt = stmt.Where("i.resolvedDate >= ?", filter.ResolvedGt)
	}
	if filter.ResolvedLt != "" {
		stmt = stmt.Where("i.resolvedDate < ?", filter.ResolvedLt)
	}
	var issues []*issueJoin
	pageResult, err := database.Paginate(ctx, stmt, database.PageParam{
		Page:      filter.Page,
		PageSize:  filter.PageSize,
		Sort:      normalizeIssueSortField(filter.Sort),
		Order:     normalizeIssueSortOrder(filter.Order),
		NeedTotal: true,
	}, &issues)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	data := make([]*scanModel.ScanIssueInfo, 0, len(pageResult.Data))
	for _, item := range pageResult.Data {
		data = append(data, mapIssueJoinToModel(item))
	}
	return &database.PageResult[*scanModel.ScanIssueInfo]{
		Data:     data,
		Total:    pageResult.Total,
		Page:     pageResult.Page,
		PageSize: pageResult.PageSize,
	}, nil
}

// ListByIDs 根据 ID 列表查询扫描问题详情
func (s *ScanIssueStore) ListByIDs(ctx context.Context, ids []int64) ([]*scanModel.ScanIssueInfo, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var issues []*issueJoin
	if err := s.db.WithContext(ctx).
		Table("ops_scan_issues AS i").
		Select("i.id, i.issueKey, i.ruleID, r.name AS ruleName, r.ruleKey AS ruleKey, "+
			"r.priority AS rulePriority, r.type AS ruleType, "+
			"i.title, i.content, i.path, i.startLine, i.endLine, i.oldCode, i.newCode, i.category, i.severity, "+
			"i.ppmID, i.pipelineExecutionID, i.repoID, i.repoBranch, "+
			"i.status, i.scanMethod, i.resolution, i.payload, i.createdByTaskID, i.updatedByTaskID, "+
			"i.createdBy, i.editedBy, pp.title AS ppmTitle, i.resolvedDate, i.closedDate, i.ignoredDate, "+
			"t.triggerID, t.triggerType").
		Joins("LEFT JOIN ops_scan_rules AS r ON i.ruleID = r.id").
		Joins("LEFT JOIN ops_scan_tasks AS t ON i.updatedByTaskID = t.id").
		Joins("LEFT JOIN ops_ppm AS pp ON i.ppmID = pp.id").
		Where("i.id IN ? AND i.deleted = ?", ids, 0).
		Find(&issues).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	data := make([]*scanModel.ScanIssueInfo, 0, len(issues))
	for _, item := range issues {
		data = append(data, mapIssueJoinToModel(item))
	}
	return data, nil
}

// UpdateResolution 批量更新 issue resolution
func (s *ScanIssueStore) UpdateResolution(ctx context.Context, ids []int64, resolution *constant.ScanIssueResolution, operator string) error {
	if len(ids) == 0 {
		return nil
	}
	// resolution 为 nil 表示清空解决方案，写入空串而非 NULL（列 NOT NULL）
	updates := map[string]any{"resolution": constant.ScanIssueResolutionNull}
	if resolution != nil {
		updates["resolution"] = resolution
	}
	if operator != "" {
		updates["editedBy"] = operator
	}
	if err := s.db.WithContext(ctx).Model(&ScanIssue{}).
		Where("id IN ? AND deleted = ?", ids, 0).
		Updates(updates).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// UpdateStatus 批量更新 issue 状态
func (s *ScanIssueStore) UpdateStatus(ctx context.Context, ids []int64, status constant.ScanIssueStatus, operator string) error {
	if len(ids) == 0 {
		return nil
	}
	updates := map[string]any{"status": status}
	if operator != "" {
		updates["editedBy"] = operator
	}
	if err := s.db.WithContext(ctx).Model(&ScanIssue{}).
		Where("id IN ? AND deleted = ?", ids, 0).
		Updates(updates).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// UpdateTimestamps 按状态写对应时间戳字段
func (s *ScanIssueStore) UpdateTimestamps(ctx context.Context, in scanModel.ScanIssueUpdateStatusReq, operator string) error {
	if len(in.IDs) == 0 {
		return nil
	}
	fields := map[string]any{}
	if in.Status != nil {
		switch *in.Status {
		case constant.ScanIssueStatusSolved:
			fields["resolvedDate"] = in.Resolved
		case constant.ScanIssueStatusClosed:
			fields["closedDate"] = datetime.NowLocalTime()
		case constant.ScanIssueStatusIgnore:
			fields["ignoredDate"] = in.Ignored
		}
	}
	if len(fields) == 0 {
		return nil
	}
	if operator != "" {
		fields["editedBy"] = operator
	}
	if err := s.db.WithContext(ctx).Model(&ScanIssue{}).
		Where("id IN ? AND deleted = ?", in.IDs, 0).
		Updates(fields).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// ResetExpiredIgnore 将过期的 ignore 问题重置为 wait 状态，清空 resolution 和 ignoredDate
func (s *ScanIssueStore) ResetExpiredIgnore(ctx context.Context, ids []int64, operator string) error {
	if len(ids) == 0 {
		return nil
	}
	updates := map[string]any{
		"status":      constant.ScanIssueStatusWait,
		"resolution":  constant.ScanIssueResolutionNull,
		"ignoredDate": int64(0),
	}
	if operator != "" {
		updates["editedBy"] = operator
	}
	if err := s.db.WithContext(ctx).Model(&ScanIssue{}).
		Where("id IN ? AND deleted = ?", ids, 0).
		Updates(updates).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (s *ScanIssueStore) ListKeysByTask(
	ctx context.Context, repoID int64, branch string, taskID uint, method constant.ScanMethod,
) ([]*scanModel.ScanIssueKey, error) {
	var issues []*ScanIssue
	if err := s.db.WithContext(ctx).
		Select("id, issueKey, createdByTaskID, updatedByTaskID, status, ignoredDate").
		Where("repoID = ? AND repoBranch = ? AND scanMethod = ? AND deleted = ?", repoID, branch, method, 0).
		Find(&issues).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	result := make([]*scanModel.ScanIssueKey, 0, len(issues))
	for _, item := range issues {
		result = append(result, &scanModel.ScanIssueKey{
			ID:              int64(item.ID),
			Key:             item.Key,
			CreatedByTaskID: item.CreatedByTaskID,
			UpdatedByTaskID: item.UpdatedByTaskID,
			Status:          item.Status,
			IgnoredDate:     item.IgnoredDate,
		})
	}
	return result, nil
}

// ListExistingKeys 查询指定 issueKey 集合中已存在的 key
func (s *ScanIssueStore) ListExistingKeys(ctx context.Context, keys []string) ([]string, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	var issueKeys []string
	if err := s.db.WithContext(ctx).
		Model(&ScanIssue{}).
		Where("issueKey IN ?", keys).
		Pluck("issueKey", &issueKeys).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return issueKeys, nil
}

// UpdateByTaskID 批量更新 issue 的 updatedByTaskID
func (s *ScanIssueStore) UpdateByTaskID(ctx context.Context, taskID uint, issueIDs []int64, operator string) error {
	if len(issueIDs) == 0 {
		return nil
	}
	updates := map[string]any{"updatedByTaskID": taskID}
	if operator != "" {
		updates["editedBy"] = operator
	}
	if err := s.db.WithContext(ctx).Model(&ScanIssue{}).
		Where("id IN ?", issueIDs).
		Updates(updates).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// DisappearedByTask 标记 issue 为已消失
func (s *ScanIssueStore) DisappearedByTask(ctx context.Context, taskID uint, issueIDs []int64, operator string) error {
	if len(issueIDs) == 0 {
		return nil
	}
	updates := map[string]any{
		"updatedByTaskID": taskID,
		"status":          constant.ScanIssueStatusClosed,
	}
	if operator != "" {
		updates["editedBy"] = operator
	}
	if err := s.db.WithContext(ctx).Model(&ScanIssue{}).
		Where("id IN ? AND deleted = ?", issueIDs, 0).
		Updates(updates).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// CreateIssueBindByTaskID 批量创建 issue 与 task 的绑定关系
func (s *ScanIssueStore) CreateIssueBindByTaskID(ctx context.Context, taskID uint, issueKeys []string) error {
	if len(issueKeys) == 0 {
		return nil
	}

	// 先根据 issueKeys 查出对应的 issue IDs
	var issues []*ScanIssue
	if err := s.db.WithContext(ctx).
		Select("id").
		Where("issueKey IN ? AND deleted = ?", issueKeys, 0).
		Find(&issues).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}

	if len(issues) == 0 {
		return nil
	}

	binds := make([]*ScanIssueTaskBind, 0, len(issues))
	for _, issue := range issues {
		binds = append(binds, &ScanIssueTaskBind{
			TaskID:  int64(taskID),
			IssueID: int64(issue.ID),
		})
	}

	if err := s.db.WithContext(ctx).Create(&binds).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// ListByTaskID 查询指定任务关联的 issue 列表（含规则信息，支持分页）
func (s *ScanIssueStore) ListByTaskID(ctx context.Context, taskID uint, filter scanModel.ScanTaskIssueReqFilter) (*database.PageResult[*scanModel.ScanIssueInfo], error) {
	stmt := s.db.WithContext(ctx).
		Table("ops_scan_issue_task_binds AS bind").
		Select("i.id, i.issueKey, i.ruleID, r.name AS ruleName, r.ruleKey AS ruleKey, "+
			"r.priority AS rulePriority, r.type AS ruleType, "+
			"i.title, i.content, i.path,  i.startLine, i.endLine, i.oldCode, i.newCode, i.category, i.severity, "+
			"i.ppmID, i.pipelineExecutionID, i.repoID, i.repoBranch, "+
			"i.status, i.scanMethod, i.resolution, i.payload, i.createdByTaskID, i.updatedByTaskID, "+
			"i.createdBy, i.editedBy, pp.title AS ppmTitle, "+
			"t.planID, p.name AS planName, i.createdDate, i.resolvedDate, i.closedDate, i.ignoredDate, "+
			"t.triggerID, t.triggerType").
		Joins("INNER JOIN ops_scan_issues AS i ON bind.issueID = i.id").
		Joins("INNER JOIN ops_scan_rules AS r ON i.ruleID = r.id").
		Joins("LEFT JOIN ops_scan_tasks AS t ON i.updatedByTaskID = t.id").
		Joins("LEFT JOIN ops_scan_plans AS p ON t.planID = p.id").
		Joins("LEFT JOIN ops_ppm AS pp ON i.ppmID = pp.id").
		Where("bind.taskID = ? AND i.deleted = ?", taskID, 0)

	if filter.ID != 0 {
		stmt = stmt.Where("i.id = ?", filter.ID)
	}
	if filter.RuleID != 0 {
		stmt = stmt.Where("i.ruleID = ?", filter.RuleID)
	}
	if filter.PpmID != 0 {
		stmt = stmt.Where("i.ppmID = ?", filter.PpmID)
	}
	if filter.Status != "" {
		stmt = stmt.Where("i.status = ?", filter.Status)
	}
	if filter.ScanMethods != nil {
		stmt = stmt.Where("i.scanMethod IN ?", filter.ScanMethods)
	}
	if filter.Path != "" {
		stmt = stmt.Where("i.path = ?", filter.Path)
	}
	if filter.Title != "" {
		stmt = stmt.Where("i.title LIKE ?", "%"+filter.Title+"%")
	}
	if filter.Content != "" {
		stmt = stmt.Where("i.content LIKE ?", "%"+filter.Content+"%")
	}
	if filter.Category != "" {
		stmt = stmt.Where("i.category = ?", filter.Category)
	}
	if filter.Severity != "" {
		stmt = stmt.Where("i.severity = ?", filter.Severity)
	}
	if filter.RulePriority != "" {
		stmt = stmt.Where("r.priority IN ?", strings.Split(filter.RulePriority, ","))
	}
	if filter.RuleType != "" {
		stmt = stmt.Where("r.type = ?", filter.RuleType)
	}
	if filter.Tool != "" {
		stmt = stmt.Where("r.plugin = ?", filter.Tool)
	}
	if filter.PlanID != 0 {
		stmt = stmt.Where("t.planID = ?", filter.PlanID)
	}
	if filter.CreateDate != "" {
		stmt = stmt.Where("DATE(i.createdDate) = ?", filter.CreateDate[:10])
	}
	if filter.DetectionStatus == "new" {
		stmt = stmt.Where("i.createdByTaskID = ?", taskID)
	}

	var issues []*issueJoin
	pageResult, err := database.Paginate(ctx, stmt, database.PageParam{
		Page:      filter.Page,
		PageSize:  filter.PageSize,
		Sort:      normalizeIssueSortField(filter.Sort),
		Order:     normalizeIssueSortOrder(filter.Order),
		NeedTotal: true,
	}, &issues)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	data := make([]*scanModel.ScanIssueInfo, 0, len(pageResult.Data))
	for _, item := range pageResult.Data {
		data = append(data, mapIssueJoinToModel(item))
	}

	return &database.PageResult[*scanModel.ScanIssueInfo]{
		Data:     data,
		Total:    pageResult.Total,
		Page:     pageResult.Page,
		PageSize: pageResult.PageSize,
	}, nil
}

// ListByTaskForResult 查询任务关联的 issue 对应的规则信息（用于质量门禁判定）
func (s *ScanIssueStore) ListByTaskForResult(ctx context.Context, taskID uint, planID int64) ([]*scanModel.Rule, error) {
	var rules []*scanModel.Rule
	if err := s.db.WithContext(ctx).
		Table("ops_scan_issue_task_binds AS bind").
		Select("r.id, r.ruleKey, r.priority, r.type").
		Joins("INNER JOIN ops_scan_issues AS i ON bind.issueID = i.id").
		Joins("INNER JOIN ops_scan_rules AS r ON i.ruleID = r.id").
		Where("bind.taskID = ? AND i.deleted = ?", taskID, 0).
		Find(&rules).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return rules, nil
}
