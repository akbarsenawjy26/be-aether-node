package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	domainAuth "aether-node/internal/domain/auth"
	userPkg "aether-node/internal/domain/user"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo        userPkg.UserRepository
	refreshTokenRepo domainAuth.RefreshTokenRepository
	jwtSecret       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthService(
	userRepo userPkg.UserRepository,
	refreshTokenRepo domainAuth.RefreshTokenRepository,
	jwtSecret string,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) domainAuth.AuthService {
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

func (s *authService) Login(ctx context.Context, req *domainAuth.LoginRequest) (*domainAuth.LoginResult, error) {
	u, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, userPkg.ErrUserNotFound) {
			return nil, domainAuth.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, domainAuth.ErrInvalidCredentials
	}

	accessToken, err := s.generateAccessToken(u)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, err
	}

	tokenHash := hashToken(refreshToken)
	refreshTokenRecord := &domainAuth.RefreshToken{
		UserGUID:  u.GUID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.refreshTokenTTL),
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshTokenRecord); err != nil {
		return nil, err
	}

	return &domainAuth.LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
	}, nil
}

func (s *authService) Logout(ctx context.Context, userGUID string) error {
	return s.refreshTokenRepo.DeleteByUserGUID(ctx, userGUID)
}

func (s *authService) Register(ctx context.Context, req *domainAuth.RegisterRequest) (*domainAuth.UserInfo, error) {
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, userPkg.ErrEmailAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &userPkg.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	return &domainAuth.UserInfo{
		GUID:      u.GUID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}, nil
}

func (s *authService) ForgotPassword(ctx context.Context, req *domainAuth.ForgotPasswordRequest) error {
	_, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, userPkg.ErrUserNotFound) {
			return nil
		}
		return err
	}
	return nil
}

func (s *authService) RefreshToken(ctx context.Context, req *domainAuth.RefreshTokenRequest) (*domainAuth.LoginResult, error) {
	tokenHash := hashToken(req.RefreshToken)

	refreshTokenRecord, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	u, err := s.userRepo.GetByGUID(ctx, refreshTokenRecord.UserGUID)
	if err != nil {
		return nil, domainAuth.ErrUserNotFound
	}

	_ = s.refreshTokenRepo.DeleteByTokenHash(ctx, tokenHash)

	accessToken, err := s.generateAccessToken(u)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, err
	}

	newTokenHash := hashToken(newRefreshToken)
	newRefreshTokenRecord := &domainAuth.RefreshToken{
		UserGUID:  u.GUID,
		TokenHash: newTokenHash,
		ExpiresAt: time.Now().Add(s.refreshTokenTTL),
	}

	if err := s.refreshTokenRepo.Create(ctx, newRefreshTokenRecord); err != nil {
		return nil, err
	}

	return &domainAuth.LoginResult{
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

func (s *authService) generateAccessToken(u *userPkg.User) (string, error) {
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
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(tokenBytes), nil
}

func hashToken(token string) string {
	return token
}
