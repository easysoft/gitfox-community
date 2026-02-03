package trigger

import "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"

type Trigger struct {
	ID          int64                  `json:"-"`
	PipelineID  int64                  `json:"pipeline_id"`
	Description string                 `json:"description"`
	Type        string                 `json:"trigger_type"`
	Secret      string                 `json:"-"`
	RepoID      int64                  `json:"repo_id"`
	CreatedBy   int64                  `json:"created_by"`
	Disabled    bool                   `json:"disabled"`
	Actions     constant.TriggerAction `json:"actions"`
}
