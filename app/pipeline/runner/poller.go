package runner

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	runnerModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/runner"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/manager"
	runnerStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/runner"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	gitfoxLogger "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/runner/docker/engine/resource"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/netutil"

	"github.com/drone/drone-go/drone"
	runnerclient "github.com/drone/runner-go/client"
	"github.com/drone/runner-go/poller"
)

func NewExecutionPoller(
	runner Runner,
	client runnerclient.Client,
	runnerStore runnerStore.Interface,
	cfg *config.Config,
) *poller.Poller {
	// Register default embedded runners to ops_runner table.
	// Each built-in runner gets a deterministic token so restarts are idempotent.
	registerDefaultRunners(runnerStore, cfg, client)

	runWithRecovery := func(ctx context.Context, stage *drone.Stage) (err error) {
		ctx = logger.WithUnwrappedZerolog(ctx)
		ctx = gitfoxLogger.WithLogger(ctx, gitfoxLogger.GetLogger(ctx))
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic received: %s", debug.Stack())
			}

			// the caller of this method (poller.Poller) discards the error - log it here
			if err != nil {
				gitfoxLogger.GetLogger(ctx).WithError(err).Errorf("An error occurred while calling runner.Run in Poller")
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

// registerDefaultRunners inserts isDefault=true records into ops_runner for each
// built-in docker runner, then feeds the tokens to the embedded client so subsequent
// Request() calls reuse them instead of generating random tokens.
func registerDefaultRunners(
	runnerStore runnerStore.Interface,
	cfg *config.Config,
	client runnerclient.Client,
) {
	n := int(cfg.Ci.ParallelWorkers)
	if n <= 0 {
		return
	}

	// Type-assert to inject tokens into the embedded client.
	tokenInit, ok := client.(manager.TokenInitializer)
	if !ok {
		return
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "default"
	}

	tokens := make([]string, 0, n)
	ctx := context.Background()

	for i := 1; i <= n; i++ {
		token := fmt.Sprintf("internal.default.%s.%d", hostname, i)

		// Idempotency: if the token already exists, reuse it (restart scenario).
		// Ensure the built-in runner stays online and enabled.
		existing, err := runnerStore.GetByToken(ctx, token)
		if err == nil && existing != nil && existing.Token == token {
			if existing.Online != constant.RunnerOnline || existing.Status != constant.RunnerEnable {
				gitfoxLogger.GetLogger(ctx).Infof("[poller] repairing default runner status, token=%s, online=%s, status=%s",
					token, existing.Online, existing.Status)
				if err := runnerStore.Update(ctx, &runnerModel.Runner{
					ID:     existing.ID,
					Online: constant.RunnerOnline,
					Status: constant.RunnerEnable,
				}); err != nil {
					gitfoxLogger.GetLogger(ctx).WithError(err).Errorf("[poller] failed to repair default runner, token=%s", token)
				}
			}
			tokens = append(tokens, token)
			continue
		}

		runner := &runnerModel.Runner{
			Name:      fmt.Sprintf("default-runner-%d", i),
			Hostname:  hostname,
			IP:        netutil.LocalIP(),
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
			Token:     token,
			Online:    constant.RunnerOnline,
			Status:    constant.RunnerEnable,
			IsDefault: true,
			CreatedBy: "system",
			EditedBy:  "system",
		}

		if err := runnerStore.Create(ctx, runner); err != nil {
			gitfoxLogger.GetLogger(ctx).WithError(err).Errorf("[poller] failed to register default runner, token=%s", token)
			continue
		}
		tokens = append(tokens, token)
	}

	if len(tokens) > 0 {
		tokenInit.InitTokens(tokens)
	}
}
