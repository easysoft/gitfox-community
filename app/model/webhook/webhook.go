package webhook

import (
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
	Created               *datetime.Local                  `json:"createdDate"`
	UpdatedBy             string                           `json:"updatedBy"`
	Updated               *datetime.Local                  `json:"updatedDate"`
	Internal              bool                             `json:"-"`
	DisplayName           string                           `json:"displayName"`
	Description           string                           `json:"description"`
	URL                   string                           `json:"url"`
	Secret                string                           `json:"-"`
	Enabled               bool                             `json:"enabled"`
	Insecure              bool                             `json:"insecure"`
	Triggers              []constant.WebhookTrigger        `json:"triggers"`
	LatestExecutionResult *constant.WebhookExecutionResult `json:"latestExecutionResult,omitempty"`
	AuthMethod            constant.WebhookAuthMethod       `json:"authMethod"`
	AuthHeader            string                           `json:"authHeader"`
}
