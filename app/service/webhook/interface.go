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
	Create(ctx context.Context, info webhookModel.Webhook) (*webhookModel.Webhook, error)
	Get(ctx context.Context, webhookID uint) (*webhookModel.Webhook, error)
	List(ctx context.Context, repoID int64, filter webhookModel.ReqForList) (*database.PageResult[*webhookModel.Webhook], error)
	Update(ctx context.Context, webhookID uint, req webhookModel.ReqForUpdate, updatedBy string) (*webhookModel.Webhook, error)
	Delete(ctx context.Context, webhookID uint) error
	ListExecutions(ctx context.Context, webhookID uint, filter webhookModel.ReqForList) (*database.PageResult[*webhookModel.Execution], error)
	GetExecution(ctx context.Context, executionID uint) (*webhookModel.Execution, error)
	RetriggerExecution(ctx context.Context, executionID uint) (*webhookModel.TriggerResult, error)
}
