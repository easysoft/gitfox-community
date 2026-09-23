// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package constant

// ArtifactFormat 制品格式枚举。
type ArtifactFormat string

func (ArtifactFormat) Enum() []interface{}                      { return toInterfaceSlice(artifactFormats) }
func (s ArtifactFormat) Sanitize() (ArtifactFormat, bool)       { return Sanitize(s, GetAllArtifactFormats) }
func GetAllArtifactFormats() ([]ArtifactFormat, ArtifactFormat) { return artifactFormats, "" }

const (
	ArtifactFormatGeneric ArtifactFormat = "file"
	ArtifactFormatRaw     ArtifactFormat = "raw"
	// ArtifactFormatMaven     ArtifactFormat = "maven"
	ArtifactFormatContainer ArtifactFormat = "container"
	// ArtifactFormatHelm      ArtifactFormat = "helm"
	// ArtifactFormatPypi      ArtifactFormat = "pypi"
	// ArtifactFormatNpm       ArtifactFormat = "npm"
)

var artifactFormats = sortEnum([]ArtifactFormat{
	ArtifactFormatGeneric,
	ArtifactFormatRaw,
	// ArtifactFormatMaven,
	ArtifactFormatContainer,
	// ArtifactFormatHelm,
	// ArtifactFormatPypi,
	// ArtifactFormatNpm,
})

// ArtifactEntityKind 制品格式实体类型枚举。
type ArtifactEntityKind string

const (
	ArtifactEntityKindFormat   ArtifactEntityKind = "format"
	ArtifactEntityKindGroup    ArtifactEntityKind = "group"
	ArtifactEntityKindPackage  ArtifactEntityKind = "package"
	ArtifactEntityKindVersion  ArtifactEntityKind = "version"
	ArtifactEntityKindAsset    ArtifactEntityKind = "asset"
	ArtifactEntityKindArtifact ArtifactEntityKind = "artifact"
)

func (ArtifactEntityKind) Enum() []interface{} { return toInterfaceSlice(artifactFormats) }
func (s ArtifactEntityKind) Sanitize() (ArtifactEntityKind, bool) {
	return Sanitize(s, GetAllArtifactEntityKinds)
}
func GetAllArtifactEntityKinds() ([]ArtifactEntityKind, ArtifactEntityKind) {
	return artifactEntityKinds, ""
}

var artifactEntityKinds = sortEnum([]ArtifactEntityKind{
	ArtifactEntityKindFormat,
	ArtifactEntityKindGroup,
	ArtifactEntityKindPackage,
	ArtifactEntityKindVersion,
	ArtifactEntityKindAsset,
	ArtifactEntityKindArtifact,
})
