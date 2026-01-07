package repo

import (
	"context"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	gitUtil "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/util"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

func (s *Service) GetMetadata(ctx context.Context, sess *auth.Session, repoid uint, pullreqnumber uint) (*repoModel.DiffStats, error) {

	// find pull request by number
	pr, err := s.repoStore.GetPullRequestByID(ctx, repoid, pullreqnumber)
	if err != nil {
		return nil, err
	}

	repo, err := s.GetRepoCheckAccessByID(ctx, sess, repoid, []constant.Permission{constant.PermissionRepoView})
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
	err = s.repoStore.UpdatePullRequestDiffStats(ctx, repoid, pullreqnumber, re)
	if err != nil {
		return nil, err
	}

	return re, nil
}

func (s *Service) GetDiffStats(ctx context.Context, sess *auth.Session, repoid uint, pathTopath string) (*repoModel.DiffStats, error) {

	repo, err := s.GetRepoCheckAccessByID(ctx, sess, repoid, []constant.Permission{constant.PermissionRepoView})
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
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
		return CompareInfo{}, err
	}
	return CompareInfo{
		BaseRef:   infos[0],
		HeadRef:   infos[1],
		MergeBase: strings.Contains(path, "..."),
	}, nil
}
