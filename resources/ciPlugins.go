// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package resources

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/utils/file"
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
		// 先去掉 query string（? 后内容），再根据路径后缀判断压缩格式
		// 例如：https://example.com/plugins.tar.gz?t=123 → 取 .tar.gz
		urlPath := url
		if idx := strings.Index(url, "?"); idx != -1 {
			urlPath = url[:idx]
		}
		suffix := ".zip"
		if strings.HasSuffix(urlPath, ".tar.gz") || strings.HasSuffix(urlPath, ".tgz") {
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
		// 为避免 CDN/代理缓存，下载前将 URL 中的 t 参数替换为当前时间戳
		// 规则：有 t 参数 → 更新为最新时间戳；无 t 参数 → 追加；无 query → 新增 ?t=
		downloadURL := withFreshTimestamp(url)
		err = file.DownloadFile(ctx, tempPath, downloadURL)
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

// withFreshTimestamp 返回带有最新时间戳的 URL，用于绕过 CDN/代理缓存。
// 规则：
//   - 已有 t 参数 → 替换为当前 Unix 时间戳
//   - 无 t 参数但有其他 query → 追加 &t=<timestamp>
//   - 无 query → 追加 ?t=<timestamp>
//
// 若 URL 解析失败（非标准格式），原样返回并追加 ?t=<timestamp>。
func withFreshTimestamp(rawURL string) string {
	ts := fmt.Sprintf("%d", time.Now().Unix())

	u, err := url.Parse(rawURL)
	if err != nil {
		// 兜底：直接拼接
		if strings.Contains(rawURL, "?") {
			return rawURL + "&t=" + ts
		}
		return rawURL + "?t=" + ts
	}

	q := u.Query()
	q.Set("t", ts)
	u.RawQuery = q.Encode()
	return u.String()
}
