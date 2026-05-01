-- Migration: 007_update_installation_points_fk.down.sql
-- Remove foreign key constraints

ALTER TABLE installation_points DROP CONSTRAINT IF EXISTS fk_installation_points_device;
ALTER TABLE installation_points DROP CONSTRAINT IF EXISTS fk_installation_points_location;
