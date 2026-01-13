package pullrequest

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/utils"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/pullrequest"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/api"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/enum"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/sha"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/util"
	gitUtil "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/util"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
)

type Service struct {
	gitSvc           git.Interface
	repoSvc          repo.Interface
	pullRequestStore pullrequest.Interface
	urlProvider      url.Provider
}

var _ Interface = (*Service)(nil)

// NewPullRequestSvc 创建 PullRequestService 实例
func NewPullRequestSvc(
	gitSvc git.Interface,
	repoSvc repo.Interface,
	pullRequestStore pullrequest.Interface,
	urlProvider url.Provider,
) *Service {
	return &Service{
		gitSvc:           gitSvc,
		repoSvc:          repoSvc,
		pullRequestStore: pullRequestStore,
		urlProvider:      urlProvider,
	}
}

func (s *Service) GetMetadata(ctx context.Context, session *auth.Session, repoid uint, pullreqnumber uint) (*repoModel.DiffStats, error) {
	// find pull request by number
	pr, err := s.pullRequestStore.GetByID(ctx, repoid, pullreqnumber)
	if err != nil {
		return nil, err
	}

	repo, err := s.repoSvc.GetRepoCheckAccessByID(ctx, session, repoid, []constant.Permission{constant.PermissionRepoView})
	if err != nil {
		return nil, err
	}

	if pr.CommitCount != nil && pr.FileCount != nil && pr.Additions != nil && pr.Deletions != nil {
		return &repoModel.DiffStats{
			Commits:      *pr.CommitCount,
			FilesChanged: *pr.FileCount,
			Additions:    *pr.Additions,
			Deletions:    *pr.Deletions,
		}, nil
	}

	// git get diff stats
	output, err := s.gitSvc.DiffStats(ctx, &git.DiffParams{
		ReadParams: git.ReadParams{RepoUID: repo.GitUID},
		BaseRef:    pr.MergeBaseSHA,
		HeadRef:    pr.SourceSHA,
	})
	if err != nil {
		return nil, gitUtil.ProcessGitErrorf(ctx, err)
	}

	re := &repoModel.DiffStats{
		Commits:      uint(output.Commits),
		FilesChanged: uint(output.FilesChanged),
		Additions:    uint(output.Additions),
		Deletions:    uint(output.Deletions),
	}

	// updata diff stats
	err = s.pullRequestStore.UpdateDiffStatsSafe(ctx, repoid, pr, re)
	if err != nil {
		return nil, err
	}

	return re, nil
}

func (s *Service) GetDiffStats(ctx context.Context, session *auth.Session, repoid uint, pathTopath string) (*repoModel.DiffStats, error) {
	repo, err := s.repoSvc.GetRepoCheckAccessByID(ctx, session, repoid, []constant.Permission{constant.PermissionRepoView})
	if err != nil {
		return nil, err
	}

	info, err := parseDiffPath(pathTopath)
	if err != nil {
		return nil, err
	}

	// git get diff stats
	output, err := s.gitSvc.DiffStats(ctx, &git.DiffParams{
		ReadParams: git.ReadParams{RepoUID: repo.GitUID},
		BaseRef:    info.BaseRef,
		HeadRef:    info.HeadRef,
		MergeBase:  info.MergeBase,
	})
	if err != nil {
		return nil, gitUtil.ProcessGitErrorf(ctx, err)
	}

	re := &repoModel.DiffStats{
		Commits:      uint(output.Commits),
		FilesChanged: uint(output.FilesChanged),
		Additions:    uint(output.Additions),
		Deletions:    uint(output.Deletions),
	}

	return re, nil
}

func (s *Service) MergeCheck(ctx context.Context, session *auth.Session, repoid uint, pathTopath string) (*repoModel.MergeCheck, error) {
	repo, err := s.repoSvc.GetRepoCheckAccessByID(ctx, session, repoid, []constant.Permission{constant.PermissionRepoView})
	if err != nil {
		return nil, err
	}

	info, err := parseDiffPath(pathTopath)
	if err != nil {
		return nil, err
	}

	writeParams, err := utils.CreateRPCInternalWriteParams(ctx, s.urlProvider, session, repo)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitCreateExecParamError)
		return nil, err
	}

	var baseSHA, headSHA string
	baseSHAs, err := s.gitSvc.ListCommitSHAs(ctx, &git.ListCommitsParams{
		ReadParams: git.ReadParams{RepoUID: writeParams.RepoUID},
		GitREF:     info.BaseRef,
		Page:       1,
		Limit:      1,
	})
	if err != nil {
		return nil, err
	}

	if len(baseSHAs) != 0 {
		baseSHA = baseSHAs[0]
	}

	headSHAs, err := s.gitSvc.ListCommitSHAs(ctx, &git.ListCommitsParams{
		ReadParams: git.ReadParams{RepoUID: writeParams.RepoUID},
		GitREF:     info.HeadRef,
		Page:       1,
		Limit:      1,
	})
	if err != nil {
		return nil, err
	}

	if len(headSHAs) != 0 {
		headSHA = headSHAs[0]
	}

	mergeOutput, err := s.gitSvc.Merge(ctx, &git.MergeParams{
		WriteParams: writeParams,
		BaseBranch:  info.BaseRef,
		HeadRepoUID: writeParams.RepoUID, // forks are not supported for now
		HeadBranch:  info.HeadRef,
	})
	if err != nil {
		data := &repoModel.MergeCheck{
			SourceSHA: headSHA,
			TargetSHA: baseSHA,
		}
		// git.Merge works with commits and error is not user-friendly
		// and here we are modify base-ref and head-ref with user input
		// values.
		// 当尝试合并两个没有共同提交历史的分支（例如它们是作为独立仓库创建的）时，就会发生这种错误
		if uErr := api.AsUnrelatedHistoriesError(err); uErr != nil {
			uErr.BaseRef = info.BaseRef
			uErr.HeadRef = info.HeadRef
			return data, uErr
		}
		if errors.GetI18nKey(err) == i18n.GitMergeCheckNoCommits {
			return data, err
		}
		return data, gitUtil.ProcessGitErrorf(ctx, err)
	}

	if len(mergeOutput.ConflictFiles) > 0 {
		return &repoModel.MergeCheck{
			Mergeable:     false,
			ConflictFiles: mergeOutput.ConflictFiles,
			SourceSHA:     headSHA,
			TargetSHA:     baseSHA,
		}, nil
	}

	return &repoModel.MergeCheck{
		Mergeable: true,
		SourceSHA: headSHA,
		TargetSHA: baseSHA,
	}, nil
}

func (s *Service) Diff(ctx context.Context, session *auth.Session, repoid uint, pullreqnumber uint, setSHAs func(sourceSHA, mergeBaseSHA string), includePatch bool, param ...repoModel.FileDiff) (resp.Stream[*git.FileDiff], error) {
	pr, err := s.pullRequestStore.GetByID(ctx, repoid, pullreqnumber)
	if err != nil {
		return nil, err
	}

	repo, err := s.repoSvc.GetRepoCheckAccessByID(ctx, session, repoid, []constant.Permission{constant.PermissionRepoView})
	if err != nil {
		return nil, err
	}

	if setSHAs != nil {
		setSHAs(pr.SourceSHA, pr.MergeBaseSHA)
	}

	var files []api.FileDiffRequest
	for _, f := range param {
		files = append(files, api.FileDiffRequest{
			Path:      f.Path,
			StartLine: f.StartLine,
		})
	}

	reader := git.NewStreamReader(s.gitSvc.Diff(ctx, &git.DiffParams{
		ReadParams:   git.ReadParams{RepoUID: repo.GitUID},
		BaseRef:      pr.MergeBaseSHA,
		HeadRef:      pr.SourceSHA,
		MergeBase:    true,
		IncludePatch: includePatch,
	}, files...))

	return reader, nil

}

func (s *Service) RawDiff(ctx context.Context, w http.ResponseWriter, session *auth.Session, repoid uint, pullreqnumber uint, setSHAs func(sourceSHA, mergeBaseSHA string), param ...repoModel.FileDiff) error {
	pr, err := s.pullRequestStore.GetByID(ctx, repoid, pullreqnumber)
	if err != nil {
		return err
	}

	repo, err := s.repoSvc.GetRepoCheckAccessByID(ctx, session, repoid, []constant.Permission{constant.PermissionRepoView})
	if err != nil {
		return err
	}

	if setSHAs != nil {
		setSHAs(pr.SourceSHA, pr.MergeBaseSHA)
	}

	var files []api.FileDiffRequest
	for _, f := range param {
		files = append(files, api.FileDiffRequest{
			Path:      f.Path,
			StartLine: f.StartLine,
		})
	}

	err = s.gitSvc.RawDiff(ctx, w, &git.DiffParams{
		ReadParams: git.ReadParams{RepoUID: repo.GitUID},
		BaseRef:    pr.MergeBaseSHA,
		HeadRef:    pr.SourceSHA,
		MergeBase:  true,
	}, files...)

	return gitUtil.ProcessGitErrorf(ctx, err)
}

func (s *Service) Commits(ctx context.Context, session *auth.Session, repoid uint, pullreqnumber uint, page, limit uint) ([]commit.Commit, int, error) {
	pr, err := s.pullRequestStore.GetByID(ctx, repoid, pullreqnumber)
	if err != nil {
		return nil, 0, err
	}

	repo, err := s.repoSvc.GetRepoCheckAccessByID(ctx, session, repoid, []constant.Permission{constant.PermissionRepoView})
	if err != nil {
		return nil, 0, err
	}

	gitRef := pr.SourceSHA
	afterRef := pr.MergeBaseSHA

	output, err := s.gitSvc.ListCommits(ctx, &git.ListCommitsParams{
		ReadParams: git.ReadParams{RepoUID: repo.GitUID},
		GitREF:     gitRef,
		After:      afterRef,
		Page:       int32(page),
		Limit:      int32(limit),
	})
	if err != nil {
		return nil, 0, gitUtil.ProcessGitErrorf(ctx, err)
	}

	commits := make([]commit.Commit, len(output.Commits))
	for i := range output.Commits {
		var commit *commit.Commit
		commit, err = utils.MapCommit(&output.Commits[i])
		if err != nil {
			return nil, 0, fmt.Errorf("failed to map commit: %w", err)
		}
		commits[i] = *commit
	}

	return commits, output.TotalCommits, nil
}

func (s *Service) Merge(ctx context.Context, session *auth.Session, repoid uint, pullreqnumber uint, params repoModel.Merge) (*repoModel.MergeResponse, *repoModel.MergeViolations, error) {
	var err error

	requiredPermission := constant.PermissionRepoPush
	if params.DryRun {
		requiredPermission = constant.PermissionRepoView
	}

	// 检查repo 权限
	repo, err := s.repoSvc.GetRepoCheckAccessByID(ctx, session, repoid, []constant.Permission{requiredPermission})
	if err != nil {
		return nil, nil, err
	}

	// TODO 加锁

	pr, err := s.pullRequestStore.GetByID(ctx, repoid, pullreqnumber)
	if err != nil {
		return nil, nil, err
	}

	// 检查pr 状态
	if pr.Status == repoModel.PullRequestStatusMerged {
		return nil, nil, errors.WithI18nKey(err, i18n.PullrequestMergeAlreadyMerged)
	}
	if pr.SourceSHA != params.SourceSHA {
		return nil, nil, errors.WithI18nKey(err, i18n.PullrequestMergeSourceShaNotMatch)
	}

	// 准备 RPC 参数
	writeParams, err := utils.CreateRPCInternalWriteParams(ctx, s.urlProvider, session, repo)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitCreateExecParamError)
		return nil, nil, err
	}

	// 模拟合并
	if params.DryRun {
		var mergeOutput git.MergeOutput

		if params.Method == "" {
			params.Method = repoModel.MergeMethodMerge
		}

		checkMergeability := func(method repoModel.MergeMethod) bool {
			switch method {
			case repoModel.MergeMethodMerge, repoModel.MergeMethodSquash:
				return pr.MergeCheckStatus == repoModel.MergeCheckStatusUnchecked
			case repoModel.MergeMethodRebase:
				return pr.RebaseCheckStatus == repoModel.MergeCheckStatusUnchecked
			case repoModel.MergeMethodFastForward: // Not important for fast-forward merge
				return true
			default:
				return true
			}
		}(params.Method)

		if checkMergeability {
			mergeOutput, err = s.gitSvc.Merge(ctx, &git.MergeParams{
				WriteParams:     writeParams,
				BaseBranch:      pr.TargetBranch,
				HeadRepoUID:     repo.GitUID,
				HeadBranch:      pr.SourceBranch,
				RefType:         enum.RefTypeUndefined,
				HeadExpectedSHA: sha.Must(params.SourceSHA),
				Method:          enum.MergeMethod(params.Method),
			})
			if err != nil {
				return nil, nil, util.ProcessGitErrorf(ctx, err)
			}
			_ = mergeOutput

			// Dry Run（试运行）模式下，更新 Pull Request 的合并状态缓存 。
			err = s.pullRequestStore.UpdateDiffStatsSafe(ctx, repoid, pr, &repoModel.DiffStats{
				HeadHash:     params.SourceSHA,
				Commits:      uint(mergeOutput.CommitCount),
				FilesChanged: uint(mergeOutput.ChangedFileCount),
				Additions:    uint(mergeOutput.Additions),
				Deletions:    uint(mergeOutput.Deletions),
			})
			if err != nil {
				// // TODO log
				return nil, nil, nil
			}
			// TODO send sse event
			return nil, nil, nil
		}

		var conflicts []string
		if params.Method == repoModel.MergeMethodRebase {
			conflicts = pr.RebaseConflicts
		} else {
			conflicts = pr.MergeConflicts
		}

		out := &repoModel.MergeResponse{
			DryRun:        true,
			Mergeable:     len(conflicts) == 0,
			ConflictFiles: conflicts,
		}

		return out, nil, nil
	}

	// --- 正式合并 ---

	// 准备提交信息
	var author *git.Identity
	switch params.Method {
	case repoModel.MergeMethodMerge:
		author = &git.Identity{
			Name:  session.User.Account,
			Email: session.User.Email,
		}
	case repoModel.MergeMethodSquash:
		author = &git.Identity{
			Name:  session.User.Account,
			Email: session.User.Email,
		} // TODO commit 真实作者
	case repoModel.MergeMethodRebase, repoModel.MergeMethodFastForward:
		author = nil
	}

	var committer *git.Identity
	switch params.Method {
	case repoModel.MergeMethodMerge, repoModel.MergeMethodSquash:
		committer = &git.Identity{
			Name:  session.Principal.DisplayName,
			Email: session.Principal.Email,
		}
	case repoModel.MergeMethodRebase:
		committer = &git.Identity{
			Name:  session.User.Account,
			Email: session.User.Email,
		}
	case repoModel.MergeMethodFastForward:
		committer = nil // Not important for fast-forward merge
	}

	// backfill commit title if none provided
	if params.Title == "" {
		switch params.Method {
		case repoModel.MergeMethodMerge:
			params.Title = fmt.Sprintf("Merge branch '%s' of %s (#%d)", pr.SourceBranch, repo.Path, pr.ID)
			// TODO 这个还保留不？
			// if pr.Flow == repoModel.PullRequestFlowZentao || strings.HasPrefix(pr.SourceBranch, "ztflow-") {
			// 	params.Title = fmt.Sprintf("Merge branch '%s' of %s (#%d)", pr.TargetBranch, repo.Path, pr.ID)
			// }
		case repoModel.MergeMethodSquash:
			params.Title = fmt.Sprintf("%s (#%d)", pr.Title, pr.ID)
		case repoModel.MergeMethodRebase, repoModel.MergeMethodFastForward:
			// Not used.
		}
	}

	// 执行git 合并操作
	now := time.Now()
	mergeOutput, err := s.gitSvc.Merge(ctx, &git.MergeParams{
		WriteParams:     writeParams,
		BaseBranch:      pr.TargetBranch,
		HeadRepoUID:     repo.GitUID,
		HeadBranch:      pr.SourceBranch,
		Message:         git.CommitMessage(params.Title, params.Message),
		Committer:       committer,
		CommitterDate:   &now,
		Author:          author,
		AuthorDate:      &now,
		RefType:         enum.RefTypeBranch,
		RefName:         pr.TargetBranch,
		HeadExpectedSHA: sha.Must(params.SourceSHA),
		Method:          enum.MergeMethod(params.Method),
	})
	if err != nil {
		return nil, nil, util.ProcessGitErrorf(ctx, err)
	}
	_ = mergeOutput

	// 处理合并冲突
	// if mergeOutput.MergeSHA.String() == "" || len(mergeOutput.ConflictFiles) > 0 {
	// 	_, err = c.pullreqStore.UpdateOptLock(ctx, pr, func(pr *types.PullReq) error {
	// 		if pr.SourceSHA != mergeOutput.HeadSHA.String() {
	// 			return errors.New("source SHA has changed")
	// 		}

	// 		// update all Merge specific information
	// 		pr.MergeBaseSHA = mergeOutput.MergeBaseSHA.String()
	// 		pr.MergeTargetSHA = ptr.String(mergeOutput.BaseSHA.String())
	// 		pr.MergeSHA = nil
	// 		pr.UpdateMergeOutcome(params.Method, mergeOutput.ConflictFiles)
	// 		pr.Stats.DiffStats = types.NewDiffStats(
	// 			mergeOutput.CommitCount,
	// 			mergeOutput.ChangedFileCount,
	// 			mergeOutput.Additions,
	// 			mergeOutput.Deletions,
	// 		)
	// 		return nil
	// 	})
	// 	if err != nil {
	// 		// non-critical error
	// 		log.Ctx(ctx).Warn().Err(err).Msg("failed to update pull request with conflict files")
	// 	} else {
	// 		if err = c.sseStreamer.Publish(ctx, targetRepo.ParentID, enum.SSETypePullRequestUpdated, pr); err != nil {
	// 			log.Ctx(ctx).Warn().Err(err).Msg("failed to publish PR changed event")
	// 		}
	// 	}

	// 	log.Ctx(ctx).Info().Msg("aborting pull request merge because of conflicts")

	// 	return nil, &repoModel.MergeViolations{
	// 		ConflictFiles:  mergeOutput.ConflictFiles,
	// 		RuleViolations: violations,
	// 		// In case of conflicting files we prioritize those for the error message.
	// 		Message: fmt.Sprintf("Merge blocked by conflicting files: %v", mergeOutput.ConflictFiles),
	// 	}, nil
	// }

	// 合并成功后首位

	// 记录活动日志

	// 触发事件 - 触发 Merged 事件，这可能会触发 Webhooks 或 CI/CD 流水线。

	// 通知 审计 埋点

	// 返回结果

	return nil, nil, nil

}

type CompareInfo struct {
	BaseRef   string
	HeadRef   string
	MergeBase bool
}

func parseDiffPath(path string) (CompareInfo, error) {
	infos := strings.SplitN(path, "...", 2)
	if len(infos) != 2 {
		infos = strings.SplitN(path, "..", 2)
	}
	if len(infos) != 2 {
		err := errors.Newf("invalid format \"%s\"", path)
		err = errors.WithI18nKey(err, i18n.PullrequestPathParamError)
		return CompareInfo{}, err
	}
	return CompareInfo{
		BaseRef:   infos[0],
		HeadRef:   infos[1],
		MergeBase: strings.Contains(path, "..."),
	}, nil
}
