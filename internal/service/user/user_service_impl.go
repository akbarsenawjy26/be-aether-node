package user

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword = errors.New("invalid password")
	ErrEmailExists     = errors.New("email already exists")
)

type userService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	// Check if email already exists
	exists, err := s.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		RoleGUID:     req.RoleGUID,
		IsActive:     true,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetUser(ctx context.Context, guid string) (*User, error) {
	return s.repo.GetByGUID(ctx, guid)
}

func (s *userService) ListUsers(ctx context.Context, params *ListParams) (*ListResult, error) {
	return s.repo.List(ctx, *params)
}

func (s *userService) UpdateUser(ctx context.Context, guid string, req *UpdateUserRequest) (*User, error) {
	user, err := s.repo.GetByGUID(ctx, guid)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Email != nil {
		// Check if new email already exists
		exists, err := s.repo.ExistsByEmail(ctx, *req.Email)
		if err != nil {
			return nil, err
		}
		if exists && *req.Email != user.Email {
			return nil, ErrEmailExists
		}
		user.Email = *req.Email
	}

	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = string(hashedPassword)
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}

	if req.LastName != nil {
		user.LastName = *req.LastName
	}

	if req.RoleGUID != nil {
		user.RoleGUID = req.RoleGUID
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) DeleteUser(ctx context.Context, guid string) error {
	return s.repo.Delete(ctx, guid)
}
