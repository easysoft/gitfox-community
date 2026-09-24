// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package file

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createTestTarGz 创建一个测试用的 tar.gz 文件，包含指定的文件列表。
// 返回生成的 tar.gz 文件路径。
func createTestTarGz(t *testing.T, files []struct {
	name    string
	content string
}) string {
	t.Helper()

	tmpDir := t.TempDir()
	tarGzPath := filepath.Join(tmpDir, "test.tar.gz")

	f, err := os.Create(tarGzPath)
	if err != nil {
		t.Fatalf("创建 tar.gz 文件失败: %v", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, file := range files {
		isDir := strings.HasSuffix(file.name, "/")

		if isDir {
			header := &tar.Header{
				Name:     file.name,
				Typeflag: tar.TypeDir,
				Mode:     0755,
			}
			if err := tw.WriteHeader(header); err != nil {
				t.Fatalf("写入目录 header 失败: %v", err)
			}
		} else {
			header := &tar.Header{
				Name:     file.name,
				Typeflag: tar.TypeReg,
				Size:     int64(len(file.content)),
				Mode:     0644,
			}
			if err := tw.WriteHeader(header); err != nil {
				t.Fatalf("写入文件 header 失败: %v", err)
			}
			if _, err := tw.Write([]byte(file.content)); err != nil {
				t.Fatalf("写入文件内容失败: %v", err)
			}
		}
	}

	return tarGzPath
}

// createTestZipForArchive 创建一个测试用的 ZIP 文件，包含指定的文件列表。
func createTestZipForArchive(t *testing.T, files []struct {
	name    string
	content string
}) string {
	t.Helper()

	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("创建 ZIP 文件失败: %v", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	for _, file := range files {
		writer, err := w.Create(file.name)
		if err != nil {
			t.Fatalf("创建 ZIP 条目失败: %v", err)
		}
		if _, err := writer.Write([]byte(file.content)); err != nil {
			t.Fatalf("写入 ZIP 内容失败: %v", err)
		}
	}

	return zipPath
}

// TestGetArchiveData_TarGz_NormalParsing 测试 tar.gz 格式的正常解析
func TestGetArchiveData_TarGz_NormalParsing(t *testing.T) {
	files := []struct {
		name    string
		content string
	}{
		{"plugins/build/go/build_go.yaml", "yaml content 1"},
		{"plugins/build/go/build_go.json", "json content 1"},
		{"plugins/build/python/build_py.yaml", "yaml content 2"},
		{"plugins/deploy/deploy.yaml", "yaml content 3"},
		{"readme.txt", "readme content"},
	}

	tarGzPath := createTestTarGz(t, files)

	root, err := GetArchiveData(tarGzPath)
	if err != nil {
		t.Fatalf("GetArchiveData 解析 tar.gz 失败: %v", err)
	}

	// 验证根节点
	if root.Type != "dir" {
		t.Errorf("根节点类型应为 'dir'，实际为: %s", root.Type)
	}
	if root.Name != "root" {
		t.Errorf("根节点名称应为 'root'，实际为: %s", root.Name)
	}

	// 验证文件查找
	file1 := root.SearchPath("plugins/build/go/build_go.yaml")
	if file1 == nil {
		t.Fatal("未找到 plugins/build/go/build_go.yaml")
	}
	if file1.Content.String() != "yaml content 1" {
		t.Errorf("文件内容不匹配，期望 'yaml content 1'，实际为 '%s'", file1.Content.String())
	}
	if file1.Suffix != ".yaml" {
		t.Errorf("文件后缀不匹配，期望 '.yaml'，实际为 '%s'", file1.Suffix)
	}

	// 验证 NextFile 遍历
	var yamlFiles []*ZipData
	for {
		f := root.NextFile(".yaml")
		if f == nil {
			break
		}
		yamlFiles = append(yamlFiles, f)
	}
	if len(yamlFiles) != 3 {
		t.Errorf("期望找到 3 个 .yaml 文件，实际找到 %d 个", len(yamlFiles))
	}

	// 验证目录查找
	dir := root.SearchPath("plugins/build")
	if dir == nil {
		t.Fatal("未找到 plugins/build 目录")
	}
	if dir.Type != "dir" {
		t.Errorf("plugins/build 类型应为 'dir'，实际为: %s", dir.Type)
	}
}

// TestGetArchiveData_Zip_ViaTarGzRouter 测试 zip 格式经由 GetArchiveData 统一入口的路由
func TestGetArchiveData_Zip_ViaTarGzRouter(t *testing.T) {
	files := []struct {
		name    string
		content string
	}{
		{"plugins/test/test.yaml", "test yaml"},
		{"plugins/test/test.json", "test json"},
	}

	zipPath := createTestZipForArchive(t, files)

	root, err := GetArchiveData(zipPath)
	if err != nil {
		t.Fatalf("GetArchiveData 解析 zip 失败: %v", err)
	}

	file := root.SearchPath("plugins/test/test.yaml")
	if file == nil {
		t.Fatal("经由 GetArchiveData 路由后，未找到 plugins/test/test.yaml")
	}
	if file.Content.String() != "test yaml" {
		t.Errorf("内容不匹配，期望 'test yaml'，实际为 '%s'", file.Content.String())
	}
}

// TestGetArchiveData_UnsupportedFormat 测试不支持的归档格式
func TestGetArchiveData_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	txtPath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(txtPath, []byte("not an archive"), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	_, err := GetArchiveData(txtPath)
	if err == nil {
		t.Error("期望对不支持的格式返回错误，但没有")
	}
	if !strings.Contains(err.Error(), "unsupported archive format") {
		t.Errorf("错误信息不匹配，实际为: %v", err)
	}
}

// TestGetArchiveData_PathTraversal 测试路径遍历攻击过滤（Zip Slip 防护）
func TestGetArchiveData_PathTraversal(t *testing.T) {
	// 构造包含恶意路径的 tar.gz
	tmpDir := t.TempDir()
	tarGzPath := filepath.Join(tmpDir, "malicious.tar.gz")

	f, err := os.Create(tarGzPath)
	if err != nil {
		t.Fatalf("创建文件失败: %v", err)
	}

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	// 写入一个包含 ".." 路径遍历的恶意条目
	maliciousHeader := &tar.Header{
		Name:     "../../etc/passwd",
		Typeflag: tar.TypeReg,
		Size:     int64(len("malicious content")),
		Mode:     0644,
	}
	if err := tw.WriteHeader(maliciousHeader); err != nil {
		t.Fatalf("写入恶意 header 失败: %v", err)
	}
	if _, err := tw.Write([]byte("malicious content")); err != nil {
		t.Fatalf("写入恶意内容失败: %v", err)
	}

	// 同时写入一个正常条目
	normalHeader := &tar.Header{
		Name:     "plugins/safe.yaml",
		Typeflag: tar.TypeReg,
		Size:     int64(len("safe content")),
		Mode:     0644,
	}
	if err := tw.WriteHeader(normalHeader); err != nil {
		t.Fatalf("写入正常 header 失败: %v", err)
	}
	if _, err := tw.Write([]byte("safe content")); err != nil {
		t.Fatalf("写入正常内容失败: %v", err)
	}

	tw.Close()
	gw.Close()
	f.Close()

	root, err := GetArchiveData(tarGzPath)
	if err != nil {
		t.Fatalf("GetArchiveData 失败: %v", err)
	}

	// 验证恶意路径条目被过滤（不应存在于目录树中）
	malicious := root.SearchPath("../../etc/passwd")
	if malicious != nil {
		t.Error("恶意路径遍历条目不应出现在目录树中，但被找到了")
	}

	// 验证正常条目存在
	safe := root.SearchPath("plugins/safe.yaml")
	if safe == nil {
		t.Error("正常条目应当存在于目录树中，但未找到")
	}
}

// TestGetArchiveData_SymlinkFiltering 测试符号链接条目过滤
func TestGetArchiveData_SymlinkFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	tarGzPath := filepath.Join(tmpDir, "symlink.tar.gz")

	f, err := os.Create(tarGzPath)
	if err != nil {
		t.Fatalf("创建文件失败: %v", err)
	}

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	// 写入一个符号链接条目
	symlinkHeader := &tar.Header{
		Name:     "evil_link",
		Typeflag: tar.TypeSymlink,
		Linkname: "/etc/shadow",
	}
	if err := tw.WriteHeader(symlinkHeader); err != nil {
		t.Fatalf("写入符号链接 header 失败: %v", err)
	}

	// 写入一个正常文件
	normalHeader := &tar.Header{
		Name:     "normal.yaml",
		Typeflag: tar.TypeReg,
		Size:     int64(len("normal")),
		Mode:     0644,
	}
	if err := tw.WriteHeader(normalHeader); err != nil {
		t.Fatalf("写入 header 失败: %v", err)
	}
	if _, err := tw.Write([]byte("normal")); err != nil {
		t.Fatalf("写入内容失败: %v", err)
	}

	tw.Close()
	gw.Close()
	f.Close()

	root, err := GetArchiveData(tarGzPath)
	if err != nil {
		t.Fatalf("GetArchiveData 失败: %v", err)
	}

	// 符号链接不应出现在树中
	link := root.SearchPath("evil_link")
	if link != nil {
		t.Error("符号链接条目不应出现在目录树中，但被找到了")
	}

	// 正常文件应存在
	normal := root.SearchPath("normal.yaml")
	if normal == nil {
		t.Error("正常文件应当存在于目录树中，但未找到")
	}
}

// TestIsTarGzFile 测试 tar.gz 文件后缀判断（大小写不敏感）
func TestIsTarGzFile(t *testing.T) {
	cases := []struct {
		path     string
		expected bool
	}{
		{"file.tar.gz", true},
		{"file.tgz", true},
		{"file.TAR.GZ", true},
		{"file.TGZ", true},
		{"file.Tar.Gz", true},
		{"/path/to/plugins.tar.gz", true},
		{"file.zip", false},
		{"file.tar", false},
		{"file.gz", false},
		{"file.txt", false},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			result := isTarGzFile(tc.path)
			if result != tc.expected {
				t.Errorf("isTarGzFile(%q) = %v，期望 %v", tc.path, result, tc.expected)
			}
		})
	}
}

// TestIsPathTraversal 测试路径遍历检测
func TestIsPathTraversal(t *testing.T) {
	cases := []struct {
		path     string
		expected bool
	}{
		{"safe/path/file.txt", false},
		{"../etc/passwd", true},
		{"plugins/../../../etc/shadow", true},
		{"plugins/build/go/file.yaml", false},
		{"..", true},
		{"..hidden/file.txt", false}, // ".." 作为目录名的一部分，不是遍历
		{"path/to/..data", false},    // 同上
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			result := isPathTraversal(tc.path)
			if result != tc.expected {
				t.Errorf("isPathTraversal(%q) = %v，期望 %v", tc.path, result, tc.expected)
			}
		})
	}
}

// TestSanitizePath 测试路径清洗
func TestSanitizePath(t *testing.T) {
	cases := []struct {
		input    string
		expected string
		safe     bool
	}{
		{"plugins/file.yaml", "plugins/file.yaml", true},
		{"/absolute/path.yaml", "absolute/path.yaml", true},
		{"\\windows\\path.yaml", "windows/path.yaml", true},
		{"../../etc/passwd", "", false},
		{"normal/../../escape", "", false},
		{"clean/path/file.txt", "clean/path/file.txt", true},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			result, safe := sanitizePath(tc.input)
			if safe != tc.safe {
				t.Errorf("sanitizePath(%q) safe = %v，期望 %v", tc.input, safe, tc.safe)
			}
			if safe && result != tc.expected {
				t.Errorf("sanitizePath(%q) = %q，期望 %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestGetArchiveData_EmptyTarGz 测试空的 tar.gz 文件
func TestGetArchiveData_EmptyTarGz(t *testing.T) {
	tmpDir := t.TempDir()
	tarGzPath := filepath.Join(tmpDir, "empty.tar.gz")

	f, err := os.Create(tarGzPath)
	if err != nil {
		t.Fatalf("创建文件失败: %v", err)
	}

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)
	tw.Close()
	gw.Close()
	f.Close()

	root, err := GetArchiveData(tarGzPath)
	if err != nil {
		t.Fatalf("解析空 tar.gz 失败: %v", err)
	}

	// 空归档应只有根节点，且无子节点
	if root.Type != "dir" {
		t.Errorf("根节点类型应为 'dir'，实际为: %s", root.Type)
	}

	file := root.NextFile(".yaml")
	if file != nil {
		t.Error("空归档中不应找到任何文件")
	}
}

// TestGetArchiveData_TgzExtension 测试 .tgz 扩展名的支持
func TestGetArchiveData_TgzExtension(t *testing.T) {
	files := []struct {
		name    string
		content string
	}{
		{"test.yaml", "tgz content"},
	}

	// 先创建 tar.gz，再复制为 .tgz
	tarGzPath := createTestTarGz(t, files)
	tgzPath := strings.TrimSuffix(tarGzPath, ".tar.gz") + ".tgz"

	content, err := os.ReadFile(tarGzPath)
	if err != nil {
		t.Fatalf("读取 tar.gz 失败: %v", err)
	}
	if err := os.WriteFile(tgzPath, content, 0644); err != nil {
		t.Fatalf("写入 .tgz 失败: %v", err)
	}

	root, err := GetArchiveData(tgzPath)
	if err != nil {
		t.Fatalf("GetArchiveData 解析 .tgz 失败: %v", err)
	}

	file := root.SearchPath("test.yaml")
	if file == nil {
		t.Error("未找到 test.yaml")
	} else if file.Content.String() != "tgz content" {
		t.Errorf("内容不匹配，期望 'tgz content'，实际为 '%s'", file.Content.String())
	}
}

// TestGetArchiveData_GetDirName 测试 tar.gz 解析后的 GetDirName 方法
func TestGetArchiveData_GetDirName(t *testing.T) {
	files := []struct {
		name    string
		content string
	}{
		{"plugins/build/go/build_go.yaml", "content"},
	}

	tarGzPath := createTestTarGz(t, files)

	root, err := GetArchiveData(tarGzPath)
	if err != nil {
		t.Fatalf("GetArchiveData 失败: %v", err)
	}

	file := root.SearchPath("plugins/build/go/build_go.yaml")
	if file == nil {
		t.Fatal("未找到目标文件")
	}

	// -1 应返回父目录名 "go"
	if dir := file.GetDirName(-1); dir != "go" {
		t.Errorf("GetDirName(-1) = %q，期望 'go'", dir)
	}
	// -2 应返回 "build"
	if dir := file.GetDirName(-2); dir != "build" {
		t.Errorf("GetDirName(-2) = %q，期望 'build'", dir)
	}
	// -3 应返回 "plugins"
	if dir := file.GetDirName(-3); dir != "plugins" {
		t.Errorf("GetDirName(-3) = %q，期望 'plugins'", dir)
	}
}
