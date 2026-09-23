// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package request

import (
	"context"
	"errors"

	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/artifact/adapter"
	model "github.com/easysoft/gitfox-community/app/model/artifact"

	artifactStore "github.com/easysoft/gitfox-community/app/store/artifact"
	storagedriver "github.com/easysoft/gitfox-community/pkg/storage/driver"

	"github.com/easysoft/gitfox-community/pkg/logger"
)

type ArtifactIndexRequest struct {
	*ArtifactRequest

	Descriptor *adapter.PackageMetaDescriptor
}

func NewIndex(view *adapter.ViewDescriptor, artifactStore artifactStore.Interface) *ArtifactIndexRequest {
	return &ArtifactIndexRequest{
		ArtifactRequest: newRequest(view, artifactStore),
	}
}

func (r *ArtifactIndexRequest) RegisterWriter(writer storagedriver.FileWriter) {
	r.writers = append(r.writers, writer)
}

func (r *ArtifactIndexRequest) Commit(ctx context.Context) error {
	log := logger.GetLogger(ctx)

	// AnonymousUser is only used for unittest
	var creatorId = int64(0)
	var creator string
	session, ok := req.GetAuthSession(ctx)
	if ok {
		creatorId = session.User.ID
		creator = session.User.Account
	}

	if r.Descriptor.MainAsset == nil {
		return errors.New("index main asset can't be nil")
	}

	modelMgr := adapter.NewModelManager(r.artifactStore, r.view, &adapter.PackageDescriptor{
		Name:      r.Descriptor.Name,
		Format:    r.Descriptor.Format,
		MainAsset: r.Descriptor.MainAsset,
		SubAssets: r.Descriptor.SubAssets,
	}, &model.CreatorInfo{ID: creatorId, UID: creator})

	// setup main asset model
	if _, err := modelMgr.Sync(ctx, r.Descriptor.MainAsset); err != nil {
		return err
	}

	// setup sub asset models
	for _, subAsset := range r.Descriptor.SubAssets {
		if _, err := modelMgr.Sync(ctx, subAsset); err != nil {
			return err
		}
	}

	// commit file writer
	for _, fw := range r.writers {
		if err := fw.Commit(ctx); err != nil {
			log.Warnf("commit file upload failed: %v", err)
			return err
		}
	}
	return nil
}

func (r *ArtifactIndexRequest) Cancel(ctx context.Context) error {
	log := logger.GetLogger(ctx)

	// clean files
	for _, fw := range r.writers {
		if err := fw.Cancel(ctx); err != nil {
			log.Warnf("cancel file upload failed: %v", err)
		}
	}
	return nil
}
