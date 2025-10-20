package audit

import (
	"encoding/json"
	"net"
	"time"

	"github.com/google/uuid"
)

type LogEntry struct {
	ID        uuid.UUID       `json:"id"`
	Event     EventType       `json:"event"`
	AccountID uuid.UUID       `json:"account_id"` // References Account.ID
	Resource  string          `json:"resource"`   // What was accessed/modified
	Action    string          `json:"action"`     // read, write, delete, etc.
	Success   bool            `json:"success"`
	IPAddress net.IP          `json:"ip_address,omitempty"`
	UserAgent string          `json:"user_agent,omitempty"`
	Details   json.RawMessage `json:"details,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type EventType string

const (
	// Authentication events
	EventUserLogin   EventType = "AUTH_LOGIN"
	EventUserLogout  EventType = "AUTH_LOGOUT"
	EventClientAuth  EventType = "AUTH_CLIENT"
	EventAuthFailure EventType = "AUTH_FAILURE"

	// Data access events
	EventDataRead   EventType = "DATA_READ"
	EventDataWrite  EventType = "DATA_WRITE"
	EventDataDelete EventType = "DATA_DELETE"

	// Administrative events
	EventPermissionChange EventType = "PERMISSION_CHANGE"
	EventAccountCreated   EventType = "ACCOUNT_CREATED"
	EventAccountDeleted   EventType = "ACCOUNT_DELETED"

	// OpenEHR specific
	EventEHRCreate         EventType = "EHR_CREATE"
	EventEHRRead           EventType = "EHR_READ"
	EventCompositionCreate EventType = "COMPOSITION_CREATE"
	EventAQLQuery          EventType = "AQL_QUERY"
)
