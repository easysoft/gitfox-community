package resources

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/file"
)

// LoadAllCIPlugins 加载所有 CI 插件。
// 返回 ZIP 读取器、清理函数和错误。调用方必须调用清理函数来释放资源。
func LoadAllCIPlugins(ctx context.Context, url string) (*zip.ReadCloser, func(), error) {
	if url == "" {
		return nil, nil, errors.New("plugins url not provided to read schemas from")
	}

	pluginsPath := url
	var cleanupFunc func() error

	// 检查是否是本地路径
	if _, err := os.Stat(url); err != nil { // 本地路径不存在 - 必须是一个远程链接
		// 使用带时间戳的文件名存储在临时目录，避免文件名冲突
		timestamp := time.Now().Format("20060102_150405")
		fileName := fmt.Sprintf("plugins_%s.zip", timestamp)
		tempPath := filepath.Join(os.TempDir(), fileName)

		cleanupFunc = func() error {
			return os.Remove(tempPath)
		}

		// 下载 ZIP 文件到临时文件
		err = file.DownloadFile(tempPath, url)
		if err != nil {
			_ = cleanupFunc()
			return nil, nil, errors.Newf("could not download remote zip: %v", err)
		}

		pluginsPath = tempPath
	}

	// 打开 ZIP 读取器
	zipReader, closeZip, err := file.ReadZipFile(pluginsPath)
	if err != nil {
		if cleanupFunc != nil {
			_ = cleanupFunc()
		}
		return nil, nil, errors.Newf("could not open zip for reading: %v", err)
	}

	// 组合清理函数
	finalCleanup := func() {
		closeZip()
		if cleanupFunc != nil {
			_ = cleanupFunc()
		}
	}

	return zipReader, finalCleanup, nil
}
