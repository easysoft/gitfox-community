package artifact

import (
	"context"
	"fmt"

	request "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter/helm"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/entity"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gorm.io/gorm"
)

// SoftRemove 软删除资产
func (s *Service) SoftRemove(ctx context.Context, req *model.SoftRemoveEntitiesReq) (*model.ArtifactNodeRemoveReport, error) {
	ids := req.EntityIDs
	report := &model.ArtifactNodeRemoveReport{Total: len(ids), Data: make([]*model.ArtifactNodeRemoveRes, len(ids))}
	for _, eid := range ids {
		if _, err := entity.Parse(ctx, s.artStore, eid); err != nil {
			logger.GetLogger(ctx).Warnf("parse artifact entity %s failed before soft remove: %v", eid, err)
			if pkgErrors.GetI18nKey(err) != i18n.ErrorDefault {
				return nil, err
			}
			return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactSoftRemoveParseError)
		}
	}

	// 取当前会话用户作为「删除人」（editedBy），匿名/系统调用时为空。
	var operator string
	if session, ok := request.GetAuthSession(ctx); ok && session != nil && session.User != nil {
		operator = session.User.Account
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for idx, eid := range ids {
			res := &model.ArtifactNodeRemoveRes{NodeID: eid, Status: model.ArtifactStatusOK}
			report.Data[idx] = res
			if err := s.softRemoveEntity(ctx, eid, operator, res); err != nil {
				if res.Status == model.ArtifactStatusOK {
					res.Status = model.ArtifactStatusUnknown
				}
				return err
			}
		}
		return nil
	}); err != nil {
		// 内层已显式带上具体 i18n key（如制品库非空 / 不存在 / 查询失败）时保留之，
		// 仅对无 key 的未知错误统一包装为 ArtifactSoftRemoveError。
		if pkgErrors.GetI18nKey(err) != i18n.ErrorDefault {
			return nil, err
		}
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactSoftRemoveError)
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

func (s *Service) softRemoveEntity(ctx context.Context, eid model.EntityID, operator string, res *model.ArtifactNodeRemoveRes) error {
	splitEntityID := eid.GetEntityID()

	switch splitEntityID.Kind {
	case constant.ArtifactEntityKindAsset:
		asset, err := s.artStore.Assets().GetById(ctx, splitEntityID.ID)
		if err != nil {
			res.Status = model.ArtifactStatusNotFound
			return nil
		}
		return s.softRemoveAsset(ctx, asset, operator, res)
	case constant.ArtifactEntityKindVersion:
		ver, err := s.artStore.Versions().GetByID(ctx, splitEntityID.ID)
		if err != nil {
			res.Status = model.ArtifactStatusNotFound
			return nil
		}
		pkg, err := s.artStore.Packages().GetByID(ctx, ver.PackageID)
		if err != nil {
			res.Status = model.ArtifactStatusNotFound
			return nil
		}
		return s.softRemoveVersion(ctx, artifactIDForVersion(pkg, ver), ver, operator, res)
	case constant.ArtifactEntityKindPackage:
		pkg, err := s.artStore.Packages().GetByID(ctx, splitEntityID.ID)
		if err != nil {
			res.Status = model.ArtifactStatusNotFound
			return nil
		}
		return s.softRemovePackage(ctx, pkg, operator, res)
	case constant.ArtifactEntityKindGroup:
		group, err := s.artStore.Groups().GetById(ctx, splitEntityID.ID)
		if err != nil {
			res.Status = model.ArtifactStatusNotFound
			return nil
		}
		return s.softRemoveGroup(ctx, group, operator, res)
	case constant.ArtifactEntityKindArtifact:
		view, err := s.artStore.Views().GetByID(ctx, splitEntityID.ID)
		if err != nil {
			res.Status = model.ArtifactStatusNotFound
			return pkgErrors.WithI18nKey(err, i18n.ArtifactSoftRemoveArtifactNotFound)
		}
		return s.softRemoveArtifact(ctx, view, operator, res)
	default:
		res.Status = model.ArtifactStatusInvalidId
		return nil
	}
}

// softRemoveAsset 软删制品资产。
//
// blob 引用保护逻辑：
//   - 先软删 asset 本身（deleted=1）
//   - 然后检查该 asset 关联的 blob 是否还被其他 active（deleted=0）asset 引用
//   - 判断方式：按 blobID 查询 ops_artifact_assets 表中 deleted=0 的记录数
//     （使用 BlobAssetsOption{BlobID} + AssetExcludeDeletedOption 组合查询）
//   - 如果还有其他 active asset 引用同一 blob → 不软删 blob，等最后一个引用者删除时再处理
//   - 如果没有任何 active asset 引用了 → 软删 blob（deleted=1），后续由 GC 物理清理
func (s *Service) softRemoveAsset(ctx context.Context, asset *model.ArtifactAsset, operator string, res *model.ArtifactNodeRemoveRes) error {
	if err := s.artStore.Assets().SoftDeleteById(ctx, asset.ID, operator); err != nil {
		return err
	}
	if asset.BlobID != 0 {
		// 查询是否还有其他 active asset 引用该 blob（当前 asset 已被软删，不会被查到）
		refs, _ := s.artStore.Assets().Search(ctx, artifactStore.BlobAssetsOption{BlobID: asset.BlobID}, artifactStore.AssetExcludeDeletedOption{})
		if len(refs) == 0 {
			// 没有其他引用了，软删 blob
			if err := s.artStore.Blobs().SoftDeleteById(ctx, asset.BlobID, operator); err != nil {
				return err
			}
		}
	}
	res.Assets++
	return nil
}

func (s *Service) softRemoveVersion(ctx context.Context, artifactID int64, ver *model.ArtifactVersion, operator string, res *model.ArtifactNodeRemoveRes) error {
	// 判空：有活 asset 时拒绝删除
	assets, err := s.artStore.Assets().Search(ctx, model.SearchAssetOption{VersionID: ver.ID}, artifactStore.AssetExcludeDeletedOption{})
	if err != nil {
		return err
	}
	if len(assets) > 0 {
		return pkgErrors.WithI18nKey(
			fmt.Errorf("version %d still has %d active asset(s)", ver.ID, len(assets)),
			i18n.ArtifactSoftRemoveVersionNotEmpty,
		)
	}
	if err = s.artStore.Versions().SoftDelete(ctx, ver, operator); err != nil {
		return err
	}
	res.Versions++
	_ = s.deleteTreeNodeByTarget(ctx, artifactID, model.ArtifactTreeNodeTypeVersion, ver.ID)
	return nil
}

func (s *Service) softRemovePackage(ctx context.Context, pkg *model.ArtifactPackage, operator string, res *model.ArtifactNodeRemoveRes) error {
	// 判空：检查所有 version 下是否有活 asset
	versions, err := s.artStore.Versions().Find(ctx, model.SearchVersionOption{PackageID: pkg.ID})
	if err != nil {
		return err
	}
	if pkg.Format == model.ArtifactContainerFormat {
		return s.softRemoveContainerPackage(ctx, pkg, versions, operator, res)
	}
	for _, ver := range versions {
		activeAssets, aErr := s.artStore.Assets().Search(ctx, model.SearchAssetOption{VersionID: ver.ID}, artifactStore.AssetExcludeDeletedOption{})
		if aErr != nil {
			return aErr
		}
		if len(activeAssets) > 0 {
			return pkgErrors.WithI18nKey(
				fmt.Errorf("package %d still has active asset(s) under version %d", pkg.ID, ver.ID),
				i18n.ArtifactSoftRemovePackageNotEmpty,
			)
		}
	}
	// 空包：级联软删 version
	for _, ver := range versions {
		if err = s.artStore.Versions().SoftDelete(ctx, ver, operator); err != nil {
			return err
		}
		res.Versions++
		_ = s.deleteTreeNodeByTarget(ctx, artifactIDForVersion(pkg, ver), model.ArtifactTreeNodeTypeVersion, ver.ID)
	}
	if err = s.artStore.Packages().SoftDelete(ctx, pkg, operator); err != nil {
		return err
	}
	res.Packages++
	_ = s.deleteTreeNodeByTarget(ctx, pkg.ArtifactID, model.ArtifactTreeNodeTypePackage, pkg.ID)
	if pkg.Format == model.ArtifactHelmFormat {
		_ = helm.NewHelmIndex(s.artStore, &adapter.ViewDescriptor{ArtifactID: pkg.ArtifactID, Store: s.fileStore, StorageID: 1}).UpdateRepo(ctx, "/artifacts/helm")
	}
	return nil
}

func (s *Service) softRemoveContainerPackage(
	ctx context.Context,
	pkg *model.ArtifactPackage,
	versions []*model.ArtifactVersion,
	operator string,
	res *model.ArtifactNodeRemoveRes,
) error {
	if len(versions) > 0 {
		res.Status = model.ArtifactStatusNotEmpty
		return pkgErrors.WithI18nKey(
			fmt.Errorf("container package %d still has %d active version(s)", pkg.ID, len(versions)),
			i18n.ArtifactSoftRemovePackageVersionExists,
		)
	}

	artifactIDs := map[int64]struct{}{}
	if pkg.ArtifactID != 0 {
		artifactIDs[pkg.ArtifactID] = struct{}{}
	}
	allVersions, err := s.artStore.Versions().Find(ctx, model.SearchVersionOption{PackageID: pkg.ID, IncludeDeleted: true})
	if err != nil {
		return err
	}
	for _, ver := range allVersions {
		artifactID := artifactIDForVersion(pkg, ver)
		if artifactID != 0 {
			artifactIDs[artifactID] = struct{}{}
		}
	}

	if err = s.artStore.Packages().SoftDelete(ctx, pkg, operator); err != nil {
		return err
	}
	res.Packages++
	for artifactID := range artifactIDs {
		_ = s.deleteTreeNodeByTarget(ctx, artifactID, model.ArtifactTreeNodeTypePackage, pkg.ID)
	}
	return nil
}

// softRemoveGroup 软删指定分组。
// 行为口径（不再做判空校验）：
//   - 当前 group 自身 + 所有未软删后代 group 一律级联软删；
//   - 链路上每一层的活 asset / version / package 一并软删（含 group 直挂 asset 与 package/version 链路下挂 asset）；
//   - asset 关联的 blob 引用计数归零时同步软删 blob，由 GC 物理清理；
//   - 每次软删 group / package / version / asset 都同步清对应 tree_node，避免前端刷树残留。
func (s *Service) softRemoveGroup(ctx context.Context, group *model.ArtifactGroup, operator string, res *model.ArtifactNodeRemoveRes) error {
	// 1) 收集当前 group + 所有未软删子 group，作为级联软删的目标集合。
	//    Groups().ListDescendants 返回不含自身，吾在头部追加当前 group 统一处理。
	descendants, err := s.artStore.Groups().ListDescendants(ctx, group.ID)
	if err != nil {
		res.Status = model.ArtifactStatusUnknown
		return pkgErrors.WithI18nKey(err, i18n.ArtifactSoftRemoveGroupQueryError)
	}
	targets := make([]*model.ArtifactGroup, 0, len(descendants)+1)
	targets = append(targets, group)
	targets = append(targets, descendants...)

	// 2) 由深到浅级联软删 targets：先清直挂 asset、再清 package/version 链路、最后软删 group 本身。
	//    Groups().ListDescendants 按 parentID 升序，反转后即从叶子起。
	for i := len(targets) - 1; i >= 0; i-- {
		g := targets[i]

		// 2.1 先软删 group 直挂 asset（generic 场景）。
		directAssets, aErr := s.artStore.Assets().Search(ctx,
			artifactStore.GroupAssetsOption{GroupID: g.ID}, artifactStore.AssetExcludeDeletedOption{})
		if aErr != nil {
			res.Status = model.ArtifactStatusUnknown
			return pkgErrors.WithI18nKey(aErr, i18n.ArtifactSoftRemoveGroupQueryError)
		}
		for _, asset := range directAssets {
			if err = s.softRemoveAsset(ctx, asset, operator, res); err != nil {
				return err
			}
		}

		// 2.2 再软删 package/version 链路下挂的 asset / version / package（raw / maven 等版本化场景）。
		pkgs, pErr := s.artStore.Packages().Search(ctx, artifactStore.GroupPackagesOption{GroupID: g.ID})
		if pErr != nil {
			return pErr
		}
		for _, pkg := range pkgs {
			if pkg.Format == model.ArtifactContainerFormat {
				continue
			}
			versions, vErr := s.artStore.Versions().Find(ctx, model.SearchVersionOption{PackageID: pkg.ID})
			if vErr != nil {
				return vErr
			}
			for _, ver := range versions {
				verAssets, vaErr := s.artStore.Assets().Search(ctx,
					model.SearchAssetOption{VersionID: ver.ID}, artifactStore.AssetExcludeDeletedOption{})
				if vaErr != nil {
					res.Status = model.ArtifactStatusUnknown
					return pkgErrors.WithI18nKey(vaErr, i18n.ArtifactSoftRemoveGroupQueryError)
				}
				for _, asset := range verAssets {
					if err = s.softRemoveAsset(ctx, asset, operator, res); err != nil {
						return err
					}
				}
				if err = s.artStore.Versions().SoftDelete(ctx, ver, operator); err != nil {
					return err
				}
				res.Versions++
				_ = s.deleteTreeNodeByTarget(ctx, g.ArtifactID, model.ArtifactTreeNodeTypeVersion, ver.ID)
			}
			if err = s.artStore.Packages().SoftDelete(ctx, pkg, operator); err != nil {
				return err
			}
			res.Packages++
			_ = s.deleteTreeNodeByTarget(ctx, g.ArtifactID, model.ArtifactTreeNodeTypePackage, pkg.ID)
		}

		// 2.3 最后软删 group 本身。
		if err = s.artStore.Groups().SoftDeleteById(ctx, g.ID, operator); err != nil {
			return err
		}
		_ = s.deleteTreeNodeByTarget(ctx, g.ArtifactID, model.ArtifactTreeNodeTypeGroup, g.ID)
	}
	// 3) 若该 format 下已无活 group，清理 format 节点（避免前端看到空的格式入口）
	remaining, _ := s.artStore.Groups().ListDescendantsForOwner(ctx, group.ArtifactID, group.Format)
	if len(remaining) == 0 {
		_ = s.deleteFormatTreeNode(ctx, group.ArtifactID, group.Format)
	}
	return nil
}

// softRemoveArtifact 软删整个制品库（view）。
// 判空口径：view 下不能存在任何未软删的 group / package / version / asset。
func (s *Service) softRemoveArtifact(ctx context.Context, view *model.ArtifactView, operator string, res *model.ArtifactNodeRemoveRes) error {
	activeAssets, err := s.artStore.Assets().Search(ctx, artifactStore.AssetViewOption{ArtifactID: view.ID}, artifactStore.AssetExcludeDeletedOption{})
	if err != nil {
		res.Status = model.ArtifactStatusUnknown
		return pkgErrors.WithI18nKey(err, i18n.ArtifactSoftRemoveArtifactQueryError)
	}
	if len(activeAssets) > 0 {
		res.Status = model.ArtifactStatusNotEmpty
		return pkgErrors.WithI18nKey(
			fmt.Errorf("artifact view %d still has %d active asset(s)", view.ID, len(activeAssets)),
			i18n.ArtifactSoftRemoveArtifactAssetExists,
		)
	}

	if view.Format == model.ArtifactContainerFormat {
		activeVersions, vErr := s.artStore.Versions().Find(ctx, model.SearchVersionOption{ArtifactLibID: view.ID})
		if vErr != nil {
			res.Status = model.ArtifactStatusUnknown
			return pkgErrors.WithI18nKey(vErr, i18n.ArtifactSoftRemoveArtifactQueryError)
		}
		if len(activeVersions) > 0 {
			res.Status = model.ArtifactStatusNotEmpty
			return pkgErrors.WithI18nKey(
				fmt.Errorf("container artifact view %d still has %d active version(s)", view.ID, len(activeVersions)),
				i18n.ArtifactSoftRemoveArtifactVersionExists,
			)
		}
	}

	if view.Format != model.ArtifactContainerFormat {
		packages, pErr := s.artStore.Packages().List(ctx, view.ID, false)
		if pErr != nil {
			res.Status = model.ArtifactStatusUnknown
			return pkgErrors.WithI18nKey(pErr, i18n.ArtifactSoftRemoveArtifactQueryError)
		}
		if len(packages) > 0 {
			res.Status = model.ArtifactStatusNotEmpty
			return pkgErrors.WithI18nKey(
				fmt.Errorf("artifact view %d still has %d active package(s)", view.ID, len(packages)),
				i18n.ArtifactSoftRemoveArtifactPackageExists,
			)
		}

		allPackages, pErr := s.artStore.Packages().List(ctx, view.ID, true)
		if pErr != nil {
			res.Status = model.ArtifactStatusUnknown
			return pkgErrors.WithI18nKey(pErr, i18n.ArtifactSoftRemoveArtifactQueryError)
		}
		for _, pkg := range allPackages {
			versions, vErr := s.artStore.Versions().Find(ctx, model.SearchVersionOption{PackageID: pkg.ID})
			if vErr != nil {
				res.Status = model.ArtifactStatusUnknown
				return pkgErrors.WithI18nKey(vErr, i18n.ArtifactSoftRemoveArtifactQueryError)
			}
			if len(versions) > 0 {
				res.Status = model.ArtifactStatusNotEmpty
				return pkgErrors.WithI18nKey(
					fmt.Errorf("artifact view %d still has active version(s) under package %d", view.ID, pkg.ID),
					i18n.ArtifactSoftRemoveArtifactVersionExists,
				)
			}
		}
	}

	groups, err := s.artStore.Groups().ListDescendantsForOwner(ctx, view.ID, view.Format)
	if err != nil {
		res.Status = model.ArtifactStatusUnknown
		return pkgErrors.WithI18nKey(err, i18n.ArtifactSoftRemoveArtifactQueryError)
	}

	if len(groups) != 0 {
		res.Status = model.ArtifactStatusNotEmpty
		return pkgErrors.WithI18nKey(
			fmt.Errorf("artifact view %d still has active group(s)", view.ID),
			i18n.ArtifactSoftRemoveArtifactGroupExists,
		)
	}

	if err = s.artStore.Views().SoftDelete(ctx, view, operator); err != nil {
		return err
	}
	return nil
}

func (s *Service) deleteTreeNodeByTarget(ctx context.Context, artifactID int64, nodeType model.ArtifactTreeNodeType, id int64) error {
	node, err := s.artStore.Nodes().GetByTarget(ctx, artifactID, nodeType, id)
	if err != nil {
		return nil
	}
	return s.artStore.Nodes().DeleteById(ctx, node.ID)
}

// deleteFormatTreeNode 删除指定 artifact 下某 format 的根树节点（path="/"）。
// 当该 format 下已无活 group 时由软删链路调用，避免前端看到空的格式入口。
// 节点不存在时静默返回（不报错），保持调用方简洁。
func (s *Service) deleteFormatTreeNode(ctx context.Context, artifactID int64, format model.ArtifactFormat) error {
	node, err := s.artStore.Nodes().GetByPath(ctx, artifactID, "/", format)
	if err != nil {
		return nil
	}
	return s.artStore.Nodes().DeleteById(ctx, node.ID)
}
