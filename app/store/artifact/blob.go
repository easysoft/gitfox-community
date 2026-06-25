package artifact

import (
	"context"
	"errors"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"

	"gorm.io/gorm"
)

var _ BlobInterface = (*blobs)(nil)

type blobs struct{ db *gorm.DB }

func (c *blobs) Create(ctx context.Context, newObj *model.CreateArtifactBlob) (*model.ArtifactBlob, error) {
	if err := validatorEmpty(newObj.StorageID, newObj.Ref); err != nil {
		return nil, err
	}
	var row artifactBlob
	row.StorageID = newObj.StorageID
	row.Ref = newObj.Ref
	row.Size = newObj.Size
	row.CreatedBy = newObj.CreatedBy
	row.CheckSum = newObj.CheckSum
	row.Metadata = newObj.Metadata
	if err := c.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return blobToModel(&row), nil
}

func (c *blobs) GetById(ctx context.Context, blobID int64) (*model.ArtifactBlob, error) {
	var row artifactBlob
	if err := c.db.WithContext(ctx).First(&row, blobID).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return blobToModel(&row), nil
}

func (c *blobs) GetByRef(ctx context.Context, ref string, storageID int64) (*model.ArtifactBlob, error) {
	if err := validatorEmpty(ref, storageID); err != nil {
		return nil, model.ErrArgsValueEmpty
	}
	var row artifactBlob
	q := artifactBlob{Ref: ref, StorageID: storageID}
	if err := c.db.WithContext(ctx).Where(&q).First(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return blobToModel(&row), nil
}

// GetActiveByChecksum 按完整 checksum JSON 串等值匹配未软删 blob，用作上传去重钩子。
// 多机/多副本上传同 sha256 在历史实现中可能落多行，取 id 最小的一条作为复用目标，保证幂等。
func (c *blobs) GetActiveByChecksum(ctx context.Context, checksum string, storageID int64) (*model.ArtifactBlob, error) {
	if checksum == "" || storageID == 0 {
		return nil, model.ErrArgsValueEmpty
	}
	var row artifactBlob
	err := c.db.WithContext(ctx).
		Where("storageID = ?", storageID).
		Where("deleted = 0").
		Where("checksum = ?", checksum).
		Order("id asc").
		First(&row).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return blobToModel(&row), nil
}

func (c *blobs) Update(ctx context.Context, upObj *model.ArtifactBlob) error {
	row := blobFromModel(upObj)
	if err := c.db.WithContext(ctx).Save(row).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *blobs) UpdateOptLock(ctx context.Context, upObj *model.ArtifactBlob, mutateFn func(blob *model.ArtifactBlob) error) (*model.ArtifactBlob, error) {
	for {
		dup := *upObj
		if err := mutateFn(&dup); err != nil {
			return nil, err
		}
		err := c.Update(ctx, &dup)
		if err == nil {
			return &dup, nil
		}
		if !errors.Is(err, pkgErrors.ErrDuplicate) {
			return nil, err
		}
		upObj, err = c.GetById(ctx, upObj.ID)
		if err != nil {
			return nil, err
		}
	}
}

// SoftDeleteById 软删 blob 并记录操作人到 editedBy（回收站 / GC 链路据此识别"删除人"）。
func (c *blobs) SoftDeleteById(ctx context.Context, blobID int64, operator string) error {
	updates := map[string]interface{}{"deleted": 1}
	if operator != "" {
		updates["editedBy"] = operator
	}
	err := c.db.WithContext(ctx).Model(&artifactBlob{ID: blobID}).
		Updates(updates).Error
	if err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *blobs) DeleteById(ctx context.Context, blobID int64) error {
	result := c.db.WithContext(ctx).Delete(&artifactBlob{}, blobID)
	if result.Error != nil {
		return database.ProcessSQLErrorf(ctx, result.Error)
	}
	if result.RowsAffected == 0 {
		return model.ErrBlobNoItemDeleted
	}
	return nil
}
