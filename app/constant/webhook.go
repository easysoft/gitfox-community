// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package constant

// WebhookParent defines different types of parents of a webhook.
type WebhookParent string

func (WebhookParent) Enum() []interface{} { return toInterfaceSlice(webhookParents) }

const (
	// WebhookParentRepo describes a repo as webhook owner.
	WebhookParentRepo WebhookParent = "repo"

	// WebhookParentSpace describes a space as webhook owner.
	WebhookParentSpace WebhookParent = "space"

	DefaultAuthMethod = "token"
	DefaultAuthHeader = "X-Gitfox-Token"
)

var webhookParents = sortEnum([]WebhookParent{
	WebhookParentRepo,
	WebhookParentSpace,
})

// WebhookExecutionResult defines the different results of a webhook execution.
type WebhookExecutionResult string

func (WebhookExecutionResult) Enum() []interface{} { return toInterfaceSlice(webhookExecutionResults) }

const (
	// WebhookExecutionResultSuccess describes a webhook execution result that succeeded.
	WebhookExecutionResultSuccess WebhookExecutionResult = "success"

	// WebhookExecutionResultRetriableError describes a webhook execution result that failed with a retriable error.
	WebhookExecutionResultRetriableError WebhookExecutionResult = "retriable_error"

	// WebhookExecutionResultFatalError describes a webhook execution result that failed with an unrecoverable error.
	WebhookExecutionResultFatalError WebhookExecutionResult = "fatal_error"
)

var webhookExecutionResults = sortEnum([]WebhookExecutionResult{
	WebhookExecutionResultSuccess,
	WebhookExecutionResultRetriableError,
	WebhookExecutionResultFatalError,
})

// WebhookTrigger defines the different types of webhook triggers available.
type WebhookTrigger string

func (WebhookTrigger) Enum() []interface{}                { return toInterfaceSlice(webhookTriggers) }
func (s WebhookTrigger) Sanitize() (WebhookTrigger, bool) { return Sanitize(s, GetAllWebhookTriggers) }

func GetAllWebhookTriggers() ([]WebhookTrigger, WebhookTrigger) {
	return webhookTriggers, "" // No default value
}

const (
	// WebhookTriggerBranchCreated gets triggered when a branch gets created.
	WebhookTriggerBranchCreated WebhookTrigger = "branch_created"
	// WebhookTriggerBranchUpdated gets triggered when a branch gets updated.
	WebhookTriggerBranchUpdated WebhookTrigger = "branch_updated"
	// WebhookTriggerBranchDeleted gets triggered when a branch gets deleted.
	WebhookTriggerBranchDeleted WebhookTrigger = "branch_deleted"

	// WebhookTriggerTagCreated gets triggered when a tag gets created.
	WebhookTriggerTagCreated WebhookTrigger = "tag_created"
	// WebhookTriggerTagUpdated gets triggered when a tag gets updated.
	WebhookTriggerTagUpdated WebhookTrigger = "tag_updated"
	// WebhookTriggerTagDeleted gets triggered when a tag gets deleted.
	WebhookTriggerTagDeleted WebhookTrigger = "tag_deleted"

	// WebhookTriggerPullReqCreated gets triggered when a pull request gets created.
	WebhookTriggerPullReqCreated WebhookTrigger = "pullreq_created"
	// WebhookTriggerPullReqReopened gets triggered when a pull request gets reopened.
	WebhookTriggerPullReqReopened WebhookTrigger = "pullreq_reopened"
	// WebhookTriggerPullReqBranchUpdated gets triggered when a pull request source branch gets updated.
	WebhookTriggerPullReqBranchUpdated WebhookTrigger = "pullreq_branch_updated"
	// WebhookTriggerPullReqClosed gets triggered when a pull request is closed.
	WebhookTriggerPullReqClosed WebhookTrigger = "pullreq_closed"
	// WebhookTriggerPullReqCommentCreated gets triggered when a pull request comment gets created.
	WebhookTriggerPullReqCommentCreated WebhookTrigger = "pullreq_comment_created"
	// WebhookTriggerPullReqCommentUpdated gets triggered when a pull request comment gets edited.
	WebhookTriggerPullReqCommentUpdated WebhookTrigger = "pullreq_comment_updated"
	// WebhookTriggerPullReqCommentStatusUpdated gets triggered when a pull request comment status gets updated.
	WebhookTriggerPullReqCommentStatusUpdated WebhookTrigger = "pullreq_comment_status_updated"
	// WebhookTriggerPullReqMerged gets triggered when a pull request is merged.
	WebhookTriggerPullReqMerged WebhookTrigger = "pullreq_merged"
	// WebhookTriggerPullReqUpdated gets triggered when a pull request gets updated.
	WebhookTriggerPullReqUpdated WebhookTrigger = "pullreq_updated"
	// WebhookTriggerPullReqLabelAssigned gets triggered when a label is assigned to a pull request.
	WebhookTriggerPullReqLabelAssigned WebhookTrigger = "pullreq_label_assigned"

	// WebhookTriggerPullReqReviewerCreated gets triggered when a reviewer is added to a pull request.
	WebhookTriggerPullReqReviewerCreated WebhookTrigger = "pullreq_reviewer_created"

	// WebhookTriggerPullReqRequiredChecksPassed gets triggered when all necessary checks of a pull request are passed.
	WebhookTriggerPullReqRequiredChecksPassed WebhookTrigger = "pullreq_required_checks_passed"

	// WebhookTriggerPullReqReviewerDeleted gets triggered when a reviewer is removed from a pull request.
	WebhookTriggerPullReqReviewerDeleted WebhookTrigger = "pullreq_reviewer_deleted"
	// WebhookTriggerPullReqReviewSubmitted gets triggered when a review is submitted for a pull request.
	WebhookTriggerPullReqReviewSubmitted WebhookTrigger = "pullreq_review_submitted"

	// WebhookTriggerArtifactCreated gets triggered when a artifact gets created.
	WebhookTriggerArtifactCreated WebhookTrigger = "artifact_created"
	// WebhookTriggerArtifactUpdated gets triggered when a artifact gets updated.
	WebhookTriggerArtifactUpdated WebhookTrigger = "artifact_updated"
	// WebhookTriggerArtifactDeleted gets triggered when a artifact gets deleted.
	WebhookTriggerArtifactDeleted WebhookTrigger = "artifact_deleted"
)

var webhookTriggers = sortEnum([]WebhookTrigger{
	WebhookTriggerBranchCreated,
	WebhookTriggerBranchUpdated,
	WebhookTriggerBranchDeleted,
	WebhookTriggerTagCreated,
	WebhookTriggerTagUpdated,
	WebhookTriggerTagDeleted,
	WebhookTriggerPullReqCreated,
	WebhookTriggerPullReqUpdated,
	WebhookTriggerPullReqReopened,
	WebhookTriggerPullReqBranchUpdated,
	WebhookTriggerPullReqClosed,
	WebhookTriggerPullReqCommentCreated,
	WebhookTriggerPullReqCommentUpdated,
	WebhookTriggerPullReqCommentStatusUpdated,
	WebhookTriggerPullReqMerged,
	WebhookTriggerPullReqLabelAssigned,
	WebhookTriggerPullReqReviewerCreated,
	WebhookTriggerPullReqReviewerDeleted,
	WebhookTriggerPullReqReviewSubmitted,
	WebhookTriggerArtifactCreated,
	WebhookTriggerArtifactUpdated,
	WebhookTriggerArtifactDeleted,
})

type WebhookAuthMethod string

func (WebhookAuthMethod) Enum() []interface{} { return toInterfaceSlice(webhookAuthMethods) }
func (s WebhookAuthMethod) Sanitize() (WebhookAuthMethod, bool) {
	return Sanitize(s, GetAllWebhookAuthMethods)
}

func GetAllWebhookAuthMethods() ([]WebhookAuthMethod, WebhookAuthMethod) {
	return webhookAuthMethods, WebhookAuthMethodNone // No default value
}

const (
	WebhookAuthMethodNone      WebhookAuthMethod = ""
	WebhookAuthMethodSignature WebhookAuthMethod = "signature"
	WebhookAuthMethodToken     WebhookAuthMethod = "token"
)

var webhookAuthMethods = sortEnum([]WebhookAuthMethod{
	WebhookAuthMethodNone,
	WebhookAuthMethodSignature,
	WebhookAuthMethodToken,
})
