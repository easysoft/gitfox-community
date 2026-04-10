package file

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
)

const (
	// maxPluginFileSize 单个插件文件最大允许读取量（100 MB）。
	// 超过此阈值的文件将被跳过，用于防御解压炸弹（Zip Bomb / Tar Bomb）攻击。
	maxPluginFileSize = 100 << 20 // 100 MB
)

// GetArchiveData 解析归档文件并返回统一的 ZipData 目录树结构。
// 当前支持 ".zip"、".tar.gz"、".tgz" 三种归档格式。
// 内部会根据文件后缀自动路由到对应的解析器。
func GetArchiveData(path string) (ZipData, error) {
	if isZipFile(path) {
		zipReader, closeFunc, err := ReadZipFile(path)
		if err != nil {
			return ZipData{}, err
		}
		defer closeFunc()
		return GetZipData(zipReader)
	}

	if isTarGzFile(path) {
		return getTarGzData(path)
	}

	return ZipData{}, errors.New("unsupported archive format")
}

// isTarGzFile 通过文件后缀（大小写不敏感）判断是否为 tar.gz 压缩包。
func isTarGzFile(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz")
}

// isPathTraversal 检查路径中是否包含目录遍历元素（".."）。
// 返回 true 表示路径不安全，应当跳过。
func isPathTraversal(name string) bool {
	for _, part := range strings.Split(name, "/") {
		if part == ".." {
			return true
		}
	}
	return false
}

// sanitizePath 对归档条目路径执行安全清洗。
// 移除前导斜杠和反斜杠，防止绝对路径注入。
// 返回清洗后的路径和是否安全的布尔值。
func sanitizePath(name string) (string, bool) {
	// 移除前导斜杠和反斜杠
	cleaned := strings.TrimLeft(name, "/\\")
	// 规范化路径分隔符（Windows 反斜杠 -> 正斜杠），跨平台一致
	cleaned = strings.ReplaceAll(cleaned, "\\", "/")

	if isPathTraversal(cleaned) {
		return "", false
	}
	return cleaned, true
}

// getTarGzData 解析 tar.gz 压缩包，构建并返回 ZipData 目录树。
// 安全措施：
//   - 过滤包含 ".." 的路径遍历条目（Zip Slip 防护）
//   - 跳过符号链接、硬链接等非常规文件类型
//   - 使用 io.LimitReader 限制单个文件最大读取量（解压炸弹防护）
func getTarGzData(path string) (ZipData, error) {
	f, err := os.Open(path)
	if err != nil {
		return ZipData{}, err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return ZipData{}, err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

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

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // 归档文件遍历结束
		}
		if err != nil {
			return ZipData{}, err
		}

		// 安全检查：跳过符号链接、硬链接等非常规条目，仅处理普通文件和目录
		switch header.Typeflag {
		case tar.TypeReg, tar.TypeDir:
			// 允许：普通文件和目录
		default:
			// 跳过：符号链接(TypeSymlink)、硬链接(TypeLink)、设备文件等
			continue
		}

		// 跳过根目录或空名条目
		if header.Name == "" || header.Name == "/" {
			continue
		}

		// 安全检查：路径清洗，防御 Zip Slip 路径遍历攻击
		safePath, ok := sanitizePath(header.Name)
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

			// 如果这个路径已经处理过（在 dirMap 中存在），跳过
			if _, exists := dirMap[currentPath]; exists {
				continue
			}

			// 判断当前节点是文件还是目录
			isLastPart := i == len(parts)-1
			isFile := isLastPart && header.Typeflag == tar.TypeReg

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

				// 安全措施：使用 LimitReader 限制单个文件的最大读取量，防御解压炸弹
				limitedReader := io.LimitReader(tr, maxPluginFileSize+1)
				n, err := io.Copy(&data.Content, limitedReader)
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
			if parent == nil {
				// 父目录不存在时自动创建（理论上不会触发，因为逐层遍历已保证）
				parent = &ZipData{
					id:    currentID,
					level: i,
					Path:  parentPath,
					Name:  parts[i-1],
					Type:  "dir",
					Next:  make(map[string]*ZipData),
				}
				currentID++
				dirMap[parentPath] = parent
			}
			if parent.Next == nil {
				parent.Next = make(map[string]*ZipData)
			}
			parent.Next.(map[string]*ZipData)[data.Name] = data

			// 如果是目录，添加到 dirMap 以便后续子节点引用
			if !isFile {
				dirMap[currentPath] = data
			}
		}
	}

	return root, nil
}
