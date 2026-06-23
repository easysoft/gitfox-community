package manager

import (
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	executionModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/execution"
	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	stageModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/stage"
	stepModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/step"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/livelog"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"github.com/drone/drone-go/drone"
	"github.com/drone/runner-go/client"
)

// toUnix converts *datetime.Local to Unix seconds, returns 0 if nil.
func toUnix(t *datetime.Local) int64 {
	if t == nil {
		return 0
	}
	return time.Time(*t).Unix()
}

// toLocal converts Unix seconds to *datetime.Local, returns nil if 0.
func toLocal(u int64) *datetime.Local {
	if u == 0 {
		return nil
	}
	tl := datetime.Local(time.Unix(u, 0))
	return &tl
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
		Before:   execution.CommitBefore,
		After:    execution.Commit,
		Started:  toUnix(execution.StartedDate),
		Finished: toUnix(execution.FinishedDate),
		Created:  toUnix(execution.CreatedDate),
		Updated:  toUnix(execution.EditedDate),
	}
}

func ConvertToDroneStageWithoutSteps(stage *stageModel.Stage) *drone.Stage {
	onSuccess := true
	onFailure := false
	switch stage.RunPolicy {
	case string(constant.RunPolicyAlways):
		onFailure = true
	case string(constant.RunPolicyOnFailure):
		onFailure = true
		onSuccess = false
	default:
	}

	return &drone.Stage{
		ID:        int64(stage.ID),
		BuildID:   int64(stage.ExecutionID),
		Number:    stage.SN,
		Name:      stage.Name,
		Type:      stage.Type,
		Status:    string(stage.Status),
		Error:     stage.ErrorLog,
		Machine:   stage.Host,
		OS:        stage.OS,
		Arch:      stage.Arch,
		Started:   toUnix(stage.Started),
		Stopped:   toUnix(stage.Finished),
		OnSuccess: onSuccess,
		OnFailure: onFailure,
		DependsOn: stage.Depends,
		Labels:    stage.Labels,
		Steps:     nil,
	}
}

func ConvertToDroneStage(stage *stageModel.StageWithSteps) *drone.Stage {

	onSuccess := true
	onFailure := false
	switch stage.RunPolicy {
	case string(constant.RunPolicyAlways):
		onFailure = true
	case string(constant.RunPolicyOnFailure):
		onFailure = true
		onSuccess = false
	default:
	}

	return &drone.Stage{
		ID:        int64(stage.ID),
		BuildID:   int64(stage.ExecutionID),
		Number:    stage.SN,
		Name:      stage.Name,
		Type:      stage.Type,
		Status:    string(stage.Status),
		Error:     stage.ErrorLog,
		Machine:   stage.Host,
		OS:        stage.OS,
		Arch:      stage.Arch,
		Started:   toUnix(stage.Started),
		Stopped:   toUnix(stage.Finished),
		OnSuccess: onSuccess,
		OnFailure: onFailure,
		DependsOn: stage.Depends,
		Labels:    stage.Labels,
		Steps:     ConvertToDroneSteps(stage.Steps),
	}
}

func ConvertToDroneSteps(steps []*stepModel.Step) []*drone.Step {
	droneSteps := make([]*drone.Step, len(steps))
	for i, step := range steps {
		droneSteps[i] = ConvertToDroneStep(step)
	}
	return droneSteps
}

func ConvertToDroneStep(step *stepModel.Step) *drone.Step {
	return &drone.Step{
		ID:        int64(step.ID),
		StageID:   int64(step.StageID),
		Number:    step.SN,
		Name:      step.Name,
		Status:    string(step.Status),
		Error:     step.ErrorLog,
		Started:   toUnix(step.Started),
		Stopped:   toUnix(step.Finished),
		DependsOn: step.Depends,
		Image:     step.Image,
	}
}

func ConvertToDroneFile(file *pipelineModel.PipelineFile) *client.File {
	return &client.File{
		Data: file.Data,
	}
}

func ConvertToDroneNetrc(netrc *Netrc) *drone.Netrc {
	if netrc == nil {
		return nil
	}

	return &drone.Netrc{
		Machine:  netrc.Machine,
		Login:    netrc.Login,
		Password: netrc.Password,
	}
}

func ConvertFromDroneStep(step *drone.Step) *stepModel.Step {
	return &stepModel.Step{
		ID:       uint(step.ID),
		StageID:  uint(step.StageID),
		SN:       step.Number,
		Name:     step.Name,
		Status:   constant.ParseCIStatus(step.Status),
		ErrorLog: step.Error,
		Started:  toLocal(step.Started),
		Finished: toLocal(step.Stopped),
		Depends:  step.DependsOn,
		Image:    step.Image,
	}
}

func ConvertFromDroneSteps(steps []*drone.Step) []*stepModel.Step {
	typesSteps := make([]*stepModel.Step, len(steps))
	for i, step := range steps {
		stepCode := ""
		stepName := step.Name
		stepBaseInfo := strings.Split(step.Name, "::")
		if len(stepBaseInfo) > 1 {
			stepCode = stepBaseInfo[0]
			stepName = stepBaseInfo[1]
		}
		typesSteps[i] = &stepModel.Step{
			ID:       uint(step.ID),
			StageID:  uint(step.StageID),
			SN:       step.Number,
			Code:     stepCode,
			Name:     stepName,
			Status:   constant.ParseCIStatus(step.Status),
			ErrorLog: step.Error,
			Started:  toLocal(step.Started),
			Finished: toLocal(step.Stopped),
			Depends:  step.DependsOn,
			Image:    step.Image,
		}
	}
	return typesSteps
}

func ConvertFromDroneStage(stage *drone.Stage) *stageModel.StageWithSteps {

	runPolicy := constant.RunPolicyOnSuccess
	switch {
	case stage.OnSuccess && stage.OnFailure:
		runPolicy = constant.RunPolicyAlways
	case stage.OnSuccess:
		runPolicy = constant.RunPolicyOnSuccess
	case stage.OnFailure:
		runPolicy = constant.RunPolicyOnFailure
	}

	return &stageModel.StageWithSteps{
		Stage: stageModel.Stage{
			ID:          uint(stage.ID),
			ExecutionID: uint(stage.BuildID),
			SN:          stage.Number,
			Name:        stage.Name,
			Type:        stage.Type,
			Status:      constant.ParseCIStatus(stage.Status),
			ErrorLog:    stage.Error,
			Host:        stage.Machine,
			OS:          stage.OS,
			Arch:        stage.Arch,
			Started:     toLocal(stage.Started),
			Finished:    toLocal(stage.Stopped),
			RunPolicy:   string(runPolicy),
			Depends:     stage.DependsOn,
			Labels:      stage.Labels,
		},
		Steps: ConvertFromDroneSteps(stage.Steps),
	}
}

func ConvertFromDroneLine(l *drone.Line) *livelog.Line {
	return &livelog.Line{
		Number:    l.Number,
		Message:   l.Message,
		Timestamp: time.Now().Unix(),
	}
}
