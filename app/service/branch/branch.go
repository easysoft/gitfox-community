package branch

import (
	"context"
	"fmt"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	branchModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
	commitModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/utils"
	branchStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	gitUtil "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/util"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
	"github.com/gin-gonic/gin"
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

// BatchArchive 批量归档分支：为每个分支创建 archived/<name> 分支，然后删除原分支。
func (s *Service) BatchArchive(
	c *gin.Context,
	ctx context.Context,
	req branchModel.ReqForArchive,
	session *authModel.Session,
	repo *repoModel.Repo,
) []branchModel.ArchiveResult {
	results := make([]branchModel.ArchiveResult, 0, len(req.BranchNames))

	writeParams, err := utils.CreateRPCInternalWriteParams(ctx, s.urlProvider, session, repo)
	if err != nil {
		for _, name := range req.BranchNames {
			results = append(results, branchModel.ArchiveResult{
				BranchName: name, Status: false,
				Message: errors.WithI18nKey(err, i18n.GitCreateExecParamError).Error(),
			})
		}
		return results
	}

	for _, branchName := range req.BranchNames {
		result := s.archiveSingle(c, ctx, branchName, repo, writeParams)
		results = append(results, result)
	}
	return results
}

// archiveSingle 归档单个分支。
func (s *Service) archiveSingle(
	c *gin.Context,
	ctx context.Context,
	branchName string,
	repo *repoModel.Repo,
	writeParams git.WriteParams,
) branchModel.ArchiveResult {
	fail := func(msg string) branchModel.ArchiveResult {
		return branchModel.ArchiveResult{BranchName: branchName, Status: false, Message: msg}
	}

	// 1. 默认分支不能归档
	if branchName == repo.DefaultBranch {
		return fail(i18n.TGin(c, i18n.BranchArchiveIsDefault))
	}

	// 2. 已经是归档分支，跳过，直接返回成功
	if strings.HasPrefix(branchName, constant.BranchArchivedPrefix) {
		return branchModel.ArchiveResult{BranchName: branchName, Status: true, Message: i18n.TGin(c, i18n.BranchArchiveSuccess)}
	}

	archivedName := constant.BranchArchivedPrefix + branchName

	// 3. 检查归档目标分支是否已存在
	_, err := s.gitPkg.GetBranch(ctx, &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: repo.GitUID},
		BranchName: archivedName,
	})
	if err == nil {
		return fail(i18n.TGin(c, i18n.BranchArchiveTargetExists, archivedName))
	}

	// 4. 创建归档分支
	_, err = s.gitPkg.CreateBranch(ctx, &git.CreateBranchParams{
		WriteParams: writeParams,
		BranchName:  archivedName,
		Target:      branchName,
	})
	if err != nil {
		return fail(i18n.TGin(c, i18n.BranchArchiveCreateError, archivedName))
	}

	// 5. 获取原分支 SHA 用于删除
	branchOut, err := s.gitPkg.GetBranch(ctx, &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: repo.GitUID},
		BranchName: branchName,
	})
	if err != nil {
		return s.rollbackArchive(c, ctx, branchName, archivedName, repo, writeParams)
	}

	// 6. 删除原分支
	err = s.gitPkg.DeleteBranch(ctx, &git.DeleteBranchParams{
		WriteParams: writeParams,
		BranchName:  branchName,
		SHA:         branchOut.Branch.SHA.String(),
	})
	if err != nil {
		return s.rollbackArchive(c, ctx, branchName, archivedName, repo, writeParams)
	}

	return branchModel.ArchiveResult{BranchName: branchName, Status: true, Message: i18n.TGin(c, i18n.BranchArchiveSuccess)}
}

// rollbackArchive 归档失败时回滚：删除已创建的归档分支。
func (s *Service) rollbackArchive(
	c *gin.Context,
	ctx context.Context,
	originalName, archivedName string,
	repo *repoModel.Repo,
	writeParams git.WriteParams,
) branchModel.ArchiveResult {
	// 获取归档分支 SHA 用于回滚删除
	archBranch, err := s.gitPkg.GetBranch(ctx, &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: repo.GitUID},
		BranchName: archivedName,
	})
	if err != nil {
		return branchModel.ArchiveResult{
			BranchName: originalName, Status: false,
			Message: i18n.TGin(c, i18n.BranchArchiveRollbackError, originalName, archivedName),
		}
	}

	err = s.gitPkg.DeleteBranch(ctx, &git.DeleteBranchParams{
		WriteParams: writeParams,
		BranchName:  archivedName,
		SHA:         archBranch.Branch.SHA.String(),
	})
	if err != nil {
		return branchModel.ArchiveResult{
			BranchName: originalName, Status: false,
			Message: i18n.TGin(c, i18n.BranchArchiveRollbackError, originalName, archivedName),
		}
	}

	return branchModel.ArchiveResult{
		BranchName: originalName, Status: false,
		Message: i18n.TGin(c, i18n.BranchArchiveDeleteError, originalName),
	}
}

// BatchActivate 批量激活归档分支：为每个 archived/<name> 创建 <name> 分支，然后删除归档分支。
func (s *Service) BatchActivate(
	c *gin.Context,
	ctx context.Context,
	req branchModel.ReqForArchive,
	session *authModel.Session,
	repo *repoModel.Repo,
) []branchModel.ActivateResult {
	results := make([]branchModel.ActivateResult, 0, len(req.BranchNames))

	writeParams, err := utils.CreateRPCInternalWriteParams(ctx, s.urlProvider, session, repo)
	if err != nil {
		for _, name := range req.BranchNames {
			results = append(results, branchModel.ActivateResult{
				BranchName: name, Status: false,
				Message: errors.WithI18nKey(err, i18n.GitCreateExecParamError).Error(),
			})
		}
		return results
	}

	for _, branchName := range req.BranchNames {
		result := s.activateSingle(c, ctx, branchName, repo, writeParams)
		results = append(results, result)
	}
	return results
}

// activateSingle 激活单个归档分支。
func (s *Service) activateSingle(
	c *gin.Context,
	ctx context.Context,
	branchName string,
	repo *repoModel.Repo,
	writeParams git.WriteParams,
) branchModel.ActivateResult {
	fail := func(msg string) branchModel.ActivateResult {
		return branchModel.ActivateResult{BranchName: branchName, Status: false, Message: msg}
	}

	// 1. 必须以 archived/ 开头
	if !strings.HasPrefix(branchName, constant.BranchArchivedPrefix) {
		return fail(i18n.TGin(c, i18n.BranchActivateNotArchived, branchName))
	}

	restoredName := strings.TrimPrefix(branchName, constant.BranchArchivedPrefix)

	// 2. 检查激活目标分支是否已存在
	_, err := s.gitPkg.GetBranch(ctx, &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: repo.GitUID},
		BranchName: restoredName,
	})
	if err == nil {
		return fail(i18n.TGin(c, i18n.BranchActivateTargetExists, restoredName))
	}

	// 3. 创建还原分支
	_, err = s.gitPkg.CreateBranch(ctx, &git.CreateBranchParams{
		WriteParams: writeParams,
		BranchName:  restoredName,
		Target:      branchName,
	})
	if err != nil {
		return fail(i18n.TGin(c, i18n.BranchActivateCreateError, restoredName))
	}

	// 4. 获取归档分支 SHA 用于删除
	archBranch, err := s.gitPkg.GetBranch(ctx, &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: repo.GitUID},
		BranchName: branchName,
	})
	if err != nil {
		// 不回滚，两个分支都保留
		return fail(i18n.TGin(c, i18n.BranchActivateDeleteError, branchName))
	}

	// 5. 删除归档分支
	err = s.gitPkg.DeleteBranch(ctx, &git.DeleteBranchParams{
		WriteParams: writeParams,
		BranchName:  branchName,
		SHA:         archBranch.Branch.SHA.String(),
	})
	if err != nil {
		// 不回滚，两个分支都保留
		return fail(i18n.TGin(c, i18n.BranchActivateDeleteError, branchName))
	}

	return branchModel.ActivateResult{BranchName: branchName, Status: true, Message: i18n.TGin(c, i18n.BranchActivateSuccess)}
}
