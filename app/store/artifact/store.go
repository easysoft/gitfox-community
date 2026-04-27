package artifact

import (
	"context"
	"encoding/json"
	"fmt"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"

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

func (s *Store) Groups() GroupInterface            { return s.groups }
func (s *Store) Views() ViewInterface              { return s.views }
func (s *Store) Packages() PackageInterface        { return s.packages }
func (s *Store) Versions() VersionInterface        { return s.versions }
func (s *Store) Assets() AssetInterface            { return s.assets }
func (s *Store) MetaAssets() MetaAssetInterface    { return s.metaAssets }
func (s *Store) Blobs() BlobInterface              { return s.blobs }
func (s *Store) Nodes() TreeNodeInterface          { return s.nodes }

func (s *Store) FindPackages(ctx context.Context, viewId int64, filter *model.ArtifactFilter) ([]*model.ArtifactListItem, error) {
	// 通过子查询先取每个 package 在当前 view 下的最新 version，再回查 package 列表。
	result := make([]*model.ArtifactListItem, 0)
	q := model.ArtifactPackage{ViewID: viewId}

	stmt := s.db.WithContext(ctx).Model(&model.ArtifactPackage{}).
		Select(`ops_artifact_packages.format, ops_artifact_packages.name, ops_artifact_packages.namespace,
			ops_artifact_versions.version, ops_artifact_versions.editedDate`).
		Joins(`JOIN (
			SELECT packageID, MAX(editedDate) AS latest_updated
				FROM ops_artifact_versions
				WHERE viewID = ?
				GROUP BY packageID) AS latest_versions
			ON ops_artifact_packages.id = latest_versions.packageID
				AND ops_artifact_packages.deleted = 0`, viewId).
		Joins(`JOIN ops_artifact_versions
			ON ops_artifact_versions.packageID = latest_versions.packageID
				AND ops_artifact_versions.editedDate = latest_versions.latest_updated
				AND ops_artifact_versions.deleted = 0`)

	if filter != nil {
		if filter.Format != "" {
			q.Format = model.ArtifactFormat(filter.Format)
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

func (s *Store) StatisticAssets(ctx context.Context, options ...SearchOption) (*model.ArtifactStatisticReport, error) {
	assets, err := s.Assets().SearchExtendBlob(ctx, options...)
	if err != nil {
		return nil, err
	}

	report := &model.ArtifactStatisticReport{
		DeleteList: make([]*model.ArtifactRecycleBlobDesc, 0),
		TagList:    make([]*model.ArtifactVersionCapacityDesc, 0),
	}
	versionCap := make(map[int64]*model.ArtifactVersionCapacityDesc)
	blobRefs := make(map[int64]int)

	for _, asset := range assets {
		if asset.BlobID == 0 {
			continue
		}
		blobRefs[asset.BlobID]++
	}

	for _, asset := range assets {
		if asset.BlobID == 0 {
			continue
		}
		if asset.Deleted > 0 {
			report.DeleteList = append(report.DeleteList, &model.ArtifactRecycleBlobDesc{
				AssetId: asset.ID,
				Path:    asset.Path,
				BlobId:  asset.BlobID,
				BlobRef: asset.Ref,
				Size:    asset.Size,
			})
			continue
		}
		if asset.VersionID.Ptr() == nil {
			continue
		}

		versionID := asset.VersionID.Int64
		capItem, ok := versionCap[versionID]
		if !ok {
			capItem = &model.ArtifactVersionCapacityDesc{VersionId: versionID}
			versionCap[versionID] = capItem
		}
		capItem.TotalSize += asset.Size
		capItem.TotalRefs++
		if blobRefs[asset.BlobID] <= 1 {
			capItem.ExclusiveSize += asset.Size
			capItem.ExclusiveRefs++
		}
	}

	for _, item := range versionCap {
		report.TagList = append(report.TagList, item)
	}
	return report, nil
}

func (s *Store) GetVersion(ctx context.Context, viewId int64,
	packageName, groupName, versionName string, format model.ArtifactFormat,
) (*model.ArtifactVersion, error) {
	dbPkg, err := s.packages.GetByName(ctx, packageName, groupName, viewId, format)
	if err != nil {
		return nil, err
	}
	dbVer, err := s.versions.GetByVersion(ctx, dbPkg.ID, viewId, versionName)
	if err != nil {
		return nil, err
	}
	return dbVer, nil
}

func (s *Store) ListAssets(ctx context.Context, versionId int64) ([]*model.ArtifactAssetsRes, error) {
	// ListAssets 拉取 version 下未删除资产，并联表补齐 blob 与创建人信息。
	result := make([]*model.ArtifactAssetsRes, 0)
	stmt := s.db.WithContext(ctx).Model(&model.ArtifactAsset{}).
		Select(`ops_artifact_assets.id,ops_artifact_assets.path,ops_artifact_assets.contentType,
			ops_artifact_assets.checkSum,ops_artifact_assets.metadata,ops_artifact_assets.createdDate,
			ops_artifact_blobs.size,ops_artifact_blobs.ref,ops_artifact_blobs.createdDate,ops_artifact_blobs.createdBy`).
		Joins(`LEFT JOIN ops_artifact_blobs on ops_artifact_assets.blobID=ops_artifact_blobs.id`).
		Where(`ops_artifact_assets.versionID = ?`, versionId).
		Where(`ops_artifact_assets.deleted = 0`)

	if err := stmt.Find(&result).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	parseAssetsRes(ctx, result)
	return result, nil
}

func (s *Store) ListGroupAssets(ctx context.Context, groupId int64) ([]*model.ArtifactAssetsRes, error) {
	// ListGroupAssets 面向 generic/raw group 级资产，查询逻辑与 version 资产分开。
	var dst []*model.ArtifactAssetsRes
	stmt := s.db.WithContext(ctx).Model(&model.ArtifactAsset{}).
		Select(`ops_artifact_assets.id,ops_artifact_assets.path,ops_artifact_assets.contentType,
			ops_artifact_assets.checkSum,ops_artifact_assets.metadata,ops_artifact_assets.createdDate,
			ops_artifact_blobs.size,ops_artifact_blobs.ref,ops_artifact_blobs.createdDate,ops_artifact_blobs.createdBy`).
		Joins(`LEFT JOIN ops_artifact_blobs on ops_artifact_assets.blobID=ops_artifact_blobs.id`).
		Where(`ops_artifact_assets.groupID = ?`, groupId).
		Where(`ops_artifact_assets.deleted = 0`)

	if err := stmt.Find(&dst).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	parseAssetsRes(ctx, dst)
	return dst, nil
}

func (s *Store) GetAsset(ctx context.Context, assetId int64) (*model.ArtifactAssetsRes, error) {
	// GetAsset 读取单个资产详情，并把字符串字段反序列化为结构化响应。
	var result model.ArtifactAssetsRes
	stmt := s.db.WithContext(ctx).Model(&model.ArtifactAsset{}).
		Select(`ops_artifact_assets.id,ops_artifact_assets.path,ops_artifact_assets.contentType,
			ops_artifact_assets.checkSum,ops_artifact_assets.metadata,ops_artifact_assets.createdDate,
			ops_artifact_blobs.size,ops_artifact_blobs.ref,ops_artifact_blobs.createdDate,ops_artifact_blobs.createdBy`).
		Joins(`LEFT JOIN ops_artifact_blobs on ops_artifact_assets.blobID=ops_artifact_blobs.id`).
		Where(`ops_artifact_assets.id = ?`, assetId).
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
