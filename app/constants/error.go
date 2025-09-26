package constants

// 内部通用错误信息（与业务状态码对应，统一错误文案）
const (
	// ErrBranchNotFound 分支不存在
	ErrBranchNotFound = "分支不存在或已删除"
	// ErrRepoNoPermission 代码库无操作权限
	ErrRepoNoPermission = "当前用户无此代码库操作权限"
	// ErrPipelineRunning 流水线正在运行（不可重复触发）
	ErrPipelineRunning = "流水线正在运行中，请勿重复触发"
	// ErrParamInvalid 参数无效
	ErrParamInvalid = "请求参数无效，请检查后重试"
)
