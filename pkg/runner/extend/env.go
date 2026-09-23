// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package extend

import (
	"fmt"

	"github.com/drone/drone-go/drone"
)

const (
	WorkSpace = "/gitfox"

	// StageTempDir is the temporary directory in a stage life cycle
	StageTempDir = WorkSpace + "/.stage"
)

func Envs(stage *drone.Stage, repo *drone.Repo) map[string]string {
	environ := map[string]string{}

	// dynamic environment file for current stage.
	environ["GITFOX_CUSTOM_ENV"] = StageTempDir + fmt.Sprintf("/%d/custom_env", stage.ID)
	environ["GITFOX_EXECUTION_ID"] = fmt.Sprintf("%d", stage.BuildID)
	environ["GITFOX_STAGE_ID"] = fmt.Sprintf("%d", stage.ID)

	// GITFOX_API_ENDPOINT / PLUGIN_TOKEN 在触发时由 triggerer 统一注入
	// （见 app/pipeline/triggerer/trigger.go），此处不再拼接，避免覆盖正确值。

	return environ
}
