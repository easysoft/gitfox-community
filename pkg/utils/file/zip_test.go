// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package file

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestReadZipFile 测试读取 zip 文件功能
func TestReadZipFile(t *testing.T) {
	// 准备测试环境：下载一个真实的 zip 文件
	testDir := "./test_zip"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer func() {
		// 清理测试目录
		_ = os.RemoveAll(testDir)
	}()

	// 下载测试用的 zip 文件
	testZipPath := filepath.Join(testDir, "test.zip")
	testURL := "https://pkg.zentao.net/gitfox/20260123/plugins.zip"

	if err := DownloadFile(context.Background(), testZipPath, testURL); err != nil {
		t.Fatalf("Failed to download test zip file: %v", err)
	}

	t.Run("ReadValidZipFile", func(t *testing.T) {
		// 测试读取有效的 zip 文件
		zipReader, closeFunc, err := ReadZipFile(testZipPath)
		if err != nil {
			t.Errorf("ReadZipFile failed: %v", err)
			return
		}

		// 确保 zipReader 不为空
		if zipReader == nil {
			t.Error("Expected zipReader to be non-nil")
			return
		}

		// 确保 closeFunc 不为空
		if closeFunc == nil {
			t.Error("Expected closeFunc to be non-nil")
			return
		}

		// 验证可以读取 zip 文件内容
		if len(zipReader.File) == 0 {
			t.Error("Expected zip file to contain files, but got 0")
		} else {
			t.Logf("Zip file contains %d files", len(zipReader.File))
			// 列出前几个文件
			for i, f := range zipReader.File {
				if i >= 3 {
					break
				}
				t.Logf("  File %d: %s (size: %d bytes)", i+1, f.Name, f.UncompressedSize64)
			}
		}

		// 测试 closeFunc 可以正常调用
		closeFunc()
		t.Log("Successfully called closeFunc")

		// 验证关闭后不能再读取（这会导致错误）
		if len(zipReader.File) > 0 {
			// 只测试第一个文件
			rc, err := zipReader.File[0].Open()
			if err == nil {
				_ = rc.Close()
				t.Error("Expected error when opening file after close, but got nil")
			} else {
				t.Logf("Got expected error after close: %v", err)
			}
		}
	})

	t.Run("ReadValidZipFileAndListContents", func(t *testing.T) {
		// 测试读取并遍历 zip 文件内容
		zipReader, closeFunc, err := ReadZipFile(testZipPath)
		if err != nil {
			t.Errorf("ReadZipFile failed: %v", err)
			return
		}
		defer closeFunc()

		// 统计文件和目录数量
		fileCount := 0
		dirCount := 0
		for _, f := range zipReader.File {
			if f.FileInfo().IsDir() {
				dirCount++
			} else {
				fileCount++
			}
		}

		t.Logf("Zip contains %d files and %d directories", fileCount, dirCount)

		if fileCount == 0 && dirCount == 0 {
			t.Error("Expected zip file to contain files or directories")
		}
	})

	t.Run("CloseMultipleTimes", func(t *testing.T) {
		// 测试多次调用 closeFunc 不会出错
		zipReader, closeFunc, err := ReadZipFile(testZipPath)
		if err != nil {
			t.Errorf("ReadZipFile failed: %v", err)
			return
		}

		if zipReader == nil || closeFunc == nil {
			t.Error("Expected valid zipReader and closeFunc")
			return
		}

		// 多次调用 closeFunc
		closeFunc()
		closeFunc()
		closeFunc()

		t.Log("Successfully called closeFunc multiple times")
	})
}

// TestReadZipFileErrorCases 测试错误情况
func TestReadZipFileErrorCases(t *testing.T) {
	t.Run("NotAZipFile", func(t *testing.T) {
		// 测试非 zip 文件
		testDir := "./test_zip_error"
		_ = os.MkdirAll(testDir, 0755)
		defer os.RemoveAll(testDir)

		// 创建一个非 zip 文件
		txtFile := filepath.Join(testDir, "test.txt")
		if err := os.WriteFile(txtFile, []byte("not a zip file"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		zipReader, closeFunc, err := ReadZipFile(txtFile)
		if err == nil {
			t.Error("Expected error for non-zip file, but got nil")
			if closeFunc != nil {
				closeFunc()
			}
			return
		}

		if err.Error() != "not a zip file" {
			t.Errorf("Expected 'not a zip file' error, got: %v", err)
		} else {
			t.Logf("Got expected error: %v", err)
		}

		if zipReader != nil {
			t.Error("Expected zipReader to be nil for error case")
		}

		if closeFunc != nil {
			t.Error("Expected closeFunc to be nil for error case")
		}
	})

	t.Run("FileWithZipExtensionButInvalidContent", func(t *testing.T) {
		// 测试扩展名是 .zip 但内容无效的文件
		testDir := "./test_zip_error"
		_ = os.MkdirAll(testDir, 0755)
		defer os.RemoveAll(testDir)

		fakeZipFile := filepath.Join(testDir, "fake.zip")
		if err := os.WriteFile(fakeZipFile, []byte("this is not valid zip content"), 0644); err != nil {
			t.Fatalf("Failed to create fake zip file: %v", err)
		}

		zipReader, closeFunc, err := ReadZipFile(fakeZipFile)
		if err == nil {
			t.Error("Expected error for invalid zip content, but got nil")
			if closeFunc != nil {
				closeFunc()
			}
			return
		}

		t.Logf("Got expected error for invalid zip content: %v", err)

		if zipReader != nil {
			t.Error("Expected zipReader to be nil for error case")
		}

		if closeFunc != nil {
			t.Error("Expected closeFunc to be nil for error case")
		}
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		// 测试不存在的文件
		nonExistentFile := "./nonexistent/file.zip"

		zipReader, closeFunc, err := ReadZipFile(nonExistentFile)
		if err == nil {
			t.Error("Expected error for non-existent file, but got nil")
			if closeFunc != nil {
				closeFunc()
			}
			return
		}

		t.Logf("Got expected error for non-existent file: %v", err)

		if zipReader != nil {
			t.Error("Expected zipReader to be nil for error case")
		}

		if closeFunc != nil {
			t.Error("Expected closeFunc to be nil for error case")
		}
	})

	t.Run("EmptyZipFile", func(t *testing.T) {
		// 测试空的 zip 文件
		testDir := "./test_zip_error"
		_ = os.MkdirAll(testDir, 0755)
		defer os.RemoveAll(testDir)

		emptyZipPath := filepath.Join(testDir, "empty.zip")

		// 创建一个空的 zip 文件
		zipFile, err := os.Create(emptyZipPath)
		if err != nil {
			t.Fatalf("Failed to create empty zip file: %v", err)
		}

		zipWriter := zip.NewWriter(zipFile)
		if err := zipWriter.Close(); err != nil {
			t.Fatalf("Failed to close zip writer: %v", err)
		}
		if err := zipFile.Close(); err != nil {
			t.Fatalf("Failed to close zip file: %v", err)
		}

		// 读取空的 zip 文件（应该成功）
		zipReader, closeFunc, err := ReadZipFile(emptyZipPath)
		if err != nil {
			t.Errorf("ReadZipFile failed for empty zip: %v", err)
			return
		}
		defer closeFunc()

		if zipReader == nil {
			t.Error("Expected zipReader to be non-nil for empty zip")
			return
		}

		if len(zipReader.File) != 0 {
			t.Errorf("Expected empty zip to have 0 files, got: %d", len(zipReader.File))
		} else {
			t.Log("Successfully read empty zip file with 0 files")
		}
	})
}

// TestIsZipFile 测试 isZipFile 辅助函数（间接测试）
func TestIsZipFileValidation(t *testing.T) {
	testCases := []struct {
		name       string
		path       string
		shouldPass bool
	}{
		{"Valid zip extension", "test.zip", true},
		{"Valid zip with path", "/path/to/file.zip", true},
		{"No extension", "testfile", false},
		{"Different extension", "test.txt", false},
		{"Different extension", "test.tar.gz", false},
		{"Case sensitive", "test.ZIP", false}, // isZipFile 是大小写敏感的
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建临时文件
			testDir := "./test_zip_validation"
			_ = os.MkdirAll(testDir, 0755)
			defer os.RemoveAll(testDir)

			filename := filepath.Join(testDir, filepath.Base(tc.path))
			if err := os.WriteFile(filename, []byte("dummy content"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			_, _, err := ReadZipFile(filename)

			if tc.shouldPass {
				// 应该通过扩展名检查（但可能因为内容无效而失败）
				// 我们只检查错误不是 "not a zip file"
				if err != nil && err.Error() == "not a zip file" {
					t.Errorf("Expected to pass extension check for %s, but got 'not a zip file' error", tc.path)
				}
			} else {
				// 应该在扩展名检查时失败
				if err == nil || err.Error() != "not a zip file" {
					t.Errorf("Expected 'not a zip file' error for %s, got: %v", tc.path, err)
				} else {
					t.Logf("Got expected 'not a zip file' error for %s", tc.path)
				}
			}
		})
	}
}

// TestGetZipData 测试解析 zip 文件为数据结构
func TestGetZipData(t *testing.T) {
	// 创建测试用的 zip 文件
	testDir := "./test_zipdata"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// 创建一个包含目录和文件的 zip
	testZipPath := filepath.Join(testDir, "test.zip")
	zipFile, err := os.Create(testZipPath)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}

	zipWriter := zip.NewWriter(zipFile)

	// 添加测试文件和目录结构
	// plugins/build/go/build_go.yaml
	// plugins/build/go/build_go.json
	// plugins/build/python/build_py.yaml
	// plugins/deploy/deploy.yaml
	// readme.txt
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

	for _, f := range files {
		writer, err := zipWriter.Create(f.name)
		if err != nil {
			t.Fatalf("Failed to create file in zip: %v", err)
		}
		_, err = writer.Write([]byte(f.content))
		if err != nil {
			t.Fatalf("Failed to write file content: %v", err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}
	if err := zipFile.Close(); err != nil {
		t.Fatalf("Failed to close zip file: %v", err)
	}

	// 读取 zip 文件
	zipReader, closeFunc, err := ReadZipFile(testZipPath)
	if err != nil {
		t.Fatalf("Failed to read zip file: %v", err)
	}
	defer closeFunc()

	// 解析为数据结构
	root, err := GetZipData(zipReader)
	if err != nil {
		t.Fatalf("Failed to get zip data: %v", err)
	}

	// 测试基本属性
	if root.Type != "dir" {
		t.Errorf("Expected root type to be 'dir', got: %s", root.Type)
	}
	if root.Name != "root" {
		t.Errorf("Expected root name to be 'root', got: %s", root.Name)
	}
	if root.Path != "" {
		t.Errorf("Expected root path to be empty, got: %s", root.Path)
	}

	t.Logf("Successfully created ZipData structure with root")

	// 测试 SearchPath
	t.Run("SearchPath", func(t *testing.T) {
		// 测试查找存在的文件
		file1 := root.SearchPath("plugins/build/go/build_go.yaml")
		if file1 == nil {
			t.Error("Expected to find plugins/build/go/build_go.yaml, but got nil")
		} else {
			if file1.Type != "file" {
				t.Errorf("Expected type 'file', got: %s", file1.Type)
			}
			if file1.Name != "build_go.yaml" {
				t.Errorf("Expected name 'build_go.yaml', got: %s", file1.Name)
			}
			if file1.Suffix != ".yaml" {
				t.Errorf("Expected suffix '.yaml', got: %s", file1.Suffix)
			}
			if file1.Content.String() != "yaml content 1" {
				t.Errorf("Expected content 'yaml content 1', got: %s", file1.Content.String())
			}
			t.Logf("Found file: %s (Path: %s, id: %d)", file1.Name, file1.Path, file1.id)
		}

		// 测试查找目录
		dir1 := root.SearchPath("plugins/build")
		if dir1 == nil {
			t.Error("Expected to find plugins/build directory, but got nil")
		} else {
			if dir1.Type != "dir" {
				t.Errorf("Expected type 'dir', got: %s", dir1.Type)
			}
			if dir1.Name != "build" {
				t.Errorf("Expected name 'build', got: %s", dir1.Name)
			}
			t.Logf("Found directory: %s (Path: %s)", dir1.Name, dir1.Path)
		}

		// 测试查找不存在的路径
		notFound := root.SearchPath("nonexistent/path")
		if notFound != nil {
			t.Error("Expected nil for non-existent path, but got result")
		}
	})

	// 测试 NextFile
	t.Run("NextFile", func(t *testing.T) {
		// 查找所有 .yaml 文件
		var yamlFiles []*ZipData
		for {
			file := root.NextFile(".yaml")
			if file == nil {
				break
			}
			yamlFiles = append(yamlFiles, file)
			t.Logf("Found yaml file: %s (Path: %s, id: %d)", file.Name, file.Path, file.id)
		}

		// 应该找到 3 个 .yaml 文件
		expectedCount := 3
		if len(yamlFiles) != expectedCount {
			t.Errorf("Expected to find %d yaml files, but got %d", expectedCount, len(yamlFiles))
		}

		// 验证不重复
		seenIDs := make(map[int]bool)
		for _, file := range yamlFiles {
			if seenIDs[file.id] {
				t.Errorf("Found duplicate file with id: %d", file.id)
			}
			seenIDs[file.id] = true
		}

		// 再次调用应该返回 nil（所有文件都已返回）
		file := root.NextFile(".yaml")
		if file != nil {
			t.Error("Expected nil after all yaml files returned, but got result")
		}

		// 检查 Read 记录
		if root.Read.Suffix != ".yaml" {
			t.Errorf("Expected Read.Suffix to be '.yaml', got: %s", root.Read.Suffix)
		}
		if len(root.Read.IDs) != expectedCount {
			t.Errorf("Expected Read.IDs to have %d entries, got: %d", expectedCount, len(root.Read.IDs))
		}
		t.Logf("Read.IDs: %v", root.Read.IDs)
	})

	// 测试 CleanRead
	t.Run("CleanRead", func(t *testing.T) {
		// 先确保 Read 不为空（从上一个测试继续）
		if len(root.Read.IDs) == 0 {
			t.Skip("Read.IDs is empty, skipping CleanRead test")
		}

		// 清空 Read
		root.CleanRead()

		// 验证已清空
		if len(root.Read.IDs) != 0 {
			t.Errorf("Expected Read.IDs to be empty after CleanRead, got length: %d", len(root.Read.IDs))
		}
		if root.Read.Suffix != "" {
			t.Errorf("Expected Read.Suffix to be empty after CleanRead, got: %s", root.Read.Suffix)
		}

		// 应该可以再次遍历
		file := root.NextFile(".yaml")
		if file == nil {
			t.Error("Expected to find yaml file after CleanRead, but got nil")
		} else {
			t.Logf("Successfully found yaml file after CleanRead: %s", file.Path)
		}
	})

	// 测试在子目录上使用 NextFile
	t.Run("NextFileOnSubdirectory", func(t *testing.T) {
		// 查找 plugins/build 目录
		buildDir := root.SearchPath("plugins/build")
		if buildDir == nil {
			t.Fatal("Failed to find plugins/build directory")
		}

		// 在子目录上查找 .yaml 文件
		file1 := buildDir.NextFile(".yaml")
		if file1 == nil {
			t.Error("Expected to find yaml file in build directory")
		} else {
			t.Logf("Found yaml file in build directory: %s (Path: %s)", file1.Name, file1.Path)
		}

		file2 := buildDir.NextFile(".yaml")
		if file2 == nil {
			t.Error("Expected to find second yaml file in build directory")
		} else {
			t.Logf("Found second yaml file in build directory: %s (Path: %s)", file2.Name, file2.Path)
		}

		// 验证不重复
		if file1.id == file2.id {
			t.Error("Found duplicate files with same id")
		}

		// 验证 buildDir 的 Read 记录
		if len(buildDir.Read.IDs) != 2 {
			t.Errorf("Expected buildDir.Read.IDs to have 2 entries, got: %d", len(buildDir.Read.IDs))
		}
	})

	// 测试混合后缀的情况
	t.Run("NextFileWithDifferentSuffixes", func(t *testing.T) {
		// 清空 root 的 Read
		root.CleanRead()

		// 先获取一个 .yaml 文件
		yaml1 := root.NextFile(".yaml")
		if yaml1 == nil {
			t.Fatal("Expected to find first yaml file")
		}
		t.Logf("Found yaml file: %s (id: %d)", yaml1.Path, yaml1.id)

		// 再获取一个 .json 文件
		json1 := root.NextFile(".json")
		if json1 == nil {
			t.Fatal("Expected to find first json file")
		}
		t.Logf("Found json file: %s (id: %d)", json1.Path, json1.id)

		// 再获取一个 .yaml 文件（应该是不同于第一个的）
		yaml2 := root.NextFile(".yaml")
		if yaml2 == nil {
			t.Fatal("Expected to find second yaml file")
		}
		t.Logf("Found second yaml file: %s (id: %d)", yaml2.Path, yaml2.id)

		// 验证不重复
		if yaml1.id == yaml2.id {
			t.Error("Got duplicate yaml files with same id")
		}
		if yaml1.id == json1.id || yaml2.id == json1.id {
			t.Error("yaml and json files should have different ids")
		}

		// 检查 Read 状态
		t.Logf("Read.IDs: %v, Read.Suffix: %s", root.Read.IDs, root.Read.Suffix)

		// 验证 Read.Suffix 是最后一次调用的后缀
		if root.Read.Suffix != ".yaml" {
			t.Errorf("Expected Read.Suffix to be '.yaml', got: %s", root.Read.Suffix)
		}

		// 这里有一个疑问：Read.IDs 应该包含所有返回过的文件id，还是只包含当前后缀的文件id？
		// 当前实现是包含所有后缀的文件id
		expectedIDCount := 3 // yaml1, json1, yaml2
		if len(root.Read.IDs) != expectedIDCount {
			t.Logf("Warning: Read.IDs contains %d entries, expected %d", len(root.Read.IDs), expectedIDCount)
			t.Logf("This might indicate that different suffixes should maintain separate read states")
		}
	})

	// 测试不同后缀的遍历是否互相干扰
	t.Run("NextFileSuffixIndependence", func(t *testing.T) {
		// 清空 root 的 Read
		root.CleanRead()

		// 遍历所有 .json 文件
		var jsonFiles []*ZipData
		for {
			file := root.NextFile(".json")
			if file == nil {
				break
			}
			jsonFiles = append(jsonFiles, file)
			t.Logf("Found json file: %s", file.Path)
		}
		t.Logf("Total json files: %d", len(jsonFiles))

		// 现在尝试遍历 .yaml 文件
		// 问题：.yaml 文件的 id 可能已经在 Read.IDs 中了吗？
		var yamlFiles []*ZipData
		for {
			file := root.NextFile(".yaml")
			if file == nil {
				break
			}
			yamlFiles = append(yamlFiles, file)
			t.Logf("Found yaml file: %s", file.Path)
		}
		t.Logf("Total yaml files: %d", len(yamlFiles))

		// 验证是否找到了预期数量的文件
		expectedYamlCount := 3
		if len(yamlFiles) != expectedYamlCount {
			t.Errorf("Expected to find %d yaml files, but got %d", expectedYamlCount, len(yamlFiles))
			t.Errorf("This indicates that different suffixes interfere with each other")
			t.Logf("Read.IDs: %v", root.Read.IDs)
		}
	})

	// 测试边界情况
	t.Run("EdgeCases", func(t *testing.T) {
		root.CleanRead()

		// 查找不存在的后缀
		notFound := root.NextFile(".xml")
		if notFound != nil {
			t.Error("Expected nil for non-existent suffix, but got result")
		}

		// 验证 Read 状态（即使没找到文件，Suffix 也应该被设置吗？）
		// 当前实现：只有找到文件时才设置 Suffix
		if root.Read.Suffix != "" {
			t.Logf("Note: Read.Suffix is set to '%s' even though no file was found", root.Read.Suffix)
		}

		// 在空目录上调用 NextFile
		emptyDir := &ZipData{
			Type: "dir",
			Name: "empty",
			Next: make(map[string]*ZipData),
		}
		file := emptyDir.NextFile(".yaml")
		if file != nil {
			t.Error("Expected nil for empty directory, but got result")
		}

		// 在文件节点上调用 NextFile（不是目录）
		yamlFile := root.SearchPath("plugins/build/go/build_go.yaml")
		if yamlFile != nil {
			result := yamlFile.NextFile(".yaml")
			// 由于 yamlFile 本身是文件且后缀匹配，应该返回自己（如果没有在 Read.IDs 中）
			if result == nil {
				t.Error("Expected file node to return itself when calling NextFile with matching suffix")
			} else if result != yamlFile {
				t.Error("Expected file node to return itself, but got different result")
			}
		}
	})
}

// TestGetDirName 测试 GetDirName 方法
func TestGetDirName(t *testing.T) {
	// 创建测试用的 zip 文件
	testDir := "./test_getdirname"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// 创建一个包含多级目录的 zip
	testZipPath := filepath.Join(testDir, "test.zip")
	zipFile, err := os.Create(testZipPath)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}

	zipWriter := zip.NewWriter(zipFile)

	// 添加测试文件，模拟真实的插件目录结构
	files := []struct {
		name    string
		content string
	}{
		{"plugins/build/go/build_go.yaml", "go plugin yaml"},
		{"plugins/build/go/build_go.json", "go plugin json"},
		{"plugins/build/docker_build/build_docker_build.yaml", "docker build yaml"},
		{"plugins/build/maven/build_maven.yaml", "maven plugin yaml"},
		{"plugins/deploy/k8s/deploy_k8s.yaml", "k8s deploy yaml"},
		{"plugins/test/unit/test_unit.yaml", "unit test yaml"},
		{"readme.txt", "readme content"},
	}

	for _, f := range files {
		writer, err := zipWriter.Create(f.name)
		if err != nil {
			t.Fatalf("Failed to create file in zip: %v", err)
		}
		_, err = writer.Write([]byte(f.content))
		if err != nil {
			t.Fatalf("Failed to write file content: %v", err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}
	if err := zipFile.Close(); err != nil {
		t.Fatalf("Failed to close zip file: %v", err)
	}

	// 读取 zip 文件并解析
	zipReader, closeFunc, err := ReadZipFile(testZipPath)
	if err != nil {
		t.Fatalf("Failed to read zip file: %v", err)
	}
	defer closeFunc()

	root, err := GetZipData(zipReader)
	if err != nil {
		t.Fatalf("Failed to get zip data: %v", err)
	}

	// 测试用例
	testCases := []struct {
		name        string
		path        string
		index       int
		expectedDir string
		description string
	}{
		{
			name:        "GetParentDir_-1",
			path:        "plugins/build/go/build_go.yaml",
			index:       -1,
			expectedDir: "go",
			description: "Get parent directory (文件的父目录)",
		},
		{
			name:        "GetGrandParentDir_-2",
			path:        "plugins/build/go/build_go.yaml",
			index:       -2,
			expectedDir: "build",
			description: "Get grandparent directory",
		},
		{
			name:        "GetRootDir_-3",
			path:        "plugins/build/go/build_go.yaml",
			index:       -3,
			expectedDir: "plugins",
			description: "Get root plugins directory",
		},
		{
			name:        "OutOfRange_-4",
			path:        "plugins/build/go/build_go.yaml",
			index:       -4,
			expectedDir: "",
			description: "Index out of range (beyond root)",
		},
		{
			name:        "DockerBuild_-1",
			path:        "plugins/build/docker_build/build_docker_build.yaml",
			index:       -1,
			expectedDir: "docker_build",
			description: "Get parent directory for docker_build",
		},
		{
			name:        "DeployK8s_-1",
			path:        "plugins/deploy/k8s/deploy_k8s.yaml",
			index:       -1,
			expectedDir: "k8s",
			description: "Get parent directory for deploy/k8s",
		},
		{
			name:        "DeployK8s_-2",
			path:        "plugins/deploy/k8s/deploy_k8s.yaml",
			index:       -2,
			expectedDir: "deploy",
			description: "Get deploy directory",
		},
		{
			name:        "RootFile_-1",
			path:        "readme.txt",
			index:       -1,
			expectedDir: "",
			description: "File in root directory (no parent)",
		},
		{
			name:        "PositiveIndex",
			path:        "plugins/build/go/build_go.yaml",
			index:       0,
			expectedDir: "",
			description: "Positive index should return empty string",
		},
		{
			name:        "PositiveIndex_1",
			path:        "plugins/build/go/build_go.yaml",
			index:       1,
			expectedDir: "",
			description: "Positive index 1 should return empty string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 查找文件
			file := root.SearchPath(tc.path)
			if file == nil {
				t.Fatalf("Failed to find file: %s", tc.path)
			}

			// 调用 GetDirName
			result := file.GetDirName(tc.index)

			// 验证结果
			if result != tc.expectedDir {
				t.Errorf("GetDirName(%d) for %s\n  Expected: %q\n  Got:      %q\n  Description: %s",
					tc.index, tc.path, tc.expectedDir, result, tc.description)
			} else {
				t.Logf("✓ GetDirName(%d) for %s = %q (%s)",
					tc.index, tc.path, result, tc.description)
			}
		})
	}

	// 测试在循环中使用 GetDirName
	t.Run("GetDirNameInLoop", func(t *testing.T) {
		root.CleanRead()

		// 遍历所有 .yaml 文件并获取其父目录
		var parentDirs []string
		for {
			file := root.NextFile(".yaml")
			if file == nil {
				break
			}

			// 跳过根目录文件
			if file.Path == "readme.txt" {
				continue
			}

			parentDir := file.GetDirName(-1)
			if parentDir != "" {
				parentDirs = append(parentDirs, parentDir)
				t.Logf("File: %s -> Parent Dir: %s", file.Path, parentDir)
			}
		}

		// 验证找到了预期的父目录
		expectedParents := []string{"go", "docker_build", "maven", "k8s", "unit"}
		if len(parentDirs) != len(expectedParents) {
			t.Errorf("Expected %d parent directories, got %d", len(expectedParents), len(parentDirs))
		}

		// 验证包含预期的目录（不验证顺序）
		parentSet := make(map[string]bool)
		for _, p := range parentDirs {
			parentSet[p] = true
		}

		for _, expected := range expectedParents {
			if !parentSet[expected] {
				t.Errorf("Expected parent directory %q not found", expected)
			}
		}
	})

	// 测试获取分组名称的实际场景
	t.Run("GetGroupName", func(t *testing.T) {
		root.CleanRead()

		// 模拟获取插件分组（倒数第二级目录）
		file := root.SearchPath("plugins/build/go/build_go.yaml")
		if file == nil {
			t.Fatal("Failed to find test file")
		}

		// 文件名
		fileName := file.GetDirName(0) // 这应该是空，因为不支持正数
		t.Logf("File name (index 0): %q", fileName)

		// 插件名称（父目录）
		pluginName := file.GetDirName(-1)
		t.Logf("Plugin name (-1): %q", pluginName)
		if pluginName != "go" {
			t.Errorf("Expected plugin name 'go', got: %q", pluginName)
		}

		// 分组名称（build、deploy、test 等）
		groupName := file.GetDirName(-2)
		t.Logf("Group name (-2): %q", groupName)
		if groupName != "build" {
			t.Errorf("Expected group name 'build', got: %q", groupName)
		}

		// 根目录
		rootDir := file.GetDirName(-3)
		t.Logf("Root dir (-3): %q", rootDir)
		if rootDir != "plugins" {
			t.Errorf("Expected root dir 'plugins', got: %q", rootDir)
		}
	})

	// 测试边界情况
	t.Run("EdgeCases", func(t *testing.T) {
		// 空路径
		emptyData := &ZipData{Path: ""}
		result := emptyData.GetDirName(-1)
		if result != "" {
			t.Errorf("Expected empty result for empty path, got: %q", result)
		}

		// 单级路径（只有文件名）
		singleLevel := &ZipData{Path: "file.txt"}
		result = singleLevel.GetDirName(-1)
		if result != "" {
			t.Errorf("Expected empty result for single-level path, got: %q", result)
		}

		// 两级路径
		twoLevel := &ZipData{Path: "dir/file.txt"}
		result = twoLevel.GetDirName(-1)
		if result != "dir" {
			t.Errorf("Expected 'dir' for two-level path, got: %q", result)
		}

		// 大的负数索引
		largeNegative := &ZipData{Path: "a/b/c/d.txt"}
		result = largeNegative.GetDirName(-100)
		if result != "" {
			t.Errorf("Expected empty result for large negative index, got: %q", result)
		}
	})
}
