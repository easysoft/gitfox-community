// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

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
