package scan

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/easysoft/gitfox-community/app/constant"
	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
	scanStore "github.com/easysoft/gitfox-community/app/store/scan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// typecheckCheck 构造一条 golangci-lint typecheck issue（锚点在 path:startLine）。
func typecheckCheck(message, path string, startLine int64) *scanModel.IssueInputCheck {
	check := &scanModel.IssueInputCheck{
		IssueInputBase: scanModel.IssueInputBase{
			RuleKey: "typecheck",
			Message: message,
			Level:   "LEVEL_MEDIUM",
		},
		IssuePayloadCheck: scanModel.IssuePayloadCheck{
			Tool:     "golangci-lint",
			Category: "CATEGORY_LINT",
			Snippet:  "package main",
			Mode:     "MODE_BLOCK",
		},
	}
	check.Location.Path = path
	check.Location.Range.StartLine = startLine
	check.Location.Range.StartByte = 99 // 锚点位置的字节偏移，展开后不应残留
	return check
}

// TestExpandTypecheckIssue 消息形态取自 golangci-lint 实际输出：
// root 包的文件带 ./ 前缀，子目录包不带（旧实现只认 ./，子目录包整包错误拆不开）。
func TestExpandTypecheckIssue(t *testing.T) {
	type want struct {
		path string
		line int64
		col  int64
		msg  string
	}
	tests := []struct {
		name string
		in   *scanModel.IssueInputCheck
		want []want
	}{
		{
			name: "root 包多条子错误",
			in: typecheckCheck(": # go-demo\n"+
				"./main.go:7:1: syntax error: non-declaration statement outside function body\n"+
				"./xxx.go:3:1: syntax error: non-declaration statement outside function body",
				"main.go", 1),
			want: []want{
				{"main.go", 7, 1, "syntax error: non-declaration statement outside function body"},
				{"xxx.go", 3, 1, "syntax error: non-declaration statement outside function body"},
			},
		},
		{
			name: "子目录包单条子错误（无 ./ 前缀）",
			in:   typecheckCheck(": # edge/pkg\npkg/a.go:4:2: undefined: Println", "pkg/a.go", 1),
			want: []want{{"pkg/a.go", 4, 2, "undefined: Println"}},
		},
		{
			name: "深嵌套目录包单条子错误",
			in:   typecheckCheck(": # edge/pkg/sub\npkg/sub/c.go:4:2: undefined: Println", "pkg/sub/c.go", 1),
			want: []want{{"pkg/sub/c.go", 4, 2, "undefined: Println"}},
		},
		{
			name: "root 包单条子错误",
			in:   typecheckCheck(": # edge\n./main.go:4:2: undefined: Println", "main.go", 1),
			want: []want{{"main.go", 4, 2, "undefined: Println"}},
		},
		{
			name: "root 包与子目录包混合",
			in: typecheckCheck(": # edge\n"+
				"./main.go:7:1: undefined: A\n"+
				"./b.go:9:1: undefined: B\n"+
				"pkg/a.go:4:2: undefined: Println",
				"main.go", 1),
			want: []want{
				{"main.go", 7, 1, "undefined: A"},
				{"b.go", 9, 1, "undefined: B"},
				{"pkg/a.go", 4, 2, "undefined: Println"},
			},
		},
		{
			name: "绝对路径子错误不展开",
			in:   typecheckCheck(": # edge\n/gitfox/991/demo666/main.go:4:2: undefined: Println", "main.go", 1),
			want: nil,
		},
		{
			name: "无子错误行保持原样",
			in:   typecheckCheck(": # edge\npackage edge: no Go files in /src/edge", "main.go", 1),
			want: nil,
		},
		{
			name: "逐行 issue（消息不含位置前缀）保持原样",
			in:   typecheckCheck("expected declaration, found akljdkf", "main.go", 7),
			want: nil,
		},
		{
			name: "非 golangci-lint 工具保持原样",
			in: func() *scanModel.IssueInputCheck {
				check := typecheckCheck("./a.php:7:1: something", "a.php", 1)
				check.Tool = "phpstan"
				return check
			}(),
			want: nil,
		},
		{
			name: "非 typecheck 规则保持原样",
			in: func() *scanModel.IssueInputCheck {
				check := typecheckCheck("./main.go:7:1: something", "main.go", 1)
				check.RuleKey = "govet"
				return check
			}(),
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandTypecheckIssue(tt.in)
			if tt.want == nil {
				require.Len(t, got, 1)
				assert.Same(t, tt.in, got[0], "未展开时应原样返回同一条 issue")
				assert.Equal(t, "package main", got[0].Snippet, "未展开时不应清空片段")
				return
			}
			require.Len(t, got, len(tt.want))
			for i, w := range tt.want {
				item := got[i]
				assert.Equal(t, w.path, item.Location.Path)
				assert.Equal(t, w.line, item.Location.Range.StartLine)
				assert.Equal(t, w.line, item.Location.Range.EndLine, "子错误按单行处理")
				assert.Equal(t, w.col, item.Location.Range.StartColumn)
				assert.Equal(t, w.msg, item.Message)
				// tool/ruleKey/等级/类别沿用原 issue
				assert.Equal(t, "golangci-lint", item.Tool)
				assert.Equal(t, "typecheck", item.RuleKey)
				assert.Equal(t, "LEVEL_MEDIUM", item.Level)
				assert.Equal(t, "CATEGORY_LINT", item.Category)
				// 锚点位置的字节偏移、片段与 blame 不适用于子错误
				assert.Zero(t, item.Location.Range.StartByte)
				assert.Zero(t, item.Location.Range.EndByte)
				assert.Empty(t, item.Snippet)
				assert.Empty(t, item.SnippetWithContext)
				assert.Zero(t, item.Location.Commit.CommitID)
				assert.Equal(t, int64(0), item.Location.Range.StartLine-tt.want[i].line, "行号不应被锚点覆盖")
			}
		})
	}
}

// fakePlanStore 等只实现 buildRuleKeyMapByTask 用到的方法，其余继承接口
// （app/store/scan/mock 中的 gomock 桩未随接口全部更新，见 issue_test.go 的说明）
type fakePlanStore struct{ scanStore.IScanPlan }

func (fakePlanStore) GetSolutionIDList(context.Context, uint) ([]int64, error) {
	return []int64{1}, nil
}

type fakeSolutionStore struct{ scanStore.ISolution }

func (fakeSolutionStore) GetByID(_ context.Context, id uint) (*scanModel.Solution, error) {
	return &scanModel.Solution{ID: int64(id), Status: constant.ScanRuleStatusEnabled}, nil
}

func (fakeSolutionStore) ListRulesets(context.Context, uint) ([]*scanModel.RuleSet, error) {
	return []*scanModel.RuleSet{{ID: 1, Status: constant.ScanRuleStatusEnabled}}, nil
}

type fakeRuleSetStore struct{ scanStore.IRuleSet }

func (fakeRuleSetStore) GetRuleIDList(context.Context, uint) ([]int64, error) {
	return []int64{10}, nil
}

type fakeRuleStore struct{ scanStore.IRule }

func (fakeRuleStore) ListByIDs(_ context.Context, ids []int64) ([]*scanModel.Rule, error) {
	rules := make([]*scanModel.Rule, 0, len(ids))
	for _, id := range ids {
		rules = append(rules, &scanModel.Rule{
			ID:      id,
			RuleKey: "typecheck",
			Lang:    constant.ScanLangGoName,
			Status:  constant.ScanRuleStatusEnabled,
		})
	}
	return rules, nil
}

// TestAnalyzeCheckIssuesExpandsTypecheck 入库链路上的一条聚合 issue 会展开成多条
// 带真实位置的 issue，Total 与展开后的条数一致，payload 里的列号落在 startColumn。
func TestAnalyzeCheckIssuesExpandsTypecheck(t *testing.T) {
	svc := &ScanTaskService{
		planStore:     fakePlanStore{},
		solutionStore: fakeSolutionStore{},
		ruleSetStore:  fakeRuleSetStore{},
		ruleStore:     fakeRuleStore{},
	}
	// 两份聚合：root 包 1 处错误、edge/pkg 包 3 处错误（含新增文件与新增函数）
	const content = `[
		{"tool":"golangci-lint","ruleKey":"typecheck",
		 "message":": # edge\n./main.go:4:2: undefined: Println",
		 "level":"LEVEL_MEDIUM","category":"CATEGORY_LINT","snippet":"package main","mode":"MODE_BLOCK",
		 "location":{"path":"main.go","range":{"startLine":1}}},
		{"tool":"golangci-lint","ruleKey":"typecheck",
		 "message":": # edge/pkg\npkg/a.go:4:2: undefined: Println\npkg/a.go:8:2: undefined: Println\npkg/b.go:4:2: undefined: Println",
		 "level":"LEVEL_MEDIUM","category":"CATEGORY_LINT","snippet":"package pkg","mode":"MODE_BLOCK",
		 "location":{"path":"pkg/a.go","range":{"startLine":1}}}
	]`

	report := &scanModel.ScanIssueReport{}
	issues, err := svc.analyzeCheckIssues(testAIContext(), []byte(content),
		&scanModel.ScanTaskInfo{RepoID: 9, ExecutionRef: "ref"}, report)
	require.NoError(t, err)

	wantLocations := []struct {
		path string
		line int64
	}{
		{"main.go", 4},
		{"pkg/a.go", 4},
		{"pkg/a.go", 8},
		{"pkg/b.go", 4},
	}
	require.Len(t, issues, len(wantLocations))
	// Total 取展开后的条数（展开前是 2），且 4 条都命中了规则
	assert.Equal(t, len(wantLocations), report.Total)
	assert.Zero(t, report.MissingRule)

	for i, w := range wantLocations {
		assert.Equal(t, w.path, issues[i].Path)
		assert.Equal(t, w.line, issues[i].StartLine)
		assert.NotEmpty(t, issues[i].Key, "展开出的子错误各自要有 issue key")
		assert.Equal(t, constant.ScanMethodCheck, issues[i].ScanMethod)
	}
	assert.Equal(t, "undefined: Println", issues[0].Title)
	assert.Equal(t, "pkg/b.go", issues[3].Path, "新增文件上的错误也要入库")

	var payload scanModel.IssuePayloadCheck
	require.NoError(t, json.Unmarshal(issues[2].Payload, &payload))
	assert.Equal(t, "pkg/a.go", payload.Location.Path)
	assert.Equal(t, int64(8), payload.Location.Range.StartLine)
	assert.Equal(t, int64(2), payload.Location.Range.StartColumn)
	assert.Zero(t, payload.Location.Range.StartByte, "列号不再写进字节偏移字段")
	assert.Empty(t, payload.Snippet, "锚点文件的片段不再串到子错误上")
	assert.Empty(t, payload.SnippetWithContext)
	assert.Equal(t, "golangci-lint", payload.Tool)
}

// TestAnalyzeCheckIssuesKeepsPluginExpandedIssues codescan 插件已在上报前展开的报告
// 不再被服务端二次处理：子错误的 message（如 `undefined: Println`）不含位置前缀，
// 位置、片段与列号原样入库，两套展开不会互相叠加。
func TestAnalyzeCheckIssuesKeepsPluginExpandedIssues(t *testing.T) {
	svc := &ScanTaskService{
		planStore:     fakePlanStore{},
		solutionStore: fakeSolutionStore{},
		ruleSetStore:  fakeRuleSetStore{},
		ruleStore:     fakeRuleStore{},
	}
	// 插件 buildFullReport 输出的真实形态：单条子错误，带自身位置、列号与片段
	const content = `[
		{"tool":"golangci-lint","ruleKey":"typecheck","message":"undefined: Println",
		 "level":"LEVEL_MEDIUM","language":"go","category":"CATEGORY_LINT",
		 "snippet":"\tPrintln(\"a2\")","snippetWithContext":"func A2() {\n\tPrintln(\"a2\")\n}",
		 "mode":"MODE_BLOCK",
		 "location":{"path":"pkg/a.go","range":{"startLine":8,"endLine":8,"startColumn":2}}}
	]`

	report := &scanModel.ScanIssueReport{}
	issues, err := svc.analyzeCheckIssues(testAIContext(), []byte(content),
		&scanModel.ScanTaskInfo{RepoID: 9, ExecutionRef: "ref"}, report)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	assert.Equal(t, 1, report.Total)
	assert.Equal(t, "pkg/a.go", issues[0].Path)
	assert.Equal(t, int64(8), issues[0].StartLine)

	var payload scanModel.IssuePayloadCheck
	require.NoError(t, json.Unmarshal(issues[0].Payload, &payload))
	assert.Equal(t, int64(2), payload.Location.Range.StartColumn)
	assert.Equal(t, "\tPrintln(\"a2\")", payload.Snippet, "插件补全的片段不应被清空")
}
