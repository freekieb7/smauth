package auth

import "github.com/google/uuid"

type User struct {
	ID              uuid.UUID
	Email           string
	PasswordHash    string
	Role            Role
	IsEmailVerified bool
	IsNew           bool
}

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleUser:
		return true
	default:
		return false
	}
}

func (r Role) String() string {
	return string(r)
}
