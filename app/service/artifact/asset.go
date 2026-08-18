package artifact

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

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

// DownloadAssetByPath 按制品完整路径下载资源。
// 空间级: viewName/groupPath/packageName/version/fileName
// 代码库级: repo/repoIdentifier/viewName/groupPath/packageName/version/fileName
func (s *Service) DownloadAssetByPath(ctx context.Context, spaceCode string, filepath string) (io.ReadCloser, *model.ArtifactBlob, string, error) {
	segments := splitAndDecode(filepath)
	if len(segments) < 2 {
		return nil, nil, "", pkgErrors.WithI18nKey(
			fmt.Errorf("invalid filepath: need at least viewName/fileName, got %d segments", len(segments)),
			i18n.ArtifactFilepathInvalid,
		)
	}

	space, err := s.spaceStore.GetByCode(ctx, spaceCode)
	if err != nil {
		return nil, nil, "", pkgErrors.WithI18nKey(err, i18n.ArtifactSpaceQueryError)
	}

	var repoID int64
	var libName string
	var remainingPath []string

	if segments[0] == "repo" {
		if len(segments) < 4 {
			return nil, nil, "", pkgErrors.WithI18nKey(
				fmt.Errorf("invalid repo-level filepath: too few segments"),
				i18n.ArtifactFilepathInvalid,
			)
		}
		repoIdentifier := segments[1]
		repo, err := s.repoStore.GetBySpaceAndIdentifier(ctx, space.ID, repoIdentifier)
		if err != nil {
			return nil, nil, "", pkgErrors.WithI18nKey(err, i18n.ArtifactRepoQueryError)
		}
		repoID = int64(repo.ID)
		libName = segments[2]
		remainingPath = segments[3:]
	} else {
		libName = segments[0]
		remainingPath = segments[1:]
	}

	artifactLib, err := s.artStore.Views().GetByCode(ctx, int64(space.ID), repoID, libName)
	if err != nil {
		return nil, nil, "", pkgErrors.WithI18nKey(err, i18n.ArtifactLibQueryError)
	}

	asset, err := s.resolveAssetInView(ctx, artifactLib, remainingPath)
	if err != nil {
		return nil, nil, "", pkgErrors.WithI18nKey(
			fmt.Errorf("artifact not found at path: %s", filepath),
			i18n.ArtifactAssetQueryError,
		)
	}

	return s.DownloadAsset(ctx, asset.ID)
}

func (s *Service) resolveAssetInView(ctx context.Context, view *model.ArtifactView, remainingPath []string) (*model.ArtifactAsset, error) {
	if view.Format != model.ArtifactFileFormat {
		return nil, fmt.Errorf("unsupported format for path download: %s", view.Format)
	}
	return s.resolveFileFormatAsset(ctx, view.ID, view.Format, remainingPath)
}

func (s *Service) resolveFileFormatAsset(ctx context.Context, artifactID int64, format model.ArtifactFormat, remainingPath []string) (*model.ArtifactAsset, error) {
	if len(remainingPath) == 0 {
		return nil, errors.New("empty remaining path")
	}

	fileName := remainingPath[len(remainingPath)-1]
	prefixSegments := remainingPath[:len(remainingPath)-1]

	if len(prefixSegments) == 0 {
		return nil, errors.New("empty prefix path")
	}

	for i := len(prefixSegments); i > 0; i-- {
		treePath := "/" + strings.Join(prefixSegments[:i], "/")
		node, err := s.artStore.Nodes().GetByPath(ctx, artifactID, treePath, format)
		if err != nil {
			continue
		}
		if !(node.TargetID > 0) {
			continue
		}

		switch node.Type {
		case model.ArtifactTreeNodeTypeVersion:
			return s.artStore.Assets().GetVersionAsset(ctx, fileName, node.TargetID)
		default:
			return s.artStore.Assets().GetGroupAsset(ctx, fileName, node.TargetID)
		}
	}

	return nil, pkgErrors.WithI18nKey(
		fmt.Errorf("artifact not found in tree: %s", strings.Join(prefixSegments, "/")),
		i18n.ArtifactAssetQueryError,
	)
}

func splitAndDecode(filepath string) []string {
	trimmed := strings.Trim(filepath, "/")
	parts := strings.Split(trimmed, "/")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		decoded, err := url.PathUnescape(p)
		if err != nil {
			result = append(result, p)
		} else {
			result = append(result, decoded)
		}
	}
	return result
}

// Upload 上传资产
func (s *Service) Upload(ctx context.Context, view *adapter.ViewDescriptor, format model.ArtifactFormat, req *http.Request) ([]*adapter.SavedAsset, error) {
	var uploader adapter.ArtifactPackageUploader

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

func (s *Service) GetAsset(ctx context.Context, assetID int64) (*model.ArtifactAssetInfo, error) {
	return s.getAssetNodeInfo(ctx, assetID)
}

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

func (s *Service) toActiveAssetNodeInfo(ctx context.Context, assetID int64) *model.ArtifactAssetInfo {
	info, err := s.getAssetNodeInfo(ctx, assetID)
	if err != nil || info == nil || info.Status != string(model.ArtifactStatusOK) {
		return nil
	}
	return info
}

func (s *Service) getAssetNodeInfo(ctx context.Context, assetID int64) (*model.ArtifactAssetInfo, error) {
	dbAsset, err := s.artStore.Assets().GetById(ctx, assetID)
	if err != nil {
		if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
			return notFoundNodeInfo(), nil
		}
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactAssetQueryError)
	}

	assetRes, err := s.artStore.GetAsset(ctx, assetID)
	if err != nil {
		if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
			return notFoundNodeInfo(), nil
		}
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactAssetQueryError)
	}

	if dbAsset.VersionID.Valid && dbAsset.VersionID.Int64 > 0 {
		info, err := s.getVersionNodeInfo(ctx, dbAsset.VersionID.Int64)
		if err != nil {
			return nil, err
		}
		if info.Status != string(model.ArtifactStatusOK) {
			return info, nil
		}

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
		} else {
			savedLink = info.Link
		}

		populateNodeInfoAssetFields(info, assetRes)

		if isContainer {
			info.Size = savedSize
			info.OS = savedOS
			info.Arch = savedArch
			info.Link = savedLink
			info.Images = savedImages
			info.Created = savedCreated
			info.Updated = savedUpdated
		} else {
			info.Link = savedLink
		}
		return info, nil
	}

	if dbAsset.GroupID.Valid && dbAsset.GroupID.Int64 > 0 {
		return s.getGroupAssetNodeInfo(ctx, dbAsset, assetRes)
	}

	if dbAsset.ArtifactID.Valid && dbAsset.ArtifactID.Int64 > 0 {
		return s.getStandaloneAssetNodeInfo(ctx, dbAsset, assetRes)
	}

	info := &model.ArtifactAssetInfo{
		Status: string(model.ArtifactStatusOK),
		Format: dbAsset.Format,
	}
	populateNodeInfoAssetFields(info, assetRes)
	return info, nil
}

func (s *Service) getStandaloneAssetNodeInfo(ctx context.Context, dbAsset *model.ArtifactAsset, assetRes *model.ArtifactAssetsRes) (*model.ArtifactAssetInfo, error) {
	lib, err := s.ParseSpaceLib(ctx, &model.SpaceRepoReq{ArtifactID: dbAsset.ArtifactID.Int64})
	if err != nil {
		return nil, err
	}

	info := &model.ArtifactAssetInfo{
		Status: string(model.ArtifactStatusOK),
		Format: dbAsset.Format,
		Metadata: &model.ArtifactPkgVerMetadata{
			Space: nodeInfoSpaceName(lib),
			Name:  filepath.Base(assetRes.Path),
		},
	}
	populateNodeInfoAssetFields(info, assetRes)
	info.Link = buildArtifactAssetLink(nodeInfoSpaceCode(lib), nodeInfoRepoIdentifier(lib), lib.ArtifactCode, string(dbAsset.Format), "", "", "", assetRes.Path)
	return info, nil
}
