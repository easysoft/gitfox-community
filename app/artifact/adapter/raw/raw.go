package raw

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter/request"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage"
)

const (
	_segmentName         = `[a-zA-Z0-9](?:[a-zA-Z0-9\-_]*[a-zA-Z0-9])?`
	_regexPackageName    = `^` + _segmentName + `$`
	_regexPackageVersion = `^[a-zA-Z0-9](?:[a-zA-Z0-9\-_\.]*[a-zA-Z0-9])?$`
	_regexPackageGroup   = `^` + _segmentName + `(?:\.` + _segmentName + `)*` + `$`
)

type uploader struct {
	uploadReq   *request.ArtifactUploadRequest
	store       storage.ContentStorage
	storeLayout adapter.StorageLayout
	descriptor  *adapter.PackageDescriptor
	bufReader   io.Reader
}

// NewUploader 创建 raw 格式的上传器。
func NewUploader(contentStore storage.ContentStorage, artStore artifactStore.Interface, view *adapter.ViewDescriptor) adapter.ArtifactPackageUploader {
	return &uploader{
		uploadReq:   request.NewUpload(view, artStore),
		store:       contentStore,
		storeLayout: adapter.StorageLayoutBlob,
		descriptor:  adapter.NewEmptyPackageDescriptor(),
	}
}

func (h *uploader) Serve(ctx context.Context, req *http.Request) (int64, error) {
	_ = req.ParseMultipartForm(32 << 20)
	if err := h.validateForm(ctx, req); err != nil {
		return 0, err
	}

	file, fh, err := req.FormFile("file")
	if err != nil {
		return 0, err
	}

	fw, ref, err := adapter.NewRandomBlobWriter(ctx, h.store)
	if err != nil {
		return 0, err
	}

	h.uploadReq.RegisterWriter(fw)
	size, hash, err := request.Write(file, fw)

	// raw 的文件路径以 form 字段 `name` 为准，兼容场景下为空则回退到上传文件名。
	fileName := strings.TrimSpace(req.FormValue("name"))
	if fileName == "" {
		fileName = fh.Filename
	}
	h.descriptor.MainAsset.Path = fileName
	h.descriptor.MainAsset.Size = size
	h.descriptor.MainAsset.Hash = hash
	h.descriptor.MainAsset.Ref = ref
	h.descriptor.MainAsset.ContentType = "application/octet-stream"
	h.descriptor.MainAsset.Kind = model.AssetKindMain
	h.descriptor.MainAsset.Attr = adapter.AttrAssetNormal
	h.descriptor.MainAsset.Format = model.ArtifactRawFormat

	h.descriptor.Format = model.ArtifactRawFormat
	return size, err
}

func (h *uploader) validateForm(_ context.Context, req *http.Request) error {
	packageName := strings.TrimSpace(req.FormValue("packageName"))
	namespace := req.FormValue("group")
	version := req.FormValue("version")

	if packageName == "" {
		return badRequest(
			fmt.Errorf("%w: required field: packageName", adapter.ErrMissFormField),
			i18n.ArtifactUploadNameRequired,
		)
	}
	if ok, _ := regexp.MatchString(_regexPackageName, packageName); !ok {
		return badRequest(
			fmt.Errorf("%w: require regex pattern: %s", adapter.ErrInvalidPackageName, _regexPackageName),
			i18n.ArtifactUploadNameInvalid,
		)
	}
	h.descriptor.Name = packageName

	if version == "" {
		return badRequest(
			fmt.Errorf("%w: required field: version", adapter.ErrMissFormField),
			i18n.ArtifactUploadVersionInvalid,
		)
	}
	if ok, _ := regexp.MatchString(_regexPackageVersion, version); !ok {
		return badRequest(
			fmt.Errorf("%w: require regex pattern: %s", adapter.ErrInvalidPackageVersion, _regexPackageVersion),
			i18n.ArtifactUploadVersionInvalid,
		)
	}
	h.descriptor.Version = version

	if namespace != "" {
		if ok, _ := regexp.MatchString(_regexPackageGroup, namespace); !ok {
			return badRequest(
				fmt.Errorf("%w: require regex pattern: %s", adapter.ErrInvalidGroupName, _regexPackageGroup),
				i18n.ArtifactUploadRawGroupInvalid,
			)
		}
		h.descriptor.Namespace = namespace
	}
	return nil
}

func (h *uploader) IsValid(ctx context.Context) error {
	u := h.uploadReq.LoadCreator(ctx)
	h.descriptor.VersionMetadata = &adapter.VersionMetadata{CreatorName: u.UID}
	h.uploadReq.Descriptor = h.descriptor
	return nil
}

func (h *uploader) Save(ctx context.Context) error {
	if err := h.uploadReq.Commit(ctx); err != nil {
		_ = h.uploadReq.Cancel(ctx)
		return err
	}
	return nil
}

func (h *uploader) Cancel(ctx context.Context) error {
	return h.uploadReq.Cancel(ctx)
}

func (h *uploader) GetAssets() []*adapter.SavedAsset {
	return h.uploadReq.Assets
}

// badRequest 将错误标记为 400 并挂上 i18n key，供 controller 层据此渲染前端提示。
func badRequest(err error, key i18n.I18nKey) error {
	return pkgErrors.WithHttpStatus(pkgErrors.WithI18nKey(err, key), http.StatusBadRequest)
}
