package webhook

import (
	"context"

	webhookModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/webhook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

type Interface interface {
	Create(ctx context.Context, info webhookModel.Webhook) (*webhookModel.Webhook, error)
	Get(ctx context.Context, webhookID uint) (*webhookModel.Webhook, error)
	List(ctx context.Context, repoID int64, filter webhookModel.ReqForList) (*database.PageResult[*webhookModel.Webhook], error)
	Update(ctx context.Context, webhookID uint, req webhookModel.ReqForUpdate, updatedBy string) (*webhookModel.Webhook, error)
	Delete(ctx context.Context, webhookID uint) error
}
