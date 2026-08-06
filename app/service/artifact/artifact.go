package artifact

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	request "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter/container"
	artModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/model"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage"
	storagedriver "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage/driver"
	urlProviderPkg "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"

	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	spaceStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/space"
	userStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"github.com/opencontainers/go-digest"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gorm.io/gorm"
)

const artifactStoragePrefix = "artifact"

const (
	JobTypeArtifactGC               = "gitfox:cleanup:artifact-garbage-collector"
	ArtifactGCJobMaxDurationSeconds = 3600
	// JobUIDArtifactGC 是制品库 GC 定时任务的唯一标识，server 启动时上 upsert 注册。
	JobUIDArtifactGC = "gitfox:recurring:artifact-gc"
	// JobCronArtifactGC 控制制品库 GC 触发时刻：北京时间每日 02:00 一次。
	// scheduler.AddRecurringWithData 在 datetime.CST 下解析 cron，因此该表达式按 CST 计算。
	JobCronArtifactGC = "0 2 * * *"
)

var _ Interface = (*Service)(nil)

// Service 是制品库的核心业务服务。
type Service struct {
	db          *gorm.DB
	artStore    artifactStore.Interface
	spaceStore  spaceStore.Interface
	repoStore   repoStore.Interface
	userStore   *userStore.Store
	fileStore   storage.ContentStorage
	scheduler   JobScheduler
	config      *config.ArtifactConfig // 制品库配置（回收站保留期、GC 间隔等）
	urlProvider urlProviderPkg.Provider
}

// NewService 创建制品库服务。使用 StorageDriver 内部构建 ContentStorage，
// 避免与 pipeline 的 ContentStorage provider 冲突。
func NewService(
	db *gorm.DB,
	artStore artifactStore.Interface,
	spaceStore spaceStore.Interface,
	repoStore repoStore.Interface,
	uStore *userStore.Store,
	d storagedriver.StorageDriver,
	scheduler JobScheduler,
	cfg *config.ArtifactConfig,
	urlProvider urlProviderPkg.Provider,
) (*Service, error) {
	fs, err := storage.NewCommonContentStore(context.Background(), d,
		storage.WithPrefix(artifactStoragePrefix))
	if err != nil {
		return nil, err
	}
	// 校验并修正配置：确保 GCInterval >= RecycleBinRetention
	cfg.Sanitize()
	return &Service{
		db:          db,
		artStore:    artStore,
		spaceStore:  spaceStore,
		repoStore:   repoStore,
		userStore:   uStore,
		fileStore:   fs,
		scheduler:   scheduler,
		config:      cfg,
		urlProvider: urlProvider,
	}, nil
}

// FileStore 返回制品库的 ContentStorage，用于外部组件（如容器镜像控制器）直接访问存储。
func (s *Service) FileStore() storage.ContentStorage {
	return s.fileStore
}

// CreateLib 创建制品库视图。
// 默认分组规则：创建制品库时不提前插入 group；各格式由上传协议按需创建或选择分组。
func (s *Service) CreateLib(ctx context.Context, req *model.CreateArtifactLibReq) (*model.ArtifactView, error) {
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
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactLibQueryError)
	}

	if len(existing) != 0 {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactLibNameDuplicate)
	}

	viewType := resolveArtifactViewType(req.SpaceID, req.RepoID)

	session, _ := request.GetAuthSession(ctx)
	// 情况4：无冲突，正常创建
	v := &model.CreateArtifactView{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		SpaceID:     req.SpaceID,
		RepoID:      req.RepoID,
		Format:      artifact.ArtifactFormat(req.Format),
		Type:        viewType,
		CreatedBy:   session.User.Account,
		EditedBy:    session.User.Account,
	}
	view, err := s.artStore.Views().Create(ctx, v)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactLibCreateError)
	}

	// file 类型：为新视图铸造一个与视图同名的根目录，便于上传时直接挂载。
	if artifact.ArtifactFormat(req.Format) == artifact.ArtifactFileFormat {
		if gErr := s.ensureDefaultGroupForView(ctx, view); gErr != nil {
			logger.GetLogger(ctx).Warnf("create default group for view %d (%s) failed: %v", view.ID, view.Name, gErr)
		}
	}

	return view, nil
}

// CreateGroup 创建分组
func (s *Service) CreateGroup(ctx context.Context, artifactID int64, names []string, format model.ArtifactFormat) (*model.ArtifactGroupEndpoint, error) {
	var result *model.ArtifactGroupEndpoint
	session, _ := req.GetAuthSession(ctx)
	err := s.db.Transaction(func(tx *gorm.DB) error {
		ge, e := s.artStore.Groups().RecursiveCreate(ctx, artifactID, names, format, session.User.Account)
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

// resolveArtifactViewType 解析制品库视图类型。
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

// ensureDefaultGroupForView 为指定视图铸造一个与视图同名的根分组，并同步 tree node。
// 使用 Groups().RecursiveCreate 保证幂等，重复调用不会建出多个同名根。
func (s *Service) ensureDefaultGroupForView(ctx context.Context, view *model.ArtifactView) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		ge, err := s.artStore.Groups().RecursiveCreate(ctx, view.ID, []string{view.Name}, view.Format, view.CreatedBy)
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

// ResolveLib 按请求中的 artifactID 或默认视图规则解析实际视图。
// 规则：
//  1. artifactID > 0：优先按 ID 精确取库
//  2. artifactID == 0 或未传：继续按现有默认 view 逻辑解析
func (s *Service) ResolveLib(ctx context.Context, req *model.SpaceRepoReq) (*model.ArtifactView, error) {
	if req == nil {
		req = &model.SpaceRepoReq{}
	}
	view, err := s.artStore.Views().GetByID(ctx, req.ArtifactID)
	if err != nil {
		if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
			return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactLibNotFound)
		}
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactLibQueryError)
	}
	if req.SpaceID > 0 && view.SpaceID != req.SpaceID {
		return nil, pkgErrors.WithI18nKey(
			fmt.Errorf("artifactID %d does not belong to spaceID %d", req.ArtifactID, req.SpaceID),
			i18n.ArtifactLibScopeMismatch,
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
				i18n.ArtifactLibScopeMismatch,
			)
		}
	}
	return view, nil
}

func (s *Service) fillViewDescriptorScope(ctx context.Context, desc *adapter.ViewDescriptor, view *model.ArtifactView) error {
	if view.SpaceID > 0 {
		space, err := s.spaceStore.GetByID(ctx, uint(view.SpaceID))
		if err != nil {
			return pkgErrors.WithI18nKey(err, i18n.ArtifactSpaceQueryError)
		}
		desc.Space = &model.SpaceRef{ID: int64(space.ID), Name: space.Name, Code: space.Code}
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

// ParseSpaceLib 解析 SpaceRepoReq 参数为 ViewDescriptor。
func (s *Service) ParseSpaceLib(ctx context.Context, req *model.SpaceRepoReq) (*adapter.ViewDescriptor, error) {
	view, err := s.ResolveLib(ctx, req)
	if err != nil {
		return nil, err
	}

	desc := &adapter.ViewDescriptor{
		ArtifactID:     view.ID,
		ArtifactName:   view.Name,
		ArtifactCode:   view.Code,
		ArtifactFormat: view.Format,
		Store:          s.fileStore,
		StorageID:      1,
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
	SpaceCode string                  // space 作用域下的空间编码，供 FullName 拼接使用
	View      *adapter.ViewDescriptor // artifact 对应的 view descriptor
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

// getVersionNodeInfo 查询 version 节点详情，并按 format 组装下载或 pull 链接。
func (s *Service) getVersionNodeInfo(ctx context.Context, versionID int64) (*model.ArtifactAssetInfo, error) {
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

	viewID := pkg.ArtifactID
	if pkg.Format == model.ArtifactContainerFormat {
		viewID = ver.ArtifactLibID
	}
	view, err := s.ParseSpaceLib(ctx, &model.SpaceRepoReq{ArtifactID: viewID})
	if err != nil {
		return nil, err
	}

	info := &model.ArtifactAssetInfo{
		Status: string(model.ArtifactStatusOK),
		Format: pkg.Format,
		Metadata: &model.ArtifactPkgVerMetadata{
			Space:   nodeInfoSpaceName(view),
			Name:    pkg.Name,
			Group:   strings.ReplaceAll(pkg.Namespace, ".", "/"),
			Version: ver.Version,
		},
	}

	if pkg.Format == model.ArtifactContainerFormat {
		if err = s.fillContainerNodeInfo(ctx, info, view, ver); err != nil {
			return nil, err
		}
		info.Link = s.buildContainerPullLink(view, pkg, ver)
		// container 格式：用 Image/Package 替代 Name/Group，语义更清晰
		info.Metadata.Image = info.Metadata.Name
		info.Metadata.Package = info.Metadata.Group
		info.Metadata.Name = ""
		info.Metadata.Group = ""
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
	info.Link = buildArtifactAssetLink(nodeInfoSpaceCode(view), nodeInfoRepoIdentifier(view), view.ArtifactCode, string(pkg.Format), pkg.Namespace, pkg.Name, ver.Version, assetRes.Path)
	return info, nil
}

func (s *Service) getPrimaryVersionAsset(ctx context.Context, versionID int64) (*model.ArtifactAssetsRes, error) {
	// 优先选择 main asset；历史数据没有 main 标记时退回到第一个资产。
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

func (s *Service) fillContainerNodeInfo(ctx context.Context, info *model.ArtifactAssetInfo, view *adapter.ViewDescriptor, ver *model.ArtifactVersion) error {
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

	info.Images = containerImagesToModel(data.Images)

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

// getGroupAssetNodeInfo 获取 group 直挂资产节点信息
func (s *Service) getGroupAssetNodeInfo(ctx context.Context, dbAsset *model.ArtifactAsset, assetRes *model.ArtifactAssetsRes) (*model.ArtifactAssetInfo, error) {
	// group 直挂资产主要用于 generic/raw 目录文件，需要反查 group 路径拼下载链接。
	group, err := s.artStore.Groups().GetEndpointById(ctx, dbAsset.GroupID.Int64)
	if err != nil {
		if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
			return notFoundNodeInfo(), nil
		}
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactGroupQueryError)
	}

	view, err := s.ParseSpaceLib(ctx, &model.SpaceRepoReq{ArtifactID: group.ArtifactID})
	if err != nil {
		return nil, err
	}

	info := &model.ArtifactAssetInfo{
		Status: string(model.ArtifactStatusOK),
		Format: dbAsset.Format,
		Metadata: &model.ArtifactPkgVerMetadata{
			Space: nodeInfoSpaceName(view),
			Name:  filepath.Base(assetRes.Path),
			Group: group.GroupPath(),
		},
	}
	populateNodeInfoAssetFields(info, assetRes)
	info.Link = buildArtifactAssetLink(nodeInfoSpaceCode(view), nodeInfoRepoIdentifier(view), view.ArtifactCode, string(dbAsset.Format), group.GroupPath(), "", "", assetRes.Path)
	return info, nil
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
func artifactIDForVersion(pkg *model.ArtifactPackage, ver *model.ArtifactVersion) int64 {
	if ver != nil && ver.ArtifactLibID != 0 {
		return ver.ArtifactLibID
	}
	if pkg != nil {
		return pkg.ArtifactID
	}
	return 0
}
