package model

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	artmodel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"

	"github.com/guregu/null"
)

type Syncer struct {
	artStore artifactStore.Interface
}

func NewSyncer(artStore artifactStore.Interface) *Syncer {
	return &Syncer{artStore: artStore}
}

func (t *Syncer) BuildTreeNode(ctx context.Context) error {
	groups, err := t.artStore.Groups().All(ctx)
	if err != nil {
		return err
	}
	for _, group := range groups {
		if _, err = t.SyncTreeNodeForGroup(ctx, group); err != nil {
			return err
		}
	}
	return nil
}

func (t *Syncer) SyncTreeNodeForGroup(ctx context.Context, group *artmodel.ArtifactGroup) (*artmodel.ArtifactTreeNode, error) {
	if currNode, err := t.artStore.Nodes().GetByTarget(ctx, group.ArtifactID, artmodel.ArtifactTreeNodeTypeGroup, group.ID); err == nil {
		return currNode, nil
	}

	var parentNode *artmodel.ArtifactTreeNode
	var err error

	if !group.ParentID.Valid || group.ParentID.Int64 == 0 {
		parentNode, err = t.SyncTreeNodeForFormat(ctx, group.ArtifactID, group.Format)
		if err != nil {
			return nil, err
		}
	} else {
		parentGroup, e := t.artStore.Groups().GetById(ctx, group.ParentID.Int64)
		if e != nil {
			return nil, e
		}
		parentNode, err = t.SyncTreeNodeForGroup(ctx, parentGroup)
		if err != nil {
			return nil, err
		}
	}

	treeNode := &artmodel.ArtifactTreeNode{
		ParentID:   null.IntFrom(parentNode.ID),
		ArtifactID: group.ArtifactID,
		Name:       group.Name,
		Path:       joinPath([]string{parentNode.Path, group.Name}, "/"),
		Type:       artmodel.ArtifactTreeNodeTypeGroup,
		Format:     group.Format,
		TargetID:   null.IntFrom(group.ID),
	}

	if err = t.upsertTreeNode(ctx, treeNode); err != nil {
		return nil, err
	}
	return treeNode, nil
}

func (t *Syncer) SyncTreeNodeForFormat(ctx context.Context, artifactID int64, format artmodel.ArtifactFormat) (*artmodel.ArtifactTreeNode, error) {
	path := "/"
	node, err := t.artStore.Nodes().GetByPath(ctx, artifactID, path, format)
	if err == nil {
		return node, nil
	}
	if !errors.Is(err, pkgErrors.ErrResourceNotFound) {
		return nil, err
	}
	newNode := &artmodel.CreateArtifactTreeNode{
		ArtifactID: artifactID,
		Path:       path,
		Name:       string(format),
		Type:       artmodel.ArtifactTreeNodeTypeFormat,
		Format:     format,
	}
	created, err := t.artStore.Nodes().Create(ctx, newNode)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (t *Syncer) SyncTreeNodeForPackage(ctx context.Context, pkg *artmodel.ArtifactPackage) (*artmodel.ArtifactTreeNode, error) {
	// 所有格式：GroupID != 0 则挂到对应 group，否则直接挂到 format 根节点。
	// raw 通过 SyncRawGroup 回填 GroupID；container 通过 getOrCreateGroup 预先建好 GroupID。
	if currNode, err := t.artStore.Nodes().GetByTarget(ctx, pkg.ArtifactID, artmodel.ArtifactTreeNodeTypePackage, pkg.ID); err == nil {
		return currNode, nil
	}

	var parentNode *artmodel.ArtifactTreeNode
	var err error

	if pkg.GroupID != 0 {
		g, e := t.artStore.Groups().GetById(ctx, pkg.GroupID)
		if e != nil {
			return nil, e
		}
		parentNode, err = t.SyncTreeNodeForGroup(ctx, g)
	} else {
		parentNode, err = t.SyncTreeNodeForFormat(ctx, pkg.ArtifactID, pkg.Format)
	}

	if err != nil {
		return nil, err
	}

	treeNode := &artmodel.ArtifactTreeNode{
		ParentID:   null.IntFrom(parentNode.ID),
		ArtifactID: pkg.ArtifactID,
		Name:       pkg.Name,
		Path:       joinPath([]string{parentNode.Path, pkg.Name}, "/"),
		Type:       artmodel.ArtifactTreeNodeTypePackage,
		Format:     pkg.Format,
		TargetID:   null.IntFrom(pkg.ID),
	}

	if err = t.upsertTreeNode(ctx, treeNode); err != nil {
		return nil, err
	}
	return treeNode, nil
}

func (t *Syncer) SyncTreeNodeForVersion(ctx context.Context, ver *artmodel.ArtifactVersion, pkg *artmodel.ArtifactPackage) (*artmodel.ArtifactTreeNode, error) {
	if ver.PackageID != pkg.ID {
		return nil, fmt.Errorf("package id does not match")
	}
	if currNode, err := t.artStore.Nodes().GetByTarget(ctx, pkg.ArtifactID, artmodel.ArtifactTreeNodeTypeVersion, ver.ID); err == nil {
		return currNode, nil
	}
	parentNode, err := t.SyncTreeNodeForPackage(ctx, pkg)
	if err != nil {
		return nil, err
	}
	treeNode := &artmodel.ArtifactTreeNode{
		ParentID:   null.IntFrom(parentNode.ID),
		ArtifactID: pkg.ArtifactID,
		Name:       ver.Version,
		Path:       joinPath([]string{parentNode.Path, ver.Version}, "/"),
		Type:       artmodel.ArtifactTreeNodeTypeVersion,
		Format:     pkg.Format,
		TargetID:   null.IntFrom(ver.ID),
	}
	if err = t.upsertTreeNode(ctx, treeNode); err != nil {
		return nil, err
	}
	return treeNode, nil
}

func (t *Syncer) upsertTreeNode(ctx context.Context, node *artmodel.ArtifactTreeNode) error {
	// 树节点以 owner+format+path 幂等；路径已存在时只纠正类型与目标实体，不重复建点。
	if upObj, e := t.artStore.Nodes().GetByPath(ctx, node.ArtifactID, node.Path, node.Format); e == nil {
		node.ID = upObj.ID
		if upObj.Type == node.Type && upObj.TargetID == node.TargetID {
			return nil
		}
		upObj.Type = node.Type
		upObj.TargetID = node.TargetID
		return t.artStore.Nodes().Update(ctx, upObj)
	} else if errors.Is(e, pkgErrors.ErrResourceNotFound) {
		created, err := t.artStore.Nodes().Create(ctx, &artmodel.CreateArtifactTreeNode{
			ParentID: node.ParentID, ArtifactID: node.ArtifactID, Name: node.Name,
			Path: node.Path, Type: node.Type, Format: node.Format, TargetID: node.TargetID,
		})
		if err != nil {
			return err
		}
		node.ID = created.ID
		return nil
	} else {
		return e
	}
}

func (t *Syncer) SyncRawGroup(ctx context.Context, pkg *artmodel.ArtifactPackage) error {
	// raw namespace 以 `.` 切分递归建组，末级 group 作为 package 的归属分组。
	groupNames := strings.Split(pkg.Namespace, ".")
	ge, e := t.artStore.Groups().RecursiveCreate(ctx, pkg.ArtifactID, groupNames, pkg.Format)
	if e != nil {
		return e
	}
	pkg.GroupID = ge.ID
	return t.artStore.Packages().Update(ctx, pkg)
}

func joinPath(arr []string, elem string) string {
	var buf bytes.Buffer
	if arr[0] != "/" {
		buf.WriteString(elem)
		buf.WriteString(strings.TrimPrefix(strings.TrimSuffix(arr[0], elem), elem))
	}
	num := len(arr)
	for i := 1; i < num; i++ {
		v := arr[i]
		buf.WriteString(elem)
		buf.WriteString(strings.TrimPrefix(strings.TrimSuffix(v, elem), elem))
	}
	return buf.String()
}

func BuildPath(namespace, name, version string) (string, error) {
	if namespace == "" && name == "" {
		return "", errors.New("empty namespace and name")
	}
	parts := make([]string, 0)
	if namespace != "" {
		parts = strings.Split(namespace, ".")
	}
	if name == "" {
		return "/" + strings.Join(parts, "/"), nil
	}
	parts = append(parts, name)
	if version == "" {
		return "/" + strings.Join(parts, "/"), nil
	}
	parts = append(parts, version)
	return "/" + strings.Join(parts, "/"), nil
}
