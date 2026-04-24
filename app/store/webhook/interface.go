package webhook

import (
	"context"

	webhookModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/webhook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

type Interface interface {
	Create(ctx context.Context, webhook *Webhook) (*webhookModel.Webhook, error)
	GetByID(ctx context.Context, id uint) (*webhookModel.Webhook, error)
	GetByRepoIDAndDisplayName(ctx context.Context, repoID int64, displayName string) (*webhookModel.Webhook, error)
	List(ctx context.Context, repoID int64, filter webhookModel.ReqForList) (*database.PageResult[*webhookModel.Webhook], error)
	ListByParentInfo(ctx context.Context, parents []webhookModel.ParentInfo) ([]*webhookModel.Webhook, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) (*webhookModel.Webhook, error)
	Delete(ctx context.Context, id uint) error
	ListExecutions(ctx context.Context, webhookID uint, filter webhookModel.ReqForList) (*database.PageResult[*webhookModel.Execution], error)
	ListExecutionsByTriggerID(ctx context.Context, triggerID string) ([]*webhookModel.Execution, error)
	GetExecutionByID(ctx context.Context, executionID uint) (*webhookModel.Execution, error)
	CreateExecution(ctx context.Context, execution *webhookModel.Execution) error
}
