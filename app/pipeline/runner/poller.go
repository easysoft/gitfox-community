package runner

import (
	"context"
	"fmt"
	"runtime/debug"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/runner/docker/engine/resource"

	gitfoxLogger "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"

	"github.com/drone/drone-go/drone"
	runnerclient "github.com/drone/runner-go/client"
	"github.com/drone/runner-go/poller"
)

func NewExecutionPoller(
	runner Runner,
	client runnerclient.Client,
) *poller.Poller {
	runWithRecovery := func(ctx context.Context, stage *drone.Stage) (err error) {
		ctx = logger.WithUnwrappedLogger(ctx)
		ctx = gitfoxLogger.WithLogger(ctx, gitfoxLogger.GetLogger(ctx))
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic received: %s", debug.Stack())
			}

			// the caller of this method (poller.Poller) discards the error - log it here
			if err != nil {
				gitfoxLogger.GetLogger(ctx).Errorw("An error occurred while calling runner.Run in Poller", "error", err)
			}
		}()
		return runner.Run(ctx, stage)
	}

	return &poller.Poller{
		Client:   client,
		Dispatch: runWithRecovery,
		Filter: &runnerclient.Filter{
			Kind: resource.Kind,
			Type: resource.Type,
			// TODO: Check if other parameters are needed.
		},
	}
}
