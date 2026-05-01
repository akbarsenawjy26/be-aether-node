package user

import (
	"context"
	domainUser "aether-node/internal/domain/user"

		"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) domainUser.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domainUser.User) error {
	if user.GUID == "" {
		user.GUID = uuid.New().String()
	}

	query := `
		INSERT INTO users (guid, email, password_hash, first_name, last_name, role_guid, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query,
		user.GUID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.RoleGUID,
		true,
		now,
		now,
	)

	return err
}

func (r *userRepository) GetByGUID(ctx context.Context, guid string) (*domainUser.User, error) {
	query := `
		SELECT guid, email, password_hash, first_name, last_name, role_guid, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE guid = $1 AND deleted_at IS NULL
	`

	user := &domainUser.User{}
	err := r.db.QueryRow(ctx, query, guid).Scan(
		&user.GUID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.RoleGUID,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	query := `
		SELECT guid, email, password_hash, first_name, last_name, role_guid, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	user := &domainUser.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.GUID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.RoleGUID,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) List(ctx context.Context, params domainUser.ListParams) (*domainUser.ListResult, error) {
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Order == "" {
		params.Order = "created_at"
	}
	if params.Sort == "" {
		params.Sort = "DESC"
	}

	offset := (params.Page - 1) * params.Limit

	// Count total
	countQuery := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 1

	if params.Search != "" {
		countQuery += ` AND (email ILIKE $1 OR first_name ILIKE $1 OR last_name ILIKE $1)`
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}

	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Get data
	query := `
		SELECT guid, email, password_hash, first_name, last_name, role_guid, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE deleted_at IS NULL
	`

	if params.Search != "" {
		query += ` AND (email ILIKE $1 OR first_name ILIKE $1 OR last_name ILIKE $1)`
	}

	query += ` ORDER BY ` + params.Order + ` ` + params.Sort
	query += ` LIMIT $` + string(rune('0'+argIdx)) + ` OFFSET $` + string(rune('0'+argIdx+1))

	args = append(args, params.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*domainUser.User, 0)
	for rows.Next() {
		user := &domainUser.User{}
		err := rows.Scan(
			&user.GUID,
			&user.Email,
			&user.PasswordHash,
			&user.FirstName,
			&user.LastName,
			&user.RoleGUID,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
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
	query := `
		UPDATE users
		SET email = $2, password_hash = $3, first_name = $4, last_name = $5, role_guid = $6, is_active = $7, updated_at = $8
		WHERE guid = $1 AND deleted_at IS NULL
	`

	user.UpdatedAt = time.Now()
	result, err := r.db.Exec(ctx, query,
		user.GUID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.RoleGUID,
		user.IsActive,
		user.UpdatedAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, guid string) error {
	query := `
		UPDATE users
		SET deleted_at = $2, updated_at = $2
		WHERE guid = $1 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query, guid, now)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	return exists, err
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, guid string, loginAt time.Time) error {
	query := `UPDATE users SET updated_at = $2 WHERE guid = $1`
	_, err := r.db.Exec(ctx, query, guid, loginAt)
	return err
}
