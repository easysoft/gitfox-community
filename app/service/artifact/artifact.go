package artifact

import (
	"archive/zip"
	"compress/flate"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter/container"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter/generic"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter/helm"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter/raw"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/entity"
	artModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/model"
	enum "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/rbac"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage"
	storagedriver "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage/driver"

	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	spaceStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/space"
	userStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	artifactperm "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/rbac/artifact"
	"github.com/opencontainers/go-digest"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gorm.io/gorm"
)

const artifactStoragePrefix = "artifact"

const (
	JobTypeArtifactContainerGC      = "gitfox:cleanup:artifact-container-garbage-collector"
	JobTypeArtifactSoftRemoveGC     = "gitfox:cleanup:artifact-soft-remove"
	JobTypeArtifactExpiredViewGC    = "gitfox:cleanup:artifact-expired-view"
	JobNameArtifactContainerGC      = "gitfox:recurring:artifact-container-gc"
	JobNameArtifactSoftRemoveGC     = "gitfox:recurring:artifact-soft-remove-gc"
	JobNameArtifactExpiredViewGC    = "gitfox:recurring:artifact-expired-view-gc"
	ArtifactGCJobMaxDurationSeconds = 3600
)

var _ Interface = (*Service)(nil)

// Service 是制品库的核心业务服务。
type Service struct {
	db         *gorm.DB
	artStore   artifactStore.Interface
	spaceStore spaceStore.Interface
	repoStore  repoStore.Interface
	userStore  *userStore.Store
	perm       *artifactperm.Manager
	fileStore  storage.ContentStorage
	scheduler  JobScheduler
	config     *config.ArtifactConfig // 制品库配置（回收站保留期、GC 间隔等）
}

// NewService 创建制品库服务。使用 StorageDriver 内部构建 ContentStorage，
// 避免与 pipeline 的 ContentStorage provider 冲突。
func NewService(
	db *gorm.DB,
	artStore artifactStore.Interface,
	spaceStore spaceStore.Interface,
	repoStore repoStore.Interface,
	uStore *userStore.Store,
	perm *artifactperm.Manager,
	d storagedriver.StorageDriver,
	scheduler JobScheduler,
	cfg *config.ArtifactConfig,
) (*Service, error) {
	fs, err := storage.NewCommonContentStore(context.Background(), d,
		storage.WithPrefix(artifactStoragePrefix))
	if err != nil {
		return nil, err
	}
	// 校验并修正配置：确保 GCInterval >= RecycleBinRetention
	cfg.Sanitize()
	return &Service{
		db:         db,
		artStore:   artStore,
		spaceStore: spaceStore,
		repoStore:  repoStore,
		userStore:  uStore,
		perm:       perm,
		fileStore:  fs,
		scheduler:  scheduler,
		config:     cfg,
	}, nil
}

// FileStore 返回制品库的 ContentStorage，用于外部组件（如容器镜像控制器）直接访问存储。
func (s *Service) FileStore() storage.ContentStorage {
	return s.fileStore
}

// CreateView 创建制品库视图。
//
// 名称唯一性规则（同一作用域 spaceID + repoID 下）：
//  1. 若存在同名的 active 视图（deleted=0）-> 拒绝，返回"名称已存在"。
//  2. 若存在同名的软删视图（deleted=1）且仍在回收站保留期内
//     （editedDate 距今 < config.RecycleBinRetention）-> 拒绝，提示用户从回收站恢复。
//  3. 若存在同名的软删视图但已超过保留期 -> 先物理删除过期 view，再正常创建。
//  4. 无同名记录 -> 正常创建。
//
// 默认分组规则：
//
//	非版本化格式（generic / raw）在视图创建成功后，会自动生成一个与视图同名、
//	同 format 的根级分组，避免用户首次上传时还要手动建组。
//	版本化格式（container / helm / pypi / npm / maven）由各自的 uploader 按协议
//	自行创建/复用分组，这里不提前插入，以免干扰协议语义。
//
// 注意：ops_artifact 表已移除严格唯一索引（uk_spaceID_repoID_name），
// 改由此方法在应用层保证名称唯一性，以支持回收站保留期的时间窗口判断。
func (s *Service) CreateView(ctx context.Context, req *model.CreateArtifactViewReq) (*model.ArtifactView, error) {
	log := logger.GetLogger(ctx)

	// Validate 已在 controller 侧调过，这里再守一次：直接复用 model 的业务校验，
	// 确保内部调用链路（非 HTTP）也能命中中文限制等规则。
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// repoID 不为空时，spaceID 必须同时存在
	if req.RepoID > 0 && req.SpaceID == 0 {
		return nil, pkgErrors.WithI18nKey(
			errors.New("spaceID is required when repoID is provided"),
			i18n.ArtifactSpaceIdRequired,
		)
	}

	// 校验 space / repo 是否存在
	if req.SpaceID > 0 {
		if _, err := s.spaceStore.GetByID(ctx, uint(req.SpaceID)); err != nil {
			return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactSpaceQueryError)
		}
	}
	if req.RepoID > 0 {
		if _, err := s.repoStore.GetByID(ctx, uint(req.RepoID)); err != nil {
			return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactRepoQueryError)
		}
	}

	// 查询同作用域下所有同名 view（ 不包含已软删的）
	existing, err := s.artStore.Views().FindByNameInScope(ctx, req.SpaceID, req.RepoID, req.Name)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactViewQueryError)
	}

	if len(existing) != 0 {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactViewNameDuplicate)
	}

	// 判断是否为该作用域下的第一个视图：复用已有的 GetDefault 方法，
	// 如果查不到默认视图，说明当前作用域还没有视图，新建的就自动设为默认。
	isDefault := !s.hasDefaultView(ctx, req.SpaceID, req.RepoID)
	viewType := resolveArtifactViewType(req.SpaceID, req.RepoID)

	// 情况4：无冲突，正常创建
	v := &model.CreateArtifactView{
		Name:        req.Name,
		Description: req.Description,
		SpaceID:     req.SpaceID,
		RepoID:      req.RepoID,
		Format:      artifact.ArtifactFormat(req.Format),
		Type:        viewType,
		Default:     isDefault,
	}
	view, err := s.artStore.Views().Create(ctx, v)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactViewCreateError)
	}

	// container：为新视图铸造一个默认根分组，名称与 format 与视图保持一致。
	if artifact.ArtifactFormat(req.Format) == artifact.ArtifactContainerFormat {
		if gErr := s.ensureDefaultGroupForView(ctx, view); gErr != nil {
			log.Warnf("create default group for view %d (%s) failed: %v", view.ID, view.Name, gErr)
		}
	}

	return view, nil
}

// ensureDefaultGroupForView 为非版本化视图创建一个与视图同名的根分组，并同步 tree node。
// 使用 Groups().RecursiveCreate：本身幂等，若已存在同名根分组会直接复用。
func (s *Service) ensureDefaultGroupForView(ctx context.Context, view *model.ArtifactView) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		ge, err := s.artStore.Groups().RecursiveCreate(ctx, view.ID, []string{view.Name}, view.Format)
		if err != nil {
			return err
		}
		if ge == nil || ge.ArtifactGroup == nil {
			return errors.New("default group endpoint is empty after create")
		}
		if _, err = artModel.NewSyncer(s.artStore).SyncTreeNodeForGroup(ctx, ge.ArtifactGroup); err != nil {
			return err
		}
		return nil
	})
}

func resolveArtifactViewType(spaceID, repoID int64) model.ArtifactViewType {
	switch {
	case repoID > 0:
		return model.ArtifactViewTypeRepo
	case spaceID > 0:
		return model.ArtifactViewTypeSpace
	default:
		return model.ArtifactViewTypeGlobal
	}
}

// hasDefaultView 检查指定作用域下是否已存在默认视图。
// 复用已有的 GetDefault / GetDefaultByRepoAndSpace 方法，无需额外的 store 查询。
func (s *Service) hasDefaultView(ctx context.Context, spaceID, repoID int64) bool {
	var err error
	if repoID > 0 {
		_, err = s.artStore.Views().GetDefaultByRepoAndSpace(ctx, repoID, spaceID)
	} else {
		_, err = s.artStore.Views().GetDefault(ctx, spaceID)
	}
	return err == nil
}

// ResolveView 按请求中的 artifactID 或默认视图规则解析实际视图。
// 规则：
//  1. artifactID > 0：优先按 ID 精确取库
//  2. artifactID == 0 或未传：继续按现有默认 view 逻辑解析
func (s *Service) ResolveView(ctx context.Context, req *model.SpaceRepoReq) (*model.ArtifactView, error) {
	if req == nil {
		req = &model.SpaceRepoReq{}
	}

	if req.ArtifactID > 0 {
		view, err := s.artStore.Views().GetByID(ctx, req.ArtifactID)
		if err != nil {
			if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
				return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactViewNotFound)
			}
			return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactViewQueryError)
		}
		if req.SpaceID > 0 && view.SpaceID != req.SpaceID {
			return nil, pkgErrors.WithI18nKey(
				fmt.Errorf("artifactID %d does not belong to spaceID %d", req.ArtifactID, req.SpaceID),
				i18n.ArtifactViewScopeMismatch,
			)
		}
		if req.RepoID > 0 {
			if req.SpaceID == 0 {
				return nil, pkgErrors.WithI18nKey(
					errors.New("spaceID is required when repoID is provided"),
					i18n.ArtifactSpaceIdRequired,
				)
			}
			if view.RepoID != req.RepoID {
				return nil, pkgErrors.WithI18nKey(
					fmt.Errorf("artifactID %d does not belong to repoID %d under spaceID %d", req.ArtifactID, req.RepoID, req.SpaceID),
					i18n.ArtifactViewScopeMismatch,
				)
			}
		}
		return view, nil
	}

	if req.RepoID > 0 {
		if req.SpaceID == 0 {
			return nil, pkgErrors.WithI18nKey(
				errors.New("spaceID is required when repoID is provided"),
				i18n.ArtifactSpaceIdRequired,
			)
		}
		view, err := s.artStore.Views().GetDefaultByRepoAndSpace(ctx, req.RepoID, req.SpaceID)
		if err != nil {
			return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactDefaultViewQueryError)
		}
		return view, nil
	}

	view, err := s.artStore.Views().GetDefault(ctx, req.SpaceID)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactDefaultViewQueryError)
	}
	return view, nil
}

func (s *Service) fillViewDescriptorScope(ctx context.Context, desc *adapter.ViewDescriptor, view *model.ArtifactView) error {
	if view.SpaceID > 0 {
		space, err := s.spaceStore.GetByID(ctx, uint(view.SpaceID))
		if err != nil {
			return pkgErrors.WithI18nKey(err, i18n.ArtifactSpaceQueryError)
		}
		desc.Space = &model.SpaceRef{ID: int64(space.ID), Name: space.Name}
	}

	if view.RepoID > 0 {
		repo, err := s.repoStore.GetByID(ctx, uint(view.RepoID))
		if err != nil {
			return pkgErrors.WithI18nKey(err, i18n.ArtifactRepoQueryError)
		}
		desc.Repo = &model.RepoRef{ID: int64(repo.ID), GitUID: repo.GitUID, Name: repo.Identifier}
	}

	return nil
}

// ParseSpaceView 解析 SpaceRepoReq 参数为 ViewDescriptor。
func (s *Service) ParseSpaceView(ctx context.Context, req *model.SpaceRepoReq) (*adapter.ViewDescriptor, error) {
	view, err := s.ResolveView(ctx, req)
	if err != nil {
		return nil, err
	}

	desc := &adapter.ViewDescriptor{
		ArtifactID: view.ID,
		Store:      s.fileStore,
		StorageID:  1,
	}
	if err = s.fillViewDescriptorScope(ctx, desc, view); err != nil {
		return nil, err
	}
	return desc, nil
}

// ContainerArtifactCtx 容器镜像定位的查询结果。
// 供 container handler 构造 ContainerReq 使用，包含下游需要的最小信息。
type ContainerArtifactCtx struct {
	Scope     string                  // "global" | "space" | "repo"
	SpaceID   int64                   // space 作用域下 > 0
	RepoID    int64                   // repo 作用域下 > 0
	View      *adapter.ViewDescriptor // artifact 对应的 view descriptor
	GroupName string                  // 容器制品库下唯一 group 的 name
}

// ScopeID 返回对应 scope 的数字 ID（global 时为 0）。
func (c *ContainerArtifactCtx) ScopeID() uint64 {
	switch c.Scope {
	case "space":
		return uint64(c.SpaceID)
	case "repo":
		return uint64(c.RepoID)
	default:
		return 0
	}
}

// resolveContainerArtifact 容器路由的统一私有解析流程。
// 步骤 1：按 (spaceID, repoID, name) 在 ops_artifact 精确查找 view 记录；
//   - ErrResourceNotFound → 包装 i18n.ArtifactContainerArtifactNotFound
//   - 其它错误 → 包装 i18n.ArtifactViewQueryError
//
// 步骤 2：调用 s.lookupContainerGroupName(ctx, view.ID) 取容器 group 名。
// 步骤 3：组装并返回 ContainerArtifactCtx（含 ViewDescriptor{ArtifactID, Store: s.fileStore, StorageID: 1}）。
func (s *Service) resolveContainerArtifact(
	ctx context.Context, scope string, spaceID, repoID int64, artifactName string,
) (*ContainerArtifactCtx, error) {
	// 1) 按 (spaceID, repoID, name) 精确查找 artifact 视图
	view, err := s.artStore.Views().GetByName(ctx, spaceID, repoID, artifactName)
	if err != nil {
		if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
			return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactContainerArtifactNotFound)
		}
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactViewQueryError)
	}

	// 2) 容器库下只有一个 group，复用已有 Groups().ListDescendantsForOwner 取 name
	groupName, err := s.lookupContainerGroupName(ctx, view.ID)
	if err != nil {
		return nil, err
	}

	// 3) 组装返回值
	return &ContainerArtifactCtx{
		Scope:     scope,
		SpaceID:   spaceID,
		RepoID:    repoID,
		View:      &adapter.ViewDescriptor{ArtifactID: view.ID, Store: s.fileStore, StorageID: 1},
		GroupName: groupName,
	}, nil
}

// ResolveGlobalArtifact 全局作用域下按 artifact name 精确定位制品库及其容器 group。
// 对应路由：/v2/global/{artifactname}/{imagename}
func (s *Service) ResolveGlobalArtifact(
	ctx context.Context, artifactName string,
) (*ContainerArtifactCtx, error) {
	return s.resolveContainerArtifact(ctx, "global", 0, 0, artifactName)
}

// ResolveSpaceArtifact 空间作用域下按 (spaceID, artifactName) 精确定位制品库及其容器 group。
// 对应路由：/v2/space{spaceID}/{artifactname}/{imagename}
func (s *Service) ResolveSpaceArtifact(
	ctx context.Context, spaceID uint64, artifactName string,
) (*ContainerArtifactCtx, error) {
	return s.resolveContainerArtifact(ctx, "space", int64(spaceID), 0, artifactName)
}

// ResolveRepoArtifact 仓库作用域下按 (repoID, artifactName) 精确定位制品库及其容器 group。
// 对应路由：/v2/repo{repoID}/{artifactname}/{imagename}
// 与现有 ParseContainerView 行为保持一致：先通过 repoStore 按 ID 反查拿到 repo，
// 再以 (int64(repo.SpaceID), int64(repoID), artifactName) 调用 resolveContainerArtifact。
// 不引入新复杂度，保留现有已验证过的查询方式。
func (s *Service) ResolveRepoArtifact(
	ctx context.Context, repoID uint64, artifactName string,
) (*ContainerArtifactCtx, error) {
	// 通过 repo 反查 spaceID（与现有 ParseContainerView 保持一致，不改动此行为）
	repo, err := s.repoStore.GetByID(ctx, uint(repoID))
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactRepoQueryError)
	}
	return s.resolveContainerArtifact(ctx, "repo", int64(repo.SpaceID), int64(repoID), artifactName)
}

// lookupContainerGroupName 读取 artifact 下唯一容器 group 的 name。
// 复用已有 store 方法 Groups().ListDescendantsForOwner：容器库下只会有一个 group，
// 直接取结果切片的第一条。不新增 store 方法，保持最小改动。
func (s *Service) lookupContainerGroupName(ctx context.Context, artifactID int64) (string, error) {
	groups, err := s.artStore.Groups().ListDescendantsForOwner(ctx, artifactID, model.ArtifactContainerFormat)
	if err != nil {
		// 查询失败统一包装 container_group_not_found
		return "", pkgErrors.WithI18nKey(err, i18n.ArtifactContainerGroupNotFound)
	}
	if len(groups) == 0 {
		// 空列表同样视作 container_group_not_found
		return "", pkgErrors.WithI18nKey(pkgErrors.ErrResourceNotFound, i18n.ArtifactContainerGroupNotFound)
	}
	// 容器只会有一个 group，取第一条即可
	return groups[0].Name, nil
}

func (s *Service) ListPackages(ctx context.Context, artifactID int64, filter *model.ArtifactFilter) ([]*model.ArtifactListItem, error) {
	res, err := s.artStore.FindPackages(ctx, artifactID, filter)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactPackageListError)
	}
	return res, nil
}

func (s *Service) ListVersions(ctx context.Context, artifactID int64,
	packageName, groupName string, format model.ArtifactFormat, opt model.SearchVersionOption,
) ([]*model.ArtifactVersion, error) {
	dbPkg, err := s.artStore.Packages().GetByName(ctx, packageName, groupName, artifactID, format)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactPackageQueryError)
	}
	opt.PackageID = dbPkg.ID
	return s.artStore.Versions().Find(ctx, opt)
}

func (s *Service) ListAssets(ctx context.Context, VersionID int64) ([]*model.ArtifactAssetsRes, error) {
	return s.artStore.ListAssets(ctx, VersionID)
}

func (s *Service) ListGroupAssets(ctx context.Context, groupID int64) ([]*model.ArtifactAssetsRes, error) {
	return s.artStore.ListGroupAssets(ctx, groupID)
}

func (s *Service) GetAsset(ctx context.Context, assetID int64) (*model.ArtifactAssetsRes, error) {
	return s.artStore.GetAsset(ctx, assetID)
}

func (s *Service) GetBlobHistory(ctx context.Context, assetID int64) ([]*model.ArtifactBlobRes, error) {
	return s.artStore.Blobs().History(ctx, assetID)
}

// ListNodeInfo 兼容参考项目 `/artifacts/nodeInfo`，按 entityIDs 批量返回节点详情。
func (s *Service) ListNodeInfo(ctx context.Context, req *model.ListNodeInfoReq) ([]*model.ArtifactNodeInfo, error) {
	ids := req.EntityIDs

	result := make([]*model.ArtifactNodeInfo, 0, len(ids))
	for _, entityID := range ids {
		info, err := s.getNodeInfoByEntityID(ctx, entityID)
		if err != nil {
			return nil, err
		}
		result = append(result, info)
	}
	return result, nil
}

func (s *Service) getNodeInfoByEntityID(ctx context.Context, entityID string) (*model.ArtifactNodeInfo, error) {
	kind, id, err := parseNodeInfoEntityID(entityID)
	if err != nil {
		return invalidNodeInfo(), nil
	}

	switch kind {
	case model.ArtifactEntityKindVersion:
		return s.getVersionNodeInfo(ctx, id)
	case model.ArtifactEntityKindAsset:
		return s.getAssetNodeInfo(ctx, id)
	default:
		return invalidNodeInfo(), nil
	}
}

func parseNodeInfoEntityID(entityID string) (model.ArtifactEntityKind, int64, error) {
	if strings.Count(entityID, ".") != 1 {
		return "", 0, fmt.Errorf("invalid artifact entity id: %s", entityID)
	}

	parts := strings.Split(entityID, ".")
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid artifact entity id number: %s", entityID)
	}
	return model.ArtifactEntityKind(parts[0]), id, nil
}

func (s *Service) getVersionNodeInfo(ctx context.Context, versionID int64) (*model.ArtifactNodeInfo, error) {
	ver, err := s.artStore.Versions().GetByID(ctx, versionID)
	if err != nil {
		if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
			return notFoundNodeInfo(), nil
		}
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactVersionQueryError)
	}

	pkg, err := s.artStore.Packages().GetByID(ctx, ver.PackageID)
	if err != nil {
		if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
			return notFoundNodeInfo(), nil
		}
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactPackageQueryError)
	}

	view, err := s.ParseSpaceView(ctx, &model.SpaceRepoReq{ArtifactID: pkg.ArtifactID})
	if err != nil {
		return nil, err
	}

	info := &model.ArtifactNodeInfo{
		Status: string(model.ArtifactStatusOK),
		Format: pkg.Format,
		Metadata: &model.ArtifactPkgVerMetadata{
			Space:   nodeInfoSpaceName(view),
			Name:    pkg.Name,
			Group:   pkg.Namespace,
			Version: ver.Version,
		},
	}

	if pkg.Format == model.ArtifactContainerFormat {
		if err = s.fillContainerNodeInfo(ctx, info, view, ver); err != nil {
			return nil, err
		}
		return info, nil
	}

	assetRes, err := s.getPrimaryVersionAsset(ctx, ver.ID)
	if err != nil {
		return nil, err
	}
	if assetRes == nil {
		return info, nil
	}

	populateNodeInfoAssetFields(info, assetRes)
	info.Link = buildArtifactAssetLink(nodeInfoSpaceName(view), string(pkg.Format), pkg.Namespace, pkg.Name, ver.Version, assetRes.Path)
	return info, nil
}

// getAssetNodeInfo 按 assetID 查询节点详情，自动关联版本 / 分组信息并拼装 ArtifactNodeInfo。
func (s *Service) getAssetNodeInfo(ctx context.Context, assetID int64) (*model.ArtifactNodeInfo, error) {
	// 查 asset 元信息（format / versionID / groupID / artifactID）
	dbAsset, err := s.artStore.Assets().GetById(ctx, assetID)
	if err != nil {
		if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
			return notFoundNodeInfo(), nil
		}
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactAssetQueryError)
	}

	// 查 asset 详情（含 blob 大小、path、contentType、checksum、创建人、时间）
	// path      → 存储路径 / container digest
	// size      → 大小（字节）
	// checksum  → md5 / sha1 / sha256 / sha512 校验值
	// creatorName → 创建人 UID
	// created / updated → 创建 / 更新时间（毫秒时间戳）
	assetRes, err := s.artStore.GetAsset(ctx, assetID)
	if err != nil {
		if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
			return notFoundNodeInfo(), nil
		}
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactAssetQueryError)
	}

	if dbAsset.VersionID.Valid && dbAsset.VersionID.Int64 > 0 {
		// asset 挂在 version 下 → 先拿 version 级信息（name 名称、version 版本、group 分组、space 所属空间）
		// container 格式在 getVersionNodeInfo 内部还会调 fillContainerNodeInfo，
		// 把 manifest 解析出的 layer 累计 size / os / arch / created / updated 写入 info。
		info, err := s.getVersionNodeInfo(ctx, dbAsset.VersionID.Int64)
		if err != nil {
			return nil, err
		}
		if info.Status != string(model.ArtifactStatusOK) {
			return info, nil
		}

		// container 格式：保存 fillContainerNodeInfo 写入的字段，防止被 populateNodeInfoAssetFields 覆盖。
		// populateNodeInfoAssetFields 会用 manifest blob 的元信息（asset 表 size/createdDate）
		// 覆盖 info.Size/Created/Updated，但 container 应保留 layer 累计大小和 manifest 时间。
		isContainer := info.Format == model.ArtifactContainerFormat
		var savedSize int64
		var savedOS, savedArch string
		var savedCreated, savedUpdated *datetime.Local
		if isContainer {
			savedSize = info.Size
			savedOS = info.OS
			savedArch = info.Arch
			savedCreated = info.Created
			savedUpdated = info.Updated
		}

		// 用 asset 数据填充 path / contentType / creatorName / link / checksum
		populateNodeInfoAssetFields(info, assetRes)

		if isContainer {
			// 恢复 manifest 解析结果：layer 累计大小、操作系统、CPU 架构、时间
			info.Size = savedSize
			info.OS = savedOS
			info.Arch = savedArch
			info.Created = savedCreated
			info.Updated = savedUpdated
		}

		if info.Metadata != nil {
			// link 下载链接（含 space / format / group / name / version / path 拼成的访问路径）
			info.Link = buildArtifactAssetLink(info.Metadata.Space, string(info.Format), info.Metadata.Group, info.Metadata.Name, info.Metadata.Version, assetRes.Path)
		}
		return info, nil
	}

	if dbAsset.GroupID.Valid && dbAsset.GroupID.Int64 > 0 {
		// asset 直挂 group（generic / raw 的目录级文件），name 取文件 basename
		return s.getGroupAssetNodeInfo(ctx, dbAsset, assetRes)
	}

	if dbAsset.ArtifactID.Valid && dbAsset.ArtifactID.Int64 > 0 {
		// asset 直挂 view（standalone，无包无版本归属）
		return s.getStandaloneAssetNodeInfo(ctx, dbAsset, assetRes)
	}

	// 兜底：没有归属信息，只填基础字段
	info := &model.ArtifactNodeInfo{
		Status: string(model.ArtifactStatusOK),
		Format: dbAsset.Format,
	}
	populateNodeInfoAssetFields(info, assetRes)
	return info, nil
}

func (s *Service) getPrimaryVersionAsset(ctx context.Context, versionID int64) (*model.ArtifactAssetsRes, error) {
	mainAssets, err := s.artStore.Assets().FindMain(ctx, versionID)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactAssetQueryError)
	}
	if len(mainAssets) > 0 {
		return s.artStore.GetAsset(ctx, mainAssets[0].ID)
	}

	assets, err := s.artStore.ListAssets(ctx, versionID)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactAssetQueryError)
	}
	if len(assets) == 0 {
		return nil, nil
	}
	return assets[0], nil
}

func (s *Service) fillContainerNodeInfo(
	ctx context.Context,
	info *model.ArtifactNodeInfo,
	view *adapter.ViewDescriptor,
	ver *model.ArtifactVersion,
) error {
	mainAssets, err := s.artStore.Assets().FindMain(ctx, ver.ID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactAssetQueryError)
	}
	if len(mainAssets) == 0 {
		return nil
	}

	dgst, err := digest.Parse(mainAssets[0].Path)
	if err != nil {
		return err
	}

	// 解析 manifest，获取各平台（OS/Arch）信息及每层 layer 累计大小
	reader := container.NewTagMetadataReader(s.artStore, view)
	data, err := reader.Parse(ctx, mainAssets[0].ContentType, dgst)
	if err != nil {
		return err
	}

	if len(data.Images) == 0 {
		return nil
	}

	// 优先选 linux/amd64；multi-arch 镜像无 linux/amd64 时取第一个平台（与参考项目一致）
	selected := data.Images[0]
	for _, image := range data.Images {
		if image.OS == "linux" && image.Arch == "amd64" {
			selected = image
			break
		}
	}

	// 大小：各 layer 之和（不含 manifest/config 文件本身）
	info.Size = selected.Size
	// 操作系统（linux / windows 等）
	info.OS = selected.OS
	// CPU 架构（amd64 / arm64 / arm/v7 等）
	info.Arch = selected.Arch
	// 创建时间取 version 创建时间，更新时间取 manifest asset 的写入时间
	info.Created = ver.CreatedDate
	info.Updated = mainAssets[0].CreatedDate
	return nil
}

func (s *Service) getGroupAssetNodeInfo(
	ctx context.Context,
	dbAsset *model.ArtifactAsset,
	assetRes *model.ArtifactAssetsRes,
) (*model.ArtifactNodeInfo, error) {
	group, err := s.artStore.Groups().GetEndpointById(ctx, dbAsset.GroupID.Int64)
	if err != nil {
		if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
			return notFoundNodeInfo(), nil
		}
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactGroupQueryError)
	}

	view, err := s.ParseSpaceView(ctx, &model.SpaceRepoReq{ArtifactID: group.ArtifactID})
	if err != nil {
		return nil, err
	}

	info := &model.ArtifactNodeInfo{
		Status: string(model.ArtifactStatusOK),
		Format: dbAsset.Format,
		Metadata: &model.ArtifactPkgVerMetadata{
			Space: nodeInfoSpaceName(view),
			Name:  filepath.Base(assetRes.Path),
			Group: group.GroupName(),
		},
	}
	populateNodeInfoAssetFields(info, assetRes)
	info.Link = buildArtifactAssetLink(nodeInfoSpaceName(view), string(dbAsset.Format), group.GroupPath(), "", "", assetRes.Path)
	return info, nil
}

func (s *Service) getStandaloneAssetNodeInfo(
	ctx context.Context,
	dbAsset *model.ArtifactAsset,
	assetRes *model.ArtifactAssetsRes,
) (*model.ArtifactNodeInfo, error) {
	view, err := s.ParseSpaceView(ctx, &model.SpaceRepoReq{ArtifactID: dbAsset.ArtifactID.Int64})
	if err != nil {
		return nil, err
	}

	info := &model.ArtifactNodeInfo{
		Status: string(model.ArtifactStatusOK),
		Format: dbAsset.Format,
		Metadata: &model.ArtifactPkgVerMetadata{
			Space: nodeInfoSpaceName(view),
			Name:  filepath.Base(assetRes.Path),
		},
	}
	populateNodeInfoAssetFields(info, assetRes)
	info.Link = buildArtifactAssetLink(nodeInfoSpaceName(view), string(dbAsset.Format), "", "", "", assetRes.Path)
	return info, nil
}

func populateNodeInfoAssetFields(info *model.ArtifactNodeInfo, asset *model.ArtifactAssetsRes) {
	if info == nil || asset == nil {
		return
	}

	info.Path = asset.Path
	info.ContentType = asset.ContentType
	info.Size = asset.Size
	info.CreatorName = asset.CreatorName
	info.Created = asset.Created
	info.Updated = asset.Updated
	info.Link = asset.Link
	info.CheckSum = asset.CheckSum
}

func buildArtifactAssetLink(spaceName, format, group, packageName, version, assetPath string) string {
	if spaceName == "" || format == "" || assetPath == "" {
		return ""
	}

	segments := []string{"_artifacts", spaceName, format}
	switch model.ArtifactFormat(format) {
	case model.ArtifactGenericFormat:
		segments = appendSlashSegments(segments, group)
	case model.ArtifactHelmFormat:
		// helm 仓库直接以文件路径挂载在格式根下
	default:
		if group != "" {
			segments = append(segments, strings.Split(strings.Trim(group, "."), ".")...)
		}
		if packageName != "" {
			segments = append(segments, packageName)
		}
		if version != "" {
			segments = append(segments, version)
		}
	}
	segments = appendSlashSegments(segments, assetPath)

	link, err := url.JoinPath("/", segments...)
	if err != nil {
		return ""
	}
	return link
}

func appendSlashSegments(dst []string, p string) []string {
	trimmed := strings.Trim(p, "/")
	if trimmed == "" {
		return dst
	}
	return append(dst, strings.Split(trimmed, "/")...)
}

func nodeInfoSpaceName(view *adapter.ViewDescriptor) string {
	if view != nil && view.Space != nil {
		return view.Space.Name
	}
	return ""
}

func invalidNodeInfo() *model.ArtifactNodeInfo {
	return &model.ArtifactNodeInfo{Status: string(model.ArtifactStatusInvalidId)}
}

func notFoundNodeInfo() *model.ArtifactNodeInfo {
	return &model.ArtifactNodeInfo{Status: string(model.ArtifactStatusNotFound)}
}

func (s *Service) DownloadAsset(ctx context.Context, assetID int64) (io.ReadCloser, *model.ArtifactBlob, error) {
	asset, err := s.artStore.Assets().GetById(ctx, assetID)
	if err != nil {
		return nil, nil, pkgErrors.WithI18nKey(err, i18n.ArtifactAssetQueryError)
	}
	blob, err := s.artStore.Blobs().GetById(ctx, asset.BlobID)
	if err != nil {
		return nil, nil, pkgErrors.WithI18nKey(err, i18n.ArtifactBlobQueryError)
	}
	blobPath := adapter.BlobPath(blob.Ref)
	reader, err := s.fileStore.Open(ctx, blobPath)
	if err != nil {
		return nil, nil, pkgErrors.WithI18nKey(err, i18n.ArtifactDownloadOpenError)
	}
	return reader, blob, nil
}

func (s *Service) Upload(ctx context.Context, view *adapter.ViewDescriptor, format model.ArtifactFormat, req *http.Request) ([]*adapter.SavedAsset, error) {
	var uploader adapter.ArtifactPackageUploader

	switch format {
	case model.ArtifactRawFormat:
		uploader = raw.NewUploader(s.fileStore, s.artStore, view)
	case model.ArtifactGenericFormat:
		uploader = generic.NewUploader(s.fileStore, s.artStore, view)
	default:
		return nil, pkgErrors.WithI18nKey(
			errors.New("unsupported artifact format"),
			i18n.ArtifactUnsupportedFormat,
		)
	}

	if _, err := uploader.Serve(ctx, req); err != nil {
		return nil, wrapUploadError(err, i18n.ArtifactUploadServeError)
	}
	if err := uploader.IsValid(ctx); err != nil {
		_ = uploader.Cancel(ctx)
		return nil, wrapUploadError(err, i18n.ArtifactUploadValidateError)
	}
	if err := uploader.Save(ctx); err != nil {
		return nil, wrapUploadError(err, i18n.ArtifactUploadSaveError)
	}
	return uploader.GetAssets(), nil
}

// wrapUploadError 若 err 已带有专属 i18n key（非默认），直接透传，
// 否则套上通用 fallback，避免覆盖 adapter 层给出的精准提示。
func wrapUploadError(err error, fallback i18n.I18nKey) error {
	if err == nil {
		return nil
	}
	if key := pkgErrors.GetI18nKey(err); key != "" && key != i18n.ErrorDefault {
		return err
	}
	return pkgErrors.WithI18nKey(err, fallback)
}

// ViewExists 检查指定 artifactID 是否存在（仅 active view）。
func (s *Service) ViewExists(ctx context.Context, artifactID int64) error {
	_, err := s.artStore.Views().GetByID(ctx, artifactID)
	if err != nil {
		return pkgErrors.WithI18nKey(err, i18n.ArtifactGroupArtifactidNotFound)
	}
	return nil
}

// GroupPathExists 检查指定视图下是否已存在完全相同的分组路径。
func (s *Service) GroupPathExists(ctx context.Context, artifactID int64, names []string, format model.ArtifactFormat) (bool, error) {
	_, err := s.artStore.Groups().GetEndpointByNames(ctx, artifactID, names, format)
	if err != nil {
		if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
			return false, nil
		}
		return false, pkgErrors.WithI18nKey(err, i18n.ArtifactGroupQueryError)
	}
	return true, nil
}

// ActiveGroupPathExists 检查指定视图下是否已存在**未软删**的同名分组路径。
// 与 GroupPathExists 的区别：软删的 group 不算"存在"，允许重新创建同名。
func (s *Service) ActiveGroupPathExists(ctx context.Context, artifactID int64, names []string, format model.ArtifactFormat) (bool, error) {
	endpoint, err := s.artStore.Groups().GetEndpointByNames(ctx, artifactID, names, format)
	if err != nil {
		if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
			return false, nil
		}
		return false, pkgErrors.WithI18nKey(err, i18n.ArtifactGroupQueryError)
	}
	// 末级 group 已软删，视为不存在
	if endpoint.ArtifactGroup != nil && endpoint.ArtifactGroup.Deleted == 1 {
		return false, nil
	}
	return true, nil
}

func (s *Service) CreateGroup(ctx context.Context, artifactID int64, names []string, format model.ArtifactFormat) (*model.ArtifactGroupEndpoint, error) {
	var result *model.ArtifactGroupEndpoint
	err := s.db.Transaction(func(tx *gorm.DB) error {
		ge, e := s.artStore.Groups().RecursiveCreate(ctx, artifactID, names, format)
		if e != nil {
			return e
		}
		// 同步树节点
		if ge == nil || ge.ArtifactGroup == nil {
			return errors.New("group endpoint is empty after create")
		}
		if _, e = artModel.NewSyncer(s.artStore).SyncTreeNodeForGroup(ctx, ge.ArtifactGroup); e != nil {
			return e
		}
		result = ge
		return nil
	})
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactGroupCreateError)
	}
	return result, nil
}

// ListTree 是通用树查询入口。
// 语义：
//  1. nodeID > 0 时，优先按节点ID定位当前节点，只支持单层查询；
//  2. path 不带 /** 时，返回“当前层”的平铺节点列表；
//  3. path 为 /** 或 /xxx/** 时，返回从该路径开始的递归树；
//  4. format=all 或留空：
//     - path == "/"：返回各 format 根节点（递归模式下展开各自子树）；
//     - path == "/xxx[/**]"：遍历所有 format 根，按 (artifactID, path, format) 逐个 GetByPath，
//     找到即进入该 format 的分支结果，未命中跳过；多 format 同名路径会合并返回。
//  5. level 表示最大可见层级：format/group/package/version/asset。
func (s *Service) ListTree(ctx context.Context, artifactID int64, filter *model.ArtifactTreeFilter) ([]*model.ArtifactTreeRes, error) {
	if filter == nil {
		filter = &model.ArtifactTreeFilter{}
	}
	level, err := normalizeTreeLevel(filter.Level)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactTreeLevelInvalid)
	}
	path, recursive := normalizeTreePath(filter.Path)

	// 1) nodeID 分支：精确定位单节点。
	if filter.NodeID > 0 {
		if recursive {
			return nil, pkgErrors.WithI18nKey(errors.New("recursive query is not supported when nodeID is provided"), i18n.ArtifactTreeNodeRecursiveConflict)
		}
		currNode, err := s.artStore.Nodes().GetById(ctx, filter.NodeID)
		if err != nil {
			return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactTreeNodeListError)
		}
		if currNode.ArtifactID != artifactID {
			return nil, pkgErrors.WithI18nKey(errors.New("node does not belong to current artifact"), i18n.ArtifactTreeNodeNotInArtifact)
		}
		if level == model.ArtifactNodeLevelFormat {
			return []*model.ArtifactTreeRes{convertTreeNodeToRes(currNode, level)}, nil
		}
		return s.listTreeCurrentLevel(ctx, currNode, level)
	}

	// 2) format 留空视为 all，统一走跨格式分支。
	isAll := filter.Format == "" || filter.Format == model.ArtifactAllFormat

	if isAll {
		roots, err := s.artStore.Nodes().ListFormats(ctx, artifactID)
		if err != nil {
			return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactTreeFormatListError)
		}
		// 2.a path 为根：保持旧语义，返回 format 森林 / 只返 roots。
		if path == model.ArtifactNodeRoot {
			if level == model.ArtifactNodeLevelFormat || !recursive {
				return convertTreeNodesToRes(roots, level), nil
			}
			return s.buildFormatForest(ctx, roots, level)
		}
		// 2.b path 指向具体节点：跨 format 逐个匹配，找到即入结果集。
		return s.listTreeForAllFormats(ctx, artifactID, path, recursive, level, roots)
	}

	// 3) 具体 format + path 分支。
	currNode, err := s.artStore.Nodes().GetByPath(ctx, artifactID, path, filter.Format)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactTreeNodeListError)
	}
	if level == model.ArtifactNodeLevelFormat {
		if path == model.ArtifactNodeRoot {
			return []*model.ArtifactTreeRes{convertTreeNodeToRes(currNode, level)}, nil
		}
		return []*model.ArtifactTreeRes{}, nil
	}
	if !recursive {
		return s.listTreeCurrentLevel(ctx, currNode, level)
	}
	if path == model.ArtifactNodeRoot {
		return s.buildTreeChildren(ctx, currNode, level)
	}
	root := convertTreeNodeToRes(currNode, level)
	root.Children, err = s.buildTreeChildren(ctx, currNode, level)
	if err != nil {
		return nil, err
	}
	root.Leaf = len(root.Children) == 0
	return []*model.ArtifactTreeRes{root}, nil
}

// listTreeForAllFormats 跨 format 按 path 查询：逐个 format 根 GetByPath，
// 命中则按 recursive / level 铺开结果，未命中的 format 静默跳过。
// 多 format 下若存在同名路径，结果会各自作为顶层元素返回。
func (s *Service) listTreeForAllFormats(
	ctx context.Context,
	artifactID int64,
	path string,
	recursive bool,
	level string,
	roots []*model.ArtifactTreeNode,
) ([]*model.ArtifactTreeRes, error) {
	data := make([]*model.ArtifactTreeRes, 0, len(roots))
	for _, root := range roots {
		currNode, err := s.artStore.Nodes().GetByPath(ctx, artifactID, path, root.Format)
		if err != nil {
			// 此 format 下不存在该 path：跳过，不让单 format 缺失拖垮整批查询。
			if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
				continue
			}
			return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactTreeNodeListError)
		}
		if level == model.ArtifactNodeLevelFormat {
			// 非根路径 + level=format 无意义，按既有单 format 语义返空。
			continue
		}
		if !recursive {
			items, lErr := s.listTreeCurrentLevel(ctx, currNode, level)
			if lErr != nil {
				return nil, lErr
			}
			data = append(data, items...)
			continue
		}
		item := convertTreeNodeToRes(currNode, level)
		children, bErr := s.buildTreeChildren(ctx, currNode, level)
		if bErr != nil {
			return nil, bErr
		}
		item.Children = children
		item.Leaf = len(children) == 0
		data = append(data, item)
	}
	return data, nil
}

// buildFormatForest 将多个 format 根节点分别递归展开，组装成整库的树森林。
// 仅在 format=all 且 path 使用递归模式时调用。
func (s *Service) buildFormatForest(ctx context.Context, roots []*model.ArtifactTreeNode, level string) ([]*model.ArtifactTreeRes, error) {
	data := make([]*model.ArtifactTreeRes, 0, len(roots))
	for _, root := range roots {
		item := convertTreeNodeToRes(root, level)
		children, err := s.buildTreeChildren(ctx, root, level)
		if err != nil {
			return nil, err
		}
		item.Children = children
		item.Leaf = len(children) == 0
		data = append(data, item)
	}
	return data, nil
}

func (s *Service) ListFormats(ctx context.Context, artifactID int64) ([]*model.ArtifactTreeNode, error) {
	return s.artStore.Nodes().ListFormats(ctx, artifactID)
}

func (s *Service) ListTreeNodes(ctx context.Context, parentID int64) ([]*model.ArtifactTreeNode, error) {
	return s.artStore.Nodes().ListByParentId(ctx, parentID)
}

// listTreeCurrentLevel 处理“非递归模式”的当前层查询。
// 它只返回 node 的直接下一层：
//  1. 普通节点返回直接 children；
//  2. group 节点在 level=asset 时，会把 group 下直接资产一并拼进来；
//  3. version 节点在 level=asset 时，直接下钻返回该 version 下的资产。
//
// 关于 container：asset 从不写入 ops_artifact_treenode（tree_sync.go 没有 SyncTreeNodeForAsset），
// 所以 container version 同样必须走 listVersionAssetAsTree（即按 versionID 直查 asset 表），
// 否则 ListByParentId 会扑空，前端永远看不到 container 的 main asset。
// 注意 container layer blob 走 isolated（VersionID=0，见 manifest_upload.go::AttrAssetIsolated），
// 不会被这里错列；这一段只会返回挂在 version 上的 main asset（manifest）。
func (s *Service) listTreeCurrentLevel(ctx context.Context, node *model.ArtifactTreeNode, level string) ([]*model.ArtifactTreeRes, error) {
	if node.Type == model.ArtifactTreeNodeTypeVersion && node.TargetID.Valid {
		if level == model.ArtifactNodeLevelAsset {
			return s.listVersionAssetAsTree(ctx, node)
		}
		return []*model.ArtifactTreeRes{convertTreeNodeToRes(node, level)}, nil
	}
	children, err := s.artStore.Nodes().ListByParentId(ctx, node.ID)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactTreeNodeListError)
	}
	// 先按 level 裁掉超过目标层级的节点，再做响应模型转换。
	children = filterTreeNodesByLevel(children, level)
	data := convertTreeNodesToRes(children, level)
	if node.Type == model.ArtifactTreeNodeTypeGroup && node.TargetID.Valid && level == model.ArtifactNodeLevelAsset {
		return s.appendGroupAssets(ctx, node.TargetID.Int64, data)
	}
	return data, nil
}

// buildTreeChildren 处理“递归模式”的子树构建。
// 递归终止条件：
//  1. 当前已到 version 且 level < asset；
//  2. 当前没有更多 children；
//  3. level 过滤后不再允许继续向下。
//
// container 同样按 version → asset 下钻：asset 不在 tree 表里，必须走 listVersionAssetAsTree
// 直查 ops_artifact_assets。container layer 走 isolated（VersionID=0），不会被这里错列。
func (s *Service) buildTreeChildren(ctx context.Context, node *model.ArtifactTreeNode, level string) ([]*model.ArtifactTreeRes, error) {
	if node.Type == model.ArtifactTreeNodeTypeVersion && node.TargetID.Valid {
		if level == model.ArtifactNodeLevelAsset {
			return s.listVersionAssetAsTree(ctx, node)
		}
		return nil, nil
	}
	children, err := s.artStore.Nodes().ListByParentId(ctx, node.ID)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactTreeNodeListError)
	}
	children = filterTreeNodesByLevel(children, level)
	data := make([]*model.ArtifactTreeRes, 0, len(children))
	for _, child := range children {
		item := convertTreeNodeToRes(child, level)
		item.Children, err = s.buildTreeChildren(ctx, child, level)
		if err != nil {
			return nil, err
		}
		item.Leaf = len(item.Children) == 0
		data = append(data, item)
	}
	if node.Type == model.ArtifactTreeNodeTypeGroup && node.TargetID.Valid && level == model.ArtifactNodeLevelAsset {
		return s.appendGroupAssets(ctx, node.TargetID.Int64, data)
	}
	return data, nil
}

// listVersionAssetAsTree 将 version 节点下的资产转换成树节点返回。
// 参考项目里 version 在 asset 级别会直接展开成文件，这里保持同样语义。
func (s *Service) listVersionAssetAsTree(ctx context.Context, node *model.ArtifactTreeNode) ([]*model.ArtifactTreeRes, error) {
	assets, err := s.artStore.ListAssets(ctx, node.TargetID.Int64)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactAssetQueryError)
	}
	info := parseArtifactTreePath(node.Path)
	data := make([]*model.ArtifactTreeRes, 0, len(assets))
	for _, asset := range assets {
		base := filepath.Base(asset.Path)
		meta := &model.ArtifactTreeNodeMeta{Id: asset.Id, Type: model.ArtifactTreeNodeTypeAsset}
		if info != nil {
			m := *info
			m.Id, m.Type = asset.Id, model.ArtifactTreeNodeTypeAsset
			meta = &m
		}
		id := (&model.ArtifactTreeNodeId{Type: model.ArtifactTreeNodeTypeAsset, Pk: asset.Id}).String()
		meta.NodeID, meta.EntityID = id, id
		data = append(data, &model.ArtifactTreeRes{Name: base, Path: strings.TrimRight(node.Path, "/") + "/" + base, Leaf: true, Format: node.Format, Metadata: meta})
	}
	return data, nil
}

// appendGroupAssets 把 generic/group 下面直接挂载的资产追加到当前结果中。
// 这一步只在 level=asset 时执行；group/package/version 级别不会把文件混入返回。
func (s *Service) appendGroupAssets(ctx context.Context, groupID int64, data []*model.ArtifactTreeRes) ([]*model.ArtifactTreeRes, error) {
	g, err := s.artStore.Groups().GetEndpointById(ctx, groupID)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactGroupQueryError)
	}
	assets, err := s.artStore.ListGroupAssets(ctx, groupID)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactAssetQueryError)
	}
	for _, asset := range assets {
		base := filepath.Base(asset.Path)
		id := (&model.ArtifactTreeNodeId{Type: model.ArtifactTreeNodeTypeAsset, Pk: asset.Id}).String()
		data = append(data, &model.ArtifactTreeRes{Name: base, Path: "/" + g.GroupPath() + "/" + base, Leaf: true, Format: model.ArtifactFormat(g.Format), Metadata: &model.ArtifactTreeNodeMeta{Id: asset.Id, Type: model.ArtifactTreeNodeTypeAsset, Name: base, NodeID: id, EntityID: id}})
	}
	return data, nil
}

// filterTreeNodesByLevel 按 level 上限过滤节点类型。
// 例如 level=package 时，只保留 group/package，不返回 version/asset。
func filterTreeNodesByLevel(inputs []*model.ArtifactTreeNode, level string) []*model.ArtifactTreeNode {
	out := make([]*model.ArtifactTreeNode, 0, len(inputs))
	for _, n := range inputs {
		if treeNodeLevelRank(n.Type) <= treeLevelRank(level) {
			out = append(out, n)
		}
	}
	return out
}

func convertTreeNodesToRes(inputs []*model.ArtifactTreeNode, level string) []*model.ArtifactTreeRes {
	out := make([]*model.ArtifactTreeRes, 0, len(inputs))
	for _, n := range inputs {
		out = append(out, convertTreeNodeToRes(n, level))
	}
	return out
}

// convertTreeNodeToRes 将存储层树节点转换为接口返回模型。
// 这里只负责“单个节点”的基础映射；children 由上层递归逻辑填充。
func convertTreeNodeToRes(n *model.ArtifactTreeNode, level string) *model.ArtifactTreeRes {
	item := &model.ArtifactTreeRes{Name: n.Name, Path: n.Path, Format: n.Format, Leaf: n.Type == model.ArtifactTreeNodeTypeAsset}
	if n.Type == model.ArtifactTreeNodeTypeVersion {
		item.Leaf = level != model.ArtifactNodeLevelAsset || n.Format == model.ArtifactContainerFormat
	}
	if n.Type == model.ArtifactTreeNodeTypeVersion {
		item.Metadata = parseArtifactTreePath(n.Path)
		if item.Metadata == nil {
			item.Metadata = &model.ArtifactTreeNodeMeta{}
		}
		item.Metadata.Type = model.ArtifactTreeNodeTypeVersion
	} else {
		item.Metadata = &model.ArtifactTreeNodeMeta{Type: n.Type}
	}
	if n.TargetID.Valid {
		item.Metadata.Id = n.TargetID.Int64
		id := (&model.ArtifactTreeNodeId{Type: n.Type, Pk: n.TargetID.Int64}).String()
		item.Metadata.NodeID, item.Metadata.EntityID = id, id
	}
	if n.Type == model.ArtifactTreeNodeTypeFormat {
		item.Metadata.Name = n.Name
		if n.ArtifactID > 0 {
			id := (&model.ArtifactTreeNodeId{Type: model.ArtifactTreeNodeTypeArtifact, Pk: n.ArtifactID}).String()
			item.Metadata.NodeID, item.Metadata.EntityID = id, id
		}
	}
	return item
}

func normalizeTreePath(path string) (string, bool) {
	p := strings.TrimSpace(path)
	if p == "" {
		return model.ArtifactNodeRoot, false
	}
	recursive := strings.HasSuffix(p, "/**")
	p = strings.TrimSuffix(p, "/**")
	if p == "" {
		p = model.ArtifactNodeRoot
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if p != "/" {
		p = strings.TrimRight(p, "/")
	}
	return p, recursive
}

func normalizeTreeLevel(level string) (string, error) {
	if level == "" {
		return model.ArtifactNodeLevelAsset, nil
	}
	switch level {
	case model.ArtifactNodeLevelFormat, model.ArtifactNodeLevelGroup, model.ArtifactNodeLevelPackage, model.ArtifactNodeLevelVersion, model.ArtifactNodeLevelAsset:
		return level, nil
	default:
		return "", fmt.Errorf("unsupported level: %s", level)
	}
}

func treeLevelRank(level string) int {
	switch level {
	case model.ArtifactNodeLevelFormat:
		return 0
	case model.ArtifactNodeLevelGroup:
		return 1
	case model.ArtifactNodeLevelPackage:
		return 2
	case model.ArtifactNodeLevelVersion:
		return 3
	default:
		return 4
	}
}

func treeNodeLevelRank(t model.ArtifactTreeNodeType) int {
	switch t {
	case model.ArtifactTreeNodeTypeFormat:
		return 0
	case model.ArtifactTreeNodeTypeGroup:
		return 1
	case model.ArtifactTreeNodeTypePackage:
		return 2
	case model.ArtifactTreeNodeTypeVersion:
		return 3
	case model.ArtifactTreeNodeTypeAsset:
		return 4
	default:
		return 0
	}
}

func parseArtifactTreePath(path string) *model.ArtifactTreeNodeMeta {
	frames := strings.Split(strings.TrimLeft(path, "/"), "/")
	if len(frames) < 2 {
		return nil
	}
	meta := &model.ArtifactTreeNodeMeta{Name: frames[len(frames)-2], Version: frames[len(frames)-1]}
	if len(frames) > 2 {
		meta.Group = strings.Join(frames[:len(frames)-2], ".")
	}
	return meta
}

func (s *Service) statistic(ctx context.Context, report *model.ArtifactStatisticReport) (*model.ArtifactStatisticResponse, error) {
	res := &model.ArtifactStatisticResponse{}

	delData := &model.ArtifactGarbageReportRes{Ids: make([]int64, 0)}
	for _, d := range report.DeleteList {
		delData.Count++
		delData.Size += d.Size
		delData.Ids = append(delData.Ids, d.AssetID)
	}
	res.GarbageCollect = delData

	if len(report.TagList) == 0 {
		res.Capacity = make([]*model.ArtifactPackageCapacityRes, 0)
		return res, nil
	}

	sort.Slice(report.TagList, func(i, j int) bool {
		return report.TagList[i].VersionID < report.TagList[j].VersionID
	})

	vids := make([]int64, 0, len(report.TagList))
	for _, t := range report.TagList {
		vids = append(vids, t.VersionID)
	}

	verInfoList, err := s.artStore.Versions().Search(ctx, artifactStore.VersionIDsOption{IDs: vids})
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactStatisticCapacityError)
	}

	capByVersion := make(map[int64]*model.ArtifactVersionCapacityDesc, len(report.TagList))
	for _, item := range report.TagList {
		capByVersion[item.VersionID] = item
	}

	pkgCapMap := make(map[int64]*model.ArtifactPackageCapacityRes)
	for _, verInfo := range verInfoList {
		pkgCap, ok := pkgCapMap[verInfo.PackageID]
		if !ok {
			pkgCap = &model.ArtifactPackageCapacityRes{
				Scope:     artifactViewScope(verInfo),
				Space:     verInfo.SpaceName,
				Repo:      verInfo.RepoName,
				Name:      verInfo.PackageName,
				Namespace: verInfo.PackageNamespace,
				Format:    verInfo.PackageFormat,
				Versions:  make([]model.ArtifactVersionCapacityRes, 0),
			}
			pkgCapMap[verInfo.PackageID] = pkgCap
		}

		tagCap, ok := capByVersion[verInfo.ID]
		if !ok {
			continue
		}
		verCap := model.ArtifactVersionCapacityRes{
			Version:       verInfo.Version,
			Size:          tagCap.TotalSize,
			ExclusiveSize: tagCap.ExclusiveSize,
		}
		pkgCap.Versions = append(pkgCap.Versions, verCap)
		pkgCap.Size += tagCap.TotalSize
		pkgCap.ExclusiveSize += tagCap.ExclusiveSize
	}

	pkgCapList := make([]*model.ArtifactPackageCapacityRes, 0, len(pkgCapMap))
	for _, pkgCap := range pkgCapMap {
		sort.Slice(pkgCap.Versions, func(i, j int) bool {
			return pkgCap.Versions[i].ExclusiveSize > pkgCap.Versions[j].ExclusiveSize
		})
		pkgCapList = append(pkgCapList, pkgCap)
	}
	sort.Slice(pkgCapList, func(i, j int) bool {
		return pkgCapList[i].Size > pkgCapList[j].Size
	})

	res.Capacity = pkgCapList
	return res, nil
}

func artifactViewScope(info *model.ArtifactVersionInfo) string {
	if info.ViewRepoID > 0 {
		return "repo"
	}
	if info.ViewSpaceID > 0 {
		return "space"
	}
	return "global"
}

// ----------------- capacity ------------------
// StatisticContainerCapacity 容器仓库存储统计
func (s *Service) StatisticContainerCapacity(ctx context.Context) (*model.ArtifactStatisticResponse, error) {
	// 容量统计不需要过滤保留期，传 0 即可
	report, err := s.artStore.StatisticAssets(ctx, 0, artifactStore.ContainerFormatOption{})
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactStatisticCapacityError)
	}
	return s.statistic(ctx, report)
}

// StatisticCapacity 非容器仓库存储统计
func (s *Service) StatisticCapacity(ctx context.Context) (*model.ArtifactStatisticResponse, error) {
	// 容量统计不需要过滤保留期，传 0 即可
	report, err := s.artStore.StatisticAssets(ctx, 0, artifactStore.ExcludeContainerOption{})
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactStatisticCapacityError)
	}
	return s.statistic(ctx, report)
}

// ----------------- summary ------------------
// StatisticSummary 仓库统计概览
func (s *Service) StatisticSummary(ctx context.Context) (map[string]*model.ArtifactStatsObject, error) {
	assets, err := s.artStore.Assets().Search(ctx, artifactStore.AssetVersionedOption{})
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactStatisticSummaryError)
	}

	versionStats := make(map[int64]*model.ArtifactStatsObject)
	for _, asset := range assets {
		if asset.VersionID.Ptr() == nil {
			continue
		}
		stats, ok := versionStats[asset.VersionID.Int64]
		if !ok {
			stats = &model.ArtifactStatsObject{}
			versionStats[asset.VersionID.Int64] = stats
		}
		stats.Total++
		if asset.Deleted > 0 {
			stats.Deleted++
		} else {
			stats.Active++
		}
	}

	vIds := make([]int64, 0, len(versionStats))
	for vID := range versionStats {
		vIds = append(vIds, vID)
	}
	verInfoList, err := s.artStore.Versions().Search(ctx, artifactStore.VersionIDsOption{IDs: vIds})
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactStatisticSummaryError)
	}

	res := map[string]*model.ArtifactStatsObject{
		"all": {},
	}
	for _, verInfo := range verInfoList {
		verStats, ok := versionStats[verInfo.ID]
		if !ok {
			logger.GetLogger(ctx).Debugf("versionID %d not found in versionStats", verInfo.ID)
			continue
		}
		key := artifactScopeKey(verInfo)
		scopeStats, ok := res[key]
		if !ok {
			scopeStats = &model.ArtifactStatsObject{}
			res[key] = scopeStats
		}
		addStats(scopeStats, verStats)
		addStats(res["all"], verStats)
	}
	return res, nil
}

func artifactScopeKey(info *model.ArtifactVersionInfo) string {
	switch artifactViewScope(info) {
	case "repo":
		return "repo:" + info.RepoName
	case "space":
		return "space:" + info.SpaceName
	default:
		return "global"
	}
}

func addStats(dst, src *model.ArtifactStatsObject) {
	dst.Total += src.Total
	dst.Active += src.Active
	dst.Deleted += src.Deleted
}

func (s *Service) GarbageCollectSoftRemove(ctx context.Context) (*model.ArtifactGarbageReportRes, error) {
	// GC 场景：传入回收站保留期，确保只清理已过期的软删资源
	report, err := s.artStore.StatisticAssets(ctx, s.config.RecycleBinRetention, artifactStore.ExcludeContainerOption{})
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactGcExecuteError)
	}
	return s.garbageCollect(ctx, report.DeleteList)
}

func (s *Service) GarbageCollectContainer(ctx context.Context) (*model.ArtifactGarbageReportRes, error) {
	// GC 场景：传入回收站保留期，确保只清理已过期的软删资源
	report, err := s.artStore.StatisticAssets(ctx, s.config.RecycleBinRetention, artifactStore.ContainerFormatOption{})
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactGcExecuteError)
	}
	return s.garbageCollect(ctx, report.DeleteList)
}

func (s *Service) garbageCollect(ctx context.Context, items []*model.ArtifactRecycleBlobDesc) (*model.ArtifactGarbageReportRes, error) {
	res := &model.ArtifactGarbageReportRes{Ids: make([]int64, 0, len(items))}
	for _, item := range items {
		if item == nil || item.AssetID == 0 || item.BlobID == 0 {
			continue
		}
		if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := s.artStore.Assets().DeleteById(ctx, item.AssetID); err != nil {
				return err
			}
			if err := s.artStore.Blobs().DeleteById(ctx, item.BlobID); err != nil {
				return err
			}
			if item.BlobRef != "" {
				return s.fileStore.Delete(ctx, adapter.BlobPath(item.BlobRef))
			}
			return nil
		}); err != nil {
			logger.GetLogger(ctx).Errorf("artifact garbage collect failed for asset %d blob %d: %v", item.AssetID, item.BlobID, err)
			return res, err
		}
		res.Count++
		res.Size += item.Size
		res.Ids = append(res.Ids, item.AssetID)
	}
	return res, nil
}

func (s *Service) SoftRemove(ctx context.Context, spaceID int64, req *model.SoftRemoveEntitiesReq) (*model.ArtifactNodeRemoveReport, error) {
	ids := req.EntityIDs
	report := &model.ArtifactNodeRemoveReport{Total: len(ids), Data: make([]*model.ArtifactNodeRemoveRes, len(ids))}
	for _, eid := range ids {
		if _, err := entity.Parse(ctx, s.artStore, eid); err != nil {
			logger.GetLogger(ctx).Warnf("parse artifact entity %s failed before soft remove: %v", eid, err)
			if pkgErrors.GetI18nKey(err) != i18n.ErrorDefault {
				return nil, err
			}
			return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactSoftRemoveParseError)
		}
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for idx, eid := range ids {
			res := &model.ArtifactNodeRemoveRes{NodeID: eid, Status: model.ArtifactStatusOK}
			report.Data[idx] = res
			if err := s.softRemoveEntity(ctx, eid, res); err != nil {
				if res.Status == model.ArtifactStatusOK {
					res.Status = model.ArtifactStatusUnknown
				}
				return err
			}
		}
		return nil
	}); err != nil {
		// 内层已显式带上具体 i18n key（如制品库非空 / 不存在 / 查询失败）时保留之，
		// 仅对无 key 的未知错误统一包装为 ArtifactSoftRemoveError。
		if pkgErrors.GetI18nKey(err) != i18n.ErrorDefault {
			return nil, err
		}
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactSoftRemoveError)
	}

	for _, item := range report.Data {
		if item.Status == model.ArtifactStatusOK {
			report.Success++
		} else {
			report.Failed++
		}
	}
	return report, nil
}

func (s *Service) softRemoveEntity(ctx context.Context, eid string, res *model.ArtifactNodeRemoveRes) error {
	if strings.Count(eid, ".") != 1 {
		res.Status = model.ArtifactStatusInvalidId
		return nil
	}
	parts := strings.Split(eid, ".")
	nodeType := model.ArtifactTreeNodeType(parts[0])
	nodeID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		res.Status = model.ArtifactStatusInvalidId
		return nil
	}

	switch nodeType {
	case model.ArtifactTreeNodeTypeAsset:
		asset, err := s.artStore.Assets().GetById(ctx, nodeID)
		if err != nil {
			res.Status = model.ArtifactStatusNotFound
			return nil
		}
		return s.softRemoveAsset(ctx, asset, res)
	case model.ArtifactTreeNodeTypeVersion:
		ver, err := s.artStore.Versions().GetByID(ctx, nodeID)
		if err != nil {
			res.Status = model.ArtifactStatusNotFound
			return nil
		}
		pkg, err := s.artStore.Packages().GetByID(ctx, ver.PackageID)
		if err != nil {
			res.Status = model.ArtifactStatusNotFound
			return nil
		}
		return s.softRemoveVersion(ctx, pkg.ArtifactID, ver, res)
	case model.ArtifactTreeNodeTypePackage:
		pkg, err := s.artStore.Packages().GetByID(ctx, nodeID)
		if err != nil {
			res.Status = model.ArtifactStatusNotFound
			return nil
		}
		return s.softRemovePackage(ctx, pkg, res)
	case model.ArtifactTreeNodeTypeGroup:
		group, err := s.artStore.Groups().GetById(ctx, nodeID)
		if err != nil {
			res.Status = model.ArtifactStatusNotFound
			return nil
		}
		return s.softRemoveGroup(ctx, group, res)
	case model.ArtifactTreeNodeTypeArtifact:
		view, err := s.artStore.Views().GetByID(ctx, nodeID)
		if err != nil {
			res.Status = model.ArtifactStatusNotFound
			return pkgErrors.WithI18nKey(err, i18n.ArtifactSoftRemoveArtifactNotFound)
		}
		return s.softRemoveArtifact(ctx, view, res)
	default:
		res.Status = model.ArtifactStatusInvalidId
		return nil
	}
}

func (s *Service) softRemoveAsset(ctx context.Context, asset *model.ArtifactAsset, res *model.ArtifactNodeRemoveRes) error {
	if err := s.artStore.Assets().SoftDeleteById(ctx, asset.ID); err != nil {
		return err
	}
	if asset.BlobID != 0 {
		if err := s.artStore.Blobs().SoftDeleteById(ctx, asset.BlobID); err != nil {
			return err
		}
	}
	res.Assets++
	return nil
}

func (s *Service) softRemoveVersion(ctx context.Context, artifactID int64, ver *model.ArtifactVersion, res *model.ArtifactNodeRemoveRes) error {
	assets, err := s.artStore.Assets().Search(ctx, model.SearchAssetOption{VersionID: ver.ID}, artifactStore.AssetExcludeDeletedOption{})
	if err != nil {
		return err
	}
	for _, asset := range assets {
		if err = s.softRemoveAsset(ctx, asset, res); err != nil {
			return err
		}
	}
	if err = s.artStore.Versions().SoftDelete(ctx, ver); err != nil {
		return err
	}
	res.Versions++
	_ = s.deleteTreeNodeByTarget(ctx, artifactID, model.ArtifactTreeNodeTypeVersion, ver.ID)
	return nil
}

func (s *Service) softRemovePackage(ctx context.Context, pkg *model.ArtifactPackage, res *model.ArtifactNodeRemoveRes) error {
	versions, err := s.artStore.Versions().Find(ctx, model.SearchVersionOption{PackageID: pkg.ID})
	if err != nil {
		return err
	}
	for _, ver := range versions {
		if err = s.softRemoveVersion(ctx, pkg.ArtifactID, ver, res); err != nil {
			return err
		}
	}
	if err = s.artStore.Packages().SoftDelete(ctx, pkg); err != nil {
		return err
	}
	res.Packages++
	_ = s.deleteTreeNodeByTarget(ctx, pkg.ArtifactID, model.ArtifactTreeNodeTypePackage, pkg.ID)
	if pkg.Format == model.ArtifactHelmFormat {
		_ = helm.NewHelmIndex(s.artStore, &adapter.ViewDescriptor{ArtifactID: pkg.ArtifactID, Store: s.fileStore, StorageID: 1}).UpdateRepo(ctx, "/artifacts/helm")
	}
	return nil
}

// softRemoveGroup 软删指定分组。
// 判空口径：分组自身 + 其所有未软删后代 group 下的活 asset 数为 0 方允许删除。
//   - 允许空的子目录（子 group 无活 asset）：一并级联软删；
//   - 任一层级存在活 asset 即拒绝，返回 ArtifactSoftRemoveGroupNotEmpty 触发事务回滚（路 B）；
//   - 顺带软删子 group 链路上的空壳 package / version（无活 asset 的上传半成品）；
//   - 每次软删 group / package / version 都同步清对应 tree_node，避免前端刷树残留。
func (s *Service) softRemoveGroup(ctx context.Context, group *model.ArtifactGroup, res *model.ArtifactNodeRemoveRes) error {
	// 1) 收集当前 group + 所有未软删子 group，作为判空与级联软删的目标集合。
	//    Groups().ListDescendants 返回不含自身，吾在头部追加当前 group 统一处理。
	descendants, err := s.artStore.Groups().ListDescendants(ctx, group.ID)
	if err != nil {
		res.Status = model.ArtifactStatusUnknown
		return pkgErrors.WithI18nKey(err, i18n.ArtifactSoftRemoveGroupQueryError)
	}
	targets := make([]*model.ArtifactGroup, 0, len(descendants)+1)
	targets = append(targets, group)
	targets = append(targets, descendants...)

	// 2) 非空校验：集合内任一 group 有活 asset（含 groupID 直挂 + package/version 链路下挂）即拒绝。
	for _, g := range targets {
		// 2.1 直挂在 group 上的活 asset（generic 场景）。
		directAssets, aErr := s.artStore.Assets().Search(ctx,
			groupAssetsOption{groupID: g.ID}, artifactStore.AssetExcludeDeletedOption{})
		if aErr != nil {
			res.Status = model.ArtifactStatusUnknown
			return pkgErrors.WithI18nKey(aErr, i18n.ArtifactSoftRemoveGroupQueryError)
		}
		if len(directAssets) > 0 {
			res.Status = model.ArtifactStatusNotEmpty
			return pkgErrors.WithI18nKey(
				fmt.Errorf("group %d still has %d active asset(s)", g.ID, len(directAssets)),
				i18n.ArtifactSoftRemoveGroupNotEmpty,
			)
		}
		// 2.2 经 package/version 链路下挂的活 asset（raw / maven 等版本化场景）。
		pkgs, pErr := s.artStore.Packages().Search(ctx, groupPackagesOption{groupID: g.ID})
		if pErr != nil {
			res.Status = model.ArtifactStatusUnknown
			return pkgErrors.WithI18nKey(pErr, i18n.ArtifactSoftRemoveGroupQueryError)
		}
		for _, pkg := range pkgs {
			versions, vErr := s.artStore.Versions().Find(ctx, model.SearchVersionOption{PackageID: pkg.ID})
			if vErr != nil {
				res.Status = model.ArtifactStatusUnknown
				return pkgErrors.WithI18nKey(vErr, i18n.ArtifactSoftRemoveGroupQueryError)
			}
			for _, ver := range versions {
				verAssets, vaErr := s.artStore.Assets().Search(ctx,
					model.SearchAssetOption{VersionID: ver.ID}, artifactStore.AssetExcludeDeletedOption{})
				if vaErr != nil {
					res.Status = model.ArtifactStatusUnknown
					return pkgErrors.WithI18nKey(vaErr, i18n.ArtifactSoftRemoveGroupQueryError)
				}
				if len(verAssets) > 0 {
					res.Status = model.ArtifactStatusNotEmpty
					return pkgErrors.WithI18nKey(
						fmt.Errorf("group %d still has active asset(s) under version %d", g.ID, ver.ID),
						i18n.ArtifactSoftRemoveGroupNotEmpty,
					)
				}
			}
		}
	}

	// 3) 空目录树：级联软删 targets 内的空壳 package/version 与 group 本身，并同步清 tree_node。
	//    由深到浅处理，保证子级先落地（Groups().ListDescendants 按 parentID 升序，反转后即从叶子起）。
	for i := len(targets) - 1; i >= 0; i-- {
		g := targets[i]
		pkgs, pErr := s.artStore.Packages().Search(ctx, groupPackagesOption{groupID: g.ID})
		if pErr != nil {
			return pErr
		}
		for _, pkg := range pkgs {
			versions, vErr := s.artStore.Versions().Find(ctx, model.SearchVersionOption{PackageID: pkg.ID})
			if vErr != nil {
				return vErr
			}
			for _, ver := range versions {
				if err = s.artStore.Versions().SoftDelete(ctx, ver); err != nil {
					return err
				}
				res.Versions++
				_ = s.deleteTreeNodeByTarget(ctx, g.ArtifactID, model.ArtifactTreeNodeTypeVersion, ver.ID)
			}
			if err = s.artStore.Packages().SoftDelete(ctx, pkg); err != nil {
				return err
			}
			res.Packages++
			_ = s.deleteTreeNodeByTarget(ctx, g.ArtifactID, model.ArtifactTreeNodeTypePackage, pkg.ID)
		}
		if err = s.artStore.Groups().SoftDeleteById(ctx, g.ID); err != nil {
			return err
		}
		_ = s.deleteTreeNodeByTarget(ctx, g.ArtifactID, model.ArtifactTreeNodeTypeGroup, g.ID)
	}
	// 4) 若该 format 下已无活 group，清理 format 节点（避免前端看到空的格式入口）
	remaining, _ := s.artStore.Groups().ListDescendantsForOwner(ctx, group.ArtifactID, group.Format)
	if len(remaining) == 0 {
		_ = s.deleteFormatTreeNode(ctx, group.ArtifactID, group.Format)
	}
	return nil
}

// deleteFormatTreeNode 删除指定 artifact 下某 format 的根节点（path="/"）。
// 当该 format 下已无活 group 时调用，避免前端看到空的格式入口。
func (s *Service) deleteFormatTreeNode(ctx context.Context, artifactID int64, format model.ArtifactFormat) error {
	node, err := s.artStore.Nodes().GetByPath(ctx, artifactID, "/", format)
	if err != nil {
		return nil
	}
	return s.artStore.Nodes().DeleteById(ctx, node.ID)
}

// softRemoveArtifact 软删整个制品库（view）。
// 判空口径：只看活 asset（AssetViewOption + AssetExcludeDeletedOption）。
//   - 允许库下存在空分组（group 不纳入判空）；
//   - 允许存在未挂 asset 的活 package / 活 version（上传中途失败留下的空壳，不阻塞删除）；
//   - 已软删的 asset 视为可忽略（由后续 GC 物理清理）。
//
// 一旦发现活 asset，立即返回 ArtifactSoftRemoveArtifactNotEmpty，
// 触发外层事务回滚（路 B 语义）：整批请求失败，不会部分成功。
//
// 清理链路：软删 package / version / group 的同时同步调 deleteTreeNodeByTarget 清对应
// ops_artifact_tree_nodes 记录，避免前端刷树残留节点。view 自身的 tree_node 无对应
// 目标实体，待 GC 阶段 hardDeleteViewResources 末尾统一清理。
func (s *Service) softRemoveArtifact(ctx context.Context, view *model.ArtifactView, res *model.ArtifactNodeRemoveRes) error {
	// 1) 非空校验：只看活 asset。活 package/version 但无 asset 的空壳允许删。
	activeAssets, err := s.artStore.Assets().Search(ctx,
		artifactStore.AssetViewOption{ArtifactID: view.ID},
		artifactStore.AssetExcludeDeletedOption{},
	)
	if err != nil {
		res.Status = model.ArtifactStatusUnknown
		return pkgErrors.WithI18nKey(err, i18n.ArtifactSoftRemoveArtifactQueryError)
	}
	if len(activeAssets) > 0 {
		res.Status = model.ArtifactStatusNotEmpty
		return pkgErrors.WithI18nKey(
			fmt.Errorf("artifact view %d still has %d active asset(s)", view.ID, len(activeAssets)),
			i18n.ArtifactSoftRemoveArtifactNotEmpty,
		)
	}

	// 2) 空库：顺带软删孤立 package / version（上传半成品），再软删 group 与 view 自身。
	packages, err := s.artStore.Packages().List(ctx, view.ID, false)
	if err != nil {
		res.Status = model.ArtifactStatusUnknown
		return pkgErrors.WithI18nKey(err, i18n.ArtifactSoftRemoveArtifactQueryError)
	}
	for _, pkg := range packages {
		versions, vErr := s.artStore.Versions().Find(ctx, model.SearchVersionOption{PackageID: pkg.ID})
		if vErr != nil {
			return vErr
		}
		for _, ver := range versions {
			if err = s.artStore.Versions().SoftDelete(ctx, ver); err != nil {
				return err
			}
			res.Versions++
			_ = s.deleteTreeNodeByTarget(ctx, view.ID, model.ArtifactTreeNodeTypeVersion, ver.ID)
		}
		if err = s.artStore.Packages().SoftDelete(ctx, pkg); err != nil {
			return err
		}
		res.Packages++
		_ = s.deleteTreeNodeByTarget(ctx, view.ID, model.ArtifactTreeNodeTypePackage, pkg.ID)
	}

	// view 与其 group 绑同一 format，直接按 view.Format 递归取全部未删 group。
	groups, err := s.artStore.Groups().ListDescendantsForOwner(ctx, view.ID, view.Format)
	if err != nil {
		res.Status = model.ArtifactStatusUnknown
		return pkgErrors.WithI18nKey(err, i18n.ArtifactSoftRemoveArtifactQueryError)
	}
	for _, g := range groups {
		if err = s.artStore.Groups().SoftDeleteById(ctx, g.ID); err != nil {
			return err
		}
		_ = s.deleteTreeNodeByTarget(ctx, view.ID, model.ArtifactTreeNodeTypeGroup, g.ID)
	}

	if err = s.artStore.Views().SoftDelete(ctx, view); err != nil {
		return err
	}
	return nil
}

func (s *Service) deleteTreeNodeByTarget(ctx context.Context, artifactID int64, nodeType model.ArtifactTreeNodeType, id int64) error {
	node, err := s.artStore.Nodes().GetByTarget(ctx, artifactID, nodeType, id)
	if err != nil {
		return nil
	}
	return s.artStore.Nodes().DeleteById(ctx, node.ID)
}

type groupPackagesOption struct{ groupID int64 }

func (o groupPackagesOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("groupID = ?", o.groupID).Where("deleted = 0")
}

type groupAssetsOption struct{ groupID int64 }

func (o groupAssetsOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("groupID = ?", o.groupID)
}

type ArchiveGroup struct {
	base      *archiveGroupItem
	subGroups []*archiveGroupItem
}

func (a *ArchiveGroup) Name() string {
	return fmt.Sprintf("%s.zip", a.base.Name)
}

type archiveGroupItem struct {
	*model.ArtifactGroup
	archivePath string
	entity      *model.ArtifactEntity
}

func (s *Service) PrepareArchiveGroup(ctx context.Context, session *authModel.Session, spaceID, repoID, groupID int64) (*ArchiveGroup, error) {
	root, err := s.artStore.Groups().GetById(ctx, groupID)
	if err != nil {
		return nil, err
	}
	rootEntity, err := entity.Parse(ctx, s.artStore, (&model.ArtifactTreeNodeId{Type: model.ArtifactTreeNodeTypeGroup, Pk: root.ID}).String())
	if err != nil {
		return nil, err
	}
	if err = s.authorizeArtifactAction(ctx, session, spaceID, repoID, rootEntity, enum.CasbinActionArtifactView); err != nil {
		return nil, err
	}

	descendants, err := s.artStore.Groups().ListDescendants(ctx, groupID)
	if err != nil {
		return nil, err
	}
	rootGroup := &archiveGroupItem{ArtifactGroup: root, archivePath: root.Name, entity: rootEntity}
	parentMap := map[int64]*archiveGroupItem{groupID: rootGroup}
	archive := &ArchiveGroup{base: rootGroup}
	denied := make(map[int64]bool)
	for _, d := range descendants {
		parent := parentMap[d.ParentID.Int64]
		if parent == nil {
			logger.GetLogger(ctx).Errorf("parent group %d does not exist in archive parent map", d.ParentID.Int64)
			continue
		}
		entityNode := parent.entity.Join(&model.ArtifactEntity{Kind: model.ArtifactEntityKindGroup, Id: d.ID, Name: d.Name})
		item := &archiveGroupItem{ArtifactGroup: d, archivePath: filepath.Join(parent.archivePath, d.Name), entity: entityNode}
		parentMap[d.ID] = item
		if denied[d.ParentID.Int64] {
			denied[d.ID] = true
			continue
		}
		if err = s.authorizeArtifactAction(ctx, session, spaceID, repoID, entityNode, enum.CasbinActionArtifactPull); err != nil {
			denied[d.ID] = true
			logger.GetLogger(ctx).Infof("group %d denied by artifact access control: %v", d.ID, err)
			continue
		}
		archive.subGroups = append(archive.subGroups, item)
	}
	return archive, nil
}

func (s *Service) authorizeArtifactAction(ctx context.Context, session *authModel.Session, spaceID, repoID int64, e *model.ArtifactEntity, action enum.CasbinAction) error {
	if session == nil || session.User == nil {
		return pkgErrors.WithI18nKey(
			errors.New("authentication required"),
			i18n.ArtifactAuthRequired,
		)
	}

	sub := rbac.NewUser(session.User.Account)
	obj := rbac.NewObject(e.PolicyLocation())

	// 三级作用域查找：repo → space → global
	// 某一级有策略则覆盖（不回落），无策略则向上继承
	if repoID > 0 {
		repoScope := rbac.NewRepoSpace(spaceID, repoID)
		if s.perm.HasPoliciesForScope(repoScope) {
			_, _, err := s.perm.Enforce(ctx, sub, repoScope, obj, action)
			return err
		}
	}
	if spaceID > 0 {
		spaceScope := rbac.NewSpace(spaceID)
		if s.perm.HasPoliciesForScope(spaceScope) {
			_, _, err := s.perm.Enforce(ctx, sub, spaceScope, obj, action)
			return err
		}
	}
	globalScope := rbac.NewSpace(0)
	_, _, err := s.perm.Enforce(ctx, sub, globalScope, obj, action)
	return err
}

func (s *Service) ArchiveGroup(ctx context.Context, archive *ArchiveGroup, output io.Writer) error {
	zipWriter := zip.NewWriter(output)
	zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})
	defer zipWriter.Close()

	for _, g := range append(archive.subGroups, archive.base) {
		assets, err := s.artStore.ListGroupAssets(ctx, g.ID)
		if err != nil {
			return err
		}
		for _, asset := range assets {
			info, err := s.fileStore.Stat(ctx, adapter.BlobPath(asset.BlobRef))
			if err != nil {
				return err
			}
			zf := zipAssetFile{asset: asset, info: info}
			header, err := zip.FileInfoHeader(zf)
			if err != nil {
				return err
			}
			header.Name = filepath.Join(g.archivePath, asset.Path)
			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return err
			}
			reader, err := s.fileStore.Open(ctx, adapter.BlobPath(asset.BlobRef))
			if err != nil {
				return err
			}
			if _, err = io.Copy(writer, reader); err != nil {
				_ = reader.Close()
				return err
			}
			_ = reader.Close()
		}
	}
	return zipWriter.Flush()
}

type zipAssetFile struct {
	asset *model.ArtifactAssetsRes
	info  storagedriver.FileInfo
}

func (z zipAssetFile) Name() string       { return z.asset.Path }
func (z zipAssetFile) Size() int64        { return z.info.Size() }
func (z zipAssetFile) Mode() fs.FileMode  { return 0644 }
func (z zipAssetFile) ModTime() time.Time { return z.info.ModTime() }
func (z zipAssetFile) IsDir() bool        { return z.info.IsDir() }
func (z zipAssetFile) Sys() any           { return nil }

func (s *Service) ListContainerImages(ctx context.Context, view *adapter.ViewDescriptor, packageName, version string) (*container.TagMetadata, error) {
	ver, err := s.artStore.GetVersion(ctx, view.ArtifactID, packageName, "", version, model.ArtifactContainerFormat)
	if err != nil {
		return nil, err
	}
	assets, err := s.artStore.Assets().Search(ctx,
		model.SearchAssetOption{VersionID: ver.ID, Kind: model.AssetKindMain},
		artifactStore.AssetExcludeDeletedOption{},
	)
	if err != nil {
		return nil, err
	}
	if len(assets) == 0 {
		return nil, pkgErrors.WithI18nKey(
			errors.New("container tag asset not found"),
			i18n.ArtifactContainerTagNotFound,
		)
	}
	dgst, err := digest.Parse(assets[0].Path)
	if err != nil {
		return nil, err
	}
	reader := container.NewTagMetadataReader(s.artStore, view)
	return reader.Parse(ctx, assets[0].ContentType, dgst)
}

// RefreshTree 全量重建制品树节点。
// 遍历顺序：groups → packages(含 raw namespace 补建 group) → versions。
// 移植自原项目 /root/repo/gitfox/app/services/artifact/sync_tree_node.go:RefreshTreeNodes。
//
// format 含义：
//   - model.ArtifactAllFormat（"all"）或空值：全量刷新所有格式
//   - 其它具体 format：仅刷新该 format 下的 groups / packages / versions
func (s *Service) RefreshTree(ctx context.Context, format string) error {
	syncer := artModel.NewSyncer(s.artStore)

	targetFormat := model.ArtifactFormat(strings.TrimSpace(format))
	refreshAll := targetFormat == "" || targetFormat == model.ArtifactAllFormat

	// 1. 同步所有 group 树节点
	groups, err := s.artStore.Groups().All(ctx)
	if err != nil {
		logger.GetLogger(ctx).Errorf("refresh tree: list groups failed: %v", err)
		return err
	}
	for _, group := range groups {
		if !refreshAll && group.Format != targetFormat {
			continue
		}
		if _, gErr := syncer.SyncTreeNodeForGroup(ctx, group); gErr != nil {
			logger.GetLogger(ctx).Errorf("refresh tree: sync group %d failed: %v", group.ID, gErr)
			return gErr
		}
	}

	// 2. 遍历所有 packages，同步 package + version 树节点
	return s.artStore.Packages().ListBatch(ctx, false, 100, func(pkg *model.ArtifactPackage) error {
		if !refreshAll && pkg.Format != targetFormat {
			return nil
		}

		// 2a. raw 格式 namespace 补建 group（namespace != "" 且 groupID==0 时）
		if pkg.Namespace != "" && pkg.GroupID == 0 {
			if sErr := syncer.SyncRawGroup(ctx, pkg); sErr != nil {
				return sErr
			}
		}

		// 2b. 同步 package 节点
		if _, pErr := syncer.SyncTreeNodeForPackage(ctx, pkg); pErr != nil {
			return pErr
		}

		// 2c. 同步该 package 下所有 version 节点
		dbVersions, vErr := s.artStore.Versions().Find(ctx, model.SearchVersionOption{PackageID: pkg.ID})
		if vErr != nil {
			return vErr
		}
		for _, ver := range dbVersions {
			if _, tErr := syncer.SyncTreeNodeForVersion(ctx, ver, pkg); tErr != nil {
				return tErr
			}
		}
		return nil
	})
}

// GarbageCollectExpiredViews 清理超过回收站保留期的已软删视图及其全部关联资源。
//
// 正确的处理顺序（参考 gitfox 「先清理子资源、最后才删父记录」模式）：
//  1. 根据 config.RecycleBinRetention 计算过期截止时间
//  2. 【仅查询】找出所有 deleted=1 且 editedDate < 截止时间 的 view ID（不立即删除）
//  3. 对每个过期 view，级联硬删其下的 packages -> versions -> assets -> blobs（含物理文件）
//  4. 关联资源清理成功后，才物理删除 view 记录本身
//
// 此顺序保证幂等性：若进程在步骤 3/4 之间崩溃，view 记录依然存在（deleted=1），
// 下次 GC 运行时会重新发现并继续处理，不会产生永久孤儿资源。
//
// 此方法由定时任务 JobTypeArtifactExpiredViewGC 自动调用。
func (s *Service) GarbageCollectExpiredViews(ctx context.Context) (int, error) {
	log := logger.GetLogger(ctx)

	// 计算回收站保留期截止时间：editedDate 早于此时间的已软删 view 将被清理
	cutoff := time.Now().Add(-s.config.RecycleBinRetention)
	log.Infof("开始清理过期已软删视图，截止时间: %s", cutoff.Format(time.RFC3339))

	// 【仅查询】找出过期 view ID，不删除——保证幂等性
	expiredArtifactIDs, err := s.artStore.Views().FindExpiredIDs(ctx, cutoff)
	if err != nil {
		return 0, pkgErrors.WithI18nKey(fmt.Errorf("query expired views failed: %w", err), i18n.ArtifactViewExpiredQueryError)
	}
	if len(expiredArtifactIDs) == 0 {
		log.Info("无过期已软删视图需要清理")
		return 0, nil
	}

	log.Infof("发现 %d 个过期已软删视图待清理: %v", len(expiredArtifactIDs), expiredArtifactIDs)

	cleaned := 0
	for _, artifactID := range expiredArtifactIDs {
		// Step 1: 先清理关联资源（packages/versions/assets/blobs/物理文件/groups/tree_nodes/meta_assets）
		if resErr := s.hardDeleteViewResources(ctx, artifactID); resErr != nil {
			// 关联资源清理失败时，保留 view 记录，下次 GC 重试
			log.Errorf("清理视图 %d 的关联资源失败，跳过本次删除（下次 GC 将重试）: %v", artifactID, resErr)
			continue
		}

		// Step 2: 关联资源清理成功后，最后才物理删除 view 记录
		if delErr := s.artStore.Views().Delete(ctx, artifactID); delErr != nil {
			log.Errorf("硬删视图 %d 记录失败（关联资源已清理，下次 GC 将重试删除 view 记录）: %v", artifactID, delErr)
			// 此处不 continue：关联资源已清理，即使 view 记录删除失败，
			// 下次 GC 也只剩删除空 view 记录，不会重复删资源（资源已不存在）
			continue
		}

		cleaned++
		log.Infof("视图 %d 清理完成", artifactID)
	}

	log.Infof("过期视图清理完成，共成功处理 %d / %d 个视图", cleaned, len(expiredArtifactIDs))
	return cleaned, nil
}

// hardDeleteViewResources 级联硬删指定 view 下的所有关联资源。
// 删除顺序：
//  1. packages (含 versions -> version assets -> blobs+物理文件)
//  2. group 级 assets (generic 格式挂载在 group 上的 assets -> blobs+物理文件)
//  3. groups
//  4. tree_nodes
//  5. meta_assets (helm chart 索引等)
func (s *Service) hardDeleteViewResources(ctx context.Context, artifactID int64) error {
	log := logger.GetLogger(ctx)

	// ---- 1. 清理 packages -> versions -> version-assets -> blobs ----
	packages, err := s.artStore.Packages().List(ctx, artifactID, true)
	if err != nil {
		return fmt.Errorf("查询 view %d 的 packages 失败: %w", artifactID, err)
	}

	for _, pkg := range packages {
		versions, vErr := s.artStore.Versions().Find(ctx, model.SearchVersionOption{
			PackageID:      pkg.ID,
			IncludeDeleted: true,
		})
		if vErr != nil {
			log.Errorf("查询 package %d 的 versions 失败: %v", pkg.ID, vErr)
			continue
		}

		for _, ver := range versions {
			if err = s.hardDeleteVersionAssets(ctx, ver.ID); err != nil {
				log.Errorf("清理 version %d 的 assets 失败: %v", ver.ID, err)
			}
			if _, err = s.artStore.Versions().DeleteByIds(ctx, ver.ID); err != nil {
				log.Errorf("硬删 version %d 失败: %v", ver.ID, err)
			}
		}

		if err = s.artStore.Packages().DeleteById(ctx, pkg.ID); err != nil {
			log.Errorf("硬删 package %d 失败: %v", pkg.ID, err)
		}
	}

	// ---- 2. 清理 view 下剩余 assets（group 级 generic assets 等）及其 blobs+物理文件 ----
	// version-linked assets 已在上面清理，这里处理 groupID>0 且 versionID=0 的 assets
	remainingAssets, err := s.artStore.Assets().SearchExtendBlob(ctx, artifactStore.AssetViewOption{ArtifactID: artifactID})
	if err != nil {
		log.Errorf("查询 view %d 的剩余 assets 失败: %v", artifactID, err)
	} else {
		for _, asset := range remainingAssets {
			if asset.BlobID != 0 && asset.Ref != "" {
				_ = s.fileStore.Delete(ctx, adapter.BlobPath(asset.Ref))
				_ = s.artStore.Blobs().DeleteById(ctx, asset.BlobID)
			}
			_ = s.artStore.Assets().DeleteById(ctx, asset.ID)
		}
	}

	// ---- 3. 清理 groups ----
	if n, gErr := s.artStore.Groups().DeleteByArtifactID(ctx, artifactID); gErr != nil {
		log.Errorf("硬删 view %d 的 groups 失败: %v", artifactID, gErr)
	} else if n > 0 {
		log.Infof("硬删 view %d 的 %d 个 groups", artifactID, n)
	}

	// ---- 4. 清理 tree_nodes ----
	if n, tErr := s.artStore.Nodes().DeleteByArtifactID(ctx, artifactID); tErr != nil {
		log.Errorf("硬删 view %d 的 tree_nodes 失败: %v", artifactID, tErr)
	} else if n > 0 {
		log.Infof("硬删 view %d 的 %d 个 tree_nodes", artifactID, n)
	}

	// ---- 5. 清理 meta_assets ----
	if n, mErr := s.artStore.MetaAssets().DeleteByArtifactID(ctx, artifactID); mErr != nil {
		log.Errorf("硬删 view %d 的 meta_assets 失败: %v", artifactID, mErr)
	} else if n > 0 {
		log.Infof("硬删 view %d 的 %d 个 meta_assets", artifactID, n)
	}

	log.Infof("视图 %d 关联资源清理完成: %d 个 packages", artifactID, len(packages))
	return nil
}

// hardDeleteVersionAssets 硬删指定 version 下的所有 assets 及关联 blobs，并清理物理文件。
func (s *Service) hardDeleteVersionAssets(ctx context.Context, versionID int64) error {
	assets, err := s.artStore.Assets().Search(ctx, model.SearchAssetOption{VersionID: versionID})
	if err != nil {
		return err
	}

	for _, asset := range assets {
		// 删除关联的 blob 及物理文件
		if asset.BlobID != 0 {
			blob, bErr := s.artStore.Blobs().GetById(ctx, asset.BlobID)
			if bErr == nil && blob.Ref != "" {
				_ = s.fileStore.Delete(ctx, adapter.BlobPath(blob.Ref))
			}
			_ = s.artStore.Blobs().DeleteById(ctx, asset.BlobID)
		}
		// 硬删 asset
		_ = s.artStore.Assets().DeleteById(ctx, asset.ID)
	}
	return nil
}
