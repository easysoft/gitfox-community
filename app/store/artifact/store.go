package artifact

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"

	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

// pickDB 选连接：调用方传入 tx 时复用事务连接，nil 时回落到 store 自身连接。
// 集中收口避免每个 store 方法都写一遍 nil 判断。
func pickDB(tx, fallback *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return fallback
}

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

// ListAssets 查询资产
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

// ListAssetsPage 分页查询资产
func (s *Store) ListAssetsPage(ctx context.Context, req *model.ListAssetReq) (*database.PageResult[*model.ArtifactAssetsRes], error) {
	stmt, err := s.listAssetsPageStmt(ctx, req)
	if err != nil {
		return nil, err
	}

	rows := make([]*model.ArtifactAssetsRes, 0)
	pageResult, err := database.Paginate(ctx, stmt, database.PageParam{
		Page:      req.Page,
		PageSize:  req.PageSize,
		Sort:      assetListSortColumn(req.Sort),
		Order:     assetListSortOrder(req.Order),
		NeedTotal: true,
	}, &rows)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	parseAssetsRes(ctx, rows)

	return &database.PageResult[*model.ArtifactAssetsRes]{
		Data:      rows,
		Total:     pageResult.Total,
		Page:      pageResult.Page,
		PageSize:  pageResult.PageSize,
		TotalPage: pageResult.TotalPage,
	}, nil
}

// ---- 资产列表分页：链路活性 + 入口路由 ----
//
// ---- 资产列表分页：入口活性 + 路由 ----
//
// 设计：业务铁律已保证"上游活则下游全活"——
//   - softRemoveVersion：必先无活 asset（artifact_remove.go:148-166）；
//   - softRemovePackage：必先所有 version 下无活 asset，再级联软删 version（同文件:168-206）；
//   - softRemoveGroup  ：级联软删子树所有 package/version/asset（同文件:254-334）；
//   - softRemoveArtifact：子树皆空才能软删（同文件:338-420）。
// 故 SQL 只需校验"入口实体自身活"+"asset 自身 deleted=0"——下游冗余校验
// 既无必要也会拖慢查询，更会因为级联尚未完成等边界把活节点误算。
//
// 不同入口活性约束：
//   - group.{id}    → 校验 group 自身 deleted=0；下挂 asset 由级联铁律保证活
//   - package.{id}  → 校验 package 自身 deleted=0；下挂 version/asset 同理
//   - version.{id}  → 校验 version 自身 deleted=0；下挂 asset 同理
//   - asset.{id}    → 仅 assets.deleted=0（已在 common filter）
//
// 回收站口径（req.Deleted=true）：跳过实体自身活性校验，但仍 join 上链拿过滤字段。

// isRecycleQuery 判定是否回收站口径。Deleted=true 即仅查软删 asset。
func isRecycleQuery(req *model.ListAssetReq) bool {
	return req.Deleted != nil && *req.Deleted
}

// joinVersionPackage 仅 join versions+packages，不施加活性约束。供 package/version 入口使用。
func joinVersionPackage(stmt *gorm.DB) *gorm.DB {
	return stmt.
		Joins(`JOIN ops_artifact_versions ON ops_artifact_versions.id = ops_artifact_assets.versionID`).
		Joins(`JOIN ops_artifact_packages ON ops_artifact_packages.id = ops_artifact_versions.packageID`)
}

// applyVersionLike 当 req.Version 非空时按版本名模糊匹配。链路必含 versions JOIN。
func applyVersionLike(stmt *gorm.DB, version string) *gorm.DB {
	if version == "" {
		return stmt
	}
	return stmt.Where(`ops_artifact_versions.version LIKE ?`, "%"+version+"%")
}

// liveEntityCond 当非回收站口径时给入口实体加 deleted=0；回收站口径不加。
func liveEntityCond(stmt *gorm.DB, table string, req *model.ListAssetReq) *gorm.DB {
	if isRecycleQuery(req) {
		return stmt
	}
	return stmt.Where(table + `.deleted = 0`)
}

// listAssetsPageStmt 资产列表分页查询：入口活性 + 路由。
func (s *Store) listAssetsPageStmt(ctx context.Context, req *model.ListAssetReq) (*gorm.DB, error) {
	splitEntityID := req.EntityID.GetEntityID()
	stmt := applyAssetListCommonFilters(s.assetListBaseStmt(ctx), req)

	switch splitEntityID.Kind {
	case constant.ArtifactEntityKindGroup:
		return s.listGroupAssetsPageStmt(ctx, stmt, splitEntityID.ID, req)
	case constant.ArtifactEntityKindVersion:
		// version.{id}：锚定 versionID，仅校验该 version 活；下挂 asset 由级联铁律保证活。
		stmt = joinVersionPackage(stmt).Where(`ops_artifact_assets.versionID = ?`, splitEntityID.ID)
		return applyVersionLike(liveEntityCond(stmt, "ops_artifact_versions", req), req.Version), nil
	case constant.ArtifactEntityKindPackage:
		// package.{id}：锚定 packageID，仅校验该 package 活；下挂 version/asset 同理。
		stmt = joinVersionPackage(stmt).Where(`ops_artifact_versions.packageID = ?`, splitEntityID.ID)
		return applyVersionLike(liveEntityCond(stmt, "ops_artifact_packages", req), req.Version), nil
	case constant.ArtifactEntityKindAsset:
		// asset.{id}：仅 assets.deleted 由 common filter 控制；version/package 仅 join 用于 Version 模糊。
		stmt = stmt.
			Joins(`LEFT JOIN ops_artifact_versions ON ops_artifact_versions.id = ops_artifact_assets.versionID`).
			Joins(`LEFT JOIN ops_artifact_packages ON ops_artifact_packages.id = ops_artifact_versions.packageID`).
			Where(`ops_artifact_assets.id = ?`, splitEntityID.ID)
		return applyVersionLike(stmt, req.Version), nil
	default:
		return stmt.Where("1 = 0"), nil
	}
}

// applyAssetListCommonFilters 资产列表公共过滤条件（与入口无关：软删态、文件名、操作人）。
func applyAssetListCommonFilters(stmt *gorm.DB, req *model.ListAssetReq) *gorm.DB {
	if req.Deleted != nil && *req.Deleted {
		stmt = stmt.Where(`ops_artifact_assets.deleted = 1`)
	} else {
		stmt = stmt.Where(`ops_artifact_assets.deleted = 0`)
	}
	if req.Name != "" {
		stmt = stmt.Where(`ops_artifact_assets.basename LIKE ?`, "%"+req.Name+"%")
	}
	if req.CreatedBy != "" {
		stmt = stmt.Where(`ops_artifact_assets.createdBy = ?`, req.CreatedBy)
	}
	if req.EditedBy != "" {
		stmt = stmt.Where(`ops_artifact_assets.editedBy = ?`, req.EditedBy)
	}
	return stmt
}

// assetListBaseStmt 资产列表基础语句：仅 SELECT 字段 + LEFT JOIN blobs 拿大小/引用。
func (s *Store) assetListBaseStmt(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx).Model(&artifactAsset{}).
		Select(`ops_artifact_assets.id,ops_artifact_assets.basename,ops_artifact_assets.contentType,
			ops_artifact_assets.checksum,
			ops_artifact_assets.createdBy,ops_artifact_assets.createdDate,
			ops_artifact_assets.editedBy,ops_artifact_assets.editedDate,
			ops_artifact_blobs.size,ops_artifact_blobs.ref`).
		Joins(`LEFT JOIN ops_artifact_blobs on ops_artifact_assets.blobID=ops_artifact_blobs.id`)
}

// listGroupAssetsPageStmt 处理 group.{id} 入口。
func (s *Store) listGroupAssetsPageStmt(ctx context.Context, stmt *gorm.DB, groupID int64, req *model.ListAssetReq) (*gorm.DB, error) {
	group, err := s.groups.GetById(ctx, groupID)
	if err != nil {
		return nil, err
	}

	groupIDs := []int64{group.ID}
	if req.More {
		descendants, e := s.groups.ListDescendants(ctx, groupID)
		if e != nil {
			return nil, e
		}
		for _, d := range descendants {
			groupIDs = append(groupIDs, d.ID)
		}
	}

	if group.Format == model.ArtifactFileFormat {
		stmt = stmt.
			Joins(`JOIN ops_artifact_groups ON ops_artifact_groups.id = ops_artifact_assets.groupID`).
			Where(`ops_artifact_assets.groupID IN ?`, groupIDs)
		return liveEntityCond(stmt, "ops_artifact_groups", req), nil
	}

	stmt = joinVersionPackage(stmt).
		Joins(`JOIN ops_artifact_groups ON ops_artifact_groups.id = ops_artifact_packages.groupID`).
		Where(`ops_artifact_packages.groupID IN ?`, groupIDs).
		Where(`ops_artifact_packages.type <> ?`, model.ArtifactContainerFormat)
	return applyVersionLike(liveEntityCond(stmt, "ops_artifact_groups", req), req.Version), nil
}

// assetListSortOrder 资产列表排序方向 默认按 升序排序
func assetListSortOrder(order string) string {
	switch strings.ToLower(order) {
	case "asc":
		return "asc"
	case "desc":
		return "desc"
	default:
		return "asc"
	}
}

// assetListSortColumn 资产列表排序字段 默认按 editedDate 排序
func assetListSortColumn(sortField string) string {
	switch sortField {
	case "":
		return "ops_artifact_assets.editedDate"
	case "id":
		return "ops_artifact_assets.id"
	case "name":
		return "ops_artifact_assets.basename"
	case "size":
		return "ops_artifact_blobs.size"
	case "createdDate":
		return "ops_artifact_assets.createdDate"
	case "editedDate":
		return "ops_artifact_assets.editedDate"
	default:
		return "ops_artifact_assets.editedDate"
	}
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
