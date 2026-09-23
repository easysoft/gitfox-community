// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package provider

import (
	"context"

	providerModel "github.com/easysoft/gitfox-community/app/model/provider"
	"github.com/easysoft/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

// Store implements the provider store interface.
type Store struct {
	db *gorm.DB
}

// NewProviderStore creates a new provider store.
func NewProviderStore(db *gorm.DB) *Store {
	return &Store{
		db: db,
	}
}

// GetByID returns a provider by its ID.
func (s *Store) GetByID(ctx context.Context, id uint) (*providerModel.Provider, error) {
	var p Provider
	err := s.db.WithContext(ctx).Where("id = ? AND deleted = 0", id).First(&p).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapToModel(p), nil
}

func (s *Store) GetByIDs(ctx context.Context, ids []uint) (map[uint]*providerModel.Provider, error) {
	var providers []*Provider
	err := s.db.WithContext(ctx).Where("id IN ? AND deleted = 0", ids).Find(&providers).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	providerModels := make(map[uint]*providerModel.Provider, len(providers))
	for _, provider := range providers {
		providerModels[provider.ID] = mapToModel(*provider)
	}
	return providerModels, nil
}

// mapToModel converts a store Provider to the model Provider.
func mapToModel(p Provider) *providerModel.Provider {
	return &providerModel.Provider{
		ID:          p.ID,
		Type:        p.Type,
		Name:        p.Name,
		URL:         p.URL,
		Token:       p.Token,
		CreatedBy:   p.CreatedBy,
		CreatedDate: p.CreatedDate,
		EditedBy:    p.EditedBy,
		EditedDate:  p.EditedDate,
	}
}
