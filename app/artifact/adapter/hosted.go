package adapter

import (
	"context"
	"net/http"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
)

type CommitWriter interface {
	Commit(ctx context.Context) error
	Cancel(ctx context.Context) error
}

type ArtifactPackageUploader interface {
	Serve(ctx context.Context, req *http.Request) (int64, error)
	IsValid(ctx context.Context) error
	Cancel(ctx context.Context) error
	Save(ctx context.Context) error
	GetAssets() []*SavedAsset
}

// ArtifactIndexUpdater 定义格式适配器在上传后刷新索引的能力。
// ArtifactIndexUpdater is an index update interface
// 每个 format-adapter 都应实现该接口。
// every format-adapter should implement it
type ArtifactIndexUpdater interface {
	// UpdatePackage 在新版本上传完成后刷新 package 级索引。
	// UpdatePackage will be called after a new version of package is uploaded
	UpdatePackage(ctx context.Context, p *model.ArtifactPackage, v *model.ArtifactView) error

	// UpdateRepo 在新版本上传完成后刷新 repository / view 级索引。
	// UpdateRepo will be called after a new version of package is uploaded
	UpdateRepo(ctx context.Context, v *model.ArtifactView) error
	CommitWriter
}

type PackageDescriptor struct {
	Name            string
	Namespace       string
	GroupID         int64
	Version         string
	VersionMetadata MetadataInterface
	Format          model.ArtifactFormat
	MainAsset       *AssetDescriptor
	SubAssets       []*AssetDescriptor
	Overwrite       bool
	Rename          bool
}

func NewEmptyPackageDescriptor() *PackageDescriptor {
	return &PackageDescriptor{MainAsset: &AssetDescriptor{}, SubAssets: make([]*AssetDescriptor, 0)}
}

func (p *PackageDescriptor) AddSub(asset *AssetDescriptor) {
	p.SubAssets = append(p.SubAssets, asset)
}

type AssetDescriptor struct {
	Attr        ModelAttribute
	Path        string
	Ref         string
	Size        int64
	Metadata    MetadataInterface
	ContentType string
	Format      model.ArtifactFormat
	Rank        model.AssetRank
	Hash        *Hash
}

type PackageMetaDescriptor struct {
	Name      string
	Format    model.ArtifactFormat
	MainAsset *AssetDescriptor
	SubAssets []*AssetDescriptor
}

func NewEmptyPackageMetaDescriptor() *PackageMetaDescriptor {
	return &PackageMetaDescriptor{MainAsset: &AssetDescriptor{}, SubAssets: make([]*AssetDescriptor, 0)}
}

func (p *PackageMetaDescriptor) AddSub(asset *AssetDescriptor) {
	p.SubAssets = append(p.SubAssets, asset)
}

type ModelAttribute string

const (
	AttrAssetNormal    ModelAttribute = "normal"
	AttrAssetIsolated  ModelAttribute = "isolated"
	AttrAssetExclusive ModelAttribute = "exclusive"
	AttrAssetIndex     ModelAttribute = "index"
)
