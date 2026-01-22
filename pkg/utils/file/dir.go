package file

import "os"

// IsDir checks if the path is directory or not.
// 检查路径是否为目录
func IsDir(path string) bool {
	file, err := os.Stat(path)
	if err != nil {
		return false
	}
	return file.IsDir()
}
