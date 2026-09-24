// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package fit

import (
	"context"
	"fmt"

	"github.com/easysoft/gitfox-community/app/service/utils"
	"github.com/easysoft/gitfox-community/app/store/token"
	"github.com/easysoft/gitfox-community/pkg/config"
)

// Service fit 服务实现
type Service struct {
	tokenStore   token.Interface
	devopsConfig config.DevopsConfig
}

var _ Interface = (*Service)(nil)

// NewService 创建 fit 服务实例
func NewService(tokenStore token.Interface, devopsConfig config.DevopsConfig) *Service {
	return &Service{
		tokenStore:   tokenStore,
		devopsConfig: devopsConfig,
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
