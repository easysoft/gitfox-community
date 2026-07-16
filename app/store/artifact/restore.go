package artifact

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

// =============================================================================
// 恢复（Restore）：与 SoftDelete 完全镜像，仅更新 deleted/editedBy 两列。
// 风格统一：列定向 Updates(map) + database.ProcessSQLErrorf 错误包装。
// 所有 Restore 都不级联——级联策略由 service 层决定。
// =============================================================================

// Restore 恢复制品库视图（将 deleted 置为 0），并把 operator 写入 editedBy（即恢复人）。
func (c *libs) Restore(ctx context.Context, viewID int64, operator string) error {
	updates := map[string]interface{}{"deleted": 0}
	if operator != "" {
		updates["editedBy"] = operator
	}
	if err := c.db.WithContext(ctx).Model(&artifactView{ID: viewID}).
		Updates(updates).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// RestoreById 恢复分组（将 deleted 置为 0）并记录恢复人到 editedBy。
func (c *groups) RestoreById(ctx context.Context, groupID int64, operator string) error {
	updates := map[string]interface{}{"deleted": 0}
	if operator != "" {
		updates["editedBy"] = operator
	}
	if err := c.db.WithContext(ctx).Model(&artifactGroup{ID: groupID}).
		Updates(updates).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// Restore 恢复制品包（将 deleted 置为 0）并记录恢复人到 editedBy。
func (c *packages) Restore(ctx context.Context, packageID int64, operator string) error {
	updates := map[string]interface{}{"deleted": 0}
	if operator != "" {
		updates["editedBy"] = operator
	}
	if err := c.db.WithContext(ctx).Model(&artifactPackage{ID: packageID}).
		Updates(updates).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// Restore 恢复制品版本（将 deleted 置为 0）并记录恢复人到 editedBy。
func (c *versions) Restore(ctx context.Context, versionID int64, operator string) error {
	updates := map[string]interface{}{"deleted": 0}
	if operator != "" {
		updates["editedBy"] = operator
	}
	if err := c.db.WithContext(ctx).Model(&artifactVersion{ID: versionID}).
		Updates(updates).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// RestoreById 恢复制品资产（将 deleted 置为 0）并记录恢复人到 editedBy。
func (c *assets) RestoreById(ctx context.Context, assetID int64, operator string) error {
	updates := map[string]interface{}{"deleted": 0}
	if operator != "" {
		updates["editedBy"] = operator
	}
	if err := c.db.WithContext(ctx).Model(&ArtifactAsset{ID: assetID}).
		Updates(updates).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// RestoreById 恢复制品 blob（将 deleted 置为 0）并记录恢复人到 editedBy。
// 不论引用计数，blob.deleted=0 不影响 GC——GC 看的是 active asset 引用数。
func (c *blobs) RestoreById(ctx context.Context, blobID int64, operator string) error {
	updates := map[string]interface{}{"deleted": 0}
	if operator != "" {
		updates["editedBy"] = operator
	}
	if err := c.db.WithContext(ctx).Model(&artifactBlob{ID: blobID}).
		Updates(updates).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}
