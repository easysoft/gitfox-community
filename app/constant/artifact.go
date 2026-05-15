package constant

// ArtifactFormat 制品格式枚举。
type ArtifactFormat string

func (ArtifactFormat) Enum() []interface{}                      { return toInterfaceSlice(artifactFormats) }
func (s ArtifactFormat) Sanitize() (ArtifactFormat, bool)       { return Sanitize(s, GetAllArtifactFormats) }
func GetAllArtifactFormats() ([]ArtifactFormat, ArtifactFormat) { return artifactFormats, "" }

const (
	ArtifactFormatGeneric ArtifactFormat = "generic"
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
