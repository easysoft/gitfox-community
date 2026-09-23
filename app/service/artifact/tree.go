// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package artifact

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	artModel "github.com/easysoft/gitfox-community/app/artifact/model"
	model "github.com/easysoft/gitfox-community/app/model/artifact"
	artifactStore "github.com/easysoft/gitfox-community/app/store/artifact"
	pkgErrors "github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

// RefreshTree 全量重建制品树节点。
// 遍历顺序：groups → packages(含 raw namespace 补建 group) → versions。
// 移植自原项目 /root/repo/gitfox/app/services/artifact/sync_tree_node.go:RefreshTreeNodes。
//
// format 含义：
//   - model.ArtifactAllFormat（"all"）或空值：全量刷新所有格式
//   - 其它具体 format：仅刷新该 format 下的 groups / packages / versions
func (s *Service) RefreshTree(ctx context.Context, format string) error {
	return s.refreshTree(ctx, format, s.artStore)
}

func (s *Service) refreshTree(ctx context.Context, format string, artStore artifactStore.Interface) error {
	syncer := artModel.NewSyncer(artStore)

	targetFormat := model.ArtifactFormat(strings.TrimSpace(format))
	refreshAll := targetFormat == "" || targetFormat == model.ArtifactAllFormat

	// 1. 同步所有 group 树节点
	groups, err := artStore.Groups().All(ctx)
	if err != nil {
		logger.GetLogger(ctx).Errorf("refresh tree: list groups failed: %v", err)
		return err
	}
	for _, group := range groups {
		if !refreshAll && group.Format != targetFormat {
			continue
		}
		_, err := syncer.SyncTreeNodeForGroup(ctx, group)
		if err != nil {
			logger.GetLogger(ctx).Errorf("refresh tree: sync group %d failed: %v", group.ID, err)
			return err
		}
	}

	// 2. 遍历所有 packages，同步 package + version 树节点
	return artStore.Packages().ListBatch(ctx, false, 100, func(pkg *model.ArtifactPackage) error {
		if !refreshAll && pkg.Format != targetFormat {
			return nil
		}

		// 2a. raw 格式 namespace 补建 group（namespace != "" 且 groupID==0 时）
		if pkg.Namespace != "" && pkg.GroupID == 0 {
			if sErr := syncer.SyncRawGroup(ctx, pkg); sErr != nil {
				return sErr
			}
		}

		// 2b. 同步 package 节点。
		// container package 只是同 scope 锚点，实际展示库由 version.ArtifactLibID 决定；
		// 这里不按 pkg.ArtifactID 预建 package 节点，避免锚点库出现空 package。
		if pkg.Format != model.ArtifactContainerFormat {
			if _, pErr := syncer.SyncTreeNodeForPackage(ctx, pkg); pErr != nil {
				return pErr
			}
		}

		// 2c. 同步该 package 下所有 version 节点。
		// container 会在 SyncTreeNodeForVersion 内按 version.ArtifactLibID 补建 package 父节点。
		dbVersions, vErr := artStore.Versions().Find(ctx, model.SearchVersionOption{PackageID: pkg.ID})
		if vErr != nil {
			return vErr
		}
		for _, ver := range dbVersions {
			if _, tErr := syncer.SyncTreeNodeForVersion(ctx, ver, pkg); tErr != nil {
				return tErr
			}
		}
		return nil
	})
}

// ListTree 是通用树查询入口。
// 语义：
//  1. nodeID > 0 时，优先按节点ID定位当前节点，只支持单层查询；
//  2. path 不带 /** 时，返回“当前层”的平铺节点列表；
//  3. path 为 /** 或 /xxx/** 时，返回从该路径开始的递归树；
//  4. format=all 或留空：
//     - path == "/"：返回各 format 根节点（递归模式下展开各自子树）；
//     - path == "/xxx[/**]"：遍历所有 format 根，按 (artifactID, path, format) 逐个 GetByPath，
//     找到即进入该 format 的分支结果，未命中跳过；多 format 同名路径会合并返回。
//  5. level 表示最大可见层级：format/group/package/version/asset。
func (s *Service) ListTree(ctx context.Context, artifactID int64, filter *model.ArtifactTreeFilter) ([]*model.ArtifactTreeRes, error) {
	if filter == nil {
		filter = &model.ArtifactTreeFilter{}
	}
	level, err := normalizeTreeLevel(filter.Level)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactTreeLevelInvalid)
	}
	path, recursive := normalizeTreePath(filter.Path)

	// 1) NodeId 分支：精确定位单节点。
	if filter.NodeId > 0 {
		if recursive {
			return nil, pkgErrors.WithI18nKey(errors.New("recursive query is not supported when NodeId is provided"), i18n.ArtifactTreeNodeRecursiveConflict)
		}
		currNode, err := s.artStore.Nodes().GetById(ctx, filter.NodeId)
		if err != nil {
			return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactTreeNodeListError)
		}
		if currNode.ArtifactID != artifactID {
			return nil, pkgErrors.WithI18nKey(errors.New("node does not belong to current artifact"), i18n.ArtifactTreeNodeNotInArtifact)
		}
		if level == model.ArtifactNodeLevelFormat {
			return []*model.ArtifactTreeRes{convertTreeNodeToRes(currNode, level)}, nil
		}
		return s.listTreeCurrentLevel(ctx, currNode, level)
	}

	// 2) format 留空视为 all，统一走跨格式分支。
	isAll := filter.Format == "" || filter.Format == model.ArtifactAllFormat

	if isAll {
		roots, err := s.artStore.Nodes().ListFormats(ctx, artifactID)
		if err != nil {
			return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactTreeFormatListError)
		}
		// 2.a path 为根：保持旧语义，返回 format 森林 / 只返 roots。
		if path == model.ArtifactNodeRoot {
			if level == model.ArtifactNodeLevelFormat || !recursive {
				return convertTreeNodesToRes(roots, level), nil
			}
			return s.buildFormatForest(ctx, roots, level)
		}
		// 2.b path 指向具体节点：跨 format 逐个匹配，找到即入结果集。
		return s.listTreeForAllFormats(ctx, artifactID, path, recursive, level, roots)
	}

	// 3) 具体 format + path 分支。
	currNode, err := s.artStore.Nodes().GetByPath(ctx, artifactID, path, filter.Format)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactTreeNodeListError)
	}
	if level == model.ArtifactNodeLevelFormat {
		if path == model.ArtifactNodeRoot {
			return []*model.ArtifactTreeRes{convertTreeNodeToRes(currNode, level)}, nil
		}
		return []*model.ArtifactTreeRes{}, nil
	}
	if !recursive {
		return s.listTreeCurrentLevel(ctx, currNode, level)
	}
	if path == model.ArtifactNodeRoot {
		return s.buildTreeChildren(ctx, currNode, level)
	}
	root := convertTreeNodeToRes(currNode, level)
	root.Children, err = s.buildTreeChildren(ctx, currNode, level)
	if err != nil {
		return nil, err
	}
	root.Leaf = len(root.Children) == 0
	return []*model.ArtifactTreeRes{root}, nil
}

// listTreeCurrentLevel 处理“非递归模式”的当前层查询。
// 它只返回 node 的直接下一层：
//  1. 普通节点返回直接 children；
//  2. group 节点在 level=asset 时，会把 group 下直接资产一并拼进来；
//  3. version 节点在 level=asset 时，直接下钻返回该 version 下的资产。
//
// 关于 container：asset 从不写入 ops_artifact_treenode（tree_sync.go 没有 SyncTreeNodeForAsset），
// 所以 container version 同样必须走 listVersionAssetAsTree（即按 versionID 直查 asset 表），
// 否则 ListByParentId 会扑空，前端永远看不到 container 的 main asset。
// 注意 container layer blob 走 isolated（VersionID=0，见 manifest_upload.go::AttrAssetIsolated），
// 不会被这里错列；这一段只会返回挂在 version 上的 main asset（manifest）。
func (s *Service) listTreeCurrentLevel(ctx context.Context, node *model.ArtifactTreeNode, level string) ([]*model.ArtifactTreeRes, error) {
	if node.Type == model.ArtifactTreeNodeTypeVersion && node.TargetID > 0 {
		if level == model.ArtifactNodeLevelAsset {
			return s.listVersionAssetAsTree(ctx, node)
		}
		return []*model.ArtifactTreeRes{convertTreeNodeToRes(node, level)}, nil
	}
	children, err := s.artStore.Nodes().ListByParentId(ctx, node.ID)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactTreeNodeListError)
	}
	// 先按 level 裁掉超过目标层级的节点，再做响应模型转换。
	children = filterTreeNodesByLevel(children, level)
	data := convertTreeNodesToRes(children, level)
	if node.Type == model.ArtifactTreeNodeTypeGroup && node.TargetID > 0 && level == model.ArtifactNodeLevelAsset {
		return s.appendGroupAssets(ctx, node.TargetID, data)
	}
	return data, nil
}

// listTreeForAllFormats 跨 format 按 path 查询：逐个 format 根 GetByPath，
// 命中则按 recursive / level 铺开结果，未命中的 format 静默跳过。
// 多 format 下若存在同名路径，结果会各自作为顶层元素返回。
func (s *Service) listTreeForAllFormats(ctx context.Context, artifactID int64, path string, recursive bool, level string, roots []*model.ArtifactTreeNode) ([]*model.ArtifactTreeRes, error) {
	data := make([]*model.ArtifactTreeRes, 0, len(roots))
	for _, root := range roots {
		currNode, err := s.artStore.Nodes().GetByPath(ctx, artifactID, path, root.Format)
		if err != nil {
			// 此 format 下不存在该 path：跳过，不让单 format 缺失拖垮整批查询。
			if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
				continue
			}
			return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactTreeNodeListError)
		}
		if level == model.ArtifactNodeLevelFormat {
			// 非根路径 + level=format 无意义，按既有单 format 语义返空。
			continue
		}
		if !recursive {
			items, lErr := s.listTreeCurrentLevel(ctx, currNode, level)
			if lErr != nil {
				return nil, lErr
			}
			data = append(data, items...)
			continue
		}
		item := convertTreeNodeToRes(currNode, level)
		children, bErr := s.buildTreeChildren(ctx, currNode, level)
		if bErr != nil {
			return nil, bErr
		}
		item.Children = children
		item.Leaf = len(children) == 0
		data = append(data, item)
	}
	return data, nil
}

// buildFormatForest 将多个 format 根节点分别递归展开，组装成整库的树森林。
// 仅在 format=all 且 path 使用递归模式时调用。
func (s *Service) buildFormatForest(ctx context.Context, roots []*model.ArtifactTreeNode, level string) ([]*model.ArtifactTreeRes, error) {
	data := make([]*model.ArtifactTreeRes, 0, len(roots))
	for _, root := range roots {
		item := convertTreeNodeToRes(root, level)
		children, err := s.buildTreeChildren(ctx, root, level)
		if err != nil {
			return nil, err
		}
		item.Children = children
		item.Leaf = len(children) == 0
		data = append(data, item)
	}
	return data, nil
}

// buildTreeChildren 处理“递归模式”的子树构建。
// 递归终止条件：
//  1. 当前已到 version 且 level < asset；
//  2. 当前没有更多 children；
//  3. level 过滤后不再允许继续向下。
//
// container 同样按 version → asset 下钻：asset 不在 tree 表里，必须走 listVersionAssetAsTree
// 直查 ops_artifact_assets。container layer 走 isolated（VersionID=0），不会被这里错列。
func (s *Service) buildTreeChildren(ctx context.Context, node *model.ArtifactTreeNode, level string) ([]*model.ArtifactTreeRes, error) {
	if node.Type == model.ArtifactTreeNodeTypeVersion && node.TargetID > 0 {
		if level == model.ArtifactNodeLevelAsset {
			return s.listVersionAssetAsTree(ctx, node)
		}
		return nil, nil
	}
	children, err := s.artStore.Nodes().ListByParentId(ctx, node.ID)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactTreeNodeListError)
	}
	children = filterTreeNodesByLevel(children, level)
	data := make([]*model.ArtifactTreeRes, 0, len(children))
	for _, child := range children {
		item := convertTreeNodeToRes(child, level)
		item.Children, err = s.buildTreeChildren(ctx, child, level)
		if err != nil {
			return nil, err
		}
		item.Leaf = len(item.Children) == 0
		data = append(data, item)
	}
	if node.Type == model.ArtifactTreeNodeTypeGroup && node.TargetID > 0 && level == model.ArtifactNodeLevelAsset {
		return s.appendGroupAssets(ctx, node.TargetID, data)
	}
	return data, nil
}

// listVersionAssetAsTree 将 version 节点下的资产转换成树节点返回。
// 参考项目里 version 在 asset 级别会直接展开成文件，这里保持同样语义。
func (s *Service) listVersionAssetAsTree(ctx context.Context, node *model.ArtifactTreeNode) ([]*model.ArtifactTreeRes, error) {
	assets, err := s.artStore.ListAssets(ctx, node.TargetID)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactAssetQueryError)
	}
	info := parseArtifactTreePath(node.Path)
	data := make([]*model.ArtifactTreeRes, 0, len(assets))
	for _, asset := range assets {
		base := filepath.Base(asset.Path)
		// asset 不在 ops_artifact_tree_nodes 表，无树节点主键 → metadata.id 留空（按 omitempty 不输出）。
		// EntityID 用 "asset.{id}" 引用串，跨接口定位实体仍走 asset.id。
		meta := &model.ArtifactTreeNodeMeta{Type: model.ArtifactTreeNodeTypeAsset}
		if info != nil {
			m := *info
			m.Id = 0
			m.Type = model.ArtifactTreeNodeTypeAsset
			meta = &m
		}
		meta.EntityID = (&model.ArtifactTreeNodeId{Type: model.ArtifactTreeNodeTypeAsset, Pk: asset.Id}).String()
		item := &model.ArtifactTreeRes{Name: base, Path: strings.TrimRight(node.Path, "/") + "/" + base, Leaf: true, Format: node.Format, Metadata: meta}
		// 填充 nodeInfo
		if nodeInfo, e := s.getAssetNodeInfo(ctx, asset.Id); e == nil && nodeInfo.Status == string(model.ArtifactStatusOK) {
			item.NodeInfo = nodeInfo
		}
		data = append(data, item)
	}
	return data, nil
}

// appendGroupAssets 把 generic/group 下面直接挂载的资产追加到当前结果中。
// 这一步只在 level=asset 时执行；group/package/version 级别不会把文件混入返回。
func (s *Service) appendGroupAssets(ctx context.Context, groupID int64, data []*model.ArtifactTreeRes) ([]*model.ArtifactTreeRes, error) {
	g, err := s.artStore.Groups().GetEndpointById(ctx, groupID)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactGroupQueryError)
	}
	assets, err := s.artStore.ListGroupAssets(ctx, groupID)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactAssetQueryError)
	}
	for _, asset := range assets {
		base := filepath.Base(asset.Path)
		// asset 不在 ops_artifact_tree_nodes 表，无树节点主键 → metadata.id 留空（按 omitempty 不输出）。
		// EntityID 用 "asset.{id}" 引用串，跨接口定位实体仍走 asset.id。
		entityID := (&model.ArtifactTreeNodeId{Type: model.ArtifactTreeNodeTypeAsset, Pk: asset.Id}).String()
		item := &model.ArtifactTreeRes{Name: base, Path: "/" + g.GroupPath() + "/" + base, Leaf: true, Format: model.ArtifactFormat(g.Format), Metadata: &model.ArtifactTreeNodeMeta{Type: model.ArtifactTreeNodeTypeAsset, Name: base, EntityID: entityID}}
		item.NodeInfo = &model.ArtifactAssetInfo{
			ID:          asset.Id,
			Status:      string(model.ArtifactStatusOK),
			Format:      model.ArtifactFormat(g.Format),
			Path:        asset.Path,
			ContentType: asset.ContentType,
			Size:        asset.Size,
			CreatorName: asset.CreatorName,
			EditedBy:    asset.EditedBy,
			Created:     asset.Created,
			Updated:     asset.Updated,
			CheckSum:    asset.CheckSum,
		}
		data = append(data, item)
	}
	return data, nil
}
