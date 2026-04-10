package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestDownloadFile 测试文件下载功能
func TestDownloadFile(t *testing.T) {
	// 测试 URL
	testURL := "https://pkg.zentao.net/gitfox/20260123/plugins.zip"

	// 创建测试目录
	testDir := "./file"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer func() {
		// 清理测试目录
		_ = os.RemoveAll(testDir)
	}()

	t.Run("DownloadToDirectory", func(t *testing.T) {
		// 测试下载到目录（自动使用 URL 中的文件名）
		dirPath := "./file/"

		err := DownloadFile(context.Background(), dirPath, testURL)
		if err != nil {
			t.Errorf("DownloadFile to directory failed: %v", err)
			return
		}

		// 验证文件是否存在
		expectedFile := filepath.Join(dirPath, "plugins.zip")
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Errorf("Expected file %s does not exist", expectedFile)
		} else {
			t.Logf("Successfully downloaded to directory: %s", expectedFile)
			// 检查文件大小
			info, _ := os.Stat(expectedFile)
			t.Logf("Downloaded file size: %d bytes", info.Size())
		}

		// 清理下载的文件
		_ = os.Remove(expectedFile)
	})

	t.Run("DownloadToSpecificFile", func(t *testing.T) {
		// 测试下载到指定文件路径
		filePath := "./file/p.zip"

		err := DownloadFile(context.Background(), filePath, testURL)
		if err != nil {
			t.Errorf("DownloadFile to specific file failed: %v", err)
			return
		}

		// 验证文件是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected file %s does not exist", filePath)
		} else {
			t.Logf("Successfully downloaded to specific file: %s", filePath)
			// 检查文件大小
			info, _ := os.Stat(filePath)
			t.Logf("Downloaded file size: %d bytes", info.Size())
		}

		// 清理下载的文件
		_ = os.Remove(filePath)
	})

	t.Run("DownloadToNonExistentDirectory", func(t *testing.T) {
		// 测试下载到不存在的目录中的文件
		filePath := "./file/subdir/test.zip"

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatalf("Failed to create parent directory: %v", err)
		}

		err := DownloadFile(context.Background(), filePath, testURL)
		if err != nil {
			t.Errorf("DownloadFile to non-existent directory failed: %v", err)
			return
		}

		// 验证文件是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected file %s does not exist", filePath)
		} else {
			t.Logf("Successfully downloaded to subdirectory: %s", filePath)
		}

		// 清理
		_ = os.RemoveAll("./file/subdir")
	})

	t.Run("DownloadOverwriteExistingFile", func(t *testing.T) {
		// 测试覆盖已存在的文件
		filePath := "./file/overwrite.zip"

		// 创建一个已存在的文件
		if err := os.WriteFile(filePath, []byte("old content"), 0644); err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		err := DownloadFile(context.Background(), filePath, testURL)
		if err != nil {
			t.Errorf("DownloadFile overwrite failed: %v", err)
			return
		}

		// 验证文件被覆盖（大小应该大于 "old content"）
		info, err := os.Stat(filePath)
		if err != nil {
			t.Errorf("Failed to stat file: %v", err)
		} else if info.Size() <= 11 {
			t.Errorf("File was not overwritten, size: %d", info.Size())
		} else {
			t.Logf("Successfully overwritten file: %s, new size: %d bytes", filePath, info.Size())
		}

		// 清理
		_ = os.Remove(filePath)
	})
}

// TestDownloadFileErrorCases 测试错误情况
func TestDownloadFileErrorCases(t *testing.T) {
	t.Run("InvalidURL", func(t *testing.T) {
		// 测试无效的 URL
		filePath := "./file/test.zip"
		invalidURL := "http://invalid-url-that-does-not-exist-12345.com/file.zip"

		// 确保目录存在
		_ = os.MkdirAll(filepath.Dir(filePath), 0755)
		defer os.RemoveAll("./file")

		err := DownloadFile(context.Background(), filePath, invalidURL)
		if err == nil {
			t.Error("Expected error for invalid URL, but got nil")
		} else {
			t.Logf("Got expected error for invalid URL: %v", err)
		}
	})

	t.Run("URLWithoutFilename", func(t *testing.T) {
		// 测试无法提取文件名的 URL（下载到目录时）
		dirPath := "./file/"
		urlWithoutFilename := "https://example.com/"

		// 确保目录存在
		_ = os.MkdirAll(dirPath, 0755)
		defer os.RemoveAll("./file")

		err := DownloadFile(context.Background(), dirPath, urlWithoutFilename)
		if err == nil {
			t.Error("Expected error for URL without filename, but got nil")
		} else if err.Error() != "cannot extract filename from url" {
			t.Errorf("Expected 'cannot extract filename from url' error, got: %v", err)
		} else {
			t.Logf("Got expected error for URL without filename: %v", err)
		}
	})

	t.Run("InvalidDirectory", func(t *testing.T) {
		// 测试下载到无效的目录路径（父目录不存在）
		filePath := "./nonexistent/subdir/file.zip"
		testURL := "https://pkg.zentao.net/gitfox/20260123/plugins.zip"

		err := DownloadFile(context.Background(), filePath, testURL)
		if err == nil {
			t.Error("Expected error for invalid directory, but got nil")
			// 清理可能创建的文件
			_ = os.RemoveAll("./nonexistent")
		} else {
			t.Logf("Got expected error for invalid directory: %v", err)
		}
	})
}

// TestCreateTempFile 测试创建临时文件功能
func TestCreateTempFile(t *testing.T) {
	t.Run("CreateTempFileSuccessfully", func(t *testing.T) {
		// 测试成功创建临时文件
		f, removeFunc, err := CreateTempFile("test-*.txt")
		if err != nil {
			t.Errorf("CreateTempFile failed: %v", err)
			return
		}

		// 验证文件对象不为空
		if f == nil {
			t.Error("Expected file to be non-nil")
			return
		}

		// 验证 removeFunc 不为空
		if removeFunc == nil {
			t.Error("Expected removeFunc to be non-nil")
			return
		}

		// 获取文件路径
		filePath := f.Name()
		t.Logf("Created temp file: %s", filePath)

		// 验证文件存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected temp file %s to exist, but it doesn't", filePath)
			return
		}

		// 写入一些数据
		testData := []byte("test content")
		n, err := f.Write(testData)
		if err != nil {
			t.Errorf("Failed to write to temp file: %v", err)
		}
		if n != len(testData) {
			t.Errorf("Expected to write %d bytes, wrote %d", len(testData), n)
		}

		// 关闭文件
		if err := f.Close(); err != nil {
			t.Errorf("Failed to close file: %v", err)
		}

		// 验证文件内容
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Failed to read temp file: %v", err)
		}
		if string(content) != string(testData) {
			t.Errorf("Expected file content '%s', got '%s'", testData, content)
		}

		// 调用 removeFunc 删除文件
		err = removeFunc()
		if err != nil {
			t.Errorf("removeFunc failed: %v", err)
			return
		}

		// 验证文件已被删除
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			t.Errorf("Expected temp file %s to be deleted, but it still exists", filePath)
		} else {
			t.Logf("Temp file successfully deleted: %s", filePath)
		}
	})

	t.Run("CreateMultipleTempFiles", func(t *testing.T) {
		// 测试创建多个临时文件
		files := []*os.File{}
		removeFuncs := []func() error{}

		// 创建 3 个临时文件
		for i := 0; i < 3; i++ {
			f, removeFunc, err := CreateTempFile("multi-test-*.dat")
			if err != nil {
				t.Errorf("CreateTempFile failed for file %d: %v", i, err)
				continue
			}

			files = append(files, f)
			removeFuncs = append(removeFuncs, removeFunc)

			t.Logf("Created temp file %d: %s", i+1, f.Name())
		}

		// 验证所有文件都存在
		for i, f := range files {
			if _, err := os.Stat(f.Name()); os.IsNotExist(err) {
				t.Errorf("Expected temp file %d to exist: %s", i, f.Name())
			}
			_ = f.Close()
		}

		// 清理所有文件
		for i, removeFunc := range removeFuncs {
			if err := removeFunc(); err != nil {
				t.Errorf("Failed to remove temp file %d: %v", i, err)
			}
		}

		// 验证所有文件都被删除
		for i, f := range files {
			if _, err := os.Stat(f.Name()); !os.IsNotExist(err) {
				t.Errorf("Expected temp file %d to be deleted: %s", i, f.Name())
			}
		}

		t.Logf("Successfully created and removed %d temp files", len(files))
	})

	t.Run("RemoveFuncMultipleTimes", func(t *testing.T) {
		// 测试多次调用 removeFunc
		f, removeFunc, err := CreateTempFile("remove-test-*.tmp")
		if err != nil {
			t.Errorf("CreateTempFile failed: %v", err)
			return
		}

		filePath := f.Name()
		t.Logf("Created temp file: %s", filePath)

		// 关闭文件
		_ = f.Close()

		// 第一次删除（应该成功）
		err = removeFunc()
		if err != nil {
			t.Errorf("First removeFunc call failed: %v", err)
		}

		// 第二次删除（文件已不存在，应该返回错误但不会 panic）
		err = removeFunc()
		if err == nil {
			t.Log("Second removeFunc call succeeded (file already deleted)")
		} else {
			t.Logf("Second removeFunc call returned expected error: %v", err)
		}

		// 验证文件不存在
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			t.Errorf("Expected temp file to be deleted: %s", filePath)
		}
	})

	t.Run("CreateTempFileWithDifferentPatterns", func(t *testing.T) {
		// 测试不同的文件名模板
		patterns := []string{
			"test.zip",
			"plugins-*.zip",
			"data-*.json",
			"*.tmp",
		}

		for _, pattern := range patterns {
			f, removeFunc, err := CreateTempFile(pattern)
			if err != nil {
				t.Errorf("CreateTempFile failed for pattern '%s': %v", pattern, err)
				continue
			}

			t.Logf("Created temp file with pattern '%s': %s", pattern, f.Name())

			_ = f.Close()
			_ = removeFunc()
		}
	})

	t.Run("WriteAndReadFromTempFile", func(t *testing.T) {
		// 测试写入和读取临时文件
		f, removeFunc, err := CreateTempFile("rw-test-*.txt")
		if err != nil {
			t.Errorf("CreateTempFile failed: %v", err)
			return
		}
		defer removeFunc()

		// 写入多行数据
		testLines := []string{
			"Line 1: Hello World",
			"Line 2: Testing temp file",
			"Line 3: This is a test",
		}

		for _, line := range testLines {
			if _, err := f.WriteString(line + "\n"); err != nil {
				t.Errorf("Failed to write line: %v", err)
			}
		}

		// 关闭文件
		if err := f.Close(); err != nil {
			t.Errorf("Failed to close file: %v", err)
		}

		// 读取文件内容
		content, err := os.ReadFile(f.Name())
		if err != nil {
			t.Errorf("Failed to read temp file: %v", err)
			return
		}

		expectedContent := ""
		for _, line := range testLines {
			expectedContent += line + "\n"
		}

		if string(content) != expectedContent {
			t.Errorf("Expected content:\n%s\nGot:\n%s", expectedContent, string(content))
		} else {
			t.Logf("Successfully wrote and read %d lines from temp file", len(testLines))
		}
	})
}
