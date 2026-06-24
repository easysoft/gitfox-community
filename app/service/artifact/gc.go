package artifact

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter"
	containerGC "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter/container/gc"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"

	"gorm.io/gorm"
)

// garbageCollectPayload 是 GC 异步任务的入参负载。
//   - Format        : 仅清指定格式（generic/raw/container 等）；空表示全量。
//   - RetentionTime : 回收站保留期（秒）。当前传 0，使用 service 配置中的默认值。
type garbageCollectPayload struct {
	Format        string `json:"format,omitempty"`
	RetentionTime int    `json:"retention_time"`
}

// =============================================================================
// 触发层（异步）：将 GC 任务投递到 jobs 调度器，handler 在工作线程中调实际执行函数
// =============================================================================

// TriggerGarbageCollect 触发指定格式或全量制品 GC 异步任务。
func (s *Service) TriggerGarbageCollect(ctx context.Context, format string) (*model.JobUIDResponse, error) {
	if s.scheduler == nil {
		return nil, pkgErrors.WithI18nKey(errors.New("artifact job scheduler is not configured"), i18n.ArtifactGcSchedulerNotConfigured)
	}

	jobName := fmt.Sprintf("%s:%d", JobTypeArtifactGC, time.Now().UnixNano())
	payload, err := json.Marshal(garbageCollectPayload{
		Format:        format,
		RetentionTime: 0,
	})
	if err != nil {
		return nil, err
	}

	if err = s.scheduler.RunJob(ctx, jobName, JobTypeArtifactGC, string(payload), ArtifactGCJobMaxDurationSeconds); err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactGcTriggerError)
	}
	return &model.JobUIDResponse{UID: jobName}, nil
}

// =============================================================================
// 执行层（同步）：cron handler / 同步链路实际执行 GC，落地到 storage
// =============================================================================

// GarbageCollect 执行指定格式或全量资源 GC（cron 自动触发 + 用户主动 API 共用）。
// container 需要按 manifest 引用图判断可回收 blob；非 container 保持软删保留期规则。
func (s *Service) GarbageCollect(ctx context.Context, format string) (*model.ArtifactGarbageReportRes, error) {
	// 特殊处理 container 格式，按 manifest 引用图判断可回收 blob。
	if format == string(model.ArtifactContainerFormat) {
		// TODO: 锁保护，避免并发 GC 任务同时上传 container 图资源
		return s.garbageCollectContainer(ctx)
	}

	// 普通格式 GC，按 assetFormatOption 过滤。
	if format != "" {
		return s.garbageCollectByOptions(ctx, artifactStore.AssetFormatOption{Format: format})
	}

	// 全量 GC 分两段：普通格式清理过期软删资源，container 清理 manifest 图不可达资源。
	nonContainer, err := s.garbageCollectByOptions(ctx, artifactStore.ExcludeContainerOption{})
	if err != nil {
		return nonContainer, err
	}

	// TODO: 锁保护，避免并发 GC 任务同时上传 container 图资源
	container, err := s.garbageCollectContainer(ctx)
	if err != nil {
		return mergeGarbageReports(nonContainer, container), err
	}

	// 收口：清理「软删 + 过期 + 空」的 group。
	// asset 反推链路只能命中"被软删 asset 牵出的 group"，对于直接软删
	// 整棵 group 子树（softRemoveGroup）的场景，需要这里主动扫表收割。
	if err = s.garbageCollectExpiredGroups(ctx); err != nil {
		return mergeGarbageReports(nonContainer, container), err
	}

	return mergeGarbageReports(nonContainer, container), nil
}

// garbageCollectExpiredGroups 主动扫描并物理清理「软删超过 RecycleBinRetention 且空」的 group。
// 每轮事务内 Search + isGroupEmpty 判定 + DeleteById；删一个就重启下一轮，
// 直到没有可删 group 为止——保证父 group 在子 group 清空后能在后续轮次被收割。
// 与 collectEmptyDeletedGroups 的差异：此处显式带 cutoff 过滤，避免误清保留期内的软删 group。
func (s *Service) garbageCollectExpiredGroups(ctx context.Context) error {
	log := logger.GetLogger(ctx)
	cutoff := datetime.NowLocalTime().Add(-s.config.RecycleBinRetention).Time()
	log.Debugf("garbageCollectExpiredGroups cutoff: %v", cutoff)

	for {
		var deletedThisRound int
		if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			txStore := artifactStore.NewStore(tx)
			groups, err := txStore.Groups().Search(ctx, artifactStore.GroupWithDeletedBeforeOption{Before: cutoff})
			if err != nil {
				return err
			}
			for _, g := range groups {
				empty, err := s.isGroupEmpty(ctx, txStore, g.ID)
				if err != nil {
					return err
				}
				if !empty {
					continue
				}
				if err = txStore.Groups().DeleteById(ctx, g.ID); err != nil {
					if errors.Is(err, pkgErrors.ErrResourceNotFound) {
						continue
					}
					return err
				}
				deletedThisRound++
			}
			return nil
		}); err != nil {
			log.Errorf("artifact garbage collect expired groups failed: %v", err)
			return err
		}
		if deletedThisRound == 0 {
			return nil
		}
		log.Debugf("artifact garbage collect expired groups: %d in this round", deletedThisRound)
	}
}

// garbageCollectByOptions 执行指定格式资源 GC。
// 1. 调用 store.StatisticAssets 统计资源
// 2. 调用 garbageCollect 执行硬删
func (s *Service) garbageCollectByOptions(ctx context.Context, options ...artifactStore.SearchOption) (*model.ArtifactGarbageReportRes, error) {
	report, err := s.StatisticAssets(ctx, s.config.RecycleBinRetention, options...)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactGcExecuteError)
	}
	return s.garbageCollect(ctx, report.DeleteList)
}

// garbageCollectContainer 执行 container 格式 GC。
func (s *Service) garbageCollectContainer(ctx context.Context) (*model.ArtifactGarbageReportRes, error) {
	report, err := containerGC.Statistic(ctx, s.artStore, s.fileStore, s.config.RecycleBinRetention)
	if err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactGcExecuteError)
	}
	return s.garbageCollect(ctx, report.DeleteList)
}

// mergeGarbageReports 合并多个 GC 报告。
func mergeGarbageReports(reports ...*model.ArtifactGarbageReportRes) *model.ArtifactGarbageReportRes {
	merged := &model.ArtifactGarbageReportRes{Ids: make([]int64, 0)}
	for _, report := range reports {
		if report == nil {
			continue
		}
		merged.Count += report.Count
		merged.Size += report.Size
		merged.Ids = append(merged.Ids, report.Ids...)
	}
	return merged
}

// garbageCollect 是 GC 的核心执行体：
//  1. 单事务硬删 asset；DeleteBlob=true 时同删 blob 记录
//  2. 累计被影响的 versionID / groupID，供后续级联清空判定
//  3. 调 collectEmptyDeletedParents 清掉「已软删且无活资源」的空壳 version/package/group
//  4. DB 事务提交后再删除物理文件，避免事务回滚时文件已不可恢复
func (s *Service) garbageCollect(ctx context.Context, items []*model.ArtifactRecycleBlobDesc) (*model.ArtifactGarbageReportRes, error) {
	res := &model.ArtifactGarbageReportRes{Ids: make([]int64, 0, len(items))}
	versionIDs := make(map[int64]struct{})
	packageIDs := make(map[int64]struct{})
	groupIDs := make(map[int64]struct{})
	txRes := &model.ArtifactGarbageReportRes{Ids: make([]int64, 0, len(items))}
	txBlobRefs := make([]string, 0, len(items))

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txStore := artifactStore.NewStore(tx)
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
			if err := txStore.Assets().DeleteById(ctx, item.AssetID); err != nil && err != model.ErrAssetNoItemDeleted {
				return err
			}
			if item.DeleteBlob {
				if err := txStore.Blobs().DeleteById(ctx, item.BlobID); err != nil && err != model.ErrBlobNoItemDeleted {
					return err
				}
			}
			txRes.Count++
			if item.DeleteBlob {
				txRes.Size += item.Size
				if item.BlobRef != "" {
					txBlobRefs = append(txBlobRefs, item.BlobRef)
				}
			}
			txRes.Ids = append(txRes.Ids, item.AssetID)
		}

		return s.collectEmptyDeletedParents(ctx, txStore, versionIDs, packageIDs, groupIDs)
	}); err != nil {
		logger.GetLogger(ctx).Errorf("artifact garbage collect failed: %v", err)
		return res, err
	}

	res = txRes
	for _, blobRef := range txBlobRefs {
		if err := s.fileStore.Delete(ctx, adapter.BlobPath(blobRef)); err != nil {
			return res, err
		}
	}
	return res, nil
}

// collectEmptyDeletedParents 自底向上收割「已软删且无活子节点」的祖先：
//   - version：若已软删且无活 asset，物理删除并把 packageID 加入待查集合
//   - package：若已软删且无活 version，物理删除并把 groupID 加入待查集合
//   - group  ：交由 collectEmptyDeletedGroups 递归处理
func (s *Service) collectEmptyDeletedParents(ctx context.Context, store artifactStore.Interface, versionIDs, packageIDs, groupIDs map[int64]struct{}) error {
	for versionID := range versionIDs {
		pkgID, deleted, err := s.versionGCState(ctx, store, versionID)
		if err != nil {
			return err
		}
		if pkgID > 0 {
			packageIDs[pkgID] = struct{}{}
		}
		if !deleted {
			continue
		}
		assets, err := store.Assets().Search(ctx, artifactStore.AssetVersionOption{VersionID: versionID})
		if err != nil {
			return err
		}
		if len(assets) == 0 {
			if _, err = store.Versions().DeleteByIds(ctx, versionID); err != nil {
				return err
			}
		}
	}

	for packageID := range packageIDs {
		groupID, deleted, err := s.packageGCState(ctx, store, packageID)
		if err != nil {
			return err
		}
		if groupID > 0 {
			groupIDs[groupID] = struct{}{}
		}
		if !deleted {
			continue
		}
		versions, err := store.Versions().Search(ctx, artifactStore.VersionPackageOption{PackageID: packageID})
		if err != nil {
			return err
		}
		if len(versions) == 0 {
			if _, err = store.Packages().DeleteByIds(ctx, packageID); err != nil {
				return err
			}
		}
	}

	return s.collectEmptyDeletedGroups(ctx, store, groupIDs)
}

// versionGCState 返回 version 的 packageID 与软删标记，未找到时静默返回 (0,false,nil)。
func (s *Service) versionGCState(ctx context.Context, store artifactStore.Interface, versionID int64) (int64, bool, error) {
	versions, err := store.Versions().Search(ctx, artifactStore.VersionIDsOption{IDs: []int64{versionID}})
	if err != nil {
		return 0, false, err
	}
	if len(versions) == 0 {
		return 0, false, nil
	}
	return versions[0].PackageID, versions[0].Deleted == 1, nil
}

// packageGCState 返回 package 的 groupID 与软删标记，未找到时静默返回 (0,false,nil)。
func (s *Service) packageGCState(ctx context.Context, store artifactStore.Interface, packageID int64) (int64, bool, error) {
	pkg, err := store.Packages().GetByID(ctx, packageID)
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
func (s *Service) collectEmptyDeletedGroups(ctx context.Context, store artifactStore.Interface, groupIDs map[int64]struct{}) error {
	changed := true
	for changed {
		changed = false
		for groupID := range groupIDs {
			group, err := store.Groups().GetById(ctx, groupID)
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
			empty, err := s.isGroupEmpty(ctx, store, groupID)
			if err != nil {
				return err
			}
			if !empty {
				delete(groupIDs, groupID)
				continue
			}
			if err = store.Groups().DeleteById(ctx, groupID); err != nil {
				return err
			}
			delete(groupIDs, groupID)
			changed = true
		}
	}
	return nil
}

// isGroupEmpty 判断 group 是否已无任何子节点（asset / package / 子 group）。
func (s *Service) isGroupEmpty(ctx context.Context, store artifactStore.Interface, groupID int64) (bool, error) {
	assets, err := store.Assets().Search(ctx, artifactStore.GroupAssetsOption{GroupID: groupID})
	if err != nil {
		return false, err
	}
	if len(assets) > 0 {
		return false, nil
	}
	packages, err := store.Packages().Search(ctx, artifactStore.PackageGroupOption{GroupID: groupID})
	if err != nil {
		return false, err
	}
	if len(packages) > 0 {
		return false, nil
	}
	children, err := store.Groups().ListChildren(ctx, groupID)
	if err != nil {
		return false, err
	}
	return len(children) == 0, nil
}

// StatisticAssets 统计制品资源容量，同时计算符合 GC 条件的软删 asset 列表。
//
// retentionTime 指定回收站保留时长：只有软删时间（editedDate）早于 now-retentionTime
// 的 asset 才会被列入 DeleteList。传入 0 则跳过时间过滤（不建议用于生产 GC）。
//
// 引用计数保护：blobRefs 只统计 active（deleted=0）asset 对 blob 的引用。
// 若一个软删 asset 的 blob 仍被其他 active asset 引用，则该 blob 不会被物理删除，
// 只删除 asset 记录本身（BlobRef 置空以跳过文件删除）。
func (s *Service) StatisticAssets(ctx context.Context, retentionTime time.Duration, options ...artifactStore.SearchOption) (*model.ArtifactStatisticReport, error) {
	log := logger.GetLogger(ctx)
	assetBlobs, err := s.artStore.Assets().SearchExtendBlob(ctx, options...)
	if err != nil {
		return nil, err
	}

	report := &model.ArtifactStatisticReport{
		DeleteList: make([]*model.ArtifactRecycleBlobDesc, 0),
		TagList:    make([]*model.ArtifactVersionCapacityDesc, 0),
	}
	versionCap := make(map[int64]*model.ArtifactVersionCapacityDesc)

	// 计算回收站保留期截止时间；retentionTime==0 时不做时间过滤
	retentionCutoff := datetime.NowLocalTime().Add(-retentionTime)
	log.Debugf("retentionTime: %v, retentionCutoff: %v", retentionTime, retentionCutoff)

	// 第一遍：只统计 active asset（deleted=0）对 blob 的引用计数。
	// 此处刻意排除 deleted>0 的 asset，确保引用计数反映的是「仍在使用」的 blob。
	activeBlobRefs := make(map[int64]int)
	for _, asset := range assetBlobs {
		if asset.BlobID == 0 {
			continue
		}

		if asset.Deleted > 0 && asset.EditedDate.Before(*retentionCutoff) {
			continue
		}

		activeBlobRefs[asset.BlobID]++
	}

	// 第二遍：分类处理每个 asset
	for _, asset := range assetBlobs {
		if asset.BlobID == 0 {
			continue
		}
		// 处理已删除 asset
		if asset.Deleted > 0 {
			log.Debugf("asset %d is deleted, editedDate: %v", asset.ID, asset.EditedDate)
			if asset.EditedDate == nil || !asset.EditedDate.Before(*retentionCutoff) {
				continue
			}

			blobRef := asset.Ref
			deleteBlob := true
			if activeBlobRefs[asset.BlobID] > 0 {
				log.Debugf("skip blob deletion for asset %d: blob %d still referenced by %d active asset(s)", asset.ID, asset.BlobID, activeBlobRefs[asset.BlobID])
				blobRef = "" // 置空 BlobRef，garbageCollect 会跳过物理文件删除
				deleteBlob = false
			}

			report.DeleteList = append(report.DeleteList, &model.ArtifactRecycleBlobDesc{
				AssetID:    asset.ID,
				VersionID:  asset.VersionID.Int64,
				GroupID:    asset.GroupID.Int64,
				ArtifactID: asset.ArtifactID.Int64,
				Path:       asset.Path,
				BlobID:     asset.BlobID,
				BlobRef:    blobRef,
				DeleteBlob: deleteBlob,
				Size:       asset.Size,
			})
			continue
		}
		if asset.VersionID.Ptr() == nil {
			continue
		}

		verId := asset.VersionID.Int64
		verDesc, ok := versionCap[verId]
		if !ok {
			versionCap[verId] = &model.ArtifactVersionCapacityDesc{
				VersionID:     verId,
				ExclusiveSize: asset.Size,
				TotalSize:     asset.Size,
			}
			continue
		}

		verDesc.ExclusiveSize += asset.Size
		verDesc.TotalSize += asset.Size
	}

	for _, item := range versionCap {
		report.TagList = append(report.TagList, item)
	}
	return report, nil
}
