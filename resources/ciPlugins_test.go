package resources

import (
	"archive/zip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// createTestZip 创建一个测试用的 ZIP 文件
func createTestZip(t *testing.T) string {
	t.Helper()

	// 创建临时目录
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")

	// 创建 ZIP 文件
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}
	defer zipFile.Close()

	// 创建 ZIP writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 添加一个测试文件到 ZIP 中
	fileWriter, err := zipWriter.Create("test.txt")
	if err != nil {
		t.Fatalf("failed to create file in zip: %v", err)
	}

	_, err = fileWriter.Write([]byte("test content"))
	if err != nil {
		t.Fatalf("failed to write to zip file: %v", err)
	}

	return zipPath
}

// TestLoadAllCIPlugins_EmptyURL 测试空 URL 的情况
func TestLoadAllCIPlugins_EmptyURL(t *testing.T) {
	ctx := context.Background()

	zipReader, cleanup, err := LoadAllCIPlugins(ctx, "")

	if err == nil {
		t.Error("expected error for empty URL, got nil")
	}

	if zipReader != nil {
		t.Error("expected nil zipReader for empty URL")
	}

	if cleanup != nil {
		t.Error("expected nil cleanup for empty URL")
	}

	expectedErrMsg := "plugins url not provided to read schemas from"
	if err != nil && err.Error() != expectedErrMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

// TestLoadAllCIPlugins_LocalFile 测试本地 ZIP 文件的情况
func TestLoadAllCIPlugins_LocalFile(t *testing.T) {
	ctx := context.Background()

	// 创建测试 ZIP 文件
	zipPath := createTestZip(t)

	// 加载 ZIP 文件
	zipReader, cleanup, err := LoadAllCIPlugins(ctx, zipPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if zipReader == nil {
		t.Fatal("expected non-nil zipReader")
	}

	if cleanup == nil {
		t.Fatal("expected non-nil cleanup function")
	}

	// 验证 ZIP 文件可以读取
	if len(zipReader.File) == 0 {
		t.Error("expected at least one file in zip")
	}

	// 调用清理函数
	cleanup()
}

// TestLoadAllCIPlugins_RemoteFile 测试远程 ZIP 文件的情况
func TestLoadAllCIPlugins_RemoteFile(t *testing.T) {
	ctx := context.Background()

	// 创建测试 ZIP 文件
	zipPath := createTestZip(t)

	// 读取 ZIP 文件内容
	zipContent, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("failed to read zip file: %v", err)
	}

	// 创建模拟 HTTP 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(zipContent)
	}))
	defer server.Close()

	// 添加文件名路径到 URL
	testURL := server.URL + "/plugins.zip"

	// 加载远程 ZIP 文件
	zipReader, cleanup, err := LoadAllCIPlugins(ctx, testURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if zipReader == nil {
		t.Fatal("expected non-nil zipReader")
	}

	if cleanup == nil {
		t.Fatal("expected non-nil cleanup function")
	}

	// 验证 ZIP 文件可以读取
	if len(zipReader.File) == 0 {
		t.Error("expected at least one file in zip")
	}

	// 验证文件内容
	file := zipReader.File[0]
	rc, err := file.Open()
	if err != nil {
		t.Fatalf("failed to open file in zip: %v", err)
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("failed to read file content: %v", err)
	}

	expectedContent := "test content"
	if string(content) != expectedContent {
		t.Errorf("expected content '%s', got '%s'", expectedContent, string(content))
	}

	// 调用清理函数
	cleanup()
}

// TestLoadAllCIPlugins_InvalidZipFile 测试无效的 ZIP 文件
func TestLoadAllCIPlugins_InvalidZipFile(t *testing.T) {
	ctx := context.Background()

	// 创建一个无效的 ZIP 文件（实际上只是一个普通文本文件）
	tmpDir := t.TempDir()
	invalidZipPath := filepath.Join(tmpDir, "invalid.zip")

	err := os.WriteFile(invalidZipPath, []byte("not a zip file"), 0644)
	if err != nil {
		t.Fatalf("failed to create invalid zip file: %v", err)
	}

	// 尝试加载无效的 ZIP 文件
	zipReader, cleanup, err := LoadAllCIPlugins(ctx, invalidZipPath)

	if err == nil {
		t.Error("expected error for invalid zip file, got nil")
	}

	if zipReader != nil {
		t.Error("expected nil zipReader for invalid zip file")
	}

	if cleanup != nil {
		cleanup()
	}
}

// TestLoadAllCIPlugins_NonExistentRemoteFile 测试不存在的远程文件
func TestLoadAllCIPlugins_NonExistentRemoteFile(t *testing.T) {
	ctx := context.Background()

	// 创建模拟 HTTP 服务器返回 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// 添加文件名路径到 URL
	testURL := server.URL + "/nonexistent.zip"

	// 尝试加载不存在的远程文件
	zipReader, cleanup, err := LoadAllCIPlugins(ctx, testURL)

	if err == nil {
		t.Error("expected error for non-existent remote file, got nil")
	}

	if zipReader != nil {
		t.Error("expected nil zipReader for non-existent remote file")
	}

	if cleanup != nil {
		cleanup()
	}
}

// TestLoadAllCIPlugins_CleanupDeletesTempFile 测试清理函数是否删除临时文件
func TestLoadAllCIPlugins_CleanupDeletesTempFile(t *testing.T) {
	ctx := context.Background()

	// 创建测试 ZIP 文件
	zipPath := createTestZip(t)

	// 读取 ZIP 文件内容
	zipContent, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("failed to read zip file: %v", err)
	}

	// 创建模拟 HTTP 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(zipContent)
	}))
	defer server.Close()

	// 添加文件名路径到 URL
	testURL := server.URL + "/plugins.zip"

	// 记录当前临时目录中的 plugins_*.zip 文件
	beforeFiles, _ := filepath.Glob(filepath.Join(os.TempDir(), "plugins_*.zip"))

	// 加载远程 ZIP 文件
	zipReader, cleanup, err := LoadAllCIPlugins(ctx, testURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cleanup == nil {
		t.Fatal("expected non-nil cleanup function")
	}

	// 查找新创建的临时文件（带时间戳的 plugins_*.zip）
	afterFiles, err := filepath.Glob(filepath.Join(os.TempDir(), "plugins_*.zip"))
	if err != nil {
		t.Fatalf("failed to list temp files: %v", err)
	}

	// 找出新增的文件
	var newFile string
	for _, f := range afterFiles {
		found := false
		for _, bf := range beforeFiles {
			if f == bf {
				found = true
				break
			}
		}
		if !found {
			newFile = f
			break
		}
	}

	if newFile == "" {
		t.Fatal("expected a new temp file to be created")
	}

	// 验证文件存在
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Fatal("expected temp file to exist before cleanup")
	}

	// 调用清理函数
	cleanup()

	// 验证临时文件已被删除
	if _, err := os.Stat(newFile); !os.IsNotExist(err) {
		t.Error("expected temp file to be deleted after cleanup")
	}

	// 验证 zipReader 已关闭（尝试访问会失败）
	if zipReader != nil && len(zipReader.File) > 0 {
		_, err := zipReader.File[0].Open()
		if err == nil {
			t.Error("expected zipReader to be closed after cleanup")
		}
	}
}
