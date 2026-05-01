-- Migration: 001_users.down.sql
-- Drop users table

DROP TABLE IF EXISTS users CASCADE;
DROP EXTENSION IF EXISTS "uuid-ossp";
