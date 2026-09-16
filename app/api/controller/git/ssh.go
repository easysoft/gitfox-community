package git

import (
	"context"
	"strings"

	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/app/model/auth"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/git/api"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

// GitServicePack 处理 SSH Git 操作 (receive-/upload-pack)
// repoRef 格式: "spaceCode/repoName"，例如 "my_space/ldg-repo" 或 "my_space/org/project"（仓库名支持包含斜杠）
func (c *GitController) GitServicePack(
	ctx context.Context,
	session *auth.Session,
	repoRef string,
	options api.ServicePackOptions,
) error {
	log := logger.GetLogger(ctx)

	// 解析 repoRef: "my_space/ldg-repo" -> spaceCode="my_space", repoName="ldg-repo"
	// 或 "my_space/org/project" -> spaceCode="my_space", repoName="org/project"
	// 使用 SplitN 确保只在第一个 / 处分割，后续的 / 都属于 repoName
	parts := strings.SplitN(repoRef, "/", 2)
	if len(parts) != 2 {
		return errors.New("invalid repository path format, expected: spaceCode/repoName")
	}

	spaceCode := parts[0]
	repoName := strings.TrimSuffix(parts[1], ".git")

	log.Infof("SSH Git operation: spaceCode=%s, repoName=%s, service=%s", spaceCode, repoName, options.Service)

	// 通过 spaceCode 查询 space
	space, err := c.spaceSvc.GetByCode(ctx, spaceCode)
	if err != nil {
		return errors.Wrapf(err, "failed to find space by code: %s", spaceCode)
	}

	// 根据操作类型确定所需权限
	permission := constant.PermissionRepoView
	if options.Service == constant.GitServiceTypeReceivePack {
		permission = constant.PermissionRepoPush
	}

	// 查询仓库并检查权限
	repo, err := c.repoSvc.GetRepoCheckAccess(ctx, session, space.ID, repoName, []constant.Permission{permission})
	if err != nil {
		return errors.Wrapf(err, "failed to get repository: %v", err)
	}

	// 调用 Git 服务执行操作
	err = c.gitSvc.GitServicePack(ctx, session, repo, options)
	if err != nil {
		return errors.Wrapf(err, "git service pack failed: %v", err)
	}

	return nil
}
