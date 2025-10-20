-- Migration: 20251020204307_create_users_table
-- Created: 2025-10-20T20:43:07+02:00
-- Description: create_users_table

-- Add your up migration SQL here

CREATE TABLE tbl_account (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL, -- 'user', 'client', 'system'
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_tbl_account_type ON tbl_account(type);
CREATE INDEX idx_tbl_account_created_at ON tbl_account(created_at);

-- User-specific data (extends account)
CREATE TABLE tbl_user (
    id UUID PRIMARY KEY REFERENCES tbl_account(id) ON DELETE CASCADE,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL,
    is_email_verified BOOLEAN NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_tbl_user_email ON tbl_user(email);
CREATE INDEX idx_tbl_user_role ON tbl_user(role);
CREATE INDEX idx_tbl_user_created_at ON tbl_user(created_at);

-- Client-specific data (extends account) 
CREATE TABLE tbl_client (
    id UUID PRIMARY KEY REFERENCES tbl_account(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    secret TEXT NOT NULL,
    redirect_uris TEXT[] NOT NULL,
    is_public BOOLEAN NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_tbl_client_name ON tbl_client(name);
CREATE INDEX idx_tbl_client_created_at ON tbl_client(created_at);

CREATE TABLE tbl_session (
    id TEXT PRIMARY KEY,
    account_id UUID REFERENCES tbl_account(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_tbl_session_account_id ON tbl_session(account_id);
CREATE INDEX idx_tbl_session_created_at ON tbl_session(created_at);

CREATE TABLE tbl_resource_server (
    id UUID PRIMARY KEY,
    url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
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
    UNIQUE (name)
);

CREATE INDEX idx_tbl_scope_resource_server_id ON tbl_scope(resource_server_id);
CREATE INDEX idx_tbl_scope_name ON tbl_scope(name);
CREATE INDEX idx_tbl_scope_created_at ON tbl_scope(created_at);

CREATE TABLE tbl_account_permission (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES tbl_account(id) ON DELETE CASCADE,
    scope_id UUID NOT NULL REFERENCES tbl_scope(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL,
    UNIQUE (account_id, scope_id)
);

CREATE INDEX idx_tbl_account_permission_account_id ON tbl_account_permission(account_id);
CREATE INDEX idx_tbl_account_permission_scope_id ON tbl_account_permission(scope_id);
CREATE INDEX idx_tbl_account_permission_created_at ON tbl_account_permission(created_at);

CREATE TABLE tbl_access_token (
    token TEXT PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES tbl_account(id) ON DELETE CASCADE,
    scopes TEXT[] NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_tbl_access_token_account_id ON tbl_access_token(account_id);
CREATE INDEX idx_tbl_access_token_expires_at ON tbl_access_token(expires_at);
CREATE INDEX idx_tbl_access_token_created_at ON tbl_access_token(created_at);

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
CREATE INDEX idx_tbl_grant_created_at ON tbl_grant(created_at);

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
CREATE INDEX idx_tbl_authorization_code_used_at ON tbl_authorization_code(used_at);
CREATE INDEX idx_tbl_authorization_code_created_at ON tbl_authorization_code(created_at);

CREATE TABLE tbl_refresh_token_chain (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES tbl_user(id) ON DELETE CASCADE,
    client_id UUID REFERENCES tbl_client(id) ON DELETE CASCADE,
    scopes TEXT[] NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_tbl_refresh_token_chain_user_id ON tbl_refresh_token_chain(user_id);
CREATE INDEX idx_tbl_refresh_token_chain_client_id ON tbl_refresh_token_chain(client_id);
CREATE INDEX idx_tbl_refresh_token_chain_created_at ON tbl_refresh_token_chain(created_at);

CREATE TABLE tbl_refresh_token (
    token TEXT PRIMARY KEY,
    chain_id UUID NOT NULL REFERENCES tbl_refresh_token_chain(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    UNIQUE (chain_id, used_at)
);

CREATE INDEX idx_tbl_refresh_token_chain_id ON tbl_refresh_token(chain_id);
CREATE INDEX idx_tbl_refresh_token_expires_at ON tbl_refresh_token(expires_at);
CREATE INDEX idx_tbl_refresh_token_used_at ON tbl_refresh_token(used_at);
CREATE INDEX idx_tbl_refresh_token_created_at ON tbl_refresh_token(created_at);

CREATE TABLE tbl_audit_log (
    id UUID PRIMARY KEY,
    event TEXT NOT NULL,
    account_id UUID NOT NULL REFERENCES tbl_account(id) ON DELETE CASCADE,
    resource TEXT NOT NULL,
    action TEXT NOT NULL,
    success BOOLEAN NOT NULL,
    ip_address INET NOT NULL,
    user_agent TEXT NOT NULL,
    details JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_tbl_audit_log_account_id ON tbl_audit_log(account_id);
CREATE INDEX idx_tbl_audit_log_event ON tbl_audit_log(event);
CREATE INDEX idx_tbl_audit_log_created_at ON tbl_audit_log(created_at);

INSERT INTO tbl_scope (id, resource_server_id, name, description, created_at)
VALUES (gen_random_uuid(), NULL, 'offline_access', 'Allow to keep you signed in', NOW()),
       (gen_random_uuid(), NULL, 'openid', 'Allow to access user identity information', NOW()),
       (gen_random_uuid(), NULL, 'email', 'Allow to access your email', NOW());

INSERT INTO tbl_account (id, type, created_at)
VALUES (gen_random_uuid(), 'system', NOW());