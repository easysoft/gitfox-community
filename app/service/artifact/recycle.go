// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package artifact

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/easysoft/gitfox-community/pkg/utils/datetime"

	model "github.com/easysoft/gitfox-community/app/model/artifact"
	pkgErrors "github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
)

// =============================================================================
// 回收站接口
//   1. ListRecycleBin —— 按 kind 分页查询回收站记录（kind 必填，五选一）
//   2. (Restore 待开斩)
// =============================================================================

// ListRecycleBin 列出回收站中的指定 kind 实体。
//
// 行为约束：
//   - 仅查 deleted=1 的记录，不再做时间窗过滤；前端按 expiresAt 自行展示距离 GC 多久
//   - kind 必须是 AllArtifactRecycleKinds 集合内的值，由 controller 已校验，service 兜底再判一次
//   - 全局视角，不接受 spaceID/repoID/artifactID 过滤
//   - 关联展示字段（artifactID/format/group/path/name）由本函数按 kind 各自反查 store 补齐
//
// 返回 (list, total, error)；total 为该 kind 的回收站总条数，供前端做分页。
func (s *Service) ListRecycleBin(
	ctx context.Context, req *model.ArtifactRecycleListReq,
) ([]*model.ArtifactRecycleItem, int64, error) {
	kind := req.Kind
	if kind == "" {
		kind = model.ArtifactRecycleKindAsset
	}
	if !isValidRecycleKind(kind) {
		return nil, 0, pkgErrors.WithI18nKey(
			fmt.Errorf("invalid recycle kind: %s", kind),
			i18n.ArtifactRecycleKindInvalid,
		)
	}

	switch kind {
	case model.ArtifactRecycleKindArtifact:
		return s.listRecycleArtifact(ctx, req.Page, req.PageSize)
	case model.ArtifactRecycleKindGroup:
		return s.listRecycleGroup(ctx, req.Page, req.PageSize)
	case model.ArtifactRecycleKindPackage:
		return s.listRecyclePackage(ctx, req.Page, req.PageSize)
	case model.ArtifactRecycleKindVersion:
		return s.listRecycleVersion(ctx, req.Page, req.PageSize)
	case model.ArtifactRecycleKindAsset:
		return s.listRecycleAsset(ctx, req.Page, req.PageSize)
	}
	// 不可达：上方已校验，留兜底防漏
	return nil, 0, pkgErrors.WithI18nKey(
		fmt.Errorf("unhandled recycle kind: %s", kind),
		i18n.ArtifactRecycleKindInvalid,
	)
}

// isValidRecycleKind 校验 kind 是否在合法枚举集合内。
func isValidRecycleKind(kind model.ArtifactRecycleKind) bool {
	for _, k := range model.AllArtifactRecycleKinds {
		if k == kind {
			return true
		}
	}
	return false
}

// expiresAt 计算回收站记录的过期时间：deletedDate + RecycleBinRetention。
// deletedDate 为空时返回 nil，前端表现为字段缺失。
func (s *Service) expiresAt(deletedAt *datetime.Local) *datetime.Local {
	if deletedAt == nil {
		return nil
	}
	t := time.Time(*deletedAt).Add(s.config.RecycleBinRetention)
	expires := datetime.Local(t)
	return &expires
}

// expiresAtFromTime 与 expiresAt 同义，但接收 *time.Time 入参（package 表 EditedDate 类型为 *time.Time）。
func (s *Service) expiresAtFromTime(deletedAt *time.Time) *datetime.Local {
	if deletedAt == nil {
		return nil
	}
	t := deletedAt.Add(s.config.RecycleBinRetention)
	expires := datetime.Local(t)
	return &expires
}

// localFromTimePtr 把 *time.Time 适配为 *datetime.Local，仅用于 package 模型字段类型与展示模型对齐。
func localFromTimePtr(t *time.Time) *datetime.Local {
	if t == nil {
		return nil
	}
	v := datetime.Local(*t)
	return &v
}

// =============================================================================
// 各 kind 的列表实现
// =============================================================================

// listRecycleArtifact 制品库（view）维度回收站。
func (s *Service) listRecycleArtifact(ctx context.Context, page, pageSize int) ([]*model.ArtifactRecycleItem, int64, error) {
	rows, total, err := s.artStore.Views().ListRecycle(ctx, page, pageSize)
	if err != nil {
		return nil, 0, pkgErrors.WithI18nKey(err, i18n.ArtifactRecycleListError)
	}
	out := make([]*model.ArtifactRecycleItem, 0, len(rows))
	for _, v := range rows {
		out = append(out, &model.ArtifactRecycleItem{
			Kind:        model.ArtifactRecycleKindArtifact,
			EntityID:    (&model.ArtifactTreeNodeId{Type: model.ArtifactTreeNodeTypeArtifact, Pk: v.ID}).String(),
			Id:          v.ID,
			Name:        v.Name,
			Format:      v.Format,
			ArtifactID:  v.ID,
			DeletedBy:   v.EditedBy,
			DeletedDate: v.EditedDate,
			ExpiresAt:   s.expiresAt(v.EditedDate),
		})
	}
	return out, total, nil
}

// listRecycleGroup 分组维度回收站。Path 由父链反查拼成「a.b.c」形式。
func (s *Service) listRecycleGroup(ctx context.Context, page, pageSize int) ([]*model.ArtifactRecycleItem, int64, error) {
	rows, total, err := s.artStore.Groups().ListRecycle(ctx, page, pageSize)
	if err != nil {
		return nil, 0, pkgErrors.WithI18nKey(err, i18n.ArtifactRecycleListError)
	}
	out := make([]*model.ArtifactRecycleItem, 0, len(rows))
	for _, g := range rows {
		path := g.Name
		if endpoint, eErr := s.artStore.Groups().GetEndpointById(ctx, g.ID); eErr == nil && endpoint != nil {
			path = endpoint.GroupName()
		}
		out = append(out, &model.ArtifactRecycleItem{
			Kind:        model.ArtifactRecycleKindGroup,
			EntityID:    (&model.ArtifactTreeNodeId{Type: model.ArtifactTreeNodeTypeGroup, Pk: g.ID}).String(),
			Id:          g.ID,
			Name:        g.Name,
			Path:        path,
			Format:      g.Format,
			ArtifactID:  g.ArtifactID,
			DeletedBy:   g.EditedBy,
			DeletedDate: g.EditedDate,
			ExpiresAt:   s.expiresAt(g.EditedDate),
		})
	}
	return out, total, nil
}

// listRecyclePackage 制品包维度回收站。Group 取 namespace（点分形式）。
func (s *Service) listRecyclePackage(ctx context.Context, page, pageSize int) ([]*model.ArtifactRecycleItem, int64, error) {
	rows, total, err := s.artStore.Packages().ListRecycle(ctx, page, pageSize)
	if err != nil {
		return nil, 0, pkgErrors.WithI18nKey(err, i18n.ArtifactRecycleListError)
	}
	out := make([]*model.ArtifactRecycleItem, 0, len(rows))
	for _, p := range rows {
		out = append(out, &model.ArtifactRecycleItem{
			Kind:        model.ArtifactRecycleKindPackage,
			EntityID:    (&model.ArtifactTreeNodeId{Type: model.ArtifactTreeNodeTypePackage, Pk: p.ID}).String(),
			Id:          p.ID,
			Name:        p.Name,
			Group:       p.Namespace,
			Format:      p.Format,
			ArtifactID:  p.ArtifactID,
			DeletedBy:   p.EditedBy,
			DeletedDate: localFromTimePtr(p.EditedDate),
			ExpiresAt:   s.expiresAtFromTime(p.EditedDate),
		})
	}
	return out, total, nil
}

// listRecycleVersion 版本维度回收站。需经 package 反查 artifactID/format/group/包名。
func (s *Service) listRecycleVersion(ctx context.Context, page, pageSize int) ([]*model.ArtifactRecycleItem, int64, error) {
	rows, total, err := s.artStore.Versions().ListRecycle(ctx, page, pageSize)
	if err != nil {
		return nil, 0, pkgErrors.WithI18nKey(err, i18n.ArtifactRecycleListError)
	}
	out := make([]*model.ArtifactRecycleItem, 0, len(rows))
	for _, v := range rows {
		item := &model.ArtifactRecycleItem{
			Kind:        model.ArtifactRecycleKindVersion,
			EntityID:    (&model.ArtifactTreeNodeId{Type: model.ArtifactTreeNodeTypeVersion, Pk: v.ID}).String(),
			Id:          v.ID,
			Version:     v.Version,
			DeletedBy:   v.EditedBy,
			DeletedDate: v.EditedDate,
			ExpiresAt:   s.expiresAt(v.EditedDate),
		}
		if pkg, pErr := s.artStore.Packages().GetByID(ctx, v.PackageID); pErr == nil && pkg != nil {
			item.Name = pkg.Name
			item.Group = pkg.Namespace
			item.Format = pkg.Format
			item.ArtifactID = artifactIDForVersion(pkg, v)
		}
		out = append(out, item)
	}
	return out, total, nil
}

// listRecycleAsset 资产维度回收站。Asset 自身可能未带 artifactID/format，需经 version/group/artifact 三类归属反查。
func (s *Service) listRecycleAsset(ctx context.Context, page, pageSize int) ([]*model.ArtifactRecycleItem, int64, error) {
	rows, total, err := s.artStore.Assets().ListRecycle(ctx, page, pageSize)
	if err != nil {
		return nil, 0, pkgErrors.WithI18nKey(err, i18n.ArtifactRecycleListError)
	}
	out := make([]*model.ArtifactRecycleItem, 0, len(rows))
	for _, a := range rows {
		item := &model.ArtifactRecycleItem{
			Kind:        model.ArtifactRecycleKindAsset,
			EntityID:    (&model.ArtifactTreeNodeId{Type: model.ArtifactTreeNodeTypeAsset, Pk: a.ID}).String(),
			Id:          a.ID,
			Name:        assetBaseName(a.Path),
			Format:      a.Format,
			DeletedBy:   a.EditedBy,
			DeletedDate: a.EditedDate,
			ExpiresAt:   s.expiresAt(a.EditedDate),
		}
		s.fillAssetOwnership(ctx, a, item)
		out = append(out, item)
	}
	return out, total, nil
}

// fillAssetOwnership 按归属链路（version → group → artifactID）回填 asset item 的 ArtifactID/Group。
// asset 可能直挂 version、group 或 view，三种来源逐一尝试，命中即返回。
func (s *Service) fillAssetOwnership(ctx context.Context, a *model.ArtifactAsset, item *model.ArtifactRecycleItem) {
	switch {
	case a.VersionID.Valid && a.VersionID.Int64 > 0:
		ver, vErr := s.artStore.Versions().GetByID(ctx, a.VersionID.Int64)
		if vErr != nil || ver == nil {
			return
		}
		item.Version = ver.Version
		pkg, pErr := s.artStore.Packages().GetByID(ctx, ver.PackageID)
		if pErr != nil || pkg == nil {
			return
		}
		item.Group = pkg.Namespace
		item.ArtifactID = artifactIDForVersion(pkg, ver)
		if item.Format == "" {
			item.Format = pkg.Format
		}
	case a.GroupID.Valid && a.GroupID.Int64 > 0:
		endpoint, eErr := s.artStore.Groups().GetEndpointById(ctx, a.GroupID.Int64)
		if eErr != nil || endpoint == nil || endpoint.ArtifactGroup == nil {
			return
		}
		item.Group = endpoint.GroupName()
		item.ArtifactID = endpoint.ArtifactGroup.ArtifactID
		if item.Format == "" {
			item.Format = endpoint.ArtifactGroup.Format
		}
	case a.ArtifactID.Valid && a.ArtifactID.Int64 > 0:
		item.ArtifactID = a.ArtifactID.Int64
	}
}

// assetBaseName 返回 asset 路径的最终段，作为前端展示的「文件名」。
// container 格式的 path 是 "sha256:..." digest，原样返回。
func assetBaseName(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx < 0 {
		return path
	}
	return path[idx+1:]
}
