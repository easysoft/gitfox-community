// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package webhook

import (
	"context"

	webhookModel "github.com/easysoft/gitfox-community/app/model/webhook"
	"github.com/easysoft/gitfox-community/pkg/database"
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
