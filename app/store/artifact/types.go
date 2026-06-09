package artifact

import (
	"time"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

// ==================== ORM 实体定义 ====================
// 以下 struct 与数据库表列一一映射，仅在 store 包内使用。
// 时间字段使用 *datetime.Local，与系统其他 store 层规范一致。
// service 层及以上使用 app/model/artifact 下的 struct（时间为 *time.Time）。

// artifactView → ops_artifact_libs
type artifactView struct {
	ID          int64           `gorm:"column:id;primaryKey"`
	Name        string          `gorm:"column:name"`
	Description string          `gorm:"column:desc"`
	SpaceID     int64           `gorm:"column:spaceID"`
	RepoID      int64           `gorm:"column:repoID"`
	Format      string          `gorm:"column:type"`
	Type        string          `gorm:"column:scope"`
	CreatedBy   string          `gorm:"column:createdBy"`
	CreatedDate *datetime.Local `gorm:"column:createdDate"`
	EditedBy    string          `gorm:"column:editedBy"`
	EditedDate  *datetime.Local `gorm:"column:editedDate"`
	Deleted     int8            `gorm:"column:deleted"`
}

func (artifactView) TableName() string { return "ops_artifact_libs" }

func (v *artifactView) BeforeCreate(_ *gorm.DB) error {
	now := datetime.NowLocalTime()
	v.CreatedDate = now
	v.EditedDate = now
	return nil
}

func (v *artifactView) BeforeUpdate(_ *gorm.DB) error {
	v.EditedDate = datetime.NowLocalTime()
	return nil
}

// artifactGroup → ops_artifact_groups
type artifactGroup struct {
	ID          int64           `gorm:"column:id;primaryKey"`
	ParentID    null.Int        `gorm:"column:parentID"`
	ArtifactID  int64           `gorm:"column:artifactLibID"`
	Name        string          `gorm:"column:name"`
	Format      string          `gorm:"column:type"`
	CreatedBy   string          `gorm:"column:createdBy"`
	CreatedDate *datetime.Local `gorm:"column:createdDate"`
	EditedBy    string          `gorm:"column:editedBy"`
	EditedDate  *datetime.Local `gorm:"column:editedDate"`
	Deleted     int8            `gorm:"column:deleted"`
}

func (artifactGroup) TableName() string { return "ops_artifact_groups" }

func (g *artifactGroup) BeforeCreate(_ *gorm.DB) error {
	now := datetime.NowLocalTime()
	g.CreatedDate = now
	g.EditedDate = now
	return nil
}

func (g *artifactGroup) BeforeUpdate(_ *gorm.DB) error {
	g.EditedDate = datetime.NowLocalTime()
	return nil
}

// artifactPackage → ops_artifact_packages
type artifactPackage struct {
	ID          int64           `gorm:"column:id;primaryKey"`
	ArtifactID  int64           `gorm:"column:artifactLibID"`
	Name        string          `gorm:"column:name"`
	Namespace   string          `gorm:"column:namespace"`
	GroupID     int64           `gorm:"column:groupID"`
	Format      string          `gorm:"column:type"`
	CreatedBy   string          `gorm:"column:createdBy"`
	CreatedDate *datetime.Local `gorm:"column:createdDate"`
	EditedBy    string          `gorm:"column:editedBy"`
	EditedDate  *datetime.Local `gorm:"column:editedDate"`
	Deleted     int8            `gorm:"column:deleted"`
}

func (artifactPackage) TableName() string { return "ops_artifact_packages" }

func (p *artifactPackage) BeforeCreate(_ *gorm.DB) error {
	now := datetime.NowLocalTime()
	p.CreatedDate = now
	p.EditedDate = now
	return nil
}

func (p *artifactPackage) BeforeUpdate(_ *gorm.DB) error {
	p.EditedDate = datetime.NowLocalTime()
	return nil
}

// artifactVersion → ops_artifact_versions
type artifactVersion struct {
	ID            int64           `gorm:"column:id;primaryKey"`
	PackageID     int64           `gorm:"column:packageID"`
	ArtifactLibID int64           `gorm:"column:artifactLibID"`
	Version       string          `gorm:"column:version"`
	Metadata      string          `gorm:"-"`
	CreatedBy     string          `gorm:"column:createdBy"`
	CreatedDate   *datetime.Local `gorm:"column:createdDate"`
	EditedBy      string          `gorm:"column:editedBy"`
	EditedDate    *datetime.Local `gorm:"column:editedDate"`
	Deleted       int8            `gorm:"column:deleted"`
}

func (artifactVersion) TableName() string { return "ops_artifact_versions" }

func (v *artifactVersion) BeforeCreate(_ *gorm.DB) error {
	now := datetime.NowLocalTime()
	v.CreatedDate = now
	v.EditedDate = now
	return nil
}

func (v *artifactVersion) BeforeUpdate(_ *gorm.DB) error {
	v.EditedDate = datetime.NowLocalTime()
	return nil
}

// artifactAsset → ops_artifact_assets
type ArtifactAsset struct {
	ID          int64           `gorm:"column:id;primaryKey"`
	VersionID   null.Int        `gorm:"column:versionID"`
	ArtifactID  null.Int        `gorm:"column:artifactLibID"`
	GroupID     null.Int        `gorm:"column:groupID"`
	Format      string          `gorm:"column:type"`
	Path        string          `gorm:"column:basename"`
	ContentType string          `gorm:"column:contentType"`
	Rank        string          `gorm:"column:rank"`
	BlobID      int64           `gorm:"column:blobID"`
	CheckSum    string          `gorm:"column:checksum"`
	CreatedBy   string          `gorm:"column:createdBy"`
	CreatedDate *datetime.Local `gorm:"column:createdDate"`
	EditedBy    string          `gorm:"column:editedBy"`
	EditedDate  *datetime.Local `gorm:"column:editedDate"`
	Deleted     int8            `gorm:"column:deleted"`
}

func (ArtifactAsset) TableName() string { return "ops_artifact_assets" }

func (a *ArtifactAsset) BeforeCreate(_ *gorm.DB) error {
	now := datetime.NowLocalTime()
	a.CreatedDate = now
	a.EditedDate = now
	return nil
}

func (a *ArtifactAsset) BeforeUpdate(_ *gorm.DB) error {
	a.EditedDate = datetime.NowLocalTime()
	return nil
}

type assetExtendRow struct {
	ArtifactAsset
	StorageID int64  `gorm:"column:storageID"`
	Ref       string `gorm:"column:ref"`
	Size      int64  `gorm:"column:size"`
}

// artifactMetaAsset → ops_artifact_meta_assets
type artifactMetaAsset struct {
	ID            int64           `gorm:"column:id;primaryKey"`
	ArtifactLibID int64           `gorm:"column:artifactLibID"`
	Format        string          `gorm:"column:type"`
	Path          string          `gorm:"column:path"`
	ContentType   string          `gorm:"column:contentType"`
	Rank          string          `gorm:"column:rank"`
	BlobID        int64           `gorm:"column:blobID"`
	CheckSum      string          `gorm:"column:checksum"`
	CreatedBy     string          `gorm:"column:createdBy"`
	CreatedDate   *datetime.Local `gorm:"column:createdDate"`
	EditedBy      string          `gorm:"column:editedBy"`
	EditedDate    *datetime.Local `gorm:"column:editedDate"`
	Deleted       int8            `gorm:"column:deleted"`
}

func (artifactMetaAsset) TableName() string { return "ops_artifact_meta_assets" }

func (m *artifactMetaAsset) BeforeCreate(_ *gorm.DB) error {
	now := datetime.NowLocalTime()
	m.CreatedDate = now
	m.EditedDate = now
	return nil
}

func (m *artifactMetaAsset) BeforeUpdate(_ *gorm.DB) error {
	m.EditedDate = datetime.NowLocalTime()
	return nil
}

// artifactBlob → ops_artifact_blobs
type artifactBlob struct {
	ID          int64           `gorm:"column:id;primaryKey"`
	StorageID   int64           `gorm:"column:storageID"`
	Ref         string          `gorm:"column:ref"`
	Size        int64           `gorm:"column:size"`
	Downloads   int64           `gorm:"column:downloads"`
	Metadata    string          `gorm:"column:metadata"`
	Deleted     int8            `gorm:"column:deleted"`
	CreatedBy   string          `gorm:"column:createdBy"`
	CreatedDate *datetime.Local `gorm:"column:createdDate"`
	EditedBy    string          `gorm:"column:editedBy"`
	EditedDate  *datetime.Local `gorm:"column:editedDate"`
	AssetID     int64           `gorm:"column:assetID"`
	CheckSum    string          `gorm:"column:checksum"`
}

func (artifactBlob) TableName() string { return "ops_artifact_blobs" }

func (b *artifactBlob) BeforeCreate(_ *gorm.DB) error {
	now := datetime.NowLocalTime()
	b.CreatedDate = now
	b.EditedDate = now
	return nil
}

func (b *artifactBlob) BeforeUpdate(_ *gorm.DB) error {
	b.EditedDate = datetime.NowLocalTime()
	return nil
}

// artifactTreeNode → ops_artifact_tree_nodes
// type 承载制品格式（generic/container/raw...），linkTable 承载树节点类型
// （format/group/package/version/asset）。linkTable 原名来自旧表设计，这里复用为
// 节点类型列，保证接口能稳定返回 entityID 前缀并按层级追加资产。
type artifactTreeNode struct {
	ID         int64    `gorm:"column:id;primaryKey"`
	ParentID   null.Int `gorm:"column:parentID"`
	ArtifactID int64    `gorm:"column:artifactLibID"`
	Name       string   `gorm:"column:nodeName"`
	Path       string   `gorm:"column:nodePath"`
	Type       string   `gorm:"column:linkTable"`
	Format     string   `gorm:"column:type"`
	TargetID   null.Int `gorm:"column:linkRecord"`
}

func (artifactTreeNode) TableName() string { return "ops_artifact_tree_nodes" }

// ==================== 转换辅助函数 ====================
// localToTime 将 *datetime.Local 转换为 *time.Time（供 toModel 使用）
func localToTime(l *datetime.Local) *time.Time {
	if l == nil {
		return nil
	}
	t := time.Time(*l)
	return &t
}

// timeToLocal 将 *time.Time 转换为 *datetime.Local（供 fromModel 使用）
func timeToLocal(t *time.Time) *datetime.Local {
	if t == nil {
		return nil
	}
	l := datetime.Local(*t)
	return &l
}

// ==================== view 转换 ====================

func viewToModel(v *artifactView) *model.ArtifactView {
	if v == nil {
		return nil
	}
	return &model.ArtifactView{
		ID:          v.ID,
		Name:        v.Name,
		Description: v.Description,
		SpaceID:     v.SpaceID,
		RepoID:      v.RepoID,
		Format:      model.ArtifactFormat(v.Format),
		Type:        model.ArtifactViewType(v.Type),
		CreatedBy:   v.CreatedBy,
		CreatedDate: v.CreatedDate,
		EditedBy:    v.EditedBy,
		EditedDate:  v.EditedDate,
		Deleted:     v.Deleted,
	}
}

// ==================== group 转换 ====================

func groupToModel(g *artifactGroup) *model.ArtifactGroup {
	if g == nil {
		return nil
	}
	return &model.ArtifactGroup{
		ID:          g.ID,
		ParentID:    g.ParentID,
		ArtifactID:  g.ArtifactID,
		Name:        g.Name,
		Format:      model.ArtifactFormat(g.Format),
		CreatedBy:   g.CreatedBy,
		CreatedDate: g.CreatedDate,
		EditedBy:    g.EditedBy,
		EditedDate:  g.EditedDate,
		Deleted:     g.Deleted,
	}
}

func groupFromModel(m *model.ArtifactGroup) *artifactGroup {
	if m == nil {
		return nil
	}
	return &artifactGroup{
		ID:          m.ID,
		ParentID:    m.ParentID,
		ArtifactID:  m.ArtifactID,
		Name:        m.Name,
		Format:      string(m.Format),
		CreatedBy:   m.CreatedBy,
		CreatedDate: m.CreatedDate,
		EditedBy:    m.EditedBy,
		EditedDate:  m.EditedDate,
		Deleted:     m.Deleted,
	}
}

func groupsToModel(gs []*artifactGroup) []*model.ArtifactGroup {
	result := make([]*model.ArtifactGroup, 0, len(gs))
	for _, g := range gs {
		result = append(result, groupToModel(g))
	}
	return result
}

// ==================== package 转换 ====================

func packageToModel(p *artifactPackage) *model.ArtifactPackage {
	if p == nil {
		return nil
	}
	return &model.ArtifactPackage{
		ID:          p.ID,
		ArtifactID:  p.ArtifactID,
		Name:        p.Name,
		Namespace:   p.Namespace,
		GroupID:     p.GroupID,
		Format:      model.ArtifactFormat(p.Format),
		CreatedBy:   p.CreatedBy,
		CreatedDate: localToTime(p.CreatedDate),
		EditedBy:    p.EditedBy,
		EditedDate:  localToTime(p.EditedDate),
		Deleted:     p.Deleted,
	}
}

func packageFromModel(m *model.ArtifactPackage) *artifactPackage {
	if m == nil {
		return nil
	}
	return &artifactPackage{
		ID:          m.ID,
		ArtifactID:  m.ArtifactID,
		Name:        m.Name,
		Namespace:   m.Namespace,
		GroupID:     m.GroupID,
		Format:      string(m.Format),
		CreatedBy:   m.CreatedBy,
		CreatedDate: timeToLocal(m.CreatedDate),
		EditedBy:    m.EditedBy,
		EditedDate:  timeToLocal(m.EditedDate),
		Deleted:     m.Deleted,
	}
}

func packagesToModel(ps []*artifactPackage) []*model.ArtifactPackage {
	result := make([]*model.ArtifactPackage, 0, len(ps))
	for _, p := range ps {
		result = append(result, packageToModel(p))
	}
	return result
}

// ==================== version 转换 ====================

func versionToModel(v *artifactVersion) *model.ArtifactVersion {
	if v == nil {
		return nil
	}
	return &model.ArtifactVersion{
		ID:            v.ID,
		ArtifactLibID: v.ArtifactLibID,
		PackageID:     v.PackageID,
		Version:       v.Version,
		Metadata:      v.Metadata,
		CreatedBy:     v.CreatedBy,
		CreatedDate:   v.CreatedDate,
		EditedBy:      v.EditedBy,
		EditedDate:    v.EditedDate,
		Deleted:       v.Deleted,
	}
}

func versionFromModel(m *model.ArtifactVersion) *artifactVersion {
	if m == nil {
		return nil
	}
	return &artifactVersion{
		ID:            m.ID,
		ArtifactLibID: m.ArtifactLibID,
		PackageID:     m.PackageID,
		Version:       m.Version,
		Metadata:      m.Metadata,
		CreatedBy:     m.CreatedBy,
		CreatedDate:   m.CreatedDate,
		EditedBy:      m.EditedBy,
		EditedDate:    m.EditedDate,
		Deleted:       m.Deleted,
	}
}

func versionsToModel(vs []*artifactVersion) []*model.ArtifactVersion {
	result := make([]*model.ArtifactVersion, 0, len(vs))
	for _, v := range vs {
		result = append(result, versionToModel(v))
	}
	return result
}

// ==================== asset 转换 ====================

func assetToModel(a *ArtifactAsset) *model.ArtifactAsset {
	if a == nil {
		return nil
	}
	return &model.ArtifactAsset{
		ID:          a.ID,
		VersionID:   a.VersionID,
		ArtifactID:  a.ArtifactID,
		GroupID:     a.GroupID,
		Format:      model.ArtifactFormat(a.Format),
		Path:        a.Path,
		ContentType: a.ContentType,
		Rank:        model.AssetRank(a.Rank),
		BlobID:      a.BlobID,
		CheckSum:    a.CheckSum,
		CreatedBy:   a.CreatedBy,
		CreatedDate: a.CreatedDate,
		EditedBy:    a.EditedBy,
		EditedDate:  a.EditedDate,
		Deleted:     a.Deleted,
	}
}

func assetFromModel(m *model.ArtifactAsset) *ArtifactAsset {
	if m == nil {
		return nil
	}
	return &ArtifactAsset{
		ID:          m.ID,
		VersionID:   m.VersionID,
		ArtifactID:  m.ArtifactID,
		GroupID:     m.GroupID,
		Format:      string(m.Format),
		Path:        m.Path,
		ContentType: m.ContentType,
		Rank:        string(m.Rank),
		BlobID:      m.BlobID,
		CheckSum:    m.CheckSum,
		CreatedBy:   m.CreatedBy,
		CreatedDate: m.CreatedDate,
		EditedBy:    m.EditedBy,
		EditedDate:  m.EditedDate,
		Deleted:     m.Deleted,
	}
}

func assetsToModel(as []*ArtifactAsset) []*model.ArtifactAsset {
	result := make([]*model.ArtifactAsset, 0, len(as))
	for _, a := range as {
		result = append(result, assetToModel(a))
	}
	return result
}

// ==================== metaAsset 转换 ====================

func metaAssetToModel(m *artifactMetaAsset) *model.ArtifactMetaAsset {
	if m == nil {
		return nil
	}
	return &model.ArtifactMetaAsset{
		ID:          m.ID,
		ArtifactID:  m.ArtifactLibID,
		Format:      model.ArtifactFormat(m.Format),
		Path:        m.Path,
		ContentType: m.ContentType,
		Rank:        model.AssetRank(m.Rank),
		BlobID:      m.BlobID,
		CheckSum:    m.CheckSum,
		CreatedBy:   m.CreatedBy,
		CreatedDate: localToTime(m.CreatedDate),
		EditedBy:    m.EditedBy,
		EditedDate:  localToTime(m.EditedDate),
		Deleted:     m.Deleted,
	}
}

func metaAssetFromModel(src *model.ArtifactMetaAsset) *artifactMetaAsset {
	if src == nil {
		return nil
	}
	return &artifactMetaAsset{
		ID:            src.ID,
		ArtifactLibID: src.ArtifactID,
		Format:        string(src.Format),
		Path:          src.Path,
		ContentType:   src.ContentType,
		Rank:          string(src.Rank),
		BlobID:        src.BlobID,
		CheckSum:      src.CheckSum,
		CreatedBy:     src.CreatedBy,
		CreatedDate:   timeToLocal(src.CreatedDate),
		EditedBy:      src.EditedBy,
		EditedDate:    timeToLocal(src.EditedDate),
		Deleted:       src.Deleted,
	}
}

// ==================== blob 转换 ====================

func blobToModel(b *artifactBlob) *model.ArtifactBlob {
	if b == nil {
		return nil
	}
	return &model.ArtifactBlob{
		ID:          b.ID,
		StorageID:   b.StorageID,
		Ref:         b.Ref,
		Size:        b.Size,
		Downloads:   b.Downloads,
		Metadata:    b.Metadata,
		Deleted:     b.Deleted,
		CreatedBy:   b.CreatedBy,
		CreatedDate: localToTime(b.CreatedDate),
		EditedBy:    b.EditedBy,
		EditedDate:  localToTime(b.EditedDate),
		AssetID:     b.AssetID,
		CheckSum:    b.CheckSum,
	}
}

func blobFromModel(m *model.ArtifactBlob) *artifactBlob {
	if m == nil {
		return nil
	}
	return &artifactBlob{
		ID:          m.ID,
		StorageID:   m.StorageID,
		Ref:         m.Ref,
		Size:        m.Size,
		Downloads:   m.Downloads,
		Metadata:    m.Metadata,
		Deleted:     m.Deleted,
		CreatedBy:   m.CreatedBy,
		CreatedDate: timeToLocal(m.CreatedDate),
		EditedBy:    m.EditedBy,
		EditedDate:  timeToLocal(m.EditedDate),
		AssetID:     m.AssetID,
		CheckSum:    m.CheckSum,
	}
}

// ==================== treeNode 转换 ====================

func nodeToModel(n *artifactTreeNode) *model.ArtifactTreeNode {
	if n == nil {
		return nil
	}
	nodeType := model.ArtifactTreeNodeType(n.Type)
	if nodeType == "" {
		nodeType = inferTreeNodeType(n)
	}
	return &model.ArtifactTreeNode{
		ID:         n.ID,
		ParentID:   n.ParentID,
		ArtifactID: n.ArtifactID,
		Name:       n.Name,
		Path:       n.Path,
		Type:       nodeType,
		Format:     model.ArtifactFormat(n.Format),
		TargetID:   n.TargetID,
	}
}

func inferTreeNodeType(n *artifactTreeNode) model.ArtifactTreeNodeType {
	if n.Path == "/" {
		return model.ArtifactTreeNodeTypeFormat
	}
	if n.TargetID.Valid && model.ArtifactFormat(n.Format) == model.ArtifactFileFormat {
		return model.ArtifactTreeNodeTypeGroup
	}
	return ""
}

func nodeFromModel(m *model.ArtifactTreeNode) *artifactTreeNode {
	if m == nil {
		return nil
	}
	return &artifactTreeNode{
		ID:         m.ID,
		ParentID:   m.ParentID,
		ArtifactID: m.ArtifactID,
		Name:       m.Name,
		Path:       m.Path,
		Type:       string(m.Type),
		Format:     string(m.Format),
		TargetID:   m.TargetID,
	}
}

func nodesToModel(ns []*artifactTreeNode) []*model.ArtifactTreeNode {
	result := make([]*model.ArtifactTreeNode, 0, len(ns))
	for _, n := range ns {
		result = append(result, nodeToModel(n))
	}
	return result
}
