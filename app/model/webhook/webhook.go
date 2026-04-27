package webhook

import (
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

type ReqForCreate struct {
	DisplayName string                     `json:"displayName"`
	Description string                     `json:"description"`
	URL         string                     `json:"url"`
	Secret      string                     `json:"secret"`
	Enabled     bool                       `json:"enabled"`
	Insecure    bool                       `json:"insecure"`
	Triggers    []constant.WebhookTrigger  `json:"triggers"`
	AuthMethod  constant.WebhookAuthMethod `json:"authMethod"`
	AuthHeader  string                     `json:"authHeader"`
}

type ReqForUpdate struct {
	DisplayName *string                     `json:"displayName"`
	Description *string                     `json:"description"`
	URL         *string                     `json:"url"`
	Secret      *string                     `json:"secret"`
	Enabled     *bool                       `json:"enabled"`
	Insecure    *bool                       `json:"insecure"`
	Triggers    []constant.WebhookTrigger   `json:"triggers"`
	AuthMethod  *constant.WebhookAuthMethod `json:"authMethod"`
	AuthHeader  *string                     `json:"authHeader"`
}

type ReqForList struct {
	Page     int    `form:"page"     default:"1"`
	PageSize int    `form:"pageSize" default:"20"`
	Sort     string `form:"sort"`
	Order    string `form:"order"`
}

// Webhook represents a webhook.
type Webhook struct {
	ID                    int64                            `json:"id"`
	ParentID              int64                            `json:"parentID"`
	ParentType            string                           `json:"parentType"`
	CreatedBy             string                           `json:"createdBy"`
	CreatedDate           *datetime.Local                  `json:"createdDate"`
	EditedBy              string                           `json:"editedBy"`
	EditedDate            *datetime.Local                  `json:"editedDate"`
	Internal              bool                             `json:"-"`
	DisplayName           string                           `json:"displayName"`
	Description           string                           `json:"description"`
	URL                   string                           `json:"url"`
	Secret                string                           `json:"secret"`
	Enabled               bool                             `json:"enabled"`
	Insecure              bool                             `json:"insecure"`
	Triggers              []constant.WebhookTrigger        `json:"triggers"`
	LatestExecutionResult *constant.WebhookExecutionResult `json:"latestExecutionResult,omitempty"`
	AuthMethod            constant.WebhookAuthMethod       `json:"authMethod"`
	AuthHeader            string                           `json:"authHeader"`
}

// Execution represents a webhook execution.
type Execution struct {
	ID             int64                           `json:"id"`
	WebhookID      int64                           `json:"webhookID"`
	TriggerType    string                          `json:"triggerType"`
	EventID        string                          `json:"eventID"`
	Result         constant.WebhookExecutionResult `json:"result"`
	Duration       int64                           `json:"duration"`
	Error          string                          `json:"error"`
	ReqURL         string                          `json:"reqUrl"`
	ReqHeaders     string                          `json:"reqHeaders"`
	ReqBody        string                          `json:"reqBody"`
	RespStatusCode int                             `json:"respStatusCode"`
	RespStatus     string                          `json:"respStatus"`
	RespHeaders    string                          `json:"respHeaders"`
	RespBody       string                          `json:"respBody"`
	RetriggerID    *int64                          `json:"retriggerID"`
	ReTriggerable  bool                            `json:"reTriggerable"`
	CreatedDate    *datetime.Local                 `json:"createdDate"`
	EditedDate     *datetime.Local                 `json:"editedDate"`
}

type WebhookParentInfo struct {
	Type constant.WebhookParent
	ID   int64
}

type TriggerResult struct {
	TriggerID   string
	TriggerType string
	Webhook     *Webhook
	Execution   *Execution
	Err         error
}

func (r *TriggerResult) Skipped() bool {
	return r.Execution == nil
}

// BaseSegment contains base info of all payloads for webhooks.
type BaseSegment struct {
	Trigger   string         `json:"trigger"`
	Repo      RepositoryInfo `json:"repo"`
	Principal string         `json:"principal"`
}

// ReferenceSegment contains the reference info for webhooks.
type ReferenceSegment struct {
	Ref ReferenceInfo `json:"ref"`
}

// ReferenceDetailsSegment contains extra details for reference related payloads for webhooks.
type ReferenceDetailsSegment struct {
	SHA string `json:"sha"`

	HeadCommit        *CommitInfo   `json:"head_commit,omitempty"`
	Commits           *[]CommitInfo `json:"commits,omitempty"`
	TotalCommitsCount int           `json:"total_commits_count,omitempty"`

	// Deprecated
	Commit *CommitInfo `json:"commit,omitempty"`
}

// ReferenceUpdateSegment contains extra details for reference update related payloads for webhooks.
type ReferenceUpdateSegment struct {
	OldSHA string `json:"old_sha"`
	Forced bool   `json:"forced"`
}

// ReferenceInfo describes a unique reference in Harness.
// It contains both the reference name as well as the repo the reference belongs to.
type ReferenceInfo struct {
	Name string         `json:"name"`
	Repo RepositoryInfo `json:"repo"`
}

type RepositoryInfo struct {
	ID            int64  `json:"id"`
	Identifier    string `json:"identifier"`
	Path          string `json:"path"`
	GitUID        string `json:"git_uid"`
	Description   string `json:"description"`
	DefaultBranch string `json:"default_branch"`
	URL           string `json:"url"`
	GitURL        string `json:"git_url"`
	GitSSHURL     string `json:"git_ssh_url"`
}

// CommitInfo describes the commit related info for a webhook payload.
// NOTE: don't use types package as we want webhook payload to be independent from API calls.
type CommitInfo struct {
	SHA       string        `json:"sha"`
	Message   string        `json:"message"`
	Author    SignatureInfo `json:"author"`
	Committer SignatureInfo `json:"committer"`

	Added    []string `json:"added"`
	Removed  []string `json:"removed"`
	Modified []string `json:"modified"`
}

// SignatureInfo describes the commit signature related info for a webhook payload.
// NOTE: don't use types package as we want webhook payload to be independent from API calls.
type SignatureInfo struct {
	Identity IdentityInfo `json:"identity"`
	When     time.Time    `json:"when"`
}

// IdentityInfo describes the signature identity related info for a webhook payload.
// NOTE: don't use types package as we want webhook payload to be independent from API calls.
type IdentityInfo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type ParentInfo struct {
	Type string
	ID   uint
}

// PullReqSegment contains details for all pull req related payloads for webhooks.
type PullReqSegment struct {
	PullReq PullReqInfo `json:"pull_req"`
}

// PullReqInfo describes the pullreq related info for a webhook payload.
// NOTE: don't use types package as we want pullreq payload to be independent from API calls.
type PullReqInfo struct {
	Number        int64                 `json:"number"`
	State         constant.PullReqState `json:"state"`
	IsDraft       bool                  `json:"is_draft"`
	Title         string                `json:"title"`
	Description   string                `json:"description"`
	SourceRepoID  int64                 `json:"source_repo_id"`
	SourceBranch  string                `json:"source_branch"`
	TargetRepoID  int64                 `json:"target_repo_id"`
	TargetBranch  string                `json:"target_branch"`
	MergeStrategy *constant.MergeMethod `json:"merge_strategy,omitempty"`
	Author        string                `json:"author"`
	PrURL         string                `json:"pr_url"`
}

// PullReqTargetReferenceSegment contains details for the pull req target reference for webhooks.
type PullReqTargetReferenceSegment struct {
	TargetRef ReferenceInfo `json:"target_ref"`
}
