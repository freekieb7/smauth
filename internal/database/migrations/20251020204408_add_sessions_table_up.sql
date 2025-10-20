-- Migration: 20251020204408_add_sessions_table
-- Created: 2025-10-20T20:44:08+02:00
-- Description: add_sessions_table

-- Add your up migration SQL here
CREATE TABLE tbl_session (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES tbl_user(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    last_used_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_tbl_session_user_id ON tbl_session(user_id);
CREATE INDEX idx_tbl_session_token_hash ON tbl_session(token_hash);
CREATE INDEX idx_tbl_session_expires_at ON tbl_session(expires_at);

