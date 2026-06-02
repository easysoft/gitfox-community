package artifact

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

// ListAssetPage 按 entityID（group.{id} / version.{id} / package.{id} / asset.{id}）分页查询。
//   - group.{id}：返回该 group 下的 asset 列表（generic 场景）
//   - version.{id}：返回该 version 下的 asset 列表（raw 场景）
//   - package.{id}：返回该 package 下的 version 列表（docker/helm 场景）
//   - asset.{id}：按 asset.id 精准定位单条资产（filter/sort/page 仍生效）
//
// 先全量查出再内存过滤 + 排序 + 分页。
func (s *Service) ListAssetPage(ctx context.Context, req *model.ListAssetReq) (interface{}, int64, error) {
	splitEntityID := req.EntityID.GetEntityID()

	if splitEntityID.Kind != constant.ArtifactEntityKindGroup &&
		splitEntityID.Kind != constant.ArtifactEntityKindVersion &&
		splitEntityID.Kind != constant.ArtifactEntityKindPackage &&
		splitEntityID.Kind != constant.ArtifactEntityKindAsset {
		return nil, 0, pkgErrors.WithI18nKey(
			fmt.Errorf("entityID must be group.{id}, version.{id}, package.{id} or asset.{id}, got %s", req.EntityID),
			i18n.ArtifactEntityIdInvalid,
		)
	}

	page, pageSize := req.Page, req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	if splitEntityID.Kind == constant.ArtifactEntityKindPackage {
		// package.{id} → 查所有 version → 各 version 下的 asset 平铺 → 分页
		return s.listPackageAssetsPage(ctx, splitEntityID.ID, req, page, pageSize)
	}
	if splitEntityID.Kind == constant.ArtifactEntityKindAsset {
		return s.listAssetByIDPage(ctx, splitEntityID.ID, req, page, pageSize)
	}
	return s.listAssetPageInner(ctx, splitEntityID, req, page, pageSize)
}

// listPackageAssetsPage 按 package.{id} 列出该 package 下所有 version 的 asset，
// 流程：查 version → 各 version 拉 asset → 平铺 → 复用 filter / sort / 分页骨架。
func (s *Service) listPackageAssetsPage(ctx context.Context, pkgID int64, req *model.ListAssetReq, page, pageSize int) (interface{}, int64, error) {
	versions, err := s.artStore.Versions().Find(ctx, model.SearchVersionOption{PackageID: pkgID})
	if err != nil {
		return nil, 0, pkgErrors.WithI18nKey(err, i18n.ArtifactVersionListError)
	}

	results := make([]*model.ArtifactAssetInfo, 0)
	for _, ver := range versions {
		assets, aErr := s.artStore.ListAssets(ctx, ver.ID)
		if aErr != nil {
			return nil, 0, pkgErrors.WithI18nKey(aErr, i18n.ArtifactAssetListError)
		}
		for _, a := range assets {
			if info := s.toActiveAssetNodeInfo(ctx, a.Id); info != nil {
				results = append(results, info)
			}
		}
	}

	return paginateAssetNodeInfoList(results, req, page, pageSize)
}

// listAssetByIDPage 按 asset.{id} 精准查询单条资产，沿用同一过滤 / 排序 / 分页骨架。
// asset 不存在或非 OK 状态时返回空列表，避免抛错破坏前端列表渲染。
func (s *Service) listAssetByIDPage(ctx context.Context, assetID int64, req *model.ListAssetReq, page, pageSize int) (interface{}, int64, error) {
	results := make([]*model.ArtifactAssetInfo, 0, 1)
	if info := s.toActiveAssetNodeInfo(ctx, assetID); info != nil {
		results = append(results, info)
	}
	return paginateAssetNodeInfoList(results, req, page, pageSize)
}

// listAssetPageInner 处理 group / version 的 asset 分页。
func (s *Service) listAssetPageInner(ctx context.Context, splitEntityID model.SplitEntityID, req *model.ListAssetReq, page, pageSize int) (interface{}, int64, error) {
	var assets []*model.ArtifactAssetsRes
	var err error
	switch splitEntityID.Kind {
	case constant.ArtifactEntityKindGroup:
		assets, err = s.listGroupAssetsForPage(ctx, splitEntityID.ID, req.More)
	case constant.ArtifactEntityKindVersion:
		assets, err = s.artStore.ListAssets(ctx, splitEntityID.ID)
	}
	if err != nil {
		return nil, 0, pkgErrors.WithI18nKey(err, i18n.ArtifactAssetListError)
	}

	results := make([]*model.ArtifactAssetInfo, 0, len(assets))
	for _, a := range assets {
		if info := s.toActiveAssetNodeInfo(ctx, a.Id); info != nil {
			results = append(results, info)
		}
	}

	return paginateAssetNodeInfoList(results, req, page, pageSize)
}

// paginateAssetNodeInfoList 公共尾段：filter → sort → 内存切片分页，供 group / version / package / asset 四种入口共用。
func paginateAssetNodeInfoList(results []*model.ArtifactAssetInfo, req *model.ListAssetReq, page, pageSize int) (interface{}, int64, error) {
	results = filterAssetNodeInfoList(results, req)
	sortAssetNodeInfoList(results, req.Sort, req.Order)
	return paginateSlice(results, page, pageSize)
}

// toActiveAssetNodeInfo 取 assetID 对应的 ArtifactAssetInfo，仅返回 status=OK 的活节点；其它情况一律返回 nil。
func (s *Service) toActiveAssetNodeInfo(ctx context.Context, assetID int64) *model.ArtifactAssetInfo {
	info, err := s.getAssetNodeInfo(ctx, assetID)
	if err != nil || info == nil || info.Status != string(model.ArtifactStatusOK) {
		return nil
	}
	return info
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

// paginateSlice 通用内存切片分页：start/end 越界保护，越界时返回空切片 + 真实 total。
func paginateSlice[T any](items []T, page, pageSize int) ([]T, int64, error) {
	total := int64(len(items))
	start := (page - 1) * pageSize
	if start >= int(total) {
		return []T{}, total, nil
	}
	end := start + pageSize
	if end > int(total) {
		end = int(total)
	}
	return items[start:end], total, nil
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
	lt, rt := left.Time(), right.Time()
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
