package repo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/githook"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/utils"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/auth/authz"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	gitAPI "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/api"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/slice"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/resources"
)

type Service struct {
	gitSvc      git.Interface
	repoStore   repo.Interface
	spaceStore  space.Interface
	urlProvider url.Provider
	authorizer  authz.Authorizer
}

var _ Interface = (*Service)(nil)

// NewRepoSvc 创建 RepoService 实例
func NewRepoSvc(
	gitSvc git.Interface,
	repoStore repo.Interface,
	spaceStore space.Interface,
	urlProvider url.Provider,
	authorizer authz.Authorizer,
) *Service {
	return &Service{
		gitSvc:      gitSvc,
		repoStore:   repoStore,
		spaceStore:  spaceStore,
		urlProvider: urlProvider,
		authorizer:  authorizer,
	}
}

func (s *Service) Create(ctx context.Context, session *auth.Session, params *repoModel.CreateRepoReq) (*repoModel.Repo, error) {
	_, err := s.spaceStore.GetByID(ctx, params.SpaceID)
	if err != nil {
		return nil, errors.NotFound.Newf("failed to find the parent space: %v", err)
	}

	gitResp, isEmpty, err := s.createGitRepository(ctx, params, session)
	if err != nil {
		return nil, errors.SystemErr.Newf("error creating repository on git: %v", err)
	}

	repo := &repo.Repository{
		Mutex:         0,
		SpaceID:       params.SpaceID,
		Identifier:    params.Identifier,
		GitUID:        gitResp.UID,
		Desc:          params.Desc,
		CreatedBy:     session.User.Account,
		ForkID:        params.ForkID,
		DefaultBranch: params.DefaultBranch,
		Empty:         isEmpty,
		ACL:           params.ACL,
	}

	repoObj, err := s.repoStore.Create(ctx, repo)
	if err != nil {
		return nil, err
	}
	repoObj.GitSSHURL = s.urlProvider.GenerateGITCloneSSHURL(ctx, repoObj.Path)
	repoObj.GitURL = s.urlProvider.GenerateGITCloneURL(ctx, repoObj.Path)
	return repoObj, nil
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
		session.User.Account,
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

func (s *Service) Update(ctx context.Context, req *repoModel.ReqForUpdate, oldRepo repoModel.Repo, session *auth.Session) (*repoModel.Repo, error) {
	// 调用 git 命令更新默认分支
	if req.DefaultBranch != "" {
		writeParams, err := utils.CreateRPCInternalWriteParams(ctx, s.urlProvider, session, &oldRepo)
		if err != nil {
			return nil, errors.WithI18nKey(err, i18n.GitCreateExecParamError)
		}
		err = s.gitSvc.UpdateDefaultBranch(ctx, &git.UpdateDefaultBranchParams{
			WriteParams: writeParams,
			BranchName:  req.DefaultBranch,
		})
		if err != nil {
			return nil, errors.WithI18nKey(err, i18n.GitExecError)
		}
	}

	// 获取 space 信息用于构造 repoPath
	spaceInfo, err := s.spaceStore.GetByID(ctx, req.SpaceID)
	if err != nil {
		return nil, errors.NotFound.Newf("failed to find the parent space: %v", err)
	}

	// 构造更新的 repository 数据
	repoData := &repo.Repository{
		SpaceID:       req.SpaceID,
		Identifier:    req.Identifier,
		Desc:          req.Desc,
		ACL:           req.ACL,
		DefaultBranch: req.DefaultBranch,
	}

	// 调用 store 层更新
	updatedRepo, err := s.repoStore.Update(ctx, repoData, oldRepo.ID)
	if err != nil {
		return nil, err
	}

	// 构造 repoPath: spaceName/repoIdentifier
	repoPath := spaceInfo.Name + "/" + updatedRepo.Identifier

	// 回填 gitURL 和 gitSSHURL
	updatedRepo.GitURL = s.urlProvider.GenerateGITCloneURL(ctx, repoPath)
	updatedRepo.GitSSHURL = s.urlProvider.GenerateGITCloneSSHURL(ctx, repoPath)

	return updatedRepo, nil
}

func (s *Service) GetRepoCheckAccess(ctx context.Context, session *auth.Session, spaceID uint, identifier string, reqPermission []constant.Permission) (*repoModel.Repo, error) {

	// get repo
	repo, err := s.GetByIdentifier(ctx, spaceID, identifier, []constant.RepoState{constant.RepoStateActive, constant.RepoStateMigrateGitPush})
	if err != nil {
		return nil, err
	}

	// check permissions
	if err := s.checkRepoPermissions(ctx, session, repo, reqPermission); err != nil {
		return nil, err
	}

	return repo, nil
}

func (s *Service) GetRepoCheckAccessByID(ctx context.Context, session *auth.Session, repoID uint, reqPermission []constant.Permission) (*repoModel.Repo, error) {

	// get repo
	repo, err := s.GetByID(ctx, session, repoID)
	if err != nil {
		return nil, err
	}

	// check permissions
	if err := s.checkRepoPermissions(ctx, session, repo, reqPermission); err != nil {
		return nil, err
	}

	return repo, nil
}

// checkRepoPermissions 检查仓库权限的公共方法
// 遍历所有请求的权限,确保用户对该仓库拥有所有所需权限
func (s *Service) checkRepoPermissions(ctx context.Context, session *auth.Session, repo *repoModel.Repo, reqPermissions []constant.Permission) error {
	for _, permission := range reqPermissions {
		ok, err := s.authorizer.Check(ctx, session,
			constant.Scope{
				ResourceType: constant.ResourceTypeRepo,
				ResourceID:   repo.ID,
			},
			permission,
		)
		if err != nil {
			return err
		}

		if !ok {
			return errors.NoType.Newf("insufficient permissions for repo %s : %v", repo.Identifier, permission)
		}
	}

	return nil
}

func (s *Service) GetByIdentifier(ctx context.Context, spaceID uint, identifier string, allowedStates []constant.RepoState) (*repoModel.Repo, error) {
	repo, err := s.repoStore.GetBySpaceAndIdentifier(ctx, spaceID, identifier)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find repository: %s", identifier)
	}

	if len(allowedStates) > 0 && !slice.Contain(allowedStates, constant.RepoState(repo.Status)) {
		return nil, errors.BadRequest.New("Repository is not ready to use.")
	}
	// 回填 gitURL 和 gitSSHURL
	repo.GitURL = s.urlProvider.GenerateGITCloneURL(ctx, repo.Path)
	repo.GitSSHURL = s.urlProvider.GenerateGITCloneSSHURL(ctx, repo.Path)

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
	// 回填 gitURL 和 gitSSHURL
	repo.GitURL = s.urlProvider.GenerateGITCloneURL(ctx, repo.Path)
	repo.GitSSHURL = s.urlProvider.GenerateGITCloneSSHURL(ctx, repo.Path)

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

func (s *Service) Archive(ctx context.Context, params gitAPI.ArchiveParams, repo repoModel.Repo, w io.Writer) error {

	return s.gitSvc.Archive(ctx, git.ArchiveParams{
		ReadParams:    git.ReadParams{RepoUID: repo.GitUID},
		ArchiveParams: params,
	}, w)
}

func (s *Service) GetBlame(ctx context.Context, req repoModel.BlameFilter, repo repoModel.Repo, filePath string, w io.Writer) error {
	gitRef := req.GitRef
	if gitRef == "" {
		gitRef = repo.DefaultBranch
	}

	return s.gitSvc.RawBlame(ctx, w, &git.BlameParams{
		ReadParams: git.ReadParams{RepoUID: repo.GitUID},
		Path:       filePath,
		GitRef:     gitRef,
		LineFrom:   req.LineFrom,
		LineTo:     req.LineTo,
	})
}

func (s *Service) Paths(ctx context.Context, req repoModel.PathsInput, repo repoModel.Repo) (*repoModel.PathsOutput, error) {
	gitRef := req.GitRef
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
	gitRef := req.GitRef
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

// Delete 删除代码库软删除
func (s *Service) Delete(ctx context.Context, session *auth.Session, id uint) error {
	_, err := s.repoStore.GetByID(ctx, id)
	if err != nil {
		return err
	}
	// 调用 store 层删除 repository 记录
	err = s.repoStore.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

// List 分页查询仓库列表
func (s *Service) List(ctx context.Context, session *auth.Session, filter repoModel.ReqFilter) ([]*repoModel.Repo, bool, error) {
	// 调用 store 层查询
	pageResult, err := s.repoStore.List(ctx, filter)
	if err != nil {
		return nil, false, err
	}

	// 为每个仓库填充 Git URL
	for _, repoObj := range pageResult.Data {
		repoObj.GitSSHURL = s.urlProvider.GenerateGITCloneSSHURL(ctx, repoObj.Path)
		repoObj.GitURL = s.urlProvider.GenerateGITCloneURL(ctx, repoObj.Path)
	}

	// 计算是否为最后一页
	isLast := pageResult.Page*pageResult.PageSize >= int(pageResult.Total)

	return pageResult.Data, isLast, nil
}
