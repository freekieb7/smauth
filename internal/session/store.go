package session

import (
	"context"
	"errors"

	"github.com/freekieb7/smauth/internal/database"
	"github.com/freekieb7/smauth/internal/http"
	"github.com/freekieb7/smauth/internal/telemetry"
	"github.com/freekieb7/smauth/internal/util"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type Store struct {
	Logger *telemetry.Logger
	DB     *database.Database
}

func NewStore(logger *telemetry.Logger, db *database.Database) Store {
	return Store{
		Logger: logger,
		DB:     db,
	}
}

func (s *Store) NewSession() (Session, error) {
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

func (s *Store) GetSession(ctx context.Context, sessionID string) (Session, error) {
	var session Session
	if err := s.DB.Conn.QueryRow(ctx, "SELECT id, account_id, data FROM tbl_session WHERE id=$1", sessionID).Scan(&session.ID, &session.AccountID, &session.Values); err != nil {
		if errors.Is(err, database.ErrNoRows) {
			return Session{}, ErrSessionNotFound
		}
		return Session{}, err
	}

	return session, nil
}

func (s *Store) SaveSession(ctx context.Context, session Session) (Session, error) {
	if session.IsNew {
		if _, err := s.DB.Conn.Exec(ctx, "INSERT INTO tbl_session (id, account_id, data, created_at) VALUES ($1, $2, $3, NOW())", session.ID, session.AccountID, session.Values); err != nil {
			s.Logger.Error("session store: failed to create session", "error", err)
			return Session{}, err
		}

		session.IsNew = false
		return session, nil
	}

	if _, err := s.DB.Conn.Exec(ctx, "UPDATE tbl_session SET account_id=$1, data=$2 WHERE id=$3", session.AccountID, session.Values, session.ID); err != nil {
		s.Logger.Error("session store: failed to update session", "error", err)
		return Session{}, err
	}

	return session, nil
}

func (s *Store) DeleteSession(ctx context.Context, sessionID string) error {
	if _, err := s.DB.Conn.Exec(ctx, "DELETE FROM tbl_session WHERE id=$1", sessionID); err != nil {
		s.Logger.Error("session store: failed to delete session", "error", err)
		return err
	}
	return nil
}

// RegenerateSession creates a new session ID for the given session and updates it in the database.
func (s *Store) RegenerateSession(ctx context.Context, session Session) (Session, error) {
	newSessionID, err := util.GenerateRandomString(32)
	if err != nil {
		s.Logger.Error("session store: failed to generate new session ID", "error", err)
		return Session{}, err
	}

	if _, err := s.DB.Conn.Exec(ctx, "UPDATE tbl_session SET id=$1 WHERE id=$2", newSessionID, session.ID); err != nil {
		s.Logger.Error("session store: failed to regenerate session ID", "error", err)
		return Session{}, err
	}

	session.ID = newSessionID
	return session, nil
}

func (s *Store) SetCookie(ctx context.Context, res *http.Response, sessionID string) {
	res.SetCookie(http.Cookie{
		Name:     "SID",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Store) ClearCookie(ctx context.Context, res *http.Response) {
	res.SetCookie(http.Cookie{
		Name:     "SID",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
