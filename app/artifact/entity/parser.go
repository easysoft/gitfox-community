package entity

import (
	"context"
	"fmt"

	"github.com/easysoft/gitfox-community/app/constant"
	model "github.com/easysoft/gitfox-community/app/model/artifact"
	artifactStore "github.com/easysoft/gitfox-community/app/store/artifact"
	pkgErrors "github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
)

// Parse 根据 entityId（格式 "kind.id"，如 "package.123"）反查出完整的 ArtifactEntity 树。
func Parse(ctx context.Context, store artifactStore.Interface, entityId model.EntityID) (*model.ArtifactEntity, error) {

	splitEntityID := entityId.GetEntityID()
	kind := splitEntityID.Kind
	switch kind {
	case constant.ArtifactEntityKindAsset, constant.ArtifactEntityKindVersion, constant.ArtifactEntityKindPackage, constant.ArtifactEntityKindGroup, constant.ArtifactEntityKindArtifact:
		p := &entityParser{artStore: store}
		return p.parse(ctx, kind, splitEntityID.ID)
	default:
		return nil, pkgErrors.WithI18nKey(
			fmt.Errorf("unsupported artifact entity type: %s", kind),
			i18n.ArtifactSoftRemoveEntityTypeUnsupported,
		)
	}
}

type entityParser struct {
	artStore artifactStore.Interface
}

func (s *entityParser) parse(ctx context.Context, eType constant.ArtifactEntityKind, pk int64) (*model.ArtifactEntity, error) {
	var err error
	elem := model.ArtifactEntity{
		Kind: eType,
		Id:   pk,
	}

	switch eType {
	case constant.ArtifactEntityKindAsset:
		asset, e := s.artStore.Assets().GetById(ctx, pk)
		if e != nil {
			return nil, e
		}

		elem.Name = asset.Path
		if asset.GroupID.Ptr() != nil && asset.GroupID.Int64 != 0 {
			elem.Parent, err = s.parse(ctx, constant.ArtifactEntityKindGroup, asset.GroupID.Int64)
		} else if asset.VersionID.Ptr() != nil && asset.VersionID.Int64 != 0 {
			elem.Parent, err = s.parse(ctx, constant.ArtifactEntityKindVersion, asset.VersionID.Int64)
		}
	case constant.ArtifactEntityKindVersion:
		ver, e := s.artStore.Versions().GetByID(ctx, pk)
		if e != nil {
			return nil, e
		}
		elem.Name = ver.Version
		elem.Parent, err = s.parse(ctx, constant.ArtifactEntityKindPackage, ver.PackageID)
	case constant.ArtifactEntityKindPackage:
		pkg, e := s.artStore.Packages().GetByID(ctx, pk)
		if e != nil {
			return nil, e
		}

		elem.Name = pkg.Name
		if pkg.GroupID != 0 {
			elem.Parent, err = s.parse(ctx, constant.ArtifactEntityKindGroup, pkg.GroupID)
		} else {
			elem.Parent = &model.ArtifactEntity{Kind: constant.ArtifactEntityKindFormat, Name: string(pkg.Format)}
		}
	case constant.ArtifactEntityKindGroup:
		group, e := s.artStore.Groups().GetById(ctx, pk)
		if e != nil {
			return nil, e
		}

		elem.Name = group.Name
		if group.ParentID.Valid && group.ParentID.Int64 != 0 {
			elem.Parent, err = s.parse(ctx, constant.ArtifactEntityKindGroup, group.ParentID.Int64)
		} else {
			elem.Parent = &model.ArtifactEntity{Kind: constant.ArtifactEntityKindFormat, Name: string(group.Format)}
		}
	case constant.ArtifactEntityKindArtifact:
		// artifact 对应 ops_artifact (view) 本身，是权限/实体树的最顶层，无父节点。
		view, e := s.artStore.Views().GetByID(ctx, pk)
		if e != nil {
			return nil, e
		}
		elem.Name = view.Name
	}

	if err != nil {
		return nil, err
	}
	return &elem, nil
}
