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

type UserService struct {
	Logger *telemetry.Logger
	DB     *database.Database
}

func NewUserService(logger *telemetry.Logger, db *database.Database) UserService {
	return UserService{
		Logger: logger,
		DB:     db,
	}
}

type ListUsersRequest struct {
	PageSize  int
	Token     string
	Direction string
}

type ListUsersResponse struct {
	Users     []User
	NextToken util.Optional[string]
	PrevToken util.Optional[string]
}

type ListUsersCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
	Direction string    `json:"direction"`
}

func (s *UserService) ListUsersPaginated(ctx context.Context, req ListUsersRequest) (ListUsersResponse, error) {
	// Set default page size if invalid
	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 25
	}

	cursor, err := decodeListUsersCursor(req.Token)
	if err != nil {
		s.Logger.Warn("Failed to decode page token", "token", req.Token, "error", err)
		return ListUsersResponse{}, errors.New("invalid page token")
	}

	// Extract direction from cursor, default to "next" for first page
	direction := "next"
	if cursor.E {
		direction = cursor.V.Direction
	}

	query := `SELECT id, email, role, is_email_verified, created_at FROM tbl_user `
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
		return ListUsersResponse{}, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Email, &user.Role, &user.IsEmailVerified, &user.CreatedAt); err != nil {
			s.Logger.Warn("Failed to scan user", "error", err)
			return ListUsersResponse{}, err
		}
		users = append(users, user)
	}

	if len(users) == 0 {
		return ListUsersResponse{
			Users: users,
		}, nil
	}

	// Check if we have more results than requested (indicates more pages)
	hasMore := len(users) > pageSize
	if hasMore {
		users = users[:pageSize] // Remove the extra record
	}

	// If fetching previous page, reverse results to maintain order
	if direction == "prev" {
		for i, j := 0, len(users)-1; i < j; i, j = i+1, j-1 {
			users[i], users[j] = users[j], users[i]
		}
	}

	// Generate next/prev tokens based on actual data availability
	var nextCursor, prevCursor util.Optional[string]

	if len(users) > 0 {
		// Generate next token (for older records)
		showNext := (direction == "next" && hasMore) || (direction == "prev")
		if showNext {
			if nextCursorStr, err := encodeListUsersCursor(ListUsersCursor{
				ID:        users[len(users)-1].ID,
				CreatedAt: users[len(users)-1].CreatedAt,
				Direction: "next",
			}); err == nil {
				nextCursor = util.Some(nextCursorStr)
			}
		}

		// Generate prev token (for newer records)
		showPrev := (direction == "prev" && hasMore) || (direction == "next" && req.Token != "")
		if showPrev {
			if prevCursorStr, err := encodeListUsersCursor(ListUsersCursor{
				ID:        users[0].ID,
				CreatedAt: users[0].CreatedAt,
				Direction: "prev",
			}); err == nil {
				prevCursor = util.Some(prevCursorStr)
			}
		}
	}

	return ListUsersResponse{
		Users:     users,
		NextToken: nextCursor,
		PrevToken: prevCursor,
	}, nil
}

func encodeListUsersCursor(cursor ListUsersCursor) (string, error) {
	data, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(data), nil
}

func decodeListUsersCursor(token string) (util.Optional[ListUsersCursor], error) {
	if token == "" {
		return util.None[ListUsersCursor](), nil
	}
	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return util.None[ListUsersCursor](), err
	}
	var cursor ListUsersCursor
	err = json.Unmarshal(data, &cursor)
	return util.Some(cursor), err
}
