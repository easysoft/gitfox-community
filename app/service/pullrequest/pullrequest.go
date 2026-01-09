package pullrequest

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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
	err = s.pullRequestStore.UpdateDiffStats(ctx, repoid, pullreqnumber, re)
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

	return util.ProcessGitErrorf(ctx, err)
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
		return nil, 0, util.ProcessGitErrorf(ctx, err)
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
