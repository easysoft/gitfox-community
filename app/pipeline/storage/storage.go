package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"time"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage"
	storagedriver "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage/driver"
)

const (
	pipelinePathFormatter      = "executions/%d/stages/%d/%s"
	cacheDataPathFormatter     = "cache/%s/data"
	cacheMetadataPathFormatter = "cache/%s/metadata.json"
)

type PipelineStorage interface {
	PutStageFile(ctx context.Context, executionId, stageId int64, fileName string, data []byte) error
	GetStageFile(ctx context.Context, executionId, stageId int64, fileName string) ([]byte, error)
	PutCacheFile(ctx context.Context, r io.Reader, meta *pipelineModel.CiCacheMetadata) error
	GetCacheFile(ctx context.Context, key string) (io.ReadCloser, error)
	ExistCacheFile(ctx context.Context, key string) (bool, error)
}

type pipelineStorage struct {
	base storage.ContentStorage
}

func (s *pipelineStorage) PutStageFile(ctx context.Context, executionId, stageId int64, fileName string, data []byte) error {
	absPath := fmt.Sprintf(pipelinePathFormatter, executionId, stageId, fileName)
	return s.base.Put(ctx, absPath, data)
}

func (s *pipelineStorage) GetStageFile(ctx context.Context, executionId, stageId int64, fileName string) ([]byte, error) {
	absPath := fmt.Sprintf(pipelinePathFormatter, executionId, stageId, fileName)
	return s.base.Get(ctx, absPath)
}

func (s *pipelineStorage) PutCacheFile(ctx context.Context, r io.Reader, meta *pipelineModel.CiCacheMetadata) error {
	blobPath := genCiCachePath(meta.Key)
	dataPath := fmt.Sprintf(cacheDataPathFormatter, blobPath)

	w, err := s.base.GetWriter(ctx, dataPath, false)
	if err != nil {
		return err
	}

	size, err := io.Copy(w, r)
	if err != nil {
		return err
	}
	meta.Size = size

	if err = s.saveCacheMetadata(ctx, meta); err != nil {
		_ = w.Cancel(ctx)
		return err
	}

	return w.Commit(ctx)
}

func (s *pipelineStorage) GetCacheFile(ctx context.Context, key string) (io.ReadCloser, error) {
	blobPath := genCiCachePath(key)
	dataPath := fmt.Sprintf(cacheDataPathFormatter, blobPath)

	reader, err := s.base.Open(ctx, dataPath)
	if err != nil {
		return nil, err
	}

	meta, err := s.getCacheMetadata(ctx, key)
	if err == nil {
		meta.Downloads += 1
		meta.PulledAt = time.Now().UnixMilli()
		_ = s.saveCacheMetadata(ctx, meta)
	}

	return reader, nil
}

func (s *pipelineStorage) ExistCacheFile(ctx context.Context, key string) (bool, error) {
	blobPath := genCiCachePath(key)
	dataPath := fmt.Sprintf(cacheDataPathFormatter, blobPath)

	_, err := s.base.Stat(ctx, dataPath)
	if err != nil {
		var pathNotFoundError storagedriver.PathNotFoundError
		if errors.As(err, &pathNotFoundError) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *pipelineStorage) saveCacheMetadata(ctx context.Context, meta *pipelineModel.CiCacheMetadata) error {
	blobPath := genCiCachePath(meta.Key)
	metadataPath := fmt.Sprintf(cacheMetadataPathFormatter, blobPath)

	metaBody, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	if err = s.base.Put(ctx, metadataPath, metaBody); err != nil {
		return err
	}
	return nil
}

func (s *pipelineStorage) getCacheMetadata(ctx context.Context, key string) (*pipelineModel.CiCacheMetadata, error) {
	blobPath := genCiCachePath(key)
	metadataPath := fmt.Sprintf(cacheMetadataPathFormatter, blobPath)

	metaBody, err := s.base.Get(ctx, metadataPath)
	if err != nil {
		logger.GetLogger(ctx).Errorw("get cache file metadata failed", "error", err, "key", key)
		return nil, err
	}

	var meta pipelineModel.CiCacheMetadata
	if err = json.Unmarshal(metaBody, &meta); err != nil {
		logger.GetLogger(ctx).Errorw("unmarshal cache metadata failed", "error", err, "key", key)
		return nil, err
	}

	return &meta, nil
}

func genCiCachePath(key string) string {
	// should never happen, validate in handler layer
	if len(key) < 4 {
		return key
	}

	return path.Join(key[0:2], key[2:4], key)
}
