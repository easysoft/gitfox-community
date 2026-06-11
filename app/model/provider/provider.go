package provider

import "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"

// Provider 外部服务模型。
type Provider struct {
	ID          uint            `json:"id"`
	Type        string          `json:"type"`
	Name        string          `json:"name"`
	URL         string          `json:"url"`
	Token       string          `json:"token"`
	CreatedBy   string          `json:"createdBy"`
	CreatedDate *datetime.Local `json:"createdDate"`
	EditedBy    string          `json:"editedBy"`
	EditedDate  *datetime.Local `json:"editedDate"`
}
