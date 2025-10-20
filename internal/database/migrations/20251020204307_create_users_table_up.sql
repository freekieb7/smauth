-- Migration: 20251020204307_create_users_table
-- Created: 2025-10-20T20:43:07+02:00
-- Description: create_users_table

-- Add your up migration SQL here
CREATE TABLE tbl_user (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_tbl_user_email ON tbl_user(email);
CREATE INDEX idx_tbl_user_created_at ON tbl_user(created_at);

