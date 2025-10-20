-- Migration: 20251020204408_add_sessions_table
-- Created: 2025-10-20T20:44:08+02:00
-- Description: add_sessions_table (rollback)

-- Add your down migration SQL here
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_token_hash;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP TABLE IF EXISTS sessions;

