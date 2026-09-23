// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package artifact

import (
	"context"

	"github.com/easysoft/gitfox-community/app/artifact/adapter"
	pkgErrors "github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
)

// ResolveGlobalArtifact 全局作用域下按 artifact code 精确定位制品库及其容器 group。
// 对应路由：/v2/global/{artifactCode}/{imagename}
func (s *Service) ResolveGlobalArtifact(ctx context.Context, artifactCode string) (*ContainerArtifactCtx, error) {
	return s.resolveContainerArtifact(ctx, "global", 0, 0, artifactCode)
}

// ResolveSpaceArtifactByCode 空间作用域下按 (spaceCode, artifactCode) 精确定位制品库。
// 对应路由：/v2/{spaceCode}/{artifactCode}/{imagename}
func (s *Service) ResolveSpaceArtifactByCode(ctx context.Context, spaceCode, artifactCode string) (*ContainerArtifactCtx, error) {
	space, err := s.spaceStore.GetByCode(ctx, spaceCode)
	if err != nil {
		return nil, err
	}
	artCtx, err := s.resolveContainerArtifact(ctx, "space", int64(space.ID), 0, artifactCode)
	if err != nil {
		return nil, err
	}
	artCtx.SpaceCode = space.Code
	return artCtx, nil
}

// ResolveRepoArtifact 仓库作用域下按 (repoID, artifactCode) 精确定位制品库及其容器 group。
// 对应路由：/v2/repo{repoID}/{artifactCode}/{imagename}
// 与现有 ParseContainerView 行为保持一致：先通过 repoStore 按 ID 反查拿到 repo，
// 再以 (int64(repo.SpaceID), int64(repoID), artifactCode) 调用 resolveContainerArtifact。
// 不引入新复杂度，保留现有已验证过的查询方式。
func (s *Service) ResolveRepoArtifact(ctx context.Context, repoID uint64, artifactCode string) (*ContainerArtifactCtx, error) {
	// 通过 repo 反查 spaceID（与现有 ParseContainerView 保持一致，不改动此行为）
	repo, err := s.repoStore.GetByID(ctx, uint(repoID))
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactRepoQueryError)
	}
	return s.resolveContainerArtifact(ctx, "repo", int64(repo.SpaceID), int64(repoID), artifactCode)
}

// resolveContainerArtifact 容器路由的统一私有解析流程。
// 步骤 1：按 (spaceID, repoID, code) 在 ops_artifact 精确查找 view 记录；
//   - ErrResourceNotFound → 包装 i18n.ArtifactContainerArtifactNotFound
//   - 其它错误 → 包装 i18n.ArtifactLibQueryError
//
// 步骤 2：组装并返回 ContainerArtifactCtx（含 ViewDescriptor{ArtifactID, Store: s.fileStore, StorageID: 1}）。
func (s *Service) resolveContainerArtifact(ctx context.Context, scope string, spaceID, repoID int64, artifactCode string) (*ContainerArtifactCtx, error) {
	// 1) 按 (spaceID, repoID, code) 精确查找 artifact 视图
	view, err := s.artStore.Views().GetByCode(ctx, spaceID, repoID, artifactCode)
	if err != nil {
		if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
			return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactContainerArtifactNotFound)
		}
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactLibQueryError)
	}

	return &ContainerArtifactCtx{
		Scope:   scope,
		SpaceID: spaceID,
		RepoID:  repoID,
		View: &adapter.ViewDescriptor{
			ArtifactID: view.ID,
			Store:      s.fileStore,
			StorageID:  1,
		},
	}, nil
}
