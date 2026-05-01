-- Migration: 003_devices.up.sql
-- Create devices table

CREATE TABLE devices (
    guid          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type          VARCHAR(100) NOT NULL,
    serial_number VARCHAR(100) NOT NULL UNIQUE,
    alias         VARCHAR(255),
    notes         TEXT,
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_devices_serial_number ON devices(serial_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_devices_type ON devices(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_devices_created_at ON devices(created_at DESC) WHERE deleted_at IS NULL;
