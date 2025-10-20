-- Migration: 20251020204307_create_users_table
-- Created: 2025-10-20T20:43:07+02:00
-- Description: create_users_table (rollback)

-- Add your down migration SQL here
DROP INDEX IF EXISTS idx_tbl_user_created_at;
DROP INDEX IF EXISTS idx_tbl_user_email;
DROP TABLE IF EXISTS tbl_user;

