// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package pipeline

import (
	"context"
	"encoding/base64"
	"fmt"

	pipelineModel "github.com/easysoft/gitfox-community/app/model/pipeline"
	"github.com/easysoft/gitfox-community/resources"
)

// CreateEmbeddedPipeline 创建内置流水线（从 embedded 资源读取 YAML 配置）
// yamlLabel: 流水线标签，对应 resources/pipeline/{label}.yaml
// pipeline: 流水线基本信息（Name, Scope, RepoID, Desc, YamlPath, DefaultBranch, CreatedBy）
func (s *Service) CreateEmbeddedPipeline(ctx context.Context, yamlLabel string, pipeline *pipelineModel.Pipeline) (*pipelineModel.Pipeline, error) {
	// 读取 embedded 的流水线配置并 base64 编码
	encodedData, err := s.readAndEncodeEmbeddedPipeline(yamlLabel)
	if err != nil {
		return nil, err
	}

	// 创建流水线
	createdPipeline, err := s.pipelineStore.CreatePipeline(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline: %w", err)
	}

	// 创建流水线内容（base64 编码后存入）
	content := &pipelineModel.PipelineContent{
		PipelineID: createdPipeline.ID,
		Data:       []byte(encodedData),
		CreatedBy:  pipeline.CreatedBy,
	}
	if _, err = s.pipelineStore.CreatePipelineContent(ctx, content); err != nil {
		return nil, fmt.Errorf("failed to create pipeline content: %w", err)
	}

	return createdPipeline, nil
}

// UpdateEmbeddedPipelineContent 更新内置流水线内容（从 embedded 资源读取 YAML 配置）
// yamlLabel: 流水线标签，对应 resources/pipeline/{label}.yaml
// pipelineID: 流水线 ID
func (s *Service) UpdateEmbeddedPipelineContent(ctx context.Context, yamlLabel string, pipelineID int64) error {
	// 读取 embedded 的流水线配置并 base64 编码
	encodedData, err := s.readAndEncodeEmbeddedPipeline(yamlLabel)
	if err != nil {
		return err
	}

	// 更新流水线内容（base64 编码后存入）
	content := &pipelineModel.PipelineContent{
		PipelineID: pipelineID,
		Data:       []byte(encodedData),
	}
	return s.pipelineStore.UpdatePipelineContent(ctx, content)
}

// readAndEncodeEmbeddedPipeline 读取 embedded 的流水线配置并进行 base64 编码
func (s *Service) readAndEncodeEmbeddedPipeline(yamlLabel string) (string, error) {
	yamlFile := "pipeline/" + yamlLabel + ".yaml"

	// 读取 embed 的流水线配置
	data, err := resources.EmbeddedPipelines.ReadFile(yamlFile)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded pipeline config %s: %w", yamlFile, err)
	}

	// base64 编码
	return base64.StdEncoding.EncodeToString(data), nil
}
