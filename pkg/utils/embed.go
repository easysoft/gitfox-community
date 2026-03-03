package utils

import (
	"embed"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

// CopyEmbeddedDir 从嵌入文件系统复制整个目录
func CopyEmbeddedDir(embeddedFS embed.FS, srcDir, dstDir string) error {

	// 创建目标目录
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return errors.WithI18nKey(errors.Wrap(err, "[embed]failed to mkdir"), i18n.ErrorEmbedFailedToMkdir)
	}

	// 规范化嵌入 FS 的源路径为 POSIX 风格（embed.FS 使用 '/'）
	srcDir = path.Clean(strings.ReplaceAll(srcDir, "\\", "/"))

	// 读取目录内容
	entries, err := embeddedFS.ReadDir(srcDir)
	if err != nil {
		return errors.WithI18nKey(errors.Wrap(err, "[embed]failed to read dir"), i18n.ErrorEmbedFailedToReadDir)
	}

	for _, entry := range entries {
		// 对嵌入 FS 使用 POSIX 路径拼接
		srcPath := path.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			// 递归复制子目录
			if err := CopyEmbeddedDir(embeddedFS, srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// 复制文件
			if err := CopyEmbeddedFile(embeddedFS, srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyEmbeddedFile 从嵌入文件系统复制文件
func CopyEmbeddedFile(embeddedFS embed.FS, srcPath, dstPath string) error {
	data, err := embeddedFS.ReadFile(srcPath)
	if err != nil {
		return errors.WithI18nKey(errors.Wrap(err, "[embed]failed to read file"), i18n.ErrorEmbedFailedToReadFile)
	}

	// 根据文件扩展名确定权限
	var perm os.FileMode = 0644
	if strings.HasSuffix(srcPath, ".sh") || strings.HasSuffix(srcPath, ".py") || strings.HasSuffix(srcPath, ".pl") {
		perm = 0755
	}

	if err := os.WriteFile(dstPath, data, perm); err != nil {
		return errors.WithI18nKey(errors.Wrap(err, "[embed]failed to write file"), i18n.ErrorEmbedFailedToWriteFile)
	}

	return nil
}
