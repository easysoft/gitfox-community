// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package publickey

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/model/auth"
	publickeyModel "github.com/easysoft/gitfox-community/app/model/publickey"
	principal "github.com/easysoft/gitfox-community/app/model/user"
	storePublicKey "github.com/easysoft/gitfox-community/app/store/publickey"
	publickeyStoreMock "github.com/easysoft/gitfox-community/app/store/publickey/mock"
	"github.com/easysoft/gitfox-community/pkg/database"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
)

type fakeUserStore struct {
	findByAccount func(ctx context.Context, account string) (*principal.User, error)
}

func (f *fakeUserStore) FindByID(ctx context.Context, id int64) (*principal.User, error) {
	return nil, nil
}

func (f *fakeUserStore) FindByAccount(ctx context.Context, account string) (*principal.User, error) {
	if f.findByAccount != nil {
		return f.findByAccount(ctx, account)
	}
	return &principal.User{Account: account}, nil
}

func (f *fakeUserStore) FindByEmail(ctx context.Context, email string) (*principal.User, error) {
	return nil, nil
}

func newSession(account string) *auth.Session {
	return &auth.Session{User: &principal.User{Account: account}}
}

func TestServiceCreate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		params    *publickeyModel.AddPublicKeyInput
		setupMock func(m *publickeyStoreMock.MockInterface)
		wantErr   bool
	}{
		{
			name: "create success",
			params: &publickeyModel.AddPublicKeyInput{
				Identifier: "key-1",
				Content:    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIKRHkkiUQkU0nkdkSOWuU4+8IMWLoeM8XdrjHDSj7zSu user@test",
			},
			setupMock: func(m *publickeyStoreMock.MockInterface) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, key storePublicKey.PublicKey) error {
					assert.Equal(t, "tester", key.Account)
					assert.Equal(t, "key-1", key.Identifier)
					assert.NotEmpty(t, key.Fingerprint)
					assert.Equal(t, "ssh-ed25519", key.Type)
					assert.Equal(t, "user@test", key.Comment)
					return nil
				})
			},
		},
		{
			name: "parse content failed",
			params: &publickeyModel.AddPublicKeyInput{
				Identifier: "key-1",
				Content:    "invalid-key-content",
			},
			setupMock: func(m *publickeyStoreMock.MockInterface) {},
			wantErr:   true,
		},
		{
			name: "store create failed",
			params: &publickeyModel.AddPublicKeyInput{
				Identifier: "key-1",
				Content:    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIKRHkkiUQkU0nkdkSOWuU4+8IMWLoeM8XdrjHDSj7zSu user@test",
			},
			setupMock: func(m *publickeyStoreMock.MockInterface) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := publickeyStoreMock.NewMockInterface(ctrl)
			tc.setupMock(mockStore)

			svc := Service{publicKeyStore: mockStore, userStore: &fakeUserStore{}}
			err := svc.Create(context.Background(), newSession("tester"), tc.params)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestServiceUpdate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		id        int64
		params    *publickeyModel.UpdatePublicKeyInput
		setupMock func(m *publickeyStoreMock.MockInterface)
		wantErr   bool
	}{
		{
			name: "update identifier and content success",
			id:   1001,
			params: &publickeyModel.UpdatePublicKeyInput{
				Identifier: "new-key",
				Content:    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIKRHkkiUQkU0nkdkSOWuU4+8IMWLoeM8XdrjHDSj7zSu user@test",
			},
			setupMock: func(m *publickeyStoreMock.MockInterface) {
				m.EXPECT().UpdateByID(gomock.Any(), "tester", int64(1001), gomock.Any()).DoAndReturn(func(ctx context.Context, account string, id int64, key storePublicKey.PublicKey) error {
					assert.Equal(t, "new-key", key.Identifier)
					assert.Equal(t, "ssh-ed25519", key.Type)
					assert.NotEmpty(t, key.Fingerprint)
					assert.Equal(t, "user@test", key.Comment)
					return nil
				})
			},
		},
		{
			name: "update identifier only success",
			id:   1002,
			params: &publickeyModel.UpdatePublicKeyInput{
				Identifier: "rename-only",
			},
			setupMock: func(m *publickeyStoreMock.MockInterface) {
				m.EXPECT().UpdateByID(gomock.Any(), "tester", int64(1002), gomock.Any()).DoAndReturn(func(ctx context.Context, account string, id int64, key storePublicKey.PublicKey) error {
					assert.Equal(t, "rename-only", key.Identifier)
					assert.Empty(t, key.Content)
					assert.Empty(t, key.Fingerprint)
					assert.Empty(t, key.Type)
					return nil
				})
			},
		},
		{
			name: "update content only success",
			id:   1003,
			params: &publickeyModel.UpdatePublicKeyInput{
				Content: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIKRHkkiUQkU0nkdkSOWuU4+8IMWLoeM8XdrjHDSj7zSu content-only@test",
			},
			setupMock: func(m *publickeyStoreMock.MockInterface) {
				m.EXPECT().UpdateByID(gomock.Any(), "tester", int64(1003), gomock.Any()).DoAndReturn(func(ctx context.Context, account string, id int64, key storePublicKey.PublicKey) error {
					assert.Empty(t, key.Identifier)
					assert.Equal(t, "ssh-ed25519", key.Type)
					assert.NotEmpty(t, key.Fingerprint)
					assert.Equal(t, "content-only@test", key.Comment)
					assert.NotEmpty(t, key.Content)
					return nil
				})
			},
		},
		{
			name: "parse content failed",
			id:   1001,
			params: &publickeyModel.UpdatePublicKeyInput{
				Identifier: "new-key",
				Content:    "invalid-key-content",
			},
			setupMock: func(m *publickeyStoreMock.MockInterface) {},
			wantErr:   true,
		},
		{
			name: "store update failed",
			id:   1001,
			params: &publickeyModel.UpdatePublicKeyInput{
				Identifier: "new-key",
				Content:    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIKRHkkiUQkU0nkdkSOWuU4+8IMWLoeM8XdrjHDSj7zSu user@test",
			},
			setupMock: func(m *publickeyStoreMock.MockInterface) {
				m.EXPECT().UpdateByID(gomock.Any(), "tester", int64(1001), gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := publickeyStoreMock.NewMockInterface(ctrl)
			tc.setupMock(mockStore)

			svc := Service{publicKeyStore: mockStore, userStore: &fakeUserStore{}}
			err := svc.Update(context.Background(), newSession("tester"), tc.id, tc.params)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestServiceDelete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		id        int64
		setupMock func(m *publickeyStoreMock.MockInterface)
		wantErr   bool
	}{
		{
			name: "delete success",
			id:   1001,
			setupMock: func(m *publickeyStoreMock.MockInterface) {
				m.EXPECT().DeleteByID(gomock.Any(), "tester", int64(1001)).Return(nil)
			},
		},
		{
			name: "delete failed",
			id:   1001,
			setupMock: func(m *publickeyStoreMock.MockInterface) {
				m.EXPECT().DeleteByID(gomock.Any(), "tester", int64(1001)).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := publickeyStoreMock.NewMockInterface(ctrl)
			tc.setupMock(mockStore)

			svc := Service{publicKeyStore: mockStore, userStore: &fakeUserStore{}}
			err := svc.Delete(context.Background(), newSession("tester"), tc.id)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestServiceList(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := publickeyStoreMock.NewMockInterface(ctrl)
	expected := &database.PageResult[*publickeyModel.PublicKey]{
		Total:    1,
		Page:     1,
		PageSize: 20,
		Data: []*publickeyModel.PublicKey{
			{Identifier: "key-1", Type: "ssh-ed25519"},
		},
	}

	mockStore.EXPECT().ListByAccount(gomock.Any(), "tester", req.ListFilter{Page: 1, PageSize: 20}).Return(expected, nil)

	svc := Service{publicKeyStore: mockStore, userStore: &fakeUserStore{}}
	result, err := svc.List(context.Background(), newSession("tester"), req.ListFilter{Page: 1, PageSize: 20})
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestServiceGetByID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		id        int64
		setupMock func(m *publickeyStoreMock.MockInterface)
		wantErr   bool
	}{
		{
			name: "get by id success",
			id:   1001,
			setupMock: func(m *publickeyStoreMock.MockInterface) {
				created := datetime.Local(time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC))
				verified := datetime.Local(time.Date(2026, 3, 1, 10, 1, 1, 0, time.UTC))
				m.EXPECT().GetByID(gomock.Any(), "tester", int64(1001)).Return(&publickeyModel.PublicKey{
					Identifier:   "key-1",
					Type:         "ssh-ed25519",
					CreatedDate:  &created,
					VerifiedDate: &verified,
				}, nil)
			},
		},
		{
			name: "get by id failed",
			id:   1001,
			setupMock: func(m *publickeyStoreMock.MockInterface) {
				m.EXPECT().GetByID(gomock.Any(), "tester", int64(1001)).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			gin.SetMode(gin.TestMode)
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest("GET", "/", nil)

			mockStore := publickeyStoreMock.NewMockInterface(ctrl)
			tc.setupMock(mockStore)

			svc := Service{publicKeyStore: mockStore, userStore: &fakeUserStore{}}
			key, err := svc.GetByID(ctx, newSession("tester"), tc.id)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, "key-1", key.Identifier)
			assert.NotEmpty(t, key.LastUsedDate)
		})
	}
}
