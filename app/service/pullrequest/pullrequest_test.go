package pullrequest

import (
	"context"
	"testing"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	gitMock "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/branch/mock"
	repoSvcMock "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo/mock"
	prStoreMock "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/pullrequest/mock"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/sha"

	"go.uber.org/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	ErrDBNotFound     = errors.New("记录不存在")
	ErrDBUpdateFailed = errors.New("数据库更新失败")
	ErrAccessDenied   = errors.New("访问被拒绝")
	ErrGitOperation   = errors.New("Git操作失败")
)

// ====================== GetMetadata 方法测试 ======================
// GetMetadataTestCase 表格驱动用例结构体
type GetMetadataTestCase struct {
	name string
	// 前置条件
	mockPRStoreGet func(m *prStoreMock.MockInterface)
	mockRepoSvc    func(m *repoSvcMock.MockInterface)
	mockGitSvc     func(m *gitMock.MockInterface)
	mockPRStoreUpd func(m *prStoreMock.MockInterface)
	inputRepoID    uint
	inputPRNumber  uint
	inputSession   *auth.Session
	// 预期结果
	expectedResult *repoModel.DiffStats
	expectedErr    error
}

// TestGetMetadata GetMetadata 单测入口
func TestGetMetadata(t *testing.T) {
	t.Parallel()

	commitCount := uint(5)
	fileCount := uint(10)
	additions := uint(100)
	deletions := uint(50)

	testCases := []GetMetadataTestCase{
		{
			name: "场景1：正常查询 - PR已有缓存的DiffStats（直接返回缓存数据）",
			mockPRStoreGet: func(m *prStoreMock.MockInterface) {
				mockPR := &repoModel.PullRequest{
					ID:           1,
					RepoID:       100,
					SourceSHA:    "abc123",
					MergeBaseSHA: "def456",
					CommitCount:  &commitCount,
					FileCount:    &fileCount,
					Additions:    &additions,
					Deletions:    &deletions,
				}
				m.EXPECT().GetByID(gomock.Any(), uint(100), uint(1)).Return(mockPR, nil)
			},
			mockRepoSvc: func(m *repoSvcMock.MockInterface) {
				mockRepo := &repoModel.Repo{
					ID:     100,
					GitUID: "repo-uid-100",
				}
				m.EXPECT().GetRepoCheckAccessByID(
					gomock.Any(),
					gomock.Any(),
					uint(100),
					[]constant.Permission{constant.PermissionRepoView},
				).Return(mockRepo, nil)
			},
			mockGitSvc: func(m *gitMock.MockInterface) {
				// 有缓存数据时不会调用Git服务
			},
			mockPRStoreUpd: func(m *prStoreMock.MockInterface) {
				// 有缓存数据时不会更新
			},
			inputRepoID:   100,
			inputPRNumber: 1,
			inputSession: &auth.Session{
				User: &principal.User{ID: 1, Account: "testuser"},
			},
			expectedResult: &repoModel.DiffStats{
				Commits:      5,
				FilesChanged: 10,
				Additions:    100,
				Deletions:    50,
			},
			expectedErr: nil,
		},
		{
			name: "场景2：正常查询 - PR无缓存DiffStats（调用Git获取并更新）",
			mockPRStoreGet: func(m *prStoreMock.MockInterface) {
				mockPR := &repoModel.PullRequest{
					ID:           2,
					RepoID:       100,
					SourceSHA:    "abc123",
					MergeBaseSHA: "def456",
					CommitCount:  nil, // 无缓存
					FileCount:    nil,
					Additions:    nil,
					Deletions:    nil,
				}
				m.EXPECT().GetByID(gomock.Any(), uint(100), uint(2)).Return(mockPR, nil)
			},
			mockRepoSvc: func(m *repoSvcMock.MockInterface) {
				mockRepo := &repoModel.Repo{
					ID:     100,
					GitUID: "repo-uid-100",
				}
				m.EXPECT().GetRepoCheckAccessByID(
					gomock.Any(),
					gomock.Any(),
					uint(100),
					[]constant.Permission{constant.PermissionRepoView},
				).Return(mockRepo, nil)
			},
			mockGitSvc: func(m *gitMock.MockInterface) {
				mockOutput := git.DiffStatsOutput{
					Commits:      3,
					FilesChanged: 8,
					Additions:    50,
					Deletions:    20,
				}
				m.EXPECT().DiffStats(gomock.Any(), gomock.Any()).Return(mockOutput, nil)
			},
			mockPRStoreUpd: func(m *prStoreMock.MockInterface) {
				m.EXPECT().UpdateDiffStatsSafe(gomock.Any(), uint(100), gomock.Any(), gomock.Any()).Return(nil)
			},
			inputRepoID:   100,
			inputPRNumber: 2,
			inputSession: &auth.Session{
				User: &principal.User{ID: 1, Account: "testuser"},
			},
			expectedResult: &repoModel.DiffStats{
				Commits:      3,
				FilesChanged: 8,
				Additions:    50,
				Deletions:    20,
			},
			expectedErr: nil,
		},
		{
			name: "场景3：PR不存在 - pullRequestStore.GetByID返回错误",
			mockPRStoreGet: func(m *prStoreMock.MockInterface) {
				m.EXPECT().GetByID(gomock.Any(), uint(100), uint(999)).Return(nil, ErrDBNotFound)
			},
			mockRepoSvc: func(m *repoSvcMock.MockInterface) {
				// PR不存在时不会调用repoSvc
			},
			mockGitSvc: func(m *gitMock.MockInterface) {
				// PR不存在时不会调用gitSvc
			},
			mockPRStoreUpd: func(m *prStoreMock.MockInterface) {
				// PR不存在时不会更新
			},
			inputRepoID:   100,
			inputPRNumber: 999,
			inputSession: &auth.Session{
				User: &principal.User{ID: 1, Account: "testuser"},
			},
			expectedResult: nil,
			expectedErr:    ErrDBNotFound,
		},
		{
			name: "场景4：权限校验失败 - repoSvc.GetRepoCheckAccessByID返回错误",
			mockPRStoreGet: func(m *prStoreMock.MockInterface) {
				mockPR := &repoModel.PullRequest{
					ID:           3,
					RepoID:       100,
					SourceSHA:    "abc123",
					MergeBaseSHA: "def456",
				}
				m.EXPECT().GetByID(gomock.Any(), uint(100), uint(3)).Return(mockPR, nil)
			},
			mockRepoSvc: func(m *repoSvcMock.MockInterface) {
				m.EXPECT().GetRepoCheckAccessByID(
					gomock.Any(),
					gomock.Any(),
					uint(100),
					[]constant.Permission{constant.PermissionRepoView},
				).Return(nil, ErrAccessDenied)
			},
			mockGitSvc: func(m *gitMock.MockInterface) {
				// 权限校验失败时不会调用gitSvc
			},
			mockPRStoreUpd: func(m *prStoreMock.MockInterface) {
				// 权限校验失败时不会更新
			},
			inputRepoID:   100,
			inputPRNumber: 3,
			inputSession: &auth.Session{
				User: &principal.User{ID: 1, Account: "testuser"},
			},
			expectedResult: nil,
			expectedErr:    ErrAccessDenied,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPRStore := prStoreMock.NewMockInterface(ctrl)
			mockRepoSvc := repoSvcMock.NewMockInterface(ctrl)
			mockGitSvc := gitMock.NewMockInterface(ctrl)

			service := &Service{
				gitSvc:           mockGitSvc,
				repoSvc:          mockRepoSvc,
				pullRequestStore: mockPRStore,
			}

			tc.mockPRStoreGet(mockPRStore)
			tc.mockRepoSvc(mockRepoSvc)
			tc.mockGitSvc(mockGitSvc)
			tc.mockPRStoreUpd(mockPRStore)

			ctx := context.Background()
			actualResult, actualErr := service.GetMetadata(ctx, tc.inputSession, tc.inputRepoID, tc.inputPRNumber)

			if tc.expectedErr != nil {
				assert.Error(t, actualErr)
				assert.True(t, errors.Is(actualErr, tc.expectedErr),
					"GetMetadata() 预期错误为 %v，实际为 %v", tc.expectedErr, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}

			assert.Equal(t, tc.expectedResult, actualResult)
		})
	}
}

// ====================== GetDiffStats 方法测试 ======================
// GetDiffStatsTestCase 表格驱动用例结构体
type GetDiffStatsTestCase struct {
	name string
	// 前置条件
	mockRepoSvc     func(m *repoSvcMock.MockInterface)
	mockGitSvc      func(m *gitMock.MockInterface)
	inputRepoID     uint
	inputPathToPath string
	inputSession    *auth.Session
	// 预期结果
	expectedResult *repoModel.DiffStats
	expectedErr    error
}

// TestGetDiffStats GetDiffStats 单测入口
func TestGetDiffStats(t *testing.T) {
	t.Parallel()

	testCases := []GetDiffStatsTestCase{
		{
			name: "场景1：正常查询 - 使用...分隔符（三点比较）",
			mockRepoSvc: func(m *repoSvcMock.MockInterface) {
				mockRepo := &repoModel.Repo{
					ID:     100,
					GitUID: "repo-uid-100",
				}
				m.EXPECT().GetRepoCheckAccessByID(
					gomock.Any(),
					gomock.Any(),
					uint(100),
					[]constant.Permission{constant.PermissionRepoView},
				).Return(mockRepo, nil)
			},
			mockGitSvc: func(m *gitMock.MockInterface) {
				mockOutput := git.DiffStatsOutput{
					Commits:      5,
					FilesChanged: 10,
					Additions:    100,
					Deletions:    50,
				}
				m.EXPECT().DiffStats(gomock.Any(), gomock.Any()).Return(mockOutput, nil)
			},
			inputRepoID:     100,
			inputPathToPath: "main...feature",
			inputSession: &auth.Session{
				User: &principal.User{ID: 1, Account: "testuser"},
			},
			expectedResult: &repoModel.DiffStats{
				Commits:      5,
				FilesChanged: 10,
				Additions:    100,
				Deletions:    50,
			},
			expectedErr: nil,
		},
		{
			name: "场景2：正常查询 - 使用..分隔符（两点比较）",
			mockRepoSvc: func(m *repoSvcMock.MockInterface) {
				mockRepo := &repoModel.Repo{
					ID:     100,
					GitUID: "repo-uid-100",
				}
				m.EXPECT().GetRepoCheckAccessByID(
					gomock.Any(),
					gomock.Any(),
					uint(100),
					[]constant.Permission{constant.PermissionRepoView},
				).Return(mockRepo, nil)
			},
			mockGitSvc: func(m *gitMock.MockInterface) {
				mockOutput := git.DiffStatsOutput{
					Commits:      3,
					FilesChanged: 5,
					Additions:    30,
					Deletions:    10,
				}
				m.EXPECT().DiffStats(gomock.Any(), gomock.Any()).Return(mockOutput, nil)
			},
			inputRepoID:     100,
			inputPathToPath: "main..feature",
			inputSession: &auth.Session{
				User: &principal.User{ID: 1, Account: "testuser"},
			},
			expectedResult: &repoModel.DiffStats{
				Commits:      3,
				FilesChanged: 5,
				Additions:    30,
				Deletions:    10,
			},
			expectedErr: nil,
		},
		{
			name: "场景3：路径格式错误 - 无分隔符",
			mockRepoSvc: func(m *repoSvcMock.MockInterface) {
				mockRepo := &repoModel.Repo{
					ID:     100,
					GitUID: "repo-uid-100",
				}
				m.EXPECT().GetRepoCheckAccessByID(
					gomock.Any(),
					gomock.Any(),
					uint(100),
					[]constant.Permission{constant.PermissionRepoView},
				).Return(mockRepo, nil)
			},
			mockGitSvc: func(m *gitMock.MockInterface) {
				// 路径格式错误时不会调用gitSvc
			},
			inputRepoID:     100,
			inputPathToPath: "invalid-path",
			inputSession: &auth.Session{
				User: &principal.User{ID: 1, Account: "testuser"},
			},
			expectedResult: nil,
			expectedErr:    errors.Newf("invalid format \"%s\"", "invalid-path"),
		},
		{
			name: "场景4：权限校验失败",
			mockRepoSvc: func(m *repoSvcMock.MockInterface) {
				m.EXPECT().GetRepoCheckAccessByID(
					gomock.Any(),
					gomock.Any(),
					uint(100),
					[]constant.Permission{constant.PermissionRepoView},
				).Return(nil, ErrAccessDenied)
			},
			mockGitSvc: func(m *gitMock.MockInterface) {
				// 权限校验失败时不会调用gitSvc
			},
			inputRepoID:     100,
			inputPathToPath: "main...feature",
			inputSession: &auth.Session{
				User: &principal.User{ID: 1, Account: "testuser"},
			},
			expectedResult: nil,
			expectedErr:    ErrAccessDenied,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepoSvc := repoSvcMock.NewMockInterface(ctrl)
			mockGitSvc := gitMock.NewMockInterface(ctrl)

			service := &Service{
				gitSvc:  mockGitSvc,
				repoSvc: mockRepoSvc,
			}

			tc.mockRepoSvc(mockRepoSvc)
			tc.mockGitSvc(mockGitSvc)

			ctx := context.Background()
			actualResult, actualErr := service.GetDiffStats(ctx, tc.inputSession, tc.inputRepoID, tc.inputPathToPath)

			if tc.expectedErr != nil {
				assert.Error(t, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}

			assert.Equal(t, tc.expectedResult, actualResult)
		})
	}
}

// ====================== Commits 方法测试 ======================
// CommitsTestCase 表格驱动用例结构体
type CommitsTestCase struct {
	name string
	// 前置条件
	mockPRStoreGet func(m *prStoreMock.MockInterface)
	mockRepoSvc    func(m *repoSvcMock.MockInterface)
	mockGitSvc     func(m *gitMock.MockInterface)
	inputRepoID    uint
	inputPRNumber  uint
	inputPage      uint
	inputLimit     uint
	inputSession   *auth.Session
	// 预期结果
	expectedCommitCount int
	expectedTotal       int
	expectedErr         error
}

// TestCommits Commits 单测入口
func TestCommits(t *testing.T) {
	t.Parallel()

	testCases := []CommitsTestCase{
		{
			name: "场景1：正常查询 - 获取PR提交列表",
			mockPRStoreGet: func(m *prStoreMock.MockInterface) {
				mockPR := &repoModel.PullRequest{
					ID:           1,
					RepoID:       100,
					SourceSHA:    "abc123",
					MergeBaseSHA: "def456",
				}
				m.EXPECT().GetByID(gomock.Any(), uint(100), uint(1)).Return(mockPR, nil)
			},
			mockRepoSvc: func(m *repoSvcMock.MockInterface) {
				mockRepo := &repoModel.Repo{
					ID:     100,
					GitUID: "repo-uid-100",
				}
				m.EXPECT().GetRepoCheckAccessByID(
					gomock.Any(),
					gomock.Any(),
					uint(100),
					[]constant.Permission{constant.PermissionRepoView},
				).Return(mockRepo, nil)
			},
			mockGitSvc: func(m *gitMock.MockInterface) {
				mockOutput := &git.ListCommitsOutput{
					Commits: []git.Commit{
						{
							SHA:     sha.Must("abc123def456abc123def456abc123def456abc1"),
							Title:   "First commit",
							Message: "First commit message",
						},
						{
							SHA:     sha.Must("def456abc123def456abc123def456abc123def4"),
							Title:   "Second commit",
							Message: "Second commit message",
						},
					},
					TotalCommits: 5,
				}
				m.EXPECT().ListCommits(gomock.Any(), gomock.Any()).Return(mockOutput, nil)
			},
			inputRepoID:   100,
			inputPRNumber: 1,
			inputPage:     1,
			inputLimit:    10,
			inputSession: &auth.Session{
				User: &principal.User{ID: 1, Account: "testuser"},
			},
			expectedCommitCount: 2,
			expectedTotal:       5,
			expectedErr:         nil,
		},
		{
			name: "场景2：PR不存在",
			mockPRStoreGet: func(m *prStoreMock.MockInterface) {
				m.EXPECT().GetByID(gomock.Any(), uint(100), uint(999)).Return(nil, ErrDBNotFound)
			},
			mockRepoSvc: func(m *repoSvcMock.MockInterface) {
				// PR不存在时不会调用repoSvc
			},
			mockGitSvc: func(m *gitMock.MockInterface) {
				// PR不存在时不会调用gitSvc
			},
			inputRepoID:   100,
			inputPRNumber: 999,
			inputPage:     1,
			inputLimit:    10,
			inputSession: &auth.Session{
				User: &principal.User{ID: 1, Account: "testuser"},
			},
			expectedCommitCount: 0,
			expectedTotal:       0,
			expectedErr:         ErrDBNotFound,
		},
		{
			name: "场景3：权限校验失败",
			mockPRStoreGet: func(m *prStoreMock.MockInterface) {
				mockPR := &repoModel.PullRequest{
					ID:           1,
					RepoID:       100,
					SourceSHA:    "abc123",
					MergeBaseSHA: "def456",
				}
				m.EXPECT().GetByID(gomock.Any(), uint(100), uint(1)).Return(mockPR, nil)
			},
			mockRepoSvc: func(m *repoSvcMock.MockInterface) {
				m.EXPECT().GetRepoCheckAccessByID(
					gomock.Any(),
					gomock.Any(),
					uint(100),
					[]constant.Permission{constant.PermissionRepoView},
				).Return(nil, ErrAccessDenied)
			},
			mockGitSvc: func(m *gitMock.MockInterface) {
				// 权限校验失败时不会调用gitSvc
			},
			inputRepoID:   100,
			inputPRNumber: 1,
			inputPage:     1,
			inputLimit:    10,
			inputSession: &auth.Session{
				User: &principal.User{ID: 1, Account: "testuser"},
			},
			expectedCommitCount: 0,
			expectedTotal:       0,
			expectedErr:         ErrAccessDenied,
		},
		{
			name: "场景4：Git操作失败",
			mockPRStoreGet: func(m *prStoreMock.MockInterface) {
				mockPR := &repoModel.PullRequest{
					ID:           1,
					RepoID:       100,
					SourceSHA:    "abc123",
					MergeBaseSHA: "def456",
				}
				m.EXPECT().GetByID(gomock.Any(), uint(100), uint(1)).Return(mockPR, nil)
			},
			mockRepoSvc: func(m *repoSvcMock.MockInterface) {
				mockRepo := &repoModel.Repo{
					ID:     100,
					GitUID: "repo-uid-100",
				}
				m.EXPECT().GetRepoCheckAccessByID(
					gomock.Any(),
					gomock.Any(),
					uint(100),
					[]constant.Permission{constant.PermissionRepoView},
				).Return(mockRepo, nil)
			},
			mockGitSvc: func(m *gitMock.MockInterface) {
				m.EXPECT().ListCommits(gomock.Any(), gomock.Any()).Return(nil, ErrGitOperation)
			},
			inputRepoID:   100,
			inputPRNumber: 1,
			inputPage:     1,
			inputLimit:    10,
			inputSession: &auth.Session{
				User: &principal.User{ID: 1, Account: "testuser"},
			},
			expectedCommitCount: 0,
			expectedTotal:       0,
			expectedErr:         ErrGitOperation,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPRStore := prStoreMock.NewMockInterface(ctrl)
			mockRepoSvc := repoSvcMock.NewMockInterface(ctrl)
			mockGitSvc := gitMock.NewMockInterface(ctrl)

			service := &Service{
				gitSvc:           mockGitSvc,
				repoSvc:          mockRepoSvc,
				pullRequestStore: mockPRStore,
			}

			tc.mockPRStoreGet(mockPRStore)
			tc.mockRepoSvc(mockRepoSvc)
			tc.mockGitSvc(mockGitSvc)

			ctx := context.Background()
			actualCommits, actualTotal, actualErr := service.Commits(ctx, tc.inputSession, tc.inputRepoID, tc.inputPRNumber, tc.inputPage, tc.inputLimit)

			if tc.expectedErr != nil {
				assert.Error(t, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}

			assert.Equal(t, tc.expectedCommitCount, len(actualCommits))
			assert.Equal(t, tc.expectedTotal, actualTotal)
		})
	}
}

// ====================== parseDiffPath 函数测试 ======================
// ParseDiffPathTestCase 表格驱动用例结构体
type ParseDiffPathTestCase struct {
	name           string
	inputPath      string
	expectedResult CompareInfo
	expectError    bool
}

// TestParseDiffPath parseDiffPath 单测入口
func TestParseDiffPath(t *testing.T) {
	t.Parallel()

	testCases := []ParseDiffPathTestCase{
		{
			name:      "场景1：三点分隔符 - merge base比较",
			inputPath: "main...feature",
			expectedResult: CompareInfo{
				BaseRef:   "main",
				HeadRef:   "feature",
				MergeBase: true,
			},
			expectError: false,
		},
		{
			name:      "场景2：两点分隔符 - 直接比较",
			inputPath: "main..feature",
			expectedResult: CompareInfo{
				BaseRef:   "main",
				HeadRef:   "feature",
				MergeBase: false,
			},
			expectError: false,
		},
		{
			name:      "场景3：带斜杠的分支名 - 三点分隔符",
			inputPath: "origin/main...origin/feature",
			expectedResult: CompareInfo{
				BaseRef:   "origin/main",
				HeadRef:   "origin/feature",
				MergeBase: true,
			},
			expectError: false,
		},
		{
			name:      "场景4：带斜杠的分支名 - 两点分隔符",
			inputPath: "release/v1.0..release/v2.0",
			expectedResult: CompareInfo{
				BaseRef:   "release/v1.0",
				HeadRef:   "release/v2.0",
				MergeBase: false,
			},
			expectError: false,
		},
		{
			name:           "场景5：无效格式 - 无分隔符",
			inputPath:      "invalid-path",
			expectedResult: CompareInfo{},
			expectError:    true,
		},
		{
			name:           "场景6：无效格式 - 单点分隔符",
			inputPath:      "main.feature",
			expectedResult: CompareInfo{},
			expectError:    true,
		},
		{
			name:      "场景7：SHA比较 - 三点分隔符",
			inputPath: "abc123...def456",
			expectedResult: CompareInfo{
				BaseRef:   "abc123",
				HeadRef:   "def456",
				MergeBase: true,
			},
			expectError: false,
		},
		{
			name:      "场景8：空baseRef - 两点分隔符",
			inputPath: "..feature",
			expectedResult: CompareInfo{
				BaseRef:   "",
				HeadRef:   "feature",
				MergeBase: false,
			},
			expectError: false,
		},
		{
			name:      "场景9：空headRef - 两点分隔符",
			inputPath: "main..",
			expectedResult: CompareInfo{
				BaseRef:   "main",
				HeadRef:   "",
				MergeBase: false,
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualResult, actualErr := parseDiffPath(tc.inputPath)

			if tc.expectError {
				assert.Error(t, actualErr)
			} else {
				assert.NoError(t, actualErr)
				assert.Equal(t, tc.expectedResult, actualResult)
			}
		})
	}
}
