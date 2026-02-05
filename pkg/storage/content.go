package storage

import (
	"context"
	"errors"
	"io"
	"sort"
	"strings"

	storagedriver "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage/driver"
)

type ContentStorage interface {
	Get(ctx context.Context, path string) ([]byte, error)
	Open(ctx context.Context, path string) (io.ReadCloser, error)
	Put(ctx context.Context, path string, contents []byte) error
	Save(ctx context.Context, path string, r io.Reader, size int64) error
	GetWriter(ctx context.Context, path string, append bool) (storagedriver.FileWriter, error)
	Stat(ctx context.Context, path string) (storagedriver.FileInfo, error)
	Delete(ctx context.Context, path string) error
	List(ctx context.Context, path string) ([]string, error)

	PrefixDir() string
}

type commonContentStore struct {
	pathPrefix string
	driver     storagedriver.StorageDriver
}

func NewCommonContentStore(ctx context.Context, driver storagedriver.StorageDriver, opts ...CommonContentStoreOption) (ContentStorage, error) {
	return newCommonContentStore(ctx, driver, opts...)
}

func newCommonContentStore(ctx context.Context, driver storagedriver.StorageDriver, opts ...CommonContentStoreOption) (*commonContentStore, error) {
	ccs := &commonContentStore{
		driver: driver,
	}
	for _, opt := range opts {
		opt.Apply(ccs)
	}
	return ccs, nil
}

func (ccs *commonContentStore) Get(ctx context.Context, path string) ([]byte, error) {
	return ccs.driver.GetContent(ctx, ccs.buildPath(path))
}

func (ccs *commonContentStore) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	fInfo, e := ccs.stat(ctx, ccs.buildPath(path))
	if e != nil {
		return nil, e
	}

	return newFileReader(ctx, ccs.driver, fInfo.Path(), fInfo.Size())
}

func (ccs *commonContentStore) Put(ctx context.Context, path string, contents []byte) error {
	return ccs.driver.PutContent(ctx, ccs.buildPath(path), contents)
}

func (ccs *commonContentStore) Save(ctx context.Context, path string, r io.Reader, size int64) error {
	writer, err := ccs.driver.Writer(ctx, ccs.buildPath(path), false)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, r)
	if err != nil {
		if cErr := writer.Cancel(ctx); cErr != nil {
			return errors.Join(err, cErr)
		}
		return err
	}
	return writer.Commit(ctx)
}

func (ccs *commonContentStore) GetWriter(ctx context.Context, path string, append bool) (storagedriver.FileWriter, error) {
	return ccs.driver.Writer(ctx, ccs.buildPath(path), append)
}

func (ccs *commonContentStore) Stat(ctx context.Context, path string) (storagedriver.FileInfo, error) {
	return ccs.stat(ctx, ccs.buildPath(path))
}

func (ccs *commonContentStore) stat(ctx context.Context, path string) (storagedriver.FileInfo, error) {
	return ccs.driver.Stat(ctx, path)
}

func (ccs *commonContentStore) PrefixDir() string {
	return ccs.pathPrefix
}

func (ccs *commonContentStore) Delete(ctx context.Context, path string) error {
	return ccs.driver.Delete(ctx, ccs.buildPath(path))
}

func (ccs *commonContentStore) List(ctx context.Context, path string) ([]string, error) {
	paths, err := ccs.driver.List(ctx, ccs.buildPath(path))
	if err != nil {
		return nil, err
	}

	sort.Strings(paths)
	return paths, nil
}

func (ccs *commonContentStore) buildPath(path string) string {
	return strings.Join([]string{
		ccs.pathPrefix, strings.TrimPrefix(path, "/"),
	}, "/")
}
