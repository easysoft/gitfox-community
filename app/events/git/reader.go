// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package events

import (
	"github.com/easysoft/gitfox-community/pkg/events"
)

func NewReaderFactory(eventsSystem *events.System) (*events.ReaderFactory[*Reader], error) {
	readerFactoryFunc := func(innerReader *events.GenericReader) (*Reader, error) {
		return &Reader{
			innerReader: innerReader,
		}, nil
	}

	return events.NewReaderFactory(eventsSystem, category, readerFactoryFunc)
}

// Reader is the event reader for this package.
// It exposes typesafe event registration methods for all events by this package.
// NOTE: Event registration methods are in the event's dedicated file.
type Reader struct {
	innerReader *events.GenericReader
}

func (r *Reader) Configure(opts ...events.ReaderOption) {
	r.innerReader.Configure(opts...)
}
