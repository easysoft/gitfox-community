package branch

import (
	"context"
	"fmt"

	branchModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/utils"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
)

type Service struct {
	urlProvider url.Provider
	gitPkg      git.Interface
}

// NewService 创建 BranchSrv 实例
func NewService(urlProvider url.Provider, git git.Interface) *Service {
	return &Service{
		urlProvider: urlProvider,
		gitPkg:      git,
	}
}

func (s *Service) Create(ctx context.Context, req branchModel.ReqForCreate) (*branchModel.RespForCreate, error) {

	// 准备 RPC 参数
	// writeParams, err := utils.CreateRPCInternalWriteParams(c, ctl.urlProvider, session, repo)
	writeParams, err := utils.CreateRPCInternalWriteParams(ctx, s.urlProvider)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitCreateExecParamError)
		return nil, err
	}

	// 调用 Git 层创建分支
	rpcOut, err := s.gitPkg.CreateBranch(ctx, &git.CreateBranchParams{
		WriteParams: writeParams,
		BranchName:  req.Name,
		Target:      req.Target,
	})
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitExecError)
		return &branchModel.RespForCreate{}, err
	}

	// 将返回结果映射为前端对象
	data, err := utils.MapBranch(rpcOut.Branch)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitMapOutputError)
		return &branchModel.RespForCreate{}, err
	}
	return &branchModel.RespForCreate{Branch: data}, nil
}

func (s *Service) List(ctx context.Context, filter branchModel.ReqFilter, repo branchModel.Branch) ([]branchModel.Branch, error) {

	// 准备 RPC 参数
	rpcOut, err := s.gitPkg.ListBranchesV2(ctx, &git.ListBranchesParams{
		// ReadParams:      git.CreateReadParams(repo),
		IncludeCommit:   filter.IncludeCommit,
		Query:           filter.Query,
		Label:           filter.Label,
		IncludeArchived: filter.IncludeArchived,
		Sort:            mapToRPCBranchSortOption(filter.Sort),
		Order:           utils.MapToRPCSortOrder(filter.Order),
		Page:            int32(filter.Page),
		PageSize:        int32(filter.Size),
	})
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitCreateExecParamError)
		return nil, err
	}
	branches := rpcOut.Branches

	response := make([]branchModel.Branch, len(branches))
	for i := range branches {
		response[i], err = utils.MapBranch(branches[i])
		if err != nil {
			return nil, fmt.Errorf("failed to map branch: %w", err)
		}

		// response[i].IsDefault = repo.DefaultBranch == branches[i].Name
	}

	return response, nil
}

func (s *Service) ListLabels(ctx context.Context, repo branchModel.Branch) ([]branchModel.Label, error) {
	labels, err := s.gitPkg.ListBranchLabelsV2(ctx, &git.ListBranchesParams{
		// ReadParams: git.CreateReadParams(repo),
	})
	if err != nil {
		return nil, fmt.Errorf("fail to get the list of branch labels from git: %w", err)
	}

	outLabels := make([]branchModel.Label, len(labels))
	for i, label := range labels {
		outLabels[i] = branchModel.Label{Name: label.Name, Total: label.Total}
	}
	return outLabels, nil
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
