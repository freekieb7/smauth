package account

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/freekieb7/smauth/internal/database"
	"github.com/freekieb7/smauth/internal/telemetry"
	"github.com/freekieb7/smauth/internal/util"
	"github.com/google/uuid"
)

type ClientService struct {
	Logger *telemetry.Logger
	DB     *database.Database
}

func NewClientService(logger *telemetry.Logger, db *database.Database) ClientService {
	return ClientService{
		Logger: logger,
		DB:     db,
	}
}

type ListClientsRequest struct {
	PageSize  int
	Token     string
	Direction string
}

type ListClientsResponse struct {
	Clients   []Client              `json:"items"`
	NextToken util.Optional[string] `json:"next_token"`
	PrevToken util.Optional[string] `json:"prev_token"`
}

type ListClientsCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
	Direction string    `json:"direction"`
}

func (s *ClientService) ListClientsPaginated(ctx context.Context, req ListClientsRequest) (ListClientsResponse, error) {
	// Set default page size if invalid
	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 25
	}

	cursor, err := decodeListClientsCursor(req.Token)
	if err != nil {
		s.Logger.Warn("Failed to decode page token", "token", req.Token, "error", err)
		return ListClientsResponse{}, errors.New("invalid page token")
	}

	// Extract direction from cursor, default to "next" for first page
	direction := "next"
	if cursor.E {
		direction = cursor.V.Direction
	}

	query := `SELECT id, secret, name, redirect_uris, is_public FROM tbl_client `
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
		return ListClientsResponse{}, err
	}
	defer rows.Close()

	clients := []Client{}
	for rows.Next() {
		var client Client
		if err := rows.Scan(&client.ID, &client.Secret, &client.Name, &client.RedirectURIs, &client.IsPublic); err != nil {
			s.Logger.Warn("Failed to scan client", "error", err)
			return ListClientsResponse{}, err
		}
		clients = append(clients, client)
	}

	if len(clients) == 0 {
		return ListClientsResponse{
			Clients: clients,
		}, nil
	}

	// Check if we have more results than requested (indicates more pages)
	hasMore := len(clients) > pageSize
	if hasMore {
		clients = clients[:pageSize] // Remove the extra record
	}

	// If fetching previous page, reverse results to maintain order
	if direction == "prev" {
		for i, j := 0, len(clients)-1; i < j; i, j = i+1, j-1 {
			clients[i], clients[j] = clients[j], clients[i]
		}
	}

	// Generate next/prev tokens based on actual data availability
	var nextCursor, prevCursor util.Optional[string]

	if len(clients) > 0 {
		// Generate next token (for older records)
		showNext := (direction == "next" && hasMore) || (direction == "prev")
		if showNext {
			if nextCursorStr, err := encodeListClientsCursor(ListClientsCursor{
				ID:        clients[len(clients)-1].ID,
				CreatedAt: clients[len(clients)-1].CreatedAt,
				Direction: "next",
			}); err == nil {
				nextCursor = util.Some(nextCursorStr)
			}
		}

		// Generate prev token (for newer records)
		showPrev := (direction == "prev" && hasMore) || (direction == "next" && req.Token != "")
		if showPrev {
			if prevCursorStr, err := encodeListClientsCursor(ListClientsCursor{
				ID:        clients[0].ID,
				CreatedAt: clients[0].CreatedAt,
				Direction: "prev",
			}); err == nil {
				prevCursor = util.Some(prevCursorStr)
			}
		}
	}

	return ListClientsResponse{
		Clients:   clients,
		NextToken: nextCursor,
		PrevToken: prevCursor,
	}, nil
}

func encodeListClientsCursor(cursor ListClientsCursor) (string, error) {
	data, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(data), nil
}

func decodeListClientsCursor(token string) (util.Optional[ListClientsCursor], error) {
	if token == "" {
		return util.None[ListClientsCursor](), nil
	}
	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return util.None[ListClientsCursor](), err
	}
	var cursor ListClientsCursor
	err = json.Unmarshal(data, &cursor)
	return util.Some(cursor), err
}
