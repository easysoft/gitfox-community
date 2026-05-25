package artifact

import (
	"context"
	stderrors "errors"
	"strconv"
	"strings"

	artModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/model"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
)

// =============================================================================
// 恢复（Restore）
//
// 设计要点（与 SoftRemove 镜像）：
//   - 入参 entityIDs 复用 "kind.id" 串；五种 kind：artifact/group/package/version/asset
//   - 单条失败继续，不整体回滚；返回 ArtifactRestoreReport（结构与软删一致）
//   - 操作人取当前 session.User.Account，沿调用链传到 store 写入 editedBy
//   - 不判断回收站过期：DB 中 deleted=1 即可恢复；过期记录由 GC 物理清理后此处自然查不到
//   - 向上级联恢复祖先（parent group 链 → view），不向下级联（与软删对称）
//   - 恢复完成后同步重建 tree_node：用 artModel.NewSyncer 调对应 SyncTreeNodeFor*
//   - asset 的 blob 关联恢复：直接 blob.deleted=0；GC 仍由 active asset 引用计数判定
// =============================================================================

// Restore 批量恢复回收站中的实体。entityIDs 为空时直接返回空 report。
func (s *Service) Restore(ctx context.Context, restoreReq *model.ArtifactRestoreEntitiesReq) (*model.ArtifactRestoreReport, error) {
	if restoreReq == nil {
		restoreReq = &model.ArtifactRestoreEntitiesReq{}
	}
	ids := restoreReq.EntityIDs
	report := &model.ArtifactRestoreReport{Total: len(ids), Data: make([]*model.ArtifactRestoreRes, len(ids))}

	// 操作人：取当前会话用户作为「恢复人」（editedBy），匿名/系统调用时为空。
	var operator string
	if session, ok := req.GetAuthSession(ctx); ok && session != nil && session.User != nil {
		operator = session.User.Account
	}

	syncer := artModel.NewSyncer(s.artStore)

	for idx, eid := range ids {
		res := &model.ArtifactRestoreRes{NodeID: eid, Status: model.ArtifactStatusOK}
		report.Data[idx] = res
		if err := s.restoreEntity(ctx, eid, operator, syncer, res); err != nil {
			// 单条失败：保留 status，继续下一条；记日志便于排查。
			logger.GetLogger(ctx).Warnf("restore entity %s failed: %v", eid, err)
			if res.Status == model.ArtifactStatusOK {
				res.Status = model.ArtifactStatusUnknown
			}
		}
	}

	for _, item := range report.Data {
		if item.Status == model.ArtifactStatusOK {
			report.Success++
		} else {
			report.Failed++
		}
	}
	return report, nil
}

// restoreEntity 解析 entityID 并按 kind 分派到具体恢复函数。
func (s *Service) restoreEntity(
	ctx context.Context, eid string, operator string,
	syncer *artModel.Syncer, res *model.ArtifactRestoreRes,
) error {
	if strings.Count(eid, ".") != 1 {
		res.Status = model.ArtifactStatusInvalidId
		return pkgErrors.WithI18nKey(
			stderrors.New("invalid entity id format"),
			i18n.ArtifactRestoreEntityIdInvalidFormat,
		)
	}
	parts := strings.SplitN(eid, ".", 2)
	kind := model.ArtifactTreeNodeType(parts[0])
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		res.Status = model.ArtifactStatusInvalidId
		return pkgErrors.WithI18nKey(
			stderrors.New("invalid entity id number"),
			i18n.ArtifactRestoreEntityIdInvalidFormat,
		)
	}

	switch kind {
	case model.ArtifactTreeNodeTypeArtifact:
		return s.restoreArtifact(ctx, id, operator, syncer, res)
	case model.ArtifactTreeNodeTypeGroup:
		return s.restoreGroup(ctx, id, operator, syncer, res)
	case model.ArtifactTreeNodeTypePackage:
		return s.restorePackage(ctx, id, operator, syncer, res)
	case model.ArtifactTreeNodeTypeVersion:
		return s.restoreVersion(ctx, id, operator, syncer, res)
	case model.ArtifactTreeNodeTypeAsset:
		return s.restoreAsset(ctx, id, operator, syncer, res)
	default:
		res.Status = model.ArtifactStatusInvalidId
		return pkgErrors.WithI18nKey(
			stderrors.New("unsupported entity kind: "+string(kind)),
			i18n.ArtifactRestoreEntityTypeUnsupported,
		)
	}
}

// =============================================================================
// 单 kind 恢复
// =============================================================================

// restoreArtifact 恢复制品库（view）。若曾自动建过默认根 group（同名同 format / parentID=0），
// 一并恢复并重建 tree_node 根节点。view 软删时只动 view 自己 + 默认 group，
// 故无须级联恢复 user-defined group / package / version / asset。
func (s *Service) restoreArtifact(
	ctx context.Context, viewID int64, operator string,
	syncer *artModel.Syncer, res *model.ArtifactRestoreRes,
) error {
	// 查看 view 是否在回收站中
	view, err := s.artStore.Views().GetByIDWithDeleted(ctx, viewID)
	if err != nil {
		res.Status = model.ArtifactStatusNotFound
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreQueryError)
	}
	if view.Deleted != 1 {
		res.Status = model.ArtifactStatusOK
		return pkgErrors.WithI18nKey(
			stderrors.New("artifact is not in recycle bin"),
			i18n.ArtifactRestoreNotInRecycleBin,
		)
	}

	// 如果现在存在重名 view，提示错误
	if views, err := s.artStore.Views().FindByNameInScope(ctx, view.SpaceID, view.RepoID, view.Name); err == nil && len(views) > 0 {
		return pkgErrors.WithI18nKey(
			stderrors.New("artifact name already exists"),
			i18n.ArtifactRestoreViewNameAlreadyExists,
		)
	}

	// 恢复 view
	if err = s.artStore.Views().Restore(ctx, view.ID, operator); err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreError)
	}

	// 默认根 group：与 view 同 format / 同名 / parentID=0；软删时由 softRemoveArtifact 一并软删。
	defaultGroup, dErr := s.findDefaultRootGroup(ctx, view)
	if dErr == nil && defaultGroup != nil && defaultGroup.Deleted == 1 {
		if rErr := s.artStore.Groups().RestoreById(ctx, defaultGroup.ID, operator); rErr != nil {
			return pkgErrors.WithI18nKey(rErr, i18n.ArtifactRestoreError)
		}
		// 同步树节点（幂等：已有则复用）。重新读一次拿最新 deleted=0 的实例。
		if g, gErr := s.artStore.Groups().GetById(ctx, defaultGroup.ID); gErr == nil && g != nil {
			if _, sErr := syncer.SyncTreeNodeForGroup(ctx, g); sErr != nil {
				logger.GetLogger(ctx).Warnf("sync tree node for default group %d failed: %v", g.ID, sErr)
			}
		}
	}
	return nil
}

// restoreGroup 恢复分组及其向上链路（父 group → view）。不向下级联。
func (s *Service) restoreGroup(
	ctx context.Context, groupID int64, operator string,
	syncer *artModel.Syncer, res *model.ArtifactRestoreRes,
) error {
	group, err := s.artStore.Groups().GetById(ctx, groupID)
	if err != nil {
		res.Status = model.ArtifactStatusNotFound
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreQueryError)
	}
	if group.Deleted != 1 {
		res.Status = model.ArtifactStatusOK
		return pkgErrors.WithI18nKey(
			stderrors.New("group is not in recycle bin"),
			i18n.ArtifactRestoreNotInRecycleBin,
		)
	}
	if err = s.artStore.Groups().RestoreById(ctx, group.ID, operator); err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreError)
	}
	if err = s.cascadeRestoreAncestorGroups(ctx, group, operator); err != nil {
		return err
	}
	if err = s.cascadeRestoreView(ctx, group.ArtifactID, operator); err != nil {
		return err
	}
	// 重建 tree_node：从根 group 一路同步到当前 group。
	fresh, gErr := s.artStore.Groups().GetById(ctx, group.ID)
	if gErr == nil && fresh != nil {
		if _, sErr := syncer.SyncTreeNodeForGroup(ctx, fresh); sErr != nil {
			logger.GetLogger(ctx).Warnf("sync tree node for group %d failed: %v", fresh.ID, sErr)
		}
	}
	return nil
}

// restorePackage 恢复制品包及其向上链路（group 链 → view）。不向下级联 version。
func (s *Service) restorePackage(
	ctx context.Context, packageID int64, operator string,
	syncer *artModel.Syncer, res *model.ArtifactRestoreRes,
) error {
	pkg, err := s.artStore.Packages().GetByID(ctx, packageID)
	if err != nil {
		res.Status = model.ArtifactStatusNotFound
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreQueryError)
	}
	if !pkg.IsDeleted() {
		res.Status = model.ArtifactStatusOK
		return pkgErrors.WithI18nKey(
			stderrors.New("package is not in recycle bin"),
			i18n.ArtifactRestoreNotInRecycleBin,
		)
	}
	if err = s.artStore.Packages().Restore(ctx, pkg.ID, operator); err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreError)
	}
	res.Packages++
	if pkg.Format == model.ArtifactContainerFormat {
		// container package 只是同 scope 锚点，实际展示节点由 version.ArtifactLibID 决定。
		// 单独恢复 package 时不恢复锚点 view，也不重建 package tree node，避免出现空 package。
		return nil
	}
	// 向上恢复 group 链 + view。
	if err = s.cascadeRestoreGroupChainByID(ctx, pkg.GroupID, operator); err != nil {
		return err
	}
	if err = s.cascadeRestoreView(ctx, pkg.ArtifactID, operator); err != nil {
		return err
	}
	// 重建 tree_node：package 节点（其内部会按需补全祖先 group 节点）。
	fresh, pErr := s.artStore.Packages().GetByID(ctx, pkg.ID)
	if pErr == nil && fresh != nil {
		if _, sErr := syncer.SyncTreeNodeForPackage(ctx, fresh); sErr != nil {
			logger.GetLogger(ctx).Warnf("sync tree node for package %d failed: %v", fresh.ID, sErr)
		}
	}
	return nil
}

// restoreVersion 恢复版本及其向上链路（package + group 链 → view）。不向下级联 asset。
func (s *Service) restoreVersion(
	ctx context.Context, versionID int64, operator string,
	syncer *artModel.Syncer, res *model.ArtifactRestoreRes,
) error {
	ver, err := s.artStore.Versions().GetByID(ctx, versionID)
	if err != nil {
		res.Status = model.ArtifactStatusNotFound
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreQueryError)
	}
	if !ver.IsDeleted() {
		res.Status = model.ArtifactStatusOK
		return pkgErrors.WithI18nKey(
			stderrors.New("version is not in recycle bin"),
			i18n.ArtifactRestoreNotInRecycleBin,
		)
	}
	if err = s.artStore.Versions().Restore(ctx, ver.ID, operator); err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreError)
	}
	res.Versions++
	// 向上：父 package（必要时含其 group 链 + view）。
	pkg, pErr := s.artStore.Packages().GetByID(ctx, ver.PackageID)
	if pErr != nil {
		return pkgErrors.WithI18nKey(pErr, i18n.ArtifactRestoreQueryError)
	}
	if pkg.IsDeleted() {
		if rErr := s.artStore.Packages().Restore(ctx, pkg.ID, operator); rErr != nil {
			return pkgErrors.WithI18nKey(rErr, i18n.ArtifactRestoreError)
		}
		res.Packages++
	}
	if err = s.cascadeRestoreGroupChainByID(ctx, pkg.GroupID, operator); err != nil {
		return err
	}
	if err = s.cascadeRestoreView(ctx, artifactIDForVersion(pkg, ver), operator); err != nil {
		return err
	}
	// 重建 tree_node：version → package → group 链。
	freshVer, vErr := s.artStore.Versions().GetByID(ctx, ver.ID)
	freshPkg, pErr2 := s.artStore.Packages().GetByID(ctx, pkg.ID)
	if vErr == nil && pErr2 == nil && freshVer != nil && freshPkg != nil {
		if _, sErr := syncer.SyncTreeNodeForVersion(ctx, freshVer, freshPkg); sErr != nil {
			logger.GetLogger(ctx).Warnf("sync tree node for version %d failed: %v", freshVer.ID, sErr)
		}
	}
	return nil
}

// restoreAsset 恢复资产、关联 blob 及其向上链路（version/group → package/group 链 → view）。
func (s *Service) restoreAsset(
	ctx context.Context, assetID int64, operator string,
	syncer *artModel.Syncer, res *model.ArtifactRestoreRes,
) error {
	asset, err := s.artStore.Assets().GetById(ctx, assetID)
	if err != nil {
		res.Status = model.ArtifactStatusNotFound
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreQueryError)
	}
	if asset.Deleted != 1 {
		res.Status = model.ArtifactStatusOK
		return pkgErrors.WithI18nKey(
			stderrors.New("asset is not in recycle bin"),
			i18n.ArtifactRestoreNotInRecycleBin,
		)
	}
	if err = s.artStore.Assets().RestoreById(ctx, asset.ID, operator); err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreError)
	}
	res.Assets++
	// blob：关联恢复（不论引用计数）。GC 仍按 active asset 数判定。
	if asset.BlobID != 0 {
		if rErr := s.artStore.Blobs().RestoreById(ctx, asset.BlobID, operator); rErr != nil {
			return pkgErrors.WithI18nKey(rErr, i18n.ArtifactRestoreError)
		}
	}

	// 向上恢复（按归属类型分支）。
	switch {
	case asset.VersionID.Valid && asset.VersionID.Int64 > 0:
		if err = s.cascadeRestoreVersionChain(ctx, asset.VersionID.Int64, operator, syncer, res); err != nil {
			return err
		}
	case asset.GroupID.Valid && asset.GroupID.Int64 > 0:
		if err = s.cascadeRestoreGroupChainByID(ctx, asset.GroupID.Int64, operator); err != nil {
			return err
		}
		// 取 group.ArtifactID 用于恢复 view。
		if g, gErr := s.artStore.Groups().GetById(ctx, asset.GroupID.Int64); gErr == nil && g != nil {
			if err = s.cascadeRestoreView(ctx, g.ArtifactID, operator); err != nil {
				return err
			}
			// 重建 tree_node 链路（asset 不入 tree_node 表，但其祖先 group 需要）。
			if _, sErr := syncer.SyncTreeNodeForGroup(ctx, g); sErr != nil {
				logger.GetLogger(ctx).Warnf("sync tree node for group %d failed: %v", g.ID, sErr)
			}
		}
	case asset.ArtifactID.Valid && asset.ArtifactID.Int64 > 0:
		if err = s.cascadeRestoreView(ctx, asset.ArtifactID.Int64, operator); err != nil {
			return err
		}
	}
	return nil
}

// =============================================================================
// 级联恢复辅助
// =============================================================================

// cascadeRestoreAncestorGroups 沿 parentID 链向上恢复所有 deleted=1 的祖先 group。
func (s *Service) cascadeRestoreAncestorGroups(ctx context.Context, group *model.ArtifactGroup, operator string) error {
	if group == nil {
		return nil
	}
	parentID := int64(0)
	if group.ParentID.Valid {
		parentID = group.ParentID.Int64
	}
	for parentID > 0 {
		parent, err := s.artStore.Groups().GetById(ctx, parentID)
		if err != nil {
			return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreQueryError)
		}
		if parent.Deleted == 1 {
			if rErr := s.artStore.Groups().RestoreById(ctx, parent.ID, operator); rErr != nil {
				return pkgErrors.WithI18nKey(rErr, i18n.ArtifactRestoreError)
			}
		}
		if parent.ParentID.Valid {
			parentID = parent.ParentID.Int64
		} else {
			parentID = 0
		}
	}
	return nil
}

// cascadeRestoreGroupChainByID 给定起点 groupID（可能为 0），自下而上恢复整链。
// groupID 为 0 表示无 group 归属（如 raw 包未挂 group），直接返回。
func (s *Service) cascadeRestoreGroupChainByID(ctx context.Context, groupID int64, operator string) error {
	if groupID <= 0 {
		return nil
	}
	group, err := s.artStore.Groups().GetById(ctx, groupID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreQueryError)
	}
	if group.Deleted == 1 {
		if rErr := s.artStore.Groups().RestoreById(ctx, group.ID, operator); rErr != nil {
			return pkgErrors.WithI18nKey(rErr, i18n.ArtifactRestoreError)
		}
	}
	return s.cascadeRestoreAncestorGroups(ctx, group, operator)
}

// cascadeRestoreView 若 view 在回收站则恢复（含其默认根 group）。
func (s *Service) cascadeRestoreView(ctx context.Context, viewID int64, operator string) error {
	if viewID <= 0 {
		return nil
	}
	view, err := s.artStore.Views().GetByIDWithDeleted(ctx, viewID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreQueryError)
	}
	if view.Deleted != 1 {
		return nil
	}
	// 如果现在存在重名 view，提示错误
	if views, err := s.artStore.Views().FindByNameInScope(ctx, view.SpaceID, view.RepoID, view.Name); err == nil && len(views) > 0 {
		return pkgErrors.WithI18nKey(
			stderrors.New("artifact name already exists"),
			i18n.ArtifactRestoreViewNameAlreadyExists,
		)
	}

	if err = s.artStore.Views().Restore(ctx, view.ID, operator); err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreError)
	}
	// view 软删时一并软删的默认 group，恢复时也对称恢复。
	defaultGroup, dErr := s.findDefaultRootGroup(ctx, view)
	if dErr == nil && defaultGroup != nil && defaultGroup.Deleted == 1 {
		if rErr := s.artStore.Groups().RestoreById(ctx, defaultGroup.ID, operator); rErr != nil {
			return pkgErrors.WithI18nKey(rErr, i18n.ArtifactRestoreError)
		}
	}
	return nil
}

// cascadeRestoreVersionChain 自 versionID 起恢复 version → package → group 链 → view，
// 并重建 tree_node。供 restoreAsset 复用。
func (s *Service) cascadeRestoreVersionChain(
	ctx context.Context, versionID int64, operator string,
	syncer *artModel.Syncer, res *model.ArtifactRestoreRes,
) error {
	ver, err := s.artStore.Versions().GetByID(ctx, versionID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreQueryError)
	}
	if ver.IsDeleted() {
		if rErr := s.artStore.Versions().Restore(ctx, ver.ID, operator); rErr != nil {
			return pkgErrors.WithI18nKey(rErr, i18n.ArtifactRestoreError)
		}
		res.Versions++
	}
	pkg, pErr := s.artStore.Packages().GetByID(ctx, ver.PackageID)
	if pErr != nil {
		return pkgErrors.WithI18nKey(pErr, i18n.ArtifactRestoreQueryError)
	}
	if pkg.IsDeleted() {
		if rErr := s.artStore.Packages().Restore(ctx, pkg.ID, operator); rErr != nil {
			return pkgErrors.WithI18nKey(rErr, i18n.ArtifactRestoreError)
		}
		res.Packages++
	}
	if err = s.cascadeRestoreGroupChainByID(ctx, pkg.GroupID, operator); err != nil {
		return err
	}
	if err = s.cascadeRestoreView(ctx, artifactIDForVersion(pkg, ver), operator); err != nil {
		return err
	}
	// 重建 tree_node：取最新（已恢复）记录。
	freshVer, vErr := s.artStore.Versions().GetByID(ctx, ver.ID)
	freshPkg, pErr2 := s.artStore.Packages().GetByID(ctx, pkg.ID)
	if vErr == nil && pErr2 == nil && freshVer != nil && freshPkg != nil {
		if _, sErr := syncer.SyncTreeNodeForVersion(ctx, freshVer, freshPkg); sErr != nil {
			logger.GetLogger(ctx).Warnf("sync tree node for version %d failed: %v", freshVer.ID, sErr)
		}
	}
	return nil
}

// findDefaultRootGroup 在 view 所在格式下查找默认根 group（同 format / 同名 / parentID=0）。
// 默认根 group 由 CreateView 在特定格式下自动建立；恢复 view 时一并还原它。
// 直接走 store.Groups().FindRootByName（含已软删记录）。找不到返回 (nil, nil)。
func (s *Service) findDefaultRootGroup(ctx context.Context, view *model.ArtifactView) (*model.ArtifactGroup, error) {
	if view == nil {
		return nil, nil
	}
	return s.artStore.Groups().FindRootByName(ctx, view.ID, view.Name, view.Format)
}
