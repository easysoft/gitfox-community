package space

import (
	"context"
	"fmt"
	"strings"

	spaceModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/space"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
)

type Service struct {
	// 依赖的存储接口等，可以在这里添加
	spaceStore space.Interface
	repoStore  repo.Interface
}

// NewService 创建 SpaceService 实例
func NewSpaceSvc(spaceStore space.Interface) *Service {
	return &Service{
		spaceStore: spaceStore,
	}
}

func (s *Service) Create(ctx context.Context, req spaceModel.CreateSpaceRequest) (*spaceModel.SpaceResponse, error) {
	// 直接将 CreateSpaceRequest 传递给 store 层，由 store 层进行数据转换并返回 SpaceResponse
	return s.spaceStore.Create(ctx, req)
}

func (s *Service) List(ctx context.Context, filter spaceModel.ListSpaceFilter) (*database.PageResult[*space.Space], error) {
	return s.spaceStore.List(ctx, filter)
}

func (s *Service) GetByID(ctx context.Context, id uint) (*space.Space, error) {
	spaceEntity, err := s.spaceStore.GetByID(ctx, id)
	if err != nil {
		// 传递Store层错误（如“space not found”）
		return nil, err
	}
	return spaceEntity, nil
}

func (s *Service) Update(ctx context.Context, id uint, req spaceModel.UpdateSpaceRequest) (*spaceModel.SpaceResponse, error) {
	// 直接将 UpdateSpaceRequest 传递给 store 层，由 store 层进行数据转换并返回 SpaceResponse
	return s.spaceStore.Update(ctx, id, req)
}

func (s *Service) Delete(ctx context.Context, id uint) error {

	// 1. 检查空间是否存在且未删除
	spaceObj, err := s.spaceStore.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if spaceObj.Deleted == 1 {
		return errors.NotFound.Newf("space %d has been deleted", id)
	}

	// 2. 检查空间是否有关联的代码库
	hasRepos, err := s.spaceStore.HasRepositories(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check repositories: %w", err)
	}
	if hasRepos {
		return errors.BadRequest.Newf("space %d has associated repositories, cannot delete", id)
	}

	// 3. 检查空间是否有关联的制品库
	hasArtifacts, err := s.spaceStore.HasArtifacts(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check artifacts: %w", err)
	}
	if hasArtifacts {
		return errors.BadRequest.Newf("space %d has associated artifacts, cannot delete", id)
	}

	// 4. 执行逻辑删除
	return s.spaceStore.Delete(ctx, id)
}

func (s *Service) Undelete(ctx context.Context, req spaceModel.UndeleteSpaceRequest) error {
	// 将 []uint16 转换为 []uint
	ids := make([]uint, len(req.IDs))
	for i, id := range req.IDs {
		ids[i] = uint(id)
	}

	if len(ids) == 0 {
		return errors.BadRequest.New("no space IDs provided")
	}

	// 调用 Store 层恢复空间
	err := s.spaceStore.Undelete(ctx, ids, req.EditedBy)
	if err != nil {
		// 检查是否是名称冲突错误
		errMsg := err.Error()
		if strings.Contains(errMsg, "already exists") {
			return errors.BadRequest.Wrapf(err, "cannot undelete spaces: name conflict")
		}
		return err
	}

	return nil
}

func (s *Service) CheckPermission() {
	// TODO: implement permission checks for spaces; placeholder to satisfy compilation.
}
