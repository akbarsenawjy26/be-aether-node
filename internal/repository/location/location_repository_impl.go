package location

import (
	"context"
	"errors"
	"time"

	"aether-node/internal/db"
	domainLocation "aether-node/internal/domain/location"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrLocationNotFound = errors.New("location not found")

type locationRepository struct {
	db *db.Queries
}

func NewLocationRepository(queries *db.Queries) domainLocation.LocationRepository {
	return &locationRepository{db: queries}
}

func (r *locationRepository) Create(ctx context.Context, loc *domainLocation.Location) error {
	if loc.GUID == "" {
		loc.GUID = uuid.New().String()
	}

	guid, _ := uuid.Parse(loc.GUID)
	now := time.Now()

	params := db.CreateLocationParams{
		Guid:      db.NewUUID(guid),
		Name:      loc.Name,
		CreatedAt: db.NewTimestamptz(now),
		UpdatedAt: db.NewTimestamptz(now),
	}

	if loc.Notes != "" {
		params.Notes = db.NewText(loc.Notes)
	}

	return r.db.CreateLocation(ctx, params)
}

func (r *locationRepository) GetByGUID(ctx context.Context, guid string) (*domainLocation.Location, error) {
	id, err := uuid.Parse(guid)
	if err != nil {
		return nil, ErrLocationNotFound
	}

	dbLoc, err := r.db.GetLocationByGUID(ctx, db.NewUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLocationNotFound
		}
		return nil, err
	}

	return db.LocationFromDB(&dbLoc), nil
}

func (r *locationRepository) List(ctx context.Context, params domainLocation.ListParams) (*domainLocation.ListResult, error) {
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Page <= 0 {
		params.Page = 1
	}

	search := "%" + params.Search + "%"
	offset := (params.Page - 1) * params.Limit

	dbLocs, err := r.db.ListLocations(ctx, db.ListLocationsParams{
		Name:   search,
		Limit:  int32(params.Limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	total, err := r.db.CountLocations(ctx)
	if err != nil {
		return nil, err
	}

	locations := make([]*domainLocation.Location, 0, len(dbLocs))
	for i := range dbLocs {
		locations = append(locations, db.LocationFromDB(&dbLocs[i]))
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

func (r *locationRepository) Update(ctx context.Context, loc *domainLocation.Location) error {
	guid, _ := uuid.Parse(loc.GUID)
	now := time.Now()

	params := db.UpdateLocationParams{
		Guid:      db.NewUUID(guid),
		Name:      loc.Name,
		UpdatedAt: db.NewTimestamptz(now),
	}

	if loc.Notes != "" {
		params.Notes = db.NewText(loc.Notes)
	}

	err := r.db.UpdateLocation(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrLocationNotFound
		}
		return err
	}
	return nil
}

func (r *locationRepository) Delete(ctx context.Context, guid string) error {
	id, err := uuid.Parse(guid)
	if err != nil {
		return ErrLocationNotFound
	}

	err = r.db.DeleteLocation(ctx, db.DeleteLocationParams{
		Guid:      db.NewUUID(id),
		DeletedAt: db.NewTimestamptz(time.Now()),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrLocationNotFound
		}
		return err
	}
	return nil
}

func (r *locationRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	exists, err := r.db.ExistsLocationByName(ctx, name)
	return bool(exists), err
}
