// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package events

import (
	"context"

	"github.com/easysoft/gitfox-community/pkg/stream"
)

// StreamProducer is an abstraction of a producer from the streams package.
type StreamProducer interface {
	Send(ctx context.Context, streamID string, payload map[string]interface{}) (string, error)
}

// StreamConsumer is an abstraction of a consumer from the streams package.
type StreamConsumer interface {
	Register(streamID string, handler stream.HandlerFunc, opts ...stream.HandlerOption) error
	Configure(opts ...stream.ConsumerOption)
	Start(ctx context.Context) error
	Errors() <-chan error
	Infos() <-chan string
}

// StreamConsumerFactoryFunc is an abstraction of a factory method for stream consumers.
type StreamConsumerFactoryFunc func(groupName string, consumerName string) (StreamConsumer, error)
