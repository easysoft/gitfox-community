// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

type StepWinQtBuild struct {
	Tool          string `json:"tool,omitempty"`
	Configuration string `json:"configuration,omitempty"`
	Arch          string `json:"arch,omitempty"`
	CustomFlags   string `json:"customFlags,omitempty"`
	OutputDir     string `json:"outputDir,omitempty"`
	FileName      string `json:"fileName,omitempty"`
}
