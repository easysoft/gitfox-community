package codereview

import (
	"context"
	"fmt"

	executionModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/execution"
	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/triggerer"
	pipelineService "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	pipelineStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
)

const (
	_codeReviewPipelineName = "_codereview"
	_codeReviewLabel        = "codereview"
	_codeReviewDescription  = "inner code review pipeline"
)

// CodeReviewService 代码审查服务
type CodeReviewService struct {
	config        *config.Config
	pipelineStore pipelineStore.Interface
	pipelineSvc   pipelineService.Interface
	triggerSvc    triggerer.Triggerer
}

// NewCodeReviewService 创建代码审查服务
func NewCodeReviewService(
	config *config.Config,
	pipelineStore pipelineStore.Interface,
	pipelineService pipelineService.Interface,
	triggerSvc triggerer.Triggerer,
) *CodeReviewService {
	return &CodeReviewService{
		config:        config,
		pipelineStore: pipelineStore,
		pipelineSvc:   pipelineService,
		triggerSvc:    triggerSvc,
	}
}

// CodeReviewInput 代码审查输入参数
type CodeReviewInput struct {
	Repo      *repoModel.Repo
	PullReqID int64
	Hook      *triggerer.Hook
}

// TriggerCodeReview 触发代码审查流水线
func (s *CodeReviewService) TriggerCodeReview(ctx context.Context, in *CodeReviewInput) (*executionModel.Execution, error) {
	// 注入代码审查参数到 Hook.Params，随流水线传递到容器环境变量
	in.Hook.Params["PPM_ID"] = fmt.Sprintf("%d", in.PullReqID)
	in.Hook.Params["REPO_ID"] = fmt.Sprintf("%d", in.Repo.ID)
	// 注入 git clone 需要的参数
	in.Hook.Params["PLUGIN_CLONE_DIR"] = in.Repo.Identifier // 供流水线中插件使用，表示代码 clone 的目录
	// 注入 gitfox 调用接口时需要的参数
	in.Hook.Params["GITFOX_API_ENDPOINT"] = s.config.URL.Base + "/api/v2"
	in.Hook.Params["PLUGIN_TOKEN"] = s.config.PipelineAuth.Password
	// 注入评审所需参数（OCR LLM 配置）
	// in.Hook.Params["OCR_LLM_PROVIDER"] = "zai"
	// in.Hook.Params["OCR_LLM_URL"] = "https://zai-new.sh.oop.cc/v1"
	// in.Hook.Params["OCR_LLM_PROTOCOL"] = "openai"
	// in.Hook.Params["OCR_LLM_API_KEY"] = "sk-uGaJPSqEgQJpXLtt46fWTXlNqA4fvcGI0pUGbCVRoryTS2G9"
	// in.Hook.Params["OCR_LLM_MODEL"] = "gpt-4o-mini"
	// in.Hook.Params["OCR_LLM_LANGUAGE"] = "中文"

	// 获取或创建代码审查流水线
	pipeline, err := s.getReviewPipeline(ctx, int64(in.Repo.ID), in.Hook)
	if err != nil {
		return nil, err
	}

	// 触发流水线执行
	execution, err := s.triggerSvc.Trigger(ctx, uint(pipeline.ID), in.Hook)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger code review pipeline: %w", err)
	}

	return execution, nil
}

// getReviewPipeline 获取或创建代码审查流水线，已存在时用最新内置 yaml 覆盖其内容
func (s *CodeReviewService) getReviewPipeline(ctx context.Context, repoID int64, hook *triggerer.Hook) (*pipelineModel.Pipeline, error) {
	if hook == nil {
		return nil, fmt.Errorf("hook is required to get code review pipeline")
	}

	pipeline, err := s.pipelineStore.GetPipelineByRepoIDAndName(ctx, repoID, _codeReviewPipelineName)
	if err != nil {
		return nil, err
	}
	if pipeline == nil {
		// 如果没有找到流水线，创建流水线及其内容
		pipeline, err = s.createPipeline(ctx, repoID, hook)
		if err != nil {
			return nil, err
		}
	} else {
		// 如果流水线已存在，更新流水线对应的 content（覆盖原有内容）
		if err := s.updatePipelineContent(ctx, pipeline.ID); err != nil {
			return nil, err
		}
	}

	return pipeline, nil
}

// createPipeline 创建代码审查流水线
func (s *CodeReviewService) createPipeline(ctx context.Context, repoID int64, hook *triggerer.Hook) (*pipelineModel.Pipeline, error) {
	// 创建流水线
	pipelineType := "repo"
	if repoID == 0 {
		pipelineType = "space"
	}
	p := &pipelineModel.Pipeline{
		Name:          _codeReviewPipelineName,
		Scope:         pipelineType,
		RepoID:        repoID,
		Desc:          _codeReviewDescription,
		YamlPath:      pipelineModel.PipelineConfigPathInnerPrefix + _codeReviewLabel + ".yaml",
		DefaultBranch: hook.Ref,
		CreatedBy:     hook.Sender,
	}

	return s.pipelineSvc.CreateEmbeddedPipeline(ctx, _codeReviewLabel, p)
}

// updatePipelineContent 更新代码审查流水线内容
func (s *CodeReviewService) updatePipelineContent(ctx context.Context, pipelineID int64) error {
	return s.pipelineSvc.UpdateEmbeddedPipelineContent(ctx, _codeReviewLabel, pipelineID)
}
