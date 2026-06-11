package provider

import (
	"context"

	providerModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/provider"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
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
