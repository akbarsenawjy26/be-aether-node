-- Migration: 001_users.up.sql
-- Create users table

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    guid            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    role_guid       UUID,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role ON users(role_guid) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON users(created_at DESC) WHERE deleted_at IS NULL;
