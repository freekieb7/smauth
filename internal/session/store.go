package session

import (
	"context"
	"errors"
	"log/slog"

	"github.com/freekieb7/smauth/internal/database"
	"github.com/freekieb7/smauth/internal/util"
)

type Store struct {
	Logger *slog.Logger
	DB     *database.Database
}

func NewStore(logger *slog.Logger, db *database.Database) Store {
	return Store{
		Logger: logger,
		DB:     db,
	}
}

func (s *Store) New() (Session, error) {
	id, err := util.GenerateRandomString(32)
	if err != nil {
		return Session{}, err
	}

	session := Session{
		ID:     id,
		Values: map[ValueKey]any{},
		IsNew:  true,
	}

	return session, nil
}

func (s *Store) Get(ctx context.Context, sessionID string) (Session, error) {
	var session Session
	if err := s.DB.Conn.QueryRow(ctx, "SELECT id, user_id, data FROM tbl_session WHERE id=$1", sessionID).Scan(&session.ID, &session.UserID, &session.Values); err != nil {
		if errors.Is(err, database.ErrNoRows) {
			session, err = s.New()
			if err != nil {
				return Session{}, err
			}
			session.ID = sessionID
			return session, nil
		}
		return Session{}, err
	}

	return session, nil
}

func (s *Store) Save(ctx context.Context, session Session) (Session, error) {
	if session.IsNew {
		if _, err := s.DB.Conn.Exec(ctx, "INSERT INTO tbl_session (id, user_id, data, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW())", session.ID, session.UserID, session.Values); err != nil {
			s.Logger.Error("session store: failed to create session", "error", err)
			return Session{}, err
		}

		session.IsNew = false
		return session, nil
	}

	if _, err := s.DB.Conn.Exec(ctx, "UPDATE tbl_session SET user_id=$1, data=$2, updated_at=NOW() WHERE id=$3", session.UserID, session.Values, session.ID); err != nil {
		s.Logger.Error("session store: failed to update session", "error", err)
		return Session{}, err
	}

	return session, nil
}

func (s *Store) Delete(ctx context.Context, sessionID string) error {
	if _, err := s.DB.Conn.Exec(ctx, "DELETE FROM tbl_session WHERE id=$1", sessionID); err != nil {
		s.Logger.Error("session store: failed to delete session", "error", err)
		return err
	}
	return nil
}
