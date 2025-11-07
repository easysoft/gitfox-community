package repo

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

type Store struct {
	db *gorm.DB
}

func NewRepoStore(db *gorm.DB) Interface {
	return &Store{
		db: db,
	}
}

func (s *Store) FindByRef(ctx context.Context, ref string) (*repo.Repository, error) {
	var repo repo.Repository //todo
	if err := s.db.WithContext(ctx).Where("ref = ?", ref).First(&repo).Error; err != nil {
		return nil, err
	}
	return &repo, nil
}
