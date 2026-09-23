// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package compiler

import (
	"github.com/easysoft/gitfox-community/pkg/runner/docker/engine2/engine"
)

const cleanStepName = "clean"

func createStageClean() *engine.Step {
	return &engine.Step{
		Name:      cleanStepName,
		Image:     "hub.zentao.net/ci/pipeline-helper:gitfox-community",
		Envs:      make(map[string]string),
		RunPolicy: engine.RunAlways,
	}
}
