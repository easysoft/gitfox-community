package constant

// RunPolicy 流水线阶段运行策略
type RunPolicy string

const (
	RunPolicyOnSuccess RunPolicy = "onSuccess"
	RunPolicyOnFailure RunPolicy = "onFailure"
	RunPolicyAlways    RunPolicy = "always"
)

// RunPolicyList 运行策略列表
var RunPolicyList = []string{
	string(RunPolicyOnSuccess),
	string(RunPolicyOnFailure),
	string(RunPolicyAlways),
}
