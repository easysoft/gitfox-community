package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	artifactModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/model"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"github.com/guregu/null"
)

type ModelManager struct {
	store   artifactStore.Interface
	view    *ViewDescriptor
	creator *model.CreatorInfo

	pkgDescriptor *PackageDescriptor
	pkgModel      *model.ArtifactPackage
	versionModel  *model.ArtifactVersion

	treeSyncer *artifactModel.Syncer
}

// NewModelManager 创建模型管理器
func NewModelManager(artStore artifactStore.Interface, view *ViewDescriptor, pkg *PackageDescriptor, creator *model.CreatorInfo) *ModelManager {
	return &ModelManager{
		store:         artStore,
		view:          view,
		creator:       creator,
		pkgDescriptor: pkg,
		treeSyncer:    artifactModel.NewSyncer(artStore),
	}
}

// Sync 按资产属性选择挂载粒度：isolated 走全局路径，exclusive/normal 走版本或 group。
func (mm *ModelManager) Sync(ctx context.Context, assetDesc *AssetDescriptor) (*SavedAsset, error) {
	var asset *model.ArtifactAsset
	var updated bool
	var err error

	switch assetDesc.Attr {
	case AttrAssetIsolated:
		asset, err = mm.createIsolatedAsset(ctx, assetDesc)
		if err != nil {
			return nil, err
		}
	case AttrAssetExclusive:
		ver, e := mm.getOrCreateVersion(ctx)
		if e != nil {
			return nil, e
		}
		asset, updated, err = mm.createExclusiveAsset(ctx, ver, assetDesc)
		if err != nil {
			return nil, err
		}
		if err = mm.updateVerTime(ctx, ver, asset); err != nil {
			return nil, err
		}
	default:
		if mm.pkgDescriptor.Format == model.ArtifactGenericFormat {
			asset, updated, err = mm.createNormalAssetToGroup(ctx, assetDesc)
			if err != nil {
				return nil, err
			}
		} else {
			ver, e := mm.getOrCreateVersion(ctx)
			if e != nil {
				return nil, e
			}
			asset, updated, err = mm.createNormalAsset(ctx, ver, assetDesc)
			if err != nil {
				return nil, err
			}
			if err = mm.updateVerTime(ctx, ver, asset); err != nil {
				return nil, err
			}
		}
	}

	saved := &SavedAsset{Object: asset, IsOverride: updated, Attr: assetDesc.Attr}
	return saved, nil
}

// createIsolatedAsset 创建孤立资产
func (mm *ModelManager) createIsolatedAsset(ctx context.Context, descriptor *AssetDescriptor) (*model.ArtifactAsset, error) {
	currAsset, assetExist, err := mm.getExistAsset(func() (*model.ArtifactAsset, error) {
		return mm.store.Assets().GetPath(ctx, descriptor.Path, descriptor.Format)
	})
	if err != nil {
		return nil, err
	}
	if assetExist {
		return mm.updateAsset(ctx, currAsset, descriptor)
	}
	newAsset := &model.ArtifactAsset{}
	if err = mm.createAsset(ctx, newAsset, descriptor); err != nil {
		return nil, err
	}
	return newAsset, nil
}

// createExclusiveAsset 创建独占资产
func (mm *ModelManager) createExclusiveAsset(ctx context.Context, ver *model.ArtifactVersion, descriptor *AssetDescriptor) (*model.ArtifactAsset, bool, error) {
	// exclusive 语义要求同一 version 最终只保留当前资产，其它资产统一软删。
	currAsset, assetExist, err := mm.getExistAsset(func() (*model.ArtifactAsset, error) {
		return mm.store.Assets().GetVersionAsset(ctx, descriptor.Path, ver.ID)
	})
	if err != nil {
		return nil, false, err
	}
	if assetExist {
		if _, err = mm.updateAsset(ctx, currAsset, descriptor); err != nil {
			return nil, false, err
		}
		if err = mm.store.Assets().SoftDeleteExcludeById(ctx, ver.ID, currAsset.ID, mm.creator.UID); err != nil {
			return nil, false, err
		}
		return currAsset, true, err
	}
	newAsset := &model.ArtifactAsset{
		GroupID:    null.IntFrom(0),
		ArtifactID: null.IntFrom(0),
		VersionID:  null.IntFrom(ver.ID),
	}
	if err = mm.createAsset(ctx, newAsset, descriptor); err != nil {
		return nil, false, err
	}
	if err = mm.store.Assets().SoftDeleteExcludeById(ctx, ver.ID, newAsset.ID, mm.creator.UID); err != nil {
		return nil, false, err
	}
	return newAsset, false, nil
}

// createNormalAsset 创建普通资产
func (mm *ModelManager) createNormalAsset(ctx context.Context, ver *model.ArtifactVersion, descriptor *AssetDescriptor) (*model.ArtifactAsset, bool, error) {
	currAsset, assetExist, err := mm.getExistAsset(func() (*model.ArtifactAsset, error) {
		return mm.store.Assets().GetVersionAsset(ctx, descriptor.Path, ver.ID)
	})
	if err != nil {
		return nil, false, err
	}
	if assetExist {
		_, err = mm.updateAsset(ctx, currAsset, descriptor)
		if err != nil {
			return nil, false, err
		}
		return currAsset, true, err
	}
	newAsset := &model.ArtifactAsset{
		GroupID:    null.IntFrom(0),
		ArtifactID: null.IntFrom(0),
		VersionID:  null.IntFrom(ver.ID),
	}
	if err = mm.createAsset(ctx, newAsset, descriptor); err != nil {
		return nil, false, err
	}
	return newAsset, false, nil
}

// createNormalAssetToGroup 创建普通资产到 group
func (mm *ModelManager) createNormalAssetToGroup(ctx context.Context, descriptor *AssetDescriptor) (*model.ArtifactAsset, bool, error) {
	groupID := mm.pkgDescriptor.GroupID
	if groupID == 0 {
		return nil, false, fmt.Errorf("asset's group is empty")
	}
	currAsset, assetExist, err := mm.getExistAsset(func() (*model.ArtifactAsset, error) {
		return mm.store.Assets().GetGroupAsset(ctx, descriptor.Path, groupID)
	})
	if err != nil {
		return nil, false, err
	}
	if assetExist {
		_, err = mm.updateAsset(ctx, currAsset, descriptor)
		if err != nil {
			return nil, false, err
		}
		return currAsset, true, err
	}
	newAsset := &model.ArtifactAsset{
		VersionID:  null.IntFrom(0),
		ArtifactID: null.IntFrom(mm.view.ArtifactID),
		GroupID:    null.IntFrom(groupID),
	}
	if err = mm.createAsset(ctx, newAsset, descriptor); err != nil {
		return nil, false, err
	}
	return newAsset, false, nil
}

// getExistAsset 获取资产是否存在
func (mm *ModelManager) getExistAsset(mutateFunc func() (*model.ArtifactAsset, error)) (*model.ArtifactAsset, bool, error) {
	currAsset, err := mutateFunc()
	if err == nil {
		return currAsset, true, nil
	}
	if errors.Is(err, pkgErrors.ErrResourceNotFound) {
		return nil, false, nil
	}
	return nil, false, err
}

// updateAsset 更新资产
func (mm *ModelManager) updateAsset(ctx context.Context, current *model.ArtifactAsset, descriptor *AssetDescriptor) (*model.ArtifactAsset, error) {
	// 更新资产时会新建 blob 并回链，旧 blob 仅做软删以保留历史引用痕迹。
	now := time.Now().UnixMilli()
	if err := mm.validateAsset(ctx, current, descriptor); err != nil {
		return nil, err
	}
	newBlob, err := mm.createBlob(ctx, descriptor, now)
	if err != nil {
		return nil, err
	}
	if err = mm.linkAssetToBlob(ctx, current.ID, newBlob); err != nil {
		return nil, err
	}
	previousBlobId := current.BlobID
	current.BlobID = newBlob.ID
	if descriptor.Hash != nil {
		current.CheckSum = descriptor.Hash.String()
	}
	if err = mm.store.Assets().Update(ctx, current); err != nil {
		return nil, err
	}
	if err = mm.store.Blobs().SoftDeleteById(ctx, previousBlobId, mm.creator.UID); err != nil {
		return nil, err
	}
	return current, nil
}

// createAsset 创建资产
func (mm *ModelManager) createAsset(ctx context.Context, newObj *model.ArtifactAsset, descriptor *AssetDescriptor) error {
	now := time.Now().UnixMilli()
	newBlob, err := mm.createBlob(ctx, descriptor, now)
	if err != nil {
		return err
	}
	session, _ := req.GetAuthSession(ctx)
	createReq := &model.CreateArtifactAsset{
		VersionID:   newObj.VersionID,
		ArtifactID:  newObj.ArtifactID,
		GroupID:     newObj.GroupID,
		Format:      descriptor.Format,
		Path:        descriptor.Path,
		Kind:        descriptor.Kind,
		ContentType: descriptor.ContentType,
		BlobID:      newBlob.ID,
		CreatedBy:   session.User.Account,
		EditedBy:    session.User.Account,
	}
	if descriptor.Hash != nil {
		createReq.CheckSum = descriptor.Hash.String()
	}
	created, err := mm.store.Assets().Create(ctx, createReq)
	if err != nil {
		return err
	}
	*newObj = *created
	return mm.linkAssetToBlob(ctx, newObj.ID, newBlob)
}

// linkAssetToBlob 链接资产到 blob
func (mm *ModelManager) linkAssetToBlob(ctx context.Context, assetID int64, blob *model.ArtifactBlob) error {
	blob.AssetID = assetID
	return mm.store.Blobs().Update(ctx, blob)
}

// createBlob 创建 blob
func (mm *ModelManager) createBlob(ctx context.Context, descriptor *AssetDescriptor, createTime int64) (*model.ArtifactBlob, error) {
	createReq := &model.CreateArtifactBlob{
		StorageID: mm.view.StorageID,
		Ref:       descriptor.Ref,
		Size:      descriptor.Size,
		CreatedBy: mm.creator.UID,
	}
	if descriptor.Hash != nil {
		createReq.CheckSum = descriptor.Hash.String()
	}
	blobModel, err := mm.store.Blobs().Create(ctx, createReq)
	if err != nil {
		return nil, err
	}
	return blobModel, nil
}

// validateAsset 验证资产是否需要更新
func (mm *ModelManager) validateAsset(ctx context.Context, asset *model.ArtifactAsset, descriptor *AssetDescriptor) error {
	// 若历史 checksum 可解析且 sha256 未变化，则拒绝重复写入相同内容。
	log := logger.GetLogger(ctx)
	var h Hash
	if err := json.Unmarshal([]byte(asset.CheckSum), &h); err != nil {
		log.Errorf("unmarshal checksum failed: %v", err)
		return nil
	}
	if h.Sha256 == descriptor.Hash.Sha256 {
		return ErrStorageFileNotChanged
	}
	return nil
}

// getOrCreateVersion 获取或创建版本
func (mm *ModelManager) getOrCreateVersion(ctx context.Context) (*model.ArtifactVersion, error) {
	log := logger.GetLogger(ctx)
	if mm.versionModel != nil {
		return mm.versionModel, nil
	}
	if mm.pkgModel == nil {
		pkg, err := mm.getOrCreatePackage(ctx)
		if err != nil {
			return nil, err
		}
		mm.pkgModel = pkg
	}

	defer func() {
		if mm.pkgModel != nil && mm.versionModel != nil {
			// 版本节点在版本建好后延迟同步，避免 package/version 半成品写入树。
			_, _ = mm.treeSyncer.SyncTreeNodeForVersion(ctx, mm.versionModel, mm.pkgModel)
		}
	}()

	currVersion, err := mm.store.Versions().GetByVersion(ctx, mm.pkgModel.ID, mm.pkgDescriptor.Version)
	if err == nil {
		if currVersion.IsDeleted() {
			if e := mm.store.Versions().Restore(ctx, currVersion.ID, mm.creator.UID); e != nil {
				return nil, e
			}
		}
		mm.versionModel = currVersion
		return currVersion, nil
	}
	if !errors.Is(err, pkgErrors.ErrResourceNotFound) {
		return nil, err
	}

	verModel := &model.CreateArtifactVersion{
		PackageID: mm.pkgModel.ID,
		Version:   mm.pkgDescriptor.Version,
		CreatedBy: mm.creator.UID,
	}
	if mm.pkgDescriptor.VersionMetadata != nil {
		metadata, e := mm.pkgDescriptor.VersionMetadata.ToJSON()
		if e == nil {
			verModel.Metadata = string(metadata)
		} else {
			log.Errorf("convert version metadata to json failed: %v", err)
		}
	}
	createdVer, err := mm.store.Versions().Create(ctx, verModel)
	if err != nil {
		return nil, err
	}
	mm.versionModel = createdVer
	return createdVer, nil
}

// updateVerTime 更新版本时间
func (mm *ModelManager) updateVerTime(ctx context.Context, verModel *model.ArtifactVersion, assetModel *model.ArtifactAsset) error {
	if verModel.ID != assetModel.VersionID.Int64 {
		return nil
	}
	verModel.EditedDate = assetModel.EditedDate
	return mm.store.Versions().Update(ctx, verModel)
}

// getOrCreatePackage 获取或创建 package
func (mm *ModelManager) getOrCreatePackage(ctx context.Context) (*model.ArtifactPackage, error) {
	if mm.pkgModel != nil {
		return mm.pkgModel, nil
	}
	desc := mm.pkgDescriptor
	currPkg, err := mm.store.Packages().GetByName(ctx, desc.Name, desc.Namespace, mm.view.ArtifactID, desc.Format)
	if err == nil {
		if currPkg.IsDeleted() {
			if e := mm.store.Packages().Restore(ctx, currPkg.ID, mm.creator.UID); e != nil {
				return nil, e
			}
		}
		mm.pkgModel = currPkg
		return currPkg, nil
	}
	if !errors.Is(err, pkgErrors.ErrResourceNotFound) {
		return nil, err
	}
	pkgModel := &model.CreateArtifactPackage{
		ArtifactID: mm.view.ArtifactID,
		Name:       desc.Name,
		Namespace:  desc.Namespace,
		GroupID:    desc.GroupID, // 来自上层 Prepare 回调（container 的 getOrCreateGroup、raw 的 namespace）
		Format:     desc.Format,
		CreatedBy:  mm.creator.UID,
	}
	createdPkg, err := mm.store.Packages().Create(ctx, pkgModel)
	if err != nil {
		return nil, err
	}
	mm.pkgModel = createdPkg

	if mm.pkgModel.Namespace != "" && mm.pkgModel.GroupID == 0 {
		// raw package 的 namespace 会被拆成 group 树，并把末级 group 回填到 package。
		if e := mm.treeSyncer.SyncRawGroup(ctx, mm.pkgModel); e != nil {
			return nil, e
		}
	}
	return createdPkg, nil
}
