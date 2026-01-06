package branch

import (
	"context"

	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	branchModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
	branchStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/slice"
)

// ***** BranchType(zt_ops_branch_type) 表在devops 库中 *****
// ***** BranchType(zt_ops_branch_type) 表 Repo 字段 存储的是 devops 库中 zt_repo 表的 id 字段*****

// BatchCreateBranchType 批量创建分支类型
func (s *Service) BatchCreateBranchType(
	ctx context.Context,
	reqs []branchModel.ReqForCreateBranchType,
	session *authModel.Session,
	repoID uint,
) ([]branchModel.BranchType, error) {
	branchTypes := make([]*branchStore.BranchType, 0, len(reqs))

	for _, req := range reqs {
		branchType := &branchStore.BranchType{
			Repo:      repoID,
			Name:      req.Name,
			Key:       req.Key,
			Prefixes:  slice.Join(req.Prefixes, ","),
			Desc:      req.Desc,
			CreatedBy: session.User.Account,
			EditedBy:  session.User.Account,
		}
		branchTypes = append(branchTypes, branchType)
	}

	results, err := s.branchStore.BatchCreateBranchType(ctx, branchTypes)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// UpdateBranchType 更新分支类型
func (s *Service) UpdateBranchType(
	ctx context.Context,
	id uint,
	req branchModel.ReqForUpdateBranchType,
	session *authModel.Session,
	repoID uint,
) (*branchModel.BranchType, error) {
	// 检查分支类型是否存在
	existing, err := s.branchStore.GetBranchTypeByID(ctx, id)
	if err != nil {
		return nil, errors.WithI18nKey(err, i18n.BranchTypeNotFound)
	}

	// 验证分支类型是否属于当前仓库
	if existing.RepoID != repoID {
		return nil, errors.WithI18nKey(errors.New("branch type not found in this repo"), i18n.BranchTypeNotFound)
	}

	// 构建更新对象
	branchType := &branchStore.BranchType{
		EditedBy: session.User.Account,
	}

	if req.Name != "" {
		branchType.Name = req.Name
	}

	if len(req.Prefixes) > 0 {
		branchType.Prefixes = slice.Join(req.Prefixes, ",")
	}

	if req.Desc != "" {
		branchType.Desc = req.Desc
	}

	result, err := s.branchStore.UpdateBranchType(ctx, branchType, id)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetBranchType 获取单个分支类型
func (s *Service) GetBranchType(
	ctx context.Context,
	id uint,
	repoID uint,
) (*branchModel.BranchType, error) {
	branchType, err := s.branchStore.GetBranchTypeByID(ctx, id)
	if err != nil {
		return nil, errors.WithI18nKey(err, i18n.BranchTypeNotFound)
	}

	// 验证分支类型是否属于当前仓库
	if branchType.RepoID != repoID {
		return nil, errors.WithI18nKey(errors.New("branch type not found in this repo"), i18n.BranchTypeNotFound)
	}

	return branchType, nil
}

// DeleteBranchType 删除分支类型
func (s *Service) DeleteBranchType(
	ctx context.Context,
	id uint,
	repoID uint,
) error {
	// 检查分支类型是否存在
	existing, err := s.branchStore.GetBranchTypeByID(ctx, id)
	if err != nil {
		return errors.WithI18nKey(err, i18n.BranchTypeNotFound)
	}

	// 验证分支类型是否属于当前仓库
	if existing.RepoID != repoID {
		return errors.WithI18nKey(errors.New("branch type not found in this repo"), i18n.BranchTypeNotFound)
	}

	err = s.branchStore.DeleteBranchType(ctx, id)
	if err != nil {
		return errors.WithI18nKey(err, i18n.BranchTypeDeleteError)
	}

	return nil
}

// ListBranchTypes 查询分支类型列表
func (s *Service) ListBranchTypes(
	ctx context.Context,
	filter branchModel.ReqFilterBranchType,
	repoID uint,
) (*database.PageResult[branchModel.BranchType], error) {
	result, err := s.branchStore.ListBranchTypes(ctx, repoID, filter)
	if err != nil {
		return nil, errors.WithI18nKey(err, i18n.BranchTypeQueryError)
	}

	return result, nil
}
