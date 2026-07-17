package artifact

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter/container"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
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
// 空间级: /api/v2/space/{spaceCode}/_{format}/{viewName}/[groupPath/][packageName/][version/]{assetPath}
// 代码库级: /api/v2/space/{spaceCode}/repo/{repoIdentifier}/_{format}/{viewName}/[groupPath/][packageName/][version/]{assetPath}
func buildArtifactAssetLink(spaceCode, repoIdentifier, viewName, format, group, packageName, version, assetPath string) string {
	if spaceCode == "" || format == "" || assetPath == "" {
		return ""
	}

	segments := []string{"api", "v2", "space", spaceCode}
	if repoIdentifier != "" {
		segments = append(segments, "repo", repoIdentifier)
	}
	segments = append(segments, "_"+format)
	if viewName != "" {
		segments = append(segments, viewName)
	}

	switch model.ArtifactFormat(format) {
	case model.ArtifactHelmFormat:
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

	escaped := make([]string, len(segments))
	for i, seg := range segments {
		escaped[i] = url.PathEscape(seg)
	}
	return "/" + strings.Join(escaped, "/")
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
		return view.Space.Code
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

// nodeInfoSpaceCode 取 nodeInfo 响应中的空间 Code。
func nodeInfoSpaceCode(view *adapter.ViewDescriptor) string {
	if view != nil && view.Space != nil {
		return view.Space.Code
	}
	return ""
}

// nodeInfoRepoIdentifier 取 nodeInfo 响应中的代码库标识。
func nodeInfoRepoIdentifier(view *adapter.ViewDescriptor) string {
	if view != nil && view.Repo != nil {
		return view.Repo.Name
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

// --- container ---

// containerImagesToModel 将 container.ImageMeta 转换为 model.ArtifactContainerImage。
func containerImagesToModel(images []*container.ImageMeta) []*model.ArtifactContainerImage {
	out := make([]*model.ArtifactContainerImage, 0, len(images))
	for _, image := range images {
		if image == nil {
			continue
		}
		out = append(out, &model.ArtifactContainerImage{
			Digest: image.Digest.String(),
			OS:     image.OS,
			Arch:   image.Arch,
			Size:   image.Size,
		})
	}
	return out
}

// --- upload ---

// wrapUploadError 若 err 已带有专属 i18n key（非默认），直接透传，
// 否则套上通用 fallback，避免覆盖 adapter 层给出的精准提示。
func wrapUploadError(err error, fallback i18n.I18nKey) error {
	if err == nil {
		return nil
	}
	if key := pkgErrors.GetI18nKey(err); key != "" && key != i18n.ErrorDefault {
		return err
	}
	return pkgErrors.WithI18nKey(err, fallback)
}
