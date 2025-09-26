package constants

// 分支相关状态码（仅内部业务使用，不对外暴露）
const (
	// BranchStatusPending 分支创建中
	BranchStatusPending = 1
	// BranchStatusSuccess 分支创建成功
	BranchStatusSuccess = 2
	// BranchStatusFailed 分支创建失败
	BranchStatusFailed = 3
	// BranchStatusDeleted 分支已删除
	BranchStatusDeleted = 4
)

// 流水线运行状态码
const (
	PipelineStatusRunning   = 10
	PipelineStatusSuccess   = 11
	PipelineStatusFailed    = 12
	PipelineStatusCancelled = 13
)

// 转换状态码为可读描述（内部辅助函数）
func GetBranchStatusDesc(status int) string {
	switch status {
	case BranchStatusPending:
		return "创建中"
	case BranchStatusSuccess:
		return "创建成功"
	case BranchStatusFailed:
		return "创建失败"
	case BranchStatusDeleted:
		return "已删除"
	default:
		return "未知状态"
	}
}
