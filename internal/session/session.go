package session

import (
	"github.com/freekieb7/smauth/internal/util"
	"github.com/google/uuid"
)

type ValueKey string

type Session struct {
	ID     string
	UserID util.Optional[uuid.UUID]
	Values map[ValueKey]any
	IsNew  bool
}
