package apikey

import (
	"context"
	domainAPIKey "aether-node/internal/domain/apikey"

	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAPIKeyNotFound = errors.New("API key not found")
var ErrAPIKeyInvalid = errors.New("invalid API key")

type apiKeyRepository struct {
	db *pgxpool.Pool
}

func NewAPIKeyRepository(db *pgxpool.Pool) domainAPIKey.APIKeyRepository {
	return &apiKeyRepository{db: db}
}

func (r *apiKeyRepository) Create(ctx context.Context, apiKey *domainAPIKey.APIKey) error {
	if apiKey.GUID == "" {
		apiKey.GUID = uuid.New().String()
	}

	query := `
		INSERT INTO api_keys (guid, key_hash, notes, expire_date, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query,
		apiKey.GUID,
		apiKey.KeyHash,
		apiKey.Notes,
		apiKey.ExpireDate,
		apiKey.IsActive,
		now,
		now,
	)

	return err
}

func (r *apiKeyRepository) GetByGUID(ctx context.Context, guid string) (*domainAPIKey.APIKey, error) {
	query := `
		SELECT guid, key_hash, notes, expire_date, is_active, created_at, updated_at, deleted_at
		FROM api_keys
		WHERE guid = $1 AND deleted_at IS NULL
	`

	apiKey := &domainAPIKey.APIKey{}
	err := r.db.QueryRow(ctx, query, guid).Scan(
		&apiKey.GUID,
		&apiKey.KeyHash,
		&apiKey.Notes,
		&apiKey.ExpireDate,
		&apiKey.IsActive,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
		&apiKey.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAPIKeyNotFound
	}
	if err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (r *apiKeyRepository) GetByKeyHash(ctx context.Context, keyHash string) (*domainAPIKey.APIKey, error) {
	query := `
		SELECT guid, key_hash, notes, expire_date, is_active, created_at, updated_at, deleted_at
		FROM api_keys
		WHERE key_hash = $1 AND deleted_at IS NULL
	`

	apiKey := &domainAPIKey.APIKey{}
	err := r.db.QueryRow(ctx, query, keyHash).Scan(
		&apiKey.GUID,
		&apiKey.KeyHash,
		&apiKey.Notes,
		&apiKey.ExpireDate,
		&apiKey.IsActive,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
		&apiKey.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAPIKeyNotFound
	}
	if err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (r *apiKeyRepository) List(ctx context.Context, params domainAPIKey.ListParams) (*domainAPIKey.ListResult, error) {
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
	countQuery := `SELECT COUNT(*) FROM api_keys WHERE deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 1

	if params.Search != "" {
		countQuery += ` AND notes ILIKE $1`
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
		SELECT guid, key_hash, notes, expire_date, is_active, created_at, updated_at, deleted_at
		FROM api_keys
		WHERE deleted_at IS NULL
	`

	if params.Search != "" {
		query += ` AND notes ILIKE $1`
	}

	query += ` ORDER BY ` + params.Order + ` ` + params.Sort
	query += ` LIMIT $` + string(rune('0'+argIdx)) + ` OFFSET $` + string(rune('0'+argIdx+1))

	args = append(args, params.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apiKeys := make([]*domainAPIKey.APIKey, 0)
	for rows.Next() {
		ak := &domainAPIKey.APIKey{}
		err := rows.Scan(
			&ak.GUID,
			&ak.KeyHash,
			&ak.Notes,
			&ak.ExpireDate,
			&ak.IsActive,
			&ak.CreatedAt,
			&ak.UpdatedAt,
			&ak.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		apiKeys = append(apiKeys, ak)
	}

	totalPages := int(total) / params.Limit
	if int(total)%params.Limit > 0 {
		totalPages++
	}

	return &domainAPIKey.ListResult{
		APIKeys:   apiKeys,
		Total:     total,
		Page:      params.Page,
		Limit:     params.Limit,
		TotalPage: totalPages,
	}, nil
}

func (r *apiKeyRepository) Update(ctx context.Context, apiKey *domainAPIKey.APIKey) error {
	query := `
		UPDATE api_keys
		SET notes = $2, expire_date = $3, is_active = $4, updated_at = $5
		WHERE guid = $1 AND deleted_at IS NULL
	`

	apiKey.UpdatedAt = time.Now()
	result, err := r.db.Exec(ctx, query,
		apiKey.GUID,
		apiKey.Notes,
		apiKey.ExpireDate,
		apiKey.IsActive,
		apiKey.UpdatedAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrAPIKeyNotFound
	}

	return nil
}

func (r *apiKeyRepository) Delete(ctx context.Context, guid string) error {
	query := `
		UPDATE api_keys
		SET deleted_at = $2, updated_at = $2
		WHERE guid = $1 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query, guid, now)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrAPIKeyNotFound
	}

	return nil
}

func (r *apiKeyRepository) ValidateKey(ctx context.Context, key string) (*domainAPIKey.APIKey, error) {
	// Hash the provided key
	keyHash := sha256.Sum256([]byte(key))
	keyHashStr := hex.EncodeToString(keyHash[:])

	apiKey, err := r.GetByKeyHash(ctx, keyHashStr)
	if err != nil {
		if errors.Is(err, ErrAPIKeyNotFound) {
			return nil, ErrAPIKeyInvalid
		}
		return nil, err
	}

	// Check if active
	if !apiKey.IsActive {
		return nil, ErrAPIKeyInvalid
	}

	// Check if expired
	if time.Now().After(apiKey.ExpireDate) {
		return nil, ErrAPIKeyInvalid
	}

	return apiKey, nil
}
