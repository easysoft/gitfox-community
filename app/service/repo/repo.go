package repo

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/slice"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/git"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
)

type Service struct {
	gitSvc    git.Interface
	repoStore repoStore.Interface
}

func NewRepoService(gitSvc git.Interface, repoStore repoStore.Interface) Interface {
	return &Service{gitSvc: gitSvc, repoStore: repoStore}
}

func (s *Service) GetRepoCheckAccess(ctx context.Context, session *auth.Session, repo_ref string, reqPermission []constant.Permission) (*repo.Repository, error) {

	// get repo
	repo, err := s.GetRepo(ctx, repo_ref, []constant.RepoState{constant.RepoStateActive, constant.RepoStateMigrateGitPush})
	if err != nil {
		return nil, err
	}

	// check repo todo
	// reqPermission

	return repo, nil
}

func (s *Service) GetRepo(ctx context.Context, repo_ref string, allowedStates []constant.RepoState) (*repo.Repository, error) {
	if repo_ref == "" {
		return nil, errors.BadRequest.New("A valid repository reference must be provided.")
	}

	repo, err := s.repoStore.FindByRef(ctx, repo_ref)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find repository: %s", repo_ref)
	}

	if len(allowedStates) > 0 && !slice.Contain(allowedStates, repo.State) {
		return nil, errors.BadRequest.New("Repository is not ready to use.")
	}

	return repo, nil
}
