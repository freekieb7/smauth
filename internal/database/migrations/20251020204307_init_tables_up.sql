-- Migration: 20251020204307_create_users_table
-- Created: 2025-10-20T20:43:07+02:00
-- Description: create_users_table

-- Add your up migration SQL here
CREATE TABLE tbl_user (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL,
    is_email_verified BOOLEAN NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_tbl_user_email ON tbl_user(email);
CREATE INDEX idx_tbl_user_role ON tbl_user(role);
CREATE INDEX idx_tbl_user_created_at ON tbl_user(created_at);

CREATE TABLE tbl_session (
    id TEXT PRIMARY KEY,
    user_id UUID REFERENCES tbl_user(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_tbl_session_user_id ON tbl_session(user_id);

CREATE TABLE tbl_resource_server (
    id UUID PRIMARY KEY,
    url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE (url)
);

CREATE INDEX idx_tbl_resource_server_url ON tbl_resource_server(url);
CREATE INDEX idx_tbl_resource_server_created_at ON tbl_resource_server(created_at);

CREATE TABLE tbl_scope (
    id UUID PRIMARY KEY,
    resource_server_id UUID REFERENCES tbl_resource_server(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE (name)
);

CREATE INDEX idx_tbl_scope_resource_server_id ON tbl_scope(resource_server_id);
CREATE INDEX idx_tbl_scope_name ON tbl_scope(name);

CREATE TABLE tbl_client (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    secret TEXT NOT NULL,
    redirect_uris TEXT[] NOT NULL,
    is_public BOOLEAN NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_tbl_client_name ON tbl_client(name);
CREATE INDEX idx_tbl_client_created_at ON tbl_client(created_at);

CREATE TABLE tbl_user_permission (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES tbl_user(id) ON DELETE CASCADE,
    scope_id UUID NOT NULL REFERENCES tbl_scope(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL,
    UNIQUE (user_id, scope_id)
);

CREATE INDEX idx_tbl_user_permission_user_id ON tbl_user_permission(user_id);
CREATE INDEX idx_tbl_user_permission_scope_id ON tbl_user_permission(scope_id);
CREATE INDEX idx_tbl_user_permission_created_at ON tbl_user_permission(created_at);

CREATE TABLE tbl_client_permission (
    id UUID PRIMARY KEY,
    client_id UUID NOT NULL REFERENCES tbl_client(id) ON DELETE CASCADE,
    scope_id UUID NOT NULL REFERENCES tbl_scope(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL,
    UNIQUE (client_id, scope_id)
);

CREATE INDEX idx_tbl_client_permission_client_id ON tbl_client_permission(client_id);
CREATE INDEX idx_tbl_client_permission_scope_id ON tbl_client_permission(scope_id);
CREATE INDEX idx_tbl_client_permission_created_at ON tbl_client_permission(created_at);

CREATE TABLE tbl_access_token (
    token TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES tbl_user(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES tbl_client(id) ON DELETE CASCADE,
    scopes TEXT[] NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_tbl_access_token_user_id ON tbl_access_token(user_id);
CREATE INDEX idx_tbl_access_token_client_id ON tbl_access_token(client_id);
CREATE INDEX idx_tbl_access_token_expires_at ON tbl_access_token(expires_at);

CREATE TABLE tbl_grant (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES tbl_user(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES tbl_client(id) ON DELETE CASCADE,
    scope_id UUID NOT NULL REFERENCES tbl_scope(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL,
    UNIQUE (user_id, client_id, scope_id)
);

CREATE INDEX idx_tbl_grant_user_id ON tbl_grant(user_id);
CREATE INDEX idx_tbl_grant_client_id ON tbl_grant(client_id);
CREATE INDEX idx_tbl_grant_scope_id ON tbl_grant(scope_id);

CREATE TABLE tbl_authorization_code (
    code TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES tbl_user(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES tbl_client(id) ON DELETE CASCADE,
    redirect_uri TEXT NOT NULL,
    scopes TEXT[] NOT NULL,
    code_challenge TEXT NOT NULL,
    code_method TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_tbl_authorization_code_user_id ON tbl_authorization_code(user_id);
CREATE INDEX idx_tbl_authorization_code_client_id ON tbl_authorization_code(client_id);
CREATE INDEX idx_tbl_authorization_code_expires_at ON tbl_authorization_code(expires_at);

CREATE TABLE tbl_refresh_token_chain (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES tbl_user(id) ON DELETE CASCADE,
    client_id UUID REFERENCES tbl_client(id) ON DELETE CASCADE,
    scopes TEXT[] NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE tbl_refresh_token (
    token TEXT PRIMARY KEY,
    chain_id UUID NOT NULL REFERENCES tbl_refresh_token_chain(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    UNIQUE (chain_id, used_at)
);

CREATE INDEX idx_tbl_refresh_token_chain_user_id ON tbl_refresh_token_chain(user_id);

INSERT INTO tbl_scope (id, resource_server_id, name, description, created_at, updated_at)
VALUES (gen_random_uuid(), NULL, 'offline_access', 'Allow to keep you signed in', NOW(), NOW()),
       (gen_random_uuid(), NULL, 'openid', 'Allow to access user identity information', NOW(), NOW()),
       (gen_random_uuid(), NULL, 'email', 'Allow to access your email', NOW(), NOW());