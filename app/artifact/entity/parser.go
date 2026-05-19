package entity

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

// Parse 根据 entityId（格式 "kind.id"，如 "package.123"）反查出完整的 ArtifactEntity 树。
func Parse(ctx context.Context, store artifactStore.Interface, entityId string) (*model.ArtifactEntity, error) {
	if strings.Count(entityId, ".") != 1 {
		return nil, pkgErrors.WithI18nKey(
			fmt.Errorf("invalid artifact entity id format: %s", entityId),
			i18n.ArtifactSoftRemoveEntityIdInvalidFormat,
		)
	}

	idx := strings.Index(entityId, ".")
	typeStr := entityId[:idx]
	idStr := entityId[idx+1:]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(
			fmt.Errorf("invalid artifact entity id number: %s", entityId),
			i18n.ArtifactSoftRemoveEntityIdInvalidNumber,
		)
	}

	kind := model.ArtifactEntityKind(typeStr)
	switch kind {
	case model.ArtifactEntityKindAsset, model.ArtifactEntityKindVersion, model.ArtifactEntityKindPackage, model.ArtifactEntityKindGroup, model.ArtifactEntityKindArtifact:
		p := &entityParser{artStore: store}
		return p.parse(ctx, kind, id)
	default:
		return nil, pkgErrors.WithI18nKey(
			fmt.Errorf("unsupported artifact entity type: %s", typeStr),
			i18n.ArtifactSoftRemoveEntityTypeUnsupported,
		)
	}
}

type entityParser struct {
	artStore artifactStore.Interface
}

func (s *entityParser) parse(ctx context.Context, eType model.ArtifactEntityKind, pk int64) (*model.ArtifactEntity, error) {
	var err error
	elem := model.ArtifactEntity{
		Kind: eType,
		Id:   pk,
	}

	switch eType {
	case model.ArtifactEntityKindAsset:
		asset, e := s.artStore.Assets().GetById(ctx, pk)
		if e != nil {
			return nil, e
		}

		elem.Name = asset.Path
		if asset.GroupID.Ptr() != nil && asset.GroupID.Int64 != 0 {
			elem.Parent, err = s.parse(ctx, model.ArtifactEntityKindGroup, asset.GroupID.Int64)
		} else if asset.VersionID.Ptr() != nil && asset.VersionID.Int64 != 0 {
			elem.Parent, err = s.parse(ctx, model.ArtifactEntityKindVersion, asset.VersionID.Int64)
		}
	case model.ArtifactEntityKindVersion:
		ver, e := s.artStore.Versions().GetByID(ctx, pk)
		if e != nil {
			return nil, e
		}
		elem.Name = ver.Version
		elem.Parent, err = s.parse(ctx, model.ArtifactEntityKindPackage, ver.PackageID)
	case model.ArtifactEntityKindPackage:
		pkg, e := s.artStore.Packages().GetByID(ctx, pk)
		if e != nil {
			return nil, e
		}

		elem.Name = pkg.Name
		if pkg.GroupID != 0 {
			elem.Parent, err = s.parse(ctx, model.ArtifactEntityKindGroup, pkg.GroupID)
		} else {
			elem.Parent = &model.ArtifactEntity{Kind: model.ArtifactEntityKindFormat, Name: string(pkg.Format)}
		}
	case model.ArtifactEntityKindGroup:
		group, e := s.artStore.Groups().GetById(ctx, pk)
		if e != nil {
			return nil, e
		}

		elem.Name = group.Name
		if group.ParentID.Valid && group.ParentID.Int64 != 0 {
			elem.Parent, err = s.parse(ctx, model.ArtifactEntityKindGroup, group.ParentID.Int64)
		} else {
			elem.Parent = &model.ArtifactEntity{Kind: model.ArtifactEntityKindFormat, Name: string(group.Format)}
		}
	case model.ArtifactEntityKindArtifact:
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
