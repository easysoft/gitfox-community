package model

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
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

// BuildTreeNode 构建所有节点的树
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

// SyncTreeNodeForGroup 同步 group 节点
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

// SyncTreeNodeForFormat 同步 format 节点
func (t *Syncer) SyncTreeNodeForFormat(ctx context.Context, artifactID int64, format artmodel.ArtifactFormat) (*artmodel.ArtifactTreeNode, error) {
	path := "/"
	node, err := t.artStore.Nodes().GetByPath(ctx, artifactID, path, format)
	if err == nil {
		if node.Type != artmodel.ArtifactTreeNodeTypeFormat || node.Format != format {
			node.Type = artmodel.ArtifactTreeNodeTypeFormat
			node.Format = format
			if err = t.artStore.Nodes().Update(ctx, node); err != nil {
				return nil, err
			}
		}
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

// SyncTreeNodeForPackage 同步 package 节点
func (t *Syncer) SyncTreeNodeForPackage(ctx context.Context, pkg *artmodel.ArtifactPackage) (*artmodel.ArtifactTreeNode, error) {
	return t.SyncTreeNodeForPackageInArtifact(ctx, pkg, pkg.ArtifactID)
}

// SyncTreeNodeForVersion 同步 version 节点
func (t *Syncer) SyncTreeNodeForVersion(ctx context.Context, ver *artmodel.ArtifactVersion, pkg *artmodel.ArtifactPackage) (*artmodel.ArtifactTreeNode, error) {
	if ver.PackageID != pkg.ID {
		return nil, fmt.Errorf("package id does not match")
	}

	artifactID := ver.ArtifactLibID
	if artifactID == 0 {
		artifactID = pkg.ArtifactID
	}

	if currNode, err := t.artStore.Nodes().GetByTarget(ctx, artifactID, artmodel.ArtifactTreeNodeTypeVersion, ver.ID); err == nil {
		return currNode, nil
	}

	parentNode, err := t.SyncTreeNodeForPackageInArtifact(ctx, pkg, artifactID)
	if err != nil {
		return nil, err
	}

	treeNode := &artmodel.ArtifactTreeNode{
		ParentID:   null.IntFrom(parentNode.ID),
		ArtifactID: artifactID,
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

// SyncTreeNodeForPackageInArtifact 同步 package 节点
func (t *Syncer) SyncTreeNodeForPackageInArtifact(ctx context.Context, pkg *artmodel.ArtifactPackage, artifactID int64) (*artmodel.ArtifactTreeNode, error) {
	if artifactID == 0 {
		artifactID = pkg.ArtifactID
	}

	if currNode, err := t.artStore.Nodes().GetByTarget(ctx, artifactID, artmodel.ArtifactTreeNodeTypePackage, pkg.ID); err == nil {
		return currNode, nil
	}

	var parentNode *artmodel.ArtifactTreeNode
	var err error

	if pkg.GroupID != 0 {
		g, e := t.artStore.Groups().GetById(ctx, pkg.GroupID)
		if e != nil {
			return nil, e
		}
		if g.ArtifactID == artifactID {
			parentNode, err = t.SyncTreeNodeForGroup(ctx, g)
		} else {
			parentNode, err = t.SyncTreeNodeForFormat(ctx, artifactID, pkg.Format)
		}
	} else {
		parentNode, err = t.SyncTreeNodeForFormat(ctx, artifactID, pkg.Format)
	}
	if err != nil {
		return nil, err
	}

	treeNode := &artmodel.ArtifactTreeNode{
		ParentID:   null.IntFrom(parentNode.ID),
		ArtifactID: artifactID,
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

// upsertTreeNode 插入或更新节点
func (t *Syncer) upsertTreeNode(ctx context.Context, node *artmodel.ArtifactTreeNode) error {
	// 树节点以 owner+format+path 幂等；路径已存在时只纠正类型与目标实体，不重复建点。
	upObj, err := t.artStore.Nodes().GetByPath(ctx, node.ArtifactID, node.Path, node.Format)
	if err != nil {
		if errors.Is(err, pkgErrors.ErrResourceNotFound) {
			created, err := t.artStore.Nodes().Create(ctx, &artmodel.CreateArtifactTreeNode{
				ParentID: node.ParentID, ArtifactID: node.ArtifactID, Name: node.Name,
				Path: node.Path, Type: node.Type, Format: node.Format, TargetID: node.TargetID,
			})
			if err != nil {
				return err
			}
			node.ID = created.ID
			return nil
		}
		return err
	}

	node.ID = upObj.ID
	if upObj.Type == node.Type && upObj.TargetID == node.TargetID {
		return nil
	}
	upObj.Type = node.Type
	upObj.TargetID = node.TargetID
	return t.artStore.Nodes().Update(ctx, upObj)
}

// SyncRawGroup 同步 raw namespace 分组
func (t *Syncer) SyncRawGroup(ctx context.Context, pkg *artmodel.ArtifactPackage) error {
	session, _ := req.GetAuthSession(ctx)
	// raw namespace 以 `.` 切分递归建组，末级 group 作为 package 的归属分组。
	groupNames := strings.Split(pkg.Namespace, ".")
	ge, e := t.artStore.Groups().RecursiveCreate(ctx, pkg.ArtifactID, groupNames, pkg.Format, session.User.Account)
	if e != nil {
		return e
	}
	pkg.GroupID = ge.ID
	return t.artStore.Packages().Update(ctx, pkg)
}

// joinPath 合并路径元素
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

// BuildPath 构建路径
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
