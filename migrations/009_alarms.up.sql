-- Migration: 009_alarms.up.sql
-- Create alarms table for tracking triggered events

CREATE TYPE alarm_status AS ENUM ('active', 'acknowledged', 'resolved');

CREATE TABLE alarms (
    guid            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_guid     UUID NOT NULL REFERENCES devices(guid) ON DELETE CASCADE,
    threshold_guid  UUID REFERENCES thresholds(guid) ON DELETE SET NULL,
    parameter_name  VARCHAR(100) NOT NULL,
    triggered_value DOUBLE PRECISION NOT NULL,
    status          alarm_status NOT NULL DEFAULT 'active',
    severity        alarm_severity NOT NULL, -- Copied from threshold at trigger time
    triggered_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    resolved_at     TIMESTAMP WITH TIME ZONE,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    acknowledged_by UUID REFERENCES users(guid) ON DELETE SET NULL,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_alarms_device_guid ON alarms(device_guid) WHERE deleted_at IS NULL;
CREATE INDEX idx_alarms_status ON alarms(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_alarms_triggered_at ON alarms(triggered_at DESC) WHERE deleted_at IS NULL;
