package repo

import (
	"context"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/paths"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

type Store struct {
	db *gorm.DB
}

func NewRepoStore(db *gorm.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) FindByRef(ctx context.Context, ref string) (*repoModel.Repo, error) {
	var repo Repository //todo
	if err := s.db.WithContext(ctx).Where("ref = ?", ref).First(&repo).Error; err != nil {
		return nil, err
	}
	return mapToModel(repo), nil
}

func (s *Store) Create(ctx context.Context, repo *Repository) (*repoModel.Repo, error) {
	if err := s.db.WithContext(ctx).Create(&repo).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapToModel(*repo), nil
}

func (s *Store) Update(ctx context.Context, repo *Repository, id uint) (*repoModel.Repo, error) {
	repo.ID = id
	if err := s.db.WithContext(ctx).Updates(repo).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// 重新查询完整记录,因为 Updates 只更新指定非零值字段,不会回填其他字段
	var updated Repository
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&updated).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return mapToModel(updated), nil
}

// mapToModel 将 repo 转换为 repoResponse
func mapToModel(data Repository) *repoModel.Repo {
	path := paths.Concatenate(fmt.Sprint(data.SpaceID), data.Identifier)
	return &repoModel.Repo{
		ID:            data.ID,
		Identifier:    data.Identifier,
		SpaceID:       data.SpaceID,
		Desc:          data.Desc,
		GitUID:        data.GitUID,
		DefaultBranch: data.DefaultBranch,
		ForkID:        data.ForkID,
		Path:          path,
		Empty:         data.Empty,
		ACL:           data.ACL,
		Mirror:        data.Mirror,
		CreatedBy:     data.CreatedBy,
		CreatedDate:   data.CreatedDate,
		EditedBy:      data.EditedBy,
		EditedDate:    data.EditedDate,
	}
}

// GetByID returns a repository by its ID.
// If no repository is found with the given ID, it returns an error.
// The repository is not checked for existence before calling this method, so if the repository does not exist, an error will be returned.
func (s *Store) GetByID(ctx context.Context, id uint) (*repoModel.Repo, error) {
	var repo Repository
	err := s.db.WithContext(ctx).Where("id = ? AND deleted = 0", id).First(&repo).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return mapToModel(repo), nil
}


func (s *Store) Delete(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Model(&Repository{}).Where("id = ?", id).Update("deleted", 1).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}