package file

import (
	"archive/zip"
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

	if err := DownloadFile(testZipPath, testURL); err != nil {
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
