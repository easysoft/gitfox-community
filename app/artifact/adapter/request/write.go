// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package request

import (
	"io"

	"github.com/easysoft/gitfox-community/app/artifact/adapter"
	storagedriver "github.com/easysoft/gitfox-community/pkg/storage/driver"
)

// Write 同时向 storage writer、hash writer 和可选额外 writer 写入数据。
func Write(r io.Reader, fw storagedriver.FileWriter, extraWriters ...io.Writer) (int64, *adapter.Hash, error) {
	hw := adapter.NewHashWriter()
	writes := []io.Writer{hw, fw}
	writes = append(writes, extraWriters...)

	mw := io.MultiWriter(writes...)
	n, err := io.Copy(mw, r)
	if err != nil {
		return 0, nil, err
	}
	return n, hw.Sum(), nil
}
