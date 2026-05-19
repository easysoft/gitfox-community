package artifact

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	enum "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"

	"gorm.io/gorm"
)

// garbageCollectPayload 是 GC 异步任务的入参负载。
//   - Entity        : 仅清指定格式（generic/raw/container 等）；为空表示全量。
//   - RetentionTime : 回收站保留期（秒）。当前传 0，使用 service 配置中的默认值。
type garbageCollectPayload struct {
	Entity        model.ArtifactFormat `json:"entity,omitempty"`
	RetentionTime int                  `json:"retention_time"`
}

// =============================================================================
// 触发层（异步）：将 GC 任务投递到 jobs 调度器，handler 在工作线程中调实际执行函数
// =============================================================================

// TriggerGarbageCollect 触发指定格式或全量制品 GC 异步任务。
// entity 留空表示清理全部 format；非空时会先经 normalizeGarbageCollectEntity 校验。
func (s *Service) TriggerGarbageCollect(ctx context.Context, entity model.ArtifactFormat) (*model.JobUIDResponse, error) {
	entity, err := normalizeGarbageCollectEntity(entity)
	if err != nil {
		return nil, err
	}
	return s.triggerGarbageCollect(ctx, JobTypeArtifactGC, entity)
}

// normalizeGarbageCollectEntity 将外部传入的 entity 统一规整为合法 ArtifactFormat。
// 空值放行，非法值包成 i18n 错误抛出，避免落到调度器后才报错。
func normalizeGarbageCollectEntity(entity model.ArtifactFormat) (model.ArtifactFormat, error) {
	raw := strings.TrimSpace(string(entity))
	if raw == "" {
		return "", nil
	}
	format, ok := enum.ArtifactFormat(raw).Sanitize()
	if !ok {
		return "", pkgErrors.WithI18nKey(
			fmt.Errorf("unsupported garbage collect entity: %s", entity),
			i18n.ArtifactViewFormatInvalid,
		)
	}
	return model.ArtifactFormat(format), nil
}

// triggerGarbageCollect 是所有 GC 触发链路的最终落点：
//  1. 校验调度器存在（否则视为配置缺失）
//  2. 以「jobType + 纳秒时间戳」拼出唯一 jobName，避免重名冲突
//  3. 序列化 payload 后投递到 jobs 调度器，handler 在工作线程中调实际执行函数
func (s *Service) triggerGarbageCollect(ctx context.Context, jobType string, entity model.ArtifactFormat) (*model.JobUIDResponse, error) {
	if s.scheduler == nil {
		return nil, pkgErrors.WithI18nKey(
			errors.New("artifact job scheduler is not configured"),
			i18n.ArtifactGcSchedulerNotConfigured,
		)
	}
	jobName := fmt.Sprintf("%s:%d", jobType, time.Now().UnixNano())
	payload, err := json.Marshal(garbageCollectPayload{
		Entity:        entity,
		RetentionTime: 0,
	})
	if err != nil {
		return nil, err
	}

	if err = s.scheduler.RunJob(ctx, jobName, jobType, string(payload), ArtifactGCJobMaxDurationSeconds); err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactGcTriggerError)
	}
	return &model.JobUIDResponse{UID: jobName}, nil
}

// =============================================================================
// 执行层（同步）：cron handler / 同步链路实际执行 GC，落地到 storage
// =============================================================================

// GarbageCollectSoftRemove 清理已软删的非容器格式资源（raw / generic / helm 等）。
// 仅清回收站保留期外的资源；container 格式由 GarbageCollectContainer 单独处理。
func (s *Service) GarbageCollectSoftRemove(ctx context.Context) (*model.ArtifactGarbageReportRes, error) {
	report, err := s.artStore.StatisticAssets(ctx, s.config.RecycleBinRetention, artifactStore.ExcludeContainerOption{})
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactGcExecuteError)
	}
	return s.garbageCollect(ctx, report.DeleteList)
}

// GarbageCollectContainer 清理 container 格式制品的孤立 blob / manifest。
// 容器协议产生的 dangling blob 由此函数处理，与软删通用 GC 互不重叠。
func (s *Service) GarbageCollectContainer(ctx context.Context) (*model.ArtifactGarbageReportRes, error) {
	report, err := s.artStore.StatisticAssets(ctx, s.config.RecycleBinRetention, artifactStore.ContainerFormatOption{})
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactGcExecuteError)
	}
	return s.garbageCollect(ctx, report.DeleteList)
}

// GarbageCollect 执行指定格式或全量软删资源的 GC。
// entity 为空时清理所有 format；否则只清指定格式，复用 garbageCollect 主流程。
func (s *Service) GarbageCollect(ctx context.Context, entity model.ArtifactFormat) (*model.ArtifactGarbageReportRes, error) {
	entity, err := normalizeGarbageCollectEntity(entity)
	if err != nil {
		return nil, err
	}
	options := make([]artifactStore.SearchOption, 0, 1)
	if entity != "" {
		options = append(options, artifactStore.AssetFormatOption{Format: entity})
	}
	report, err := s.artStore.StatisticAssets(ctx, s.config.RecycleBinRetention, options...)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactGcExecuteError)
	}
	return s.garbageCollect(ctx, report.DeleteList)
}

// garbageCollect 是 GC 的核心执行体：
//  1. 单事务硬删 asset；DeleteBlob=true 时同删 blob 记录与物理文件
//  2. 累计被影响的 versionID / groupID，供后续级联清空判定
//  3. 调 collectEmptyDeletedParents 清掉「已软删且无活资源」的空壳 version/package/group
func (s *Service) garbageCollect(ctx context.Context, items []*model.ArtifactRecycleBlobDesc) (*model.ArtifactGarbageReportRes, error) {
	res := &model.ArtifactGarbageReportRes{Ids: make([]int64, 0, len(items))}
	versionIDs := make(map[int64]struct{})
	packageIDs := make(map[int64]struct{})
	groupIDs := make(map[int64]struct{})
	for _, item := range items {
		if item == nil || item.AssetID == 0 || item.BlobID == 0 {
			continue
		}
		if item.VersionID > 0 {
			versionIDs[item.VersionID] = struct{}{}
		}
		if item.GroupID > 0 {
			groupIDs[item.GroupID] = struct{}{}
		}
		if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := s.artStore.Assets().DeleteById(ctx, item.AssetID); err != nil {
				return err
			}
			if item.DeleteBlob {
				if err := s.artStore.Blobs().DeleteById(ctx, item.BlobID); err != nil {
					return err
				}
				if item.BlobRef != "" {
					return s.fileStore.Delete(ctx, adapter.BlobPath(item.BlobRef))
				}
			}
			return nil
		}); err != nil {
			logger.GetLogger(ctx).Errorf("artifact garbage collect failed for asset %d blob %d: %v", item.AssetID, item.BlobID, err)
			return res, err
		}
		res.Count++
		res.Size += item.Size
		res.Ids = append(res.Ids, item.AssetID)
	}
	if err := s.collectEmptyDeletedParents(ctx, versionIDs, packageIDs, groupIDs); err != nil {
		return res, err
	}
	return res, nil
}

// collectEmptyDeletedParents 自底向上收割「已软删且无活子节点」的祖先：
//   - version：若已软删且无活 asset，物理删除并把 packageID 加入待查集合
//   - package：若已软删且无活 version，物理删除并把 groupID 加入待查集合
//   - group  ：交由 collectEmptyDeletedGroups 递归处理
func (s *Service) collectEmptyDeletedParents(ctx context.Context, versionIDs, packageIDs, groupIDs map[int64]struct{}) error {
	for versionID := range versionIDs {
		pkgID, deleted, err := s.versionGCState(ctx, versionID)
		if err != nil {
			return err
		}
		if pkgID > 0 {
			packageIDs[pkgID] = struct{}{}
		}
		if !deleted {
			continue
		}
		assets, err := s.artStore.Assets().Search(ctx, artifactStore.AssetVersionOption{VersionID: versionID})
		if err != nil {
			return err
		}
		if len(assets) == 0 {
			if _, err = s.artStore.Versions().DeleteByIds(ctx, versionID); err != nil {
				return err
			}
		}
	}

	for packageID := range packageIDs {
		groupID, deleted, err := s.packageGCState(ctx, packageID)
		if err != nil {
			return err
		}
		if groupID > 0 {
			groupIDs[groupID] = struct{}{}
		}
		if !deleted {
			continue
		}
		versions, err := s.artStore.Versions().Search(ctx, artifactStore.VersionPackageOption{PackageID: packageID})
		if err != nil {
			return err
		}
		if len(versions) == 0 {
			if _, err = s.artStore.Packages().DeleteByIds(ctx, packageID); err != nil {
				return err
			}
		}
	}

	return s.collectEmptyDeletedGroups(ctx, groupIDs)
}

// versionGCState 返回 version 的 packageID 与软删标记，未找到时静默返回 (0,false,nil)。
func (s *Service) versionGCState(ctx context.Context, versionID int64) (int64, bool, error) {
	versions, err := s.artStore.Versions().Search(ctx, artifactStore.VersionIDsOption{IDs: []int64{versionID}})
	if err != nil {
		return 0, false, err
	}
	if len(versions) == 0 {
		return 0, false, nil
	}
	return versions[0].PackageID, versions[0].Deleted == 1, nil
}

// packageGCState 返回 package 的 groupID 与软删标记，未找到时静默返回 (0,false,nil)。
func (s *Service) packageGCState(ctx context.Context, packageID int64) (int64, bool, error) {
	pkg, err := s.artStore.Packages().GetByID(ctx, packageID)
	if err != nil {
		if errors.Is(err, pkgErrors.ErrResourceNotFound) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return pkg.GroupID, pkg.Deleted == 1, nil
}

// collectEmptyDeletedGroups 迭代清理空的软删 group，并把父 group 拉入下一轮。
// 用 changed 标志循环到稳态：每删一个 group 都重新检查祖先是否变空。
// 跳过条件：未软删 / 非空 / 已不存在。
func (s *Service) collectEmptyDeletedGroups(ctx context.Context, groupIDs map[int64]struct{}) error {
	changed := true
	for changed {
		changed = false
		for groupID := range groupIDs {
			group, err := s.artStore.Groups().GetById(ctx, groupID)
			if err != nil {
				if errors.Is(err, pkgErrors.ErrResourceNotFound) {
					delete(groupIDs, groupID)
					continue
				}
				return err
			}
			parentID := group.ParentID.Int64
			if parentID > 0 {
				groupIDs[parentID] = struct{}{}
			}
			if group.Deleted != 1 {
				delete(groupIDs, groupID)
				continue
			}
			empty, err := s.isGroupEmpty(ctx, groupID)
			if err != nil {
				return err
			}
			if !empty {
				delete(groupIDs, groupID)
				continue
			}
			if err = s.artStore.Groups().DeleteById(ctx, groupID); err != nil {
				return err
			}
			delete(groupIDs, groupID)
			changed = true
		}
	}
	return nil
}

// isGroupEmpty 判断 group 是否已无任何子节点（asset / package / 子 group）。
// 三类逐项查，任一非空立即返回 false，避免无谓的二次查询。
func (s *Service) isGroupEmpty(ctx context.Context, groupID int64) (bool, error) {
	assets, err := s.artStore.Assets().Search(ctx, artifactStore.GroupAssetsOption{GroupID: groupID})
	if err != nil {
		return false, err
	}
	if len(assets) > 0 {
		return false, nil
	}
	packages, err := s.artStore.Packages().Search(ctx, artifactStore.PackageGroupOption{GroupID: groupID})
	if err != nil {
		return false, err
	}
	if len(packages) > 0 {
		return false, nil
	}
	children, err := s.artStore.Groups().ListChildren(ctx, groupID)
	if err != nil {
		return false, err
	}
	return len(children) == 0, nil
}

// deleteFormatTreeNode 删除指定 artifact 下某 format 的根树节点（path="/"）。
// 当该 format 下已无活 group 时由软删链路调用，避免前端看到空的格式入口。
// 节点不存在时静默返回（不报错），保持调用方简洁。
func (s *Service) deleteFormatTreeNode(ctx context.Context, artifactID int64, format model.ArtifactFormat) error {
	node, err := s.artStore.Nodes().GetByPath(ctx, artifactID, "/", format)
	if err != nil {
		return nil
	}
	return s.artStore.Nodes().DeleteById(ctx, node.ID)
}
