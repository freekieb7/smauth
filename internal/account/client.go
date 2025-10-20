package account

import (
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID           uuid.UUID
	Secret       string
	Name         string
	RedirectURIs []string
	IsPublic     bool
	IsNew        bool
	CreatedAt    time.Time
}
