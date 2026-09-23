// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package branch

import (
	"context"
	"fmt"
	"strings"

	"github.com/easysoft/gitfox-community/app/constant"
	authModel "github.com/easysoft/gitfox-community/app/model/auth"
	branchModel "github.com/easysoft/gitfox-community/app/model/branch"
	commitModel "github.com/easysoft/gitfox-community/app/model/commit"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	"github.com/easysoft/gitfox-community/app/service/utils"
	branchStore "github.com/easysoft/gitfox-community/app/store/branch"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/git"
	gitUtil "github.com/easysoft/gitfox-community/pkg/git/util"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/easysoft/gitfox-community/pkg/utils/url"
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
	req.Name = utils.DecodeGitRef(req.Name)
	req.Source = utils.DecodeGitRef(req.Source)

	// 禁止创建 archived/ 前缀的分支
	if strings.HasPrefix(req.Name, constant.BranchArchivedPrefix) {
		return nil, errors.WithI18nKey(
			fmt.Errorf("branch name cannot start with %s", constant.BranchArchivedPrefix),
			i18n.BranchArchivedPrefixForbidden,
		)
	}

	// 准备 RPC 参数
	writeParams, err := utils.CreateRPCInternalWriteParams(ctx, s.urlProvider, session, repo)
	if err != nil {
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

	// ArchivedOnly 模式下不按类型/前缀过滤
	if filter.ArchivedOnly {
		filter.Prefix = ""
		filter.BranchTypeIDs = nil
	}

	// 构建 Labels 列表
	var labels []string
	if filter.Prefix != "" {
		labels = append(labels, filter.Prefix)
	}

	// 如果传入了 BranchTypeIDs，查询对应的前缀并追加到 labels
	if len(filter.BranchTypeIDs) > 0 {
		branchTypes, err := s.branchStore.GetBranchTypesByIDs(ctx, repo.ID, filter.BranchTypeIDs)
		if err != nil {
			return nil, false, errors.WithI18nKey(err, i18n.BranchTypeQueryError)
		}
		if len(branchTypes) != len(filter.BranchTypeIDs) {
			return nil, false, errors.WithI18nKey(
				fmt.Errorf("some branch type IDs not found"),
				i18n.BranchTypeNotFound,
			)
		}
		for _, bt := range branchTypes {
			labels = append(labels, bt.Prefixes...)
		}
	}

	// 准备 RPC 参数
	rpcOut, isLast, err := s.gitPkg.ListBranchesV2(ctx, &git.ListBranchesParams{
		ReadParams:      git.ReadParams{RepoUID: repo.GitUID},
		IncludeCommit:   filter.IncludeCommit,
		Query:           filter.Query,
		Labels:          labels,
		IncludeArchived: filter.IncludeArchived,
		ArchivedOnly:    filter.ArchivedOnly,
		Sort:            mapToRPCBranchSortOption(filter.Sort),
		Order:           utils.MapToRPCSortOrder(filter.Order),
		Page:            int32(filter.Page),
		PageSize:        int32(filter.PageSize),
	})
	if err != nil {
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

	// 计算分支与默认分支的 divergence（ahead/behind 提交数），默认分支自身不参与计算
	if filter.MaxDivergence > 0 && len(response) > 0 {
		divRequests := make([]git.CommitDivergenceRequest, 0, len(response))
		// 记录非默认分支在 response 中的索引，用于回填结果
		nonDefaultIdx := make([]int, 0, len(response))
		for i, b := range response {
			if b.IsDefault {
				continue
			}
			nonDefaultIdx = append(nonDefaultIdx, i)
			divRequests = append(divRequests, git.CommitDivergenceRequest{
				From: b.Name,
				To:   repo.DefaultBranch,
			})
		}

		if len(divRequests) > 0 {
			divOut, divErr := s.gitPkg.GetCommitDivergences(ctx, &git.GetCommitDivergencesParams{
				ReadParams: git.ReadParams{RepoUID: repo.GitUID},
				MaxCount:   int32(filter.MaxDivergence),
				Requests:   divRequests,
			})
			if divErr == nil {
				// divRequests 与 nonDefaultIdx 一一对应，直接按索引回填
				for j, idx := range nonDefaultIdx {
					if j < len(divOut.Divergences) {
						response[idx].CommitDivergence = &commitModel.CommitDivergence{
							Ahead:  divOut.Divergences[j].Ahead,
							Behind: divOut.Divergences[j].Behind,
						}
					}
				}
			}
		}
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
	name = utils.DecodeGitRef(name)

	// 准备 RPC 参数
	writeParams, err := utils.CreateRPCInternalWriteParams(ctx, s.urlProvider, session, repo)
	if err != nil {
		return err
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

