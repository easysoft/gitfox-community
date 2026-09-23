// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

type StepBuildAndPushDockerImage struct {
	Mode              string              `json:"mode,omitempty"`
	DriverOpt         string              `json:"driverOpt,omitempty"`
	Dockerfile        string              `json:"dockerfile,omitempty"`
	Context           string              `json:"context,omitempty"`
	Platforms         []string            `json:"platforms,omitempty"`
	BuildArgs         map[string]string   `json:"buildArgs,omitempty"`
	Labels            map[string]string   `json:"labels,omitempty"`
	Repository        *DockerRepository   `json:"repository,omitempty"`
	ExtraRepositories []*DockerRepository `json:"extraRepositories,omitempty"`
}
