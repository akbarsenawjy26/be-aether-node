-- Migration: 009_alarms.down.sql
-- Drop alarms table and enum

DROP TABLE IF EXISTS alarms;
DROP TYPE IF EXISTS alarm_status;
