// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package stepbuild

import (
	harness "github.com/easysoft/gitfox-community/pkg/pipeline/spec"
	"github.com/easysoft/gitfox-community/pkg/pipeline/step/factory"
)

func ConvertBuild(step *harness.Step) (*harness.StepBuild, error) {
	c, err := factory.GetConverter(step.Type)
	if err != nil {
		return nil, err
	}

	return c.Convert(step)
}
