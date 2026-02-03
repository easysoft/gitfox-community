package manager

import (
	"time"

	executionModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/execution"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"github.com/drone/drone-go/drone"
)

// toUnix converts *datetime.Local to Unix seconds, returns 0 if nil.
func toUnix(t *datetime.Local) int64 {
	if t == nil {
		return 0
	}
	return time.Time(*t).Unix()
}

func ConvertToDroneBuild(execution *executionModel.Execution) *drone.Build {
	return &drone.Build{
		ID:       int64(execution.ID),
		Trigger:  execution.Trigger,
		Parent:   int64(execution.Parent),
		Status:   string(execution.Status),
		Error:    execution.Error,
		Ref:      execution.Ref,
		Params:   execution.Params,
		Started:  toUnix(execution.Started),
		Finished: toUnix(execution.Finished),
		Created:  toUnix(execution.CreatedDate),
		Updated:  toUnix(execution.EditedDate),
	}
}
