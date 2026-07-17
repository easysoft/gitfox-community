package adapter

import (
	"context"
	"errors"
	"path"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage"
	storagedriver "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage/driver"

	"github.com/google/uuid"
)

type StorageFileWriter struct {
	Layout StorageLayout
	Writer storagedriver.FileWriter
	Key    string
	Path   string
}

func (s *StorageFileWriter) SetPath(p string) {
	s.Path = p
}

type StorageFileWriterList []*StorageFileWriter

func NewStorageFileWriter(ctx context.Context, path string, append bool, store storage.ContentStorage) (*StorageFileWriter, error) {
	fileWriter, err := store.GetWriter(ctx, path, append)
	if err != nil {
		return nil, err
	}
	return &StorageFileWriter{Writer: fileWriter, Key: path}, nil
}

func NewStorageBlobWriter(ctx context.Context, store storage.ContentStorage, ref string, append bool) (*StorageFileWriter, error) {
	fileWriter, err := store.GetWriter(ctx, BlobPath(ref), append)
	if err != nil {
		return nil, err
	}
	return &StorageFileWriter{Writer: fileWriter, Key: ref}, nil
}

func NewRandomBlobWriter(ctx context.Context, store storage.ContentStorage) (storagedriver.FileWriter, string, error) {
	ref := GenerateSha256()
	storePath := BlobPath(ref)

	_, err := store.Stat(ctx, storePath)
	if err == nil {
		return nil, "", ErrStorageFileAlreadyExists
	}

	var notFoundErr storagedriver.PathNotFoundError
	if errors.As(err, &notFoundErr) {
		fileWriter, e := store.GetWriter(ctx, storePath, false)
		if e != nil {
			return nil, ref, e
		}
		return fileWriter, ref, nil
	}
	return nil, ref, err
}

func NewTemporaryWriter(ctx context.Context, store storage.ContentStorage, temporaryId string, append bool) (storagedriver.FileWriter, error) {
	temporaryPath := TemporaryPath(temporaryId)
	fileWriter, err := store.GetWriter(ctx, temporaryPath, append)
	if err != nil {
		return nil, err
	}
	return fileWriter, nil
}

type StorageLayout string

const (
	StorageLayoutBlob StorageLayout = "blob"
	StorageLayoutPath StorageLayout = "path"
)

func BlobPath(p string) string {
	// blob ref 默认按前 4 位分片成两级目录；若已带 `/` 则视为完整存储路径直接复用。
	if len(p) < 4 {
		return p
	}
	if strings.Contains(p, "/") {
		return p
	}
	return path.Join(p[0:2], p[2:4], p)
}

func GenerateSha256() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func TemporaryPath(p string) string {
	return path.Join("_uploads", BlobPath(strings.ReplaceAll(p, "-", "")))
}

type ViewDescriptor struct {
	ArtifactID     int64
	ArtifactName   string
	ArtifactCode   string
	ArtifactFormat artifact.ArtifactFormat
	StorageID      int64
	Store          storage.ContentStorage
	Space          *model.SpaceRef
	Repo           *model.RepoRef
}

type SavedAsset struct {
	Object     *model.ArtifactAsset
	IsOverride bool
	Attr       ModelAttribute
	// AlreadyExists 标记本次 Sync 命中"同名同内容"分支：未真正落新数据，
	// 直接复用已有 asset。Controller 层据此切换响应文案为"文件已存在"。
	AlreadyExists bool
}
