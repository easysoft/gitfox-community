package jobs

import (
	"bytes"
	"encoding/gob"

	jobModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/jobs"
	pkgJob "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/jobs"
)

// encodeStateChange encodes a ScheduleJob into a StateChange message.
// 将 ScheduleJob 编码为 StateChange 消息
func encodeStateChange(job *jobModel.ScheduleJob) ([]byte, error) {
	stateChange := &pkgJob.StateChange{
		Name:     job.Name,
		Type:     job.Type,
		State:    string(job.State),
		Progress: job.RunProgress,
		Result:   job.Result,
		Failure:  job.LastFailureError,
	}

	buffer := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buffer).Encode(stateChange); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
