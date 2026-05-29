package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	artifactModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/model"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

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
		if mm.pkgDescriptor.Format == model.ArtifactFileFormat {
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
//
// raw 上传语义（用户面）：同 version 下若已存在同名资产：
//  1. md5/sha1/sha256 三摘要全等 → 视为重复，返回 ErrAssetDuplicate（带 i18n key），上层据此渲染 400；
//  2. 任一摘要不同 → 自动追加 `(n)` 后缀（n=1 起递增至空位）保留两份，descriptor.Path 同步改写。
//
// 不再走旧的"覆盖更新"分支，避免老 blob 被悄悄替换、用户失去原文件。
func (mm *ModelManager) createNormalAsset(ctx context.Context, ver *model.ArtifactVersion, descriptor *AssetDescriptor) (*model.ArtifactAsset, bool, error) {
	currAsset, assetExist, err := mm.getExistAsset(func() (*model.ArtifactAsset, error) {
		return mm.store.Assets().GetVersionAsset(ctx, descriptor.Path, ver.ID)
	})
	if err != nil {
		return nil, false, err
	}
	if assetExist {
		if isSameContent(currAsset, descriptor) {
			return nil, false, duplicateAssetError(descriptor.Path)
		}
		newPath, e := mm.resolveAssetPath(descriptor.Path, func(p string) (bool, error) {
			return mm.versionAssetExists(ctx, p, ver.ID)
		})
		if e != nil {
			return nil, false, e
		}
		descriptor.Path = newPath
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
//
// generic/file 上传语义（用户面）：同 group 下若已存在同名资产：
//  1. md5/sha1/sha256 三摘要全等 → 视为重复，返回 ErrAssetDuplicate；
//  2. 任一摘要不同 → 自动追加 `(n)` 后缀直至空位，并用新名建一份新的 asset。
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
		// 如果内容相同，直接返回重复错误。
		if isSameContent(currAsset, descriptor) {
			return nil, false, duplicateAssetError(descriptor.Path)
		}
		// 如果内容不同，追加后缀。
		newPath, e := mm.resolveAssetPath(descriptor.Path, func(p string) (bool, error) {
			return mm.groupAssetExists(ctx, p, groupID)
		})
		if e != nil {
			return nil, false, e
		}
		descriptor.Path = newPath
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

// isSameContent 比对已落库 asset 与新上传 descriptor 是否为同一份内容。
// 仅当 md5/sha1/sha256 三摘要全部解析成功且全部相等时返回 true，
// 任一字段缺失或解析失败一律按"内容不同"处理（让重命名分支接管）。
func isSameContent(asset *model.ArtifactAsset, descriptor *AssetDescriptor) bool {
	if asset == nil || descriptor == nil || descriptor.Hash == nil {
		return false
	}
	var h Hash
	if err := json.Unmarshal([]byte(asset.CheckSum), &h); err != nil {
		return false
	}
	if h.Md5 == "" || h.Sha1 == "" || h.Sha256 == "" {
		return false
	}
	return h.Md5 == descriptor.Hash.Md5 &&
		h.Sha1 == descriptor.Hash.Sha1 &&
		h.Sha256 == descriptor.Hash.Sha256
}

// duplicateAssetError 构造重复上传错误：带 400 状态与 i18n key，供 controller 渲染。
func duplicateAssetError(path string) error {
	err := fmt.Errorf("%w: %s", ErrAssetDuplicate, path)
	return pkgErrors.WithHttpStatus(
		pkgErrors.WithI18nKey(err, i18n.ArtifactUploadFileDuplicate),
		400,
	)
}

// resolveAssetPath 在 basePath 基础上追加 `(n)` 后缀直至命中空位。
// 拆扩展名优先匹配复合扩展白名单（见 splitNameExt），其余回落到最后一个 dot：
//   - foo.tar.gz   → foo(1).tar.gz   （命中 .tar.gz 白名单，走 first-dot 保留完整复合扩展）
//   - foo.min.js   → foo(1).min.js   （命中 .min.js 白名单）
//   - foo.txt      → foo(1).txt      （单扩展，last-dot 与 first-dot 等价）
//   - version.1.2.3.txt → version.1.2.3(1).txt（无白名单命中，按 last-dot 切，避免误判点分版本号）
//   - foo          → foo(1)
//   - .bashrc      → .bashrc(1)      （首字符为 dot 时不视为扩展名）
//
// existsFunc 用于判断候选路径是否已被占用；为防止异常分组下的死循环，加 1000 次硬上限。
func (mm *ModelManager) resolveAssetPath(basePath string, existsFunc func(string) (bool, error)) (string, error) {
	stem, ext := splitNameExt(basePath)
	const maxAttempts = 1000
	for i := 1; i <= maxAttempts; i++ {
		candidate := fmt.Sprintf("%s(%d)%s", stem, i, ext)
		exists, err := existsFunc(candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("resolve duplicate-safe path failed for %q after %d attempts", basePath, maxAttempts)
}

// compoundExtensions 复合扩展名白名单。命中后走 first-dot 切分以保留完整后缀。
// 命中规则：basename（去掉前导隐藏点后）以白名单条目结尾，且白名单本身的前导 dot
// 不是 basename 的首字符（避免 ".tar.gz" 这种纯扩展名文件被错切）。
// 排在前面的更长，避免 ".js" 抢先命中 ".min.js"。
var compoundExtensions = []string{
	".tar.gz",
	".tar.bz2",
	".tar.xz",
	".tar.zst",
	".min.js",
	".min.css",
	".spec.ts",
	".test.ts",
	".d.ts",
}

// splitNameExt 按以下顺序拆分文件名与扩展名：
//  1. 命中复合扩展白名单 → first-dot 风格切（保留 .tar.gz / .min.js 等完整复合扩展）；
//  2. 其余 → last-dot 风格切（避免 version.1.2.3.txt 类点分版本号被误判）；
//  3. 首字符为 dot 的隐藏文件（如 .bashrc）不视为扩展名分隔符。
func splitNameExt(name string) (string, string) {
	base := path.Base(name)
	dir := strings.TrimSuffix(name, base)

	// 隐藏文件首字符为 dot：剥离前导 dot 后再做后缀匹配，避免 ".tar.gz" 整段被吃掉。
	scan := base
	leadingDots := 0
	for leadingDots < len(scan) && scan[leadingDots] == '.' {
		leadingDots++
	}
	scanBody := scan[leadingDots:]

	for _, ce := range compoundExtensions {
		if len(scanBody) > len(ce) && strings.HasSuffix(strings.ToLower(scanBody), ce) {
			cut := len(scan) - len(ce)
			return dir + base[:cut], base[cut:]
		}
	}

	idx := strings.LastIndex(base, ".")
	if idx <= 0 {
		return name, ""
	}
	return dir + base[:idx], base[idx:]
}

// versionAssetExists 判断 version 维度下是否已存在指定 path 的 active 资产。
func (mm *ModelManager) versionAssetExists(ctx context.Context, p string, versionID int64) (bool, error) {
	_, err := mm.store.Assets().GetVersionAsset(ctx, p, versionID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, pkgErrors.ErrResourceNotFound) {
		return false, nil
	}
	return false, err
}

// groupAssetExists 判断 group 维度下是否已存在指定 path 的 active 资产。
func (mm *ModelManager) groupAssetExists(ctx context.Context, p string, groupID int64) (bool, error) {
	_, err := mm.store.Assets().GetGroupAsset(ctx, p, groupID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, pkgErrors.ErrResourceNotFound) {
		return false, nil
	}
	return false, err
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
		Rank:        descriptor.Rank,
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

	currVersion, err := mm.store.Versions().GetByVersion(ctx, mm.pkgModel.ID, mm.view.ArtifactID, mm.pkgDescriptor.Version)
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
		PackageID:     mm.pkgModel.ID,
		ArtifactLibID: mm.view.ArtifactID,
		Version:       mm.pkgDescriptor.Version,
		CreatedBy:     mm.creator.UID,
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
	var currPkg *model.ArtifactPackage
	var err error
	if desc.Format == model.ArtifactContainerFormat {
		unlock, e := mm.store.Packages().LockArtifactScopePackage(ctx, mm.view.ArtifactID, desc.Namespace, desc.Name, desc.Format)
		if e != nil {
			return nil, e
		}
		defer unlock()
		currPkg, err = mm.store.Packages().GetByNameInArtifactScope(ctx, desc.Name, desc.Namespace, mm.view.ArtifactID, desc.Format)
	} else {
		currPkg, err = mm.store.Packages().GetByName(ctx, desc.Name, desc.Namespace, mm.view.ArtifactID, desc.Format)
	}
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
