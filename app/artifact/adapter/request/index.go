// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package request

import (
	"context"
	"errors"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"

	artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
	storagedriver "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage/driver"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
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
	session, ok := req.GetAuthSession(ctx)
	if ok {
		creatorId = session.User.ID
	}

	if r.Descriptor.MainAsset == nil {
		return errors.New("index main asset can't be nil")
	}

	modelMgr := adapter.NewModelManager(r.artifactStore, r.view, &adapter.PackageDescriptor{
		Name:      r.Descriptor.Name,
		Format:    r.Descriptor.Format,
		MainAsset: r.Descriptor.MainAsset,
		SubAssets: r.Descriptor.SubAssets,
	}, &model.CreatorInfo{ID: creatorId})

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
