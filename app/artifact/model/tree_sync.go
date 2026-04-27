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
	if currNode, err := t.artStore.Nodes().GetByTarget(ctx, group.ViewID, artmodel.ArtifactTreeNodeTypeGroup, group.ID); err == nil {
		return currNode, nil
	}

	var parentNode *artmodel.ArtifactTreeNode
	var err error

	if group.ParentID.Ptr() == nil {
		parentNode, err = t.SyncTreeNodeForFormat(ctx, group.ViewID, group.Format)
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
		ParentID: null.IntFrom(parentNode.ID),
		ViewID:  group.ViewID,
		Name:     group.Name,
		Path:     joinPath([]string{parentNode.Path, group.Name}, "/"),
		Type:     artmodel.ArtifactTreeNodeTypeGroup,
		Format:   group.Format,
		TargetId: null.IntFrom(group.ID),
	}

	if err = t.upsertTreeNode(ctx, treeNode); err != nil {
		return nil, err
	}
	return treeNode, nil
}

func (t *Syncer) SyncTreeNodeForFormat(ctx context.Context, viewId int64, format artmodel.ArtifactFormat) (*artmodel.ArtifactTreeNode, error) {
	path := "/"
	node, err := t.artStore.Nodes().GetByPath(ctx, viewId, path, format)
	if err == nil {
		return node, nil
	}
	if !errors.Is(err, pkgErrors.ErrResourceNotFound) {
		return nil, err
	}
	newNode := &artmodel.ArtifactTreeNode{
		ViewID: viewId,
		Path:    path,
		Name:    string(format),
		Type:    artmodel.ArtifactTreeNodeTypeFormat,
		Format:  format,
	}
	if err = t.artStore.Nodes().Create(ctx, newNode); err != nil {
		return nil, err
	}
	return newNode, nil
}

func (t *Syncer) SyncTreeNodeForPackage(ctx context.Context, pkg *artmodel.ArtifactPackage) (*artmodel.ArtifactTreeNode, error) {
	// raw 格式的 package 可挂到 group，下其它格式始终直接挂到 format 根节点。
	if currNode, err := t.artStore.Nodes().GetByTarget(ctx, pkg.ViewID, artmodel.ArtifactTreeNodeTypePackage, pkg.ID); err == nil {
		return currNode, nil
	}

	var parentNode *artmodel.ArtifactTreeNode
	var err error

	if pkg.Format == artmodel.ArtifactRawFormat {
		if pkg.GroupID == 0 {
			parentNode, err = t.SyncTreeNodeForFormat(ctx, pkg.ViewID, pkg.Format)
		} else {
			g, e := t.artStore.Groups().GetById(ctx, pkg.GroupID)
			if e != nil {
				return nil, e
			}
			parentNode, err = t.SyncTreeNodeForGroup(ctx, g)
		}
	} else {
		parentNode, err = t.SyncTreeNodeForFormat(ctx, pkg.ViewID, pkg.Format)
	}

	if err != nil {
		return nil, err
	}

	treeNode := &artmodel.ArtifactTreeNode{
		ParentID: null.IntFrom(parentNode.ID),
		ViewID:  pkg.ViewID,
		Name:     pkg.Name,
		Path:     joinPath([]string{parentNode.Path, pkg.Name}, "/"),
		Type:     artmodel.ArtifactTreeNodeTypePackage,
		Format:   pkg.Format,
		TargetId: null.IntFrom(pkg.ID),
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
	if currNode, err := t.artStore.Nodes().GetByTarget(ctx, pkg.ViewID, artmodel.ArtifactTreeNodeTypeVersion, ver.ID); err == nil {
		return currNode, nil
	}
	parentNode, err := t.SyncTreeNodeForPackage(ctx, pkg)
	if err != nil {
		return nil, err
	}
	treeNode := &artmodel.ArtifactTreeNode{
		ParentID: null.IntFrom(parentNode.ID),
		ViewID:  pkg.ViewID,
		Name:     ver.Version,
		Path:     joinPath([]string{parentNode.Path, ver.Version}, "/"),
		Type:     artmodel.ArtifactTreeNodeTypeVersion,
		Format:   pkg.Format,
		TargetId: null.IntFrom(ver.ID),
	}
	if err = t.upsertTreeNode(ctx, treeNode); err != nil {
		return nil, err
	}
	return treeNode, nil
}

func (t *Syncer) upsertTreeNode(ctx context.Context, node *artmodel.ArtifactTreeNode) error {
	// 树节点以 owner+format+path 幂等；路径已存在时只纠正类型与目标实体，不重复建点。
	if upObj, e := t.artStore.Nodes().GetByPath(ctx, node.ViewID, node.Path, node.Format); e == nil {
		node.ID = upObj.ID
		if upObj.Type == node.Type && upObj.TargetId == node.TargetId {
			return nil
		}
		upObj.Type = node.Type
		upObj.TargetId = node.TargetId
		return t.artStore.Nodes().Update(ctx, upObj)
	} else if errors.Is(e, pkgErrors.ErrResourceNotFound) {
		return t.artStore.Nodes().Create(ctx, node)
	} else {
		return e
	}
}

func (t *Syncer) SyncRawGroup(ctx context.Context, pkg *artmodel.ArtifactPackage) error {
	// raw namespace 以 `.` 切分递归建组，末级 group 作为 package 的归属分组。
	groupNames := strings.Split(pkg.Namespace, ".")
	ge, e := t.artStore.Groups().RecursiveCreate(ctx, pkg.ViewID, groupNames, pkg.Format)
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
