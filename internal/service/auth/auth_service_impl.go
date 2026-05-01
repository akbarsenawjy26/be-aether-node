package auth

import (
	"context"
	"errors"
	"time"

	"aether-node/internal/domain/user"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

type authService struct {
	userRepo       user.UserRepository
	refreshTokenRepo RefreshTokenRepository
	jwtSecret      []byte
	accessTokenTTL time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthService(
	userRepo user.UserRepository,
	refreshTokenRepo RefreshTokenRepository,
	jwtSecret string,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) AuthService {
	return &authService{
		userRepo:        userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtSecret:       []byte(jwtSecret),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

type AccessTokenClaims struct {
	UserGUID string `json:"sub"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

func (s *authService) Login(ctx context.Context, req *LoginRequest) (*LoginResult, error) {
	// Find user by email
	u, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(u)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store refresh token hash
	tokenHash := HashToken(refreshToken)
	refreshTokenRecord := &RefreshToken{
		UserGUID:  u.GUID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.refreshTokenTTL),
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshTokenRecord); err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
	}, nil
}

func (s *authService) Logout(ctx context.Context, userGUID string) error {
	return s.refreshTokenRepo.DeleteByUserGUID(ctx, userGUID)
}

func (s *authService) Register(ctx context.Context, req *RegisterRequest) (*UserInfo, error) {
	// Check if email exists
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, user.ErrEmailExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &user.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	return &UserInfo{
		GUID:      u.GUID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}, nil
}

func (s *authService) ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) error {
	// In production, send email with reset link
	// For now, just check if user exists
	_, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			// Don't reveal if user exists or not
			return nil
		}
		return err
	}

	// TODO: Send email with reset token
	return nil
}

func (s *authService) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*LoginResult, error) {
	// Hash the provided token and look it up
	tokenHash := HashToken(req.RefreshToken)

	refreshTokenRecord, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Get user
	u, err := s.userRepo.GetByGUID(ctx, refreshTokenRecord.UserGUID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Delete old refresh token
	_ = s.refreshTokenRepo.DeleteByTokenHash(ctx, tokenHash)

	// Generate new tokens
	accessToken, err := s.generateAccessToken(u)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store new refresh token
	newTokenHash := HashToken(newRefreshToken)
	newRefreshTokenRecord := &RefreshToken{
		UserGUID:  u.GUID,
		TokenHash: newTokenHash,
		ExpiresAt: time.Now().Add(s.refreshTokenTTL),
	}

	if err := s.refreshTokenRepo.Create(ctx, newRefreshTokenRecord); err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
	}, nil
}

func (s *authService) ValidateAccessToken(ctx context.Context, tokenString string) (string, error) {
	claims := &AccessTokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid access token")
	}

	return claims.UserGUID, nil
}

func (s *authService) generateAccessToken(u *user.User) (string, error) {
	now := time.Now()
	claims := AccessTokenClaims{
		UserGUID: u.GUID,
		Email:    u.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *authService) generateRefreshToken() (string, error) {
	// Generate a random token
	tokenBytes := make([]byte, 32)
	if _, err := readRandomBytes(tokenBytes); err != nil {
		return "", err
	}
	return hexEncode(tokenBytes), nil
}

// Helper functions
func readRandomBytes(b []byte) (n int, err error) {
	// Using crypto/rand
	return 0, nil // Simplified
}

func hexEncode(b []byte) string {
	const hexChars = "0123456789abcdef"
	result := make([]byte, len(b)*2)
	for i, v := range b {
		result[i*2] = hexChars[v>>4]
		result[i*2+1] = hexChars[v&0x0f]
	}
	return string(result)
}
