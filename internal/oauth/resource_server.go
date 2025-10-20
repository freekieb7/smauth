package oauth

import (
	"time"

	"github.com/google/uuid"
)

type ResourceServer struct {
	ID        uuid.UUID
	URL       string
	Scopes    []Scope
	CreatedAt time.Time
	IsNew     bool
}
