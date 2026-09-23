// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      liyang <liyang@easycorp.ltd>

package statistics

import (
	"context"
	"sort"
	"strings"

	statisticsModel "github.com/easysoft/gitfox-community/app/model/statistics"
	userModel "github.com/easysoft/gitfox-community/app/model/user"
	repoStore "github.com/easysoft/gitfox-community/app/store/repo"
	"github.com/easysoft/gitfox-community/pkg/git"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

// CommitCounter 统计代码库的提交次数。
type CommitCounter interface {
	CountCommits(ctx context.Context, params *git.CountCommitsParams) (*git.CountCommitsOutput, error)
}

// UserFinder 根据账号或邮箱查询用户。
type UserFinder interface {
	FindByAccount(ctx context.Context, account string) (*userModel.User, error)
	FindByEmail(ctx context.Context, email string) (*userModel.User, error)
}

type Service struct {
	repoStore     repoStore.Interface
	commitCounter CommitCounter
	userFinder    UserFinder
}

// NewService 创建统计服务实例。
func NewService(store repoStore.Interface, commitCounter CommitCounter, userFinder UserFinder) *Service {
	return &Service{
		repoStore:     store,
		commitCounter: commitCounter,
		userFinder:    userFinder,
	}
}

// GetCommitStatistics 遍历未删除的代码库执行git命令统计提交次数，支持按时间统计和按提交者统计。
func (s *Service) GetCommitStatistics(ctx context.Context, req statisticsModel.ReqForCommitStatistics) (*statisticsModel.RespForCommitStatistics, error) {
	repos, err := s.repoStore.ListAllGitRepos(ctx)
	if err != nil {
		return nil, err
	}

	committerFilter := s.resolveCommitterFilter(ctx, req.User)

	dailyCounts := make(map[string]int64)
	authorCounts := make(map[string]int64)
	for _, repo := range repos {
		/* git仓库目录由gitUID生成，长度不足4位无法定位仓库目录。 */
		if repo == nil || len(repo.GitUID) < 4 {
			continue
		}

		output, err := s.commitCounter.CountCommits(ctx, &git.CountCommitsParams{
			ReadParams: git.ReadParams{RepoUID: repo.GitUID},
			Since:      req.BeginTime,
			Until:      req.EndTime,
			Committer:  committerFilter,
		})
		if err != nil {
			/* 代码库可能正在导入或已不在本地，跳过该代码库继续统计其他代码库。 */
			logger.GetLogger(ctx).Warnf("failed to count commits of repo %d(%s): %v", repo.ID, repo.Identifier, err)
			continue
		}
		if output == nil {
			continue
		}

		for _, stat := range output.CommitStats {
			dailyCounts[stat.Date] += int64(stat.Commits)
			if req.ByUser {
				authorCounts[stat.Author] += int64(stat.Commits)
			}
		}
	}

	resp := &statisticsModel.RespForCommitStatistics{
		Stats:   buildDailyCounts(dailyCounts),
		Authors: make([]statisticsModel.AuthorCount, 0),
	}
	if req.ByUser {
		resp.Authors = s.buildAuthorCounts(ctx, authorCounts)
	}

	return resp, nil
}

// resolveCommitterFilter 解析按人筛选的提交人，传入账号时转换为邮箱。
func (s *Service) resolveCommitterFilter(ctx context.Context, user string) string {
	if user == "" {
		return ""
	}
	if strings.Contains(user, "@") {
		return user
	}

	if user, err := s.userFinder.FindByAccount(ctx, user); err == nil && user != nil && user.Email != "" {
		return user.Email
	}

	return user
}

// buildDailyCounts 将按天统计结果转换为按日期升序排列的切片。
func buildDailyCounts(counts map[string]int64) []statisticsModel.DailyCount {
	dates := make([]string, 0, len(counts))
	for date := range counts {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	stats := make([]statisticsModel.DailyCount, 0, len(dates))
	for _, date := range dates {
		stats = append(stats, statisticsModel.DailyCount{Date: date, Count: counts[date]})
	}

	return stats
}

// buildAuthorCounts 将按提交者统计结果转换为切片，并补充提交者的账号和姓名。
func (s *Service) buildAuthorCounts(ctx context.Context, counts map[string]int64) []statisticsModel.AuthorCount {
	emails := make([]string, 0, len(counts))
	for email := range counts {
		emails = append(emails, email)
	}
	sort.Strings(emails)

	authors := make([]statisticsModel.AuthorCount, 0, len(emails))
	for _, email := range emails {
		author := statisticsModel.AuthorCount{Email: email, Count: counts[email]}
		if user, err := s.userFinder.FindByEmail(ctx, email); err == nil && user != nil {
			author.Account = user.Account
			author.Name = user.DisplayName
		}
		authors = append(authors, author)
	}

	sort.SliceStable(authors, func(i, j int) bool {
		if authors[i].Count != authors[j].Count {
			return authors[i].Count > authors[j].Count
		}
		return authors[i].Email < authors[j].Email
	})

	return authors
}
