package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/livelog"
	urlprovider "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"

	"github.com/drone/drone-go/drone"
	"github.com/drone/runner-go/client"
)

type embedded struct {
	config      *config.Config
	urlProvider urlprovider.Provider
	manager     ExecutionManager
}

var _ client.Client = (*embedded)(nil)

func NewEmbeddedClient(
	manager ExecutionManager,
	urlProvider urlprovider.Provider,
	config *config.Config,
) client.Client {
	return &embedded{
		config:      config,
		urlProvider: urlProvider,
		manager:     manager,
	}
}

// Join notifies the server the runner is joining the cluster.
// Since the runner is embedded, this can just return nil.
func (e *embedded) Join(_ context.Context, _ string) error {
	return nil
}

// Leave notifies the server the runner is leaving the cluster.
// Since the runner is embedded, this can just return nil.
func (e *embedded) Leave(_ context.Context, _ string) error {
	return nil
}

// Ping sends a ping message to the server to test connectivity.
// Since the runner is embedded, this can just return nil.
func (e *embedded) Ping(_ context.Context, _ string) error {
	return nil
}

// Request requests the next available build stage for execution.
func (e *embedded) Request(ctx context.Context, args *client.Filter) (*drone.Stage, error) {
	request := &Request{
		Type:   args.Type,
		OS:     args.OS,
		Arch:   args.Arch,
		Labels: args.Labels,
		Metadata: map[string]string{
			"hostname":     "default",
			"ipaddress":    "127.0.0.1",
			"worker_total": fmt.Sprintf("%d", e.config.Ci.ParallelWorkers),
			"internal":     "true",
		},
	}
	stage, err := e.manager.Request(ctx, request)
	if err != nil {
		return nil, err
	}
	return ConvertToDroneStageWithoutSteps(stage), nil
}

// Accept accepts the build stage for execution.
func (e *embedded) Accept(ctx context.Context, s *drone.Stage) error {
	stage, err := e.manager.Accept(ctx, s.ID, s.Machine)
	if err != nil {
		return err
	}
	*s = *ConvertToDroneStageWithoutSteps(stage)
	return err
}

// Detail gets the build stage details for execution.
func (e *embedded) Detail(ctx context.Context, stage *drone.Stage) (*client.Context, error) {
	details, err := e.manager.Details(ctx, stage.ID)
	if err != nil {
		return nil, err
	}

	return &client.Context{
		Build:  ConvertToDroneBuild(&details.Execution.Execution),
		Stage:  ConvertToDroneStage(details.Stage),
		Config: ConvertToDroneFile(details.Config),
		Netrc:  ConvertToDroneNetrc(details.Netrc),
		System: &drone.System{
			Proto: e.urlProvider.GetAPIProto(ctx),
			Host:  e.urlProvider.GetAPIHostname(ctx),
		},
		// Secrets: ConvertToDroneSecrets(details.Secrets),
	}, nil
}

// Update updates the build stage.
func (e *embedded) Update(ctx context.Context, stage *drone.Stage) error {
	var err error
	convertedStage := ConvertFromDroneStage(stage)
	status := constant.ParseCIStatus(stage.Status)
	if status == constant.CIStatusPending || status == constant.CIStatusRunning {
		err = e.manager.BeforeStage(ctx, convertedStage)
	} else {
		err = e.manager.AfterStage(ctx, convertedStage)
	}
	*stage = *ConvertToDroneStage(convertedStage)
	return err
}

// UpdateStep updates the build step.
func (e *embedded) UpdateStep(ctx context.Context, step *drone.Step) error {
	var err error
	convertedStep := ConvertFromDroneStep(step)
	status := constant.ParseCIStatus(step.Status)
	if status == constant.CIStatusPending || status == constant.CIStatusRunning {
		err = e.manager.BeforeStep(ctx, convertedStep)
	} else {
		err = e.manager.AfterStep(ctx, convertedStep)
	}
	*step = *ConvertToDroneStep(convertedStep)
	return err
}

// Watch watches for build cancellation requests.
func (e *embedded) Watch(ctx context.Context, executionID int64) (bool, error) {
	return e.manager.Watch(ctx, executionID)
}

// Batch batch writes logs to the streaming logs.
func (e *embedded) Batch(ctx context.Context, stepID int64, lines []*drone.Line) error {
	for _, l := range lines {
		line := ConvertFromDroneLine(l)
		err := e.manager.Write(ctx, stepID, line)
		if err != nil {
			return err
		}
	}
	return nil
}

// Upload uploads the full logs to the server.
func (e *embedded) Upload(ctx context.Context, stepID int64, l []*drone.Line) error {
	var buffer bytes.Buffer
	lines := []livelog.Line{}
	for _, line := range l {
		lines = append(lines, *ConvertFromDroneLine(line))
	}
	out, err := json.Marshal(lines)
	if err != nil {
		return err
	}
	_, err = buffer.Write(out)
	if err != nil {
		return err
	}
	return e.manager.UploadLogs(ctx, stepID, &buffer)
}

// UploadCard uploads a card to drone server.
func (e *embedded) UploadCard(_ context.Context, _ int64, _ *drone.CardInput) error {
	// Implement UploadCard logic here
	return nil // Replace with appropriate error handling and logic
}
