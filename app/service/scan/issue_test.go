package scan

import (
	"context"
	"testing"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	executionModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/execution"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
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
	branches []string // 实际写入的 repoBranch
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
		f.branches = append(f.branches, item.RepoBranch)
	}
	return nil
}

func (f *fakeIssueStore) Create(_ context.Context, issue *scanStore.ScanIssue) error {
	f.createCalls++
	if f.dupAt > 0 && f.createCalls == f.dupAt {
		return errors.ErrDuplicate
	}
	f.created = append(f.created, issue.Key)
	f.branches = append(f.branches, issue.RepoBranch)
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
	issues, err := svc.analyzeAIIssues(testAIContext(), content, 9, scanModel.AIScopePPM(7), report)

	require.NoError(t, err)
	require.Len(t, issues, 1)
	assert.Equal(t, constant.ScanMethodAI, issues[0].ScanMethod)
	assert.Equal(t, "a.go", issues[0].Path)
	assert.Equal(t, int64(3), issues[0].StartLine)
	assert.Equal(t, int64(4), issues[0].EndLine)
	// key 里带着 ppmID，确保同一 PR 的意见能稳定去重
	assert.Equal(t, scanModel.GenerateIssueKey(9, "", "ai/ppm/7/a.go/3/4"), issues[0].Key)
	assert.Equal(t, 1, report.Total)
}

// MR（ppmID）与无 MR（executionID）来自两张互相独立的自增表，同一仓库内取值可能相同。
// 两者作用域不同，即便 path 与行号完全一致也必须落在不同的 key 上，否则后者会被
// insertAIIssues 的去重逻辑静默丢弃。
func TestAnalyzeAIIssuesScopeSpacesDoNotCollide(t *testing.T) {
	svc := &ScanTaskService{}
	content := []byte(`{"status":"complete","comments":[` +
		`{"path":"a.go","content":"c1","start_line":3,"end_line":4,` +
		`"category":"bug","severity":"high"}]}`)

	ppmReport := &scanModel.ScanIssueReport{}
	ppmIssues, err := svc.analyzeAIIssues(testAIContext(), content, 9, scanModel.AIScopePPM(1), ppmReport)
	require.NoError(t, err)
	require.Len(t, ppmIssues, 1)

	execReport := &scanModel.ScanIssueReport{}
	execIssues, err := svc.analyzeAIIssues(testAIContext(), content, 9, scanModel.AIScopeExecution(1), execReport)
	require.NoError(t, err)
	require.Len(t, execIssues, 1)

	assert.NotEqual(t, ppmIssues[0].Key, execIssues[0].Key)
}

func TestAnalyzeAIIssuesSkipsEmptyComments(t *testing.T) {
	svc := &ScanTaskService{}
	for _, content := range [][]byte{
		[]byte(`{"status":"complete","comments":[]}`),
		[]byte(`{"status":"skipped"}`),
	} {
		report := &scanModel.ScanIssueReport{}
		issues, err := svc.analyzeAIIssues(testAIContext(), content, 9, scanModel.AIScopePPM(7), report)
		require.NoError(t, err)
		assert.Empty(t, issues)
	}
}

// repoBranch 一列只有一种格式：有 MR 时 ppm.SourceBranch 是短分支名，落库前要补成
// 完整 ref，与 codescan 的 ExecutionRef、无 MR 的 execution.Ref 对齐；已经带前缀的
// 输入不得重复叠加。
func TestSaveAIIssuesNormalizesRepoBranch(t *testing.T) {
	cases := []struct {
		name         string
		sourceBranch string
		want         string
	}{
		{name: "短分支名补全", sourceBranch: "dev", want: "refs/heads/dev"},
		{name: "带斜杠的分支名", sourceBranch: "feature/scan", want: "refs/heads/feature/scan"},
		{name: "已是完整 ref 不重复叠加", sourceBranch: "refs/heads/dev", want: "refs/heads/dev"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeIssueStore{}
			svc := &ScanTaskService{issueStore: store}
			ppm := testAIPullRequest()
			ppm.SourceBranch = tt.sourceBranch

			_, _, err := svc.saveAIIssues(testAIContext(), []*scanModel.ScanIssue{testAIIssue("k1")}, ppm)

			require.NoError(t, err)
			assert.Equal(t, []string{tt.want}, store.branches)
		})
	}
}

// 无 MR 场景 execution.Ref 本身即完整 ref，原样落库
func TestSaveAIIssuesForExecutionKeepsFullRef(t *testing.T) {
	store := &fakeIssueStore{}
	svc := &ScanTaskService{issueStore: store}

	_, _, err := svc.saveAIIssuesForExecution(testAIContext(),
		[]*scanModel.ScanIssue{testAIIssue("k1")},
		&executionModel.Execution{ID: 3, Ref: "refs/heads/main", CreatedBy: "tester"}, 9)

	require.NoError(t, err)
	assert.Equal(t, []string{"refs/heads/main"}, store.branches)
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
