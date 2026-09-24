// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package helm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/easysoft/gitfox-community/app/artifact/adapter"
	"github.com/easysoft/gitfox-community/app/artifact/adapter/request"
	"github.com/easysoft/gitfox-community/app/store/artifact"
	artifactStore "github.com/easysoft/gitfox-community/app/store/artifact"

	"github.com/easysoft/gitfox-community/pkg/storage"

	model "github.com/easysoft/gitfox-community/app/model/artifact"

	"github.com/easysoft/gitfox-community/pkg/logger"
)

const (
	_chartPkgType = "application/x-compressed-tar"

	// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#rfc-1035-label-names
	_regexChartName = `^[a-z][a-z0-9\-]+[a-z0-9]$`
)

type helmUploader struct {
	uploadReq *request.ArtifactUploadRequest
	//base      *base.HostedUploader
	artStore artifact.Interface
	view     *adapter.ViewDescriptor
	store    storage.ContentStorage

	storeLayout adapter.StorageLayout

	descriptor *adapter.PackageDescriptor
	bufReader  io.Reader
}

func NewUploader(contentStore storage.ContentStorage, artStore artifactStore.Interface, view *adapter.ViewDescriptor) adapter.ArtifactPackageUploader {
	return &helmUploader{
		uploadReq: request.NewUpload(view, artStore),
		artStore:  artStore,
		view:      view,
		store:     contentStore,

		storeLayout: adapter.StorageLayoutBlob,
		descriptor:  adapter.NewEmptyPackageDescriptor(),
	}
}

func (h *helmUploader) Serve(ctx context.Context, req *http.Request) (int64, error) {
	_ = req.ParseMultipartForm(32 << 20)
	file, _, err := req.FormFile("chart")
	if err != nil {
		return 0, err
	}

	fw, ref, err := adapter.NewRandomBlobWriter(ctx, h.store)
	if err != nil {
		return 0, err
	}

	buf := bytes.NewBuffer(nil)
	h.bufReader = buf

	h.uploadReq.RegisterWriter(fw)
	size, hash, err := request.Write(file, fw, buf)

	h.descriptor.MainAsset.Size = size
	h.descriptor.MainAsset.Hash = hash
	h.descriptor.MainAsset.Ref = ref
	h.descriptor.MainAsset.ContentType = _chartPkgType
	h.descriptor.MainAsset.Rank = model.AssetRankMain
	h.descriptor.MainAsset.Attr = adapter.AttrAssetNormal
	h.descriptor.MainAsset.Format = model.ArtifactHelmFormat
	return size, err
}

func (h *helmUploader) IsValid(ctx context.Context) error {
	chart, err := ParseChart(h.bufReader)
	if err != nil {
		return fmt.Errorf("invalid package content: %s", err.Error())
	}
	logger.GetLogger(ctx).Infof("chart: %+v", chart)

	if ok, _ := regexp.MatchString(_regexChartName, chart.Name()); !ok {
		return adapter.ErrInvalidPackageName
	}

	h.descriptor.MainAsset.Path = fmt.Sprintf("charts/%s-%s.tgz", chart.Metadata.Name, chart.Metadata.Version)

	h.descriptor.Name = chart.Metadata.Name
	h.descriptor.Namespace = ""
	h.descriptor.Version = chart.Metadata.Version
	h.descriptor.Format = model.ArtifactHelmFormat
	h.descriptor.MainAsset.Metadata = &AssetMetadata{data: chart.Metadata}
	uploader := h.uploadReq.LoadCreator(ctx)
	h.descriptor.VersionMetadata = &adapter.VersionMetadata{CreatorName: uploader.UID}
	h.uploadReq.Descriptor = h.descriptor

	return nil
}

func (h *helmUploader) Save(ctx context.Context) error {
	if err := h.uploadReq.Commit(ctx); err != nil {
		_ = h.uploadReq.Cancel(ctx)
		return err
	}
	return nil
}

func (h *helmUploader) Cancel(ctx context.Context) error {
	return h.uploadReq.Cancel(ctx)
}

func (h *helmUploader) GetAssets() []*adapter.SavedAsset {
	return h.uploadReq.Assets
}
