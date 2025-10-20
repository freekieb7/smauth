-- Migration: 20251105105304_microsoft-sso
-- Created: 2025-11-05T10:53:04+01:00
-- Description: microsoft-sso

-- Add your up migration SQL here

CREATE TABLE tbl_jwks (
    id UUID PRIMARY KEY,
    origin TEXT NOT NULL, -- e.g., 'microsoft', 'oauth_provider_x'
    kid TEXT NOT NULL,
    kty TEXT NOT NULL,
    alg TEXT NOT NULL,
    use TEXT NOT NULL,
    n TEXT NOT NULL,
    e TEXT NOT NULL,
    x5c TEXT[] NOT NULL,
    x5t TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);