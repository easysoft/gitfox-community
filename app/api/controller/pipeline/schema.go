package pipeline

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// swagger 类型引用
var (
	_ pipelineModel.GroupedPlugins
)

// @Summary 查询流水线步骤分组
// @Description 查询流水线步骤分组
// @Tags pipelines
// @Accept application/json
// @Produce application/json
// @Success 200 {object} resp.ListResponse{data=[]pipelineModel.GroupedPlugins,listArgs=resp.ListNoTotalArgs} "查询成功"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /pipeline/steps/grouped [get]
func (c *Controller) GroupSteps(ctx *gin.Context) {
	// 调用 service 层查询插件分组数据
	groupedPlugins, err := c.pipelineSvc.GetGroupedPlugins(ctx.Request.Context())
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.ListSuccess(ctx, groupedPlugins, nil, i18n.PipelineQuerySchemaGroupSuccess, nil)
}

// @Summary 查询流水线步骤表单配置
// @Description 查询流水线步骤表单配置
// @Tags pipelines
// @Accept application/json
// @Produce application/json
// @Param name path string true "步骤名称"
// @Success 200 {object} resp.ObjectResponse "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipeline/steps/{name}/schema [get]
func (c *Controller) GetSchema(ctx *gin.Context) {
	// 获取路径参数 name
	stepName, err := req.PathParam(ctx, req.PathParamSchemaName)
	if err != nil {
		return
	}

	// 调用 service 层根据 name 查询插件
	plugin, err := c.pipelineSvc.GetPluginByName(ctx.Request.Context(), stepName)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	// 返回插件的 json 字段内容
	resp.OK(ctx, i18n.PipelineQueryStepSchemaSuccess, plugin.Json)
}
