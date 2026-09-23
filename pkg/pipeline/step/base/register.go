// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package base

import "github.com/easysoft/gitfox-community/pkg/pipeline/step/factory"

func Register() {
	factory.RegisterConverter("build_go", &buildGoStep{})
	factory.RegisterConverter("build_yarn", &buildYarnStep{})
	factory.RegisterConverter("build_maven", &buildMavenStep{})
	factory.RegisterConverter("build_gradle", &buildGradleStep{})
	factory.RegisterConverter("build_npm", &buildNpmStep{})
}
