package audit

import (
	"context"

	"github.com/freekieb7/smauth/internal/database"
	"github.com/freekieb7/smauth/internal/telemetry"
)

type Store struct {
	Logger *telemetry.Logger
	DB     *database.Database
}

func NewStore(logger *telemetry.Logger, db *database.Database) Store {
	return Store{Logger: logger, DB: db}
}

func (s *Store) SaveLogEntry(ctx context.Context, entry LogEntry) error {
	_, err := s.DB.Conn.Exec(ctx, `
		INSERT INTO tbl_audit_log (id, event, account_id, resource, action, success, ip_address, user_agent, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, entry.ID, entry.Event, entry.AccountID, entry.Resource, entry.Action, entry.Success, entry.IPAddress, entry.UserAgent, entry.Details, entry.CreatedAt.UTC())
	return err
}
