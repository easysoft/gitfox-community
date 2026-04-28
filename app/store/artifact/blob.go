package artifact

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"

	"github.com/rs/zerolog/log"
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

func (c *blobs) GetById(ctx context.Context, blobId int64) (*model.ArtifactBlob, error) {
	var row artifactBlob
	if err := c.db.WithContext(ctx).First(&row, blobId).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return blobToModel(&row), nil
}

func (c *blobs) GetByRef(ctx context.Context, ref string, storageId int64) (*model.ArtifactBlob, error) {
	if err := validatorEmpty(ref, storageId); err != nil {
		return nil, model.ErrArgsValueEmpty
	}
	var row artifactBlob
	q := artifactBlob{Ref: ref, StorageID: storageId}
	if err := c.db.WithContext(ctx).Where(&q).First(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return blobToModel(&row), nil
}

func (c *blobs) GetByHistory(ctx context.Context, assetId, createdAt int64) (*model.ArtifactBlob, error) {
	var row artifactBlob
	tm := time.UnixMilli(createdAt)
	// 用原始 time.Time 查询，GORM 会调用 datetime.Local.Value() 做存储转换
	if err := c.db.WithContext(ctx).
		Where("assetID = ? AND createdDate = ?", assetId, tm).
		First(&row).Error; err != nil {
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

func (c *blobs) SoftDeleteById(ctx context.Context, blobId int64) error {
	err := c.db.WithContext(ctx).Model(&artifactBlob{ID: blobId}).
		UpdateColumn("deleted", 1).Error
	if err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *blobs) DeleteById(ctx context.Context, blobId int64) error {
	result := c.db.WithContext(ctx).Delete(&artifactBlob{}, blobId)
	if result.Error != nil {
		return database.ProcessSQLErrorf(ctx, result.Error)
	}
	if result.RowsAffected == 0 {
		return model.ErrBlobNoItemDeleted
	}
	return nil
}

func (c *blobs) History(ctx context.Context, assetId int64) ([]*model.ArtifactBlobRes, error) {
	var dst []*model.ArtifactBlobRes
	stmt := c.db.WithContext(ctx).Model(&artifactBlob{}).
		Select(`ops_artifact_blobs.id,ops_artifact_assets.path,ops_artifact_blobs.size,ops_artifact_blobs.ref,
			ops_artifact_blobs.createdBy,
			ops_artifact_blobs.checkSum,ops_artifact_blobs.metadata,ops_artifact_blobs.createdDate,ops_artifact_blobs.deleted`).
		Joins(`LEFT JOIN ops_artifact_assets on ops_artifact_blobs.assetID=ops_artifact_assets.id`).
		Where(`ops_artifact_blobs.assetID = ?`, assetId).
		Order("ops_artifact_blobs.createdDate DESC")

	if err := stmt.Find(&dst).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	for _, a := range dst {
		var checkSum model.CheckSumRes
		if err := json.Unmarshal([]byte(a.CheckSumString), &checkSum); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("parse artifact blob checksum failed")
		}
		a.CheckSum = &checkSum

		if a.MetadataString != "" {
			var metadata map[string]string
			if err := json.Unmarshal([]byte(a.MetadataString), &metadata); err != nil {
				log.Ctx(ctx).Error().Err(err).Msg("parse artifact blob metadata failed")
			}
			a.Metadata = metadata
		}

		a.Creator = a.CreatorName
	}
	return dst, nil
}
