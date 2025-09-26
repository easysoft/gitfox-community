package constants

// API 路径前缀（内部统一管理，避免硬编码）
const (
	// APIV1Prefix API v1 版本前缀
	APIV1Prefix = "/api/v1"
	// CodebasePath 代码库模块路径前缀
	CodebasePath = APIV1Prefix + "/repos"
	// PipelinePath 流水线模块路径前缀
	PipelinePath = APIV1Prefix + "/pipelines"
	// ArtifactPath 制品库模块路径前缀
	ArtifactPath = APIV1Prefix + "/artifacts"
)

// 内部缓存键前缀（避免缓存键冲突）
const (
	CacheBranchPrefix   = "devops:branch:"
	CachePipelinePrefix = "devops:pipeline:"
	CacheUserPrefix     = "devops:user:"
)
