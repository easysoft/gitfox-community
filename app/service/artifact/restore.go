// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package artifact

import (
	"context"
	stderrors "errors"

	artModel "github.com/easysoft/gitfox-community/app/artifact/model"
	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/app/model/artifact"
	model "github.com/easysoft/gitfox-community/app/model/artifact"
	pkgErrors "github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/easysoft/gitfox-community/pkg/logger"

	"github.com/easysoft/gitfox-community/app/api/utils/req"
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

// BatchRestore 批量恢复回收站中的实体。entityIDs 为空时直接返回空 report。
func (s *Service) BatchRestore(ctx context.Context, restoreReq *model.ArtifactEntitiesReq) (*model.ArtifactRestoreReport, error) {

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
		err := s.restoreEntity(ctx, eid, operator, syncer, res)
		if err != nil {
			logger.GetLogger(ctx).Warnf("restore entity %s failed: %v", eid, err)
			res.Error = err
			if key := pkgErrors.GetI18nKey(err); key != "" {
				res.Reason = i18n.TWithContext(ctx, string(key))
				report.Message += res.Reason + " "
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

// Restore 恢复回收站中的实体（entityID）
func (s *Service) Restore(ctx context.Context, restoreReq model.ArtifactEntitieReq) (*model.ArtifactRestoreRes, error) {
	// 操作人：取当前会话用户作为「恢复人」（editedBy），匿名/系统调用时为空。
	var operator string
	if session, ok := req.GetAuthSession(ctx); ok && session != nil && session.User != nil {
		operator = session.User.Account
	}

	syncer := artModel.NewSyncer(s.artStore)
	res := &model.ArtifactRestoreRes{NodeID: restoreReq.EntityID, Status: model.ArtifactStatusOK}
	err := s.restoreEntity(ctx, restoreReq.EntityID, operator, syncer, res)
	if err != nil {
		logger.GetLogger(ctx).Warnf("restore entity %s failed: %v", restoreReq.EntityID, err)
		res.Error = err
		if key := pkgErrors.GetI18nKey(err); key != "" {
			res.Reason = i18n.TWithContext(ctx, string(key))
		}
	}

	return res, nil
}

// restoreEntity 解析 entityID 并按 kind 分派到具体恢复函数。
func (s *Service) restoreEntity(ctx context.Context, entityID model.EntityID, operator string, syncer *artModel.Syncer, res *model.ArtifactRestoreRes) error {
	splitEntityID := entityID.GetEntityID()

	// 根据 kind 分派到具体恢复函数
	switch splitEntityID.Kind {
	case constant.ArtifactEntityKindArtifact:
		return s.restoreArtifact(ctx, splitEntityID.ID, operator, syncer, res)
	case constant.ArtifactEntityKindGroup:
		return s.restoreGroup(ctx, splitEntityID.ID, operator, syncer, res)
	case constant.ArtifactEntityKindPackage:
		return s.restorePackage(ctx, splitEntityID.ID, operator, syncer, res)
	case constant.ArtifactEntityKindVersion:
		return s.restoreVersion(ctx, splitEntityID.ID, operator, syncer, res)
	case constant.ArtifactEntityKindAsset:
		return s.restoreAsset(ctx, splitEntityID.ID, operator, syncer, res)
	default:
		res.Status = model.ArtifactStatusInvalidId
		return pkgErrors.WithI18nKey(
			stderrors.New("unsupported entity kind: "+string(splitEntityID.Kind)),
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
func (s *Service) restoreArtifact(ctx context.Context, viewID int64, operator string, syncer *artModel.Syncer, res *model.ArtifactRestoreRes) error {
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
		dupErr := pkgErrors.WithI18nKey(
			stderrors.New("artifact name already exists"),
			i18n.ArtifactRestoreViewNameAlreadyExists,
		)
		res.Status = model.ArtifactStatusNameAlreadyExists
		return dupErr
	}

	// 恢复 view
	if err = s.artStore.Views().Restore(ctx, view.ID, operator); err != nil {
		wrapped := pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreError)
		res.Status = model.ArtifactStatusUnknown
		return wrapped
	}

	// 默认根 group：与 view 同 format / 同名 / parentID=0；软删时由 softRemoveArtifact 一并软删。
	if artifact.ArtifactFormat(view.Format) == artifact.ArtifactFileFormat {
		if gErr := s.ensureDefaultGroupForView(ctx, view); gErr != nil {
			logger.GetLogger(ctx).Warnf("create default group for view %d (%s) failed: %v", view.ID, view.Name, gErr)
		}
	}

	return nil
}

// restoreGroup 恢复分组及其向上链路（父 group → view）。不向下级联。
// 若同位置（artifactLibID + parentID）已有同名活分组，直接拒绝恢复。
func (s *Service) restoreGroup(ctx context.Context, groupID int64, operator string, syncer *artModel.Syncer, res *model.ArtifactRestoreRes) error {
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

	parentID := int64(0)
	if group.ParentID.Valid {
		parentID = group.ParentID.Int64
	}
	if err = s.ensureGroupNameAvailable(ctx, group.ArtifactID, parentID, group.Format, group.Name); err != nil {
		res.Status = classifyRestoreErr(err)
		return err
	}

	// 恢复 group
	if err = s.artStore.Groups().RestoreById(ctx, group.ID, operator); err != nil {
		wrapped := pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreError)
		res.Status = model.ArtifactStatusUnknown
		return wrapped
	}
	// 恢复 group 链
	if err = s.cascadeRestoreAncestorGroups(ctx, group, operator); err != nil {
		res.Status = classifyRestoreErr(err)
		return err
	}
	// 恢复 view
	if err = s.cascadeRestoreView(ctx, group.ArtifactID, operator); err != nil {
		res.Status = classifyRestoreErr(err)
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
		wrapped := wrapRestoreDuplicateError(err, i18n.ArtifactRestorePackageNameAlreadyExists)
		res.Status = model.ArtifactStatusNameAlreadyExists
		return wrapped
	}
	res.Packages++
	if pkg.Format == model.ArtifactContainerFormat {
		// container package 只是同 scope 锚点，实际展示节点由 version.ArtifactLibID 决定。
		// 单独恢复 package 时不恢复锚点 view，也不重建 package tree node，避免出现空 package。
		return nil
	}
	// 向上恢复 group 链 + view。
	if err = s.cascadeRestoreGroupChainByID(ctx, pkg.GroupID, operator); err != nil {
		res.Status = classifyRestoreErr(err)
		return err
	}
	if err = s.cascadeRestoreView(ctx, pkg.ArtifactID, operator); err != nil {
		res.Status = classifyRestoreErr(err)
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
		wrapped := wrapRestoreDuplicateError(err, i18n.ArtifactRestoreVersionAlreadyExists)
		res.Status = model.ArtifactStatusNameAlreadyExists
		return wrapped
	}
	res.Versions++
	// 向上：父 package（必要时含其 group 链 + view）。
	pkg, pErr := s.artStore.Packages().GetByID(ctx, ver.PackageID)
	if pErr != nil {
		res.Status = model.ArtifactStatusNotFound
		return pkgErrors.WithI18nKey(pErr, i18n.ArtifactRestoreQueryError)
	}
	if pkg.IsDeleted() {
		if rErr := s.artStore.Packages().Restore(ctx, pkg.ID, operator); rErr != nil {
			wrapped := wrapRestoreDuplicateError(rErr, i18n.ArtifactRestorePackageNameAlreadyExists)
			res.Status = model.ArtifactStatusNameAlreadyExists
			return wrapped
		}
		res.Packages++
	}
	if err = s.cascadeRestoreGroupChainByID(ctx, pkg.GroupID, operator); err != nil {
		res.Status = classifyRestoreErr(err)
		return err
	}
	if err = s.cascadeRestoreView(ctx, artifactIDForVersion(pkg, ver), operator); err != nil {
		res.Status = classifyRestoreErr(err)
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
// 若同 (versionID/groupID, basename) 已有活资产，直接拒绝恢复。
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

	if err = s.ensureAssetNameAvailable(ctx, asset); err != nil {
		res.Status = classifyRestoreErr(err)
		return err
	}

	if err = s.artStore.Assets().RestoreById(ctx, asset.ID, operator); err != nil {
		wrapped := pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreError)
		res.Status = model.ArtifactStatusUnknown
		return wrapped
	}
	res.Assets++
	// blob：关联恢复（不论引用计数）。GC 仍按 active asset 数判定。
	if asset.BlobID != 0 {
		if rErr := s.artStore.Blobs().RestoreById(ctx, asset.BlobID, operator); rErr != nil {
			wrapped := pkgErrors.WithI18nKey(rErr, i18n.ArtifactRestoreError)
			res.Status = model.ArtifactStatusUnknown
			return wrapped
		}
	}

	// 向上恢复（按归属类型分支）。
	switch {
	case asset.VersionID.Valid && asset.VersionID.Int64 > 0:
		if err = s.cascadeRestoreVersionChain(ctx, asset.VersionID.Int64, operator, syncer, res); err != nil {
			res.Status = classifyRestoreErr(err)
			return err
		}
	case asset.GroupID.Valid && asset.GroupID.Int64 > 0:
		if err = s.cascadeRestoreGroupChainByID(ctx, asset.GroupID.Int64, operator); err != nil {
			res.Status = classifyRestoreErr(err)
			return err
		}
		// 取 group.ArtifactID 用于恢复 view。
		if g, gErr := s.artStore.Groups().GetById(ctx, asset.GroupID.Int64); gErr == nil && g != nil {
			if err = s.cascadeRestoreView(ctx, g.ArtifactID, operator); err != nil {
				res.Status = classifyRestoreErr(err)
				return err
			}
			// 重建 tree_node 链路（asset 不入 tree_node 表，但其祖先 group 需要）。
			if _, sErr := syncer.SyncTreeNodeForGroup(ctx, g); sErr != nil {
				logger.GetLogger(ctx).Warnf("sync tree node for group %d failed: %v", g.ID, sErr)
			}
		}
	case asset.ArtifactID.Valid && asset.ArtifactID.Int64 > 0:
		if err = s.cascadeRestoreView(ctx, asset.ArtifactID.Int64, operator); err != nil {
			res.Status = classifyRestoreErr(err)
			return err
		}
	}
	return nil
}

// =============================================================================
// 级联恢复辅助
// =============================================================================

// cascadeRestoreAncestorGroups 沿 parentID 链向上恢复所有 deleted=1 的祖先 group。
// 每恢复一节都先查活：撞名活节点时给被恢复 group 改名共存。
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
			grandParent := int64(0)
			if parent.ParentID.Valid {
				grandParent = parent.ParentID.Int64
			}
			if rErr := s.ensureGroupNameAvailable(ctx, parent.ArtifactID, grandParent, parent.Format, parent.Name); rErr != nil {
				return rErr
			}
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
// 起点节点撞名时直接拒绝恢复。
func (s *Service) cascadeRestoreGroupChainByID(ctx context.Context, groupID int64, operator string) error {
	if groupID <= 0 {
		return nil
	}
	group, err := s.artStore.Groups().GetById(ctx, groupID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreQueryError)
	}
	if group.Deleted == 1 {
		parentID := int64(0)
		if group.ParentID.Valid {
			parentID = group.ParentID.Int64
		}
		if rErr := s.ensureGroupNameAvailable(ctx, group.ArtifactID, parentID, group.Format, group.Name); rErr != nil {
			return rErr
		}
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

	if artifact.ArtifactFormat(view.Format) == artifact.ArtifactFileFormat {
		if gErr := s.ensureDefaultGroupForView(ctx, view); gErr != nil {
			logger.GetLogger(ctx).Warnf("create default group for view %d (%s) failed: %v", view.ID, view.Name, gErr)
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
			return wrapRestoreDuplicateError(rErr, i18n.ArtifactRestoreVersionAlreadyExists)
		}
		res.Versions++
	}
	pkg, pErr := s.artStore.Packages().GetByID(ctx, ver.PackageID)
	if pErr != nil {
		return pkgErrors.WithI18nKey(pErr, i18n.ArtifactRestoreQueryError)
	}
	if pkg.IsDeleted() {
		if rErr := s.artStore.Packages().Restore(ctx, pkg.ID, operator); rErr != nil {
			return wrapRestoreDuplicateError(rErr, i18n.ArtifactRestorePackageNameAlreadyExists)
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

// =============================================================================
// 重名拒绝：恢复时若同位置已有活节点，直接报错
// =============================================================================

// ensureGroupNameAvailable 检查给定位置（artifactLibID + parentID + format + name）是否已有活分组。
//
// parentID == 0 表示根分组，按 FindRootByName 取出再判 Deleted；
// parentID > 0 走 ListChildren（含软删）后内存过滤出活子节点。
// 命中活分组时返回 ArtifactRestoreGroupNameAlreadyExists。
func (s *Service) ensureGroupNameAvailable(ctx context.Context, artifactLibID int64, parentID int64, format model.ArtifactFormat, name string) error {
	if parentID > 0 {
		children, err := s.artStore.Groups().ListChildren(ctx, parentID)
		if err != nil {
			return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreQueryError)
		}
		for _, c := range children {
			if c == nil {
				continue
			}
			if c.Name == name && c.Deleted == 0 {
				return pkgErrors.WithI18nKey(
					stderrors.New("group name already exists at parent: "+name),
					i18n.ArtifactRestoreGroupNameAlreadyExists,
				)
			}
		}
		return nil
	}
	root, err := s.artStore.Groups().FindRootByName(ctx, artifactLibID, name, format)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreQueryError)
	}
	if root != nil && root.Deleted == 0 {
		return pkgErrors.WithI18nKey(
			stderrors.New("root group name already exists: "+name),
			i18n.ArtifactRestoreGroupNameAlreadyExists,
		)
	}
	return nil
}

// ensureAssetNameAvailable 检查 asset 在 (versionID/groupID, basename) 上是否已有活资产。
// 仅按 versionID / groupID 单一归属判定；asset.ArtifactID 归属（meta 资产）跳过查重。
// 命中时返回 ArtifactRestoreAssetNameAlreadyExists。
func (s *Service) ensureAssetNameAvailable(ctx context.Context, asset *model.ArtifactAsset) error {
	if asset == nil || asset.Path == "" {
		return nil
	}
	var (
		row *model.ArtifactAsset
		err error
	)
	switch {
	case asset.VersionID.Valid && asset.VersionID.Int64 > 0:
		row, err = s.artStore.Assets().GetVersionAsset(ctx, asset.Path, asset.VersionID.Int64)
	case asset.GroupID.Valid && asset.GroupID.Int64 > 0:
		row, err = s.artStore.Assets().GetGroupAsset(ctx, asset.Path, asset.GroupID.Int64)
	default:
		return nil
	}
	if err != nil {
		if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
			return nil
		}
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreQueryError)
	}
	// 命中的若是恢复目标本身（理论上不会，因为 GetVersionAsset/GetGroupAsset 已过 deleted=0），跳过。
	if row != nil && row.ID == asset.ID {
		return nil
	}
	return pkgErrors.WithI18nKey(
		stderrors.New("asset name already exists at location: "+asset.Path),
		i18n.ArtifactRestoreAssetNameAlreadyExists,
	)
}

// wrapRestoreDuplicateError 将 store 层抛出的唯一约束冲突包装成业务 i18n key。
// 非唯一冲突原样返回；nil 直接返回。
func wrapRestoreDuplicateError(err error, dupKey i18n.I18nKey) error {
	if err == nil {
		return nil
	}
	if pkgErrors.Is(err, pkgErrors.ErrDuplicate) {
		return pkgErrors.WithI18nKey(err, dupKey)
	}
	return pkgErrors.WithI18nKey(err, i18n.ArtifactRestoreError)
}

// classifyRestoreErr 按 err 携带的 i18n key 分类 restore 链路状态：
// 重名 key → NameAlreadyExists；其余 → Unknown。
func classifyRestoreErr(err error) model.ArtifactStatus {
	if err == nil {
		return model.ArtifactStatusOK
	}
	switch pkgErrors.GetI18nKey(err) {
	case i18n.ArtifactRestoreViewNameAlreadyExists,
		i18n.ArtifactRestoreGroupNameAlreadyExists,
		i18n.ArtifactRestoreAssetNameAlreadyExists,
		i18n.ArtifactRestorePackageNameAlreadyExists,
		i18n.ArtifactRestoreVersionAlreadyExists:
		return model.ArtifactStatusNameAlreadyExists
	}
	return model.ArtifactStatusUnknown
}
