-- Migration: 20251020222515_add_clients_table
-- Created: 2025-10-20T22:25:15+02:00
-- Description: add_clients_table

-- Add your up migration SQL here

CREATE TABLE tbl_client (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    secret TEXT NOT NULL,
    redirect_uris TEXT[] NOT NULL,
    scopes TEXT[] NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_client_name ON tbl_client(name);
CREATE INDEX idx_client_created_at ON tbl_client(created_at);