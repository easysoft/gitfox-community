package webhook

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// Webhook 对应数据库表 ops_webhooks
type Webhook struct {
	ID               uint            `gorm:"column:id"`
	SpaceID          int64           `gorm:"column:spaceID"`
	RepoID           int64           `gorm:"column:repoID"`
	DisplayName      string          `gorm:"column:displayName"`
	Desc             string          `gorm:"column:desc"`
	URL              string          `gorm:"column:url"`
	Secret           string          `gorm:"column:secret"`
	IsSSL            bool            `gorm:"column:isSSL"`
	Triggers         string          `gorm:"column:triggers"`
	Internal         bool            `gorm:"column:internal"`
	AuthMethod       string          `gorm:"column:authMethod"`
	AuthHeader       string          `gorm:"column:authHeader"`
	LatestExecResult string          `gorm:"column:latestExecResult"`
	Enabled          bool            `gorm:"column:enabled"`
	CreatedBy        string          `gorm:"column:createdBy"`
	CreatedDate      *datetime.Local `gorm:"column:createdDate"`
	EditedBy         string          `gorm:"column:editedBy"`
	EditedDate       *datetime.Local `gorm:"column:editedDate"`
	Deleted          uint8           `gorm:"column:deleted"`
}

func (w *Webhook) TableName() string {
	return "ops_webhooks"
}

func (w *Webhook) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	w.CreatedDate = now
	w.EditedDate = now
	return nil
}

func (w *Webhook) BeforeUpdate(tx *gorm.DB) error {
	w.EditedDate = datetime.NowLocalTime()
	return nil
}

// Execution 对应数据库表 ops_webhook_executions
type Execution struct {
	ID             uint            `gorm:"column:id"`
	WebhookID      uint            `gorm:"column:webhookID"`
	TriggerType    string          `gorm:"column:triggerType"`
	EventID        string          `gorm:"column:eventID"`
	Result         string          `gorm:"column:result"`
	Duration       int64           `gorm:"column:duration"`
	Error          string          `gorm:"column:error"`
	ReqURL         string          `gorm:"column:reqUrl"`
	ReqHeaders     string          `gorm:"column:reqHeaders"`
	ReqBody        string          `gorm:"column:reqBody"`
	RespStatusCode int             `gorm:"column:respStatusCode"`
	RespStatus     string          `gorm:"column:respStatus"`
	RespHeaders    string          `gorm:"column:respHeaders"`
	RespBody       string          `gorm:"column:respBody"`
	RetriggerID    *uint           `gorm:"column:retriggerID"`
	Retriggerable  bool            `gorm:"column:retriggerable"`
	CreatedDate    *datetime.Local `gorm:"column:createdDate"`
	EditedDate     *datetime.Local `gorm:"column:editedDate"`
}

func (e *Execution) TableName() string {
	return "ops_webhook_executions"
}
