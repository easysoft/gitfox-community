package scan

import (
	"context"
	"testing"

	"github.com/easysoft/gitfox-community/app/constant"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
	scanStore "github.com/easysoft/gitfox-community/app/store/scan"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeIssueStore 只实现 saveAIIssues 用到的方法，其余方法继承自接口
// （app/store/scan/mock 中的 MockIScanIssue 尚未随接口更新，这里不依赖它）
type fakeIssueStore struct {
	scanStore.IScanIssue

	existingKeys []string
	existingErr  error

	batchErr   error
	batchSizes []int

	dupAt       int // 第 N 次 Create 返回重复键错误，0 表示不返回
	createCalls int

	created []string // 实际写入的 issue key
}

func (f *fakeIssueStore) ListExistingKeys(_ context.Context, _ []string) ([]string, error) {
	return f.existingKeys, f.existingErr
}

func (f *fakeIssueStore) BatchCreate(_ context.Context, issues []*scanStore.ScanIssue) error {
	f.batchSizes = append(f.batchSizes, len(issues))
	if f.batchErr != nil {
		return f.batchErr
	}
	for _, item := range issues {
		f.created = append(f.created, item.Key)
	}
	return nil
}

func (f *fakeIssueStore) Create(_ context.Context, issue *scanStore.ScanIssue) error {
	f.createCalls++
	if f.dupAt > 0 && f.createCalls == f.dupAt {
		return errors.ErrDuplicate
	}
	f.created = append(f.created, issue.Key)
	return nil
}

func testAIContext() context.Context {
	return logger.WithLogger(context.Background(), logger.NewNopLogger())
}

func testAIIssue(key string) *scanModel.ScanIssue {
	return &scanModel.ScanIssue{
		Key:        key,
		Title:      "title",
		Path:       "a.go",
		StartLine:  1,
		EndLine:    1,
		ScanMethod: constant.ScanMethodAI,
	}
}

func testAIPullRequest() *repoModel.PullRequest {
	return &repoModel.PullRequest{ID: 7, RepoID: 9, SourceBranch: "dev", CreatedBy: "tester"}
}

// analyzeAIIssues 只按意见数量判断，partial（部分文件评审失败）等状态同样要入库
func TestAnalyzeAIIssuesAcceptsNonCompleteStatus(t *testing.T) {
	svc := &ScanTaskService{}
	content := []byte(`{"status":"partial","comments":[` +
		`{"path":"a.go","content":"c1","suggestion_code":"s1","existing_code":"e1",` +
		`"start_line":3,"end_line":4,"category":"bug","severity":"high"}]}`)

	report := &scanModel.ScanIssueReport{}
	issues, err := svc.analyzeAIIssues(testAIContext(), content, 9, 7, report)

	require.NoError(t, err)
	require.Len(t, issues, 1)
	assert.Equal(t, constant.ScanMethodAI, issues[0].ScanMethod)
	assert.Equal(t, "a.go", issues[0].Path)
	assert.Equal(t, int64(3), issues[0].StartLine)
	assert.Equal(t, int64(4), issues[0].EndLine)
	// key 里带着 ppmID，确保同一 PR 的意见能稳定去重
	assert.Equal(t, scanModel.GenerateIssueKey(9, "", "ai/7/a.go/3/4"), issues[0].Key)
	assert.Equal(t, 1, report.Total)
}

func TestAnalyzeAIIssuesSkipsEmptyComments(t *testing.T) {
	svc := &ScanTaskService{}
	for _, content := range [][]byte{
		[]byte(`{"status":"complete","comments":[]}`),
		[]byte(`{"status":"skipped"}`),
	} {
		report := &scanModel.ScanIssueReport{}
		issues, err := svc.analyzeAIIssues(testAIContext(), content, 9, 7, report)
		require.NoError(t, err)
		assert.Empty(t, issues)
	}
}

// saveAIIssues 返回实际插入/跳过的条数，而不是输入条数
func TestSaveAIIssuesCounts(t *testing.T) {
	t.Run("全部插入", func(t *testing.T) {
		store := &fakeIssueStore{}
		svc := &ScanTaskService{issueStore: store}

		added, skipped, err := svc.saveAIIssues(testAIContext(),
			[]*scanModel.ScanIssue{testAIIssue("k1"), testAIIssue("k2")}, testAIPullRequest())

		require.NoError(t, err)
		assert.Equal(t, 2, added)
		assert.Equal(t, 0, skipped)
		assert.Equal(t, []string{"k1", "k2"}, store.created)
		assert.Equal(t, []int{2}, store.batchSizes)
	})

	t.Run("key 已存在则跳过", func(t *testing.T) {
		store := &fakeIssueStore{existingKeys: []string{"k1"}}
		svc := &ScanTaskService{issueStore: store}

		added, skipped, err := svc.saveAIIssues(testAIContext(),
			[]*scanModel.ScanIssue{testAIIssue("k1")}, testAIPullRequest())

		require.NoError(t, err)
		assert.Equal(t, 0, added)
		assert.Equal(t, 1, skipped)
		assert.Empty(t, store.created)
		assert.Empty(t, store.batchSizes)
	})

	t.Run("批内重复 key 只插一条", func(t *testing.T) {
		store := &fakeIssueStore{}
		svc := &ScanTaskService{issueStore: store}

		added, skipped, err := svc.saveAIIssues(testAIContext(),
			[]*scanModel.ScanIssue{testAIIssue("k1"), testAIIssue("k1")}, testAIPullRequest())

		require.NoError(t, err)
		assert.Equal(t, 1, added)
		assert.Equal(t, 1, skipped)
		assert.Equal(t, []string{"k1"}, store.created)
	})

	t.Run("整批冲突时逐条插入，不再整批丢弃", func(t *testing.T) {
		store := &fakeIssueStore{batchErr: errors.ErrDuplicate, dupAt: 2}
		svc := &ScanTaskService{issueStore: store}

		added, skipped, err := svc.saveAIIssues(testAIContext(),
			[]*scanModel.ScanIssue{testAIIssue("k1"), testAIIssue("k2")}, testAIPullRequest())

		require.NoError(t, err)
		assert.Equal(t, 1, added)
		assert.Equal(t, 1, skipped)
		assert.Equal(t, 2, store.createCalls)
		assert.Equal(t, []string{"k1"}, store.created)
	})

	t.Run("非重复键错误照样返回", func(t *testing.T) {
		store := &fakeIssueStore{batchErr: errors.ErrDataTooLong}
		svc := &ScanTaskService{issueStore: store}

		_, _, err := svc.saveAIIssues(testAIContext(),
			[]*scanModel.ScanIssue{testAIIssue("k1")}, testAIPullRequest())

		require.Error(t, err)
		assert.ErrorIs(t, err, errors.ErrDataTooLong)
	})
}
