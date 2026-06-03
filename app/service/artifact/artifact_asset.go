package artifact

import (
	"context"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

// ListAssetPage 按 entityID（group.{id} / version.{id} / package.{id} / asset.{id}）分页查询。
//   - group.{id}：返回该 group 下的 asset 列表（generic 场景）
//   - version.{id}：返回该 version 下的 asset 列表（raw 场景）
//   - package.{id}：返回该 package 下的 version 列表（docker/helm 场景）
//   - asset.{id}：按 asset.id 精准定位单条资产（filter/sort/page 仍生效）
//
// 分页/排序统一在 store 层完成（参见 *database.PageResult），service 只负责
// 把命中的 asset 翻译成 ArtifactAssetInfo。OS/Arch 等运行时字段不在数据库
// 列上，强行参与排序/过滤会导致 total 失真，故此处不再支持。
func (s *Service) ListAssetPage(ctx context.Context, req *model.ListAssetReq) (*database.PageResult[*model.ArtifactAssetInfo], error) {
	splitEntityID := req.EntityID.GetEntityID()

	if splitEntityID.Kind != constant.ArtifactEntityKindGroup &&
		splitEntityID.Kind != constant.ArtifactEntityKindVersion &&
		splitEntityID.Kind != constant.ArtifactEntityKindPackage &&
		splitEntityID.Kind != constant.ArtifactEntityKindAsset {
		return nil, pkgErrors.WithI18nKey(
			fmt.Errorf("entityID must be group.{id}, version.{id}, package.{id} or asset.{id}, got %s", req.EntityID),
			i18n.ArtifactEntityIdInvalid,
		)
	}

	pageResult, err := s.artStore.ListAssetsPage(ctx, req)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactAssetListError)
	}

	results := make([]*model.ArtifactAssetInfo, 0, len(pageResult.Data))
	for _, a := range pageResult.Data {
		if info := s.toActiveAssetNodeInfo(ctx, a.Id); info != nil {
			results = append(results, info)
		}
	}

	return &database.PageResult[*model.ArtifactAssetInfo]{
		Data:      results,
		Total:     pageResult.Total,
		Page:      pageResult.Page,
		PageSize:  pageResult.PageSize,
		TotalPage: pageResult.TotalPage,
	}, nil
}

// toActiveAssetNodeInfo 取 assetID 对应的 ArtifactAssetInfo，仅返回 status=OK 的活节点；其它情况一律返回 nil。
func (s *Service) toActiveAssetNodeInfo(ctx context.Context, assetID int64) *model.ArtifactAssetInfo {
	info, err := s.getAssetNodeInfo(ctx, assetID)
	if err != nil || info == nil || info.Status != string(model.ArtifactStatusOK) {
		return nil
	}
	return info
}
