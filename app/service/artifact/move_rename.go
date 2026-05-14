package artifact

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

// movePackageInner 在单事务（由调用方包裹）内完成一次 package 移动。
// 语义：更新 package.ArtifactID / GroupID / Namespace，并重建 tree node。
//
//   - container 格式：忽略传入的 tGroupID，目标 group 直接取目标 view 下唯一的 container group。
//   - 非 container 格式：tGroupID == 0 视为挂到 format 根；tGroupID > 0 时目标 group 必须属于目标 view，
//     且格式要与源 package 一致。
func (s *Service) movePackageInner(ctx context.Context, pkgID, tArtifactID, tGroupID int64) error {
	// 1. 读源 package
	srcPkg, err := s.artStore.Packages().GetByID(ctx, pkgID)
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
	srcView, err := s.artStore.Views().GetByID(ctx, srcPkg.ArtifactID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveTargetInvalid)
	}
	tView, err := s.artStore.Views().GetByID(ctx, tArtifactID)
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
		// container：目标库下只有一个 group，直接取
		groups, err := s.artStore.Groups().ListDescendantsForOwner(ctx, tArtifactID, model.ArtifactContainerFormat)
		if err != nil {
			return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveTargetInvalid)
		}
		if len(groups) == 0 {
			return pkgErrors.WithI18nKey(
				fmt.Errorf("no container group under artifact %d", tArtifactID),
				i18n.ArtifactMoveTargetInvalid,
			)
		}
		targetGroupID = groups[0].ID
		namespace = ""
	default:
		if tGroupID > 0 {
			endpoint, err := s.artStore.Groups().GetEndpointById(ctx, tGroupID)
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
		} else {
			// tGroupID == 0：挂到 format 根
			targetGroupID = 0
			namespace = ""
		}
	}

	// 5. 同级冲突检查
	if _, err = s.artStore.Packages().GetByName(ctx, srcPkg.Name, namespace, tArtifactID, srcPkg.Format); err == nil {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("package %q already exists at target", srcPkg.Name),
			i18n.ArtifactMoveSiblingConflict,
		)
	} else if !pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactPackageQueryError)
	}

	// 6. 主表更新
	oldArtifactID := srcPkg.ArtifactID
	srcPkg.ArtifactID = tArtifactID
	srcPkg.GroupID = targetGroupID
	srcPkg.Namespace = namespace
	if err = s.artStore.Packages().Update(ctx, srcPkg); err != nil {
		return err
	}

	// 6.1 同步更新该 package 下所有 version 的 asset 的 artifactID
	versions, err := s.artStore.Versions().Find(ctx, model.SearchVersionOption{PackageID: srcPkg.ID})
	if err != nil {
		return err
	}
	for _, ver := range versions {
		assets, err := s.artStore.Assets().Search(ctx, model.SearchAssetOption{VersionID: ver.ID})
		if err != nil {
			return err
		}
		for _, asset := range assets {
			asset.ArtifactID = null.IntFrom(tArtifactID)
			asset.GroupID = null.IntFrom(targetGroupID)
			if err = s.artStore.Assets().Update(ctx, asset); err != nil {
				return err
			}
		}
	}

	// 7. tree node 同步：清掉源侧 package 子树，再调用 RefreshTree 重建。
	if node, e := s.artStore.Nodes().GetByTarget(ctx, oldArtifactID, model.ArtifactTreeNodeTypePackage, srcPkg.ID); e == nil && node != nil {
		if err = s.artStore.Nodes().RecurseDeleteById(ctx, node.ID); err != nil {
			return err
		}
	}
	return s.RefreshTree(ctx, string(srcPkg.Format))
}

// MovePackage 移动 package 到目标 artifact（及可选目标 group），整个过程在单事务内完成。
// container 格式会忽略传入的 tGroupID，实际落到目标 artifact 唯一的 container group。
func (s *Service) MovePackage(ctx context.Context, pkgID, tArtifactID, tGroupID int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.movePackageInner(ctx, pkgID, tArtifactID, tGroupID)
	})
}

// moveAssetInner 在调用方事务内完成一次组资产移动。
// 仅支持组资产（asset.GroupID 有效、versionID 无效）；版本资产随 package / version 移动，不在本方法范畴。
func (s *Service) moveAssetInner(ctx context.Context, assetID, tArtifactID, tGroupID int64) error {
	// 1. 读源 asset
	srcAsset, err := s.artStore.Assets().GetById(ctx, assetID)
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
	if srcAsset.Format == artifact.ArtifactContainerFormat {
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
	srcView, err := s.artStore.Views().GetByID(ctx, srcAsset.ArtifactID.Int64)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveTargetInvalid)
	}
	tView, err := s.artStore.Views().GetByID(ctx, tArtifactID)
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
	tGroup, err := s.artStore.Groups().GetById(ctx, tGroupID)
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
	if _, err = s.artStore.Assets().GetGroupAsset(ctx, srcAsset.Path, tGroupID); err == nil {
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
	return s.artStore.Assets().Update(ctx, srcAsset)
}

// MoveAsset 移动组资产到目标 group，整个过程在单事务内完成。
func (s *Service) MoveAsset(ctx context.Context, assetID, tArtifactID, tGroupID int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.moveAssetInner(ctx, assetID, tArtifactID, tGroupID)
	})
}

// moveGroupInner 在调用方事务内完成一次 group 移动。
// 语义：更新 group.ArtifactID / ParentID，跨 artifact 时子树内所有 group / asset / package 的 artifactID 一起改。
func (s *Service) moveGroupInner(ctx context.Context, groupID, tArtifactID, tParentGroupID int64) error {
	// 1. 读源 group
	srcGroup, err := s.artStore.Groups().GetById(ctx, groupID)
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

	// 2. 读源/目标 view，校验 format 严格一致
	srcView, err := s.artStore.Views().GetByID(ctx, srcGroup.ArtifactID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveTargetInvalid)
	}
	tView, err := s.artStore.Views().GetByID(ctx, tArtifactID)
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
		descendants, _ := s.artStore.Groups().ListDescendants(ctx, groupID)
		for _, d := range descendants {
			if d.ID == tParentGroupID {
				return pkgErrors.WithI18nKey(
					fmt.Errorf("target parent %d is a descendant of source group %d", tParentGroupID, groupID),
					i18n.ArtifactMoveCycleDetected,
				)
			}
		}
		// 验证目标父组存在且属于目标 view
		tParent, err := s.artStore.Groups().GetById(ctx, tParentGroupID)
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
		if _, err = s.artStore.Groups().GetSubNode(ctx, tParentGroupID, srcGroup.Name); err == nil {
			return pkgErrors.WithI18nKey(
				fmt.Errorf("group %q already exists under target parent", srcGroup.Name),
				i18n.ArtifactMoveSiblingConflict,
			)
		}
	} else {
		if _, err = s.artStore.Groups().GetRoot(ctx, tArtifactID, srcGroup.Name, srcGroup.Format); err == nil {
			return pkgErrors.WithI18nKey(
				fmt.Errorf("root group %q already exists in target artifact", srcGroup.Name),
				i18n.ArtifactMoveSiblingConflict,
			)
		}
	}

	// 5. 主表更新：源 group
	oldArtifactID := srcGroup.ArtifactID
	s.db.WithContext(ctx).Model(&artifactGroupRow{}).Where("id = ?", groupID).
		Updates(map[string]interface{}{
			"artifactID": tArtifactID,
			"parentID":   tParentGroupID,
		})

	// 6. 跨 artifact 时，子树内所有 group / asset / package 的 artifactID 一起改
	if oldArtifactID != tArtifactID {
		descendants, _ := s.artStore.Groups().ListDescendants(ctx, groupID)
		subtreeIDs := []int64{groupID}
		for _, d := range descendants {
			subtreeIDs = append(subtreeIDs, d.ID)
			s.db.WithContext(ctx).Model(&artifactGroupRow{}).Where("id = ?", d.ID).
				UpdateColumn("artifactID", tArtifactID)
		}
		// 子树下的 assets
		s.db.WithContext(ctx).Table("ops_artifact_assets").
			Where("groupID IN ?", subtreeIDs).
			UpdateColumn("artifactID", tArtifactID)
		// 子树下的 packages
		s.db.WithContext(ctx).Table("ops_artifact_packages").
			Where("groupID IN ?", subtreeIDs).
			UpdateColumn("artifactID", tArtifactID)
	}

	// 7. tree 同步：删源侧 tree node，重建
	if node, e := s.artStore.Nodes().GetByTarget(ctx, oldArtifactID, model.ArtifactTreeNodeTypeGroup, groupID); e == nil && node != nil {
		if err = s.artStore.Nodes().RecurseDeleteById(ctx, node.ID); err != nil {
			return err
		}
	}
	return s.RefreshTree(ctx, string(srcGroup.Format))
}

// MoveGroup 移动 group 到目标 artifact（及可选目标父 group），整个过程在单事务内完成。
func (s *Service) MoveGroup(ctx context.Context, groupID, tArtifactID, tParentGroupID int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.moveGroupInner(ctx, groupID, tArtifactID, tParentGroupID)
	})
}

// renameAssetInner 在调用方事务内完成一次组资产重命名。
func (s *Service) renameAssetInner(ctx context.Context, assetID int64, newName string) error {
	// 1. 读源 asset
	srcAsset, err := s.artStore.Assets().GetById(ctx, assetID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactRenameSourceDeleted)
	}
	if srcAsset.Deleted == 1 {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("asset %d is soft-deleted", assetID),
			i18n.ArtifactRenameSourceDeleted,
		)
	}
	if srcAsset.Format == artifact.ArtifactContainerFormat {
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
	if _, err = s.artStore.Assets().GetGroupAsset(ctx, newName, srcAsset.GroupID.Int64); err == nil {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("asset %q already exists in group", newName),
			i18n.ArtifactRenameSiblingConflict,
		)
	} else if !pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactAssetQueryError)
	}

	// 3. 更新 path
	srcAsset.Path = newName
	return s.artStore.Assets().Update(ctx, srcAsset)
}

// RenameAsset 重命名组资产，整个过程在单事务内完成。
func (s *Service) RenameAsset(ctx context.Context, assetID int64, newName string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.renameAssetInner(ctx, assetID, newName)
	})
}

// renameGroupInner 在调用方事务内完成一次 group 重命名。
func (s *Service) renameGroupInner(ctx context.Context, groupID int64, newName string) error {
	// 1. 读源 group
	srcGroup, err := s.artStore.Groups().GetById(ctx, groupID)
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
		if _, err = s.artStore.Groups().GetSubNode(ctx, srcGroup.ParentID.Int64, newName); err == nil {
			return pkgErrors.WithI18nKey(
				fmt.Errorf("group %q already exists under parent", newName),
				i18n.ArtifactRenameSiblingConflict,
			)
		}
	} else {
		if _, err = s.artStore.Groups().GetRoot(ctx, srcGroup.ArtifactID, newName, srcGroup.Format); err == nil {
			return pkgErrors.WithI18nKey(
				fmt.Errorf("root group %q already exists", newName),
				i18n.ArtifactRenameSiblingConflict,
			)
		}
	}

	// 3. 主表更新
	if err := s.artStore.Groups().UpdateName(ctx, groupID, newName); err != nil {
		return err
	}

	// 4. tree 同步：删源侧 tree node，重建
	if node, e := s.artStore.Nodes().GetByTarget(ctx, srcGroup.ArtifactID, model.ArtifactTreeNodeTypeGroup, groupID); e == nil && node != nil {
		if err = s.artStore.Nodes().RecurseDeleteById(ctx, node.ID); err != nil {
			return err
		}
	}
	return s.RefreshTree(ctx, string(srcGroup.Format))
}

// RenameGroup 重命名 group，整个过程在单事务内完成。
func (s *Service) RenameGroup(ctx context.Context, groupID int64, newName string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.renameGroupInner(ctx, groupID, newName)
	})
}

// artifactGroupRow 用于 GORM 直接操作 ops_artifact_groups 表。
type artifactGroupRow struct {
	ID         int64    `gorm:"primaryKey;column:id"`
	ParentID   null.Int `gorm:"column:parentID"`
	ArtifactID int64    `gorm:"column:artifactID"`
	Name       string   `gorm:"column:name"`
	Format     string   `gorm:"column:format"`
	Deleted    int8     `gorm:"column:deleted"`
}

func (artifactGroupRow) TableName() string { return "ops_artifact_groups" }

// parseEntityRef 解析 "kind.id" 形式的 entityID。
// 失败时返回带 i18n key 的错误。
func parseEntityRef(entityID string) (model.ArtifactEntityKind, int64, error) {
	parts := strings.SplitN(entityID, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", 0, pkgErrors.WithI18nKey(
			fmt.Errorf("invalid entity id: %q", entityID),
			i18n.ArtifactEntityIdInvalid,
		)
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		return "", 0, pkgErrors.WithI18nKey(
			fmt.Errorf("invalid entity id: %q", entityID),
			i18n.ArtifactEntityIdInvalid,
		)
	}
	return model.ArtifactEntityKind(parts[0]), id, nil
}

// moveByKind 按实体类型派发到对应的 Move 方法（不带事务，由调用方包裹）。
func (s *Service) moveByKind(ctx context.Context, kind model.ArtifactEntityKind, sourceID, tArtifactID, tGroupID int64) error {
	switch kind {
	case model.ArtifactEntityKindPackage:
		return s.movePackageInner(ctx, sourceID, tArtifactID, tGroupID)
	case model.ArtifactEntityKindAsset:
		return s.moveAssetInner(ctx, sourceID, tArtifactID, tGroupID)
	case model.ArtifactEntityKindGroup:
		return s.moveGroupInner(ctx, sourceID, tArtifactID, tGroupID)
	default:
		return pkgErrors.WithI18nKey(
			fmt.Errorf("unsupported entity type for move: %s", kind),
			i18n.ArtifactMoveEntityTypeInvalid,
		)
	}
}

// renameByKind 按实体类型派发到对应的 Rename 方法（不带事务，由调用方包裹）。
// package / version 不允许重命名；container 格式的制品库也暂不允许重命名。
func (s *Service) renameByKind(ctx context.Context, kind model.ArtifactEntityKind, sourceID int64, newName string) error {
	switch kind {
	case model.ArtifactEntityKindAsset:
		return s.renameAssetInner(ctx, sourceID, newName)
	case model.ArtifactEntityKindGroup:
		return s.renameGroupInner(ctx, sourceID, newName)
	case model.ArtifactEntityKindArtifact:
		return s.renameArtifactInner(ctx, sourceID, newName)
	case model.ArtifactEntityKindPackage:
		return pkgErrors.WithI18nKey(
			fmt.Errorf("rename is not supported for packages"),
			i18n.ArtifactRenameVersionedNotSupported,
		)
	case model.ArtifactEntityKindVersion:
		return pkgErrors.WithI18nKey(
			fmt.Errorf("rename is not supported for versions"),
			i18n.ArtifactRenameVersionedNotSupported,
		)
	default:
		return pkgErrors.WithI18nKey(
			fmt.Errorf("unsupported entity type for rename: %s", kind),
			i18n.ArtifactRenameEntityTypeInvalid,
		)
	}
}

// renameArtifactInner 重命名制品库（view），同步更新 container 格式下的 group name 和 package namespace。
func (s *Service) renameArtifactInner(ctx context.Context, artifactID int64, newName string) error {
	// 1. 读 view
	view, err := s.artStore.Views().GetByID(ctx, artifactID)
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
	trimmed, err := artifact.ValidateViewName(newName, view.Format)
	if err != nil {
		return err
	}
	newName = trimmed

	// 2. 同名检查：同作用域下不能有同名活 view
	existing, err := s.artStore.Views().FindByNameInScope(ctx, view.SpaceID, view.RepoID, newName)
	if err != nil {
		return err
	}
	for _, v := range existing {
		if v.Deleted == 0 && v.ID != artifactID {
			return pkgErrors.WithI18nKey(
				fmt.Errorf("artifact view %q already exists", newName),
				i18n.ArtifactRenameSiblingConflict,
			)
		}
	}

	oldName := view.Name
	var oldContainerGroupID int64
	if view.Format == model.ArtifactContainerFormat || view.Format == model.ArtifactAllFormat {
		if group, e := s.artStore.Groups().GetRoot(ctx, artifactID, oldName, model.ArtifactContainerFormat); e == nil {
			oldContainerGroupID = group.ID
		}
	}

	// 3. 更新 view name
	if err = s.artStore.Views().UpdateName(ctx, artifactID, newName); err != nil {
		return err
	}

	// 4. container 格式：同步更新 group name 和 package namespace
	if view.Format == model.ArtifactContainerFormat || view.Format == model.ArtifactAllFormat {
		if err = s.artStore.Groups().UpdateNameByCondition(ctx, artifactID, oldName, newName, string(model.ArtifactContainerFormat)); err != nil {
			return err
		}
		if err = s.artStore.Packages().UpdateNamespaceByCondition(ctx, artifactID, oldName, newName, string(model.ArtifactContainerFormat)); err != nil {
			return err
		}
		if oldContainerGroupID > 0 {
			if node, e := s.artStore.Nodes().GetByTarget(ctx, artifactID, model.ArtifactTreeNodeTypeGroup, oldContainerGroupID); e == nil && node != nil {
				if err = s.artStore.Nodes().RecurseDeleteById(ctx, node.ID); err != nil {
					return err
				}
			}
			if err = s.RefreshTree(ctx, string(model.ArtifactContainerFormat)); err != nil {
				return err
			}
		}
	}

	return nil
}

// MoveEntity 解析 entityID 并派发到对应的 Move 方法，整个过程在单事务内完成。
func (s *Service) MoveEntity(ctx context.Context, req *model.MoveEntityReq) error {
	kind, id, err := parseEntityRef(req.EntityID)
	if err != nil {
		return err
	}
	// 移动操作必须指定目标 group
	if req.TargetGroupID == 0 {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("targetGroupID is required for move"),
			i18n.ArtifactMoveTargetGroupRequired,
		)
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.moveByKind(ctx, kind, id, req.TargetArtifactID, req.TargetGroupID)
	})
}

// RenameEntity 解析 entityID 并派发到对应的 Rename 方法，整个过程在单事务内完成。
func (s *Service) RenameEntity(ctx context.Context, req *model.RenameEntityReq) error {
	kind, id, err := parseEntityRef(req.EntityID)
	if err != nil {
		return err
	}
	// 重命名操作必须指定 newName
	if req.NewName == "" {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("newName is required for rename"),
			i18n.ArtifactRenameNewNameRequired,
		)
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.renameByKind(ctx, kind, id, req.NewName)
	})
}

// RelocateEntity 同时支持移动 + 重命名，单事务内先移动后改名。
//
// 语义：
//   - TargetArtifactID == 0 或与当前一致，且 TargetGroupID == 0 或与当前一致 → 不移动；
//   - NewName == "" 或与当前一致 → 不改名；
//   - 两者都不变 → 返回 ArtifactRelocateNoChange。
func (s *Service) RelocateEntity(ctx context.Context, req *model.RelocateEntityReq) error {
	kind, id, err := parseEntityRef(req.EntityID)
	if err != nil {
		return err
	}

	// 取当前实体的 artifactID / groupID / name，用于与传入参数比对，判断是否真有变化。
	var (
		curArtifactID int64
		curGroupID    int64
		curName       string
	)
	switch kind {
	case model.ArtifactEntityKindGroup:
		g, e := s.artStore.Groups().GetById(ctx, id)
		if e != nil {
			return pkgErrors.WithI18nKey(e, i18n.ArtifactMoveSourceDeleted)
		}
		curArtifactID = g.ArtifactID
		if g.ParentID.Valid {
			curGroupID = g.ParentID.Int64
		}
		curName = g.Name
	case model.ArtifactEntityKindAsset:
		a, e := s.artStore.Assets().GetById(ctx, id)
		if e != nil {
			return pkgErrors.WithI18nKey(e, i18n.ArtifactMoveSourceDeleted)
		}
		if a.ArtifactID.Valid {
			curArtifactID = a.ArtifactID.Int64
		}
		if a.GroupID.Valid {
			curGroupID = a.GroupID.Int64
		}
		curName = a.Path
	case model.ArtifactEntityKindPackage:
		p, e := s.artStore.Packages().GetByID(ctx, id)
		if e != nil {
			return pkgErrors.WithI18nKey(e, i18n.ArtifactMoveSourceDeleted)
		}
		curArtifactID = p.ArtifactID
		curGroupID = p.GroupID
		curName = p.Name
	case model.ArtifactEntityKindArtifact:
		v, e := s.artStore.Views().GetByID(ctx, id)
		if e != nil {
			return pkgErrors.WithI18nKey(e, i18n.ArtifactMoveSourceDeleted)
		}
		curArtifactID = v.ID
		curName = v.Name
	default:
		return pkgErrors.WithI18nKey(
			fmt.Errorf("unsupported entity type: %s", kind),
			i18n.ArtifactMoveEntityTypeInvalid,
		)
	}

	// 判断是否需要移动（traget 不传或与当前一致 → 不移动）
	tArtifactID := req.TargetArtifactID
	if tArtifactID == 0 {
		tArtifactID = curArtifactID
	}
	tGroupID := req.TargetGroupID
	needMove := tArtifactID != curArtifactID || (req.TargetGroupID != 0 && req.TargetGroupID != curGroupID)

	// 判断是否需要重命名（NewName 为空或与当前一致 → 不改名）
	needRename := req.NewName != "" && req.NewName != curName

	if !needMove && !needRename {
		// 既不移动也不改名，直接成功返回，不视为错误
		return nil
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if needMove {
			// 移动时若未显式指定 tGroupID，沿用当前父级（保持层级不变）
			if tGroupID == 0 {
				tGroupID = curGroupID
			}
			if err := s.moveByKind(ctx, kind, id, tArtifactID, tGroupID); err != nil {
				return err
			}
		}
		if needRename {
			if err := s.renameByKind(ctx, kind, id, req.NewName); err != nil {
				return err
			}
		}
		return nil
	})
}
