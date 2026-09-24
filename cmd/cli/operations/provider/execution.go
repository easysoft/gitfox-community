// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package provider

import (
	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/app/api/controller/execution"
	"github.com/easysoft/gitfox-community/app/pipeline/canceler"
	"github.com/easysoft/gitfox-community/app/pipeline/manager"
	"github.com/easysoft/gitfox-community/app/pipeline/scheduler"
	"github.com/easysoft/gitfox-community/app/pipeline/triggerer"
	executionSvc "github.com/easysoft/gitfox-community/app/service/execution"
	stageSvc "github.com/easysoft/gitfox-community/app/service/stage"
	stepSvc "github.com/easysoft/gitfox-community/app/service/step"
	executionStore "github.com/easysoft/gitfox-community/app/store/execution"
	stageStore "github.com/easysoft/gitfox-community/app/store/stage"
	stepStore "github.com/easysoft/gitfox-community/app/store/step"
	"github.com/easysoft/gitfox-community/pkg/database/dbtx"
)

var ExecutionWireSet = wire.NewSet(
	execution.NewController,
	triggerer.New,
	scheduler.New,
	canceler.New,
	wire.Bind(new(manager.ExecutionManager), new(*manager.Manager)),
	manager.New,
	manager.NewEmbeddedClient,

	// Execution service
	wire.Bind(new(executionSvc.Interface), new(*executionSvc.Service)),
	executionSvc.NewExecutionService,
	// Execution store
	wire.Bind(new(executionStore.Interface), new(*executionStore.Store)),
	executionStore.NewExecutionStore,

	// Stage service
	wire.Bind(new(stageSvc.Interface), new(*stageSvc.Service)),
	stageSvc.NewStageService,
	// Stage store
	wire.Bind(new(stageStore.Interface), new(*stageStore.Store)),
	stageStore.NewStageStore,

	// Step service
	wire.Bind(new(stepSvc.Interface), new(*stepSvc.Service)),
	stepSvc.NewStepService,
	// Step store
	wire.Bind(new(stepStore.Interface), new(*stepStore.Store)),
	stepStore.NewStepStore,

	// Transaction runner
	wire.Bind(new(dbtx.TxRunner), new(*dbtx.TxRunnerImpl)),
	dbtx.NewTxRunner,
)
