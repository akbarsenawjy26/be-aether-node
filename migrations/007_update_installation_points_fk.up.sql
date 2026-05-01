-- Migration: 007_update_installation_points_fk.up.sql
-- Add foreign key constraints with proper indexing

-- Add foreign key from installation_points.device_guid -> devices.guid
ALTER TABLE installation_points
ADD CONSTRAINT fk_installation_points_device
FOREIGN KEY (device_guid) REFERENCES devices(guid) ON DELETE RESTRICT;

-- Add foreign key from installation_points.location_guid -> locations.guid
ALTER TABLE installation_points
ADD CONSTRAINT fk_installation_points_location
FOREIGN KEY (location_guid) REFERENCES locations(guid) ON DELETE RESTRICT;

-- Add foreign key from users.role_guid -> roles.guid (if roles table exists)
-- ALTER TABLE users
-- ADD CONSTRAINT fk_users_role
-- FOREIGN KEY (role_guid) REFERENCES roles(guid) ON DELETE SET NULL;
