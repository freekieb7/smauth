package oauth

import (
	"time"

	"github.com/freekieb7/smauth/internal/util"
	"github.com/google/uuid"
)

type AccessToken struct {
	Token     string
	ClientID  util.Optional[uuid.UUID]
	UserID    util.Optional[uuid.UUID]
	Scopes    []string
	ExpiresAt time.Time
	IsNew     bool
}
