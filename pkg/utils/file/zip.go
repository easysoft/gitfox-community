// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package file

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/easysoft/gitfox-community/pkg/errors"
)

// isZipFile checks if the file is a zip file.
func isZipFile(path string) bool {
	return strings.HasSuffix(path, ".zip")
}

// ReadZipFile reads a zip file and returns a zip.ReadCloser and a close function.
func ReadZipFile(path string) (*zip.ReadCloser, func(), error) {
	if !isZipFile(path) {
		return nil, nil, errors.New("not a zip file")
	}

	zipReader, err := zip.OpenReader(path)
	if err != nil {
		return nil, nil, err
	}

	closeFunc := func() {
		_ = zipReader.Close()
	}

	return zipReader, closeFunc, nil
}

type ZipData struct {
	id      int
	level   int
	Path    string // 文件或目录的完整路径，如: plugins/build/go/build_go.yaml
	Type    string // file  dir
	Name    string // xxx.yaml or xxx.json
	Suffix  string // .yaml or .json
	Content bytes.Buffer
	Next    any // 多级嵌套
	Read    ReadInfo
}
type ReadInfo struct {
	IDs    []int
	Suffix string
}

// SearchPath 查找指定path 并返回data
func (zd *ZipData) SearchPath(path string) *ZipData {
	// 如果当前节点的路径匹配，返回当前节点
	if zd.Path == path {
		return zd
	}

	// 如果当前节点不是目录或 Next 为空，返回 nil
	if zd.Type != "dir" || zd.Next == nil {
		return nil
	}

	// 递归查找子节点
	nextMap, ok := zd.Next.(map[string]*ZipData)
	if !ok {
		return nil
	}

	for _, child := range nextMap {
		if result := child.SearchPath(path); result != nil {
			return result
		}
	}

	return nil
}

// GetDirName 根据负数索引返回路径中的目录名
// 例如：plugins/build/go/build_go.yaml
// index = -1 返回 "go" (倒数第二个部分，文件的父目录)
// index = -2 返回 "build"
// index = -3 返回 "plugins"
func (zd *ZipData) GetDirName(index int) string {
	if index >= 0 {
		return ""
	}

	parts := strings.Split(zd.Path, "/")
	if len(parts) == 0 {
		return ""
	}

	// 计算实际索引位置
	// -1 表示倒数第二个（因为最后一个是文件名）
	actualIndex := len(parts) - 1 + index

	if actualIndex < 0 || actualIndex >= len(parts) {
		return ""
	}

	return parts[actualIndex]
}

// NextFile 查找并返回 data 内指定后缀的文件，返回后 id 记录在 Read 中，不重复返回，Read 中都已经返回则返回 nil
func (zd *ZipData) NextFile(suffix string) *ZipData {
	result := zd.nextFileRecursive(func(f *ZipData) bool {
		return f.Type == "file" && f.Suffix == suffix
	}, zd)
	if result != nil {
		// 在调用者节点上记录 Read 信息
		zd.Read.IDs = append(zd.Read.IDs, result.id)
		zd.Read.Suffix = suffix
	}
	return result
}

// NextFileByName 查找并返回 data 内指定文件名的文件，返回后 id 记录在 Read 中，不重复返回。
// 用于文件名固定的场景（如新格式插件包内的 bin.json / ui.json）。
func (zd *ZipData) NextFileByName(name string) *ZipData {
	result := zd.nextFileRecursive(func(f *ZipData) bool {
		return f.Type == "file" && f.Name == name
	}, zd)
	if result != nil {
		// 在调用者节点上记录 Read 信息
		zd.Read.IDs = append(zd.Read.IDs, result.id)
		zd.Read.Suffix = filepath.Ext(name)
	}
	return result
}

// nextFileRecursive 递归查找文件（内部方法）
// match: 匹配谓词，决定当前节点是否为目标文件
// caller: 最初的调用者，用于检查 Read 状态
func (zd *ZipData) nextFileRecursive(match func(*ZipData) bool, caller *ZipData) *ZipData {
	// 如果当前节点是文件且匹配
	if match(zd) {
		// 检查该文件是否已经在 caller 的 Read 中返回过
		for _, readID := range caller.Read.IDs {
			if readID == zd.id {
				return nil
			}
		}
		// 未返回过，返回该文件
		return zd
	}

	// 如果当前节点不是目录或 Next 为空，返回 nil
	if zd.Type != "dir" || zd.Next == nil {
		return nil
	}

	// 递归查找子节点
	nextMap, ok := zd.Next.(map[string]*ZipData)
	if !ok {
		return nil
	}

	for _, child := range nextMap {
		if result := child.nextFileRecursive(match, caller); result != nil {
			return result
		}
	}

	return nil
}

// CleanRead 清空Read (递归清空所有子节点)
func (zd *ZipData) CleanRead() {
	zd.Read.IDs = nil
	zd.Read.Suffix = ""

	// 递归清空子节点的 Read
	if zd.Type == "dir" && zd.Next != nil {
		if nextMap, ok := zd.Next.(map[string]*ZipData); ok {
			for _, child := range nextMap {
				child.CleanRead()
			}
		}
	}
}

// GetZipData 解析 zip.ReadCloser，构建并返回 ZipData 目录树。
// 安全措施：
//   - 过滤包含 ".." 的路径遍历条目（Zip Slip 防护）
//   - 使用 io.LimitReader 限制单个文件最大读取量（解压炸弹防护）
func GetZipData(zipFile *zip.ReadCloser) (ZipData, error) {
	// 全局自增 id，用于标识每个节点的唯一性
	currentID := 0
	root := ZipData{
		id:    currentID,
		level: 0,
		Path:  "",
		Type:  "dir",
		Name:  "root",
		Next:  make(map[string]*ZipData),
	}
	currentID++

	// dirMap 用于快速定位已创建的目录节点，避免重复创建
	dirMap := make(map[string]*ZipData)
	dirMap[""] = &root

	for _, file := range zipFile.File {
		// 跳过根目录或空名条目
		if file.Name == "" || file.Name == "/" {
			continue
		}

		// 安全检查：路径清洗，防御 Zip Slip 路径遍历攻击
		safePath, ok := sanitizePath(file.Name)
		if !ok {
			continue // 包含 ".." 的恶意路径，静默跳过
		}

		// 移除末尾的斜杠，统一格式
		cleanPath := strings.TrimSuffix(safePath, "/")
		parts := strings.Split(cleanPath, "/")

		// 逐层遍历路径，确保所有父目录节点都已创建
		for i := 0; i < len(parts); i++ {
			currentPath := strings.Join(parts[:i+1], "/")
			parentPath := ""
			if i > 0 {
				parentPath = strings.Join(parts[:i], "/")
			}

			// 如果这个路径已经处理过，跳过
			if _, exists := dirMap[currentPath]; exists {
				continue
			}

			// 判断当前节点是文件还是目录
			isLastPart := i == len(parts)-1
			isFile := isLastPart && !file.FileInfo().IsDir()

			data := &ZipData{
				id:    currentID,
				level: i + 1,
				Path:  currentPath,
				Name:  parts[i],
			}
			currentID++

			if isFile {
				data.Type = "file"
				data.Suffix = filepath.Ext(parts[i])

				// 读取文件内容（使用 LimitReader 进行解压炸弹防护）
				fc, err := file.Open()
				if err != nil {
					return ZipData{}, err
				}
				limitedReader := io.LimitReader(fc, maxPluginFileSize+1)
				n, err := io.Copy(&data.Content, limitedReader)
				fc.Close()
				if err != nil {
					return ZipData{}, err
				}
				// 如果实际读取量超过上限，说明文件过大，跳过该文件
				if n > maxPluginFileSize {
					continue
				}
			} else {
				data.Type = "dir"
				data.Next = make(map[string]*ZipData)
			}

			// 将当前节点挂载到父目录
			parent := dirMap[parentPath]
			if parent.Next == nil {
				parent.Next = make(map[string]*ZipData)
			}
			parent.Next.(map[string]*ZipData)[data.Name] = data

			// 如果是目录，添加到 dirMap 以便后续引用
			if !isFile {
				dirMap[currentPath] = data
			}
		}
	}

	return root, nil
}

// ZipDir walks root and streams a ZIP archive to w.
func ZipDir(ctx context.Context, root string, w io.Writer) (err error) {
	if w == nil {
		return os.ErrInvalid
	}

	zipWriter := zip.NewWriter(w)
	defer func() {
		if closeErr := zipWriter.Close(); err == nil {
			err = closeErr
		}
	}()

	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if path == root {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		// 仅处理目录、普通文件、符号链接；其它类型跳过
		if !info.IsDir() && !info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)
		if info.IsDir() {
			header.Name += "/"
		} else if info.Mode().IsRegular() {
			header.Method = zip.Deflate
		}

		archiveEntry, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			_, err = io.WriteString(archiveEntry, target)
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(archiveEntry, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
	return err
}
