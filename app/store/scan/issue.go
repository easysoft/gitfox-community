package scan

import (
	"context"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

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
			"i.message, i.path, i.line, i.repoID, i.repoBranch, "+
			"i.status, i.scanMethod, i.resolution, i.payload, i.createdByTaskID, i.updatedByTaskID, t.planID, p.name AS planName, i.createdDate").
		Joins("INNER JOIN ops_scan_rules AS r ON i.ruleID = r.id").
		Joins("LEFT JOIN ops_scan_tasks AS t ON i.updatedByTaskID = t.id").
		Joins("LEFT JOIN ops_scan_plans AS p ON t.planID = p.id").
		Where("i.id = ? AND i.deleted = ?", id, false).
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
			"i.message, i.path, i.line, i.repoID, i.repoBranch, "+
			"i.status, i.scanMethod, i.resolution, i.payload, i.createdByTaskID, i.updatedByTaskID, t.planID, p.name AS planName, i.createdDate").
		Joins("INNER JOIN ops_scan_rules AS r ON i.ruleID = r.id").
		Joins("LEFT JOIN ops_scan_tasks AS t ON i.updatedByTaskID = t.id").
		Joins("LEFT JOIN ops_scan_plans AS p ON t.planID = p.id").
		Where("i.deleted = ?", false)

	if filter.ID != 0 {
		stmt = stmt.Where("i.id = ?", filter.ID)
	}
	if filter.RepoID != 0 {
		stmt = stmt.Where("i.repoID = ?", filter.RepoID)
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
	if filter.ScanMethod != "" {
		stmt = stmt.Where("i.scanMethod = ?", filter.ScanMethod)
	}
	if filter.File != "" {
		stmt = stmt.Where("i.path = ?", filter.File)
	}
	if filter.Message != "" {
		stmt = stmt.Where("i.message LIKE ?", "%"+filter.Message+"%")
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

	var issues []*issueJoin
	pageResult, err := database.Paginate(ctx, stmt, database.PageParam{
		Page:      filter.Page,
		PageSize:  filter.PageSize,
		Sort:      "i.id",
		Order:     "desc",
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
			"i.message, i.path, i.line, i.repoID, i.repoBranch, "+
			"i.status, i.scanMethod, i.resolution, i.payload, i.createdByTaskID, i.updatedByTaskID").
		Joins("INNER JOIN ops_scan_rules AS r ON i.ruleID = r.id").
		Where("i.id IN ? AND i.deleted = ?", ids, false).
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
func (s *ScanIssueStore) UpdateResolution(ctx context.Context, ids []int64, resolution *constant.ScanIssueResolution) error {
	if len(ids) == 0 {
		return nil
	}
	if err := s.db.WithContext(ctx).Model(&ScanIssue{}).
		Where("id IN ? AND deleted = ?", ids, false).
		Updates(map[string]any{"resolution": resolution}).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// UpdateStatus 批量更新 issue 状态
func (s *ScanIssueStore) UpdateStatus(ctx context.Context, ids []int64, status constant.ScanIssueStatus) error {
	if len(ids) == 0 {
		return nil
	}
	if err := s.db.WithContext(ctx).Model(&ScanIssue{}).
		Where("id IN ? AND deleted = ?", ids, false).
		Update("status", status).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// ListKeysByTask 查询指定仓库+分支+方法下已有的 issue keys
func (s *ScanIssueStore) ListKeysByTask(
	ctx context.Context, repoID int64, branch string, taskID uint, method constant.ScanMethod,
) ([]*scanModel.ScanIssueKey, error) {
	var issues []*ScanIssue
	if err := s.db.WithContext(ctx).
		Select("id, issueKey, createdByTaskID, updatedByTaskID, status").
		Where("repoID = ? AND repoBranch = ? AND scanMethod = ? AND deleted = ?", repoID, branch, method, false).
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
		})
	}
	return result, nil
}

// UpdateByTaskID 批量更新 issue 的 updatedByTaskID
func (s *ScanIssueStore) UpdateByTaskID(ctx context.Context, taskID uint, issueIDs []int64) error {
	if len(issueIDs) == 0 {
		return nil
	}
	if err := s.db.WithContext(ctx).Model(&ScanIssue{}).
		Where("id IN ?", issueIDs).
		Update("updatedByTaskID", taskID).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// DisappearedByTask 标记 issue 为已消失
func (s *ScanIssueStore) DisappearedByTask(ctx context.Context, taskID uint, issueIDs []int64) error {
	if len(issueIDs) == 0 {
		return nil
	}
	if err := s.db.WithContext(ctx).Model(&ScanIssue{}).
		Where("id IN ? AND deleted = ?", issueIDs, false).
		Updates(map[string]any{
			"updatedByTaskID": taskID,
			"status":          constant.ScanIssueStatusClosed,
		}).Error; err != nil {
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
		Where("issueKey IN ? AND deleted = ?", issueKeys, false).
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
			"i.message, i.path, i.line, i.repoID, i.repoBranch, "+
			"i.status, i.scanMethod, i.resolution, i.payload, i.createdByTaskID, i.updatedByTaskID").
		Joins("INNER JOIN ops_scan_issues AS i ON bind.issueID = i.id").
		Joins("INNER JOIN ops_scan_rules AS r ON i.ruleID = r.id").
		Where("bind.taskID = ? AND i.deleted = ?", taskID, false)

	if filter.ID != 0 {
		stmt = stmt.Where("i.id = ?", filter.ID)
	}
	if filter.RuleID != 0 {
		stmt = stmt.Where("i.ruleID = ?", filter.RuleID)
	}
	if filter.Status != "" {
		stmt = stmt.Where("i.status = ?", filter.Status)
	}
	if filter.ScanMethod != "" {
		stmt = stmt.Where("i.scanMethod = ?", filter.ScanMethod)
	}
	if filter.File != "" {
		stmt = stmt.Where("i.path = ?", filter.File)
	}
	if filter.Message != "" {
		stmt = stmt.Where("i.message LIKE ?", "%"+filter.Message+"%")
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
		Sort:      "i.id",
		Order:     "desc",
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
		Where("bind.taskID = ? AND i.deleted = ?", taskID, false).
		Find(&rules).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return rules, nil
}
