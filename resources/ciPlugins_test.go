// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package resources

import (
	"archive/zip"
	"context"
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

// TestDownloadCIPlugins_EmptyURL 测试空 URL 的情况
func TestDownloadCIPlugins_EmptyURL(t *testing.T) {
	ctx := context.Background()

	path, cleanup, err := DownloadCIPlugins(ctx, "")
	if err == nil {
		t.Error("expected error for empty URL, got nil")
	}
	if path != "" {
		t.Error("expected empty string for empty URL")
	}
	if cleanup != nil {
		t.Error("expected nil cleanup for empty URL")
	}
}

// TestDownloadCIPlugins_LocalFile 测试本地 ZIP 文件的情况
func TestDownloadCIPlugins_LocalFile(t *testing.T) {
	ctx := context.Background()
	zipPath := createTestZip(t)

	path, cleanup, err := DownloadCIPlugins(ctx, zipPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}
	if cleanup == nil {
		t.Fatal("expected non-nil cleanup function")
	}

	cleanup()
}

// TestDownloadCIPlugins_RemoteFile 测试远程 ZIP 文件的情况
func TestDownloadCIPlugins_RemoteFile(t *testing.T) {
	ctx := context.Background()
	zipPath := createTestZip(t)

	zipContent, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("failed to read zip file: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(zipContent)
	}))
	defer server.Close()

	testURL := server.URL + "/plugins.zip"

	path, cleanup, err := DownloadCIPlugins(ctx, testURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}
	if cleanup == nil {
		t.Fatal("expected non-nil cleanup function")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("downloaded file does not exist")
	}

	cleanup()
}

// TestDownloadCIPlugins_NonExistentRemoteFile 测试不存在的远程文件
func TestDownloadCIPlugins_NonExistentRemoteFile(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	testURL := server.URL + "/nonexistent.zip"

	path, cleanup, err := DownloadCIPlugins(ctx, testURL)
	if err == nil {
		t.Error("expected error for non-existent remote file, got nil")
	}
	if path != "" {
		t.Error("expected empty string for non-existent remote file")
	}
	if cleanup != nil {
		cleanup()
	}
}

// TestDownloadCIPlugins_CleanupDeletesTempFile 测试清理函数是否删除临时文件
func TestDownloadCIPlugins_CleanupDeletesTempFile(t *testing.T) {
	ctx := context.Background()
	zipPath := createTestZip(t)

	zipContent, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("failed to read zip file: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(zipContent)
	}))
	defer server.Close()

	testURL := server.URL + "/plugins.zip"

	path, cleanup, err := DownloadCIPlugins(ctx, testURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cleanup == nil {
		t.Fatal("expected non-nil cleanup function")
	}

	// 验证文件存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected temp file to exist before cleanup")
	}

	// 调用清理函数
	cleanup()

	// 验证临时文件已被删除
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected temp file to be deleted after cleanup")
	}
}
