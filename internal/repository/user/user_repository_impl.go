package user

import (
	"context"
	"errors"
	"time"

	"aether-node/internal/db"
	domainUser "aether-node/internal/domain/user"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type userRepository struct {
	db *db.Queries
}

func NewUserRepository(pool *pgxpool.Pool) domainUser.UserRepository {
	return &userRepository{db: db.New(pool)}
}

func (r *userRepository) Create(ctx context.Context, user *domainUser.User) error {
	if user.GUID == "" {
		user.GUID = uuid.New().String()
	}

	guid, _ := uuid.Parse(user.GUID)
	now := time.Now()

	params := db.CreateUserParams{
		Guid:         db.NewUUID(guid),
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		IsActive:     user.IsActive,
		CreatedAt:    db.NewTimestamptz(now),
		UpdatedAt:    db.NewTimestamptz(now),
	}

	if user.RoleGUID != nil {
		id, _ := uuid.Parse(*user.RoleGUID)
		params.RoleGuid = db.NewUUID(id)
	}

	return r.db.CreateUser(ctx, params)
}

func (r *userRepository) GetByGUID(ctx context.Context, guid string) (*domainUser.User, error) {
	id, err := uuid.Parse(guid)
	if err != nil {
		return nil, ErrUserNotFound
	}

	dbUser, err := r.db.GetUserByGUID(ctx, db.NewUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return db.UserFromDB(&dbUser), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	dbUser, err := r.db.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return db.UserFromDB(&dbUser), nil
}

func (r *userRepository) List(ctx context.Context, params domainUser.ListParams) (*domainUser.ListResult, error) {
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Page <= 0 {
		params.Page = 1
	}

	search := "%" + params.Search + "%"
	offset := (params.Page - 1) * params.Limit

	dbUsers, err := r.db.ListUsers(ctx, db.ListUsersParams{
		Email:  search,
		Limit:  int32(params.Limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	total, err := r.db.CountUsers(ctx)
	if err != nil {
		return nil, err
	}

	users := make([]*domainUser.User, 0, len(dbUsers))
	for i := range dbUsers {
		users = append(users, db.UserFromDB(&dbUsers[i]))
	}

	totalPages := int(total) / params.Limit
	if int(total)%params.Limit > 0 {
		totalPages++
	}

	return &domainUser.ListResult{
		Users:     users,
		Total:     total,
		Page:      params.Page,
		Limit:     params.Limit,
		TotalPage: totalPages,
	}, nil
}

func (r *userRepository) Update(ctx context.Context, user *domainUser.User) error {
	guid, _ := uuid.Parse(user.GUID)
	now := time.Now()

	params := db.UpdateUserParams{
		Guid:         db.NewUUID(guid),
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		IsActive:     user.IsActive,
		UpdatedAt:    db.NewTimestamptz(now),
	}

	if user.RoleGUID != nil {
		id, _ := uuid.Parse(*user.RoleGUID)
		params.RoleGuid = db.NewUUID(id)
	}

	err := r.db.UpdateUser(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, guid string) error {
	id, err := uuid.Parse(guid)
	if err != nil {
		return ErrUserNotFound
	}

	err = r.db.DeleteUser(ctx, db.DeleteUserParams{
		Guid:      db.NewUUID(id),
		DeletedAt: db.NewTimestamptz(time.Now()),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	exists, err := r.db.ExistsUserByEmail(ctx, email)
	return bool(exists), err
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, guid string, loginAt time.Time) error {
	id, _ := uuid.Parse(guid)
	return r.db.UpdateUserLastLogin(ctx, db.UpdateUserLastLoginParams{
		Guid:      db.NewUUID(id),
		UpdatedAt: db.NewTimestamptz(loginAt),
	})
}
