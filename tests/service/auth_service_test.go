package service_test

import (
	"context"
	"testing"
	"time"

	domainAuth "aether-node/internal/domain/auth"
	userPkg "aether-node/internal/domain/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock for userPkg.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *userPkg.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByGUID(ctx context.Context, guid string) (*userPkg.User, error) {
	args := m.Called(ctx, guid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userPkg.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*userPkg.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userPkg.User), args.Error(1)
}

func (m *MockUserRepository) List(ctx context.Context, params userPkg.ListParams) (*userPkg.ListResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userPkg.ListResult), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *userPkg.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, guid string) error {
	args := m.Called(ctx, guid)
	return args.Error(0)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, guid string, loginAt time.Time) error {
	args := m.Called(ctx, guid, loginAt)
	return args.Error(0)
}

// MockRefreshTokenRepository is a mock for domainAuth.RefreshTokenRepository
type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *domainAuth.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domainAuth.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainAuth.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) DeleteByUserGUID(ctx context.Context, userGUID string) error {
	args := m.Called(ctx, userGUID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	args := m.Called(ctx, tokenHash)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) UpdateLastUsed(ctx context.Context, guid string, usedAt string) error {
	args := m.Called(ctx, guid, usedAt)
	return args.Error(0)
}

// TestLoginRequest tests the LoginRequest structure
func TestLoginRequest(t *testing.T) {
	req := &domainAuth.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	assert.Equal(t, "test@example.com", req.Email)
	assert.Equal(t, "password123", req.Password)
}

// TestRegisterRequest tests the RegisterRequest structure
func TestRegisterRequest(t *testing.T) {
	req := &domainAuth.RegisterRequest{
		Email:     "newuser@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	assert.Equal(t, "newuser@example.com", req.Email)
	assert.Equal(t, "password123", req.Password)
	assert.Equal(t, "John", req.FirstName)
	assert.Equal(t, "Doe", req.LastName)
}

// TestRefreshTokenRequest tests the RefreshTokenRequest structure
func TestRefreshTokenRequest(t *testing.T) {
	req := &domainAuth.RefreshTokenRequest{
		RefreshToken: "refresh-token-123",
	}

	assert.Equal(t, "refresh-token-123", req.RefreshToken)
}

// TestUserInfo tests the UserInfo structure
func TestUserInfo(t *testing.T) {
	userInfo := &domainAuth.UserInfo{
		GUID:      "guid-123",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	assert.Equal(t, "guid-123", userInfo.GUID)
	assert.Equal(t, "test@example.com", userInfo.Email)
	assert.Equal(t, "John", userInfo.FirstName)
	assert.Equal(t, "Doe", userInfo.LastName)
}

// TestLoginResult tests the LoginResult structure
func TestLoginResult(t *testing.T) {
	result := &domainAuth.LoginResult{
		AccessToken:  "access-token-123",
		RefreshToken: "refresh-token-456",
		ExpiresIn:    3600,
	}

	assert.Equal(t, "access-token-123", result.AccessToken)
	assert.Equal(t, "refresh-token-456", result.RefreshToken)
	assert.Equal(t, int64(3600), result.ExpiresIn)
}

// TestRefreshToken tests the RefreshToken structure
func TestRefreshToken(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	token := &domainAuth.RefreshToken{
		GUID:      "token-guid-123",
		UserGUID:  "user-guid-456",
		TokenHash: "hashed-token",
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}

	assert.Equal(t, "token-guid-123", token.GUID)
	assert.Equal(t, "user-guid-456", token.UserGUID)
	assert.Equal(t, "hashed-token", token.TokenHash)
	assert.Equal(t, expiresAt, token.ExpiresAt)
	assert.Equal(t, now, token.CreatedAt)
}

// TestAuthServiceInterfaces tests that the service implements the correct interface
func TestAuthServiceInterfaces(t *testing.T) {
	// This test verifies that the domain interfaces are properly defined
	var _ domainAuth.AuthService = (*mockAuthService)(nil)
}

// mockAuthService is a minimal implementation for interface checking
type mockAuthService struct{}

func (m *mockAuthService) Login(ctx context.Context, req *domainAuth.LoginRequest) (*domainAuth.LoginResult, error) {
	return nil, nil
}
func (m *mockAuthService) Logout(ctx context.Context, userGUID string) error {
	return nil
}
func (m *mockAuthService) Register(ctx context.Context, req *domainAuth.RegisterRequest) (*domainAuth.UserInfo, error) {
	return nil, nil
}
func (m *mockAuthService) ForgotPassword(ctx context.Context, req *domainAuth.ForgotPasswordRequest) error {
	return nil
}
func (m *mockAuthService) RefreshToken(ctx context.Context, req *domainAuth.RefreshTokenRequest) (*domainAuth.LoginResult, error) {
	return nil, nil
}
func (m *mockAuthService) ValidateAccessToken(ctx context.Context, tokenString string) (string, error) {
	return "", nil
}
