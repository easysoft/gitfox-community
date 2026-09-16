package importer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/easysoft/gitfox-community/app/constant"
	jobModel "github.com/easysoft/gitfox-community/app/model/jobs"
	providerModel "github.com/easysoft/gitfox-community/app/model/provider"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	"github.com/easysoft/gitfox-community/app/service/encrypt"
	jobsSvc "github.com/easysoft/gitfox-community/app/service/jobs"
	jobsStoreMock "github.com/easysoft/gitfox-community/app/store/jobs/mock"
	providerStoreMock "github.com/easysoft/gitfox-community/app/store/provider/mock"
	repoStoreMock "github.com/easysoft/gitfox-community/app/store/repo/mock"
	pkgErrors "github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/lock"
	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/easysoft/gitfox-community/pkg/pubsub"
	svnPkg "github.com/easysoft/gitfox-community/pkg/svn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestBuildProviderFromRepo(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		repo        *repoModel.Repo
		provider    *providerModel.Provider
		assertFunc  func(t *testing.T, actual repoModel.Provider)
		expectError string
	}

	svnPassword := base64.StdEncoding.EncodeToString([]byte("secret"))
	svnConnector, err := json.Marshal(repoModel.SVNConnectInfo{Slug: "repo-a", User: "alice", Password: svnPassword})
	require.NoError(t, err)
	gitConnector, err := json.Marshal(repoModel.GitConnectInfo{Slug: "team/repo-a", ProjectID: "123"})
	require.NoError(t, err)

	testCases := []testCase{
		{
			name: "成功-Git provider payload 构建",
			repo: &repoModel.Repo{
				SCMType:   constant.RepoScmTypeGit,
				Connector: gitConnector,
			},
			provider: &providerModel.Provider{ID: 1, Type: string(constant.RepoProviderTypeGitLab), URL: "https://git.example.com", Token: "token-a"},
			assertFunc: func(t *testing.T, actual repoModel.Provider) {
				assert.Equal(t, constant.RepoProviderTypeGitLab, actual.Type)
				assert.Equal(t, uint(1), actual.ID)
				assert.Equal(t, "https://git.example.com", actual.Host)
				assert.Equal(t, "token-a", actual.Token)
				assert.Equal(t, "team/repo-a", actual.Slug)
				assert.Equal(t, "123", actual.ProjectID)
			},
		},
		{
			name: "成功-Git provider payload 构建（projectID 为空）",
			repo: &repoModel.Repo{
				SCMType: constant.RepoScmTypeGit,
				Connector: mustJSON(t, repoModel.GitConnectInfo{
					Slug:      "team/repo-b",
					ProjectID: "",
				}),
			},
			provider: &providerModel.Provider{ID: 9, Type: string(constant.RepoProviderTypeGitea), URL: "https://gitea.example.com", Token: "token-b"},
			assertFunc: func(t *testing.T, actual repoModel.Provider) {
				assert.Equal(t, constant.RepoProviderTypeGitea, actual.Type)
				assert.Equal(t, uint(9), actual.ID)
				assert.Equal(t, "https://gitea.example.com", actual.Host)
				assert.Equal(t, "token-b", actual.Token)
				assert.Equal(t, "team/repo-b", actual.Slug)
				assert.Equal(t, "", actual.ProjectID)
			},
		},
		{
			name: "成功-SVN provider payload 构建并解码密码",
			repo: &repoModel.Repo{
				SCMType:   constant.RepoScmTypeSvn,
				Connector: svnConnector,
			},
			provider: &providerModel.Provider{ID: 2, Type: string(constant.RepoProviderTypeSVN), URL: "https://svn.example.com", Token: "unused"},
			assertFunc: func(t *testing.T, actual repoModel.Provider) {
				assert.Equal(t, constant.RepoProviderTypeSVN, actual.Type)
				assert.Equal(t, uint(2), actual.ID)
				assert.Equal(t, "https://svn.example.com", actual.Host)
				assert.Equal(t, "repo-a", actual.Slug)
				assert.Equal(t, "alice", actual.Username)
				assert.Equal(t, "secret", actual.Password)
			},
		},
		{
			name: "失败-Git connector 非法",
			repo: &repoModel.Repo{
				SCMType:   constant.RepoScmTypeGit,
				Connector: []byte("not-json"),
			},
			provider:    &providerModel.Provider{Type: string(constant.RepoProviderTypeGitLab), URL: "https://git.example.com"},
			expectError: "failed to unmarshal Git connect info",
		},
		{
			name: "失败-SVN 密码非法",
			repo: &repoModel.Repo{
				SCMType: constant.RepoScmTypeSvn,
				Connector: mustJSON(t, repoModel.SVNConnectInfo{
					Slug:     "repo-a",
					User:     "alice",
					Password: "not-base64",
				}),
			},
			provider:    &providerModel.Provider{Type: string(constant.RepoProviderTypeSVN), URL: "https://svn.example.com"},
			expectError: "failed to decode SVN password",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual, err := buildProviderFromRepo(tc.repo, tc.provider)
			if tc.expectError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectError)
				return
			}

			require.NoError(t, err)
			tc.assertFunc(t, actual)
		})
	}
}

func TestMigrate(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		setup     func(t *testing.T, ctrl *gomock.Controller) *Repository
		expectErr string
	}

	testCases := []testCase{
		{
			name: "失败-GetMigrate 返回错误",
			setup: func(t *testing.T, ctrl *gomock.Controller) *Repository {
				repoStore := repoStoreMock.NewMockInterface(ctrl)
				repoStore.EXPECT().GetMigrate(gomock.Any()).Return(nil, assert.AnError)
				return &Repository{repoStore: repoStore}
			},
			expectErr: assert.AnError.Error(),
		},
		{
			name: "成功-无迁移仓库直接返回",
			setup: func(t *testing.T, ctrl *gomock.Controller) *Repository {
				repoStore := repoStoreMock.NewMockInterface(ctrl)
				repoStore.EXPECT().GetMigrate(gomock.Any()).Return([]*repoModel.Repo{}, nil)
				return &Repository{repoStore: repoStore}
			},
		},
		{
			name: "失败-全部 provider 未找到",
			setup: func(t *testing.T, ctrl *gomock.Controller) *Repository {
				repoStore := repoStoreMock.NewMockInterface(ctrl)
				providerStore := providerStoreMock.NewMockInterface(ctrl)
				repos := []*repoModel.Repo{{ID: 11, ProviderID: 5, SCMType: constant.RepoScmTypeGit, Connector: mustJSON(t, repoModel.GitConnectInfo{Slug: "team/repo"})}}
				repoStore.EXPECT().GetMigrate(gomock.Any()).Return(repos, nil)
				providerStore.EXPECT().GetByIDs(gomock.Any(), []uint{5}).Return(map[uint]*providerModel.Provider{}, nil)
				return &Repository{repoStore: repoStore, providerStore: providerStore}
			},
			expectErr: "migrate scheduling completed all failed",
		},
		{
			name: "成功-SVN 仓库成功入队导入任务",
			setup: func(t *testing.T, ctrl *gomock.Controller) *Repository {
				repoStore := repoStoreMock.NewMockInterface(ctrl)
				providerStore := providerStoreMock.NewMockInterface(ctrl)
				jobStore := jobsStoreMock.NewMockInterface(ctrl)
				encrypter, err := encrypt.Provide("", false)
				require.NoError(t, err)

				scheduler, err := jobsSvc.NewScheduler(
					jobStore,
					nil,
					lock.NewInMemory(lock.Config{}),
					pubsub.NewMemory(),
					logger.NewNopLogger(),
					"test-instance",
					1,
					time.Minute,
				)
				require.NoError(t, err)

				repos := []*repoModel.Repo{{
					ID:         14,
					ProviderID: 5,
					SCMType:    constant.RepoScmTypeSvn,
					Status:     constant.RepoStateImporting,
					Mirror:     true,
					Connector: mustJSON(t, repoModel.SVNConnectInfo{
						Slug:     "repo-a",
						User:     "alice",
						Password: base64.StdEncoding.EncodeToString([]byte("secret")),
					}),
				}}
				repoStore.EXPECT().GetMigrate(gomock.Any()).Return(repos, nil)
				providerStore.EXPECT().GetByIDs(gomock.Any(), []uint{5}).Return(map[uint]*providerModel.Provider{
					5: {ID: 5, Type: string(constant.RepoProviderTypeSVN), URL: "https://svn.example.com"},
				}, nil)
				jobStore.EXPECT().GetByName(gomock.Any(), "import-repo-14").Return(nil, pkgErrors.ErrResourceNotFound)
				jobStore.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, job *jobModel.ScheduleJob) error {
					assert.Equal(t, "import-repo-14", job.Name)
					assert.Equal(t, JobType, job.Type)
					assert.Equal(t, jobModel.StateScheduled, job.State)
					assert.Equal(t, 2700, job.MaxDurationSeconds)
					return nil
				})

				return &Repository{
					repoStore:     repoStore,
					providerStore: providerStore,
					svn: &fakeSVNService{
						infoRevisionFn: func(_ context.Context, rawURL, username, password string) (string, error) {
							assert.Equal(t, "https://svn.example.com/repo-a", rawURL)
							assert.Equal(t, "alice", username)
							assert.Equal(t, "secret", password)
							return "HEAD", nil
						},
					},
					encrypter: encrypter,
					scheduler: scheduler,
				}
			},
		},
		{
			name: "成功-Git 镜像仓库成功入队导入任务",
			setup: func(t *testing.T, ctrl *gomock.Controller) *Repository {
				repoStore := repoStoreMock.NewMockInterface(ctrl)
				providerStore := providerStoreMock.NewMockInterface(ctrl)
				jobStore := jobsStoreMock.NewMockInterface(ctrl)
				scmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/api/v1/user":
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte(`{"login":"alice"}`))
					case "/api/v1/repos/team/repo-git-mirror":
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte(`{
							"name":"repo-git-mirror",
							"full_name":"team/repo-git-mirror",
							"default_branch":"main",
							"clone_url":"https://gitea.example.com/team/repo-git-mirror.git",
							"private":true,
							"owner":{"login":"team"}
						}`))
					default:
						http.NotFound(w, r)
					}
				}))
				t.Cleanup(scmServer.Close)
				encrypter, err := encrypt.Provide("", false)
				require.NoError(t, err)

				scheduler, err := jobsSvc.NewScheduler(
					jobStore,
					nil,
					lock.NewInMemory(lock.Config{}),
					pubsub.NewMemory(),
					logger.NewNopLogger(),
					"test-instance",
					1,
					time.Minute,
				)
				require.NoError(t, err)

				repos := []*repoModel.Repo{{
					ID:         15,
					ProviderID: 6,
					SCMType:    constant.RepoScmTypeGit,
					Status:     constant.RepoStateImporting,
					Mirror:     true,
					Connector: mustJSON(t, repoModel.GitConnectInfo{
						Slug:      "team/repo-git-mirror",
						ProjectID: "456",
					}),
				}}
				repoStore.EXPECT().GetMigrate(gomock.Any()).Return(repos, nil)
				providerStore.EXPECT().GetByIDs(gomock.Any(), []uint{6}).Return(map[uint]*providerModel.Provider{
					6: {
						ID:    6,
						Type:  string(constant.RepoProviderTypeGitea),
						URL:   scmServer.URL,
						Token: "token-git",
					},
				}, nil)
				jobStore.EXPECT().GetByName(gomock.Any(), "import-repo-15").Return(nil, pkgErrors.ErrResourceNotFound)
				jobStore.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, job *jobModel.ScheduleJob) error {
					assert.Equal(t, "import-repo-15", job.Name)
					assert.Equal(t, JobType, job.Type)
					assert.Equal(t, jobModel.StateScheduled, job.State)
					assert.Equal(t, 2700, job.MaxDurationSeconds)
					return nil
				})

				return &Repository{
					repoStore:     repoStore,
					providerStore: providerStore,
					svn:           &fakeSVNService{},
					encrypter:     encrypter,
					scheduler:     scheduler,
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := tc.setup(t, ctrl)
			ctx := logger.WithLogger(context.Background(), logger.NewNopLogger())
			err := service.Migrate(ctx)

			if tc.expectErr == "" {
				assert.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectErr)
		})
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	b, err := json.Marshal(value)
	require.NoError(t, err)
	return b
}

type fakeSVNService struct {
	infoRevisionFn func(ctx context.Context, url, username, password string) (string, error)
}

func (f *fakeSVNService) FileURL(svnUID string) string { return svnUID }

func (f *fakeSVNService) Info(ctx context.Context, params *svnPkg.InfoParams) (string, error) {
	return "", nil
}

func (f *fakeSVNService) InfoRevision(ctx context.Context, url, username, password string) (string, error) {
	if f.infoRevisionFn != nil {
		return f.infoRevisionFn(ctx, url, username, password)
	}
	return "", nil
}

func (f *fakeSVNService) GetLatestRevision(ctx context.Context, svnUID, username, password string) (string, error) {
	return "", nil
}

func (f *fakeSVNService) InfoXML(ctx context.Context, url, revision, username, password string) (string, error) {
	return "", nil
}

func (f *fakeSVNService) GetPathInfo(ctx context.Context, svnUID, revision, username, password string) (string, error) {
	return "", nil
}

func (f *fakeSVNService) Checkout(ctx context.Context, params *svnPkg.CheckoutParams) error {
	return nil
}

func (f *fakeSVNService) Update(ctx context.Context, params *svnPkg.UpdateParams) error { return nil }

func (f *fakeSVNService) CreateRepository(ctx context.Context, svnUID string) error { return nil }

func (f *fakeSVNService) RemoveRepository(svnUID string) error { return nil }

func (f *fakeSVNService) RemoveWorkingCopy(svnUID string) error { return nil }

func (f *fakeSVNService) SyncInit(ctx context.Context, params *svnPkg.SyncInitParams) error {
	return nil
}

func (f *fakeSVNService) Sync(ctx context.Context, params *svnPkg.SyncParams) error { return nil }

func (f *fakeSVNService) SyncCopyRevProps(ctx context.Context, params *svnPkg.SyncParams) error {
	return nil
}

func (f *fakeSVNService) Log(ctx context.Context, params *svnPkg.LogParams) (string, error) {
	return "", nil
}

func (f *fakeSVNService) ListAllCommits(ctx context.Context, svnUID string) (string, error) {
	return "", nil
}

func (f *fakeSVNService) Diff(ctx context.Context, params *svnPkg.DiffParams) (string, error) {
	return "", nil
}

func (f *fakeSVNService) Ls(ctx context.Context, params *svnPkg.LsParams) (string, error) {
	return "", nil
}

func (f *fakeSVNService) Cat(ctx context.Context, params *svnPkg.CatParams) (string, error) {
	return "", nil
}

func (f *fakeSVNService) Blame(ctx context.Context, params *svnPkg.BlameParams) (string, error) {
	return "", nil
}

func (f *fakeSVNService) Export(ctx context.Context, params *svnPkg.ExportParams, w io.Writer) (bool, error) {
	return false, nil
}
