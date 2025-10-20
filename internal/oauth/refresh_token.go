package oauth

import (
	"github.com/google/uuid"
)

type RefreshTokenChain struct {
	ID       uuid.UUID
	ClientID uuid.UUID
	UserID   uuid.UUID
	Scopes   []string
	IsNew    bool
}

type RefreshToken struct {
	Token string
	IsNew bool
}
