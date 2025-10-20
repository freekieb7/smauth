package oauth

import (
	"github.com/google/uuid"
)

type AuthorizationCode struct {
	Code          string
	ClientID      uuid.UUID
	UserID        uuid.UUID
	RedirectURI   string
	Scopes        []string
	CodeChallenge string
	CodeMethod    string
	IsNew         bool
}
