// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package container

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/easysoft/gitfox-community/app/artifact/adapter"
	"github.com/easysoft/gitfox-community/app/artifact/adapter/request"
	model "github.com/easysoft/gitfox-community/app/model/artifact"
	artifactStore "github.com/easysoft/gitfox-community/app/store/artifact"
	"github.com/easysoft/gitfox-community/pkg/storage"

	"github.com/easysoft/gitfox-community/pkg/logger"
)

type manifestUploader struct {
	digest    string
	tag       string
	uploadReq *request.ArtifactUploadRequest
	//base      *base.HostedUploader
	store storage.ContentStorage

	storeLayout adapter.StorageLayout

	descriptor *adapter.PackageDescriptor
	bufReader  io.Reader
}

func NewManifestUploader(artStore artifactStore.Interface, view *adapter.ViewDescriptor, dgst string) *manifestUploader {
	up := &manifestUploader{
		digest:      dgst,
		uploadReq:   request.NewUpload(view, artStore),
		store:       view.Store,
		storeLayout: adapter.StorageLayoutBlob,
		descriptor:  adapter.NewEmptyPackageDescriptor(),
	}
	return up
}

func (h *manifestUploader) Prepare(fn func(desc *adapter.PackageDescriptor)) {
	h.descriptor.Format = model.ArtifactContainerFormat
	fn(h.descriptor)
}

func (h *manifestUploader) Descriptor() *adapter.PackageDescriptor {
	return h.descriptor
}

func (h *manifestUploader) Serve(ctx context.Context, req *http.Request) (int64, error) {
	mediaType := req.Header.Get("Content-Type")
	fw, ref, err := adapter.NewRandomBlobWriter(ctx, h.store)
	if err != nil {
		return 0, err
	}

	buf := bytes.NewBuffer(nil)
	h.bufReader = buf

	h.uploadReq.RegisterWriter(fw)
	size, hash, err := request.Write(req.Body, fw, buf)
	if err != nil {
		return 0, err
	}

	h.descriptor.MainAsset.Size = size
	h.descriptor.MainAsset.Hash = hash
	h.descriptor.MainAsset.Ref = ref
	h.descriptor.MainAsset.ContentType = mediaType
	h.descriptor.MainAsset.Rank = model.AssetRankMain
	if h.digest != "" {
		// saved manifest as a blob, shared cross spaces
		h.descriptor.MainAsset.Attr = adapter.AttrAssetIsolated
	} else {
		// manifest is a tag, should contain only one asset
		h.descriptor.MainAsset.Attr = adapter.AttrAssetExclusive

		// save metadata for container tag
		uploader := h.uploadReq.LoadCreator(ctx)
		h.descriptor.VersionMetadata = &adapter.VersionMetadata{CreatorName: uploader.UID}
	}

	h.descriptor.Format = model.ArtifactContainerFormat
	return size, err
}

func (h *manifestUploader) IsValid(ctx context.Context) error {
	logger.GetLogger(ctx).Infof("digest: %s", h.digest)
	logger.GetLogger(ctx).Infof("hash sha256: %s", h.descriptor.MainAsset.Hash.Sha256)
	h.descriptor.MainAsset.Path = fmt.Sprintf("sha256:%s", h.descriptor.MainAsset.Hash.Sha256)
	h.descriptor.MainAsset.Format = model.ArtifactContainerFormat
	h.uploadReq.Descriptor = h.descriptor
	return nil
}

func (h *manifestUploader) Save(ctx context.Context) error {
	if err := h.uploadReq.Commit(ctx); err != nil {
		logger.GetLogger(ctx).Errorf("【container push】commit manifest failed: err=%v", err)
		_ = h.uploadReq.Cancel(ctx)
		return err
	}
	return nil
}

func (h *manifestUploader) Cancel(ctx context.Context) error {
	return h.uploadReq.Cancel(ctx)
}

func (h *manifestUploader) GetAssets() []*adapter.SavedAsset {
	return h.uploadReq.Assets
}
