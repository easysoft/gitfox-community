// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package s3

import (
	"context"
	"io"
)

// Interface provides an interface for the persistent log store backend.
type Interface interface {
	// Find returns a log stream from the datastore.
	Find(ctx context.Context, filePath string) (io.ReadCloser, error)

	// Create writes copies the log stream from Reader r to the datastore.
	Create(ctx context.Context, filePath string, r io.Reader) error

	// Update copies the log stream from Reader r to the datastore.
	Update(ctx context.Context, filePath string, r io.Reader) error

	// Delete purges the log stream from the datastore.
	Delete(ctx context.Context, filePath string) error
}
