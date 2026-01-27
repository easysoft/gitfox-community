package commit

import (
	"context"

	commitModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/utils"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
)

type service struct {
	git git.Interface
}

func newService(git git.Interface) Service {
	return &service{git: git}
}

// FindRef finds information about a commit in Harness for the git ref.
// This is using the branch only as the ref at the moment, can be changed
// when needed to take any ref (like sha, tag).
func (f *service) FindRef(
	ctx context.Context,
	repo *repoModel.Repo,
	branch string,
) (*commitModel.Commit, error) {
	readParams := git.ReadParams{
		RepoUID: repo.GitUID,
	}
	branchOutput, err := f.git.GetBranch(ctx, &git.GetBranchParams{
		ReadParams: readParams,
		BranchName: branch,
	})
	if err != nil {
		return nil, err
	}

	// convert the RPC commit output to a types.Commit.
	return utils.MapCommit(branchOutput.Branch.Commit)
}

// FindCommit finds information about a commit in Harness for the git SHA.
func (f *service) FindCommit(
	ctx context.Context,
	repo *repoModel.Repo,
	rawSHA string,
) (*commitModel.Commit, error) {
	readParams := git.ReadParams{
		RepoUID: repo.GitUID,
	}
	commitOutput, err := f.git.GetCommit(ctx, &git.GetCommitParams{
		ReadParams: readParams,
		Revision:   rawSHA,
	})
	if err != nil {
		return nil, err
	}

	// convert the RPC commit output to a types.Commit.
	return utils.MapCommit(&commitOutput.Commit)
}

// GetChangedFiles returns a list of files changed between two commits.
func (f *service) GetChangedFiles(
	ctx context.Context,
	repo *repoModel.Repo,
	before, after string,
) ([]string, error) {
	readParams := git.ReadParams{
		RepoUID: repo.GitUID,
	}

	// 使用 DiffFileNames 获取变更文件列表
	output, err := f.git.DiffFileNames(ctx, &git.DiffParams{
		ReadParams: readParams,
		BaseRef:    before,
		HeadRef:    after,
	})
	if err != nil {
		return nil, err
	}

	return output.Files, nil
}
