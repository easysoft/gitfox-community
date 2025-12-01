package git

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/api"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

// GitServicePack 处理 SSH Git 操作 (receive-/upload-pack)
// repoRef 格式: "spaceID/repoName"，例如 "1/ldg-repo"
func (c *GitController) GitServicePack(
	ctx context.Context,
	session *auth.Session,
	repoRef string,
	options api.ServicePackOptions,
) error {
	log := logger.GetLogger(ctx)

	// 解析 repoRef: "1/ldg-repo" -> spaceID=1, repoName="ldg-repo"
	parts := strings.SplitN(repoRef, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository path format, expected: spaceID/repoName")
	}

	// 解析 spaceID
	spaceID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid spaceID: %s", parts[0])
	}

	repoName := strings.TrimSuffix(parts[1], ".git")

	log.Infof("SSH Git operation: spaceID=%d, repoName=%s, service=%s", spaceID, repoName, options.Service)

	// 根据操作类型确定所需权限
	permission := constant.PermissionRepoView
	if options.Service == constant.GitServiceTypeReceivePack {
		permission = constant.PermissionRepoPush
	}

	// 查询仓库并检查权限
	repo, err := c.repoSvc.GetRepoCheckAccess(ctx, session, uint(spaceID), repoName, []constant.Permission{permission})
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	// 调用 Git 服务执行操作
	err = c.gitSvc.GitServicePack(ctx, session, repo.GitUID, options)
	if err != nil {
		return fmt.Errorf("git service pack failed: %w", err)
	}

	return nil
}
