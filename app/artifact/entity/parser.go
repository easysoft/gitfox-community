package entity

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
)

// Parse 根据 entityId（格式 "kind.id"，如 "package.123"）反查出完整的 ArtifactEntity 树。
func Parse(ctx context.Context, store artifactStore.Interface, entityId string) (*model.ArtifactEntity, error) {
	if strings.Count(entityId, ".") != 1 {
		return nil, fmt.Errorf("invalid node_id %s", entityId)
	}

	idx := strings.Index(entityId, ".")
	typeStr := entityId[:idx]
	idStr := entityId[idx+1:]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, err
	}

	kind := model.ArtifactEntityKind(typeStr)
	switch kind {
	case model.ArtifactEntityKindAsset, model.ArtifactEntityKindVersion, model.ArtifactEntityKindPackage, model.ArtifactEntityKindGroup:
		p := &entityParser{artStore: store}
		return p.parse(ctx, kind, id)
	default:
		return nil, fmt.Errorf("unsupported artifact entity '%s'", entityId)
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
		if asset.GroupID.Ptr() != nil {
			elem.Parent, err = s.parse(ctx, model.ArtifactEntityKindGroup, asset.GroupID.Int64)
		} else if asset.VersionID.Ptr() != nil {
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
		if group.ParentID.Ptr() != nil {
			elem.Parent, err = s.parse(ctx, model.ArtifactEntityKindGroup, group.ParentID.Int64)
		} else {
			elem.Parent = &model.ArtifactEntity{Kind: model.ArtifactEntityKindFormat, Name: string(group.Format)}
		}
	}

	if err != nil {
		return nil, err
	}
	return &elem, nil
}
