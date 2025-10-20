package oauth

import (
	"time"

	"github.com/google/uuid"
)

type AccessToken struct {
	Token     string
	AccountID uuid.UUID
	Scopes    []string
	ExpiresAt time.Time
	IsNew     bool
}
