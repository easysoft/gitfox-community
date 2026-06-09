package artifact

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter/generic"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter/raw"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

// DownloadAsset 下载资产
func (s *Service) DownloadAsset(ctx context.Context, assetID int64) (io.ReadCloser, *model.ArtifactBlob, string, error) {
	// 下载走 asset -> blob -> content store 三段定位，避免直接信任外部传入路径。
	asset, err := s.artStore.Assets().GetById(ctx, assetID)
	if err != nil {
		return nil, nil, "", pkgErrors.WithI18nKey(err, i18n.ArtifactAssetQueryError)
	}
	blob, err := s.artStore.Blobs().GetById(ctx, asset.BlobID)
	if err != nil {
		return nil, nil, "", pkgErrors.WithI18nKey(err, i18n.ArtifactBlobQueryError)
	}
	blobPath := adapter.BlobPath(blob.Ref)
	reader, err := s.fileStore.Open(ctx, blobPath)
	if err != nil {
		return nil, nil, "", pkgErrors.WithI18nKey(err, i18n.ArtifactDownloadOpenError)
	}
	return reader, blob, asset.Path, nil
}

// Upload 上传资产
func (s *Service) Upload(ctx context.Context, view *adapter.ViewDescriptor, format model.ArtifactFormat, req *http.Request) ([]*adapter.SavedAsset, error) {
	var uploader adapter.ArtifactPackageUploader

	// 当前标准上传入口只处理 raw/generic；版本化协议由各自协议控制器处理。
	switch format {
	case model.ArtifactRawFormat:
		uploader = raw.NewUploader(s.fileStore, s.artStore, view)
	case model.ArtifactFileFormat:
		uploader = generic.NewUploader(s.fileStore, s.artStore, view)
	default:
		return nil, pkgErrors.WithI18nKey(
			errors.New("unsupported artifact format"),
			i18n.ArtifactUnsupportedFormat,
		)
	}

	if _, err := uploader.Serve(ctx, req); err != nil {
		return nil, wrapUploadError(err, i18n.ArtifactUploadServeError)
	}
	if err := uploader.IsValid(ctx); err != nil {
		_ = uploader.Cancel(ctx)
		return nil, wrapUploadError(err, i18n.ArtifactUploadValidateError)
	}
	if err := uploader.Save(ctx); err != nil {
		return nil, wrapUploadError(err, i18n.ArtifactUploadSaveError)
	}
	return uploader.GetAssets(), nil
}

// GetAsset 返回单个资产的 nodeInfo 详情。
func (s *Service) GetAsset(ctx context.Context, assetID int64) (*model.ArtifactAssetInfo, error) {
	return s.getAssetNodeInfo(ctx, assetID)
}

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

// getAssetNodeInfo 按 assetID 查询节点详情，自动关联版本 / 分组信息并拼装 ArtifactNodeInfo。
func (s *Service) getAssetNodeInfo(ctx context.Context, assetID int64) (*model.ArtifactAssetInfo, error) {
	// 查 asset 元信息（format / versionID / groupID / artifactID）
	dbAsset, err := s.artStore.Assets().GetById(ctx, assetID)
	if err != nil {
		if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
			return notFoundNodeInfo(), nil
		}
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactAssetQueryError)
	}

	// 查 asset 详情（含 blob 大小、path、contentType、checksum、创建人、时间）
	// path      → 存储路径 / container digest
	// size      → 大小（字节）
	// checksum  → md5 / sha1 / sha256 / sha512 校验值
	// creatorName → 创建人 UID
	// created / updated → 创建 / 更新时间（毫秒时间戳）
	assetRes, err := s.artStore.GetAsset(ctx, assetID)
	if err != nil {
		if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
			return notFoundNodeInfo(), nil
		}
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactAssetQueryError)
	}

	if dbAsset.VersionID.Valid && dbAsset.VersionID.Int64 > 0 {
		// asset 挂在 version 下 → 先拿 version 级信息（name 名称、version 版本、group 分组、space 所属空间）
		// container 格式在 getVersionNodeInfo 内部还会调 fillContainerNodeInfo，
		// 把 manifest 解析出的 layer 累计 size / os / arch / created / updated 写入 info。
		info, err := s.getVersionNodeInfo(ctx, dbAsset.VersionID.Int64)
		if err != nil {
			return nil, err
		}
		if info.Status != string(model.ArtifactStatusOK) {
			return info, nil
		}

		// container 格式：保存 fillContainerNodeInfo 写入的字段，防止被 populateNodeInfoAssetFields 覆盖。
		// populateNodeInfoAssetFields 会用 manifest blob 的元信息（asset 表 size/createdDate）
		// 覆盖 info.Size/Created/Updated，但 container 应保留 layer 累计大小和 manifest 时间。
		isContainer := info.Format == model.ArtifactContainerFormat
		var savedSize int64
		var savedOS, savedArch, savedLink string
		var savedImages []*model.ArtifactContainerImage
		var savedCreated, savedUpdated *datetime.Local
		if isContainer {
			savedSize = info.Size
			savedOS = info.OS
			savedArch = info.Arch
			savedLink = info.Link
			savedImages = info.Images
			savedCreated = info.Created
			savedUpdated = info.Updated
		}

		// 用 asset 数据填充 path / contentType / creatorName / link / checksum
		populateNodeInfoAssetFields(info, assetRes)

		if isContainer {
			// 恢复 manifest 解析结果：layer 累计大小、操作系统、CPU 架构、时间
			info.Size = savedSize
			info.OS = savedOS
			info.Arch = savedArch
			info.Link = savedLink
			info.Images = savedImages
			info.Created = savedCreated
			info.Updated = savedUpdated
		}

		if info.Metadata != nil && !isContainer {
			// link 下载链接（含 space / format / group / name / version / path 拼成的访问路径）
			info.Link = buildArtifactAssetLink(info.Metadata.Space, string(info.Format), info.Metadata.Group, info.Metadata.Name, info.Metadata.Version, assetRes.Path)
		}
		return info, nil
	}

	if dbAsset.GroupID.Valid && dbAsset.GroupID.Int64 > 0 {
		// asset 直挂 group（generic / raw 的目录级文件），name 取文件 basename
		return s.getGroupAssetNodeInfo(ctx, dbAsset, assetRes)
	}

	if dbAsset.ArtifactID.Valid && dbAsset.ArtifactID.Int64 > 0 {
		// asset 直挂 view（standalone，无包无版本归属）
		return s.getStandaloneAssetNodeInfo(ctx, dbAsset, assetRes)
	}

	// 兜底：没有归属信息，只填基础字段
	info := &model.ArtifactAssetInfo{
		Status: string(model.ArtifactStatusOK),
		Format: dbAsset.Format,
	}
	populateNodeInfoAssetFields(info, assetRes)
	return info, nil
}

// getStandaloneAssetNodeInfo 获取孤立资产节点信息
func (s *Service) getStandaloneAssetNodeInfo(ctx context.Context, dbAsset *model.ArtifactAsset, assetRes *model.ArtifactAssetsRes) (*model.ArtifactAssetInfo, error) {
	// standalone asset 不属于 package/version/group，只按 view 级路径返回。
	view, err := s.ParseSpaceView(ctx, &model.SpaceRepoReq{ArtifactID: dbAsset.ArtifactID.Int64})
	if err != nil {
		return nil, err
	}

	info := &model.ArtifactAssetInfo{
		Status: string(model.ArtifactStatusOK),
		Format: dbAsset.Format,
		Metadata: &model.ArtifactPkgVerMetadata{
			Space: nodeInfoSpaceName(view),
			Name:  filepath.Base(assetRes.Path),
		},
	}
	populateNodeInfoAssetFields(info, assetRes)
	info.Link = buildArtifactAssetLink(nodeInfoSpaceName(view), string(dbAsset.Format), "", "", "", assetRes.Path)
	return info, nil
}
