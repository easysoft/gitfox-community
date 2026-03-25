package fit

import (
	"context"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	repoSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/utils"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/token"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
)

// Service fit 服务实现
type Service struct {
	tokenStore   token.Interface
	devopsConfig config.DevopsConfig
	repoSvc      repoSvc.Interface
}

var _ Interface = (*Service)(nil)

// NewService 创建 fit 服务实例
func NewService(tokenStore token.Interface, devopsConfig config.DevopsConfig, repoSvc repoSvc.Interface) *Service {
	return &Service{
		tokenStore:   tokenStore,
		devopsConfig: devopsConfig,
		repoSvc:      repoSvc,
	}
}

// GetBranchPermission 从 tokenStore 获取 DevOps Token，再调用外部系统获取分支权限
func (s *Service) GetBranchPermission(ctx context.Context, repoID uint, branchName string) (*utils.BranchPermissionResult, error) {
	entry, err := s.tokenStore.GetDevopsToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get devops token: %w", err)
	}

	result, err := utils.GetBranchPermission(ctx, s.devopsConfig.URL, int64(repoID), branchName, entry.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch permission for repo %d branch %s: %w", repoID, branchName, err)
	}

	return result, nil
}

// GetRepoByName 根据空间 ID 和仓库名获取仓库信息（带权限校验）
func (s *Service) GetRepoByName(ctx context.Context, session *auth.Session, spaceID uint, repoName string) (*repoModel.Repo, error) {
	return s.repoSvc.GetRepoCheckAccess(ctx, session, spaceID, repoName, []constant.Permission{constant.PermissionRepoView})
}
