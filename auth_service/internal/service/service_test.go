package service

import (
	"context"
	"github.com/jaam8/web_calculator/auth_service/internal/service/utils"
	errs "github.com/jaam8/web_calculator/common-lib/errors"
	"github.com/jaam8/web_calculator/common-lib/gen/auth_service"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type MockCacheAdapter struct {
	mock.Mock
}

func (m *MockCacheAdapter) SaveToken(token, userID string, refresh bool) error {
	args := m.Called(token, userID, refresh)
	return args.Error(0)
}

func (m *MockCacheAdapter) GetToken(token string, refresh bool) (string, error) {
	args := m.Called(token, refresh)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockCacheAdapter) DeleteToken(token string, refresh bool) error {
	args := m.Called(token, refresh)
	return args.Error(0)
}

type MockStorageAdapter struct {
	mock.Mock
}

func (m *MockStorageAdapter) RegisterUser(login string, hashPassword string) (string, error) {
	args := m.Called(login, "hashedpassword")
	return args.Get(0).(string), args.Error(1)
}

func (m *MockStorageAdapter) LoginUser(login string) (string, string, error) {
	args := m.Called(login)
	return args.Get(0).(string), args.Get(1).(string), args.Error(2)
}

func TestAuthService_RegisterUser(t *testing.T) {
	tests := []struct {
		name          string
		login         string
		password      string
		expectedError error
		mockSetup     func(storage *MockStorageAdapter)
	}{
		{
			name:          "successful registration",
			login:         "testuser",
			password:      "password",
			expectedError: nil,
			mockSetup: func(storage *MockStorageAdapter) {
				storage.On("RegisterUser", "testuser", "hashedpassword").Return("userID", nil)
			},
		},
		{
			name:          "user already exists",
			login:         "existinguser",
			password:      "password",
			expectedError: errs.ErrUserAlreadyExists,
			mockSetup: func(storage *MockStorageAdapter) {
				storage.On("RegisterUser", "existinguser", "hashedpassword").Return("", errs.ErrUserAlreadyExists)
			},
		},
		{
			name:          "empty login",
			login:         "",
			password:      "password",
			expectedError: errs.ErrEmptyLogin,
		},
		{
			name:          "empty password",
			login:         "testuser",
			password:      "",
			expectedError: errs.ErrEmptyPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := new(MockStorageAdapter)
			cache := new(MockCacheAdapter)
			if tt.mockSetup != nil {
				tt.mockSetup(storage)
			}

			service := NewAuthService(storage, cache, "secret",
				time.Second, time.Second)

			req := &auth_service.RegisterRequest{
				Login:    tt.login,
				Password: tt.password,
			}

			ctx := context.Background()
			ctx, _ = logger.New(ctx)
			resp, err := service.Register(ctx, req)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "userID", resp.UserId)
			}

			cache.AssertExpectations(t)
			storage.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name          string
		login         string
		password      string
		expectedError error
		mockSetup     func(storage *MockStorageAdapter, cache *MockCacheAdapter)
	}{
		{
			name:          "successful login",
			login:         "testuser",
			password:      "password",
			expectedError: nil,
			mockSetup: func(storage *MockStorageAdapter, cache *MockCacheAdapter) {
				realHash, _ := utils.GenerateHash("password")

				storage.
					On("LoginUser", "testuser").
					Return("userID", realHash, nil)

				cache.On("SaveToken", mock.Anything, "userID", false).Return(nil)
				cache.On("SaveToken", mock.Anything, "userID", true).Return(nil)
			},
		},
		{
			name:          "wrong password",
			login:         "testuser",
			password:      "password",
			expectedError: errs.ErrWrongPassword,
			mockSetup: func(storage *MockStorageAdapter, cache *MockCacheAdapter) {
				storage.On("LoginUser", "testuser").Return(
					"userID", "wrongpassword", nil)
			},
		},
		{
			name:          "user not found",
			login:         "existinguser",
			password:      "password",
			expectedError: errs.ErrUserNotFound,
			mockSetup: func(storage *MockStorageAdapter, cache *MockCacheAdapter) {
				storage.On("LoginUser", "existinguser").Return("", "", errs.ErrUserNotFound)
			},
		},
		{
			name:          "empty login",
			login:         "",
			password:      "password",
			expectedError: errs.ErrEmptyLogin,
		},
		{
			name:          "empty password",
			login:         "testuser",
			password:      "",
			expectedError: errs.ErrEmptyPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := new(MockStorageAdapter)
			cache := new(MockCacheAdapter)
			if tt.mockSetup != nil {
				tt.mockSetup(storage, cache)
			}

			service := NewAuthService(storage, cache, "secret",
				time.Second, time.Second)

			req := &auth_service.LoginRequest{
				Login:    tt.login,
				Password: tt.password,
			}

			ctx := context.Background()
			ctx, _ = logger.New(ctx)
			resp, err := service.Login(ctx, req)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "userID", resp.UserId)
			}

			cache.AssertExpectations(t)
			storage.AssertExpectations(t)
		})
	}
}

func TestAuthService_Refresh(t *testing.T) {
	tests := []struct {
		name         string
		refreshToken string
		mockSetup    func(cache *MockCacheAdapter)
		expectedErr  error
	}{
		{
			name:         "empty token",
			refreshToken: "",
			expectedErr:  errs.ErrInvalidToken,
		},
		{
			name: "token expired",
			refreshToken: func() string {
				token, err := utils.GenerateJWT("user123", "secret", true, time.Millisecond)
				time.Sleep(500 * time.Millisecond)
				require.NoError(t, err)
				return token
			}(),
			expectedErr: errs.ErrTokenExpired,
		},
		{
			name: "successful refresh",
			refreshToken: func() string {
				token, err := utils.GenerateJWT("user123", "secret", true, time.Minute)
				require.NoError(t, err)
				return token
			}(),
			mockSetup: func(cache *MockCacheAdapter) {
				cache.On("GetToken", mock.Anything, true).
					Return("user123", nil)
				cache.On("SaveToken", mock.Anything, "user123", false).
					Return(nil)
				cache.On("SaveToken", mock.Anything, "user123", true).
					Return(nil)
				cache.On("DeleteToken", mock.Anything, true).
					Return(nil)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := new(MockCacheAdapter)
			if tt.mockSetup != nil {
				tt.mockSetup(cache)
			}

			service := NewAuthService(nil, cache, "secret",
				time.Minute, time.Second)

			req := &auth_service.RefreshRequest{RefreshToken: tt.refreshToken}

			ctx := context.Background()
			ctx, _ = logger.New(ctx)
			resp, err := service.Refresh(ctx, req)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.Nil(t, resp)
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotEmpty(t, resp.AccessToken)
				require.NotEmpty(t, resp.RefreshToken)
			}

			cache.AssertExpectations(t)
		})
	}
}
