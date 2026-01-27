package constant

// TriggerAction defines the different actions on triggers will fire.
type TriggerAction string

// These are similar to enums defined in webhook enum but can diverge
// as these are different entities.
const (
	// TriggerActionBranchCreated gets triggered when a branch gets created.
	TriggerActionBranchCreated TriggerAction = "branch_created"
	// TriggerActionBranchUpdated gets triggered when a branch gets updated.
	TriggerActionBranchUpdated TriggerAction = "branch_updated"

	// TriggerActionTagCreated gets triggered when a tag gets created.
	TriggerActionTagCreated TriggerAction = "tag_created"
	// TriggerActionTagUpdated gets triggered when a tag gets updated.
	TriggerActionTagUpdated TriggerAction = "tag_updated"

	// TriggerActionPullReqCreated gets triggered when a pull request gets created.
	TriggerActionPullReqCreated TriggerAction = "pullreq_created"
	// TriggerActionPullReqReopened gets triggered when a pull request gets reopened.
	TriggerActionPullReqReopened TriggerAction = "pullreq_reopened"
	// TriggerActionPullReqBranchUpdated gets triggered when a pull request source branch gets updated.
	TriggerActionPullReqBranchUpdated TriggerAction = "pullreq_branch_updated"
	// TriggerActionPullReqClosed gets triggered when a pull request is closed.
	TriggerActionPullReqClosed TriggerAction = "pullreq_closed"
	// TriggerActionPullReqMerged gets triggered when a pull request is merged.
	TriggerActionPullReqMerged TriggerAction = "pullreq_merged"

	TriggerActionCron TriggerAction = "cron"
)

func (TriggerAction) Enum() []interface{}               { return toInterfaceSlice(triggerActions) }
func (t TriggerAction) Sanitize() (TriggerAction, bool) { return Sanitize(t, GetAllTriggerActions) }
func (t TriggerAction) GetTriggerEvent() TriggerEvent {
	if t == TriggerActionPullReqCreated ||
		t == TriggerActionPullReqBranchUpdated ||
		t == TriggerActionPullReqReopened ||
		t == TriggerActionPullReqClosed ||
		t == TriggerActionPullReqMerged {
		return TriggerEventPullRequest
	}
	if t == TriggerActionTagCreated || t == TriggerActionTagUpdated {
		return TriggerEventTag
	}
	if t == "" {
		return TriggerEventManual
	}
	if t == TriggerActionCron {
		return TriggerEventCron
	}
	return TriggerEventPush
}

func GetAllTriggerActions() ([]TriggerAction, TriggerAction) {
	return triggerActions, "" // No default value
}

var triggerActions = sortEnum([]TriggerAction{
	TriggerActionBranchCreated,
	TriggerActionBranchUpdated,
	TriggerActionTagCreated,
	TriggerActionTagUpdated,
	TriggerActionPullReqCreated,
	TriggerActionPullReqReopened,
	TriggerActionPullReqBranchUpdated,
	TriggerActionPullReqClosed,
	TriggerActionPullReqMerged,
})

// Trigger types.
const (
	TriggerHook = "@hook"
	TriggerCron = "@cron"
)
