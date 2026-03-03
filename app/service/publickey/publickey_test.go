package publickey

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	publickeyModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/publickey"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	storePublicKey "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/publickey"
	publickeyStoreMock "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/publickey/mock"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
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

func (f *fakeUserStore) CompareHashAndPassword(ctx context.Context, hashedPassword, password []byte) error {
	return nil
}

func (f *fakeUserStore) IssueUserPAT(ctx context.Context, user *principal.User, identifier string, lifetime *time.Duration) (*auth.TokenResponse, error) {
	return nil, nil
}

func (f *fakeUserStore) IssueAppPAT(ctx context.Context, user *principal.User, identifier string, lifetime *time.Duration) (*auth.TokenResponse, error) {
	return nil, nil
}

func (f *fakeUserStore) IssueAdminPAT() (string, error) {
	return "", nil
}

func newSession(account string) *auth.Session {
	return &auth.Session{User: &principal.User{Account: account}}
}

func TestServiceCreatePublicKey(t *testing.T) {
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
			err := svc.CreatePublicKey(context.Background(), newSession("tester"), tc.params)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestServiceUpdatePublicKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		id        int64
		params    *publickeyModel.UpdatePublicKeyInput
		setupMock func(m *publickeyStoreMock.MockInterface)
		wantErr   bool
	}{
		{
			name: "update success",
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
			err := svc.UpdatePublicKey(context.Background(), newSession("tester"), tc.id, tc.params)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestServiceDeletePublicKey(t *testing.T) {
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
			err := svc.DeletePublicKey(context.Background(), newSession("tester"), tc.id)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestServiceListPublicKeys(t *testing.T) {
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
	result, err := svc.ListPublicKeys(context.Background(), newSession("tester"), req.ListFilter{Page: 1, PageSize: 20})
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
