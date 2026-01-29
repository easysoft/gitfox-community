package jobs

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pubsub"
)

const (
	PubSubTopicCancelJob   = "gitfox:job:cancel_job"
	PubSubTopicStateChange = "gitfox:job:state_change"
)

// publishStateChange publishes a StateChange message to the PubSub topic.
// 发布 StateChange 消息到 PubSub 主题
func publishStateChange(ctx context.Context, publisher pubsub.Publisher, payload []byte) error {
	err := publisher.Publish(ctx, PubSubTopicStateChange, payload)
	if err != nil {
		return errors.Wrapf(err, "failed to publish StateChange")
	}

	return nil
}
