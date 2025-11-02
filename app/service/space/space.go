package space

import (
	"context"
	"fmt"
	"strings"

	spaceModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/space"
	pkg_errors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
)

type SpaceSvc struct {
	// 依赖的存储接口等，可以在这里添加
	spaceStore space.SpaceInterface
}

// NewService 创建 SpaceService 实例
func NewSpaceSvc(spaceStore space.SpaceInterface) *SpaceSvc {
	return &SpaceSvc{
		spaceStore: spaceStore,
	}
}

func (s *SpaceSvc) Create(ctx context.Context, req spaceModel.CreateSpaceRequest) (*spaceModel.SpaceResponse, error) {
	// 直接将 CreateSpaceRequest 传递给 store 层，由 store 层进行数据转换并返回 SpaceResponse
	return s.spaceStore.Create(ctx, req)
}

func (s *SpaceSvc) List(ctx context.Context, filter spaceModel.ListSpaceFilter) ([]spaceModel.SpaceResponse, error) {
	// Implementation logic here
	return nil, nil
}

func (s *SpaceSvc) Get(ctx context.Context, id uint) (*spaceModel.SpaceResponse, error) {
	// Implementation logic here
	return nil, nil
}

func (s *SpaceSvc) Update(ctx context.Context, id uint, req spaceModel.UpdateSpaceRequest) (*spaceModel.SpaceResponse, error) {
	// 直接将 UpdateSpaceRequest 传递给 store 层，由 store 层进行数据转换并返回 SpaceResponse
	return s.spaceStore.Update(ctx, id, req)
}

func (s *SpaceSvc) Delete(ctx context.Context, id uint) error {

	// 1. 检查空间是否存在且未删除
	spaceObj, err := s.spaceStore.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if spaceObj.Deleted == 1 {
		return pkg_errors.NotFound.Newf("space %d has been deleted", id)
	}

	// 2. 检查空间是否有关联的代码库
	hasRepos, err := s.spaceStore.HasRepositories(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check repositories: %w", err)
	}
	if hasRepos {
		return pkg_errors.BadRequest.Newf("space %d has associated repositories, cannot delete", id)
	}

	// 3. 检查空间是否有关联的制品库
	hasArtifacts, err := s.spaceStore.HasArtifacts(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check artifacts: %w", err)
	}
	if hasArtifacts {
		return pkg_errors.BadRequest.Newf("space %d has associated artifacts, cannot delete", id)
	}

	// 4. 执行逻辑删除
	return s.spaceStore.Delete(ctx, id)
}

func (s *SpaceSvc) Undelete(ctx context.Context, req spaceModel.UndeleteSpaceRequest) error {
	// 将 []uint16 转换为 []uint
	ids := make([]uint, len(req.IDs))
	for i, id := range req.IDs {
		ids[i] = uint(id)
	}

	if len(ids) == 0 {
		return pkg_errors.BadRequest.New("no space IDs provided")
	}

	// 调用 Store 层恢复空间
	err := s.spaceStore.Undelete(ctx, ids, req.EditedBy)
	if err != nil {
		// 检查是否是名称冲突错误
		errMsg := err.Error()
		if strings.Contains(errMsg, "already exists") {
			return pkg_errors.BadRequest.Wrapf(err, "cannot undelete spaces: name conflict")
		}
		return err
	}

	return nil
}
