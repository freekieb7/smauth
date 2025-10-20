-- Migration: 20251020222515_add_clients_table
-- Created: 2025-10-20T22:25:15+02:00
-- Description: add_clients_table (rollback)

-- Add your down migration SQL here

DROP INDEX IF EXISTS idx_client_name;
DROP INDEX IF EXISTS idx_client_created_at;
DROP TABLE IF EXISTS tbl_client;