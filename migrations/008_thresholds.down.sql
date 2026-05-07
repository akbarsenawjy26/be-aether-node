-- Migration: 008_thresholds.down.sql
-- Drop thresholds table and enum

DROP TABLE IF EXISTS thresholds;
DROP TYPE IF EXISTS alarm_severity;
