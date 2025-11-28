package file

import (
	"os"
	"path/filepath"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
)

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
