package req

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"github.com/gin-gonic/gin"
)

const (
	QueryParamPipelineOs     = "os"
	QueryParamPipelineArch   = "arch"
	QueryParamPipelineRunner = "runner"
	QueryParamStepCategory   = "category"

	QueryParamPipelineBuildEnvType     = "type"
	QueryParamPipelineBuildEnvQuery    = "query"
	QueryParamPipelineBuildEnvCategory = "category"

	PathParamBuildEnvId = "env_id"
	PathParamSchemaName = "name"
)

func ParsePipelineOs(ctx *gin.Context) constant.PipelineOs {
	result, ok := constant.PipelineOs(ctx.Query(QueryParamPipelineOs)).Sanitize()
	if !ok {
		return constant.PipelineOsLinux
	}
	return result
}

func ParsePipelineRunner(ctx *gin.Context, runner string) constant.PipelineRunner {
	result, ok := constant.PipelineRunner(ctx.Query(QueryParamPipelineRunner)).Sanitize()
	if !ok {
		if runner == "kubernetes" {
			return constant.PipelineRunnerKubernetes
		} else {
			return constant.PipelineRunnerDocker
		}
	}

	return result
}

func ParsePipelineStepFilterFromRequest(ctx *gin.Context, runner string) (*pipelineModel.StepFilter, error) {
	return &pipelineModel.StepFilter{
		Name:     ctx.Query(QueryParamName),
		Category: ctx.Query(QueryParamStepCategory),
	}, nil
}
