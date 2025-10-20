package oauth

import "github.com/google/uuid"

type ResourceServer struct {
	ID     uuid.UUID
	URL    string
	Scopes []Scope
	IsNew  bool
}
