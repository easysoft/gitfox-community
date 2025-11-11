package repo

import (
	"context"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
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

func (s *Store) Create(ctx context.Context, repo *Repository) (*repoModel.CreateRepoResp, error) {
	if err := s.db.WithContext(ctx).Create(&repo).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapToModel(*repo), nil
}

// mapToModel 将 Space 转换为 SpaceResponse
func mapToModel(data Repository) *repoModel.CreateRepoResp {
	return &repoModel.CreateRepoResp{
		ID:            data.ID,
		Identifier:    data.Identifier,
		SpaceID:       data.SpaceID,
		Desc:          data.Desc,
		GitUID:        data.GitUID,
		DefaultBranch: data.DefaultBranch,
		ForkID:        data.ForkID,
		Empty:         data.Empty,
		ACL:           data.ACL,
		Mirror:        data.Mirror,
		CreatedBy:     data.CreatedBy,
		CreatedDate:   data.CreatedDate,
		EditedBy:      data.EditedBy,
		EditedDate:    data.EditedDate,
	}
}
