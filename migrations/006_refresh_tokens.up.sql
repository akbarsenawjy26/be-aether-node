-- Migration: 006_refresh_tokens.up.sql
-- Create refresh_tokens table

CREATE TABLE refresh_tokens (
    guid        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_guid   UUID NOT NULL REFERENCES users(guid) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL,
    expires_at  TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_guid ON refresh_tokens(user_guid);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
