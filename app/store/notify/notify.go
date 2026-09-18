package notify

import (
	"context"

	"github.com/easysoft/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

type Store struct {
	db *gorm.DB
}

func NewNotifyStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(ctx context.Context, notify *Notify) error {
	if err := s.db.WithContext(ctx).Create(notify).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (s *Store) Update(ctx context.Context, id uint, status string) error {
	if err := s.db.WithContext(ctx).Model(&Notify{}).
		Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}
