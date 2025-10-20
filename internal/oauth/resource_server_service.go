package oauth

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

type ResourceServerService struct {
	Logger *telemetry.Logger
	DB     *database.Database
}

func NewResourceServerService(logger *telemetry.Logger, db *database.Database) ResourceServerService {
	return ResourceServerService{
		Logger: logger,
		DB:     db,
	}
}

// Pagination types for client listing
type ListResourceServersRequest struct {
	PageSize  int
	Token     string
	Direction string
}

type ListResourceServersResponse struct {
	ResourceServers []ResourceServer
	NextToken       util.Optional[string]
	PrevToken       util.Optional[string]
}

type ListResourceServersCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
	Direction string    `json:"direction"`
}

func (s *ResourceServerService) ListResourceServersPaginated(ctx context.Context, req ListResourceServersRequest) (ListResourceServersResponse, error) {
	// Set default page size if invalid
	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 25
	}

	cursor, err := decodeListResourceServersCursor(req.Token)
	if err != nil {
		s.Logger.Warn("Failed to decode page token", "token", req.Token, "error", err)
		return ListResourceServersResponse{}, errors.New("invalid page token")
	}

	// Extract direction from cursor, default to "next" for first page
	direction := "next"
	if cursor.E {
		direction = cursor.V.Direction
	}

	query := `SELECT id, url, created_at FROM tbl_resource_server `
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
		s.Logger.Warn("Failed to query resource servers", "error", err)
		return ListResourceServersResponse{}, err
	}
	defer rows.Close()

	resourceServers := []ResourceServer{}
	for rows.Next() {
		var server ResourceServer
		if err := rows.Scan(&server.ID, &server.URL, &server.CreatedAt); err != nil {
			s.Logger.Warn("Failed to scan resource server", "error", err)
			return ListResourceServersResponse{}, err
		}
		resourceServers = append(resourceServers, server)
	}

	if len(resourceServers) == 0 {
		return ListResourceServersResponse{
			ResourceServers: resourceServers,
		}, nil
	}

	// Check if we have more results than requested (indicates more pages)
	hasMore := len(resourceServers) > pageSize
	if hasMore {
		resourceServers = resourceServers[:pageSize] // Remove the extra record
	}

	// If fetching previous page, reverse results to maintain order
	if direction == "prev" {
		for i, j := 0, len(resourceServers)-1; i < j; i, j = i+1, j-1 {
			resourceServers[i], resourceServers[j] = resourceServers[j], resourceServers[i]
		}
	}

	// Generate next/prev tokens based on actual data availability
	var nextCursor, prevCursor util.Optional[string]

	if len(resourceServers) > 0 {
		// Generate next token (for older records)
		showNext := (direction == "next" && hasMore) || (direction == "prev")
		if showNext {
			if nextCursorStr, err := encodeListResourceServersCursor(ListResourceServersCursor{
				ID:        resourceServers[len(resourceServers)-1].ID,
				CreatedAt: resourceServers[len(resourceServers)-1].CreatedAt,
				Direction: "next",
			}); err == nil {
				nextCursor = util.Some(nextCursorStr)
			}
		}

		// Generate prev token (for newer records)
		showPrev := (direction == "prev" && hasMore) || (direction == "next" && req.Token != "")
		if showPrev {
			if prevCursorStr, err := encodeListResourceServersCursor(ListResourceServersCursor{
				ID:        resourceServers[0].ID,
				CreatedAt: resourceServers[0].CreatedAt,
				Direction: "prev",
			}); err == nil {
				prevCursor = util.Some(prevCursorStr)
			}
		}
	}

	return ListResourceServersResponse{
		ResourceServers: resourceServers,
		NextToken:       nextCursor,
		PrevToken:       prevCursor,
	}, nil
}

func encodeListResourceServersCursor(cursor ListResourceServersCursor) (string, error) {
	data, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(data), nil
}

func decodeListResourceServersCursor(token string) (util.Optional[ListResourceServersCursor], error) {
	if token == "" {
		return util.None[ListResourceServersCursor](), nil
	}
	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return util.None[ListResourceServersCursor](), err
	}
	var cursor ListResourceServersCursor
	err = json.Unmarshal(data, &cursor)
	return util.Some(cursor), err
}
