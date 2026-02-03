package file

import (
	"archive/zip"
	"bytes"
	"io"
	"path/filepath"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
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

// NextFile 查找并返回data 内的指定后缀的文件，返回后id记录在 Read 中，不重复返回 ，Read 中都已经返回则返回nil
func (zd *ZipData) NextFile(suffix string) *ZipData {
	result := zd.nextFileRecursive(suffix, zd)
	if result != nil {
		// 在调用者节点上记录 Read 信息
		zd.Read.IDs = append(zd.Read.IDs, result.id)
		zd.Read.Suffix = suffix
	}
	return result
}

// nextFileRecursive 递归查找文件（内部方法）
// caller: 最初的调用者，用于检查 Read 状态
func (zd *ZipData) nextFileRecursive(suffix string, caller *ZipData) *ZipData {
	// 如果当前节点是文件且后缀匹配
	if zd.Type == "file" && zd.Suffix == suffix {
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
		if result := child.nextFileRecursive(suffix, caller); result != nil {
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

func GetZipData(zipFile *zip.ReadCloser) (ZipData, error) {
	// 全局自增 id
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

	// 构建目录树
	dirMap := make(map[string]*ZipData)
	dirMap[""] = &root

	for _, file := range zipFile.File {
		// 跳过根目录
		if file.Name == "" || file.Name == "/" {
			continue
		}

		// 移除末尾的斜杠
		cleanPath := strings.TrimSuffix(file.Name, "/")
		parts := strings.Split(cleanPath, "/")

		// 遍历路径的每一部分，确保所有父目录都存在
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

			// 判断是文件还是目录
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

				// 读取文件内容
				fc, err := file.Open()
				if err != nil {
					return ZipData{}, err
				}
				_, err = io.Copy(&data.Content, fc)
				fc.Close()
				if err != nil {
					return ZipData{}, err
				}
			} else {
				data.Type = "dir"
				data.Next = make(map[string]*ZipData)
			}

			// 添加到父目录
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
