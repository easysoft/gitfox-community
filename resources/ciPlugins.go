package resources

import (
	"archive/zip"
	"context"
	"os"

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
		// 创建临时文件，使用 * 确保文件名以 .zip 结尾
		f, removeFunc, err := file.CreateTempFile("plugins-*.zip")
		if err != nil {
			return nil, nil, err
		}
		defer removeFunc()
		tempPath := f.Name()

		// 关闭文件以便 DownloadFile 可以写入
		_ = f.Close()

		// 下载 ZIP 文件到临时文件
		err = file.DownloadFile(tempPath, url)
		if err != nil {
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
