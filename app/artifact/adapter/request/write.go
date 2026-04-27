package request

import (
	"io"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter"
	storagedriver "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage/driver"
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
