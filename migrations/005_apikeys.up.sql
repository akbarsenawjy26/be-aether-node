-- Migration: 005_apikeys.up.sql
-- Create apikeys table

CREATE TABLE apikeys (
    guid        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_hash    VARCHAR(255) NOT NULL,
    notes       TEXT,
    expire_date TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_apikeys_key_hash ON apikeys(key_hash) WHERE deleted_at IS NULL;
CREATE INDEX idx_apikeys_expire_date ON apikeys(expire_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_apikeys_is_active ON apikeys(is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_apikeys_created_at ON apikeys(created_at DESC) WHERE deleted_at IS NULL;
