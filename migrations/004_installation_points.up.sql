-- Migration: 004_installation_points.up.sql
-- Create installation_points table

CREATE TABLE installation_points (
    guid          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(255) NOT NULL,
    device_guid   UUID NOT NULL REFERENCES devices(guid) ON DELETE RESTRICT,
    location_guid UUID NOT NULL REFERENCES locations(guid) ON DELETE RESTRICT,
    notes         TEXT,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_installation_points_device ON installation_points(device_guid) WHERE deleted_at IS NULL;
CREATE INDEX idx_installation_points_location ON installation_points(location_guid) WHERE deleted_at IS NULL;
CREATE INDEX idx_installation_points_name ON installation_points(name) WHERE deleted_at IS NULL;
CREATE INDEX idx_installation_points_created_at ON installation_points(created_at DESC) WHERE deleted_at IS NULL;
