package scan

import (
	"testing"

	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildCommitRange 覆盖非内嵌流水线（create-by-runner）返回给插件的提交范围：
// shas 是 branch 最新两笔（倒序），增量取 次新/最新，其余一律 最新/最新（两值相同即全量）。
func TestBuildCommitRange(t *testing.T) {
	t.Parallel()

	const (
		latest  = "sha-latest"
		second  = "sha-second"
		branch  = "main"
		asFull  = "full"
		asIncre = "incremental"
	)

	tests := []struct {
		name           string
		scanType       constant.ScanType
		shas           []string
		wantBefore     string
		wantAfter      string
		wantErrMessage string
	}{
		{
			name:       "全量计划取最新一笔且两值相同",
			scanType:   constant.ScanTypeFull,
			shas:       []string{latest, second},
			wantBefore: latest,
			wantAfter:  latest,
		},
		{
			name:       "全量计划仅一笔提交",
			scanType:   constant.ScanTypeFull,
			shas:       []string{latest},
			wantBefore: latest,
			wantAfter:  latest,
		},
		{
			name:       "增量计划取次新与最新",
			scanType:   constant.ScanTypeIncremental,
			shas:       []string{latest, second},
			wantBefore: second,
			wantAfter:  latest,
		},
		{
			name:       "增量计划仅一笔提交退化为全量语义",
			scanType:   constant.ScanTypeIncremental,
			shas:       []string{latest},
			wantBefore: latest,
			wantAfter:  latest,
		},
		{
			name:           "增量计划无提交报错",
			scanType:       constant.ScanTypeIncremental,
			shas:           nil,
			wantErrMessage: "no commits found on branch " + branch,
		},
		{
			name:           "全量计划无提交报错",
			scanType:       constant.ScanTypeFull,
			shas:           []string{},
			wantErrMessage: "no commits found on branch " + branch,
		},
		{
			name:       "未知 scanType 按全量语义",
			scanType:   constant.ScanType("whatever"),
			shas:       []string{latest, second},
			wantBefore: latest,
			wantAfter:  latest,
		},
		{
			name:       "空 scanType 按全量语义",
			scanType:   constant.ScanType(""),
			shas:       []string{latest, second},
			wantBefore: latest,
			wantAfter:  latest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			before, after, err := buildCommitRange(tt.scanType, branch, tt.shas)
			if tt.wantErrMessage != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErrMessage)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantBefore, before)
			assert.Equal(t, tt.wantAfter, after)
		})
	}
}

// TestBuildLatestTwoCommitRange 覆盖增量口径的纯函数核心：最新为 after、次新为 before；
// 分支只有一笔时两值相同（增量无从谈起）；无提交时报错。
func TestBuildLatestTwoCommitRange(t *testing.T) {
	t.Parallel()

	const (
		latest = "sha-latest"
		second = "sha-second"
		branch = "main"
	)

	tests := []struct {
		name           string
		shas           []string
		wantBefore     string
		wantAfter      string
		wantErrMessage string
	}{
		{
			name:       "两笔提交取次新与最新",
			shas:       []string{latest, second},
			wantBefore: second,
			wantAfter:  latest,
		},
		{
			name:       "仅一笔提交两值相同",
			shas:       []string{latest},
			wantBefore: latest,
			wantAfter:  latest,
		},
		{
			name:           "无提交报错",
			shas:           nil,
			wantErrMessage: "no commits found on branch " + branch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			before, after, err := buildLatestTwoCommitRange(branch, tt.shas)
			if tt.wantErrMessage != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErrMessage)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantBefore, before)
			assert.Equal(t, tt.wantAfter, after)
		})
	}
}
