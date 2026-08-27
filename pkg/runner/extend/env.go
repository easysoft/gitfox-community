// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

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
