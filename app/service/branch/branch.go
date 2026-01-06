package branch

import (
	"context"
	"fmt"

	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	branchModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/utils"
	branchStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	gitUtil "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/util"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
)

type Service struct {
	urlProvider url.Provider
	gitPkg      git.Interface
	branchStore branchStore.Interface
}

var _ Interface = (*Service)(nil)

// NewService 创建 BranchSrv 实例
func NewService(urlProvider url.Provider, git git.Interface, branchStore branchStore.Interface) *Service {
	return &Service{
		urlProvider: urlProvider,
		gitPkg:      git,
		branchStore: branchStore,
	}
}

func (s *Service) Create(
	ctx context.Context,
	req branchModel.ReqForCreate,
	session *authModel.Session,
	repo *repoModel.Repo,
) (*branchModel.Branch, error) {

	// 准备 RPC 参数
	writeParams, err := utils.CreateRPCInternalWriteParams(ctx, s.urlProvider, session, repo)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitCreateExecParamError)
		return nil, err
	}

	// 调用 Git 层创建分支
	rpcOut, err := s.gitPkg.CreateBranch(ctx, &git.CreateBranchParams{
		WriteParams: writeParams,
		BranchName:  req.Name,
		Target:      req.Source,
	})
	if err != nil {
		return &branchModel.Branch{}, gitUtil.ProcessGitErrorf(ctx, err)
	}

	// 将返回结果映射为前端对象
	data, err := utils.MapBranch(rpcOut.Branch)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitMapOutputError)
		return &branchModel.Branch{}, err
	}
	return &data, nil
}

func (s *Service) List(ctx context.Context, filter branchModel.ReqFilter, repo *repoModel.Repo) ([]branchModel.Branch, bool, error) {

	// 准备 RPC 参数
	rpcOut, isLast, err := s.gitPkg.ListBranchesV2(ctx, &git.ListBranchesParams{
		ReadParams:      git.ReadParams{RepoUID: repo.GitUID},
		IncludeCommit:   filter.IncludeCommit,
		Query:           filter.Query,
		Label:           filter.Prefix,
		IncludeArchived: filter.IncludeArchived,
		Sort:            mapToRPCBranchSortOption(filter.Sort),
		Order:           utils.MapToRPCSortOrder(filter.Order),
		Page:            int32(filter.Page),
		PageSize:        int32(filter.PageSize),
	})
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitCreateExecParamError)
		return nil, isLast, err
	}

	branches := rpcOut.Branches
	response := make([]branchModel.Branch, len(branches))
	for i := range branches {
		response[i], err = utils.MapBranch(branches[i])
		if err != nil {
			return nil, isLast, fmt.Errorf("failed to map branch: %w", err)
		}

		response[i].IsDefault = repo.DefaultBranch == branches[i].Name
	}

	return response, isLast, nil
}

func (s *Service) ListPrefixes(ctx context.Context, repo repoModel.Repo) ([]branchModel.Prefix, error) {
	labels, err := s.gitPkg.ListBranchLabelsV2(ctx, &git.ListBranchesParams{
		ReadParams: git.ReadParams{RepoUID: repo.GitUID},
	})
	if err != nil {
		err = errors.WithI18nKey(err, i18n.BranchListPrefixError)
		return nil, err
	}

	outLabels := make([]branchModel.Prefix, len(labels))
	for i, label := range labels {
		outLabels[i] = branchModel.Prefix{Name: label.Name, Total: label.Total}
	}
	return outLabels, nil
}

func (s *Service) Delete(
	ctx context.Context,
	name string,
	session *authModel.Session,
	repo *repoModel.Repo,
) error {
	// 准备 RPC 参数
	writeParams, err := utils.CreateRPCInternalWriteParams(ctx, s.urlProvider, session, repo)
	if err != nil {
		return errors.WithI18nKey(err, i18n.GitCreateExecParamError)
	}

	// 获取分支信息以取得 SHA
	branchOut, err := s.gitPkg.GetBranch(ctx, &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: repo.GitUID},
		BranchName: name,
	})
	if err != nil {
		return errors.WithI18nKey(err, i18n.GitExecError)
	}

	// 调用 Git 层删除分支
	err = s.gitPkg.DeleteBranch(ctx, &git.DeleteBranchParams{
		WriteParams: writeParams,
		BranchName:  name,
		SHA:         branchOut.Branch.SHA.String(),
	})
	if err != nil {
		return errors.WithI18nKey(err, i18n.GitExecError)
	}

	return nil
}

func mapToRPCBranchSortOption(o string) git.BranchSortOption {
	switch o {
	case "date":
		return git.BranchSortOptionDate
	case "name":
		return git.BranchSortOptionName
	default:
		return git.BranchSortOptionDefault
	}
}
