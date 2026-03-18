package scan

import (
	"context"

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
		Where("id IN ?", issueIDs).
		Updates(map[string]interface{}{
			"updatedByTaskID": taskID,
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
