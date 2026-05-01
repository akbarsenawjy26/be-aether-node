package user

import (
	"context"

	domainUser "aether-node/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo domainUser.UserRepository
}

func NewUserService(repo domainUser.UserRepository) domainUser.UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, req *domainUser.CreateUserRequest) (*domainUser.User, error) {
	exists, err := s.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domainUser.ErrEmailAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domainUser.User{
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

func (s *userService) GetUser(ctx context.Context, guid string) (*domainUser.User, error) {
	return s.repo.GetByGUID(ctx, guid)
}

func (s *userService) ListUsers(ctx context.Context, params *domainUser.ListParams) (*domainUser.ListResult, error) {
	return s.repo.List(ctx, *params)
}

func (s *userService) UpdateUser(ctx context.Context, guid string, req *domainUser.UpdateUserRequest) (*domainUser.User, error) {
	user, err := s.repo.GetByGUID(ctx, guid)
	if err != nil {
		return nil, err
	}

	if req.Email != nil {
		exists, err := s.repo.ExistsByEmail(ctx, *req.Email)
		if err != nil {
			return nil, err
		}
		if exists && *req.Email != user.Email {
			return nil, domainUser.ErrEmailAlreadyExists
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
