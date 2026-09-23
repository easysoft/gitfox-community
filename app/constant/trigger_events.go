// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package constant

// TriggerEvent defines the different kinds of events in triggers.
type TriggerEvent string

// Hook event constants.
const (
	TriggerEventCron        TriggerEvent = "cron"
	TriggerEventManual      TriggerEvent = "manual"
	TriggerEventPush        TriggerEvent = "push"
	TriggerEventPullRequest TriggerEvent = "pull_request"
	TriggerEventTag         TriggerEvent = "tag"
)

// Enum returns all possible TriggerEvent values.
func (TriggerEvent) Enum() []interface{} {
	return toInterfaceSlice(triggerEvents)
}

// Sanitize validates and returns a sanitized TriggerEvent value.
func (event TriggerEvent) Sanitize() (TriggerEvent, bool) {
	return Sanitize(event, GetAllTriggerEvents)
}

// GetAllTriggerEvents returns all possible TriggerEvent values and a default value.
func GetAllTriggerEvents() ([]TriggerEvent, TriggerEvent) {
	return triggerEvents, TriggerEventManual
}

// List of all TriggerEvent values.
var triggerEvents = sortEnum([]TriggerEvent{
	TriggerEventCron,
	TriggerEventManual,
	TriggerEventPush,
	TriggerEventPullRequest,
	TriggerEventTag,
})
