// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package constant

// RunPolicy 流水线阶段运行策略
type RunPolicy string

const (
	RunPolicyOnSuccess RunPolicy = "onSuccess"
	RunPolicyOnFailure RunPolicy = "onFailure"
	RunPolicyAlways    RunPolicy = "always"
)

const (
	PipelineEngineGitFox  = "gitfox"
	PipelineEngineGitLab  = "gitlab"
	PipelineEngineJenkins = "jenkins"
)

// RunPolicyList 运行策略列表
var RunPolicyList = []string{
	string(RunPolicyOnSuccess),
	string(RunPolicyOnFailure),
	string(RunPolicyAlways),
}

// UserAgentPipeline is the HTTP user agent used by pipeline clone steps.
const UserAgentPipeline = "gitfox-pipeline"
