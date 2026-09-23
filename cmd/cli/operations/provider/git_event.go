// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package provider

import (
	gitevents "github.com/easysoft/gitfox-community/app/events/git"
	"github.com/easysoft/gitfox-community/pkg/events"

	"github.com/google/wire"
)

// GitEventWireSet provides a wire set for this package.
var GitEventWireSet = wire.NewSet(
	ProvideGitEventReaderFactory,
	ProvideGitEventReporter,
)

func ProvideGitEventReaderFactory(eventsSystem *events.System) (*events.ReaderFactory[*gitevents.Reader], error) {
	return gitevents.NewReaderFactory(eventsSystem)
}

func ProvideGitEventReporter(eventsSystem *events.System) (*gitevents.Reporter, error) {
	return gitevents.NewReporter(eventsSystem)
}
