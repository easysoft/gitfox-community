// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package events

import (
	"errors"

	"github.com/easysoft/gitfox-community/pkg/events"
)

// Reporter is the event reporter for this package.
// It exposes typesafe send methods for all events of this package.
// NOTE: Event send methods are in the event's dedicated file.
type Reporter struct {
	innerReporter *events.GenericReporter
}

func NewReporter(eventsSystem *events.System) (*Reporter, error) {
	innerReporter, err := events.NewReporter(eventsSystem, category)
	if err != nil {
		return nil, errors.New("failed to create new GenericReporter from event system")
	}

	return &Reporter{
		innerReporter: innerReporter,
	}, nil
}
