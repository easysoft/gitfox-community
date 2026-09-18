package statistics

import (
	"context"
	"errors"
	"testing"

	commitModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	statisticsModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/statistics"
	userModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo/mock"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"go.uber.org/mock/gomock"
)

type fakeCommitCounter struct {
	outputs  map[string]*git.CountCommitsOutput
	requests []*git.CountCommitsParams
	errs     map[string]error
}

func (f *fakeCommitCounter) CountCommits(ctx context.Context, params *git.CountCommitsParams) (*git.CountCommitsOutput, error) {
	f.requests = append(f.requests, params)
	if err := f.errs[params.RepoUID]; err != nil {
		return nil, err
	}
	return f.outputs[params.RepoUID], nil
}

type fakeUserFinder struct {
	byAccount map[string]*userModel.User
	byEmail   map[string]*userModel.User
}

func (f *fakeUserFinder) FindByAccount(ctx context.Context, account string) (*userModel.User, error) {
	if user, ok := f.byAccount[account]; ok {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (f *fakeUserFinder) FindByEmail(ctx context.Context, email string) (*userModel.User, error) {
	if user, ok := f.byEmail[email]; ok {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func newCommitStats(repoUID string, stats ...commitModel.CommitCount) *git.CountCommitsOutput {
	return &git.CountCommitsOutput{CommitStats: stats, AuthorCount: map[string]int{}, Total: len(stats)}
}

func commitCount(date, author string, commits int) commitModel.CommitCount {
	return commitModel.CommitCount{Date: date, Author: author, CommitCountData: commitModel.CommitCountData{Commits: commits}}
}

// TestGetCommitStatistics 验证按时间汇总多个代码库的提交次数。
func TestGetCommitStatistics(t *testing.T) {
	store := mock.NewMockInterface(gomock.NewController(t))
	store.EXPECT().ListAllGitRepos(gomock.Any()).Return([]*repoModel.Repo{
		{ID: 1, Identifier: "repo1", GitUID: "uid1"},
		{ID: 2, Identifier: "repo2", GitUID: "uid2"},
	}, nil)
	counter := &fakeCommitCounter{
		outputs: map[string]*git.CountCommitsOutput{
			"uid1": newCommitStats("uid1", commitCount("2026-09-15", "admin@example.com", 2), commitCount("2026-09-16", "dev@example.com", 1)),
			"uid2": newCommitStats("uid2", commitCount("2026-09-15", "dev@example.com", 3)),
		},
		errs: map[string]error{},
	}
	svc := NewService(store, counter, &fakeUserFinder{})

	resp, err := svc.GetCommitStatistics(context.Background(), statisticsModel.ReqForCommitStatistics{
		BeginTime: "2026-09-15",
		EndTime:   "2026-09-16",
	})
	if err != nil {
		t.Fatalf("GetCommitStatistics returned error: %v", err)
	}

	if len(resp.Stats) != 2 {
		t.Fatalf("expected 2 days, got %d(%+v)", len(resp.Stats), resp.Stats)
	}
	if resp.Stats[0].Date != "2026-09-15" || resp.Stats[0].Count != 5 {
		t.Errorf("unexpected first day: %+v", resp.Stats[0])
	}
	if resp.Stats[1].Date != "2026-09-16" || resp.Stats[1].Count != 1 {
		t.Errorf("unexpected second day: %+v", resp.Stats[1])
	}
	if len(resp.Authors) != 0 {
		t.Errorf("expected no authors when byUser is false, got %+v", resp.Authors)
	}
	if len(counter.requests) != 2 || counter.requests[0].Since != "2026-09-15" || counter.requests[0].Until != "2026-09-16" {
		t.Errorf("unexpected git requests: %+v", counter.requests)
	}
}

// TestGetCommitStatisticsByUser 验证按提交者统计时返回每个提交者的信息。
func TestGetCommitStatisticsByUser(t *testing.T) {
	store := mock.NewMockInterface(gomock.NewController(t))
	store.EXPECT().ListAllGitRepos(gomock.Any()).Return([]*repoModel.Repo{{ID: 1, Identifier: "repo1", GitUID: "uid1"}}, nil)
	counter := &fakeCommitCounter{
		outputs: map[string]*git.CountCommitsOutput{
			"uid1": newCommitStats("uid1", commitCount("2026-09-15", "admin@example.com", 1), commitCount("2026-09-15", "dev@example.com", 4)),
		},
		errs: map[string]error{},
	}
	userFinder := &fakeUserFinder{byEmail: map[string]*userModel.User{
		"admin@example.com": {Account: "admin", DisplayName: "管理员", Email: "admin@example.com"},
	}}
	svc := NewService(store, counter, userFinder)

	resp, err := svc.GetCommitStatistics(context.Background(), statisticsModel.ReqForCommitStatistics{
		BeginTime: "2026-09-15",
		EndTime:   "2026-09-15",
		ByUser:    true,
	})
	if err != nil {
		t.Fatalf("GetCommitStatistics returned error: %v", err)
	}

	if len(resp.Authors) != 2 {
		t.Fatalf("expected 2 authors, got %d(%+v)", len(resp.Authors), resp.Authors)
	}
	if resp.Authors[0].Email != "dev@example.com" || resp.Authors[0].Count != 4 {
		t.Errorf("unexpected first author: %+v", resp.Authors[0])
	}
	if resp.Authors[1].Account != "admin" || resp.Authors[1].Name != "管理员" || resp.Authors[1].Count != 1 {
		t.Errorf("unexpected second author: %+v", resp.Authors[1])
	}
}

// TestGetCommitStatisticsWithUserFilter 验证按人筛选时账号会转换为邮箱下发到git命令。
func TestGetCommitStatisticsWithUserFilter(t *testing.T) {
	store := mock.NewMockInterface(gomock.NewController(t))
	store.EXPECT().ListAllGitRepos(gomock.Any()).Return([]*repoModel.Repo{{ID: 1, Identifier: "repo1", GitUID: "uid1"}}, nil)
	counter := &fakeCommitCounter{outputs: map[string]*git.CountCommitsOutput{}, errs: map[string]error{}}
	userFinder := &fakeUserFinder{byAccount: map[string]*userModel.User{
		"admin": {Account: "admin", DisplayName: "管理员", Email: "admin@example.com"},
	}}
	svc := NewService(store, counter, userFinder)

	if _, err := svc.GetCommitStatistics(context.Background(), statisticsModel.ReqForCommitStatistics{
		BeginTime: "2026-09-15",
		EndTime:   "2026-09-16",
		User:      "admin",
	}); err != nil {
		t.Fatalf("GetCommitStatistics returned error: %v", err)
	}

	if len(counter.requests) != 1 || counter.requests[0].Committer != "admin@example.com" {
		t.Errorf("expected committer filter admin@example.com, got %+v", counter.requests)
	}
}

// TestGetCommitStatisticsSkipErrorRepo 验证单个代码库统计失败时不影响其他代码库。
func TestGetCommitStatisticsSkipErrorRepo(t *testing.T) {
	store := mock.NewMockInterface(gomock.NewController(t))
	store.EXPECT().ListAllGitRepos(gomock.Any()).Return([]*repoModel.Repo{
		{ID: 1, Identifier: "repo1", GitUID: "uid1"},
		{ID: 2, Identifier: "repo2", GitUID: "uid2"},
	}, nil)
	counter := &fakeCommitCounter{
		outputs: map[string]*git.CountCommitsOutput{
			"uid2": newCommitStats("uid2", commitCount("2026-09-16", "dev@example.com", 2)),
		},
		errs: map[string]error{"uid1": errors.New("repository not found")},
	}
	svc := NewService(store, counter, &fakeUserFinder{})

	resp, err := svc.GetCommitStatistics(context.Background(), statisticsModel.ReqForCommitStatistics{
		BeginTime: "2026-09-16",
		EndTime:   "2026-09-16",
	})
	if err != nil {
		t.Fatalf("GetCommitStatistics returned error: %v", err)
	}
	if len(resp.Stats) != 1 || resp.Stats[0].Count != 2 {
		t.Errorf("unexpected stats: %+v", resp.Stats)
	}
}

// TestGetCommitStatisticsListReposError 验证查询代码库失败时直接返回错误。
func TestGetCommitStatisticsListReposError(t *testing.T) {
	store := mock.NewMockInterface(gomock.NewController(t))
	store.EXPECT().ListAllGitRepos(gomock.Any()).Return(nil, errors.New("query repos failed"))
	svc := NewService(store, &fakeCommitCounter{}, &fakeUserFinder{})

	if _, err := svc.GetCommitStatistics(context.Background(), statisticsModel.ReqForCommitStatistics{
		BeginTime: "2026-09-15",
		EndTime:   "2026-09-16",
	}); err == nil {
		t.Error("expected error when listing repos failed, got nil")
	}
}
