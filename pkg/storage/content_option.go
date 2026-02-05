package storage

import (
	"path"
	"strings"
)

type CommonContentStoreOption interface {
	Apply(store *commonContentStore)
}

type prefixDirOption struct {
	dir string
}

func (o prefixDirOption) Apply(store *commonContentStore) {
	store.pathPrefix = "/" + strings.TrimPrefix(path.Clean(o.dir), "/")
}

func WithPrefix(dir string) CommonContentStoreOption {
	return prefixDirOption{dir}
}
