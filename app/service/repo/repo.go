package repo

import (
	"bytes"
	"context"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/slice"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
)

type Service struct {
	gitSvc    git.Interface
	repoStore repo.Interface
}

var _ Interface = (*Service)(nil)

// NewService 创建 SpaceService 实例
func NewRepoSvc(gitSvc git.Interface, repoStore repo.Interface) *Service {
	return &Service{
		gitSvc:    gitSvc,
		repoStore: repoStore,
	}
}

func (s *Service) Create(ctx context.Context, req *repoModel.CreateRepoReq) (*repoModel.CreateRepoResp, error) {
	var (
		err     error
		content []byte
	)
	files := make([]git.File, 0, 3) // readme, gitignore, licence
	if req.Readme {
		content = s.createReadme(req.Identifier, req.Desc)
		files = append(files, git.File{
			Path:    "README.md",
			Content: content,
		})
	}
	return &repoModel.CreateRepoResp{}, err
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

	if len(allowedStates) > 0 && !slice.Contain(allowedStates, constant.RepoState(repo.Status)) {
		return nil, errors.BadRequest.New("Repository is not ready to use.")
	}

	return repo, nil
}



func (s *Service) createReadme(name, description string) []byte {
	content := bytes.Buffer{}
	content.WriteString("# " + name + "\n")
	if description != "" {
		content.WriteString(description)
	}
	return content.Bytes()
}