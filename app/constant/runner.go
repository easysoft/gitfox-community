// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package constant

// RunnerStatus Runner 在线/离线状态
type RunnerStatus string

const (
	RunnerOnline  RunnerStatus = "online"
	RunnerOffline RunnerStatus = "offline"
)

// RunnerActiveStatus Runner 启用/禁用状态
type RunnerActiveStatus string

const (
	RunnerEnable  RunnerActiveStatus = "enable"
	RunnerDisable RunnerActiveStatus = "disable"
)
