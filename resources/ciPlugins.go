package resources

import (
	"context"
	"os"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/file"
)

// DownloadCIPlugins 下载 CI 插件压缩包。
// 返回压缩包本地路径，清理函数和错误。调用方必须调用清理函数来删除临时文件。
func DownloadCIPlugins(ctx context.Context, url string) (string, func(), error) {
	if url == "" {
		return "", nil, errors.New("plugins url not provided to download from")
	}

	pluginsPath := url
	var cleanupFunc func()

	// 检查是否是本地路径
	if _, err := os.Stat(url); err != nil { // 本地路径不存在 - 必须是一个远程链接
		// 创建临时文件，使用 * 因为不知道是 zip 还是 tar.gz 结尾
		// 我们可以根据 URL 后缀简单判断
		suffix := ".zip"
		if strings.HasSuffix(url, ".tar.gz") || strings.HasSuffix(url, ".tgz") {
			suffix = ".tar.gz"
		}
		
		f, removeFunc, err := file.CreateTempFile("plugins-*" + suffix)
		if err != nil {
			return "", nil, err
		}
		cleanupFunc = func() { removeFunc() }
		tempPath := f.Name()

		// 关闭文件以便 DownloadFile 可以写入
		_ = f.Close()

		// 下载文件到临时文件
		err = file.DownloadFile(ctx, tempPath, url)
		if err != nil {
			cleanupFunc()
			return "", nil, errors.Newf("could not download remote archive: %v", err)
		}

		pluginsPath = tempPath
	}

	// 组合清理函数
	finalCleanup := func() {
		if cleanupFunc != nil {
			cleanupFunc()
		}
	}

	return pluginsPath, finalCleanup, nil
}
