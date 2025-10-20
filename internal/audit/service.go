package audit

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/freekieb7/smauth/internal/database"
	"github.com/freekieb7/smauth/internal/telemetry"
	"github.com/freekieb7/smauth/internal/util"
	"github.com/google/uuid"
)

type Service struct {
	Logger *telemetry.Logger
	DB     *database.Database
	Store  *Store
}

func NewService(logger *telemetry.Logger, db *database.Database, store *Store) Service {
	return Service{
		Logger: logger,
		DB:     db,
		Store:  store,
	}
}

type LogEventRequest struct {
	Event     EventType
	AccountID uuid.UUID
	Resource  string
	Action    string
	Success   bool
	IPAddress string
	UserAgent string
	Details   map[string]any
}

func (s *Service) LogEvent(ctx context.Context, req LogEventRequest) error {
	// Create a new log entry
	details, err := json.Marshal(req.Details)
	if err != nil {
		s.Logger.Error("Failed to marshal audit log details", "error", err, "event", string(req.Event), "account_id", req.AccountID)
		return err
	}

	host, _, err := net.SplitHostPort(req.IPAddress)
	if err != nil {
		panic(err)
	}

	ip := net.ParseIP(host)

	id, err := uuid.NewV7()
	if err != nil {
		s.Logger.Error("Failed to generate UUID for audit log", "error", err, "event", string(req.Event), "account_id", req.AccountID)
		return err
	}

	entry := LogEntry{
		ID:        id,
		Event:     req.Event,
		AccountID: req.AccountID,
		Resource:  req.Resource,
		Action:    req.Action,
		Success:   req.Success,
		IPAddress: ip,
		UserAgent: req.UserAgent,
		Details:   details,
		CreatedAt: time.Now(),
	}

	if err := s.Store.SaveLogEntry(ctx, entry); err != nil {
		s.Logger.Error("Failed to create audit log", "error", err, "event", string(req.Event), "account_id", req.AccountID)
		return err
	}

	return nil
}

// Pagination types for audit log listing
type ListLogEntriesRequest struct {
	PageSize int
	Token    string
}

type ListLogEntriesResponse struct {
	LogEntries []LogEntry
	NextToken  util.Optional[string]
	PrevToken  util.Optional[string]
}

type ListLogEntriesCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
	Direction string    `json:"direction"`
}

func (s *Service) ListLogEntriesPaginated(ctx context.Context, req ListLogEntriesRequest) (ListLogEntriesResponse, error) {
	// Set default page size if invalid
	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 25
	}

	cursor, err := decodeListLogEntriesCursor(req.Token)
	if err != nil {
		s.Logger.Warn("Failed to decode page token", "token", req.Token, "error", err)
		return ListLogEntriesResponse{}, errors.New("invalid page token")
	}

	// Extract direction from cursor, default to "next" for first page
	direction := "next"
	if cursor.E {
		direction = cursor.V.Direction
	}

	query := `SELECT id, event, account_id, resource, action, success, ip_address, user_agent, details, created_at FROM tbl_audit_log `
	args := []any{}
	argIdx := 1

	order := "DESC"
	cmp := "<"

	if direction == "prev" {
		order = "ASC"
		cmp = ">"
	}

	if cursor.E {
		query += fmt.Sprintf("WHERE (created_at, id) %s ($%d, $%d) ", cmp, argIdx, argIdx+1)
		args = append(args, cursor.V.CreatedAt, cursor.V.ID)
		argIdx += 2
	}

	query += fmt.Sprintf("ORDER BY created_at %s, id %s ", order, order)
	query += fmt.Sprintf("LIMIT $%d", argIdx)
	args = append(args, pageSize+1) // Fetch one extra to detect more pages

	rows, err := s.DB.Conn.Query(ctx, query, args...)
	if err != nil {
		s.Logger.Warn("Failed to query audit logs", "error", err)
		return ListLogEntriesResponse{}, err
	}
	defer rows.Close()

	logEntries := []LogEntry{}
	for rows.Next() {
		var entry LogEntry
		if err := rows.Scan(&entry.ID, &entry.Event, &entry.AccountID, &entry.Resource, &entry.Action, &entry.Success, &entry.IPAddress, &entry.UserAgent, &entry.Details, &entry.CreatedAt); err != nil {
			s.Logger.Warn("Failed to scan log entry", "error", err)
			return ListLogEntriesResponse{}, err
		}
		logEntries = append(logEntries, entry)
	}

	if len(logEntries) == 0 {
		return ListLogEntriesResponse{
			LogEntries: logEntries,
		}, nil
	}

	// Check if we have more results than requested (indicates more pages)
	hasMore := len(logEntries) > pageSize
	if hasMore {
		logEntries = logEntries[:pageSize] // Remove the extra record
	}

	// If fetching previous page, reverse results to maintain order
	if direction == "prev" {
		for i, j := 0, len(logEntries)-1; i < j; i, j = i+1, j-1 {
			logEntries[i], logEntries[j] = logEntries[j], logEntries[i]
		}
	}

	// Generate next/prev tokens based on actual data availability
	var nextCursor, prevCursor util.Optional[string]

	if len(logEntries) > 0 {
		// Generate next token (for older records)
		showNext := (direction == "next" && hasMore) || (direction == "prev")
		if showNext {
			if nextCursorStr, err := encodeListLogEntriesCursor(ListLogEntriesCursor{
				ID:        logEntries[len(logEntries)-1].ID,
				CreatedAt: logEntries[len(logEntries)-1].CreatedAt,
				Direction: "next",
			}); err == nil {
				nextCursor = util.Some(nextCursorStr)
			}
		}

		// Generate prev token (for newer records)
		showPrev := (direction == "prev" && hasMore) || (direction == "next" && req.Token != "")
		if showPrev {
			if prevCursorStr, err := encodeListLogEntriesCursor(ListLogEntriesCursor{
				ID:        logEntries[0].ID,
				CreatedAt: logEntries[0].CreatedAt,
				Direction: "prev",
			}); err == nil {
				prevCursor = util.Some(prevCursorStr)
			}
		}
	}

	return ListLogEntriesResponse{
		LogEntries: logEntries,
		NextToken:  nextCursor,
		PrevToken:  prevCursor,
	}, nil
}

func encodeListLogEntriesCursor(cursor ListLogEntriesCursor) (string, error) {
	data, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(data), nil
}

func decodeListLogEntriesCursor(token string) (util.Optional[ListLogEntriesCursor], error) {
	if token == "" {
		return util.None[ListLogEntriesCursor](), nil
	}
	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return util.None[ListLogEntriesCursor](), err
	}
	var cursor ListLogEntriesCursor
	err = json.Unmarshal(data, &cursor)
	return util.Some(cursor), err
}
