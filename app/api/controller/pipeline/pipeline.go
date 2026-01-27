package pipeline

import (
	pipelineSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
)

// Controller 处理与 pipeline 相关的 HTTP 请求
type Controller struct {
	ciConfig      config.CiConfig
	pluginService pipelineSvc.Interface
}

// NewController 创建流水线处理器实例
func NewController(ciConfig config.CiConfig, pluginService pipelineSvc.Interface) *Controller {
	return &Controller{
		ciConfig:      ciConfig,
		pluginService: pluginService,
	}
}
