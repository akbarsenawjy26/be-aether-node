-- Migration: 002_locations.up.sql
-- Create locations table

CREATE TABLE locations (
    guid        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(255) NOT NULL,
    notes       TEXT,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_locations_name ON locations(name) WHERE deleted_at IS NULL;
CREATE INDEX idx_locations_created_at ON locations(created_at DESC) WHERE deleted_at IS NULL;
