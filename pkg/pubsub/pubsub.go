// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package pubsub

import "context"

type Publisher interface {
	// Publish topic to message broker with payload.
	Publish(ctx context.Context, topic string, payload []byte,
		options ...PublishOption) error
}

type PubSub interface {
	Publisher
	// Subscribe consumer to process the topic with payload, this should be
	// blocking operation.
	Subscribe(ctx context.Context, topic string,
		handler func(payload []byte) error, options ...SubscribeOption) Consumer
}

type Consumer interface {
	Subscribe(ctx context.Context, topics ...string) error
	Unsubscribe(ctx context.Context, topics ...string) error
	Close() error
}
