// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package factory

import (
	"fmt"

	"github.com/easysoft/gitfox-community/pkg/contracts"
)

var _factory = map[string]contracts.PipelineStepBuildConverter{}

func RegisterConverter(name string, converter contracts.PipelineStepBuildConverter) {
	_factory[name] = converter
}

func GetConverter(name string) (contracts.PipelineStepBuildConverter, error) {
	converter, ok := _factory[name]
	if !ok {
		return nil, fmt.Errorf("project converter not found for %s", name)
	}
	return converter, nil
}
