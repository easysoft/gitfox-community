// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package artifact

import (
	"context"
	"fmt"

	"github.com/easysoft/gitfox-community/app/constant"
	model "github.com/easysoft/gitfox-community/app/model/artifact"
	artifactStore "github.com/easysoft/gitfox-community/app/store/artifact"
	pkgErrors "github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

// RelocateEntity 同时支持移动 + 重命名，单事务内先移动后改名。
//
// 语义：
//   - TargetGroupID 未传 → 不移动，即使 TargetArtifactID 有值也只执行重命名；
//   - TargetGroupID 已传，且 TargetArtifactID == 0 或与当前一致 → 同库移动；
//   - TargetGroupID 显式传 0 → 移动到目标制品库根目录（仅 generic group 支持）；
//   - NewName == "" 或与当前一致 → 不改名；
//   - 两者都不变 → 返回 ArtifactRelocateNoChange。
func (s *Service) RelocateEntity(ctx context.Context, req *model.RelocateEntityReq) error {
	splitEntityID := req.EntityID.GetEntityID()

	// 取当前实体的 artifactID / groupID / name，用于与传入参数比对，判断是否真有变化。
	var curArtifactID int64
	var curGroupID int64
	var curName string
	var curFormat model.ArtifactFormat

	switch splitEntityID.Kind {
	case constant.ArtifactEntityKindGroup:
		group, err := s.artStore.Groups().GetById(ctx, splitEntityID.ID)
		if err != nil {
			return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveSourceDeleted)
		}
		curArtifactID = group.ArtifactID
		if group.ParentID.Valid {
			curGroupID = group.ParentID.Int64
		}
		curName = group.Name
		curFormat = group.Format
	case constant.ArtifactEntityKindAsset:
		asset, err := s.artStore.Assets().GetById(ctx, splitEntityID.ID)
		if err != nil {
			return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveSourceDeleted)
		}
		if asset.ArtifactID.Valid {
			curArtifactID = asset.ArtifactID.Int64
		}
		if asset.GroupID.Valid {
			curGroupID = asset.GroupID.Int64
		}
		curName = asset.Path
		curFormat = asset.Format
	case constant.ArtifactEntityKindPackage:
		pack, err := s.artStore.Packages().GetByID(ctx, splitEntityID.ID)
		if err != nil {
			return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveSourceDeleted)
		}
		curArtifactID = pack.ArtifactID
		curGroupID = pack.GroupID
		curName = pack.Name
		curFormat = pack.Format
	case constant.ArtifactEntityKindArtifact:
		version, err := s.artStore.Views().GetByID(ctx, splitEntityID.ID)
		if err != nil {
			return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveSourceDeleted)
		}
		curArtifactID = version.ID
		curName = version.Name
		curFormat = version.Format
	default:
		return pkgErrors.WithI18nKey(
			fmt.Errorf("unsupported entity type: %s", splitEntityID.Kind),
			i18n.ArtifactMoveEntityTypeInvalid,
		)
	}

	// 判断是否需要移动：targetGroupID 是移动意图的开关。
	// 未传 targetGroupID 时，即使 targetArtifactID 有值也不移动，保证纯重命名请求不会被拦截。
	tArtifactID := req.TargetArtifactID
	if tArtifactID == 0 {
		tArtifactID = curArtifactID
	}
	var targetGroupID int64
	targetGroupIDSpecified := req.TargetGroupID != nil
	if targetGroupIDSpecified {
		targetGroupID = *req.TargetGroupID
	}
	needMove := targetGroupIDSpecified && (tArtifactID != curArtifactID || targetGroupID != curGroupID)

	// 判断是否需要重命名（NewName 为空或与当前一致 → 不改名）
	needRename := req.NewName != "" && req.NewName != curName

	if !needMove && !needRename {
		return nil
	}
	if curFormat == model.ArtifactContainerFormat {
		if needMove {
			return pkgErrors.WithI18nKey(
				fmt.Errorf("container artifacts do not support move"),
				i18n.ArtifactMoveContainerNotSupported,
			)
		}
		if needRename {
			return pkgErrors.WithI18nKey(
				fmt.Errorf("container artifacts do not support rename"),
				i18n.ArtifactRenameContainerNotSupported,
			)
		}
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if needMove {
			// 移动时若未显式指定 targetGroupID，沿用当前父级（保持层级不变）。
			// 显式指定 0 则表示移动到目标制品库根目录，由下层按实体类型和格式校验是否支持。
			if !targetGroupIDSpecified {
				targetGroupID = curGroupID
			}
			if err := s.moveByKind(ctx, tx, splitEntityID, tArtifactID, targetGroupID); err != nil {
				return err
			}
		}
		if needRename {
			if err := s.renameByKind(ctx, tx, splitEntityID, req.NewName); err != nil {
				return err
			}
		}
		return nil
	})
}

// moveByKind 按实体类型派发到对应的 Move 方法（不带事务，由调用方包裹）。
// tx 由 RelocateEntity 的 Transaction 回调传入，向下透传给各 inner，
// 让 store 写操作复用同一事务连接。
func (s *Service) moveByKind(ctx context.Context, tx *gorm.DB, splitEntityID model.SplitEntityID, tArtifactID, tGroupID int64) error {
	switch splitEntityID.Kind {
	case constant.ArtifactEntityKindPackage:
		return s.movePackageInner(ctx, tx, splitEntityID.ID, tArtifactID, tGroupID)
	case constant.ArtifactEntityKindAsset:
		return s.moveAssetInner(ctx, tx, splitEntityID.ID, tArtifactID, tGroupID)
	case constant.ArtifactEntityKindGroup:
		return s.moveGroupInner(ctx, tx, splitEntityID.ID, tArtifactID, tGroupID)
	default:
		return pkgErrors.WithI18nKey(
			fmt.Errorf("unsupported entity type for move: %s", splitEntityID.Kind),
			i18n.ArtifactMoveEntityTypeInvalid,
		)
	}
}

// renameByKind 按实体类型派发到对应的 Rename 方法（不带事务，由调用方包裹）。
// package / version 不允许重命名；container 格式的制品库也暂不允许重命名。
func (s *Service) renameByKind(ctx context.Context, tx *gorm.DB, splitEntityID model.SplitEntityID, newName string) error {
	switch splitEntityID.Kind {
	case constant.ArtifactEntityKindAsset:
		return s.renameAssetInner(ctx, tx, splitEntityID.ID, newName)
	case constant.ArtifactEntityKindGroup:
		return s.renameGroupInner(ctx, tx, splitEntityID.ID, newName)
	case constant.ArtifactEntityKindArtifact:
		return s.renameArtifactInner(ctx, tx, splitEntityID.ID, newName)
	case constant.ArtifactEntityKindPackage:
		return pkgErrors.WithI18nKey(
			fmt.Errorf("rename is not supported for packages"),
			i18n.ArtifactRenameVersionedNotSupported,
		)
	case constant.ArtifactEntityKindVersion:
		return pkgErrors.WithI18nKey(
			fmt.Errorf("rename is not supported for versions"),
			i18n.ArtifactRenameVersionedNotSupported,
		)
	default:
		return pkgErrors.WithI18nKey(
			fmt.Errorf("unsupported entity type for rename: %s", splitEntityID.Kind),
			i18n.ArtifactRenameEntityTypeInvalid,
		)
	}
}

// movePackageInner 在单事务（由调用方包裹）内完成一次 package 移动。
// 语义：更新 package.ArtifactID / GroupID / Namespace，并重建 tree node。
//
//   - container 格式：忽略传入的 tGroupID，直接落到目标 view 的 format 根节点。
//   - 非 container 格式：目标 group 必须属于目标 view，且格式要与源 package 一致。
func (s *Service) movePackageInner(ctx context.Context, tx *gorm.DB, pkgID, tArtifactID, tGroupID int64) error {
	txStore := artifactStore.NewStore(tx)

	// 1. 读源 package
	srcPkg, err := txStore.Packages().GetByID(ctx, pkgID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveSourceDeleted)
	}
	if srcPkg.IsDeleted() {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("package %d is soft-deleted", pkgID),
			i18n.ArtifactMoveSourceDeleted,
		)
	}

	// 2. 读源/目标 view，校验 format 一致（不允许跨格式制品库）
	srcView, err := txStore.Views().GetByID(ctx, srcPkg.ArtifactID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveTargetInvalid)
	}
	tView, err := txStore.Views().GetByID(ctx, tArtifactID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveTargetInvalid)
	}

	// 3. 源/目标制品库 format 必须严格一致
	if srcView.Format != tView.Format {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("source view format %s does not match target view format %s", srcView.Format, tView.Format),
			i18n.ArtifactMoveFormatMismatch,
		)
	}

	// 4. 解析目标 group 与 namespace
	var targetGroupID int64
	var namespace string
	switch srcPkg.Format {
	case model.ArtifactContainerFormat:
		// container：package 不再挂 group，version.ArtifactLibID 决定展示在哪个制品库。
		targetGroupID = 0
		namespace = ""
	default:
		if tGroupID <= 0 {
			return pkgErrors.WithI18nKey(
				fmt.Errorf("targetGroupID is required for package move"),
				i18n.ArtifactMoveTargetGroupRequired,
			)
		}
		endpoint, err := txStore.Groups().GetEndpointById(ctx, tGroupID)
		if err != nil {
			return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveTargetInvalid)
		}
		// 目标 group 必须属于目标 view 且格式一致
		if endpoint.ArtifactID != tArtifactID || endpoint.Format != srcPkg.Format {
			return pkgErrors.WithI18nKey(
				fmt.Errorf("target group %d does not belong to artifact %d or format mismatch", tGroupID, tArtifactID),
				i18n.ArtifactMoveTargetInvalid,
			)
		}
		targetGroupID = tGroupID
		namespace = endpoint.GroupName()
	}

	// 5. 同级冲突检查
	var existing *model.ArtifactPackage
	if srcPkg.Format == model.ArtifactContainerFormat {
		existing, err = txStore.Packages().GetByNameInArtifactScope(ctx, srcPkg.Name, namespace, tArtifactID, srcPkg.Format)
	} else {
		existing, err = txStore.Packages().GetByName(ctx, srcPkg.Name, namespace, tArtifactID, srcPkg.Format)
	}
	if err == nil && existing.ID != srcPkg.ID {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("package %q already exists at target", srcPkg.Name),
			i18n.ArtifactMoveSiblingConflict,
		)
	} else if err != nil && !pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactPackageQueryError)
	}

	// 6. 主表更新
	oldArtifactID := srcPkg.ArtifactID
	srcPkg.ArtifactID = tArtifactID
	srcPkg.GroupID = targetGroupID
	srcPkg.Namespace = namespace
	if err = txStore.Packages().Update(ctx, srcPkg); err != nil {
		return err
	}

	// 6.1 同步更新该 package 下所有 version 的 asset 的 artifactID
	versions, err := txStore.Versions().Find(ctx, model.SearchVersionOption{PackageID: srcPkg.ID})
	if err != nil {
		return err
	}
	for _, ver := range versions {
		ver.ArtifactLibID = tArtifactID
		if err = txStore.Versions().Update(ctx, ver); err != nil {
			return err
		}
		assets, err := txStore.Assets().Search(ctx, model.SearchAssetOption{VersionID: ver.ID})
		if err != nil {
			return err
		}
		for _, asset := range assets {
			asset.ArtifactID = null.IntFrom(tArtifactID)
			asset.GroupID = null.IntFrom(targetGroupID)
			if err = txStore.Assets().Update(ctx, asset); err != nil {
				return err
			}
		}
	}

	// 7. tree node 同步：清掉源侧 package 子树，再调用 RefreshTree 重建。
	if node, e := txStore.Nodes().GetByTarget(ctx, oldArtifactID, model.ArtifactTreeNodeTypePackage, srcPkg.ID); e == nil && node != nil {
		if err = txStore.Nodes().RecurseDeleteById(ctx, node.ID); err != nil {
			return err
		}
	}
	return s.refreshTree(ctx, string(srcPkg.Format), txStore)
}

// moveAssetInner 在调用方事务内完成一次组资产移动。
// 仅支持组资产（asset.GroupID 有效、versionID 无效）；版本资产随 package / version 移动，不在本方法范畴。
func (s *Service) moveAssetInner(ctx context.Context, tx *gorm.DB, assetID, tArtifactID, tGroupID int64) error {
	txStore := artifactStore.NewStore(tx)

	// 1. 读源 asset
	srcAsset, err := txStore.Assets().GetById(ctx, assetID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveSourceDeleted)
	}
	if srcAsset.Deleted == 1 {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("asset %d is soft-deleted", assetID),
			i18n.ArtifactMoveSourceDeleted,
		)
	}
	// 仅接受组资产：GroupID 必填、不得挂在 version 下
	if srcAsset.Format == model.ArtifactContainerFormat {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("asset %d is not a group asset", assetID),
			i18n.ArtifactMoveEntityTypeInvalid,
		)
	}

	// 2. 读源/目标 view，校验 format 严格一致
	if !srcAsset.ArtifactID.Valid || srcAsset.ArtifactID.Int64 == 0 {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("asset %d has no associated artifactID", assetID),
			i18n.ArtifactMoveSourceDeleted,
		)
	}
	srcView, err := txStore.Views().GetByID(ctx, srcAsset.ArtifactID.Int64)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveTargetInvalid)
	}
	tView, err := txStore.Views().GetByID(ctx, tArtifactID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveTargetInvalid)
	}
	// 3. 源/目标制品库 format 必须严格一致
	if srcView.Format != tView.Format {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("source view format %s does not match target view format %s", srcView.Format, tView.Format),
			i18n.ArtifactMoveFormatMismatch,
		)
	}

	// 4. 目标 group 必须有效，且属于目标 view、格式一致
	if tGroupID <= 0 {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("targetGroupID is required for group asset move"),
			i18n.ArtifactMoveTargetInvalid,
		)
	}
	tGroup, err := txStore.Groups().GetById(ctx, tGroupID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveTargetInvalid)
	}
	if tGroup.ArtifactID != tArtifactID || tGroup.Format != srcAsset.Format {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("target group %d does not belong to artifact %d or format mismatch", tGroupID, tArtifactID),
			i18n.ArtifactMoveTargetInvalid,
		)
	}

	// 5. 同级冲突：同一 group 下同名文件
	if _, err = txStore.Assets().GetGroupAsset(ctx, srcAsset.Path, tGroupID); err == nil {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("asset %q already exists under target group", srcAsset.Path),
			i18n.ArtifactMoveSiblingConflict,
		)
	} else if !pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactAssetQueryError)
	}

	// 6. 主表更新
	srcAsset.ArtifactID = null.IntFrom(tArtifactID)
	srcAsset.GroupID = null.IntFrom(tGroupID)
	return txStore.Assets().Update(ctx, srcAsset)
}

// moveGroupInner 在调用方事务内完成一次 group 移动。
// 语义：更新 group.ArtifactID / ParentID，跨 artifact 时子树内所有 group / asset / package 的 artifactID 一起改。
func (s *Service) moveGroupInner(ctx context.Context, tx *gorm.DB, groupID, tArtifactID, tParentGroupID int64) error {
	txStore := artifactStore.NewStore(tx)

	// 1. 读源 group
	srcGroup, err := txStore.Groups().GetById(ctx, groupID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveSourceDeleted)
	}
	if srcGroup.Deleted == 1 {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("group %d is soft-deleted", groupID),
			i18n.ArtifactMoveSourceDeleted,
		)
	}
	// 版本化格式不允许移动
	if model.IsVersionedFormat(srcGroup.Format) {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("versioned format %s group cannot be moved", srcGroup.Format),
			i18n.ArtifactMoveFormatMismatch,
		)
	}
	if tParentGroupID == 0 && srcGroup.Format != model.ArtifactFileFormat {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("only file groups can be moved to root, got format %s", srcGroup.Format),
			i18n.ArtifactMoveTargetInvalid,
		)
	}

	// 2. 读源/目标 view，校验 format 严格一致
	srcView, err := txStore.Views().GetByID(ctx, srcGroup.ArtifactID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveTargetInvalid)
	}
	tView, err := txStore.Views().GetByID(ctx, tArtifactID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveTargetInvalid)
	}
	if srcView.Format != tView.Format {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("source view format %s does not match target view format %s", srcView.Format, tView.Format),
			i18n.ArtifactMoveFormatMismatch,
		)
	}

	// 3. 循环引用检查：目标父组不能是源组本身或其后代
	if tParentGroupID > 0 {
		if tParentGroupID == groupID {
			return pkgErrors.WithI18nKey(
				fmt.Errorf("cannot move group into itself"),
				i18n.ArtifactMoveCycleDetected,
			)
		}
		descendants, _ := txStore.Groups().ListDescendants(ctx, groupID)
		for _, d := range descendants {
			if d.ID == tParentGroupID {
				return pkgErrors.WithI18nKey(
					fmt.Errorf("target parent %d is a descendant of source group %d", tParentGroupID, groupID),
					i18n.ArtifactMoveCycleDetected,
				)
			}
		}
		// 验证目标父组存在且属于目标 view
		tParent, err := txStore.Groups().GetById(ctx, tParentGroupID)
		if err != nil {
			return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveTargetInvalid)
		}
		if tParent.ArtifactID != tArtifactID {
			return pkgErrors.WithI18nKey(
				fmt.Errorf("target parent group %d does not belong to artifact %d", tParentGroupID, tArtifactID),
				i18n.ArtifactMoveTargetInvalid,
			)
		}
	}

	// 4. 同级冲突
	if tParentGroupID > 0 {
		if _, err = txStore.Groups().GetSubNode(ctx, tParentGroupID, srcGroup.Name); err == nil {
			return pkgErrors.WithI18nKey(
				fmt.Errorf("group %q already exists under target parent", srcGroup.Name),
				i18n.ArtifactMoveSiblingConflict,
			)
		}
	} else {
		if _, err = txStore.Groups().GetRoot(ctx, tArtifactID, srcGroup.Name, srcGroup.Format); err == nil {
			return pkgErrors.WithI18nKey(
				fmt.Errorf("root group %q already exists in target artifact", srcGroup.Name),
				i18n.ArtifactMoveSiblingConflict,
			)
		}
	}

	// 5. 主表更新：txStore 已绑定事务连接，后续 store 调用都会落在同一事务内。
	oldArtifactID := srcGroup.ArtifactID
	if mErr := txStore.Groups().MoveLocation(ctx, groupID, tArtifactID, tParentGroupID); mErr != nil {
		if pkgErrors.Is(mErr, pkgErrors.ErrResourceNotFound) {
			return pkgErrors.WithI18nKey(mErr, i18n.ArtifactMoveSourceDeleted)
		}
		return pkgErrors.WithI18nKey(mErr, i18n.ArtifactMoveTargetInvalid)
	}

	// 6. 跨 artifact 时，子树内所有 group / asset / package 的 artifactLibID 一起改。
	//    Groups / Assets / Packages 三个 store 都暴露 MoveArtifactLibByIDs / ByGroupIDs，
	//    service 不再直接拼 SQL。
	if oldArtifactID != tArtifactID {
		descendants, _ := txStore.Groups().ListDescendants(ctx, groupID)
		subtreeIDs := []int64{groupID}
		for _, d := range descendants {
			subtreeIDs = append(subtreeIDs, d.ID)
		}
		if mErr := txStore.Groups().MoveArtifactLibByIDs(ctx, subtreeIDs, tArtifactID); mErr != nil {
			return pkgErrors.WithI18nKey(mErr, i18n.ArtifactMoveTargetInvalid)
		}
		if mErr := txStore.Assets().MoveArtifactLibByGroupIDs(ctx, subtreeIDs, tArtifactID); mErr != nil {
			return pkgErrors.WithI18nKey(mErr, i18n.ArtifactMoveTargetInvalid)
		}
		if mErr := txStore.Packages().MoveArtifactLibByGroupIDs(ctx, subtreeIDs, tArtifactID); mErr != nil {
			return pkgErrors.WithI18nKey(mErr, i18n.ArtifactMoveTargetInvalid)
		}
	}

	// 7. tree 同步：删源侧 tree node，重建
	if node, e := txStore.Nodes().GetByTarget(ctx, oldArtifactID, model.ArtifactTreeNodeTypeGroup, groupID); e == nil && node != nil {
		if err = txStore.Nodes().RecurseDeleteById(ctx, node.ID); err != nil {
			return err
		}
	}
	return s.refreshTree(ctx, string(srcGroup.Format), txStore)
}

// renameAssetInner 在调用方事务内完成一次组资产重命名。
func (s *Service) renameAssetInner(ctx context.Context, tx *gorm.DB, assetID int64, newName string) error {
	txStore := artifactStore.NewStore(tx)

	// 1. 读源 asset
	srcAsset, err := txStore.Assets().GetById(ctx, assetID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRenameSourceDeleted)
	}
	if srcAsset.Deleted == 1 {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("asset %d is soft-deleted", assetID),
			i18n.ArtifactRenameSourceDeleted,
		)
	}
	if srcAsset.Format == model.ArtifactContainerFormat {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("container asset %d cannot be renamed", assetID),
			i18n.ArtifactRenameEntityTypeInvalid,
		)
	}
	// 仅组资产
	if !srcAsset.GroupID.Valid || srcAsset.GroupID.Int64 == 0 {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("asset %d is not a group asset", assetID),
			i18n.ArtifactRenameEntityTypeInvalid,
		)
	}

	// 2. 同级冲突：同 group 下同名文件
	if _, err = txStore.Assets().GetGroupAsset(ctx, newName, srcAsset.GroupID.Int64); err == nil {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("asset %q already exists in group", newName),
			i18n.ArtifactRenameSiblingConflict,
		)
	} else if !pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactAssetQueryError)
	}

	// 3. 更新 path
	srcAsset.Path = newName
	return txStore.Assets().Update(ctx, srcAsset)
}

// renameGroupInner 在调用方事务内完成一次 group 重命名。
func (s *Service) renameGroupInner(ctx context.Context, tx *gorm.DB, groupID int64, newName string) error {
	// 1. 读源 group
	txStore := artifactStore.NewStore(tx)
	srcGroup, err := txStore.Groups().GetById(ctx, groupID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRenameSourceDeleted)
	}
	if srcGroup.Deleted == 1 {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("group %d is soft-deleted", groupID),
			i18n.ArtifactRenameSourceDeleted,
		)
	}
	// 版本化格式不允许重命名
	if model.IsVersionedFormat(srcGroup.Format) {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("versioned format %s group cannot be renamed", srcGroup.Format),
			i18n.ArtifactRenameVersionedNotSupported,
		)
	}

	// 2. 同级冲突
	if srcGroup.ParentID.Valid && srcGroup.ParentID.Int64 > 0 {
		if _, err = txStore.Groups().GetSubNode(ctx, srcGroup.ParentID.Int64, newName); err == nil {
			return pkgErrors.WithI18nKey(
				fmt.Errorf("group %q already exists under parent", newName),
				i18n.ArtifactRenameSiblingConflict,
			)
		}
	} else {
		if _, err = txStore.Groups().GetRoot(ctx, srcGroup.ArtifactID, newName, srcGroup.Format); err == nil {
			return pkgErrors.WithI18nKey(
				fmt.Errorf("root group %q already exists", newName),
				i18n.ArtifactRenameSiblingConflict,
			)
		}
	}

	// 3. 主表更新
	if err := txStore.Groups().UpdateName(ctx, groupID, newName); err != nil {
		return err
	}

	// 4. tree 同步：删源侧 tree node，重建
	if node, e := txStore.Nodes().GetByTarget(ctx, srcGroup.ArtifactID, model.ArtifactTreeNodeTypeGroup, groupID); e == nil && node != nil {
		if err = txStore.Nodes().RecurseDeleteById(ctx, node.ID); err != nil {
			return err
		}
	}
	return s.refreshTree(ctx, string(srcGroup.Format), txStore)
}

// renameArtifactInner 重命名制品库（view）。
func (s *Service) renameArtifactInner(ctx context.Context, tx *gorm.DB, artifactID int64, newName string) error {
	txStore := artifactStore.NewStore(tx)

	// 1. 读 view
	view, err := txStore.Views().GetByID(ctx, artifactID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRenameSourceDeleted)
	}
	if view.Deleted == 1 {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("artifact %d is soft-deleted", artifactID),
			i18n.ArtifactRenameSourceDeleted,
		)
	}

	// 1.1 名称合法性校验（复用 CreateView 的规则）
	trimmed, err := model.ValidateLibName(newName, view.Format)
	if err != nil {
		return err
	}
	newName = trimmed

	// 2. 同名检查：同作用域下不能有同名活 view
	existing, err := txStore.Views().FindByNameInScope(ctx, view.SpaceID, view.RepoID, newName)
	if err != nil {
		return err
	}
	for _, v := range existing {
		if v.Deleted == 0 && v.ID != artifactID {
			return pkgErrors.WithI18nKey(
				fmt.Errorf("artifact view %q already exists", newName),
				i18n.ArtifactLibNameDuplicate,
			)
		}
	}

	// 3. 更新 view name
	if err = txStore.Views().UpdateName(ctx, artifactID, newName); err != nil {
		return err
	}

	return nil
}
