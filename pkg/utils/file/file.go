package file

import (
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

// DownloadFile will download the file exist in url to a local file.
// 下载文件到本地 path: 本地文件路径(或者是目录) url: 远程文件URL
func DownloadFile(path string, url string) error {
	if IsDir(path) {
		// 如果是目录，则从 URL 中提取文件名并拼接路径
		_, fileName := filepath.Split(url)
		if fileName == "" {
			return errors.New("cannot extract filename from url")
		}
		path = filepath.Join(path, fileName)
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp != nil && resp.Body != nil {
		defer func() { _ = resp.Body.Close() }()
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
