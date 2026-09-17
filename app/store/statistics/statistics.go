package statistics

import (
	"context"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

// listPageSize 单次查询代码库的分页大小。
const listPageSize = 200

var _ Interface = (*Store)(nil)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

// ListGitRepos 获取所有未删除的git代码库。
func (s *Store) ListGitRepos(ctx context.Context) ([]*repoModel.Repo, error) {
	repos := make([]*repoModel.Repo, 0)

	page := 1
	for {
		var rows []*repoStore.Repository
		query := s.db.WithContext(ctx).Model(&repoStore.Repository{}).
			Where("deleted = ?", 0).
			Where("scmType = ?", "git").
			Order("id ASC")

		pageResult, err := database.PaginateDefault(ctx, query, page, listPageSize, &rows)
		if err != nil {
			return nil, database.ProcessSQLErrorf(ctx, err)
		}

		for _, row := range rows {
			repos = append(repos, &repoModel.Repo{
				ID:            row.ID,
				SpaceID:       row.SpaceID,
				Identifier:    row.Identifier,
				SCMType:       row.SCMType,
				GitUID:        row.GitUID,
				DefaultBranch: row.DefaultBranch,
				Status:        row.Status,
				Mirror:        row.Mirror,
				Deleted:       uint(row.Deleted),
			})
		}

		if int64(page*listPageSize) >= pageResult.Total || len(rows) == 0 {
			break
		}
		page++
	}

	return repos, nil
}
