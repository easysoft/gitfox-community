package pullrequest

import (
	"context"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

type Store struct {
	db *gorm.DB
}

func NewPullRequestStore(db *gorm.DB) *Store {
	return &Store{
		db: db,
	}
}

// GetByID 根据ID获取PR
func (s *Store) GetByID(ctx context.Context, targetRepoID uint, id uint) (*repoModel.PullRequest, error) {
	var pr PullRequest
	if err := s.db.WithContext(ctx).First(&pr, "targetRepoID = ? AND id = ?", targetRepoID, id).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapPullRequestToModel(pr), nil
}

// UpdateStatus 更新PR状态
func (s *Store) UpdateStatus(ctx context.Context, targetRepoID uint, id uint, status repoModel.PullRequestStatus, editedBy string) error {
	updates := map[string]interface{}{
		"status":     status,
		"editedBy":   editedBy,
		"editedDate": datetime.NowLocalTime(),
	}
	if err := s.db.WithContext(ctx).Model(&PullRequest{}).Where("targetRepoID = ? AND id = ?", targetRepoID, id).Updates(updates).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// UpdateMergeInfo 更新合并信息
func (s *Store) UpdateMergeInfo(
	ctx context.Context,
	targetRepoID uint,
	id uint,
	mergedBy string,
	mergeMethod repoModel.MergeMethod,
	mergeSHA string,
	editedBy string,
) error {
	updates := map[string]interface{}{
		"status":      repoModel.PullRequestStatusMerged,
		"mergedBy":    mergedBy,
		"mergeMethod": mergeMethod,
		"mergeSHA":    mergeSHA,
		"editedBy":    editedBy,
		"editedDate":  datetime.NowLocalTime(),
	}
	if err := s.db.WithContext(ctx).Model(&PullRequest{}).Where("targetRepoID = ? AND id = ?", targetRepoID, id).Updates(updates).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// UpdateMergeCheckStatus 更新合并检查状态
func (s *Store) UpdateMergeCheckStatus(
	ctx context.Context,
	targetRepoID uint,
	id uint,
	status repoModel.MergeCheckStatus,
	conflicts *string,
) error {
	updates := map[string]interface{}{
		"mergeCheckStatus": status,
		"mergeConflicts":   conflicts,
		"editedDate":       datetime.NowLocalTime(),
	}
	if err := s.db.WithContext(ctx).Model(&PullRequest{}).Where("targetRepoID = ? AND id = ?", targetRepoID, id).Updates(updates).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// UpdateRebaseCheckStatus 更新变基检查状态
func (s *Store) UpdateRebaseCheckStatus(
	ctx context.Context,
	targetRepoID uint,
	id uint,
	status repoModel.RebaseCheckStatus,
	conflicts *string,
) error {
	updates := map[string]interface{}{
		"rebaseCheckStatus": status,
		"rebaseConflicts":   conflicts,
		"editedDate":        datetime.NowLocalTime(),
	}
	if err := s.db.WithContext(ctx).Model(&PullRequest{}).Where("targetRepoID = ? AND id = ?", targetRepoID, id).Updates(updates).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// UpdateDiffStats 更新PR差异统计信息
func (s *Store) UpdateDiffStats(
	ctx context.Context,
	targetRepoID uint,
	id uint,
	diffStats *repoModel.DiffStats,
) error {
	updates := map[string]interface{}{
		"commitCount": diffStats.Commits,
		"fileCount":   diffStats.FilesChanged,
		"additions":   diffStats.Additions,
		"deletions":   diffStats.Deletions,
		"editedDate":  datetime.NowLocalTime(),
	}
	if err := s.db.WithContext(ctx).Model(&PullRequest{}).Where("targetRepoID = ? AND id = ?", targetRepoID, id).Updates(updates).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// CountByTargetRepoID 统计仓库的PR数量
func (s *Store) CountByTargetRepoID(ctx context.Context, targetRepoID uint) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&PullRequest{}).Where("targetRepoID = ?", targetRepoID).Count(&count).Error; err != nil {
		return 0, database.ProcessSQLErrorf(ctx, err)
	}
	return count, nil
}

// CountByStatus 按状态统计PR数量
func (s *Store) CountByStatus(ctx context.Context, targetRepoID uint, status repoModel.PullRequestStatus) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&PullRequest{}).
		Where("targetRepoID = ? AND status = ?", targetRepoID, status).
		Count(&count).Error; err != nil {
		return 0, database.ProcessSQLErrorf(ctx, err)
	}
	return count, nil
}

// mapPullRequestToModel 将 store 模型转换为 API 模型
func mapPullRequestToModel(data PullRequest) *repoModel.PullRequest {
	return &repoModel.PullRequest{
		ID:                data.ID,
		Title:             data.Title,
		Desc:              data.Desc,
		RepoID:            data.RepoID,
		SourceRepoID:      data.SourceRepoID,
		SourceBranch:      data.SourceBranch,
		SourceSHA:         data.SourceSHA,
		TargetRepoID:      data.TargetRepoID,
		TargetBranch:      data.TargetBranch,
		MergedBy:          data.MergedBy,
		MergeMethod:       repoModel.MergeMethod(data.MergeMethod),
		MergeTargetSHA:    data.MergeTargetSHA,
		MergeBaseSHA:      data.MergeBaseSHA,
		MergeSHA:          data.MergeSHA,
		MergeCheckStatus:  repoModel.MergeCheckStatus(data.MergeCheckStatus),
		MergeConflicts:    data.MergeConflicts,
		RebaseCheckStatus: repoModel.RebaseCheckStatus(data.RebaseCheckStatus),
		RebaseConflicts:   data.RebaseConflicts,
		CommitCount:       data.CommitCount,
		FileCount:         data.FileCount,
		Flow:              data.Flow,
		ReviewFlowID:      data.ReviewFlowID,
		Additions:         data.Additions,
		Deletions:         data.Deletions,
		Status:            repoModel.PullRequestStatus(data.Status),
		CreatedBy:         data.CreatedBy,
		CreatedDate:       data.CreatedDate,
		EditedBy:          data.EditedBy,
		EditedDate:        data.EditedDate,
	}
}
