package artifact

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"

	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

// Store 是 artifact store 的聚合实现。
// Store is the main artifact store implementation.
type Store struct {
	db         *gorm.DB
	groups     *groups
	views      *views
	packages   *packages
	versions   *versions
	assets     *assets
	metaAssets *metaAssets
	blobs      *blobs
	nodes      *treeNode
}

// NewStore 根据数据库连接初始化各类 artifact 子仓储。
// NewStore creates a new artifact store.
func NewStore(db *gorm.DB) *Store {
	return &Store{
		db:         db,
		groups:     &groups{db: db},
		views:      &views{db: db},
		packages:   &packages{db: db},
		versions:   &versions{db: db},
		assets:     &assets{db: db},
		metaAssets: &metaAssets{db: db},
		blobs:      &blobs{db: db},
		nodes:      &treeNode{db: db},
	}
}

func (s *Store) Groups() GroupInterface         { return s.groups }
func (s *Store) Views() ViewInterface           { return s.views }
func (s *Store) Packages() PackageInterface     { return s.packages }
func (s *Store) Versions() VersionInterface     { return s.versions }
func (s *Store) Assets() AssetInterface         { return s.assets }
func (s *Store) MetaAssets() MetaAssetInterface { return s.metaAssets }
func (s *Store) Blobs() BlobInterface           { return s.blobs }
func (s *Store) Nodes() TreeNodeInterface       { return s.nodes }

func (s *Store) FindPackages(ctx context.Context, artifactID int64, filter *model.ArtifactFilter) ([]*model.ArtifactListItem, error) {
	// 通过子查询先取每个 package 在当前 view 下的最新 version，再回查 package 列表。
	result := make([]*model.ArtifactListItem, 0)
	q := artifactPackage{ArtifactID: artifactID}

	stmt := s.db.WithContext(ctx).Model(&artifactPackage{}).
		Select(`ops_artifact_packages.type, ops_artifact_packages.name, ops_artifact_packages.namespace,
			ops_artifact_versions.version, ops_artifact_versions.editedDate`).
		Joins(`JOIN (
			SELECT v.packageID, MAX(v.editedDate) AS latest_updated
				FROM ops_artifact_versions v
				JOIN ops_artifact_packages p ON v.packageID = p.id
				WHERE p.artifactLibID = ? AND v.deleted = 0
				GROUP BY v.packageID) AS latest_versions
			ON ops_artifact_packages.id = latest_versions.packageID
				AND ops_artifact_packages.deleted = 0`, artifactID).
		Joins(`JOIN ops_artifact_versions
			ON ops_artifact_versions.packageID = latest_versions.packageID
				AND ops_artifact_versions.editedDate = latest_versions.latest_updated
				AND ops_artifact_versions.deleted = 0`)

	if filter != nil {
		if filter.Format != "" {
			q.Format = string(filter.Format)
		}
		if filter.Query != "" {
			stmt = stmt.Where("name LIKE ?", fmt.Sprintf("%%%s%%", filter.Query))
		}
	}

	stmt.Where(&q).Order("latest_versions.latest_updated desc").
		Limit(int(database.Limit(filter.Size))).
		Offset(int(database.Offset(filter.Page, filter.Size))).
		Find(&result)

	for id, p := range result {
		result[id].DisplayName = p.Name
		if p.Namespace != "" {
			result[id].DisplayName = p.Namespace + ":" + p.Name
		}
	}

	return result, nil
}

// StatisticAssets 统计制品资源容量，同时计算符合 GC 条件的软删 asset 列表。
//
// retentionTime 指定回收站保留时长：只有软删时间（editedDate）早于 now-retentionTime
// 的 asset 才会被列入 DeleteList。传入 0 则跳过时间过滤（不建议用于生产 GC）。
//
// 引用计数保护：blobRefs 只统计 active（deleted=0）asset 对 blob 的引用。
// 若一个软删 asset 的 blob 仍被其他 active asset 引用，则该 blob 不会被物理删除，
// 只删除 asset 记录本身（BlobRef 置空以跳过文件删除）。
func (s *Store) StatisticAssets(ctx context.Context, retentionTime time.Duration, options ...SearchOption) (*model.ArtifactStatisticReport, error) {
	assets, err := s.Assets().SearchExtendBlob(ctx, options...)
	if err != nil {
		return nil, err
	}

	report := &model.ArtifactStatisticReport{
		DeleteList: make([]*model.ArtifactRecycleBlobDesc, 0),
		TagList:    make([]*model.ArtifactVersionCapacityDesc, 0),
	}
	versionCap := make(map[int64]*model.ArtifactVersionCapacityDesc)

	// 第一遍：只统计 active asset（deleted=0）对 blob 的引用计数。
	// 此处刻意排除 deleted>0 的 asset，确保引用计数反映的是「仍在使用」的 blob。
	activeBlobRefs := make(map[int64]int)
	for _, asset := range assets {
		if asset.BlobID == 0 || asset.Deleted > 0 {
			continue
		}
		activeBlobRefs[asset.BlobID]++
	}

	// 计算回收站保留期截止时间；retentionTime==0 时不做时间过滤
	var retentionCutoff *datetime.Local
	if retentionTime > 0 {
		retentionCutoff = datetime.NowLocalTime().Add(-retentionTime)
	}

	// 第二遍：分类处理每个 asset
	for _, asset := range assets {
		if asset.BlobID == 0 {
			continue
		}
		if asset.Deleted > 0 {
			// 保留期检查：editedDate（软删时间）必须早于截止时间
			if retentionTime > 0 {
				if asset.EditedDate == nil || !asset.EditedDate.Before(*retentionCutoff) {
					// 仍在保留期内，跳过
					continue
				}
			}

			// 引用计数检查：blob 是否仍被 active asset 引用
			// 若有 active 引用则保留物理文件（BlobRef 置空），只删 asset 记录
			blobRef := asset.Ref
			deleteBlob := true
			if activeBlobRefs[asset.BlobID] > 0 {
				logger.GetLogger(ctx).Debugf(
					"skip blob deletion for asset %d: blob %d still referenced by %d active asset(s)",
					asset.ID, asset.BlobID, activeBlobRefs[asset.BlobID],
				)
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

		versionID := asset.VersionID.Int64
		capItem, ok := versionCap[versionID]
		if !ok {
			capItem = &model.ArtifactVersionCapacityDesc{VersionID: versionID}
			versionCap[versionID] = capItem
		}
		capItem.TotalSize += asset.Size
		capItem.TotalRefs++
		if activeBlobRefs[asset.BlobID] <= 1 {
			capItem.ExclusiveSize += asset.Size
			capItem.ExclusiveRefs++
		}
	}

	for _, item := range versionCap {
		report.TagList = append(report.TagList, item)
	}
	return report, nil
}

func (s *Store) GetVersion(ctx context.Context, artifactID int64,
	packageName, groupName, versionName string, format model.ArtifactFormat,
) (*model.ArtifactVersion, error) {
	dbPkg, err := s.packages.GetByName(ctx, packageName, groupName, artifactID, format)
	if err != nil {
		return nil, err
	}
	dbVer, err := s.versions.GetByVersion(ctx, dbPkg.ID, versionName)
	if err != nil {
		return nil, err
	}
	return dbVer, nil
}

func (s *Store) ListAssets(ctx context.Context, VersionID int64) ([]*model.ArtifactAssetsRes, error) {
	// ListAssets 拉取 version 下未删除资产，并联表补齐 blob 大小与引用信息。
	result := make([]*model.ArtifactAssetsRes, 0)
	stmt := s.db.WithContext(ctx).Model(&artifactAsset{}).
		Select(`ops_artifact_assets.id,ops_artifact_assets.basename,ops_artifact_assets.contentType,
			ops_artifact_assets.checksum,
			ops_artifact_assets.createdBy,ops_artifact_assets.createdDate,
			ops_artifact_assets.editedBy,ops_artifact_assets.editedDate,
			ops_artifact_blobs.size,ops_artifact_blobs.ref`).
		Joins(`LEFT JOIN ops_artifact_blobs on ops_artifact_assets.blobID=ops_artifact_blobs.id`).
		Where(`ops_artifact_assets.versionID = ?`, VersionID).
		Where(`ops_artifact_assets.deleted = 0`)

	if err := stmt.Find(&result).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	parseAssetsRes(ctx, result)
	return result, nil
}

func (s *Store) ListGroupAssets(ctx context.Context, groupID int64) ([]*model.ArtifactAssetsRes, error) {
	// ListGroupAssets 面向 generic/raw group 级资产，查询逻辑与 version 资产分开。
	var dst []*model.ArtifactAssetsRes
	stmt := s.db.WithContext(ctx).Model(&artifactAsset{}).
		Select(`ops_artifact_assets.id,ops_artifact_assets.basename,ops_artifact_assets.contentType,
			ops_artifact_assets.checksum,
			ops_artifact_assets.createdBy,ops_artifact_assets.createdDate,
			ops_artifact_assets.editedBy,ops_artifact_assets.editedDate,
			ops_artifact_blobs.size,ops_artifact_blobs.ref`).
		Joins(`LEFT JOIN ops_artifact_blobs on ops_artifact_assets.blobID=ops_artifact_blobs.id`).
		Where(`ops_artifact_assets.groupID = ?`, groupID).
		Where(`ops_artifact_assets.deleted = 0`)

	if err := stmt.Find(&dst).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	parseAssetsRes(ctx, dst)
	return dst, nil
}

func (s *Store) GetAsset(ctx context.Context, assetID int64) (*model.ArtifactAssetsRes, error) {
	// GetAsset 读取单个资产详情，并把字符串字段反序列化为结构化响应。
	var result model.ArtifactAssetsRes
	stmt := s.db.WithContext(ctx).Model(&artifactAsset{}).
		Select(`ops_artifact_assets.id,ops_artifact_assets.basename,ops_artifact_assets.contentType,
			ops_artifact_assets.checksum,
			ops_artifact_assets.createdBy,ops_artifact_assets.createdDate,
			ops_artifact_assets.editedBy,ops_artifact_assets.editedDate,
			ops_artifact_blobs.size,ops_artifact_blobs.ref`).
		Joins(`LEFT JOIN ops_artifact_blobs on ops_artifact_assets.blobID=ops_artifact_blobs.id`).
		Where(`ops_artifact_assets.id = ?`, assetID).
		Where(`ops_artifact_assets.deleted = 0`)

	if err := stmt.Take(&result).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	var checkSum model.CheckSumRes
	if err := json.Unmarshal([]byte(result.CheckSumString), &checkSum); err == nil {
		result.CheckSum = &checkSum
	}
	if result.MetadataString != "" {
		var metadata map[string]string
		if err := json.Unmarshal([]byte(result.MetadataString), &metadata); err == nil {
			result.Metadata = metadata
		}
	}

	return &result, nil
}

// parseAssetsRes 会把字符串存储的 checksum 与 metadata 解码成响应结构。
// parseAssetsRes parses checksum and metadata from string fields.
func parseAssetsRes(ctx context.Context, items []*model.ArtifactAssetsRes) {
	for _, a := range items {
		var checkSum model.CheckSumRes
		if err := json.Unmarshal([]byte(a.CheckSumString), &checkSum); err != nil {
			logger.GetLogger(ctx).Errorf("parse artifact asset checksum failed: %v", err)
		}
		a.CheckSum = &checkSum

		if a.MetadataString != "" {
			var metadata map[string]string
			if err := json.Unmarshal([]byte(a.MetadataString), &metadata); err != nil {
				logger.GetLogger(ctx).Errorf("parse artifact asset metadata failed: %v", err)
			}
			a.Metadata = metadata
		}

		a.Creator = a.CreatorName
	}
}
