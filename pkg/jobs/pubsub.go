// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package jobs

import (
	"context"

	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/pubsub"
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
