package location

import (
	"context"
	domainLocation "aether-node/internal/domain/location"

	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrLocationNotFound = errors.New("location not found")

type locationRepository struct {
	db *pgxpool.Pool
}

func NewLocationRepository(db *pgxpool.Pool) domainLocation.LocationRepository {
	return &locationRepository{db: db}
}

func (r *locationRepository) Create(ctx context.Context, location *domainLocation.Location) error {
	if location.GUID == "" {
		location.GUID = uuid.New().String()
	}

	query := `
		INSERT INTO locations (guid, name, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query,
		location.GUID,
		location.Name,
		location.Notes,
		now,
		now,
	)

	return err
}

func (r *locationRepository) GetByGUID(ctx context.Context, guid string) (*domainLocation.Location, error) {
	query := `
		SELECT guid, name, notes, created_at, updated_at, deleted_at
		FROM locations
		WHERE guid = $1 AND deleted_at IS NULL
	`

	location := &domainLocation.Location{}
	err := r.db.QueryRow(ctx, query, guid).Scan(
		&location.GUID,
		&location.Name,
		&location.Notes,
		&location.CreatedAt,
		&location.UpdatedAt,
		&location.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrLocationNotFound
	}
	if err != nil {
		return nil, err
	}

	return location, nil
}

func (r *locationRepository) List(ctx context.Context, params domainLocation.ListParams) (*domainLocation.ListResult, error) {
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
	countQuery := `SELECT COUNT(*) FROM locations WHERE deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 1

	if params.Search != "" {
		countQuery += ` AND (name ILIKE $1 OR notes ILIKE $1)`
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
		SELECT guid, name, notes, created_at, updated_at, deleted_at
		FROM locations
		WHERE deleted_at IS NULL
	`

	if params.Search != "" {
		query += ` AND (name ILIKE $1 OR notes ILIKE $1)`
	}

	query += ` ORDER BY ` + params.Order + ` ` + params.Sort
	query += ` LIMIT $` + string(rune('0'+argIdx)) + ` OFFSET $` + string(rune('0'+argIdx+1))

	args = append(args, params.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	locations := make([]*domainLocation.Location, 0)
	for rows.Next() {
		loc := &domainLocation.Location{}
		err := rows.Scan(
			&loc.GUID,
			&loc.Name,
			&loc.Notes,
			&loc.CreatedAt,
			&loc.UpdatedAt,
			&loc.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}

	totalPages := int(total) / params.Limit
	if int(total)%params.Limit > 0 {
		totalPages++
	}

	return &domainLocation.ListResult{
		Locations: locations,
		Total:     total,
		Page:      params.Page,
		Limit:     params.Limit,
		TotalPage: totalPages,
	}, nil
}

func (r *locationRepository) Update(ctx context.Context, location *domainLocation.Location) error {
	query := `
		UPDATE locations
		SET name = $2, notes = $3, updated_at = $4
		WHERE guid = $1 AND deleted_at IS NULL
	`

	location.UpdatedAt = time.Now()
	result, err := r.db.Exec(ctx, query,
		location.GUID,
		location.Name,
		location.Notes,
		location.UpdatedAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrLocationNotFound
	}

	return nil
}

func (r *locationRepository) Delete(ctx context.Context, guid string) error {
	query := `
		UPDATE locations
		SET deleted_at = $2, updated_at = $2
		WHERE guid = $1 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query, guid, now)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrLocationNotFound
	}

	return nil
}

func (r *locationRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM locations WHERE name = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, name).Scan(&exists)
	return exists, err
}
