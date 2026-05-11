package generic

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
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
	_segmentName       = `[a-zA-Z0-9](?:[a-zA-Z0-9\-_]*[a-zA-Z0-9])?`
	_regexPackageGroup = `^` + _segmentName + `(?:\.` + _segmentName + `)*` + `$`
)

type uploader struct {
	uploadReq   *request.ArtifactUploadRequest
	store       storage.ContentStorage
	artStore    artifactStore.Interface
	view        *adapter.ViewDescriptor
	storeLayout adapter.StorageLayout
	descriptor  *adapter.PackageDescriptor
	bufReader   io.Reader
}

// NewUploader 创建 generic 格式的上传器。
func NewUploader(contentStore storage.ContentStorage, artStore artifactStore.Interface, view *adapter.ViewDescriptor) adapter.ArtifactPackageUploader {
	return &uploader{
		uploadReq:   request.NewUpload(view, artStore),
		store:       contentStore,
		artStore:    artStore,
		view:        view,
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

	// 文件名一律以上传文件本身为准（fh.Filename），不接受 form 里的 name 覆盖，
	// 以避免客户端传入 name="swagger" 等无后缀值，导致 asset.path 丢失扩展名，
	// 进而在文件树 / Content-Disposition 下载时还原不出正确的文件类型。
	h.descriptor.MainAsset.Path = fh.Filename
	h.descriptor.MainAsset.Size = size
	h.descriptor.MainAsset.Hash = hash
	h.descriptor.MainAsset.Ref = ref
	h.descriptor.MainAsset.ContentType = "application/octet-stream"
	h.descriptor.MainAsset.Kind = model.AssetKindMain
	h.descriptor.MainAsset.Attr = adapter.AttrAssetNormal

	h.descriptor.Format = model.ArtifactGenericFormat
	return size, err
}

func (h *uploader) validateForm(_ context.Context, req *http.Request) error {
	group := req.FormValue("group")
	groupId := req.FormValue("groupId")

	if group == "" && groupId == "" {
		return badRequest(
			fmt.Errorf("%w: required parameter 'group' or 'groupId'", adapter.ErrMissFormField),
			i18n.ArtifactUploadGroupRequired,
		)
	}

	artifactID := h.view.ArtifactID
	if groupId != "" {
		gId, err := strconv.ParseInt(groupId, 10, 64)
		if err != nil {
			return badRequest(
				fmt.Errorf("%w: %s", adapter.ErrInvalidGroupId, groupId),
				i18n.ArtifactUploadGroupIdInvalid,
			)
		}
		g, err := h.artStore.Groups().GetById(req.Context(), gId)
		if err != nil {
			return badRequest(
				fmt.Errorf("%w: %s: %v", adapter.ErrInvalidGroupId, groupId, err),
				i18n.ArtifactUploadGroupIdInvalid,
			)
		}
		if g.ArtifactID != artifactID {
			return badRequest(
				fmt.Errorf("%w: groupId=%s does not belong to artifactID=%d", adapter.ErrInvalidGroupId, groupId, artifactID),
				i18n.ArtifactUploadGroupArtifactMismatch,
			)
		}
		h.descriptor.GroupId = gId
	} else {
		if ok, _ := regexp.MatchString(_regexPackageGroup, group); !ok {
			return badRequest(
				adapter.ErrInvalidGroupName,
				i18n.ArtifactUploadGroupNameInvalid,
			)
		}
		g, err := h.artStore.Groups().GetEndpointByNames(req.Context(), artifactID, strings.Split(group, "."), model.ArtifactGenericFormat)
		if err != nil {
			return badRequest(
				fmt.Errorf("%w: group=%q not found under artifactID=%d: %v", adapter.ErrInvalidGroupName, group, artifactID, err),
				i18n.ArtifactUploadGroupPathNotFound,
			)
		}
		if g.ArtifactGroup == nil || g.ArtifactGroup.ArtifactID != artifactID {
			return badRequest(
				fmt.Errorf("%w: group=%q does not belong to artifactID=%d", adapter.ErrInvalidGroupName, group, artifactID),
				i18n.ArtifactUploadGroupArtifactMismatch,
			)
		}
		h.descriptor.GroupId = g.ID
	}

	return nil
}

// badRequest 将错误标记为 400 状态并附带 i18n key，供 controller 层据此选择响应状态码。
func badRequest(err error, key i18n.I18nKey) error {
	return pkgErrors.WithHttpStatus(pkgErrors.WithI18nKey(err, key), http.StatusBadRequest)
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
