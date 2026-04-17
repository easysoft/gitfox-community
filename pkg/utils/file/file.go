package file

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
)

// FindFilesRecursive 递归查找目录下所有指定文件名的文件路径
func FindFilesRecursive(dir, target string) ([]string, error) {
	// 1. 校验目录是否有效（存在且为目录）
	info, err := os.Stat(dir)
	if err != nil {
		return nil, errors.Newf("目录无效: %s", err.Error())
	}
	if !info.IsDir() {
		return nil, errors.New("指定路径不是目录")
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, errors.Newf("无法获取绝对路径: %s", err.Error())
	}

	var result []string

	// 3. 递归遍历目录
	err = filepath.WalkDir(absDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return filepath.SkipDir // 跳过当前目录，不影响其他目录遍历
		}
		if !d.IsDir() && d.Name() == target {
			absPath, _ := filepath.Abs(path)
			result = append(result, absPath)
		}
		return nil
	})

	if err != nil {
		return nil, errors.Newf("遍历目录失败: %s", err.Error())
	}

	return result, nil
}

// DownloadFile 下载远程文件到本地。
// 支持通过 context 进行超时控制和取消操作。
// path: 本地文件路径(或者是目录) url: 远程文件URL ctx: 上下文（用于超时和取消控制）
func DownloadFile(ctx context.Context, path string, url string) error {
	if IsDir(path) {
		// 如果是目录，则从 URL 中提取文件名并拼接路径
		_, fileName := filepath.Split(url)
		if fileName == "" {
			return errors.New("cannot extract filename from url")
		}
		path = filepath.Join(path, fileName)
	}

	// 使用 http.NewRequestWithContext 替代废弃的 http.Get，支持 context 取消
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return errors.Newf("could not create HTTP request: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp != nil && resp.Body != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.Newf("received non-2xx status code: %d", resp.StatusCode)
	}

	// 创建或者覆盖
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return errors.Newf("could not copy response body output to file: %v", err)
	}
	return nil
}

// CreateTempFile creates a temporary file with the given pattern and returns the file, a remove function, and any error.
// 创建临时文件 pattern: 文件名模板（例如 "plugins.zip"）
// 返回: 文件对象, 删除函数, 错误信息
func CreateTempFile(pattern string) (*os.File, func() error, error) {
	f, err := os.CreateTemp(os.TempDir(), pattern)
	if err != nil {
		return nil, nil, errors.Newf("could not create temp file: %v", err)
	}

	removeFunc := func() error {
		_ = f.Close()
		return os.Remove(f.Name())
	}

	return f, removeFunc, nil
}
