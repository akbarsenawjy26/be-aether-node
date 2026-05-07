-- Migration: 008_thresholds.up.sql
-- Create thresholds table for alarm settings

CREATE TYPE alarm_severity AS ENUM ('info', 'warning', 'critical');

CREATE TABLE thresholds (
    guid            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_guid     UUID NOT NULL REFERENCES devices(guid) ON DELETE CASCADE,
    parameter_name  VARCHAR(100) NOT NULL, -- e.g., 'pm2_5', 'temperature'
    min_value       DOUBLE PRECISION,
    max_value       DOUBLE PRECISION,
    severity        alarm_severity NOT NULL DEFAULT 'warning',
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_thresholds_device_guid ON thresholds(device_guid) WHERE deleted_at IS NULL;
CREATE INDEX idx_thresholds_parameter_name ON thresholds(parameter_name) WHERE deleted_at IS NULL;
