package pullrequest

import (
	"context"
	"errors"

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

// Create 创建PR
func (s *Store) Create(ctx context.Context, pr *repoModel.PullRequest) error {
	storePR := mapModelToStore(pr)
	if err := s.db.WithContext(ctx).Create(storePR).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	pr.ID = storePR.ID
	return nil
}

// GetByID 根据ID获取PR
func (s *Store) GetByID(ctx context.Context, targetRepoID uint, id uint) (*repoModel.PullRequest, error) {
	var pr PullRequest
	if err := s.db.WithContext(ctx).First(&pr, "targetRepoID = ? AND id = ?", targetRepoID, id).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapPullRequestToModel(pr), nil
}

// List 根据过滤条件查询PR列表
func (s *Store) List(ctx context.Context, filter *repoModel.PullRequestFilter) ([]*repoModel.PullRequest, error) {
	var prs []PullRequest

	query := s.db.WithContext(ctx).Model(&PullRequest{})

	if filter.RepoID != 0 {
		query = query.Where("repoID = ?", filter.RepoID)
	}
	if filter.SourceRepoID != 0 {
		query = query.Where("sourceRepoID = ?", filter.SourceRepoID)
	}
	if filter.TargetRepoID != 0 {
		query = query.Where("targetRepoID = ?", filter.TargetRepoID)
	}
	if filter.SourceBranch != "" {
		query = query.Where("sourceBranch = ?", filter.SourceBranch)
	}
	if filter.TargetBranch != "" {
		query = query.Where("targetBranch = ?", filter.TargetBranch)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.CreatedBy != "" {
		query = query.Where("createdBy = ?", filter.CreatedBy)
	}

	if filter.PageSize > 0 {
		query = query.Limit(filter.PageSize)
		if filter.Page > 0 {
			query = query.Offset(filter.Page * filter.PageSize)
		}
	}

	if err := query.Find(&prs).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	result := make([]*repoModel.PullRequest, len(prs))
	for i, pr := range prs {
		result[i] = mapPullRequestToModel(pr)
	}
	return result, nil
}

// UpdateOptLock 使用简化的更新方式更新PR（无乐观锁）
func (s *Store) UpdateOptLock(
	ctx context.Context,
	pr *repoModel.PullRequest,
	mutateFn func(pr *repoModel.PullRequest) error,
) (*repoModel.PullRequest, error) {
	dup := *pr

	if err := mutateFn(&dup); err != nil {
		return nil, err
	}

	storePR := mapModelToStore(&dup)
	storePR.EditedDate = datetime.NowLocalTime()

	if err := s.db.WithContext(ctx).Model(&PullRequest{}).
		Where("targetRepoID = ? AND id = ?", dup.TargetRepoID, dup.ID).
		Updates(storePR).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return &dup, nil
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
	status repoModel.MergeCheckStatus,
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

// ResetMergeCheckStatus resets the pull request's mergeability status to unchecked
// for all PRs where target branch points to targetBranch.
func (s *Store) ResetMergeCheckStatus(
	ctx context.Context,
	targetRepoID uint,
	targetBranch string,
) error {
	updates := map[string]interface{}{
		"editedDate":        datetime.NowLocalTime(),
		"mergeTargetSHA":    "",
		"mergeSHA":          "",
		"mergeCheckStatus":  repoModel.MergeCheckStatusUnchecked,
		"mergeConflicts":    nil,
		"rebaseCheckStatus": repoModel.MergeCheckStatusUnchecked,
		"rebaseConflicts":   nil,
		"commitCount":       nil,
		"fileCount":         nil,
	}

	err := s.db.WithContext(ctx).
		Model(&PullRequest{}).
		Where("targetRepoID = ? AND targetBranch = ? AND status NOT IN (?, ?)",
			targetRepoID, targetBranch,
			repoModel.PullRequestStatusClosed, repoModel.PullRequestStatusMerged).
		Updates(updates).Error

	if err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// UpdateDiffStats 安全的更新PR差异统计信息
func (s *Store) UpdateDiffStatsSafe(
	ctx context.Context,
	targetRepoID uint,
	pr *repoModel.PullRequest,
	diffStats *repoModel.DiffStats,
) (*repoModel.PullRequest, error) {
	err := s.db.Transaction(func(tx *gorm.DB) error {

		if pr.Status != repoModel.PullRequestStatusOpen {
			return errors.New("Pull request must be open")
		}

		updates := map[string]interface{}{
			"commitCount": diffStats.Commits,
			"fileCount":   diffStats.FilesChanged,
			"additions":   diffStats.Additions,
			"deletions":   diffStats.Deletions,
		}

		if diffStats.MergeTargetSHA != "" {
			updates["mergeTargetSHA"] = diffStats.MergeTargetSHA
		}
		if diffStats.SourceSHA != "" {
			updates["sourceSHA"] = diffStats.SourceSHA
		}
		if diffStats.MergeBaseSHA != "" {
			updates["mergeBaseSHA"] = diffStats.MergeBaseSHA
		}
		if diffStats.MergeSHA != nil {
			updates["mergeSHA"] = *diffStats.MergeSHA
		}

		if diffStats.MergeCheckStatus != "" {
			updates["mergeCheckStatus"] = diffStats.MergeCheckStatus
			updates["mergeConflicts"] = StringSlice(diffStats.MergeConflicts)
		}

		if diffStats.RebaseCheckStatus != "" {
			updates["rebaseCheckStatus"] = diffStats.RebaseCheckStatus
			updates["rebaseConflicts"] = StringSlice(diffStats.RebaseConflicts)
		}

		if err := tx.WithContext(ctx).Model(&PullRequest{}).Where("targetRepoID = ? AND id = ?", targetRepoID, pr.ID).Updates(updates).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	pr, err = s.GetByID(ctx, targetRepoID, pr.ID)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return pr, nil
}

// UpdateMergeAfter 安全的更新PR合并信息
func (s *Store) UpdateMergeAfterSafe(
	ctx context.Context,
	targetRepoID uint,
	pr *repoModel.PullRequest,
	mergeAfter *repoModel.MergeAfter,
) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {

		if pr.Status != repoModel.PullRequestStatusOpen {
			return errors.New("Pull request must be open")
		}

		updates := map[string]interface{}{
			"status":            repoModel.PullRequestStatusMerged,
			"mergeCheckStatus":  repoModel.MergeCheckStatusMergeable,
			"mergeConflicts":    nil,
			"rebaseCheckStatus": repoModel.MergeCheckStatusMergeable,
			"rebaseConflicts":   nil,

			"commitCount": mergeAfter.Commits,
			"fileCount":   mergeAfter.FilesChanged,
			"additions":   mergeAfter.Additions,
			"deletions":   mergeAfter.Deletions,

			"mergedDate":     mergeAfter.MergedDate,
			"mergedBy":       mergeAfter.MergedBy,
			"editedBy":       mergeAfter.MergedBy,
			"editedDate":     datetime.NowLocalTime(),
			"mergeMethod":    mergeAfter.MergeMethod,
			"sourceSHA":      mergeAfter.HeadHash,
			"mergeTargetSHA": mergeAfter.MergeTargetSHA,
			"mergeBaseSHA":   mergeAfter.MergeBaseSHA,
			"mergeSHA":       mergeAfter.MergeSHA,
		}
		if err := tx.WithContext(ctx).Model(&PullRequest{}).Where("targetRepoID = ? AND id = ?", targetRepoID, pr.ID).Updates(updates).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
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

// ListReviewerAccountsByRequestID 根据请求ID（ppmID）查询评审人账号列表
func (s *Store) ListReviewerAccountsByRequestID(ctx context.Context, requestID uint) ([]string, error) {
	var reviewers []RequestReviewer
	if err := s.db.WithContext(ctx).
		Model(&RequestReviewer{}).
		Select("account").
		Where("requestID = ?", requestID).
		Find(&reviewers).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	accounts := make([]string, 0, len(reviewers))
	for _, r := range reviewers {
		if r.Account != "" {
			accounts = append(accounts, r.Account)
		}
	}
	return accounts, nil
}

// mapPullRequestToModel 将 store 模型转换为 API 模型
func mapPullRequestToModel(data PullRequest) *repoModel.PullRequest {
	return &repoModel.PullRequest{
		ID:                data.ID,
		Title:             data.Title,
		Desc:              data.Desc.String,
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
		MergeConflicts:    []string(data.MergeConflicts),
		RebaseCheckStatus: repoModel.MergeCheckStatus(data.RebaseCheckStatus),
		RebaseConflicts:   []string(data.RebaseConflicts),
		CommitCount:       data.CommitCount,
		FileCount:         data.FileCount,
		Flow:              repoModel.PullRequestFlow(data.Flow),
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

// mapModelToStore 将 API 模型转换为 store 模型
func mapModelToStore(data *repoModel.PullRequest) *PullRequest {
	return &PullRequest{
		ID:                data.ID,
		Title:             data.Title,
		Desc:              database.NullIfEmpty(data.Desc),
		RepoID:            data.RepoID,
		SourceRepoID:      data.SourceRepoID,
		SourceBranch:      data.SourceBranch,
		SourceSHA:         data.SourceSHA,
		TargetRepoID:      data.TargetRepoID,
		TargetBranch:      data.TargetBranch,
		MergedBy:          data.MergedBy,
		MergeMethod:       string(data.MergeMethod),
		MergeTargetSHA:    data.MergeTargetSHA,
		MergeBaseSHA:      data.MergeBaseSHA,
		MergeSHA:          data.MergeSHA,
		MergeCheckStatus:  string(data.MergeCheckStatus),
		MergeConflicts:    StringSlice(data.MergeConflicts),
		RebaseCheckStatus: string(data.RebaseCheckStatus),
		RebaseConflicts:   StringSlice(data.RebaseConflicts),
		CommitCount:       data.CommitCount,
		FileCount:         data.FileCount,
		Flow:              int8(data.Flow),
		ReviewFlowID:      data.ReviewFlowID,
		Additions:         data.Additions,
		Deletions:         data.Deletions,
		Status:            string(data.Status),
		CreatedBy:         data.CreatedBy,
		CreatedDate:       data.CreatedDate,
		EditedBy:          data.EditedBy,
		EditedDate:        data.EditedDate,
	}
}
