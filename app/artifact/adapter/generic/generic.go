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

	if h.descriptor.MainAsset.Path == "" {
		h.descriptor.MainAsset.Path = fh.Filename
	}
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
	name := req.FormValue("name")
	group := req.FormValue("group")
	groupId := req.FormValue("id")

	if name != "" {
		h.descriptor.MainAsset.Path = name
	}

	if group == "" && groupId == "" {
		return fmt.Errorf("%w: required parameter 'group' or 'id'", adapter.ErrMissFormField)
	}

	viewId := h.view.ViewID
	if groupId != "" {
		gId, err := strconv.ParseInt(groupId, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: %s", adapter.ErrInvalidGroupId, groupId)
		}
		g, err := h.artStore.Groups().GetById(req.Context(), gId)
		if err != nil {
			return err
		}
		if g.ViewID != viewId {
			return fmt.Errorf("%w: %s", adapter.ErrInvalidGroupId, groupId)
		}
		h.descriptor.GroupId = gId
	} else {
		if ok, _ := regexp.MatchString(_regexPackageGroup, group); !ok {
			return adapter.ErrInvalidGroupName
		}
		g, err := h.artStore.Groups().GetEndpointByNames(req.Context(), viewId, strings.Split(group, "."), model.ArtifactGenericFormat)
		if err != nil {
			return err
		}
		h.descriptor.GroupId = g.ID
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
