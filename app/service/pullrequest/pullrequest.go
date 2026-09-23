// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package pullrequest

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	"github.com/easysoft/gitfox-community/app/constant"
	gitevents "github.com/easysoft/gitfox-community/app/events/git"
	pullreqevents "github.com/easysoft/gitfox-community/app/events/pullreq"
	"github.com/easysoft/gitfox-community/app/githook"
	"github.com/easysoft/gitfox-community/app/model/auth"
	"github.com/easysoft/gitfox-community/app/model/commit"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	"github.com/easysoft/gitfox-community/app/service/locker"
	"github.com/easysoft/gitfox-community/app/service/repo"
	"github.com/easysoft/gitfox-community/app/service/sse"
	"github.com/easysoft/gitfox-community/app/service/utils"
	"github.com/easysoft/gitfox-community/app/store/pullrequest"
	repoStore "github.com/easysoft/gitfox-community/app/store/repo"
	"github.com/easysoft/gitfox-community/pkg/auth/authn"
	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/events"
	"github.com/easysoft/gitfox-community/pkg/git"
	"github.com/easysoft/gitfox-community/pkg/git/api"
	"github.com/easysoft/gitfox-community/pkg/git/enum"
	"github.com/easysoft/gitfox-community/pkg/git/sha"
	"github.com/easysoft/gitfox-community/pkg/git/util"
	gitUtil "github.com/easysoft/gitfox-community/pkg/git/util"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/easysoft/gitfox-community/pkg/pubsub"
	"github.com/easysoft/gitfox-community/pkg/stream"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
	"github.com/easysoft/gitfox-community/pkg/utils/url"
	"github.com/gotidy/ptr"
)

type Service struct {
	gitSvc             git.Interface
	repoSvc            repo.Interface
	repoStore          repoStore.Interface
	locker             *locker.Locker
	pullRequestStore   pullrequest.Interface
	pullreqEvReporter  *pullreqevents.Reporter
	sseStreamer        sse.Streamer
	urlProvider        url.Provider
	cancelMutex        sync.Mutex
	cancelMergeability map[string]context.CancelFunc
	pubsub             pubsub.PubSub
	pushReqCleaner     PushRequestCleaner
}

const (
	fastForwardNotPossibleMessage = "target branch has diverged from the source branch, fast-forward not possible"
	// sourceOutOfDateMessage matches the precondition failure raised when the recorded source SHA no
	// longer equals the head branch tip (e.g. new commits were pushed after the merge check ran).
	sourceOutOfDateMessage = "which doesn't match expected SHA"
)

var _ Interface = (*Service)(nil)

// NewPullRequestSvc 创建 PullRequestService 实例
func NewPullRequestSvc(
	ctx context.Context,
	gitReaderFactory *events.ReaderFactory[*gitevents.Reader],
	pullReqEvFactory *events.ReaderFactory[*pullreqevents.Reader],
	pullReqEvReporter *pullreqevents.Reporter,
	cfg *config.Config,
	gitSvc git.Interface,
	repoSvc repo.Interface,
	repoStore repoStore.Interface,
	locker *locker.Locker,
	pullRequestStore pullrequest.Interface,
	sseStreamer sse.Streamer,
	bus pubsub.PubSub,
	urlProvider url.Provider,
	pushReqCleaner PushRequestCleaner,
) (*Service, error) {
	service := Service{
		pullreqEvReporter:  pullReqEvReporter,
		gitSvc:             gitSvc,
		repoSvc:            repoSvc,
		repoStore:          repoStore,
		locker:             locker,
		pullRequestStore:   pullRequestStore,
		sseStreamer:        sseStreamer,
		cancelMergeability: make(map[string]context.CancelFunc),
		pubsub:             bus,
		urlProvider:        urlProvider,
		pushReqCleaner:     pushReqCleaner,
	}

	var err error

	// handle git branch events to trigger specific pull request events
	const groupGit = "gitfox:pullreq:git"
	_, err = gitReaderFactory.Launch(ctx, groupGit, cfg.InstanceID,
		func(r *gitevents.Reader) error {
			const idleTimeout = 15 * time.Second
			r.Configure(
				stream.WithConcurrency(1),
				stream.WithHandlerOptions(
					stream.WithIdleTimeout(idleTimeout),
					stream.WithMaxRetries(3),
				))

			_ = r.RegisterBranchUpdated(service.triggerPREventOnBranchUpdate)
			_ = r.RegisterBranchDeleted(service.closePullReqOnBranchDelete)

			return nil
		})
	if err != nil {
		return nil, err
	}

	// pull request ref maintenance
	const groupPullReqHeadRef = "gitfox:pullreq:headref"
	_, err = pullReqEvFactory.Launch(ctx, groupPullReqHeadRef, cfg.InstanceID,
		func(r *pullreqevents.Reader) error {
			const idleTimeout = 10 * time.Second
			r.Configure(
				stream.WithConcurrency(1),
				stream.WithHandlerOptions(
					stream.WithIdleTimeout(idleTimeout),
					stream.WithMaxRetries(3),
				))

			_ = r.RegisterCreated(service.createHeadRefOnCreated)
			_ = r.RegisterBranchUpdated(service.updateHeadRefOnBranchUpdate)
			_ = r.RegisterReopened(service.updateHeadRefOnReopen)

			return nil
		})
	if err != nil {
		return nil, err
	}

	// mergeability check
	const groupPullReqMergeable = "gitfox:pullreq:mergeable"
	_, err = pullReqEvFactory.Launch(ctx, groupPullReqMergeable, cfg.InstanceID,
		func(r *pullreqevents.Reader) error {
			const idleTimeout = 30 * time.Second
			r.Configure(
				stream.WithConcurrency(3),
				stream.WithHandlerOptions(
					stream.WithIdleTimeout(idleTimeout),
					stream.WithMaxRetries(2),
				))

			_ = r.RegisterCreated(service.mergeCheckOnCreated)
			_ = r.RegisterBranchUpdated(service.mergeCheckOnBranchUpdate)
			_ = r.RegisterReopened(service.mergeCheckOnReopen)
			_ = r.RegisterClosed(service.mergeCheckOnClosed)
			_ = r.RegisterMerged(service.mergeCheckOnMerged)

			return nil
		})
	if err != nil {
		return nil, err
	}

	// cancel any previous pr mergeability check
	// payload is oldsha.
	_ = bus.Subscribe(ctx, cancelMergeCheckKey, func(payload []byte) error {
		oldSHA := string(payload)
		if oldSHA == "" {
			return nil
		}

		service.cancelMutex.Lock()
		defer service.cancelMutex.Unlock()

		cancel := service.cancelMergeability[oldSHA]
		if cancel != nil {
			cancel()
		}

		delete(service.cancelMergeability, oldSHA)

		return nil
	}, pubsub.WithChannelNamespace("pullreq"))

	return &service, nil
}

func processPullRequestMergeError(ctx context.Context, method repoModel.MergeMethod, err error) error {
	if err == nil {
		return nil
	}

	// Fast-forward merge only succeeds when the target tip is an ancestor of the source tip.
	// Translate this specific git conflict into a user-facing i18n error for PR merges.
	if method == repoModel.MergeMethodFastForward &&
		strings.Contains(err.Error(), fastForwardNotPossibleMessage) {
		return errors.WithI18nKey(errors.New("target branch latest commit is not an ancestor of source branch"), i18n.GitFastForwardTargetNotAncestor)
	}

	// The recorded source SHA no longer matches the head branch tip (new commits were pushed after
	// the merge check, or the review branch was rewritten). Surface an actionable message instead
	// of the raw "git command failed" fallback, which gives the user no way to recover.
	if strings.Contains(err.Error(), sourceOutOfDateMessage) {
		return errors.WithI18nKey(errors.New("source branch is out of date, please refresh and retry"), i18n.GitMergeSourceOutOfDate)
	}

	return util.ProcessGitErrorf(ctx, err)
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

	if pr.CommitCount != 0 && pr.FileCount != 0 && pr.Additions != 0 && pr.Deletions != 0 {
		return &repoModel.DiffStats{
			MergeSHA:     ptr.String(""),
			Commits:      pr.CommitCount,
			FilesChanged: pr.FileCount,
			Additions:    pr.Additions,
			Deletions:    pr.Deletions,
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
	_, err = s.pullRequestStore.UpdateDiffStatsSafe(ctx, repoid, pr, re)
	if err != nil {
		return nil, err
	}

	re.MergeSHA = ptr.String("")

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

	mergeBaseResult, err := s.gitSvc.MergeBase(ctx, git.MergeBaseParams{
		ReadParams: git.ReadParams{RepoUID: repo.GitUID},
		Ref1:       info.BaseRef,
		Ref2:       info.HeadRef,
	})
	if err != nil {
		return nil, gitUtil.ProcessGitErrorf(ctx, err)
	}

	re := &repoModel.DiffStats{
		MergeSHA:     ptr.String(""),
		Commits:      uint(output.Commits),
		MergeBaseSHA: mergeBaseResult.MergeBaseSHA.String(),
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

func (s *Service) Commits(ctx context.Context, session *auth.Session, repoid uint, pullreqnumber uint, page, pageSize uint) ([]commit.Commit, int, error) {
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
		Limit:      int32(pageSize),
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

func (s *Service) Merge(ctx context.Context, session *auth.Session, repoid uint, pullreqID uint, params repoModel.Merge) (*repoModel.MergeResponse, error) {
	var err error
	logger := logger.GetLogger(ctx)

	requiredPermission := constant.PermissionRepoPush
	if params.DryRun {
		requiredPermission = constant.PermissionRepoView
	}

	// 检查repo 权限
	repo, err := s.repoSvc.GetRepoCheckAccessByID(ctx, session, repoid, []constant.Permission{requiredPermission})
	if err != nil {
		return nil, err
	}
	// 加锁
	const timeout = 3 * time.Minute
	var lockID int64 // 0 means locking repo level for prs (for actual merging)
	if params.DryRun {
		lockID = int64(pullreqID) // dryrun doesn't need repo level lock
	}

	unlock, err := s.locker.LockPR(ctx, int64(repo.ID), lockID, timeout+30*time.Second)
	if err != nil {
		return nil, errors.WithI18nKey(err, i18n.PullrequestMergeLockError)
	}
	defer unlock()

	pr, err := s.pullRequestStore.GetByID(ctx, repoid, pullreqID)
	if err != nil {
		return nil, err
	}

	// 检查pr 状态
	if pr.Status == repoModel.PullRequestStatusMerged {
		return nil, errors.WithI18nKey(err, i18n.PullrequestMergeAlreadyMerged)
	}
	if pr.SourceSHA != params.SourceSHA {
		return nil, errors.WithI18nKey(err, i18n.PullrequestMergeSourceShaNotMatch)
	}

	// 准备 RPC 参数
	writeParams, err := utils.CreateRPCInternalWriteParams(ctx, s.urlProvider, session, repo)
	if err != nil {
		return nil, err
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
				return pr.MergeCheckStatus == repoModel.MergeCheckStatusUnchecked || pr.MergeCheckStatus == ""
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
				return nil, processPullRequestMergeError(ctx, params.Method, err)
			}

			df := &repoModel.DiffStats{
				SourceSHA:      mergeOutput.HeadSHA.String(),
				MergeTargetSHA: mergeOutput.BaseSHA.String(),
				MergeBaseSHA:   mergeOutput.MergeBaseSHA.String(),
				MergeSHA:       ptr.String(""),
				HeadHash:       mergeOutput.HeadSHA.String(),
				Commits:        uint(mergeOutput.CommitCount),
				FilesChanged:   uint(mergeOutput.ChangedFileCount),
				Additions:      uint(mergeOutput.Additions),
				Deletions:      uint(mergeOutput.Deletions),
			}
			df.UpdateMergeOutcome(params.Method, mergeOutput.ConflictFiles)

			// Dry Run（试运行）模式下，更新 Pull Request 的合并状态缓存 。
			newPr, err := s.pullRequestStore.UpdateDiffStatsSafe(ctx, repoid, pr, df)
			if err != nil {
				logger.Errorf("failed to update unchecked pull request diff stats,err:%v", err)
			} else {
				pr = newPr
			}

			// TODO: send sse event
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

		return out, nil
	}

	// --- 正式合并 ---

	// 准备提交信息
	// Author（作者）：原始代码的编写者
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
		} // TODO: commit 真实作者
	case repoModel.MergeMethodRebase, repoModel.MergeMethodFastForward:
		author = nil
	}

	// Committer（提交者）：将代码提交到仓库的人
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

			params.Title = fmt.Sprintf("Merge branch '%s' of %s (#%d)", pr.TargetBranch, repo.Path, pr.ID)
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
		Message:         params.Title,
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
		return nil, processPullRequestMergeError(ctx, params.Method, err)
	}

	// 处理合并冲突
	if mergeOutput.MergeSHA.String() == "" || len(mergeOutput.ConflictFiles) > 0 {
		df := &repoModel.DiffStats{
			SourceSHA:      mergeOutput.HeadSHA.String(),
			MergeTargetSHA: mergeOutput.BaseSHA.String(),
			MergeBaseSHA:   mergeOutput.MergeBaseSHA.String(),
			MergeSHA:       ptr.String(""),
			HeadHash:       mergeOutput.HeadSHA.String(),
			Commits:        uint(mergeOutput.CommitCount),
			FilesChanged:   uint(mergeOutput.ChangedFileCount),
			Additions:      uint(mergeOutput.Additions),
			Deletions:      uint(mergeOutput.Deletions),
		}
		df.UpdateMergeOutcome(params.Method, mergeOutput.ConflictFiles)
		newPr, err := s.pullRequestStore.UpdateDiffStatsSafe(ctx, repoid, pr, df)
		if err != nil {
			logger.Errorf("failed to update pull request with conflict files,err:%v", err)
		} else {
			pr = newPr
		}

		// TODO: send sse event

		logger.Errorf("aborting pull request merge because of conflicts,conflictFiles:%v", mergeOutput.ConflictFiles)

		err = errors.Newf("Merge blocked by conflicting files: %v", mergeOutput.ConflictFiles)
		return nil, errors.WithI18nKey(err, i18n.PullrequestMergeViolations)
	}

	logger.Infof("successfully merged PR")

	mergeSHA := mergeOutput.MergeSHA.String()
	// 合并成功后收尾工作
	err = s.pullRequestStore.UpdateMergeAfterSafe(ctx, repoid, pr, &repoModel.MergeAfter{
		DiffStats: repoModel.DiffStats{
			HeadHash:       mergeOutput.HeadSHA.String(),
			Commits:        uint(mergeOutput.CommitCount),
			FilesChanged:   uint(mergeOutput.ChangedFileCount),
			Additions:      uint(mergeOutput.Additions),
			Deletions:      uint(mergeOutput.Deletions),
			MergeTargetSHA: mergeOutput.BaseSHA.String(),
			MergeBaseSHA:   mergeOutput.MergeBaseSHA.String(),
			MergeSHA:       &mergeSHA,
		},
		MergedBy:    &session.User.Account,
		MergedDate:  datetime.NowLocalTime(),
		MergeMethod: repoModel.MergeMethod(params.Method),
	})
	if err != nil {
		return nil, err
	}

	// 触发 Merged 事件，这可能会触发 Webhooks 或 CI/CD 流水线。
	s.pullreqEvReporter.Merged(ctx, &pullreqevents.MergedPayload{
		Base: pullreqevents.Base{
			PullReqID:    int64(pr.ID),
			SourceRepoID: int64(pr.SourceRepoID),
			TargetRepoID: int64(pr.TargetRepoID),
			Account:      session.User.Account,
		},
		TargetBranch: pr.TargetBranch,
		MergeMethod:  params.Method,
		MergeSHA:     mergeOutput.MergeSHA.String(),
		TargetSHA:    mergeOutput.BaseSHA.String(),
		SourceSHA:    mergeOutput.HeadSHA.String(),
	})

	// 清理代码推送请求（付费功能）产生的临时分支和引用
	var branchDeleted bool
	if s.pushReqCleaner != nil {
		branchDeleted = s.pushReqCleaner.Cleanup(ctx, pr)
	}

	// 返回结果
	return &repoModel.MergeResponse{
		SHA:           mergeOutput.MergeSHA.String(),
		BranchDeleted: branchDeleted,
	}, nil
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
		BaseRef:   utils.DecodeGitRef(infos[0]),
		HeadRef:   utils.DecodeGitRef(infos[1]),
		MergeBase: strings.Contains(path, "..."),
	}, nil
}

// CreateSystemRPCWriteParams creates base write parameters for system internal write operations.
// It uses the admin user for system-level git operations.
func CreateSystemRPCWriteParams(
	ctx context.Context,
	urlProvider url.Provider,
	repoID int64,
	repoGITUID string,
) (git.WriteParams, error) {
	// generate envars (add everything githook CLI needs for execution)
	envVars, err := githook.GenerateEnvironmentVariables(
		ctx,
		urlProvider.GetInternalAPIURL(ctx),
		repoID,
		authn.HeaderAppAdmin,
		false,
		true,
	)
	if err != nil {
		return git.WriteParams{}, fmt.Errorf("failed to generate git hook environment variables: %w", err)
	}

	return git.WriteParams{
		Actor: git.Identity{
			Name:  authn.HeaderAppAdmin,
			Email: authn.HeaderAdminEmail,
		},
		RepoUID: repoGITUID,
		EnvVars: envVars,
	}, nil
}

// TriggerEvent 根据类型触发 Pull Request 事件
// type: create - 触发 Created 事件
// type: close - 触发 Closed 事件
// type: reopen - 触发 Reopened 事件
func (s *Service) TriggerEvent(
	ctx context.Context,
	session *auth.Session,
	repoID uint,
	pullreqID uint,
	params repoModel.EventTrigger,
) error {
	// 校验事件触发类型
	if err := params.Check(); err != nil {
		return err
	}

	pr, err := s.pullRequestStore.GetByID(ctx, repoID, pullreqID)
	if err != nil {
		return err
	}

	account := session.User.Account

	switch params.Type {
	case repoModel.EventTriggerTypeCreate:
		s.pullreqEvReporter.Created(ctx, &pullreqevents.CreatedPayload{
			Base: pullreqevents.Base{
				PullReqID:    int64(pr.ID),
				SourceRepoID: int64(pr.SourceRepoID),
				TargetRepoID: int64(pr.TargetRepoID),
				Account:      account,
			},
			SourceBranch: pr.SourceBranch,
			TargetBranch: pr.TargetBranch,
			SourceSHA:    pr.SourceSHA,
			MergeBaseSHA: pr.MergeBaseSHA,
		})
	case repoModel.EventTriggerTypeClose:
		s.pullreqEvReporter.Closed(ctx, &pullreqevents.ClosedPayload{
			Base: pullreqevents.Base{
				PullReqID:    int64(pr.ID),
				SourceRepoID: int64(pr.SourceRepoID),
				TargetRepoID: int64(pr.TargetRepoID),
				Account:      account,
			},
			SourceBranch: pr.SourceBranch,
			SourceSHA:    pr.SourceSHA,
		})

		// 清理代码推送请求（付费功能）产生的临时分支和引用
		if s.pushReqCleaner != nil {
			s.pushReqCleaner.Cleanup(ctx, pr)
		}
	case repoModel.EventTriggerTypeReopen:
		s.pullreqEvReporter.Reopened(ctx, &pullreqevents.ReopenedPayload{
			Base: pullreqevents.Base{
				PullReqID:    int64(pr.ID),
				SourceRepoID: int64(pr.SourceRepoID),
				TargetRepoID: int64(pr.TargetRepoID),
				Account:      account,
			},
			SourceBranch: pr.SourceBranch,
			SourceSHA:    pr.SourceSHA,
			MergeBaseSHA: pr.MergeBaseSHA,
		})
	}

	return nil
}
