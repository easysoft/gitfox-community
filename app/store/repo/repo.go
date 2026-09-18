package repo

import (
	"context"
	stderr "errors"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/paths"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
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

// errVersionConflict 表示更新时乐观锁版本冲突，由 UpdateOptLock 内部重试。
var errVersionConflict = stderr.New("version conflict")

// UpdateOptLock 使用乐观锁机制更新仓库。
// mutateFn 会在每次尝试时基于最新的仓库快照被调用；若基于 mutex 的条件更新命中 0 行，
// 方法会重新拉取最新状态并重新执行 mutateFn 直至成功或 mutateFn 返回错误。
func (s *Store) UpdateOptLock(
	ctx context.Context,
	repo *repoModel.Repo,
	mutateFn func(repo *repoModel.Repo) error,
) (*repoModel.Repo, error) {
	for {
		dup := *repo

		if err := mutateFn(&dup); err != nil {
			return nil, err
		}

		updated, err := s.updateOptLock(ctx, &dup)
		if err == nil {
			return updated, nil
		}
		if !stderr.Is(err, errVersionConflict) {
			return nil, err
		}

		latest, err := s.GetByID(ctx, repo.ID)
		if err != nil {
			return nil, err
		}
		repo = latest
	}
}

// updateOptLock 执行一次基于 mutex 的条件更新。命中 0 行视为版本冲突。
func (s *Store) updateOptLock(ctx context.Context, repo *repoModel.Repo) (*repoModel.Repo, error) {
	updates := map[string]any{
		"spaceID":       repo.SpaceID,
		"name":          repo.Identifier,
		"desc":          repo.Desc,
		"product":       repo.Product,
		"gitUID":        repo.GitUID,
		"defaultBranch": repo.DefaultBranch,
		"forkID":        repo.ForkID,
		"acl":           repo.ACL,
		"mirror":        repo.Mirror,
		"status":        repo.Status,
		"editedBy":      repo.EditedBy,
	}

	result := s.db.WithContext(ctx).Model(&Repository{}).
		Where("id = ? AND deleted = 0", repo.ID).
		Updates(updates)
	if result.Error != nil {
		return nil, database.ProcessSQLErrorf(ctx, result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, errVersionConflict
	}

	var refreshed Repository
	if err := s.db.WithContext(ctx).Where("id = ?", repo.ID).First(&refreshed).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapToModel(refreshed), nil
}

// mapToModel 将 repo 转换为 repoResponse
func mapToModel(data Repository) *repoModel.Repo {
	var path string
	if data.SpaceCode != "" {
		path = paths.Concatenate(data.SpaceCode, data.Identifier)
	}

	return &repoModel.Repo{
		ID:            data.ID,
		Identifier:    data.Identifier,
		SpaceID:       data.SpaceID,
		Desc:          data.Desc,
		SCMType:       data.SCMType,
		Product:       data.Product,
		GitUID:        data.GitUID,
		DefaultBranch: data.DefaultBranch,
		ForkID:        data.ForkID,
		Path:          path,
		ACL:           data.ACL,
		Mirror:        data.Mirror != 0,
		ProviderID:    data.ProviderID,
		Connector:     data.Connector,
		Status:        data.Status,
		CreatedBy:     data.CreatedBy,
		CreatedDate:   data.CreatedDate,
		EditedBy:      data.EditedBy,
		EditedDate:    data.EditedDate,
	}
}

// GetByID returns a repository by its ID.
func (s *Store) GetByID(ctx context.Context, id uint) (*repoModel.Repo, error) {
	var repo Repository
	err := s.db.WithContext(ctx).
		Model(&Repository{}).
		Select("ops_repo.*, ops_space.code AS spaceCode").
		Joins("LEFT JOIN ops_space ON ops_space.id = ops_repo.spaceID").
		Where("ops_repo.id = ? AND ops_repo.deleted = 0", id).
		First(&repo).Error
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

// HardDelete 物理删除仓库记录，供导入失败等需要彻底清理的场景使用。
func (s *Store) HardDelete(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Where("id = ?", id).Delete(&Repository{}).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// GetBySpaceAndIdentifier 通过 spaceID 和 identifier 查找仓库
func (s *Store) GetBySpaceAndIdentifier(ctx context.Context, spaceID uint, identifier string) (*repoModel.Repo, error) {
	var repo Repository
	err := s.db.WithContext(ctx).Where("spaceID = ? AND name = ? AND deleted = 0", spaceID, identifier).First(&repo).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapToModel(repo), nil
}

// List 分页查询仓库列表
func (s *Store) List(ctx context.Context, filter repoModel.ReqFilter) (*database.PageResult[*repoModel.Repo], error) {
	tx := s.db.WithContext(ctx).
		Model(&Repository{}).
		Select("ops_repo.*, ops_space.code AS spaceCode").
		Joins("LEFT JOIN ops_space ON ops_space.id = ops_repo.spaceID")

	// 默认只查询未删除的仓库
	tx = tx.Where("ops_repo.deleted = ?", 0)
	// 如果有 query 参数，模糊搜索 identifier
	if filter.Query != "" {
		tx = tx.Where("ops_repo.name LIKE ?", "%"+filter.Query+"%")
	}
	if filter.SpaceID != 0 {
		tx = tx.Where("ops_repo.spaceID = ?", filter.SpaceID)
	}

	// 执行分页查询（按 ID 降序排序，使用表限定列避免联查后字段歧义）
	tx = tx.Order("ops_repo.id DESC")

	var repos []*Repository
	pageResult, err := database.PaginateDefault(ctx, tx, filter.Page, filter.PageSize, &repos)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// 转换为 model 类型
	repoModels := make([]*repoModel.Repo, len(repos))
	for i, repo := range repos {
		repoModels[i] = mapToModel(*repo)
	}

	// 构造返回结果
	return &database.PageResult[*repoModel.Repo]{
		Data:     repoModels,
		Total:    pageResult.Total,
		Page:     pageResult.Page,
		PageSize: pageResult.PageSize,
	}, nil
}

// 查询迁移的历史代码库
func (s *Store) GetMigrate(ctx context.Context) ([]*repoModel.Repo, error) {
	var repos []*Repository
	err := s.db.WithContext(ctx).Where("gitUID like ? AND deleted = 0", fmt.Sprintf("%s%%", constant.MigrateRepoGitUID)).Find(&repos).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	// 转换为 model 类型
	repoModels := make([]*repoModel.Repo, len(repos))
	for i, repo := range repos {
		repoModels[i] = mapToModel(*repo)
	}

	return repoModels, nil
}

// ListAllGitRepos 一次性查询所有未删除的git代码库。
func (s *Store) ListAllGitRepos(ctx context.Context) ([]*repoModel.Repo, error) {
	var repos []*Repository
	err := s.db.WithContext(ctx).
		Where("deleted = ?", 0).
		Where("scmType = ?", "git").
		Order("id ASC").
		Find(&repos).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// 转换为 model 类型
	repoModels := make([]*repoModel.Repo, len(repos))
	for i, repo := range repos {
		repoModels[i] = mapToModel(*repo)
	}

	return repoModels, nil
}
