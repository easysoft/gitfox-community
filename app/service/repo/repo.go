package repo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/githook"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/utils"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/space"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/slice"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/resources"
)

type Service struct {
	gitSvc      git.Interface
	repoStore   repo.Interface
	spaceStore  space.Interface
	urlProvider url.Provider
}

var _ Interface = (*Service)(nil)

// NewService 创建 SpaceService 实例
func NewRepoSvc(gitSvc git.Interface, repoStore repo.Interface, spaceStore space.Interface, urlProvider url.Provider) *Service {
	return &Service{
		gitSvc:      gitSvc,
		repoStore:   repoStore,
		spaceStore:  spaceStore,
		urlProvider: urlProvider,
	}
}

func (s *Service) Create(ctx context.Context, req *repoModel.CreateRepoReq) (*repoModel.Repo, error) {
	_, err := s.spaceStore.GetByID(ctx, req.SpaceID)
	if err != nil {
		return nil, errors.NotFound.Newf("failed to find the parent space: %v", err)
	}
	// session, ok := reqPkg.GetAuthSession(ctx)
	// if !ok {
	// 	return nil, errors.SystemErr.Newf("error getting auth session: %w", err)
	// }

	//TODO : mock session
	session := &auth.Session{
		User: &principal.User{
			ID:      1,
			Email:   "actor@chandao.com",
			Account: "actor",
		},
	}

	gitResp, isEmpty, err := s.createGitRepository(ctx, req, session)
	if err != nil {
		return nil, errors.SystemErr.Newf("error creating repository on git: %v", err)
	}

	repo := &repo.Repository{
		Mutex:         0,
		SpaceID:       req.SpaceID,
		Identifier:    req.Identifier,
		GitUID:        gitResp.UID,
		Desc:          req.Desc,
		CreatedBy:     session.User.Account,
		ForkID:        req.ForkID,
		DefaultBranch: req.DefaultBranch,
		Empty:         isEmpty,
		ACL:           req.ACL,
	}

	return s.repoStore.Create(ctx, repo)
}

func (s *Service) createGitRepository(ctx context.Context, req *repoModel.CreateRepoReq, session *auth.Session) (*git.CreateRepositoryOutput, bool, error) {
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
	if req.License != "" && req.License != "none" {
		content, err = resources.Licence.ReadFile(fmt.Sprintf("license/%s.txt", req.License))
		if err != nil {
			return nil, false, errors.Wrapf(err, "failed to read license '%s'", req.License)
		}
		files = append(files, git.File{
			Path:    "LICENSE",
			Content: content,
		})
	}

	if req.Ignore != "" {
		content, err = resources.Gitignore.ReadFile(fmt.Sprintf("gitignore/%s.gitignore", req.Ignore))
		if err != nil {
			return nil, false, errors.Wrapf(err, "failed to read git ignore '%s'", req.Ignore)
		}
		files = append(files, git.File{
			Path:    ".gitignore",
			Content: content,
		})
	}

	// generate envars (add everything githook CLI needs for execution)
	envVars, err := githook.GenerateEnvironmentVariables(
		ctx,
		s.urlProvider.GetInternalAPIURL(ctx),
		0,
		session.User.ID,
		true,
		true,
	)
	if err != nil {
		return nil, false, errors.SystemErr.Newf("failed to generate git hook environment variables: %v", err)
	}
	now := time.Now()
	resp, err := s.gitSvc.CreateRepository(ctx, &git.CreateRepositoryParams{
		Actor: git.Identity{
			Name:  session.User.Account,
			Email: session.User.Email,
		},
		EnvVars:       envVars,
		DefaultBranch: req.DefaultBranch,
		Files:         files,
		Author: &git.Identity{
			Name:  session.User.Account,
			Email: session.User.Email,
		},
		AuthorDate: &now,
		Committer: &git.Identity{
			Name:  session.User.Account,
			Email: session.User.Email,
		},
		CommitterDate: &now,
	})
	if err != nil {
		return nil, false, errors.Wrapf(err, "failed to create repo")
	}
	return resp, len(files) == 0, nil
}

func (s *Service) GetRepoCheckAccess(ctx context.Context, session *auth.Session, repoRef string, reqPermission []constant.Permission) (*repoModel.Repo, error) {

	// get repo
	repo, err := s.GetRepo(ctx, repoRef, []constant.RepoState{constant.RepoStateActive, constant.RepoStateMigrateGitPush})
	if err != nil {
		return nil, err
	}

	// check repo todo
	// reqPermission

	return repo, nil
}

func (s *Service) GetRepo(ctx context.Context, repoRef string, allowedStates []constant.RepoState) (*repoModel.Repo, error) {
	if repoRef == "" {
		return nil, errors.BadRequest.New("A valid repository reference must be provided.")
	}

	repo, err := s.repoStore.FindByRef(ctx, repoRef)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find repository: %s", repoRef)
	}

	if len(allowedStates) > 0 && !slice.Contain(allowedStates, constant.RepoState(repo.Status)) {
		return nil, errors.BadRequest.New("Repository is not ready to use.")
	}

	return repo, nil
}

func (s *Service) GetByID(ctx context.Context, session *auth.Session, id uint) (*repoModel.Repo, error) {
	if id < 1 {
		return nil, errors.BadRequest.New("A valid repository id must be provided.")
	}

	repo, err := s.repoStore.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find repository: %d", id)
	}

	// _, err = s.GetRepoCheckAccess(ctx, session, repo.Ref, []constant.Permission{constant.PermissionRepoRead})
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

func (s *Service) GetDiff(ctx context.Context, req repoModel.ReqForDiff, repo repoModel.Repo, w io.Writer) error {
	info, err := utils.ParseDiffPath(req.Range)
	if err != nil {
		return err
	}
	return s.gitSvc.RawDiff(ctx, w, &git.DiffParams{
		ReadParams: git.ReadParams{RepoUID: repo.GitUID},
		BaseRef:    info.BaseRef,
		HeadRef:    info.HeadRef,
		MergeBase:  info.MergeBase,
	}, req.Files...)
}

func (s *Service) Paths(ctx context.Context, req repoModel.PathsInput, repo repoModel.Repo) (*repoModel.PathsOutput, error) {
	var gitRef string
	if req.GitRef == "" {
		gitRef = repo.DefaultBranch
	}
	rpcOut, err := s.gitSvc.ListPaths(ctx, &git.ListPathsParams{
		ReadParams:         git.ReadParams{RepoUID: repo.GitUID},
		GitREF:             gitRef,
		IncludeDirectories: req.IncludeDirectories,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list git paths: %w", err)
	}

	return &repoModel.PathsOutput{
		Files:       rpcOut.Files,
		Directories: rpcOut.Directories,
	}, nil
}

func (s *Service) PathsDetails(ctx context.Context, req repoModel.PathsDetailsInput, repo repoModel.Repo) (*repoModel.PathsDetailsOutput, error) {
	var gitRef string
	if req.GitRef == "" {
		gitRef = repo.DefaultBranch
	}

	result, err := s.gitSvc.PathsDetails(ctx, git.PathsDetailsParams{
		ReadParams: git.ReadParams{RepoUID: repo.GitUID},
		GitREF:     gitRef,
		Paths:      req.Paths,
	})
	if err != nil {
		return nil, err
	}

	return &repoModel.PathsDetailsOutput{
		Details: result.Details,
	}, nil

}
