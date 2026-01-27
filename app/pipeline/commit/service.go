package commit

import (
	"context"

	commitModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
)

type (
	// Service provides access to commit information via
	// the SCM provider. Today, this is Harness but it can
	// be extendible to any SCM provider.
	Service interface {
		// ref is the ref to fetch the commit from, eg refs/heads/master
		FindRef(ctx context.Context, repo *repoModel.Repo, ref string) (*commitModel.Commit, error)

		// FindCommit returns information about a commit in a repo.
		FindCommit(ctx context.Context, repo *repoModel.Repo, sha string) (*commitModel.Commit, error)

		// GetChangedFiles returns a list of files changed between two commits.
		GetChangedFiles(ctx context.Context, repo *repoModel.Repo, before, after string) ([]string, error)
	}
)
