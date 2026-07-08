package space

import (
	"context"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	spaceModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/sse"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/space"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/auth/authz"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
)

type Service struct {
	spaceStore space.Interface
	repoStore  repo.Interface
	authorizer authz.Authorizer

	sseSvc sse.Streamer
}

// NewService 创建 SpaceService 实例
func NewSpaceSvc(spaceStore space.Interface, repoStore repo.Interface, authorizer authz.Authorizer, sseService sse.Streamer) *Service {
	return &Service{
		spaceStore: spaceStore,
		repoStore:  repoStore,
		authorizer: authorizer,
		sseSvc:     sseService,
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
		return nil, err
	}
	return spaceEntity, nil
}

func (s *Service) GetByCode(ctx context.Context, code string) (*space.Space, error) {
	spaceEntity, err := s.spaceStore.GetByCode(ctx, code)
	if err != nil {
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

func (s *Service) Restore(ctx context.Context, id uint) error {
	// 调用 Store 层恢复空间
	err := s.spaceStore.Restore(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetSpaceCheckAccess(ctx context.Context, session *auth.Session, spaceCode string, permissions []constant.Permission) (*space.Space, error) {
	spaceObj, err := s.spaceStore.GetByCode(ctx, spaceCode)
	if err != nil {
		return nil, err
	}

	// 检查权限
	for _, v := range permissions {
		ok, err := s.authorizer.Check(ctx, session,
			constant.Scope{
				ResourceType: constant.ResourceTypeSpace,
				ResourceID:   spaceObj.ID,
			},
			v,
		)
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, constant.ErrUnauthorized
		}
	}

	return spaceObj, nil
}

func (s *Service) Events(ctx context.Context, spaceID uint) (<-chan *sse.Event, <-chan error, func(context.Context) error, error) {
	_, err := s.GetByID(ctx, spaceID)
	if err != nil {
		return nil, nil, nil, err
	}

	chEvents, chErr, sseCancel := s.sseSvc.Stream(ctx, int64(spaceID))

	return chEvents, chErr, sseCancel, nil
}
