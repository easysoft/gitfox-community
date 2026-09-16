package pipeline

import (
	"github.com/easysoft/gitfox-community/app/service/pipeline"
	"github.com/easysoft/gitfox-community/app/service/repo"
	"github.com/easysoft/gitfox-community/pkg/config"
)

// Controller 处理与 pipeline 相关的 HTTP 请求
type Controller struct {
	ciConfig    config.CiConfig
	pipelineSvc pipeline.Interface
}

// NewController 创建流水线处理器实例
func NewController(ciConfig config.CiConfig, pipelineSvc pipeline.Interface, repoSvc repo.Interface) *Controller {
	return &Controller{
		ciConfig:    ciConfig,
		pipelineSvc: pipelineSvc,
	}
}
