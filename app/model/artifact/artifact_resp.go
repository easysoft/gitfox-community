package artifact

import (
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"

	"github.com/guregu/null"
)

// ==================== 查询过滤条件 ====================
// ==================== Filters ====================

// ArtifactFilter 定义制品列表查询参数。
// ArtifactFilter define artifacts query parameters.
type ArtifactFilter struct {
	Page   int    `json:"page"`
	Size   int    `json:"size"`
	Query  string `json:"query"`
	Format string `json:"format"`
	Sort   string `json:"sort"`
	Order  string `json:"order"`
}

// ArtifactVersionFilter 定义版本列表查询参数。
// ArtifactVersionFilter define artifact versions query parameters.
type ArtifactVersionFilter struct {
	Page    int    `json:"page"`
	Size    int    `json:"size"`
	Query   string `json:"query"`
	Package string `json:"package"`
	Group   string `json:"group"`
}

type ArtifactTreeFilter struct {
	NodeId int64          `json:"nodeID"`
	Path   string         `json:"path"`
	Format ArtifactFormat `json:"format"`
	Level  string         `json:"level"`
}

// ==================== 响应模型 ====================
// ==================== Response Types ====================

type ArtifactListItem struct {
	Format      string          `gorm:"column:type" json:"format"`
	Name        string          `gorm:"column:name" json:"name"`
	Namespace   string          `gorm:"column:namespace" json:"namespace"`
	DisplayName string          `gorm:"-" json:"displayName"`
	Version     string          `gorm:"column:version" json:"version"`
	UpdateTime  *datetime.Local `gorm:"column:editedDate" json:"editedDate"`
}

type ArtifactAssetsRes struct {
	Id             int64           `gorm:"column:id" json:"id"`
	Path           string          `gorm:"column:basename" json:"path"`
	ContentType    string          `gorm:"column:contentType" json:"contentType"`
	Size           int64           `gorm:"column:size" json:"size"`
	BlobRef        string          `gorm:"column:ref" json:"-"`
	Creator        string          `gorm:"-" json:"-"`
	CreatorName    string          `gorm:"column:createdBy" json:"createdBy"`
	EditedBy       string          `gorm:"column:editedBy" json:"editedBy"`
	CheckSumString string          `gorm:"column:checksum" json:"-"`
	MetadataString string          `gorm:"-" json:"-"`
	Created        *datetime.Local `gorm:"column:createdDate" json:"createdDate"`
	Updated        *datetime.Local `gorm:"column:editedDate" json:"editedDate"`

	CheckSum *CheckSumRes      `gorm:"-" json:"checksum"`
	Metadata map[string]string `gorm:"-" json:"metadata"`
	Link     string            `gorm:"-" json:"link"`
	EntityID string            `gorm:"-" json:"entityID"`

	// ==================== 业务回填字段（由 service 层基于 nodeinfo 链路统一填充）====================
	// Name 文件名 / 制品名（asset path 的 basename）
	Name string `gorm:"-" json:"name,omitempty"`
	// PackageName 所属包名（version 类资产；非 version 资产为空）
	PackageName string `gorm:"-" json:"packageName,omitempty"`
	// Group 所属目录（generic/raw 为点分路径；container 为镜像目录；不存在则空）
	Group string `gorm:"-" json:"group,omitempty"`
	// Version 版本号（version 类资产；非 version 资产为空）
	Version string `gorm:"-" json:"version,omitempty"`
	// OS / Arch 仅 container 格式由 manifest 解析后回填
	OS   string `gorm:"-" json:"os,omitempty"`
	Arch string `gorm:"-" json:"arch,omitempty"`
}

type ArtifactBlobRes struct {
	Id             int64           `gorm:"column:id" json:"id"`
	Path           string          `gorm:"column:basename" json:"path"`
	Size           int64           `gorm:"column:size" json:"size"`
	BlobRef        string          `gorm:"column:ref" json:"-"`
	CreatedBy      string          `gorm:"column:createdBy" json:"createdBy"`
	CheckSumString string          `gorm:"column:checksum" json:"-"`
	MetadataString string          `gorm:"column:metadata" json:"-"`
	Created        *datetime.Local `gorm:"column:createdDate" json:"createdDate"`
	Deleted        null.Int        `gorm:"column:deleted" json:"deleted"`

	CheckSum *CheckSumRes      `gorm:"-" json:"checksum"`
	Metadata map[string]string `gorm:"-" json:"metadata"`
	Link     string            `gorm:"-" json:"link"`
	EntityID string            `gorm:"-" json:"entityID"`
}

type CheckSumRes struct {
	Md5    string `json:"md5"`
	Sha1   string `json:"sha1"`
	SHA256 string `json:"sha256"`
	SHA512 string `json:"sha512"`
}

type ArtifactVersionsRes struct {
	Version     string `json:"version"`
	CreatorName string `json:"createdBy"`
	Updated     int64  `json:"editedDate"`
}

// ==================== 树节点响应 ====================
// ==================== Tree Node Response ====================

type ArtifactTreeNodeType string

const (
	ArtifactTreeNodeTypeFormat   ArtifactTreeNodeType = "format"
	ArtifactTreeNodeTypeGroup    ArtifactTreeNodeType = "group"
	ArtifactTreeNodeTypePackage  ArtifactTreeNodeType = "package"
	ArtifactTreeNodeTypeVersion  ArtifactTreeNodeType = "version"
	ArtifactTreeNodeTypeAsset    ArtifactTreeNodeType = "asset"
	ArtifactTreeNodeTypeArtifact ArtifactTreeNodeType = "artifact"
)

type ArtifactTreeRes struct {
	Name     string                `json:"name"`
	Path     string                `json:"path"`
	Leaf     bool                  `json:"leaf"`
	Format   ArtifactFormat        `json:"format"`
	Metadata *ArtifactTreeNodeMeta `json:"metadata,omitempty"`
	NodeInfo *ArtifactAssetInfo    `json:"nodeInfo,omitempty"`
	Children []*ArtifactTreeRes    `json:"children,omitempty"`
}

type ArtifactTreeNodeMeta struct {
	// Id 是 ops_artifact_tree_nodes.id（树节点主键）。
	// asset 节点不写入 tree_nodes 表，故该字段留空（按 omitempty 不输出）。
	// 跨接口实体引用请用 EntityID（"kind.pk" 形式，如 asset.5）。
	Id      int64                `json:"id,omitempty"`
	Type    ArtifactTreeNodeType `json:"type"`
	Name    string               `json:"name,omitempty"`
	Group   string               `json:"group,omitempty"`
	Version string               `json:"version,omitempty"`

	// EntityID 是 "kind.pk" 形式的实体引用串（如 group.11 / asset.5），
	// 用于跨接口定位实体（移动 / 重命名 / 软删 / 详情等接口的入参）。
	EntityID string `json:"entityID,omitempty"`
}

type ArtifactTreeNodeId struct {
	Type ArtifactTreeNodeType `json:"type"`
	Pk   int64                `json:"pk"`
}

func (nid *ArtifactTreeNodeId) String() string {
	return fmt.Sprintf("%s.%d", nid.Type, nid.Pk)
}

// ==================== 回收站 ====================
// ==================== Recycle Bin ====================

// ArtifactRecycleKind 回收站列表的实体类型枚举，与 EntityKind 同集合，仅用于 query 入参校验。
type ArtifactRecycleKind string

const (
	ArtifactRecycleKindArtifact ArtifactRecycleKind = "artifact"
	ArtifactRecycleKindGroup    ArtifactRecycleKind = "group"
	ArtifactRecycleKindPackage  ArtifactRecycleKind = "package"
	ArtifactRecycleKindVersion  ArtifactRecycleKind = "version"
	ArtifactRecycleKindAsset    ArtifactRecycleKind = "asset"
)

// AllArtifactRecycleKinds 全量合法 kind 集合，供 controller 校验。
var AllArtifactRecycleKinds = []ArtifactRecycleKind{
	ArtifactRecycleKindArtifact,
	ArtifactRecycleKindGroup,
	ArtifactRecycleKindPackage,
	ArtifactRecycleKindVersion,
	ArtifactRecycleKindAsset,
}

// ArtifactRecycleListReq 回收站列表请求。
//   - Kind  : 必填，缺省走 asset；只接受 AllArtifactRecycleKinds 集合内的值
//   - Page  : 缺省 1
//   - PageSize: 缺省 20，上限 200（store 层兜底）
type ArtifactRecycleListReq struct {
	Kind     ArtifactRecycleKind `form:"kind" json:"kind"`
	Page     int                 `form:"page" json:"page"`
	PageSize int                 `form:"pageSize" json:"pageSize"`
}

// ArtifactRecycleItem 回收站列表项。字段语义：
//   - Kind         : 实体类型
//   - EntityID     : "kind.id" 实体引用串，与软删 / 恢复接口入参一致
//   - Id           : 实体表自身主键
//   - Name         : 实体名（asset/group → name；package → 包名；version → 所属包名；artifact → 视图名）
//   - Version      : 仅 version 类填
//   - Group        : package/version 反查所属 group 的点分路径（namespace 等价）
//   - Path         : 仅 group 类填，从根到自身的点分路径
//   - Format       : 实体所属制品格式
//   - ArtifactID   : 自身或经反查得到的所属制品库 ID
//   - DeletedBy    : 软删人（取 editedBy）
//   - DeletedDate  : 软删时间（取 editedDate）
//   - ExpiresAt    : DeletedDate + RecycleBinRetention，到期后由 GC 物理清理
type ArtifactRecycleItem struct {
	Kind        ArtifactRecycleKind `json:"kind"`
	EntityID    string              `json:"entityID"`
	Id          int64               `json:"id"`
	Name        string              `json:"name,omitempty"`
	Version     string              `json:"version,omitempty"`
	Group       string              `json:"group,omitempty"`
	Path        string              `json:"path,omitempty"`
	Format      ArtifactFormat      `json:"format,omitempty"`
	ArtifactID  int64               `json:"artifactID,omitempty"`
	DeletedBy   string              `json:"deletedBy,omitempty"`
	DeletedDate *datetime.Local     `json:"deletedDate,omitempty"`
	ExpiresAt   *datetime.Local     `json:"expiresAt,omitempty"`
}

// ==================== 节点详情响应 ====================
// ==================== Node Info ====================

type ArtifactStatus string

const (
	ArtifactStatusOK                ArtifactStatus = "ok"
	ArtifactStatusInvalidId         ArtifactStatus = "invalid_id"
	ArtifactStatusNotFound          ArtifactStatus = "not found"
	ArtifactStatusNotEmpty          ArtifactStatus = "not_empty"
	ArtifactStatusNameAlreadyExists ArtifactStatus = "name_already_exists"
	ArtifactStatusUnknown           ArtifactStatus = "unknown"
)

// ArtifactAssetInfo
// 字段说明：
//   - 名称（Name）      → Metadata.Name（包名）
//   - 版本（Version）   → Metadata.Version
//   - 分组（Group）     → Metadata.Group（命名空间 / 二级目录）
//   - 所属空间          → Metadata.Space
//   - 大小（Size）      → Size（字节）
//   - 校验值（CheckSum）→ CheckSum（含 md5/sha1/sha256/sha512）
//   - 创建人            → CreatorName（JSON: createdBy）
//   - 创建时间          → Created（JSON: createdDate）
//   - 最后更新时间      → Updated（JSON: editedDate）
//   - 系统/架构         → OS / Arch（仅 container 格式填充，解析自 manifest 元数据；
//     multi-arch 镜像优先取 linux/amd64，否则取第一个平台）
type ArtifactAssetInfo struct {
	// 资产 ID
	ID int64 `json:"id,omitempty"`
	// 查询状态：ok / not found / invalid_id / unknown
	Status string `json:"status"`
	// 制品格式：container / helm / raw / generic 等
	Format ArtifactFormat `json:"format"`
	// 包/版本基础元数据（名称、版本、分组、所属空间）
	Metadata *ArtifactPkgVerMetadata `json:"metadata"`

	// 文件存储路径（container 为 sha256:xxx digest，其它为相对路径）
	Path string `json:"path"`
	// MIME 类型（如 application/vnd.docker.distribution.manifest.v2+json）
	ContentType string `json:"contentType"`
	// 文件大小（字节）
	Size int64 `json:"size"`
	// 创建人（用户 UID）
	CreatorName string `json:"createdBy"`
	// 更新人（用户 UID）
	EditedBy string `json:"editedBy"`
	// 创建时间
	Created *datetime.Local `json:"createdDate"`
	// 最后更新时间
	Updated *datetime.Local `json:"editedDate"`
	// 资产下载链接
	Link string `json:"link"`
	// 校验值（md5 / sha1 / sha256 / sha512），由上传时计算并存储
	CheckSum *CheckSumRes `json:"checksum"`
	// 操作系统（仅 container 格式）：linux / windows 等
	OS string `json:"os,omitempty"`
	// CPU 架构（仅 container 格式）：amd64 / arm64 / arm/v7 等
	Arch string `json:"arch,omitempty"`
	// 多平台镜像列表（仅 container 格式）：manifest list / OCI index 会返回多个平台。
	Images []*ArtifactContainerImage `json:"images,omitempty"`
}

type ArtifactContainerImage struct {
	Digest string `json:"digest"`
	OS     string `json:"os"`
	Arch   string `json:"arch"`
	Size   int64  `json:"size"`
}

// ArtifactPkgVerMetadata 制品包/版本的基础定位元数据。
type ArtifactPkgVerMetadata struct {
	// 所属空间名（global 时为 "global"）
	Space string `json:"space"`
	// 包名（非 container 格式使用）
	Name string `json:"name,omitempty"`
	// 分组 / 命名空间（非 container 格式使用）
	Group string `json:"group,omitempty"`
	// 版本号（docker tag / helm version / raw 版本标识）
	Version string `json:"version"`
	// 镜像名（仅 container 格式）
	Image string `json:"image,omitempty"`
	// 容器包名 / namespace（仅 container 格式）
	Package string `json:"package,omitempty"`
}

// ==================== 删除与回收统计 ====================
// ==================== Remove / GC ====================

type ArtifactNodeRemoveRes struct {
	NodeID   EntityID       `json:"nodeID"`
	Status   ArtifactStatus `json:"status"`
	Packages int            `json:"packages,omitempty"`
	Versions int            `json:"versions,omitempty"`
	Assets   int            `json:"assets,omitempty"`
}

type ArtifactNodeRemoveReport struct {
	Total   int                      `json:"total"`
	Success int                      `json:"success"`
	Failed  int                      `json:"failed"`
	Data    []*ArtifactNodeRemoveRes `json:"data"`
}

// ArtifactRestoreRes 单条恢复结果。结构与 ArtifactNodeRemoveRes 同形：
//   - NodeID         : 入参的 EntityID 串（kind.id）
//   - Status         : 处理状态（ok / not_found / invalid_id / name_already_exists / unknown 等）
//   - Reason         : 失败时的 i18n key，前端按 key 取多语言文案；成功时为空
//   - Packages/Versions/Assets : 本次恢复链路上各类实体的累计影响计数
type ArtifactRestoreRes struct {
	NodeID   EntityID       `json:"nodeID"`
	Error    error          `json:"-"`
	Status   ArtifactStatus `json:"status"`
	Reason   string         `json:"reason,omitempty"`
	Packages int            `json:"packages,omitempty"`
	Versions int            `json:"versions,omitempty"`
	Assets   int            `json:"assets,omitempty"`
}

// ArtifactRestoreReport 批量恢复汇总，结构与 ArtifactNodeRemoveReport 对称。
type ArtifactRestoreReport struct {
	Message string                `json:"-"`
	Total   int                   `json:"total"`
	Success int                   `json:"success"`
	Failed  int                   `json:"failed"`
	Data    []*ArtifactRestoreRes `json:"data"`
}

type JobUIDResponse struct {
	UID string `json:"uid"`
}

type ArtifactGarbageReportRes struct {
	Count int64   `json:"count"`
	Size  int64   `json:"size"`
	Ids   []int64 `json:"ids"`
}

type ArtifactStatisticResponse struct {
	GarbageCollect *ArtifactGarbageReportRes     `json:"garbageCollect"`
	Capacity       []*ArtifactPackageCapacityRes `json:"capacity"`
}

type ArtifactPackageCapacityRes struct {
	Scope         string                       `json:"scope"`
	Space         string                       `json:"space"`
	Repo          string                       `json:"repo,omitempty"`
	Name          string                       `json:"name"`
	Namespace     string                       `json:"namespace"`
	Format        ArtifactFormat               `json:"format"`
	Size          int64                        `json:"size"`
	ExclusiveSize int64                        `json:"exclusiveSize"`
	Versions      []ArtifactVersionCapacityRes `json:"versions"`
}

type ArtifactVersionCapacityRes struct {
	Version       string `json:"version"`
	Size          int64  `json:"size"`
	ExclusiveSize int64  `json:"exclusiveSize"`
}

type ArtifactStatsObject struct {
	Total   int64 `json:"total"`
	Deleted int64 `json:"deleted"`
	Active  int64 `json:"active"`
}
