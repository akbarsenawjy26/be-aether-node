package db

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"aether-node/internal/domain/apikey"
	"aether-node/internal/domain/auth"
	"aether-node/internal/domain/device"
	"aether-node/internal/domain/installation_point"
	"aether-node/internal/domain/location"
	"aether-node/internal/domain/user"
)

// ─── User ────────────────────────────────────────────────────────────────────

func UserFromDB(in *User) *user.User {
	if in == nil {
		return nil
	}
	out := &user.User{
		Email:        in.Email,
		PasswordHash: in.PasswordHash,
		FirstName:    in.FirstName,
		LastName:     in.LastName,
		IsActive:     in.IsActive,
	}
	if in.Guid.Valid {
		out.GUID = uuid.UUID(in.Guid.Bytes).String()
	}
	if in.RoleGuid.Valid {
		s := uuid.UUID(in.RoleGuid.Bytes).String()
		out.RoleGUID = &s
	}
	if in.CreatedAt.Valid {
		out.CreatedAt = in.CreatedAt.Time
	}
	if in.UpdatedAt.Valid {
		out.UpdatedAt = in.UpdatedAt.Time
	}
	if in.DeletedAt.Valid {
		out.DeletedAt = &in.DeletedAt.Time
	}
	return out
}

// ─── Device ──────────────────────────────────────────────────────────────────

func DeviceFromDB(in *Device) *device.Device {
	if in == nil {
		return nil
	}
	out := &device.Device{
		Type:         in.Type,
		SerialNumber: in.SerialNumber,
		IsActive:     in.IsActive,
	}
	if in.Guid.Valid {
		out.GUID = uuid.UUID(in.Guid.Bytes).String()
	}
	if in.Alias.Valid {
		out.Alias = in.Alias.String
	}
	if in.Notes.Valid {
		out.Notes = in.Notes.String
	}
	if in.CreatedAt.Valid {
		out.CreatedAt = in.CreatedAt.Time
	}
	if in.UpdatedAt.Valid {
		out.UpdatedAt = in.UpdatedAt.Time
	}
	if in.DeletedAt.Valid {
		out.DeletedAt = &in.DeletedAt.Time
	}
	return out
}

// ─── Location ────────────────────────────────────────────────────────────────

func LocationFromDB(in *Location) *location.Location {
	if in == nil {
		return nil
	}
	out := &location.Location{
		Name: in.Name,
	}
	if in.Guid.Valid {
		out.GUID = uuid.UUID(in.Guid.Bytes).String()
	}
	if in.Notes.Valid {
		out.Notes = in.Notes.String
	}
	if in.CreatedAt.Valid {
		out.CreatedAt = in.CreatedAt.Time
	}
	if in.UpdatedAt.Valid {
		out.UpdatedAt = in.UpdatedAt.Time
	}
	if in.DeletedAt.Valid {
		out.DeletedAt = &in.DeletedAt.Time
	}
	return out
}

// ─── InstallationPoint ───────────────────────────────────────────────────────

func InstallationPointFromDB(in *InstallationPoint) *installation_point.InstallationPoint {
	if in == nil {
		return nil
	}
	out := &installation_point.InstallationPoint{
		Name: in.Name,
	}
	if in.Guid.Valid {
		out.GUID = uuid.UUID(in.Guid.Bytes).String()
	}
	if in.DeviceGuid.Valid {
		out.DeviceGUID = uuid.UUID(in.DeviceGuid.Bytes).String()
	}
	if in.LocationGuid.Valid {
		out.LocationGUID = uuid.UUID(in.LocationGuid.Bytes).String()
	}
	if in.Notes.Valid {
		out.Notes = in.Notes.String
	}
	if in.CreatedAt.Valid {
		out.CreatedAt = in.CreatedAt.Time
	}
	if in.UpdatedAt.Valid {
		out.UpdatedAt = in.UpdatedAt.Time
	}
	if in.DeletedAt.Valid {
		out.DeletedAt = &in.DeletedAt.Time
	}
	return out
}

// ─── APIKey ──────────────────────────────────────────────────────────────────

func APIKeyFromDB(in *Apikey) *apikey.APIKey {
	if in == nil {
		return nil
	}
	out := &apikey.APIKey{
		KeyHash:  in.KeyHash,
		IsActive: in.IsActive,
	}
	if in.Guid.Valid {
		out.GUID = uuid.UUID(in.Guid.Bytes).String()
	}
	if in.Notes.Valid {
		out.Notes = in.Notes.String
	}
	if in.ExpireDate.Valid {
		out.ExpireDate = in.ExpireDate.Time
	}
	if in.CreatedAt.Valid {
		out.CreatedAt = in.CreatedAt.Time
	}
	if in.UpdatedAt.Valid {
		out.UpdatedAt = in.UpdatedAt.Time
	}
	if in.DeletedAt.Valid {
		out.DeletedAt = &in.DeletedAt.Time
	}
	return out
}

// ─── RefreshToken ───────────────────────────────────────────────────────────

func RefreshTokenFromDB(in *RefreshToken) *auth.RefreshToken {
	if in == nil {
		return nil
	}
	out := &auth.RefreshToken{
		TokenHash: in.TokenHash,
	}
	if in.Guid.Valid {
		out.GUID = uuid.UUID(in.Guid.Bytes).String()
	}
	if in.UserGuid.Valid {
		out.UserGUID = uuid.UUID(in.UserGuid.Bytes).String()
	}
	if in.ExpiresAt.Valid {
		out.ExpiresAt = in.ExpiresAt.Time
	}
	if in.CreatedAt.Valid {
		out.CreatedAt = in.CreatedAt.Time
	}
	return out
}

// ─── Constructor helpers ─────────────────────────────────────────────────────

func NewUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func NewNullableUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func NewTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func NewText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}
