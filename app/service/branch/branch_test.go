package branch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	branchModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	gitMock "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/branch/mock"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/sha"
	urlMock "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url/mock"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	testSHA     = sha.Must("aabbccddee0011223344556677889900aabbccdd")
	errNotFound = errors.New("branch not found")
)

func newTestRepo(defaultBranch string) *repoModel.Repo {
	return &repoModel.Repo{
		ID:            1,
		GitUID:        "test-repo-uid",
		DefaultBranch: defaultBranch,
	}
}

func newTestSession() *authModel.Session {
	return &authModel.Session{
		User: &principal.User{
			ID:      1,
			Account: "tester",
			Email:   "tester@test.com",
		},
	}
}

func newTestGinContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c
}

func setupService(ctrl *gomock.Controller) (*Service, *gitMock.MockInterface) {
	mockGit := gitMock.NewMockInterface(ctrl)
	mockURL := urlMock.NewMockProvider(ctrl)
	mockURL.EXPECT().GetInternalAPIURL(gomock.Any()).Return("http://localhost").AnyTimes()
	svc := &Service{gitPkg: mockGit, urlProvider: mockURL}
	return svc, mockGit
}

// --- BatchArchive Tests ---

func TestBatchArchive_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockGit := setupService(ctrl)
	repo := newTestRepo("main")
	session := newTestSession()
	ginCtx := newTestGinContext()

	// 1. 检查 archived/dev 不存在
	checkGet := mockGit.EXPECT().GetBranch(gomock.Any(), &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: "test-repo-uid"},
		BranchName: "archived/dev",
	}).Return(nil, errNotFound)
	// 2. CreateBranch archived/dev
	createCall := mockGit.EXPECT().CreateBranch(gomock.Any(), gomock.Any()).Return(
		&git.CreateBranchOutput{Branch: git.Branch{Name: "archived/dev", SHA: testSHA}}, nil,
	).After(checkGet)
	// 3. GetBranch dev (for SHA)
	getCall := mockGit.EXPECT().GetBranch(gomock.Any(), &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: "test-repo-uid"},
		BranchName: "dev",
	}).Return(
		&git.GetBranchOutput{Branch: git.Branch{Name: "dev", SHA: testSHA}}, nil,
	).After(createCall)
	// 4. DeleteBranch dev
	mockGit.EXPECT().DeleteBranch(gomock.Any(), gomock.Any()).Return(nil).After(getCall)

	req := branchModel.ReqForArchive{BranchNames: []string{"dev"}}
	results := svc.BatchArchive(ginCtx, context.Background(), req, session, repo)

	assert.Len(t, results, 1)
	assert.True(t, results[0].Status)
	assert.Equal(t, "dev", results[0].BranchName)
	assert.Contains(t, results[0].Message, "archive_success")
}

func TestBatchArchive_DefaultBranch(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _ := setupService(ctrl)
	repo := newTestRepo("main")
	session := newTestSession()
	ginCtx := newTestGinContext()

	req := branchModel.ReqForArchive{BranchNames: []string{"main"}}
	results := svc.BatchArchive(ginCtx, context.Background(), req, session, repo)

	assert.Len(t, results, 1)
	assert.False(t, results[0].Status)
	assert.Contains(t, results[0].Message, "archive_is_default")
}

func TestBatchArchive_AlreadyArchived(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _ := setupService(ctrl)
	repo := newTestRepo("main")
	session := newTestSession()
	ginCtx := newTestGinContext()

	req := branchModel.ReqForArchive{BranchNames: []string{"archived/feature"}}
	results := svc.BatchArchive(ginCtx, context.Background(), req, session, repo)

	assert.Len(t, results, 1)
	assert.True(t, results[0].Status)
	assert.Contains(t, results[0].Message, "archive_success")
}

func TestBatchArchive_TargetExists(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockGit := setupService(ctrl)
	repo := newTestRepo("main")
	session := newTestSession()
	ginCtx := newTestGinContext()

	// archived/dev already exists
	mockGit.EXPECT().GetBranch(gomock.Any(), &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: "test-repo-uid"},
		BranchName: "archived/dev",
	}).Return(
		&git.GetBranchOutput{Branch: git.Branch{Name: "archived/dev", SHA: testSHA}}, nil,
	)

	req := branchModel.ReqForArchive{BranchNames: []string{"dev"}}
	results := svc.BatchArchive(ginCtx, context.Background(), req, session, repo)

	assert.Len(t, results, 1)
	assert.False(t, results[0].Status)
	assert.Contains(t, results[0].Message, "archive_target_exists")
}

func TestBatchArchive_CreateFailed(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockGit := setupService(ctrl)
	repo := newTestRepo("main")
	session := newTestSession()
	ginCtx := newTestGinContext()

	// archived/dev does not exist
	checkGet := mockGit.EXPECT().GetBranch(gomock.Any(), &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: "test-repo-uid"},
		BranchName: "archived/dev",
	}).Return(nil, errNotFound)
	// CreateBranch fails
	mockGit.EXPECT().CreateBranch(gomock.Any(), gomock.Any()).Return(
		nil, errors.New("internal error"),
	).After(checkGet)

	req := branchModel.ReqForArchive{BranchNames: []string{"dev"}}
	results := svc.BatchArchive(ginCtx, context.Background(), req, session, repo)

	assert.Len(t, results, 1)
	assert.False(t, results[0].Status)
	assert.Contains(t, results[0].Message, "archive_create_error")
}

func TestBatchArchive_DeleteFailedWithRollback(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockGit := setupService(ctrl)
	repo := newTestRepo("main")
	session := newTestSession()
	ginCtx := newTestGinContext()

	// 1. Check archived/dev not exist
	checkGet := mockGit.EXPECT().GetBranch(gomock.Any(), &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: "test-repo-uid"},
		BranchName: "archived/dev",
	}).Return(nil, errNotFound)
	// 2. CreateBranch archived/dev succeeds
	createCall := mockGit.EXPECT().CreateBranch(gomock.Any(), gomock.Any()).Return(
		&git.CreateBranchOutput{Branch: git.Branch{Name: "archived/dev", SHA: testSHA}}, nil,
	).After(checkGet)
	// 3. GetBranch dev (for delete SHA)
	getDevCall := mockGit.EXPECT().GetBranch(gomock.Any(), &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: "test-repo-uid"},
		BranchName: "dev",
	}).Return(
		&git.GetBranchOutput{Branch: git.Branch{Name: "dev", SHA: testSHA}}, nil,
	).After(createCall)
	// 4. DeleteBranch dev fails
	deleteCall := mockGit.EXPECT().DeleteBranch(gomock.Any(), gomock.Any()).Return(
		errors.New("delete failed"),
	).After(getDevCall)
	// 5. Rollback: GetBranch archived/dev
	rollbackGet := mockGit.EXPECT().GetBranch(gomock.Any(), &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: "test-repo-uid"},
		BranchName: "archived/dev",
	}).Return(
		&git.GetBranchOutput{Branch: git.Branch{Name: "archived/dev", SHA: testSHA}}, nil,
	).After(deleteCall)
	// 6. Rollback: DeleteBranch archived/dev succeeds
	mockGit.EXPECT().DeleteBranch(gomock.Any(), gomock.Any()).Return(nil).After(rollbackGet)

	req := branchModel.ReqForArchive{BranchNames: []string{"dev"}}
	results := svc.BatchArchive(ginCtx, context.Background(), req, session, repo)

	assert.Len(t, results, 1)
	assert.False(t, results[0].Status)
	assert.Contains(t, results[0].Message, "archive_delete_error")
}

func TestBatchArchive_PartialFailContinues(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockGit := setupService(ctrl)
	repo := newTestRepo("main")
	session := newTestSession()
	ginCtx := newTestGinContext()

	// main is default → fail without git calls
	// dev → succeed: check + create + get + delete
	checkGet := mockGit.EXPECT().GetBranch(gomock.Any(), &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: "test-repo-uid"},
		BranchName: "archived/dev",
	}).Return(nil, errNotFound)
	createCall := mockGit.EXPECT().CreateBranch(gomock.Any(), gomock.Any()).Return(
		&git.CreateBranchOutput{Branch: git.Branch{Name: "archived/dev", SHA: testSHA}}, nil,
	).After(checkGet)
	getCall := mockGit.EXPECT().GetBranch(gomock.Any(), &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: "test-repo-uid"},
		BranchName: "dev",
	}).Return(
		&git.GetBranchOutput{Branch: git.Branch{Name: "dev", SHA: testSHA}}, nil,
	).After(createCall)
	mockGit.EXPECT().DeleteBranch(gomock.Any(), gomock.Any()).Return(nil).After(getCall)

	req := branchModel.ReqForArchive{BranchNames: []string{"main", "dev"}}
	results := svc.BatchArchive(ginCtx, context.Background(), req, session, repo)

	assert.Len(t, results, 2)
	assert.False(t, results[0].Status, "default branch should fail")
	assert.True(t, results[1].Status, "dev should succeed")
}

// --- BatchActivate Tests ---

func TestBatchActivate_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockGit := setupService(ctrl)
	repo := newTestRepo("main")
	session := newTestSession()
	ginCtx := newTestGinContext()

	// 1. Check dev not exist
	checkGet := mockGit.EXPECT().GetBranch(gomock.Any(), &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: "test-repo-uid"},
		BranchName: "dev",
	}).Return(nil, errNotFound)
	// 2. CreateBranch dev
	createCall := mockGit.EXPECT().CreateBranch(gomock.Any(), gomock.Any()).Return(
		&git.CreateBranchOutput{Branch: git.Branch{Name: "dev", SHA: testSHA}}, nil,
	).After(checkGet)
	// 3. GetBranch archived/dev (for SHA)
	getCall := mockGit.EXPECT().GetBranch(gomock.Any(), &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: "test-repo-uid"},
		BranchName: "archived/dev",
	}).Return(
		&git.GetBranchOutput{Branch: git.Branch{Name: "archived/dev", SHA: testSHA}}, nil,
	).After(createCall)
	// 4. DeleteBranch archived/dev
	mockGit.EXPECT().DeleteBranch(gomock.Any(), gomock.Any()).Return(nil).After(getCall)

	req := branchModel.ReqForArchive{BranchNames: []string{"archived/dev"}}
	results := svc.BatchActivate(ginCtx, context.Background(), req, session, repo)

	assert.Len(t, results, 1)
	assert.True(t, results[0].Status)
	assert.Equal(t, "archived/dev", results[0].BranchName)
	assert.Contains(t, results[0].Message, "activate_success")
}

func TestBatchActivate_NotArchived(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _ := setupService(ctrl)
	repo := newTestRepo("main")
	session := newTestSession()
	ginCtx := newTestGinContext()

	req := branchModel.ReqForArchive{BranchNames: []string{"dev"}}
	results := svc.BatchActivate(ginCtx, context.Background(), req, session, repo)

	assert.Len(t, results, 1)
	assert.False(t, results[0].Status)
	assert.Contains(t, results[0].Message, "activate_not_archived")
}

func TestBatchActivate_TargetExists(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockGit := setupService(ctrl)
	repo := newTestRepo("main")
	session := newTestSession()
	ginCtx := newTestGinContext()

	// dev already exists
	mockGit.EXPECT().GetBranch(gomock.Any(), &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: "test-repo-uid"},
		BranchName: "dev",
	}).Return(
		&git.GetBranchOutput{Branch: git.Branch{Name: "dev", SHA: testSHA}}, nil,
	)

	req := branchModel.ReqForArchive{BranchNames: []string{"archived/dev"}}
	results := svc.BatchActivate(ginCtx, context.Background(), req, session, repo)

	assert.Len(t, results, 1)
	assert.False(t, results[0].Status)
	assert.Contains(t, results[0].Message, "activate_target_exists")
}

func TestBatchActivate_CreateFailed(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockGit := setupService(ctrl)
	repo := newTestRepo("main")
	session := newTestSession()
	ginCtx := newTestGinContext()

	// dev does not exist
	checkGet := mockGit.EXPECT().GetBranch(gomock.Any(), &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: "test-repo-uid"},
		BranchName: "dev",
	}).Return(nil, errNotFound)
	// CreateBranch fails
	mockGit.EXPECT().CreateBranch(gomock.Any(), gomock.Any()).Return(
		nil, errors.New("internal error"),
	).After(checkGet)

	req := branchModel.ReqForArchive{BranchNames: []string{"archived/dev"}}
	results := svc.BatchActivate(ginCtx, context.Background(), req, session, repo)

	assert.Len(t, results, 1)
	assert.False(t, results[0].Status)
	assert.Contains(t, results[0].Message, "activate_create_error")
}

func TestBatchActivate_DeleteFailedNoRollback(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockGit := setupService(ctrl)
	repo := newTestRepo("main")
	session := newTestSession()
	ginCtx := newTestGinContext()

	// 1. Check dev not exist
	checkGet := mockGit.EXPECT().GetBranch(gomock.Any(), &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: "test-repo-uid"},
		BranchName: "dev",
	}).Return(nil, errNotFound)
	// 2. CreateBranch dev succeeds
	createCall := mockGit.EXPECT().CreateBranch(gomock.Any(), gomock.Any()).Return(
		&git.CreateBranchOutput{Branch: git.Branch{Name: "dev", SHA: testSHA}}, nil,
	).After(checkGet)
	// 3. GetBranch archived/dev
	getCall := mockGit.EXPECT().GetBranch(gomock.Any(), &git.GetBranchParams{
		ReadParams: git.ReadParams{RepoUID: "test-repo-uid"},
		BranchName: "archived/dev",
	}).Return(
		&git.GetBranchOutput{Branch: git.Branch{Name: "archived/dev", SHA: testSHA}}, nil,
	).After(createCall)
	// 4. DeleteBranch archived/dev fails
	mockGit.EXPECT().DeleteBranch(gomock.Any(), gomock.Any()).Return(
		errors.New("delete failed"),
	).After(getCall)

	req := branchModel.ReqForArchive{BranchNames: []string{"archived/dev"}}
	results := svc.BatchActivate(ginCtx, context.Background(), req, session, repo)

	assert.Len(t, results, 1)
	assert.False(t, results[0].Status)
	assert.Contains(t, results[0].Message, "activate_delete_error")
}
