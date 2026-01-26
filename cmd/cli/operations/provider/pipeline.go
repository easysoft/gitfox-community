package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/pipeline"
	pipelineSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	pipelineStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/pipeline"
)

var PipelineWireSet = wire.NewSet(
	pipeline.NewController,
	wire.Bind(new(pipelineSvc.PluginService), new(*pipelineSvc.Service)),
	pipelineSvc.NewPluginService,
	wire.Bind(new(pipelineStore.Interface), new(*pipelineStore.Store)),
	pipelineStore.NewPluginStore,
)
