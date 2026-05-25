package artifact

import (
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

// filterVersionList 按请求条件过滤 package 下的版本列表。
func filterVersionList(list []*model.VersionListItem, req *model.ListAssetReq) []*model.VersionListItem {
	if req.CreatedBy == "" && req.EditedBy == "" && req.Version == "" {
		return list
	}
	filtered := make([]*model.VersionListItem, 0, len(list))
	for _, item := range list {
		if req.CreatedBy != "" && item.CreatedBy != req.CreatedBy {
			continue
		}
		if req.EditedBy != "" && item.EditedBy != req.EditedBy {
			continue
		}
		if req.Version != "" && !strings.Contains(item.Version, req.Version) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

// sortVersionList 按前端传入字段排序版本列表，默认按 editedDate 倒序。
func sortVersionList(list []*model.VersionListItem, sortField, order string) {
	if sortField == "" {
		sortField = "editedDate"
	}
	ascending := listSortAscending(sortField, order)
	sort.Slice(list, func(i, j int) bool {
		var cmp int
		switch sortField {
		case "createdDate":
			cmp = compareTime(list[i].CreatedDate, list[j].CreatedDate)
		case "editedDate":
			cmp = compareTime(list[i].EditedDate, list[j].EditedDate)
		case "name", "version":
			cmp = strings.Compare(strings.ToLower(list[i].Version), strings.ToLower(list[j].Version))
		case "size":
			cmp = compareInt64(list[i].Size, list[j].Size)
		case "id":
			cmp = compareInt64(list[i].ID, list[j].ID)
		default:
			cmp = compareTime(list[i].EditedDate, list[j].EditedDate)
		}
		if cmp == 0 {
			cmp = compareInt64(list[i].ID, list[j].ID)
		}
		if ascending {
			return cmp < 0
		}
		return cmp > 0
	})
}

// filterAssetNodeInfoList 按名称、平台和操作人过滤资产详情列表。
func filterAssetNodeInfoList(list []*model.ArtifactAssetInfo, req *model.ListAssetReq) []*model.ArtifactAssetInfo {
	filtered := make([]*model.ArtifactAssetInfo, 0, len(list))
	for _, info := range list {
		if req.Name != "" && !strings.Contains(strings.ToLower(info.Path), strings.ToLower(req.Name)) {
			continue
		}
		if req.OS != "" && info.OS != req.OS {
			continue
		}
		if req.Arch != "" && info.Arch != req.Arch {
			continue
		}
		if req.CreatedBy != "" && info.CreatorName != req.CreatedBy {
			continue
		}
		if req.EditedBy != "" && info.EditedBy != req.EditedBy {
			continue
		}
		filtered = append(filtered, info)
	}
	return filtered
}

// sortAssetNodeInfoList 按资产字段排序，默认按 editedDate 倒序。
func sortAssetNodeInfoList(list []*model.ArtifactAssetInfo, sortField, order string) {
	if sortField == "" {
		sortField = "editedDate"
	}
	ascending := listSortAscending(sortField, order)
	sort.Slice(list, func(i, j int) bool {
		var cmp int
		switch sortField {
		case "createdDate":
			cmp = compareTime(list[i].Created, list[j].Created)
		case "editedDate":
			cmp = compareTime(list[i].Updated, list[j].Updated)
		case "name":
			cmp = strings.Compare(assetSortName(list[i]), assetSortName(list[j]))
		case "size":
			cmp = compareInt64(list[i].Size, list[j].Size)
		default:
			cmp = strings.Compare(list[i].Path, list[j].Path)
		}
		if cmp == 0 {
			cmp = strings.Compare(list[i].Path, list[j].Path)
		}
		if ascending {
			return cmp < 0
		}
		return cmp > 0
	})
}

// listSortAscending 返回列表排序方向；未传 order 时名称类升序、时间类倒序。
func listSortAscending(sortField, order string) bool {
	switch strings.ToLower(order) {
	case "asc":
		return true
	case "desc":
		return false
	}
	return sortField == "id" || sortField == "name" || sortField == "version"
}

// compareTime 比较可空时间，nil 按零值时间处理。
func compareTime(left, right *datetime.Local) int {
	lt, rt := timeVal(left), timeVal(right)
	switch {
	case lt.Before(rt):
		return -1
	case lt.After(rt):
		return 1
	default:
		return 0
	}
}

// compareInt64 比较 int64，便于排序 tie-break。
func compareInt64(left, right int64) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

// assetSortName 提取资产排序名称，优先使用 nodeInfo 元数据名称。
func assetSortName(info *model.ArtifactAssetInfo) string {
	if info == nil {
		return ""
	}
	if info.Metadata != nil && info.Metadata.Name != "" {
		return strings.ToLower(info.Metadata.Name)
	}
	if info.Path != "" {
		return strings.ToLower(filepath.Base(info.Path))
	}
	return ""
}

// timeVal 将项目本地时间类型转为标准 time.Time。
func timeVal(t *datetime.Local) time.Time {
	if t == nil {
		return time.Time{}
	}
	return time.Time(*t)
}

// addLinkForGeneric 为 generic 分组资产补齐下载链接和实体 ID。
func addLinkForGeneric(data []*model.ArtifactAssetsRes, spaceName, groupPath, format string) error {
	if format == "" {
		return nil
	}
	if spaceName == "" {
		spaceName = "global"
	}
	for _, res := range data {
		segments := []string{"_artifacts", spaceName, format}
		if groupPath != "" {
			segments = append(segments, strings.Split(strings.Trim(groupPath, "/"), "/")...)
		}
		if res.Path != "" {
			segments = append(segments, strings.Split(strings.Trim(res.Path, "/"), "/")...)
		}
		link, err := url.JoinPath("/", segments...)
		if err != nil {
			return err
		}
		res.Link = link
		res.EntityID = (&model.ArtifactTreeNodeId{Type: model.ArtifactTreeNodeTypeAsset, Pk: res.Id}).String()
	}
	return nil
}

// applyNodeInfoToAsset 将 nodeInfo 解析出的业务字段回填到资产列表项。
func applyNodeInfoToAsset(a *model.ArtifactAssetsRes, info *model.ArtifactAssetInfo) {
	if a == nil || info == nil || info.Status != string(model.ArtifactStatusOK) {
		return
	}
	a.EntityID = (&model.ArtifactTreeNodeId{Type: model.ArtifactTreeNodeTypeAsset, Pk: a.Id}).String()
	a.Link = info.Link
	a.OS = info.OS
	a.Arch = info.Arch
	if info.Size > 0 {
		a.Size = info.Size
	}
	if info.Created != nil {
		a.Created = info.Created
	}
	if info.Updated != nil {
		a.Updated = info.Updated
	}
	if info.Metadata != nil {
		a.Name = info.Metadata.Name
		a.PackageName = info.Metadata.Name
		a.Group = info.Metadata.Group
		a.Version = info.Metadata.Version
	}
}

// parseNodeInfoEntityID 解析 nodeInfo 入参中的 entityID。
func parseNodeInfoEntityID(entityID string) (model.ArtifactEntityKind, int64, error) {
	if strings.Count(entityID, ".") != 1 {
		return "", 0, fmt.Errorf("invalid artifact entity id: %s", entityID)
	}

	parts := strings.Split(entityID, ".")
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid artifact entity id number: %s", entityID)
	}
	return model.ArtifactEntityKind(parts[0]), id, nil
}

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
	baseURL := *registryURL
	// 循环剥掉尾部 /v2：uiURL 可能本身就配了 /v2，叠加 GenerateRegistryURL 的 JoinPath("v2") 后 path 会变成 /v2/v2。
	baseURL.Path = trimTrailingV2(baseURL.Path)
	baseURL.RawQuery = ""
	baseURL.Fragment = ""
	if baseURL.Host == "" {
		return strings.Trim(baseURL.Path, "/")
	}

	ref := baseURL.Host
	if p := strings.Trim(baseURL.EscapedPath(), "/"); p != "" {
		ref += "/" + p
	}
	return ref
}

// appendSlashSegments 将斜杠路径拆成 URL path 段。
func appendSlashSegments(dst []string, p string) []string {
	trimmed := strings.Trim(p, "/")
	if trimmed == "" {
		return dst
	}
	return append(dst, strings.Split(trimmed, "/")...)
}

// trimTrailingV2 去掉路径尾部所有 "/v2" 后缀。
// uiURL 若本身已带 /v2，GenerateRegistryURL JoinPath("v2") 会再叠一层，需要循环剥净。
func trimTrailingV2(p string) string {
	for {
		p = strings.TrimRight(p, "/")
		next := strings.TrimSuffix(p, "/v2")
		if next == p {
			break
		}
		p = next
	}
	return p
}

// nodeInfoSpaceName 取 nodeInfo 响应中的空间名。
func nodeInfoSpaceName(view *adapter.ViewDescriptor) string {
	if view != nil && view.Space != nil {
		return view.Space.Name
	}
	return ""
}

// invalidNodeInfo 返回 entityID 无效的 nodeInfo 响应。
func invalidNodeInfo() *model.ArtifactAssetInfo {
	return &model.ArtifactAssetInfo{Status: string(model.ArtifactStatusInvalidId)}
}

// notFoundNodeInfo 返回资源不存在的 nodeInfo 响应。
func notFoundNodeInfo() *model.ArtifactAssetInfo {
	return &model.ArtifactAssetInfo{Status: string(model.ArtifactStatusNotFound)}
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

// convertTreeNodesToRes 批量转换树节点为接口响应。
func convertTreeNodesToRes(inputs []*model.ArtifactTreeNode, level string) []*model.ArtifactTreeRes {
	out := make([]*model.ArtifactTreeRes, 0, len(inputs))
	for _, n := range inputs {
		out = append(out, convertTreeNodeToRes(n, level))
	}
	return out
}

// convertTreeNodeToRes 将存储层树节点转换为接口返回模型。
//
// metadata.id 始终为 ops_artifact_tree_nodes.id（树节点主键）；
// asset 节点不写入 tree_nodes 表，由调用方决定是否输出（默认留空）。
// metadata.entityID 是 "kind.pk" 实体引用串，供跨接口定位实体使用。
func convertTreeNodeToRes(n *model.ArtifactTreeNode, level string) *model.ArtifactTreeRes {
	item := &model.ArtifactTreeRes{Name: n.Name, Path: n.Path, Format: n.Format, Leaf: n.Type == model.ArtifactTreeNodeTypeAsset}
	if n.Type == model.ArtifactTreeNodeTypeVersion {
		item.Leaf = level != model.ArtifactNodeLevelAsset || n.Format == model.ArtifactContainerFormat
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
