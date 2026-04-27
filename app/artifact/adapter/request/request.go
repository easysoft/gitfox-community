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

// ArtifactRequest 封装制品上传请求的基础设施（view、store、writer 列表）。
type ArtifactRequest struct {
	view          *adapter.ViewDescriptor
	storageId     int64
	artifactStore artifactStore.Interface
	modelMgr      *adapter.ModelManager

	writers []storagedriver.FileWriter
}

func newRequest(view *adapter.ViewDescriptor, store artifactStore.Interface) *ArtifactRequest {
	return &ArtifactRequest{
		view:          view,
		storageId:     view.StorageID,
		artifactStore: store,
		writers:       make([]storagedriver.FileWriter, 0),
	}
}

// ArtifactUploadRequest 管理一次完整上传：表单解析 → 存储写入 → 模型同步 → 提交。
type ArtifactUploadRequest struct {
	*ArtifactRequest
	creator    *model.CreatorInfo
	Descriptor *adapter.PackageDescriptor
	Assets     []*adapter.SavedAsset
}

func NewUpload(view *adapter.ViewDescriptor, store artifactStore.Interface) *ArtifactUploadRequest {
	return &ArtifactUploadRequest{
		ArtifactRequest: newRequest(view, store),
	}
}

func (r *ArtifactUploadRequest) RegisterWriter(writer storagedriver.FileWriter) {
	r.writers = append(r.writers, writer)
}

// LoadCreator 从上下文中提取当前操作人信息，用于记录 blob 创建者。
func (r *ArtifactUploadRequest) LoadCreator(ctx context.Context) *model.CreatorInfo {
	session, ok := req.GetAuthSession(ctx)
	if !ok {
		return &model.CreatorInfo{ID: 0, UID: "anonymous", DisplayName: "anonymous"}
	}
	if session.User != nil {
		return &model.CreatorInfo{
			ID:          session.User.ID,
			UID:         session.User.Account,
			DisplayName: session.User.Account,
		}
	}
	if session.Principal != nil {
		return &model.CreatorInfo{
			ID:          0,
			UID:         session.Principal.DisplayName,
			DisplayName: session.Principal.DisplayName,
		}
	}
	return &model.CreatorInfo{ID: 0, UID: "anonymous", DisplayName: "anonymous"}
}

// Commit 执行模型同步 + 提交所有 storage writer。
func (r *ArtifactUploadRequest) Commit(ctx context.Context) error {
	log := logger.GetLogger(ctx)

	modelMgr := adapter.NewModelManager(r.artifactStore, r.view, r.Descriptor, r.LoadCreator(ctx))
	if savedAsset, err := modelMgr.Sync(ctx, r.Descriptor.MainAsset); err != nil {
		if !errors.Is(err, adapter.ErrStorageFileNotChanged) {
			return err
		}
	} else {
		if savedAsset.Attr != adapter.AttrAssetIsolated {
			r.Assets = append(r.Assets, savedAsset)
		}
	}

	for _, subAsset := range r.Descriptor.SubAssets {
		if _, err := modelMgr.Sync(ctx, subAsset); err != nil {
			if !errors.Is(err, adapter.ErrStorageFileNotChanged) {
				return err
			}
			continue
		}
	}

	// commit file writer
	log.Debugf("commit file writer for artifact %s", r.Descriptor.Name)
	for _, fw := range r.writers {
		if err := fw.Commit(ctx); err != nil {
			log.Warnf("commit file upload failed: %v", err)
			return err
		}
		if err := fw.Close(); err != nil {
			log.Warnf("commit file close failed: %v", err)
		}
	}

	return nil
}

func (r *ArtifactUploadRequest) Cancel(ctx context.Context) error {
	log := logger.GetLogger(ctx)
	for _, fw := range r.writers {
		if err := fw.Cancel(ctx); err != nil {
			log.Warnf("cancel file upload failed: %v", err)
		}
	}
	return nil
}
