package repo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/easysoft/gitfox-community/app/githook"

	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/app/model/auth"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	"github.com/easysoft/gitfox-community/app/service/utils"
	providerStore "github.com/easysoft/gitfox-community/app/store/provider"
	"github.com/easysoft/gitfox-community/app/store/repo"
	"github.com/easysoft/gitfox-community/app/store/space"
	"github.com/easysoft/gitfox-community/pkg/auth/authz"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/git"
	gitAPI "github.com/easysoft/gitfox-community/pkg/git/api"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	svnPkg "github.com/easysoft/gitfox-community/pkg/svn"
	"github.com/easysoft/gitfox-community/pkg/utils/slice"
	"github.com/easysoft/gitfox-community/pkg/utils/url"
	"github.com/easysoft/gitfox-community/resources"
)

type Service struct {
	gitSvc        git.Interface
	svnSvc        svnPkg.Interface
	repoStore     repo.Interface
	spaceStore    space.Interface
	urlProvider   url.Provider
	authorizer    authz.Authorizer
	providerStore providerStore.Interface
}

var _ Interface = (*Service)(nil)

// NewRepoSvc 创建 RepoService 实例
func NewRepoSvc(
	gitSvc git.Interface,
	svnSvc svnPkg.Interface,
	repoStore repo.Interface,
	spaceStore space.Interface,
	urlProvider url.Provider,
	authorizer authz.Authorizer,
	providerStore providerStore.Interface,
) *Service {
	return &Service{
		gitSvc:        gitSvc,
		svnSvc:        svnSvc,
		repoStore:     repoStore,
		spaceStore:    spaceStore,
		urlProvider:   urlProvider,
		authorizer:    authorizer,
		providerStore: providerStore,
	}
}

// fillGitURLs 回填 GitURL 和 GitSSHURL。
// 对于 Mirror 仓库，通过 ProviderID 获取 provider 的 URL 并结合 Connector 中的 slug 拼接；
// 对于普通仓库，使用 repo.Path 生成 URL。
func (s *Service) fillGitURLs(ctx context.Context, repo *repoModel.Repo) {
	if repo.Mirror {
		provider, err := s.providerStore.GetByID(ctx, repo.ProviderID)
		if err != nil || provider == nil {
			return
		}

		// SVN 镜像仓库
		if repo.SCMType == constant.RepoScmTypeSvn {
			svnConnectInfo := repoModel.SVNConnectInfo{}
			if err := json.Unmarshal(repo.Connector, &svnConnectInfo); err != nil {
				return
			}
			repo.SvnURL = s.urlProvider.GenerateMirrorSVNCloneURL(ctx, provider.URL, svnConnectInfo.Slug)
			repo.GitURL = repo.SvnURL
			return
		}

		// Git 镜像仓库
		gitConnectInfo := repoModel.GitConnectInfo{}
		if err := json.Unmarshal(repo.Connector, &gitConnectInfo); err != nil {
			return
		}

		repo.GitURL = s.urlProvider.GenerateMirrorGITCloneURL(ctx, provider.URL, gitConnectInfo.Slug)
		repo.GitSSHURL = s.urlProvider.GenerateMirrorGITCloneSSHURL(ctx, provider.URL, gitConnectInfo.Slug)
		return
	}

	repo.GitURL = s.urlProvider.GenerateGITCloneURL(ctx, repo.Path)
	repo.GitSSHURL = s.urlProvider.GenerateGITCloneSSHURL(ctx, repo.Path)
}

func (s *Service) Create(ctx context.Context, session *auth.Session, params *repoModel.CreateRepoReq) (*repoModel.Repo, error) {
	_, err := s.spaceStore.GetByID(ctx, params.SpaceID)
	if err != nil {
		return nil, errors.WithI18nKey(nil, i18n.I18nKey(i18n.RepoSpaceNotFound))
	}

	gitResp, _, err := s.createGitRepository(ctx, params, session)
	if err != nil {
		return nil, err
	}

	repo := &repo.Repository{
		SpaceID:       params.SpaceID,
		Identifier:    params.Identifier,
		GitUID:        gitResp.UID,
		Desc:          params.Desc,
		Product:       params.Product,
		CreatedBy:     session.User.Account,
		ForkID:        params.ForkID,
		DefaultBranch: params.DefaultBranch,
		ACL:           params.ACL,
		Status:        constant.RepoStateActive, // 新创建的仓库直接设为 active 状态
	}

	repoObj, err := s.repoStore.Create(ctx, repo)
	if err != nil {
		return nil, err
	}
	s.fillGitURLs(ctx, repoObj)
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
			return nil, false, errors.WithI18nKey(nil, i18n.RepoLicenseError)
		}
		files = append(files, git.File{
			Path:    "LICENSE",
			Content: content,
		})
	}

	if req.Ignore != "" {
		content, err = resources.Gitignore.ReadFile(fmt.Sprintf("gitignore/%s.gitignore", req.Ignore))
		if err != nil {
			return nil, false, errors.WithI18nKey(err, i18n.RepoGitignoreInvalid)
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
		return nil, false, errors.WithI18nKey(err, i18n.RepoCreateError)
	}
	return resp, len(files) == 0, nil
}

func (s *Service) Update(ctx context.Context, req *repoModel.ReqForUpdate, oldRepo repoModel.Repo, session *auth.Session) (*repoModel.Repo, error) {

	if !oldRepo.Mirror {
		// 调用 git 命令更新默认分支
		req.DefaultBranch = utils.DecodeGitRef(req.DefaultBranch)
		if req.DefaultBranch != "" {
			if strings.HasPrefix(req.DefaultBranch, constant.BranchArchivedPrefix) {
				return nil, errors.WithI18nKey(
					fmt.Errorf("default branch cannot start with %s", constant.BranchArchivedPrefix),
					i18n.RepoDefaultBranchArchivedForbidden,
				)
			}
			writeParams, err := utils.CreateRPCInternalWriteParams(ctx, s.urlProvider, session, &oldRepo)
			if err != nil {
				return nil, err
			}
			err = s.gitSvc.UpdateDefaultBranch(ctx, &git.UpdateDefaultBranchParams{
				WriteParams: writeParams,
				BranchName:  req.DefaultBranch,
			})
			if err != nil {
				return nil, errors.WithI18nKey(err, i18n.GitExecError)
			}
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
		Desc:          req.Desc,
		ACL:           req.ACL,
		Product:       req.Product,
		Identifier:    req.Identifier,
		DefaultBranch: req.DefaultBranch,
	}

	// 调用 store 层更新
	updatedRepo, err := s.repoStore.Update(ctx, repoData, oldRepo.ID)
	if err != nil {
		return nil, err
	}

	// 构造 repoPath: spaceCode/repoIdentifier
	repoPath := spaceInfo.Code + "/" + updatedRepo.Identifier

	// 回填 gitURL 和 gitSSHURL
	if updatedRepo.Mirror {
		s.fillGitURLs(ctx, updatedRepo)
	} else {
		updatedRepo.GitURL = s.urlProvider.GenerateGITCloneURL(ctx, repoPath)
		updatedRepo.GitSSHURL = s.urlProvider.GenerateGITCloneSSHURL(ctx, repoPath)
	}

	return updatedRepo, nil
}

func (s *Service) GetRepoCheckAccess(ctx context.Context, session *auth.Session, spaceID uint, identifier string, reqPermission []constant.Permission) (*repoModel.Repo, error) {

	// get repo
	repo, err := s.GetByIdentifier(ctx, spaceID, identifier, []string{constant.RepoStateActive})
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

func (s *Service) GetByIdentifier(ctx context.Context, spaceID uint, identifier string, allowedStates []string) (*repoModel.Repo, error) {
	repo, err := s.repoStore.GetBySpaceAndIdentifier(ctx, spaceID, identifier)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find repository: %s", identifier)
	}

	if len(allowedStates) > 0 && !slice.Contain(allowedStates, repo.Status) {
		return nil, errors.BadRequest.New("Repository is not ready to use.")
	}
	// 回填 gitURL 和 gitSSHURL
	s.fillGitURLs(ctx, repo)

	return repo, nil
}

func (s *Service) GetByID(ctx context.Context, session *auth.Session, id uint) (*repoModel.Repo, error) {
	if id < 1 {
		return nil, errors.BadRequest.New("A valid repository id must be provided.")
	}

	repo, err := s.repoStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// 回填 gitURL 和 gitSSHURL
	s.fillGitURLs(ctx, repo)

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
	gitRef := utils.DecodeGitRef(req.GitRef)
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
	gitRef := utils.DecodeGitRef(req.GitRef)
	if gitRef == "" {
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
	gitRef := utils.DecodeGitRef(req.GitRef)
	if gitRef == "" {
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
		s.fillGitURLs(ctx, repoObj)
	}

	// 计算是否为最后一页
	isLast := pageResult.Page*pageResult.PageSize >= int(pageResult.Total)

	return pageResult.Data, isLast, nil
}
