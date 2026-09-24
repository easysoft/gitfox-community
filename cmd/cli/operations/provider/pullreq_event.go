// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package provider

import (
	pullreqevents "github.com/easysoft/gitfox-community/app/events/pullreq"
	"github.com/easysoft/gitfox-community/pkg/events"

	"github.com/google/wire"
)

// PullReqEventWireSet provides a wire set for this package.
var PullReqEventWireSet = wire.NewSet(
	ProvidePullReqReaderFactory,
	ProvidePullReqReporter,
)

func ProvidePullReqReaderFactory(eventsSystem *events.System) (*events.ReaderFactory[*pullreqevents.Reader], error) {
	return pullreqevents.NewReaderFactory(eventsSystem)
}

func ProvidePullReqReporter(eventsSystem *events.System) (*pullreqevents.Reporter, error) {
	return pullreqevents.NewReporter(eventsSystem)
}
