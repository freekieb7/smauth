-- Migration: 20251020204307_create_users_table
-- Created: 2025-10-20T20:43:07+02:00
-- Description: create_users_table (rollback)

-- Add your down migration SQL here

DROP TABLE tbl_audit_log;
DROP TABLE tbl_refresh_token;
DROP TABLE tbl_refresh_token_chain;
DROP TABLE tbl_authorization_code;
DROP TABLE tbl_grant;
DROP TABLE tbl_access_token;
DROP TABLE tbl_account_permission;
DROP TABLE tbl_client CASCADE;
DROP TABLE tbl_scope;
DROP TABLE tbl_resource_server CASCADE;
DROP TABLE tbl_session;
DROP TABLE tbl_user CASCADE;
DROP TABLE tbl_account CASCADE;