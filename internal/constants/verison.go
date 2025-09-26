package constants

// 版本标识常量（统一管理所有版本名称，避免硬编码）
const (
	// VersionBase 免费版版本标识
	VersionBase = "pms"
	// VersionEnterprise 企业版版本标识
	VersionEnterprise = "biz"
	// VersionUltimate 旗舰版版本标识
	VersionUltimate = "ipd"
)

// 所有支持的版本列表（用于参数校验）
var SupportedVersions = []string{
	VersionBase,
	VersionEnterprise,
	VersionUltimate,
}
