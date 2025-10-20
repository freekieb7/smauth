package oauth

import (
	"strings"

	"github.com/google/uuid"
)

type Scope struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

func (s Scope) PlainScopeName() string {
	parts := strings.Split(s.Name, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return s.Name
}
