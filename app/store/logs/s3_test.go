package logs

import (
	"bytes"
	"context"
	"io"
	"testing"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
)

const (
	testBucket    = "pangu"                            // TODO: fill in your S3 bucket name
	testPrefix    = "gitgox/test/"                     // TODO: fill in your key prefix
	testEndpoint  = "http://cdn.zcorp.cc:9000"         // TODO: fill in your S3 endpoint, e.g. "https://s3.amazonaws.com"
	testRegion    = "cn-qingdao"                       // TODO: fill in your S3 region
	testAccessKey = "aoYt0DH7e8yBVz32"                 // TODO: fill in your access key
	testSecretKey = "WLvoz7tqjMGD9GJBaT3xCyEwg4qeEo6O" // TODO: fill in your secret key
	testPathStyle = true                               // TODO: set to true if using path-style access (e.g. MinIO)
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	if testBucket == "" || testEndpoint == "" {
		t.Skip("skipping S3 integration test: testBucket or testEndpoint not configured")
	}
	s3cfg := &config.S3Config{
		Bucket:    testBucket,
		Prefix:    testPrefix,
		Endpoint:  testEndpoint,
		Region:    testRegion,
		AccessKey: testAccessKey,
		SecretKey: testSecretKey,
		PathStyle: testPathStyle,
	}
	store, err := NewS3Store(s3cfg)
	if err != nil {
		t.Fatalf("NewS3Store failed: %v", err)
	}
	return store
}

func newTestParam(stepID string) *pipeline.S3Logs {
	return &pipeline.S3Logs{
		PipelineID:  "pipeline-1",
		ExecutionID: "exec-1",
		StageID:     "stage-1",
		StepID:      stepID,
	}
}

func TestStore_Create(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	param := newTestParam("step-create")

	content := "hello, this is a test log line\n"
	err := store.Create(ctx, param, bytes.NewBufferString(content))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// cleanup
	defer func() {
		_ = store.Delete(ctx, param)
	}()

	rc, err := store.Find(ctx, param)
	if err != nil {
		t.Fatalf("Find after Create failed: %v", err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if string(got) != content {
		t.Errorf("content mismatch: got %q, want %q", string(got), content)
	}
}

func TestStore_Find(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	param := newTestParam("step-find")

	content := "find test log content\n"
	if err := store.Create(ctx, param, bytes.NewBufferString(content)); err != nil {
		t.Fatalf("Create setup failed: %v", err)
	}
	defer func() {
		_ = store.Delete(ctx, param)
	}()

	rc, err := store.Find(ctx, param)
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if string(got) != content {
		t.Errorf("content mismatch: got %q, want %q", string(got), content)
	}
}

func TestStore_Find_NotExist(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	param := newTestParam("step-not-exist")

	_, err := store.Find(ctx, param)
	if err == nil {
		t.Fatal("expected error when finding non-existent log, got nil")
	}
}

func TestStore_Update(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	param := newTestParam("step-update")

	original := "original log content\n"
	if err := store.Create(ctx, param, bytes.NewBufferString(original)); err != nil {
		t.Fatalf("Create setup failed: %v", err)
	}
	defer func() {
		_ = store.Delete(ctx, param)
	}()

	updated := "updated log content\n"
	if err := store.Update(ctx, param, bytes.NewBufferString(updated)); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	rc, err := store.Find(ctx, param)
	if err != nil {
		t.Fatalf("Find after Update failed: %v", err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if string(got) != updated {
		t.Errorf("content mismatch after update: got %q, want %q", string(got), updated)
	}
}

func TestStore_Delete(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	param := newTestParam("step-delete")

	content := "log to be deleted\n"
	if err := store.Create(ctx, param, bytes.NewBufferString(content)); err != nil {
		t.Fatalf("Create setup failed: %v", err)
	}

	if err := store.Delete(ctx, param); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := store.Find(ctx, param)
	if err == nil {
		t.Fatal("expected error when finding deleted log, got nil")
	}
}

func TestStore_CreateAndDelete_LargeContent(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	param := newTestParam("step-large")

	// generate ~1MB log content
	var buf bytes.Buffer
	line := "2024-01-01T00:00:00Z INFO this is a repeated log line for large content test\n"
	for buf.Len() < 1<<20 {
		buf.WriteString(line)
	}
	content := buf.String()

	if err := store.Create(ctx, param, bytes.NewBufferString(content)); err != nil {
		t.Fatalf("Create large content failed: %v", err)
	}
	defer func() {
		_ = store.Delete(ctx, param)
	}()

	rc, err := store.Find(ctx, param)
	if err != nil {
		t.Fatalf("Find large content failed: %v", err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if len(got) != len(content) {
		t.Errorf("large content length mismatch: got %d bytes, want %d bytes", len(got), len(content))
	}
}

func TestStore_CRUD_Lifecycle(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	param := newTestParam("step-lifecycle")

	// 1. Create
	v1 := "lifecycle: version 1\n"
	if err := store.Create(ctx, param, bytes.NewBufferString(v1)); err != nil {
		t.Fatalf("step 1 Create failed: %v", err)
	}

	// 2. Find and verify
	rc, err := store.Find(ctx, param)
	if err != nil {
		t.Fatalf("step 2 Find failed: %v", err)
	}
	got, _ := io.ReadAll(rc)
	rc.Close()
	if string(got) != v1 {
		t.Errorf("step 2 content mismatch: got %q, want %q", string(got), v1)
	}

	// 3. Update
	v2 := "lifecycle: version 2\n"
	if err := store.Update(ctx, param, bytes.NewBufferString(v2)); err != nil {
		t.Fatalf("step 3 Update failed: %v", err)
	}

	// 4. Find and verify updated content
	rc, err = store.Find(ctx, param)
	if err != nil {
		t.Fatalf("step 4 Find failed: %v", err)
	}
	got, _ = io.ReadAll(rc)
	rc.Close()
	if string(got) != v2 {
		t.Errorf("step 4 content mismatch: got %q, want %q", string(got), v2)
	}

	// 5. Delete
	if err := store.Delete(ctx, param); err != nil {
		t.Fatalf("step 5 Delete failed: %v", err)
	}

	// 6. Verify deleted
	_, err = store.Find(ctx, param)
	if err == nil {
		t.Fatal("step 6 expected error after Delete, got nil")
	}
}
