package artifact

import (
	"fmt"

	enum "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"

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
	Path   string         `json:"path"`
	Format ArtifactFormat `json:"format"`
	Level  string         `json:"level"`
}

// ==================== 响应模型 ====================
// ==================== Response Types ====================

type ArtifactListItem struct {
	Format      string `gorm:"column:format" json:"format"`
	Name        string `gorm:"column:name" json:"name"`
	Namespace   string `gorm:"column:namespace" json:"namespace"`
	DisplayName string `gorm:"-" json:"display_name"`
	Version     string `gorm:"column:version" json:"version"`
	UpdateTime  int64  `gorm:"column:editedDate" json:"update_time"`
}


type ArtifactAssetsRes struct {
	Id             int64  `gorm:"column:id" json:"id"`
	Path           string `gorm:"column:path" json:"path"`
	ContentType    string `gorm:"column:contentType" json:"content_type"`
	Size           int64  `gorm:"column:size" json:"size"`
	BlobRef        string `gorm:"column:ref" json:"-"`
	Creator        string `gorm:"-" json:"creator"`
	CreatorName    string `gorm:"column:createdBy" json:"creator_name"`
	CheckSumString string `gorm:"column:checkSum" json:"-"`
	MetadataString string `gorm:"column:metadata" json:"-"`
	Created        int64  `gorm:"column:createdDate" json:"created"`
	Updated        int64  `gorm:"column:editedDate" json:"updated"`

	CheckSum *CheckSumRes      `gorm:"-" json:"checksum"`
	Metadata map[string]string `gorm:"-" json:"metadata"`
	Link     string            `gorm:"-" json:"link"`
	EntityId string            `gorm:"-" json:"entity_id"`
}

type ArtifactBlobRes struct {
	Id             int64    `gorm:"column:id" json:"id"`
	Path           string   `gorm:"column:path" json:"path"`
	Size           int64    `gorm:"column:size" json:"size"`
	BlobRef        string   `gorm:"column:ref" json:"-"`
	Creator        string   `gorm:"-" json:"creator"`
	CreatorName    string   `gorm:"column:createdBy" json:"creator_name"`
	CheckSumString string   `gorm:"column:checkSum" json:"-"`
	MetadataString string   `gorm:"column:metadata" json:"-"`
	Created        int64    `gorm:"column:createdDate" json:"created"`
	Deleted        null.Int `gorm:"column:deleted" json:"deleted"`

	CheckSum *CheckSumRes      `gorm:"-" json:"checksum"`
	Metadata map[string]string `gorm:"-" json:"metadata"`
	Link     string            `gorm:"-" json:"link"`
	EntityId string            `gorm:"-" json:"entity_id"`
}

type CheckSumRes struct {
	Md5    string `json:"md5"`
	Sha1   string `json:"sha1"`
	SHA256 string `json:"sha256"`
	SHA512 string `json:"sha512"`
}

type ArtifactVersionsRes struct {
	Version     string `json:"version"`
	CreatorName string `json:"creator_name"`
	Updated     int64  `json:"updated"`
}

// ==================== 树节点响应 ====================
// ==================== Tree Node Response ====================

type ArtifactTreeNodeType string

const (
	ArtifactTreeNodeTypeFormat    ArtifactTreeNodeType = "format"
	ArtifactTreeNodeTypeGroup     ArtifactTreeNodeType = "group"
	ArtifactTreeNodeTypeDirectory ArtifactTreeNodeType = "directory"
	ArtifactTreeNodeTypePackage   ArtifactTreeNodeType = "package"
	ArtifactTreeNodeTypeVersion   ArtifactTreeNodeType = "version"
	ArtifactTreeNodeTypeAsset     ArtifactTreeNodeType = "asset"
)

type ArtifactTreeRes struct {
	Name     string                `json:"name"`
	Path     string                `json:"path"`
	Leaf     bool                  `json:"leaf"`
	Format   ArtifactFormat        `json:"format"`
	Metadata *ArtifactTreeNodeMeta `json:"metadata,omitempty"`
}

type ArtifactTreeNodeMeta struct {
	Id      int64                `json:"id,omitempty"`
	Type    ArtifactTreeNodeType `json:"type"`
	Name    string               `json:"name,omitempty"`
	Group   string               `json:"group,omitempty"`
	Version string               `json:"version,omitempty"`

	// NodeId 已废弃，调用方应改用 EntityId。
	// Deprecated NodeId, use EntityId instead
	NodeId   string `json:"node_id,omitempty"`
	EntityId string `json:"entity_id,omitempty"`
}

type ArtifactTreeNodeId struct {
	Type ArtifactTreeNodeType `json:"type"`
	Pk   int64                `json:"pk"`
}

func (nid *ArtifactTreeNodeId) String() string {
	return fmt.Sprintf("%s.%d", nid.Type, nid.Pk)
}

// ==================== 节点详情响应 ====================
// ==================== Node Info ====================

type ArtifactStatus string

const (
	ArtifactStatusOK        ArtifactStatus = "ok"
	ArtifactStatusInvalidId ArtifactStatus = "invalid_id"
	ArtifactStatusNotFound  ArtifactStatus = "not found"
	ArtifactStatusUnknown   ArtifactStatus = "unknown"
)

type ArtifactNodeInfo struct {
	Status   string                  `json:"status"`
	Format   ArtifactFormat          `json:"format"`
	Metadata *ArtifactPkgVerMetadata `json:"metadata"`

	Path        string       `json:"path"`
	ContentType string       `json:"content_type"`
	Size        int64        `json:"size"`
	CreatorName string       `json:"creator_name"`
	Created     int64        `json:"created"`
	Updated     int64        `json:"updated"`
	Link        string       `json:"link"`
	CheckSum    *CheckSumRes `json:"checksum"`
}

type ArtifactPkgVerMetadata struct {
	Space   string `json:"space"`
	Name    string `json:"name"`
	Group   string `json:"group"`
	Version string `json:"version"`
}

// ==================== 删除与回收统计 ====================
// ==================== Remove / GC ====================

type ArtifactNodeRemoveRes struct {
	NodeId   string         `json:"node_id"`
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

type JobUIDResponse struct {
	UID string `json:"uid"`
}

type ArtifactGarbageReportRes struct {
	Count int64   `json:"count"`
	Size  int64   `json:"size"`
	Ids   []int64 `json:"ids"`
}

type ArtifactStatisticResponse struct {
	GarbageCollect *ArtifactGarbageReportRes     `json:"garbage_collect"`
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
	ExclusiveSize int64                        `json:"exclusive_size"`
	Versions      []ArtifactVersionCapacityRes `json:"versions"`
}

type ArtifactVersionCapacityRes struct {
	Version       string `json:"version"`
	Size          int64  `json:"size"`
	ExclusiveSize int64  `json:"exclusive_size"`
}

type ArtifactStatsObject struct {
	Total   int64 `json:"total"`
	Deleted int64 `json:"deleted"`
	Active  int64 `json:"active"`
}

// ==================== 权限响应 ====================
// ==================== Permission Response ====================

type ArtifactPermissions struct {
	Private bool                `json:"private"`
	Rules   []*ArtifactPermRule `json:"rules"`
}

type ArtifactEntityPermissions struct {
	ArtifactPermissions
	Inherit ArtifactPermInherit `json:"inherit"`
}

type ArtifactPermInherit struct {
	Inherited       bool     `json:"inherited"`
	InheritEntityId string   `json:"inherit_entity_id"`
	InheritFrom     string   `json:"inherit_from"`
	InheritParents  []string `json:"inherit_parents"`
}

type ArtifactPermRule struct {
	Subjects []*CasbinSubject  `json:"subjects"`
	Role     enum.CasbinRole   `json:"role"`
	Scope    enum.CasbinScope  `json:"scope"`
}
