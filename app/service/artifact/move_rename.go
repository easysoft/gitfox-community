package artifact

import (
	"context"
	"fmt"

	artModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/model"
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

	// 2. 读目标 view
	tView, err := s.artStore.Views().GetByID(ctx, tArtifactID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveTargetInvalid)
	}

	// 3. 格式匹配：目标库格式必须与源包一致，或目标库为 all
	if tView.Format != srcPkg.Format && tView.Format != model.ArtifactAllFormat {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("package format %s does not match target view format %s", srcPkg.Format, tView.Format),
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

	// 7. tree node 同步：清掉源侧 package 节点（版本节点会随之孤立，交给 Syncer 重建覆盖）
	syncer := artModel.NewSyncer(s.artStore)
	if node, e := s.artStore.Nodes().GetByTarget(ctx, oldArtifactID, model.ArtifactTreeNodeTypePackage, srcPkg.ID); e == nil && node != nil {
		_ = s.artStore.Nodes().DeleteById(ctx, node.ID)
	}
	if _, err = syncer.SyncTreeNodeForPackage(ctx, srcPkg); err != nil {
		return err
	}
	versions, err := s.artStore.Versions().Find(ctx, model.SearchVersionOption{PackageID: srcPkg.ID})
	if err != nil {
		return err
	}
	for _, ver := range versions {
		if _, err = syncer.SyncTreeNodeForVersion(ctx, ver, srcPkg); err != nil {
			return err
		}
	}
	return nil
}

// MovePackage 移动 package 到目标 artifact（及可选目标 group），整个过程在单事务内完成。
// container 格式会忽略传入的 tGroupID，实际落到目标 artifact 唯一的 container group。
func (s *Service) MovePackage(ctx context.Context, pkgID, tArtifactID, tGroupID int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.movePackageInner(ctx, pkgID, tArtifactID, tGroupID)
	})
}

// moveAssetInner 在调用方事务内完成一次组资产移动。
// 仅支持组资产（asset.GroupID 有效、VersionID 无效）；版本资产随 package / version 移动，不在本方法范畴。
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

	// 2. 目标 view
	tView, err := s.artStore.Views().GetByID(ctx, tArtifactID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveTargetInvalid)
	}
	// 3. 格式匹配
	if tView.Format != srcAsset.Format {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("asset format %s does not match target view format %s", srcAsset.Format, tView.Format),
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

	// 2. 目标 view
	tView, err := s.artStore.Views().GetByID(ctx, tArtifactID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactMoveTargetInvalid)
	}
	if tView.Format != srcGroup.Format && tView.Format != model.ArtifactAllFormat {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("group format %s does not match target view format %s", srcGroup.Format, tView.Format),
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
	syncer := artModel.NewSyncer(s.artStore)
	if node, e := s.artStore.Nodes().GetByTarget(ctx, oldArtifactID, model.ArtifactTreeNodeTypeGroup, groupID); e == nil && node != nil {
		_ = s.artStore.Nodes().RecurseDeleteById(ctx, node.ID)
	}
	// 重新读取更新后的 group
	updatedGroup, _ := s.artStore.Groups().GetById(ctx, groupID)
	if updatedGroup != nil {
		_, _ = syncer.SyncTreeNodeForGroup(ctx, updatedGroup)
	}
	// 子树 group 也重建
	descendants, _ := s.artStore.Groups().ListDescendants(ctx, groupID)
	for _, d := range descendants {
		_, _ = syncer.SyncTreeNodeForGroup(ctx, d)
	}
	return nil
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
	s.db.WithContext(ctx).Model(&artifactGroupRow{}).Where("id = ?", groupID).
		UpdateColumn("name", newName)

	// 4. tree 同步：删源侧 tree node，重建
	syncer := artModel.NewSyncer(s.artStore)
	if node, e := s.artStore.Nodes().GetByTarget(ctx, srcGroup.ArtifactID, model.ArtifactTreeNodeTypeGroup, groupID); e == nil && node != nil {
		_ = s.artStore.Nodes().RecurseDeleteById(ctx, node.ID)
	}
	updatedGroup, _ := s.artStore.Groups().GetById(ctx, groupID)
	if updatedGroup != nil {
		_, _ = syncer.SyncTreeNodeForGroup(ctx, updatedGroup)
	}
	// 子树也重建
	descendants, _ := s.artStore.Groups().ListDescendants(ctx, groupID)
	for _, d := range descendants {
		_, _ = syncer.SyncTreeNodeForGroup(ctx, d)
	}
	return nil
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

// MoveEntity 按 entityType 派发到对应的 Move 方法。
func (s *Service) MoveEntity(ctx context.Context, req *model.MoveEntityReq) error {
	switch req.EntityType {
	case "package":
		return s.MovePackage(ctx, req.SourceID, req.TargetArtifactID, req.TargetGroupID)
	case "asset":
		return s.MoveAsset(ctx, req.SourceID, req.TargetArtifactID, req.TargetGroupID)
	case "group":
		return s.MoveGroup(ctx, req.SourceID, req.TargetArtifactID, req.TargetGroupID)
	default:
		return pkgErrors.WithI18nKey(
			fmt.Errorf("unsupported entity type for move: %s", req.EntityType),
			i18n.ArtifactMoveEntityTypeInvalid,
		)
	}
}

// RenameEntity 按 entityType 派发到对应的 Rename 方法。
func (s *Service) RenameEntity(ctx context.Context, req *model.RenameEntityReq) error {
	switch req.EntityType {
	case "asset":
		return s.RenameAsset(ctx, req.SourceID, req.NewName)
	case "group":
		return s.RenameGroup(ctx, req.SourceID, req.NewName)
	case "package":
		// package 重命名不支持（版本化包名参与分发协议）
		return pkgErrors.WithI18nKey(
			fmt.Errorf("rename is not supported for packages"),
			i18n.ArtifactRenameVersionedNotSupported,
		)
	default:
		return pkgErrors.WithI18nKey(
			fmt.Errorf("unsupported entity type for rename: %s", req.EntityType),
			i18n.ArtifactRenameEntityTypeInvalid,
		)
	}
}
