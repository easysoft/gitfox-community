package constants

// 版本标识常量（统一管理所有版本名称，避免硬编码）
const (
	// VersionBase 免费版版本标识
	VersionBase = "lite"
)

// 所有支持的版本列表（用于参数校验）
var SupportedVersions = []string{
	VersionBase,
}
