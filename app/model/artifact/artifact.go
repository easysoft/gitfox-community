package artifact

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	enum "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

// ==================== 格式与资产类型 ====================
// ==================== Format & Kind ====================

type ArtifactFormat string

const (
	ArtifactAllFormat ArtifactFormat = "all"

	// ArtifactGenericFormat 表示按目录/文件层级直接存储任意文件。
	// ArtifactGenericFormat is used for store any files as folder/file style
	ArtifactGenericFormat ArtifactFormat = "generic"

	// ArtifactRawFormat 表示按 raw 仓库语义存储文件，要求 package 与 version。
	// ArtifactRawFormat is used for store any files as maven style, require package and version
	ArtifactRawFormat ArtifactFormat = "raw"

	ArtifactMavenFormat     ArtifactFormat = "maven"
	ArtifactContainerFormat ArtifactFormat = "container"
	ArtifactHelmFormat      ArtifactFormat = "helm"
	ArtifactPypiFormat      ArtifactFormat = "pypi"
	ArtifactNpmFormat       ArtifactFormat = "npm"
)

var AllArtifactFormatList = []ArtifactFormat{ArtifactGenericFormat, ArtifactRawFormat, ArtifactMavenFormat, ArtifactContainerFormat, ArtifactHelmFormat}

type AssetKind string

const (
	AssetKindMain AssetKind = "main"
	AssetKindSub  AssetKind = "subordinate"
)

func ValidateAssetKind(kind AssetKind) error {
	if kind != AssetKindMain && kind != AssetKindSub {
		return fmt.Errorf("unidentified kind '%s'", kind)
	}
	return nil
}

const (
	ArtifactNodeRoot = "/"

	ArtifactNodeLevelAsset   = "asset"
	ArtifactNodeLevelVersion = "version"
)

// ==================== 权限实体类型枚举 ====================
// ==================== Enum: Entity Kind ====================

type ArtifactEntityKind string

func (s ArtifactEntityKind) String() string {
	return string(s)
}

const (
	ArtifactEntityKindFormat  ArtifactEntityKind = "format"
	ArtifactEntityKindGroup   ArtifactEntityKind = "group"
	ArtifactEntityKindPackage ArtifactEntityKind = "package"
	ArtifactEntityKindVersion ArtifactEntityKind = "version"
	ArtifactEntityKindAsset   ArtifactEntityKind = "asset"
	ArtifactEntityEmpty       ArtifactEntityKind = "_empty"
)

// CasbinSubject 保存 Casbin 策略中的主体信息。
type CasbinSubject struct {
	Type        string `json:"type"`
	Id          int64  `json:"id"`
	UID         string `json:"uid"`
	DisplayName string `json:"displayName,omitempty"`
}

// ==================== 核心实体定义 ====================
// ==================== Core Entities ====================

type ArtifactView struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	SpaceID     int64           `json:"spaceId"`
	RepoID      int64           `json:"repoId"`
	Default     bool            `json:"isDefault"`
	CreatedBy   string          `json:"createdBy"`
	CreatedDate *datetime.Local `json:"createdDate"`
	EditedBy    string          `json:"editedBy"`
	EditedDate  *datetime.Local `json:"editedDate"`
	Deleted     int8            `json:"-"`
}
type CreateArtifactView struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SpaceID     int64  `json:"spaceId"`
	RepoID      int64  `json:"repoId"`
	Default     bool   `json:"isDefault"`
}

type ArtifactGroup struct {
	ID          int64          `json:"id"`
	ParentID    null.Int       `json:"parentId"`
	ViewID      int64          `json:"viewId"`
	Name        string         `json:"name"`
	Format      ArtifactFormat `json:"format"`
	CreatedBy   string         `json:"createdBy"`
	CreatedDate *time.Time     `json:"createdDate"`
	EditedBy    string         `json:"editedBy"`
	EditedDate  *time.Time     `json:"editedDate"`
	Deleted     int8           `json:"-"`
}

type CreateArtifactGroup struct {
	ParentID null.Int       `json:"parentId"`
	ViewID   int64          `json:"viewId"`
	Name     string         `json:"name"`
	Format   ArtifactFormat `json:"format"`
}

type ArtifactPackage struct {
	ID        int64  `json:"id"`
	ViewID    int64  `json:"viewId"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`

	// GroupID 用于标记 package 所属 group，主要给 raw / maven 风格使用，替代 namespace 的层级含义。
	// GroupId used for define group of the package, used for raw maven style, instead of the namespace field
	// 这里使用 int64，便于后续构建索引和关联查询。
	// convert to int64 for build index
	GroupID int64 `json:"groupId"`

	Format      ArtifactFormat `json:"format"`
	CreatedBy   string         `json:"createdBy"`
	CreatedDate *time.Time     `json:"createdDate"`
	EditedBy    string         `json:"editedBy"`
	EditedDate  *time.Time     `json:"editedDate"`
	Deleted     int8           `json:"-"`
}

type CreateArtifactPackage struct {
	ViewID    int64          `json:"viewId"`
	Name      string         `json:"name"`
	Namespace string         `json:"namespace"`
	Format    ArtifactFormat `json:"format"`
}

func (p *ArtifactPackage) IsDeleted() bool {
	return p.Deleted == 1
}

type ArtifactVersion struct {
	ID        int64  `json:"id"`
	PackageID int64  `json:"packageId"`
	Version   string `json:"version"`

	ViewID int64 `json:"viewId"`
	// Metadata 保存 version 级元数据，例如版本创建者等附加信息。
	// Metadata store version-level metadata. example: version creator
	Metadata string `json:"metadata"`

	CreatedBy   string     `json:"createdBy"`
	CreatedDate *time.Time `json:"createdDate"`
	EditedBy    string     `json:"editedBy"`
	EditedDate  *time.Time `json:"editedDate"`
	Deleted     int8       `json:"-"`
}

type CreateArtifactVersion struct {
	PackageID int64  `json:"packageId"`
	Version   string `json:"version"`
	ViewID    int64  `json:"viewId"`
	Metadata  string `json:"metadata"`
}

type ArtifactVersionInfo struct {
	ID      int64  `gorm:"column:id"`
	Version string `gorm:"column:version"`
	ViewID  int64  `gorm:"column:viewID"`
	Deleted int64  `gorm:"column:deleted"`

	PackageId        int64          `gorm:"column:packageID"`
	PackageName      string         `gorm:"column:name"`
	PackageNamespace string         `gorm:"column:namespace"`
	PackageFormat    ArtifactFormat `gorm:"column:format"`

	ViewSpaceID int64  `gorm:"column:spaceID"`
	ViewRepoID  int64  `gorm:"column:repoID"`
	SpaceName   string `gorm:"column:space_name"`
	RepoName    string `gorm:"column:repo_name"`
}

func (v *ArtifactVersion) IsDeleted() bool {
	return v.Deleted == 1
}

type ArtifactAsset struct {
	ID int64 `json:"id"`

	// VersionID 绑定具体 package(format) 与 view 下的版本资产。
	// VersionID binding with package(with format), view
	VersionID null.Int `json:"versionId"`

	// ViewID 用于无版本资产，例如格式级 index 文件。
	// ViewID used for none-version assets, like index file for a format
	ViewID null.Int `json:"viewId"`

	// GroupID 用于无版本的文件类资产挂到 group 节点。
	// GroupId used for file format assets, without version
	GroupID null.Int `json:"groupId"`

	// Format 也用于标识无版本资产所属的仓库格式。
	// Format is also used for none-version assets
	Format ArtifactFormat `json:"format"`

	Path        string    `json:"path"`
	ContentType string    `json:"contentType"`
	Kind        AssetKind `json:"kind"`
	// Metadata 保存 asset 级元数据，例如 helm chart 信息，用于后续重建索引。
	// Metadata store asset-level metadata. example: helm chart metadata, for rebuild index
	Metadata    string     `json:"metadata"`
	BlobID      int64      `json:"blobId"`
	CheckSum    string     `json:"checkSum"`
	CreatedBy   string     `json:"createdBy"`
	CreatedDate *time.Time `json:"createdDate"`
	EditedBy    string     `json:"editedBy"`
	EditedDate  *time.Time `json:"editedDate"`
	Deleted     int8       `json:"-"`
}

type CreateArtifactAsset struct {
	VersionID   null.Int       `json:"versionId"`
	ViewID      null.Int       `json:"viewId"`
	GroupID     null.Int       `json:"groupId"`
	Format      ArtifactFormat `json:"format"`
	Path        string         `json:"path"`
	ContentType string         `json:"contentType"`
	Kind        AssetKind      `json:"kind"`
	Metadata    string         `json:"metadata"`
	BlobID      int64          `json:"blobId"`
	CheckSum    string         `json:"checkSum"`
}

type ArtifactAssetExtendBlob struct {
	ArtifactAsset
	BlobID    int64  `gorm:"column:blobID"`
	StorageID int64  `gorm:"column:storageID"`
	Ref       string `gorm:"column:ref"`
	Size      int64  `gorm:"column:size"`
}

type ArtifactMetaAsset struct {
	ID          int64          `json:"id"`
	ViewID      int64          `json:"viewId"`
	Format      ArtifactFormat `json:"format"`
	Path        string         `json:"path"`
	ContentType string         `json:"contentType"`
	Kind        AssetKind      `json:"kind"`
	BlobID      int64          `json:"blobId"`
	CheckSum    string         `json:"checkSum"`
	CreatedBy   string         `json:"createdBy"`
	CreatedDate *time.Time     `json:"createdDate"`
	EditedBy    string         `json:"editedBy"`
	EditedDate  *time.Time     `json:"editedDate"`
	Deleted     int8           `json:"-"`
}

type CreateArtifactMetaAsset struct {
	ViewID      int64          `json:"viewId"`
	Format      ArtifactFormat `json:"format"`
	Path        string         `json:"path"`
	ContentType string         `json:"contentType"`
	Kind        AssetKind      `json:"kind"`
	BlobID      int64          `json:"blobId"`
	CheckSum    string         `json:"checkSum"`
}

type ArtifactBlob struct {
	ID        int64  `json:"id"`
	StorageID int64  `json:"storageId"`
	Ref       string `json:"ref"`
	Size      int64  `json:"size"`
	Downloads int64  `json:"downloads"`
	// Metadata 保存 blob 文件级元数据，例如容器上传进度或哈希状态。
	// Metadata store file-level metadata. example: container upload progress hash state
	Metadata    string     `json:"metadata"`
	Deleted     int8       `json:"-"`
	CreatedBy   string     `json:"createdBy"`
	CreatedDate *time.Time `json:"createdDate"`
	EditedBy    string     `json:"editedBy"`
	EditedDate  *time.Time `json:"editedDate"`
	AssetId     int64      `json:"assetId"`
	CheckSum    string     `json:"checkSum"`
}

type CreateArtifactBlob struct {
	StorageID int64  `json:"storageId"`
	Ref       string `json:"ref"`
	Size      int64  `json:"size"`
	CreatedBy string `json:"createdBy"`
	CheckSum  string `json:"checkSum"`
	Metadata  string `json:"metadata"`
}

type ArtifactTreeNode struct {
	ID       int64                `json:"id"`
	ParentID null.Int             `json:"parentId"`
	ViewID   int64                `json:"viewId"`
	Name     string               `json:"name"`
	Path     string               `json:"path"`
	Type     ArtifactTreeNodeType `json:"type"`
	Format   ArtifactFormat       `json:"format"`
	TargetId null.Int             `json:"targetId"`
}

type CreateArtifactTreeNode struct {
	ParentID null.Int             `json:"parentId"`
	ViewID   int64                `json:"viewId"`
	Name     string               `json:"name"`
	Path     string               `json:"path"`
	Type     ArtifactTreeNodeType `json:"type"`
	Format   ArtifactFormat       `json:"format"`
	TargetId null.Int             `json:"targetId"`
}

func (n *ArtifactTreeNode) Key() string {
	return fmt.Sprintf("%d:%s:%s", n.ViewID, n.Format, n.Path)
}

func (n *ArtifactTreeNode) IsRoot() bool {
	return n.Path == "/"
}

// Parent 返回当前节点的父节点抽象；根下一级会回落为 format 节点。
/*
Parent return the object parent node

	inputs:
	/a/b/c -> /a/b
	/a/b -> /a
	/a -> /
*/
func (n *ArtifactTreeNode) Parent() *ArtifactTreeNode {
	p := ArtifactTreeNode{
		ViewID: n.ViewID,
		Format: n.Format,
	}

	parentPath := filepath.Dir(n.Path)
	p.Name = filepath.Base(parentPath)
	p.Path = parentPath

	if p.Path == "/" {
		p.Name = string(n.Format)
		p.Type = ArtifactTreeNodeTypeFormat
	} else {
		p.Name = filepath.Base(parentPath)
		p.Type = ArtifactTreeNodeTypeDirectory
	}
	return &p
}

// ==================== 权限路径实体 ====================
// ==================== Entity (for permission path) ====================

type ArtifactEntity struct {
	Kind   ArtifactEntityKind
	Name   string
	Id     int64
	Parent *ArtifactEntity
}

func (e *ArtifactEntity) loc() string {
	if e.Kind != ArtifactEntityKindFormat {
		return fmt.Sprintf("%s:%d", e.Kind, e.Id)
	} else {
		return e.Name
	}
}

func (e *ArtifactEntity) Path() string {
	var frames []string

	node := e
	for node.Parent != nil {
		frames = append(frames, node.Name)
		node = node.Parent
	}
	frames = append(frames, node.Name)

	var buff bytes.Buffer
	count := len(frames)
	for i := 0; i < count; i++ {
		if i > 0 {
			buff.WriteString("/")
		}
		buff.WriteString(frames[count-1-i])
	}
	return buff.String()
}

func (e *ArtifactEntity) Join(entries ...*ArtifactEntity) *ArtifactEntity {
	count := len(entries)
	latest := e
	for i := 0; i < count; i++ {
		n := entries[i]
		parent := *latest
		n.Parent = &parent
		latest = n
	}
	return latest
}

func (e *ArtifactEntity) PolicyLocation() string {
	var frames []string

	node := e
	for node.Parent != nil {
		frames = append(frames, node.loc())
		node = node.Parent
	}
	frames = append(frames, node.loc())

	var buff bytes.Buffer
	count := len(frames)
	for i := 0; i < count; i++ {
		if i > 0 {
			buff.WriteString("/")
		}
		buff.WriteString(frames[count-1-i])
	}
	return buff.String()
}

// ==================== Entity 工具方法 ====================
// ==================== Entity Helpers ====================

// ParseEntityKind 将 entityId 字符串（格式 "kind.id"，如 "package.123"）解析为 ArtifactEntity。
// 用于中间件从 URL 路径参数快速构建 entity，无需查库。
// 解析失败时返回仓库级 mock entity。
func ParseEntityKind(entityId string) *ArtifactEntity {
	parts := strings.SplitN(entityId, ".", 2)
	if len(parts) != 2 {
		return &ArtifactEntity{Kind: ArtifactEntityKindFormat, Name: "repository"}
	}
	kind := ArtifactEntityKind(parts[0])
	switch kind {
	case ArtifactEntityKindGroup, ArtifactEntityKindPackage, ArtifactEntityKindVersion, ArtifactEntityKindAsset:
		return &ArtifactEntity{Kind: kind, Name: entityId}
	default:
		return &ArtifactEntity{Kind: ArtifactEntityKindFormat, Name: "repository"}
	}
}

// ==================== 组节点端点 ====================
// ==================== Group Endpoint ====================

type ArtifactGroupEndpoint struct {
	Parents []*ArtifactGroup
	*ArtifactGroup
}

func (e *ArtifactGroupEndpoint) IsDeleted() bool {
	var deleted bool
	for _, parent := range e.Parents {
		if parent.Deleted == 1 {
			deleted = true
			break
		}
	}

	return deleted || e.Deleted == 1
}

func (e *ArtifactGroupEndpoint) GroupName() string {
	return strings.Join(e.names(), ".")
}

func (e *ArtifactGroupEndpoint) GroupPath() string {
	return strings.Join(e.names(), "/")
}

func (e *ArtifactGroupEndpoint) names() []string {
	names := make([]string, 0)
	for _, parent := range e.Parents {
		names = append(names, parent.Name)
	}
	names = append(names, e.Name)
	return names
}

// ==================== 搜索选项 ====================
// ==================== Search Options ====================

type SearchVersionOption struct {
	PackageId      int64
	ViewId         int64
	Page           int    `json:"page"`
	Size           int    `json:"size"`
	Query          string `json:"query"`
	IncludeDeleted bool   `json:"includeDeleted"`
}

func (opt SearchVersionOption) Apply(db *gorm.DB) *gorm.DB {
	q := ArtifactVersion{}
	if opt.PackageId != 0 {
		q.PackageID = opt.PackageId
	}

	if opt.ViewId != 0 {
		q.ViewID = opt.ViewId
	}

	return db.Where(&q)
}

type SearchAssetOption struct {
	VersionId int64
	Kind      AssetKind
	Path      string
}

func (opt SearchAssetOption) Apply(db *gorm.DB) *gorm.DB {
	q := ArtifactAsset{
		VersionID: null.IntFromPtr(&opt.VersionId),
	}

	if opt.Kind != ("") {
		q.Kind = opt.Kind
	}

	return db.Where(&q)
}

// ==================== 树节点响应模型 ====================
// ==================== Tree Node ====================

type TreeNode struct {
	Id       int64       `json:"id"`
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	Children []*TreeNode `json:"children,omitempty"`
	IsLeaf   bool        `json:"isLeaf"`
}

// ==================== 回收与统计模型 ====================
// ==================== Recycle / Statistic ====================

type ArtifactRecycleBlobDesc struct {
	AssetId int64
	Path    string
	BlobId  int64
	BlobRef string
	Size    int64
}

type ArtifactVersionCapacityDesc struct {
	AssetId       int64
	VersionId     int64
	ExclusiveSize int64
	TotalSize     int64
	ExclusiveRefs int
	TotalRefs     int
}

type ArtifactStatisticReport struct {
	DeleteList []*ArtifactRecycleBlobDesc
	TagList    []*ArtifactVersionCapacityDesc
}

// ==================== 错误定义 ====================
// ==================== Errors ====================

var (
	ErrPkgNotFound      = errors.New("artifact package not found")
	ErrPkgNoItemDeleted = errors.New("no artifact package was deleted")

	ErrPkgVersionNotFound      = errors.New("artifact package version not found")
	ErrPkgVersionNoItemDeleted = errors.New("no artifact package was deleted")

	ErrAssetNotFound      = errors.New("artifact asset not found")
	ErrAssetNoItemDeleted = errors.New("no artifact asset was deleted")

	ErrMetaAssetNotFound      = errors.New("artifact meta asset not found")
	ErrMetaAssetNoItemDeleted = errors.New("no artifact meta asset was deleted")

	ErrBlobNotFound      = errors.New("artifact blob not found")
	ErrBlobNoItemDeleted = errors.New("no artifact blob was deleted")

	ErrViewDefaultNotFound = errors.New("default view not found")
	ErrViewNotFound        = errors.New("special view not found")
	ErrArgsValueEmpty      = errors.New("args value should not empty")
)

// ==================== 权限请求模型 ====================
// ==================== Permission Request ====================

// PermissionRequest 是设置制品库权限的请求体。
type PermissionRequest struct {
	Private bool       `json:"private"`
	Rules   []PermRule `json:"rules"`
}

// PermRule 是权限规则中的单条规则，包含用户列表、角色和作用域。
type PermRule struct {
	Accounts []string         `json:"accounts"`
	Role     enum.CasbinRole  `json:"role"`
	Scope    enum.CasbinScope `json:"scope"`
}
