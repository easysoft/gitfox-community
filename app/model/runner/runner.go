package runner

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

// Runner 业务模型
type Runner struct {
	ID          uint                        `json:"id"`
	Name        string                      `json:"name"`
	Desc        string                      `json:"desc"`
	Version     string                      `json:"version"`
	Hostname    string                      `json:"hostname"`
	IP          string                      `json:"ip"`
	OS          string                      `json:"os"`
	Arch        string                      `json:"arch"`
	Labels      []string                    `json:"labels"`
	Token       string                      `json:"token,omitempty"`
	HeartBeat   int                         `json:"heartBeat"`
	Online      constant.RunnerStatus       `json:"online"`
	Status      constant.RunnerActiveStatus `json:"status"`
	IsDefault   bool                        `json:"isDefault"`
	CreatedBy   string                      `json:"createdBy"`
	CreatedDate *datetime.Local             `json:"createdDate"`
	EditedBy    string                      `json:"editedBy"`
	EditedDate  *datetime.Local             `json:"editedDate"`
}

// ReqFilter query parameters.
type ReqFilter struct {
	Query    string `json:"query"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Status   string `json:"status"`
	Online   string `json:"online"`
	Sort     string `json:"sort"`
	Order    string `json:"order"`
	Page     int    `json:"page" default:"1"`
	PageSize int    `json:"pageSize" default:"20"`
}

// ReqForUpdateBasicInfo 更新 runner 基础信息请求
type ReqForUpdateBasicInfo struct {
	Name string `json:"name" binding:"required"`
	Desc string `json:"desc"`
}

// ReqForUpdateStatus 更新 runner 状态请求
type ReqForUpdateStatus struct {
	Status constant.RunnerActiveStatus `json:"status" binding:"required,oneof=enable disable"`
}

// ReqForRegisterRunner 注册 runner 请求
type RegisterRunnerReq struct {
	Name     string            `json:"name" binding:"required"`
	Desc     string            `json:"desc"`
	Version  string            `json:"version"`
	Hostname string            `json:"hostname"`
	IP       string            `json:"ip"`
	OS       string            `json:"os"`
	Arch     string            `json:"arch"`
	Labels   map[string]string `json:"labels"`
}

// GenerateTokenResponse 生成 token 响应
type GenerateTokenResponse struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}
