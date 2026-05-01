-- Migration: 006_refresh_tokens.down.sql
-- Drop refresh_tokens table

DROP TABLE IF EXISTS refresh_tokens CASCADE;
