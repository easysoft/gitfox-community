// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package artifact

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/easysoft/gitfox-community/app/constant"
	enum "github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

const (
	DefaultArtifactLibName string = "Default"
	DefaultArtifactLibCode string = "default"
)

// DefaultArtifactLibViewName 返回默认制品库的展示名：按 locale 取本地化名称。
// 普通建仓（app/api/controller/repo/repo.go）与导入建仓（app/service/importer）共用本函数，
// 避免同一功能出现两种展示名。
// 翻译缺失时 i18n 返回的是 key 本身或 "[locale:key]" 占位串，都不能作为库名落库，
// 统一回退到 DefaultArtifactLibName（与 en-US 的取值一致）。
func DefaultArtifactLibViewName(locale string) string {
	key := string(i18n.ArtifactDefaultViewName)
	name := i18n.TWithLocale(locale, key)
	if name == "" || name == key || strings.HasPrefix(name, "[") {
		return DefaultArtifactLibName
	}
	return name
}

// ==================== 格式与资产类型 ====================
// ==================== Format & Rank ====================

type ArtifactFormat string

const (
	ArtifactAllFormat ArtifactFormat = "all"

	// ArtifactGenericFormat 表示按目录/文件层级直接存储任意文件。
	// ArtifactGenericFormat is used for store any files as folder/file style
	ArtifactFileFormat ArtifactFormat = "file"

	// ArtifactRawFormat 表示按 raw 仓库语义存储文件，要求 package 与 version。
	// ArtifactRawFormat is used for store any files as maven style, require package and version
	ArtifactRawFormat ArtifactFormat = "raw"

	ArtifactMavenFormat     ArtifactFormat = "maven"
	ArtifactContainerFormat ArtifactFormat = "container"
	ArtifactHelmFormat      ArtifactFormat = "helm"
	ArtifactPypiFormat      ArtifactFormat = "pypi"
	ArtifactNpmFormat       ArtifactFormat = "npm"
)

var AllArtifactFormatList = []ArtifactFormat{ArtifactFileFormat, ArtifactRawFormat, ArtifactMavenFormat, ArtifactContainerFormat, ArtifactHelmFormat}

type AssetRank string

const (
	AssetRankMain AssetRank = "main"
	AssetRankSub  AssetRank = "subordinate"

//Format 是否有 sub asset 说明
// generic 没有 单个文件，全部 main
// raw 没有 单个文件，全部 main
// container（Docker/OCI） 没有 manifest/config/layer blob 都是 main
// helm 几乎没有（chart 主体一个 tgz） main 居多，prov 签名才会是 sub
// maven 有 每个 jar 带 .asc/.md5/.sha1/.sha256/.sha512
// pypi/npm 视实现 一般 main 为主

// Path Rank
// maven-metadata.xml main
// maven-metadata.xml.md5 sub
// maven-metadata.xml.sha1 sub
// maven-metadata.xml.sha256 sub
// maven-metadata.xml.sha512 sub
)

func ValidateAssetRank(rank AssetRank) error {
	if rank != AssetRankMain && rank != AssetRankSub {
		return fmt.Errorf("unidentified rank '%s'", rank)
	}
	return nil
}

const (
	ArtifactNodeRoot = "/"

	ArtifactNodeLevelFormat  = "format"
	ArtifactNodeLevelGroup   = "group"
	ArtifactNodeLevelPackage = "package"
	ArtifactNodeLevelVersion = "version"
	ArtifactNodeLevelAsset   = "asset"
)

// ==================== 权限实体类型枚举 ====================
// ==================== Enum: Entity Kind ====================

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
	ID          int64            `json:"id"`
	Name        string           `json:"name"`
	Code        string           `json:"code"`
	Description string           `json:"description"`
	SpaceID     int64            `json:"spaceID"`
	RepoID      int64            `json:"repoID"`
	Format      ArtifactFormat   `json:"format"`
	Type        ArtifactViewType `json:"type"`
	Default     bool             `json:"isDefault"`
	CreatedBy   string           `json:"createdBy"`
	CreatedDate *datetime.Local  `json:"createdDate"`
	EditedBy    string           `json:"editedBy"`
	EditedDate  *datetime.Local  `json:"editedDate"`
	Deleted     int8             `json:"-"`
}

type ArtifactViewType string

const (
	ArtifactViewTypeGlobal ArtifactViewType = "global"
	ArtifactViewTypeSpace  ArtifactViewType = "space"
	ArtifactViewTypeRepo   ArtifactViewType = "repo"
)

type CreateArtifactView struct {
	Name        string           `json:"name"`
	Code        string           `json:"code"`
	Description string           `json:"description"`
	SpaceID     int64            `json:"spaceID"`
	RepoID      int64            `json:"repoID"`
	Type        ArtifactViewType `json:"type"`
	Format      ArtifactFormat   `json:"format"`
	EditedBy    string           `json:"editedBy"`
	CreatedBy   string           `json:"createdBy"`
}

type ArtifactGroup struct {
	ID          int64           `json:"id"`
	ParentID    null.Int        `json:"parentID"`
	ArtifactID  int64           `json:"artifactID"`
	Name        string          `json:"name"`
	Format      ArtifactFormat  `json:"format"`
	CreatedBy   string          `json:"createdBy"`
	CreatedDate *datetime.Local `json:"createdDate"`
	EditedBy    string          `json:"editedBy"`
	EditedDate  *datetime.Local `json:"editedDate"`
	Deleted     int8            `json:"-"`
}

type CreateArtifactGroup struct {
	ParentID   null.Int       `json:"parentID"`
	ArtifactID int64          `json:"artifactID"`
	Name       string         `json:"name"`
	Format     ArtifactFormat `json:"format"`
	CreatedBy  string         `json:"createdBy"`
	EditedBy   string         `json:"editedBy"`
}

type ArtifactPackage struct {
	ID         int64  `json:"id"`
	ArtifactID int64  `json:"artifactID"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`

	// GroupID 用于标记 package 所属 group，主要给 raw / maven 风格使用，替代 namespace 的层级含义。
	// GroupID used for define group of the package, used for raw maven style, instead of the namespace field
	// 这里使用 int64，便于后续构建索引和关联查询。
	// convert to int64 for build index
	GroupID int64 `json:"groupID"`

	Format      ArtifactFormat `json:"format"`
	CreatedBy   string         `json:"createdBy"`
	CreatedDate *time.Time     `json:"createdDate"`
	EditedBy    string         `json:"editedBy"`
	EditedDate  *time.Time     `json:"editedDate"`
	Deleted     int8           `json:"-"`
}

type CreateArtifactPackage struct {
	ArtifactID int64          `json:"artifactID"`
	Name       string         `json:"name"`
	Namespace  string         `json:"namespace"`
	GroupID    int64          `json:"groupID"`
	Format     ArtifactFormat `json:"format"`
	CreatedBy  string         `json:"createdBy"`
}

func (p *ArtifactPackage) IsDeleted() bool {
	return p.Deleted == 1
}

type ArtifactVersion struct {
	ID            int64  `json:"id"`
	ArtifactLibID int64  `json:"artifactLibID"`
	PackageID     int64  `json:"packageID"`
	Version       string `json:"version"`

	// Metadata 保存 version 级元数据，例如版本创建者等附加信息。
	// Metadata store version-level metadata. example: version creator
	Metadata string `json:"metadata"`

	CreatedBy   string          `json:"createdBy"`
	CreatedDate *datetime.Local `json:"createdDate"`
	EditedBy    string          `json:"editedBy"`
	EditedDate  *datetime.Local `json:"editedDate"`
	Deleted     int8            `json:"-"`
}

type CreateArtifactVersion struct {
	ArtifactLibID int64  `json:"artifactLibID"`
	PackageID     int64  `json:"packageID"`
	Version       string `json:"version"`
	Metadata      string `json:"metadata"`
	CreatedBy     string `json:"createdBy"`
}

type ArtifactVersionInfo struct {
	ID            int64  `gorm:"column:id"`
	Version       string `gorm:"column:version"`
	ArtifactLibID int64  `gorm:"column:artifactLibID"`
	Deleted       uint8  `gorm:"column:deleted"`

	PackageID        int64          `gorm:"column:packageID"`
	PackageName      string         `gorm:"column:name"`
	PackageNamespace string         `gorm:"column:namespace"`
	PackageFormat    ArtifactFormat `gorm:"column:type"`

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
	VersionID null.Int `json:"versionID" gorm:"column:versionID"`

	// ArtifactID 用于无版本资产，例如格式级 index 文件。
	// ArtifactID used for none-version assets, like index file for a format
	ArtifactID null.Int `json:"artifactID" gorm:"column:artifactLibID"`

	// GroupID 用于无版本的文件类资产挂到 group 节点。
	// GroupID used for file format assets, without version
	GroupID null.Int `json:"groupID" gorm:"column:groupID"`

	// Format 也用于标识无版本资产所属的仓库格式。
	// Format is also used for none-version assets
	Format ArtifactFormat `json:"format"`

	Path        string    `json:"path"`
	ContentType string    `json:"contentType"`
	Rank        AssetRank `json:"rank"`
	// Metadata 保存 asset 级元数据，例如 helm chart 信息，用于后续重建索引。
	// Metadata store asset-level metadata. example: helm chart metadata, for rebuild index
	Metadata    string          `json:"metadata"`
	BlobID      int64           `json:"blobID"`
	CheckSum    string          `json:"checkSum"`
	CreatedBy   string          `json:"createdBy"`
	CreatedDate *datetime.Local `json:"createdDate"`
	EditedBy    string          `json:"editedBy"`
	EditedDate  *datetime.Local `json:"editedDate"`
	Deleted     int8            `json:"-"`
}

type CreateArtifactAsset struct {
	VersionID   null.Int       `json:"versionID"`
	ArtifactID  null.Int       `json:"artifactID"`
	GroupID     null.Int       `json:"groupID"`
	Format      ArtifactFormat `json:"format"`
	Path        string         `json:"path"`
	ContentType string         `json:"contentType"`
	Rank        AssetRank      `json:"rank"`
	BlobID      int64          `json:"blobID"`
	CheckSum    string         `json:"checkSum"`
	CreatedBy   string         `json:"createdBy"`
	EditedBy    string         `json:"editedBy"`
}

type ArtifactAssetExtendBlob struct {
	ArtifactAsset
	BlobID    int64  `json:"blobID"`
	StorageID int64  `json:"storageID"`
	Ref       string `json:"ref"`
	Size      int64  `json:"size"`
}

type ArtifactMetaAsset struct {
	ID          int64          `json:"id"`
	ArtifactID  int64          `json:"artifactID"`
	Format      ArtifactFormat `json:"format"`
	Path        string         `json:"path"`
	ContentType string         `json:"contentType"`
	Rank        AssetRank      `json:"rank"`
	BlobID      int64          `json:"blobID"`
	CheckSum    string         `json:"checkSum"`
	CreatedBy   string         `json:"createdBy"`
	CreatedDate *time.Time     `json:"createdDate"`
	EditedBy    string         `json:"editedBy"`
	EditedDate  *time.Time     `json:"editedDate"`
	Deleted     int8           `json:"-"`
}

type CreateArtifactMetaAsset struct {
	ArtifactID  int64          `json:"artifactID"`
	Format      ArtifactFormat `json:"format"`
	Path        string         `json:"path"`
	ContentType string         `json:"contentType"`
	Rank        AssetRank      `json:"rank"`
	BlobID      int64          `json:"blobID"`
	CheckSum    string         `json:"checkSum"`
}

type ArtifactBlob struct {
	ID        int64  `json:"id"`
	StorageID int64  `json:"storageID"`
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
	AssetID     int64      `json:"assetID"`
	CheckSum    string     `json:"checkSum"`
}

type CreateArtifactBlob struct {
	StorageID   int64      `json:"storageID"`
	Ref         string     `json:"ref"`
	Size        int64      `json:"size"`
	CreatedDate *time.Time `json:"createdDate"`
	CreatedBy   string     `json:"createdBy"`
	CheckSum    string     `json:"checkSum"`
	Metadata    string     `json:"metadata"`
}

type ArtifactTreeNode struct {
	ID         int64                `json:"id"`
	ParentID   null.Int             `json:"parentID"`
	ArtifactID int64                `json:"artifactID"`
	Name       string               `json:"name"`
	Path       string               `json:"path"`
	Type       ArtifactTreeNodeType `json:"type"`
	Format     ArtifactFormat       `json:"format"`
	TargetID   int64                `json:"targetID"`
}

type CreateArtifactTreeNode struct {
	ParentID   null.Int             `json:"parentID"`
	ArtifactID int64                `json:"artifactID"`
	Name       string               `json:"name"`
	Path       string               `json:"path"`
	Type       ArtifactTreeNodeType `json:"type"`
	Format     ArtifactFormat       `json:"format"`
	TargetID   int64                `json:"targetID"`
}

func (n *ArtifactTreeNode) Key() string {
	return fmt.Sprintf("%d:%s:%s", n.ArtifactID, n.Format, n.Path)
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
		ArtifactID: n.ArtifactID,
		Format:     n.Format,
	}

	parentPath := filepath.Dir(n.Path)
	p.Name = filepath.Base(parentPath)
	p.Path = parentPath

	if p.Path == "/" {
		p.Name = string(n.Format)
		p.Type = ArtifactTreeNodeTypeFormat
	} else {
		p.Name = filepath.Base(parentPath)
		p.Type = ArtifactTreeNodeTypeGroup
	}
	return &p
}

// ==================== 权限路径实体 ====================
// ==================== Entity (for permission path) ====================

type ArtifactEntity struct {
	Kind   constant.ArtifactEntityKind
	Name   string
	Id     int64
	Parent *ArtifactEntity
}

func (e *ArtifactEntity) loc() string {
	if e.Kind != constant.ArtifactEntityKindFormat {
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
	ArtifactLibID  int64
	PackageID      int64
	Page           int    `json:"page"`
	Size           int    `json:"size"`
	Query          string `json:"query"`
	IncludeDeleted bool   `json:"includeDeleted"`
}

func (opt SearchVersionOption) Apply(db *gorm.DB) *gorm.DB {
	q := ArtifactVersion{}
	if opt.ArtifactLibID != 0 {
		q.ArtifactLibID = opt.ArtifactLibID
	}
	if opt.PackageID != 0 {
		q.PackageID = opt.PackageID
	}

	return db.Where(&q)
}

type SearchAssetOption struct {
	VersionID int64
	Rank      AssetRank
	Path      string
}

func (opt SearchAssetOption) Apply(db *gorm.DB) *gorm.DB {
	q := ArtifactAsset{
		VersionID: null.IntFromPtr(&opt.VersionID),
	}

	if opt.Rank != ("") {
		q.Rank = opt.Rank
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

type ArtifactStatisticReport struct {
	DeleteList []*ArtifactRecycleBlobDesc
	TagList    []*ArtifactVersionCapacityDesc
}

type ArtifactRecycleBlobDesc struct {
	AssetID    int64  // 要硬删的 asset ID
	VersionID  int64  // asset 所属 version
	GroupID    int64  // asset 所属 group
	ArtifactID int64  // asset 所属制品库
	Path       string // asset 路径
	BlobID     int64  // asset 关联的 blob ID
	BlobRef    string // 物理文件 ref，DeleteBlob=true 时用于删文件
	DeleteBlob bool   // 是否连 blob 记录和物理文件一起删
	Size       int64  // blob size
}

type ArtifactVersionCapacityDesc struct {
	AssetID       int64 // 当前代码没有赋值使用
	VersionID     int64 // 聚合维度：version ID
	ExclusiveSize int64 // 独占容量
	TotalSize     int64 // 总容量
	ExclusiveRefs int   // 独占 asset 数
	TotalRefs     int   // 总 asset 数
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
	Accounts      []string         `json:"accounts"`
	AllowAccounts []string         `json:"allowAccounts"`
	DenyAccounts  []string         `json:"denyAccounts"`
	Role          enum.CasbinRole  `json:"role"`
	Scope         enum.CasbinScope `json:"scope"`
}

// RelocateEntityReq 移动并/或重命名实体请求。
// 语义：
//   - TargetGroupID 未传 → 不移动，即使 TargetArtifactID 有值也只执行重命名；
//   - TargetGroupID 已传，且 TargetArtifactID == 0 或与当前一致 → 同库移动；
//     TargetGroupID 显式传 0 → 移动到目标制品库根目录（仅 generic group 支持）；
//     两者皆不变即「不移动」。
//   - NewName == "" 或与当前一致 → 不改名。
//   - 移动 + 改名同时存在时，先移动后改名，整个过程在单事务内完成。
type RelocateEntityReq struct {
	EntityID         EntityID `json:"entityID"`         // 实体引用，形如 "group.11" / "asset.5"
	TargetArtifactID int64    `json:"targetArtifactID"` // 目标制品库 ID（可选）
	TargetGroupID    *int64   `json:"targetGroupID"`    // 目标 group ID（可选；generic group 显式传 0 表示根目录）
	NewName          string   `json:"newName"`          // 新名称（可选）
}

func (req *RelocateEntityReq) Validate() error {
	if err := req.EntityID.Validate(); err != nil {
		return err
	}
	return nil
}

// IsVersionedFormat 判断给定格式是否为版本化格式。
// 版本化格式（container / helm / pypi / npm / maven）采用 package + version 的分发语义，
// 其包名参与分发协议（如 pull URL），因此不允许重命名；组/包的结构层级也需要做特殊处理。
// 其余格式（generic / raw、以及 ArtifactAllFormat 或未知值）视为非版本化，返回 false。
// IsVersionedFormat reports whether the given artifact format is a versioned one.
func IsVersionedFormat(f ArtifactFormat) bool {
	switch f {
	case ArtifactContainerFormat, ArtifactHelmFormat, ArtifactPypiFormat, ArtifactNpmFormat, ArtifactMavenFormat:
		return true
	default:
		return false
	}
}
