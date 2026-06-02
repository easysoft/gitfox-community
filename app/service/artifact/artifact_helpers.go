package artifact

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
)

// --- 制品相关 ---

// populateNodeInfoAssetFields 将资产基础字段填充到 nodeInfo 响应。
func populateNodeInfoAssetFields(info *model.ArtifactAssetInfo, asset *model.ArtifactAssetsRes) {
	if info == nil || asset == nil {
		return
	}

	info.Path = asset.Path
	info.ID = asset.Id
	info.ContentType = asset.ContentType
	info.Size = asset.Size
	info.CreatorName = asset.CreatorName
	info.EditedBy = asset.EditedBy
	info.Created = asset.Created
	info.Updated = asset.Updated
	info.Link = asset.Link
	info.CheckSum = asset.CheckSum
}

// buildArtifactAssetLink 生成普通制品资产的下载路径。
func buildArtifactAssetLink(spaceName, format, group, packageName, version, assetPath string) string {
	if spaceName == "" || format == "" || assetPath == "" {
		return ""
	}

	segments := []string{"_artifacts", spaceName, format}
	switch model.ArtifactFormat(format) {
	case model.ArtifactFileFormat:
		segments = appendSlashSegments(segments, group)
	case model.ArtifactHelmFormat:
		// helm 仓库直接以文件路径挂载在格式根下。
	default:
		if group != "" {
			segments = append(segments, strings.Split(strings.Trim(group, "."), ".")...)
		}
		if packageName != "" {
			segments = append(segments, packageName)
		}
		if version != "" {
			segments = append(segments, version)
		}
	}
	segments = appendSlashSegments(segments, assetPath)

	link, err := url.JoinPath("/", segments...)
	if err != nil {
		return ""
	}
	return link
}

// buildContainerPullLink 生成容器镜像可直接 docker pull 的引用。
func (s *Service) buildContainerPullLink(view *adapter.ViewDescriptor, pkg *model.ArtifactPackage, ver *model.ArtifactVersion) string {
	if view == nil || pkg == nil || ver == nil || view.ArtifactName == "" || pkg.Name == "" || ver.Version == "" {
		return ""
	}

	// "global/docker_global_dev/pipeline-helper:v1n"
	imageRef := strings.Join([]string{
		containerDomainID(view),
		view.ArtifactName,
		pkg.Name,
	}, "/") + ":" + ver.Version

	registry := s.containerRegistryReference()
	if registry == "" {
		return imageRef
	}
	return strings.TrimRight(registry, "/") + "/" + imageRef
}

// containerDomainID 根据视图作用域生成容器路由中的 domainid。
func containerDomainID(view *adapter.ViewDescriptor) string {
	if view != nil && view.Repo != nil && view.Repo.ID > 0 {
		return "repo" + strconv.FormatInt(view.Repo.ID, 10)
	}
	if view != nil && view.Space != nil && view.Space.ID > 0 {
		return "space" + strconv.FormatInt(view.Space.ID, 10)
	}
	return "global"
}

// containerRegistryReference 生成 Docker 引用需要的 registry 前缀。
func (s *Service) containerRegistryReference() string {
	if s == nil || s.urlProvider == nil {
		return ""
	}

	registryURL := s.urlProvider.GenerateRegistryURL()
	if registryURL == nil {
		return ""
	}

	return registryURL.Host
}

// appendSlashSegments 将斜杠路径拆成 URL path 段。
func appendSlashSegments(dst []string, p string) []string {
	trimmed := strings.Trim(p, "/")
	if trimmed == "" {
		return dst
	}
	return append(dst, strings.Split(trimmed, "/")...)
}

// nodeInfoSpaceName 取 nodeInfo 响应中的空间名。
func nodeInfoSpaceName(view *adapter.ViewDescriptor) string {
	if view != nil && view.Space != nil {
		return view.Space.Name
	}
	return ""
}

// notFoundNodeInfo 返回资源不存在的 nodeInfo 响应。
func notFoundNodeInfo() *model.ArtifactAssetInfo {
	return &model.ArtifactAssetInfo{Status: string(model.ArtifactStatusNotFound)}
}

// convertTreeNodesToRes 批量转换树节点为接口响应。
func convertTreeNodesToRes(inputs []*model.ArtifactTreeNode, level string) []*model.ArtifactTreeRes {
	out := make([]*model.ArtifactTreeRes, 0, len(inputs))
	for _, n := range inputs {
		out = append(out, convertTreeNodeToRes(n, level))
	}
	return out
}

// --- Tree相关 ---

// normalizeTreeLevel 规范化树查询层级参数。
func normalizeTreeLevel(level string) (string, error) {
	if level == "" {
		return model.ArtifactNodeLevelAsset, nil
	}
	switch level {
	case model.ArtifactNodeLevelFormat, model.ArtifactNodeLevelGroup, model.ArtifactNodeLevelPackage, model.ArtifactNodeLevelVersion, model.ArtifactNodeLevelAsset:
		return level, nil
	default:
		return "", fmt.Errorf("unsupported level: %s", level)
	}
}

// normalizeTreePath 规范化树查询路径，并识别是否递归查询。
func normalizeTreePath(path string) (string, bool) {
	p := strings.TrimSpace(path)
	if p == "" {
		return model.ArtifactNodeRoot, false
	}
	recursive := strings.HasSuffix(p, "/**")
	p = strings.TrimSuffix(p, "/**")
	if p == "" {
		p = model.ArtifactNodeRoot
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if p != "/" {
		p = strings.TrimRight(p, "/")
	}
	return p, recursive
}

// filterTreeNodesByLevel 按层级裁剪树节点。
func filterTreeNodesByLevel(inputs []*model.ArtifactTreeNode, level string) []*model.ArtifactTreeNode {
	out := make([]*model.ArtifactTreeNode, 0, len(inputs))
	for _, n := range inputs {
		if treeNodeLevelRank(n.Type) <= treeLevelRank(level) {
			out = append(out, n)
		}
	}
	return out
}

// treeLevelRank 返回请求层级的排序等级。
func treeLevelRank(level string) int {
	switch level {
	case model.ArtifactNodeLevelFormat:
		return 0
	case model.ArtifactNodeLevelGroup:
		return 1
	case model.ArtifactNodeLevelPackage:
		return 2
	case model.ArtifactNodeLevelVersion:
		return 3
	default:
		return 4
	}
}

// treeNodeLevelRank 返回树节点类型的排序等级。
func treeNodeLevelRank(t model.ArtifactTreeNodeType) int {
	switch t {
	case model.ArtifactTreeNodeTypeFormat:
		return 0
	case model.ArtifactTreeNodeTypeGroup:
		return 1
	case model.ArtifactTreeNodeTypePackage:
		return 2
	case model.ArtifactTreeNodeTypeVersion:
		return 3
	case model.ArtifactTreeNodeTypeAsset:
		return 4
	default:
		return 0
	}
}

// parseArtifactTreePath 从树节点路径中解析包、分组和版本信息。
func parseArtifactTreePath(path string) *model.ArtifactTreeNodeMeta {
	frames := strings.Split(strings.TrimLeft(path, "/"), "/")
	if len(frames) < 2 {
		return nil
	}
	meta := &model.ArtifactTreeNodeMeta{Name: frames[len(frames)-2], Version: frames[len(frames)-1]}
	if len(frames) > 2 {
		meta.Group = strings.Join(frames[:len(frames)-2], ".")
	}
	return meta
}

// convertTreeNodeToRes 将存储层树节点转换为接口返回模型。
//
// metadata.id 始终为 ops_artifact_tree_nodes.id（树节点主键）；
// asset 节点不写入 tree_nodes 表，由调用方决定是否输出（默认留空）。
// metadata.entityID 是 "kind.pk" 实体引用串，供跨接口定位实体使用。
func convertTreeNodeToRes(n *model.ArtifactTreeNode, level string) *model.ArtifactTreeRes {
	item := &model.ArtifactTreeRes{Name: n.Name, Path: n.Path, Format: n.Format, Leaf: n.Type == model.ArtifactTreeNodeTypeAsset}
	if n.Type == model.ArtifactTreeNodeTypeVersion {
		// version 节点是否叶子完全取决于查询是否下钻到 asset 级：
		//   - level != asset：用户没要 asset 层，version 视为叶子；
		//   - level == asset：version 下还有 manifest/file asset 子节点，必须保持非叶子。
		// container 的 version 同样挂着 manifest blob asset，不做格式特殊化。
		item.Leaf = level != model.ArtifactNodeLevelAsset
	}
	if n.Type == model.ArtifactTreeNodeTypeVersion {
		item.Metadata = parseArtifactTreePath(n.Path)
		if item.Metadata == nil {
			item.Metadata = &model.ArtifactTreeNodeMeta{}
		}
		item.Metadata.Type = model.ArtifactTreeNodeTypeVersion
	} else {
		item.Metadata = &model.ArtifactTreeNodeMeta{Type: n.Type}
	}
	// Id 取 tree_nodes.id（树节点主键）；EntityID 用 "kind.pk" 实体串。
	if n.ID > 0 {
		item.Metadata.Id = n.ID
	}
	if n.TargetID.Valid {
		item.Metadata.EntityID = (&model.ArtifactTreeNodeId{Type: n.Type, Pk: n.TargetID.Int64}).String()
	}
	if n.Type == model.ArtifactTreeNodeTypeFormat {
		item.Metadata.Name = n.Name
		if n.ArtifactID > 0 {
			item.Metadata.EntityID = (&model.ArtifactTreeNodeId{Type: model.ArtifactTreeNodeTypeArtifact, Pk: n.ArtifactID}).String()
		}
	}
	return item
}

// --- 统计相关 ---

// artifactViewScope 将版本统计信息转换为作用域标识。
func artifactViewScope(info *model.ArtifactVersionInfo) string {
	if info.ViewRepoID > 0 {
		return "repo"
	}
	if info.ViewSpaceID > 0 {
		return "space"
	}
	return "global"
}

// artifactScopeKey 生成统计汇总使用的作用域 key。
func artifactScopeKey(info *model.ArtifactVersionInfo) string {
	switch artifactViewScope(info) {
	case "repo":
		return "repo:" + info.RepoName
	case "space":
		return "space:" + info.SpaceName
	default:
		return "global"
	}
}

// addStats 将一组统计累加到目标统计对象。
func addStats(dst, src *model.ArtifactStatsObject) {
	dst.Total += src.Total
	dst.Active += src.Active
	dst.Deleted += src.Deleted
}
