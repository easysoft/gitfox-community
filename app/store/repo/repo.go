package repo

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"

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

func (s *Store) FindByRef(ctx context.Context, ref string) (*Repository, error) {
	var repo Repository //todo
	if err := s.db.WithContext(ctx).Where("ref = ?", ref).First(&repo).Error; err != nil {
		return nil, err
	}
	return &repo, nil
}

func (s *Store) Create(ctx context.Context, repo *Repository) error {
	return nil
}

// GetByID returns a repository by its ID.
// If no repository is found with the given ID, it returns an error.
// The repository is not checked for existence before calling this method, so if the repository does not exist, an error will be returned.
func (s *Store) GetByID(ctx context.Context, id uint) (*Repository, error) {
	var repo Repository
	repo.ID = id
	err := s.db.WithContext(ctx).First(&repo).Where("deleted = 0").Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound.Wrapf(err, "repository %d not found", id)
	}

	if err != nil {
		return nil, err
	}

	return &repo, nil
}
