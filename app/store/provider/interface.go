// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package provider

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=$GOFILE -destination=mock/provider_store.go -package=mock

import (
	"context"

	providerModel "github.com/easysoft/gitfox-community/app/model/provider"
)

// Interface defines the provider store operations.
type Interface interface {
	// GetByID returns a provider by its ID.
	GetByID(ctx context.Context, id uint) (*providerModel.Provider, error)
	GetByIDs(ctx context.Context, ids []uint) (map[uint]*providerModel.Provider, error)
}
