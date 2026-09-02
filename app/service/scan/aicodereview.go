package scan

import (
	"context"
	"fmt"

	executionModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/execution"
	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/triggerer"
	pipelineService "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/utils"
	pipelineStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/pipeline"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/token"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
)

const (
	_codeReviewPipelineName = "_codereview"
	_codeReviewLabel        = "codereview"
	_codeReviewDescription  = "inner code review pipeline"
)

// AIReviewService AI 代码评审服务
type AIReviewService struct {
	aireviewStore scanStore.IAIReview
	config        *config.Config
	pipelineStore pipelineStore.Interface
	pipelineSvc   pipelineService.Interface
	triggerSvc    triggerer.Triggerer
	tokenStore    token.Interface
}

var _ IAIReview = (*AIReviewService)(nil)

func NewService(
	aireviewStore scanStore.IAIReview,
	config *config.Config,
	pipelineStore pipelineStore.Interface,
	pipelineService pipelineService.Interface,
	triggerSvc triggerer.Triggerer,
	tokenStore token.Interface,
) *AIReviewService {
	return &AIReviewService{
		aireviewStore: aireviewStore,
		config:        config,
		pipelineStore: pipelineStore,
		pipelineSvc:   pipelineService,
		triggerSvc:    triggerSvc,
		tokenStore:    tokenStore,
	}
}

func (s *AIReviewService) GetAIReviewRules(ctx context.Context) ([]*model.AIReviewRule, error) {
	rules, err := s.aireviewStore.GetRules(ctx)
	if err != nil {
		return rules, err
	}

	return rules, nil
}

// CodeReviewInput 代码审查输入参数
type CodeReviewInput struct {
	Repo      *repoModel.Repo
	PullReqID int64
	Hook      *triggerer.Hook
}

// TriggerCodeReview 触发代码审查流水线
func (s *AIReviewService) TriggerCodeReview(ctx context.Context, in *CodeReviewInput) (*executionModel.Execution, error) {
	// 注入代码审查参数到 Hook.Params，随流水线传递到容器环境变量
	in.Hook.Params["PPM_ID"] = fmt.Sprintf("%d", in.PullReqID)
	in.Hook.Params["REPO_ID"] = fmt.Sprintf("%d", in.Repo.ID)
	// 注入 git clone 需要的参数
	in.Hook.Params["PLUGIN_CLONE_DIR"] = in.Repo.Identifier // 供流水线中插件使用，表示代码 clone 的目录
	// 注入 gitfox 调用接口时需要的参数
	in.Hook.Params["GITFOX_API_ENDPOINT"] = s.config.URL.Base + "/api/v2"
	in.Hook.Params["PLUGIN_TOKEN"] = s.config.PipelineAuth.Password
	// 注入评审所需参数（OCR LLM 配置），从 DevOps 平台获取 AI 模型设置
	entry, err := s.tokenStore.GetDevopsToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get devops token: %w", err)
	}
	aiSetting, err := utils.GetAIModelSetting(ctx, s.config.Devops.URL, entry.Key)
	if err != nil || !aiSetting.Enable {
		return nil, fmt.Errorf("failed to get AI model setting: %w", err)
	}
	in.Hook.Params["OCR_LLM_PROVIDER"] = "zai"
	in.Hook.Params["OCR_LLM_URL"] = aiSetting.ZaiURL
	in.Hook.Params["OCR_LLM_PROTOCOL"] = "openai"
	in.Hook.Params["OCR_LLM_API_KEY"] = aiSetting.ZaiToken
	in.Hook.Params["OCR_LLM_MODEL"] = aiSetting.Model
	if aiSetting.ZaiLanguage == "chinese" {
		in.Hook.Params["OCR_LLM_LANGUAGE"] = "中文"
	} else {
		in.Hook.Params["OCR_LLM_LANGUAGE"] = "English"
	}

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
func (s *AIReviewService) getReviewPipeline(ctx context.Context, repoID int64, hook *triggerer.Hook) (*pipelineModel.Pipeline, error) {
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
func (s *AIReviewService) createPipeline(ctx context.Context, repoID int64, hook *triggerer.Hook) (*pipelineModel.Pipeline, error) {
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
func (s *AIReviewService) updatePipelineContent(ctx context.Context, pipelineID int64) error {
	return s.pipelineSvc.UpdateEmbeddedPipelineContent(ctx, _codeReviewLabel, pipelineID)
}
