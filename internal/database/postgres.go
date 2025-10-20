package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/freekieb7/smauth/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Database struct {
	Conn *pgx.Conn
}

func New() *Database {
	return &Database{}
}

func (db *Database) Connect(connString string) error {
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return err
	}

	db.Conn = conn
	return nil
}

func (db *Database) Close() error {
	return db.Conn.Close(context.Background())
}

func (db *Database) Ping(ctx context.Context) error {
	return db.Conn.Ping(ctx)
}

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (db *Database) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := db.Conn.Query(ctx, "SELECT id, email, password_hash, name, created_at, updated_at FROM tbl_user")
	if err != nil {
		return nil, fmt.Errorf("db: failed to query users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db: failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db: failed to iterate users: %w", err)
	}

	return users, nil
}

func (db *Database) GetUserByID(ctx context.Context, userID uuid.UUID) (User, error) {
	return db.GetUser(ctx, GetUserParams{
		ID: util.Some(userID),
	})
}

func (db *Database) GetUserByEmail(ctx context.Context, email string) (User, error) {
	return db.GetUser(ctx, GetUserParams{
		Email: util.Some(email),
	})
}

type GetUserParams struct {
	ID    util.Optional[uuid.UUID]
	Email util.Optional[string]
}

func (db *Database) GetUser(ctx context.Context, params GetUserParams) (User, error) {
	var user User
	var query strings.Builder
	var args []any
	argNum := 1

	query.WriteString("SELECT id, email, password_hash, name, created_at, updated_at FROM tbl_user WHERE 1=1")

	if params.ID.Some {
		query.WriteString(fmt.Sprintf(" AND id = $%d", argNum))
		args = append(args, params.ID.Data)
		argNum++
	}

	if params.Email.Some {
		query.WriteString(fmt.Sprintf(" AND email = $%d", argNum))
		args = append(args, params.Email.Data)
		argNum++
	}

	err := db.Conn.QueryRow(ctx, query.String(), args...).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return user, fmt.Errorf("db: failed to get user: %w", err)
	}

	return user, nil
}

type CreateUserParams struct {
	Email        string
	PasswordHash string
	Name         string
}

func (db *Database) CreateUser(ctx context.Context, params CreateUserParams) (User, error) {
	user := User{
		ID:           uuid.New(),
		Email:        params.Email,
		PasswordHash: params.PasswordHash,
		Name:         params.Name,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := db.Conn.Exec(ctx,
		"INSERT INTO tbl_user (id, email, password_hash, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		user.ID, user.Email, user.PasswordHash, user.Name, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return user, fmt.Errorf("db: failed to create user: %w", err)
	}

	return user, nil
}

type UpdateUserParams struct {
	Email        util.Optional[string]
	PasswordHash util.Optional[string]
	Name         util.Optional[string]
}

func (db *Database) UpdateUserByID(ctx context.Context, id uuid.UUID, params UpdateUserParams) error {
	var query strings.Builder
	var args []any
	argNum := 1

	query.WriteString("UPDATE tbl_user SET ")

	if params.Email.Some {
		query.WriteString(fmt.Sprintf("email = $%d, ", argNum))
		args = append(args, params.Email.Data)
		argNum++
	}

	if params.PasswordHash.Some {
		query.WriteString(fmt.Sprintf("password_hash = $%d, ", argNum))
		args = append(args, params.PasswordHash.Data)
		argNum++
	}

	if params.Name.Some {
		query.WriteString(fmt.Sprintf("name = $%d, ", argNum))
		args = append(args, params.Name.Data)
		argNum++
	}

	if len(args) == 0 {
		// Nothing to update
		return nil
	}

	// Append updated_at
	query.WriteString(fmt.Sprintf("updated_at = $%d ", argNum))
	args = append(args, time.Now())
	argNum++

	// Where clause
	query.WriteString(fmt.Sprintf("WHERE id = $%d", argNum))
	args = append(args, id)

	_, err := db.Conn.Exec(ctx, query.String(), args...)
	return err
}

func (db *Database) DeleteUserByID(ctx context.Context, userID uuid.UUID) error {
	_, err := db.Conn.Exec(ctx, "DELETE FROM tbl_user WHERE id = $1", userID)
	return err
}

type Client struct {
	ID           uuid.UUID
	Name         string
	Secret       string
	RedirectURIs []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (db *Database) ListClients(ctx context.Context) ([]Client, error) {
	rows, err := db.Conn.Query(ctx, "SELECT id, name, secret, redirect_uris, created_at, updated_at FROM tbl_client")
	if err != nil {
		return nil, fmt.Errorf("db: failed to query clients: %w", err)
	}
	defer rows.Close()

	var clients []Client
	for rows.Next() {
		var client Client
		if err := rows.Scan(&client.ID, &client.Name, &client.Secret, &client.RedirectURIs, &client.CreatedAt, &client.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db: failed to scan client: %w", err)
		}
		clients = append(clients, client)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db: failed to iterate clients: %w", err)
	}

	return clients, nil
}

type CreateClientParams struct {
	Name         string
	Secret       string
	Scopes       []string
	RedirectURIs []string
}

func (db *Database) CreateClient(ctx context.Context, params CreateClientParams) (Client, error) {
	client := Client{
		ID:           uuid.New(),
		Name:         params.Name,
		Secret:       params.Secret,
		RedirectURIs: params.RedirectURIs,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := db.Conn.Exec(ctx,
		"INSERT INTO tbl_client (id, name, secret, redirect_uris, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		client.ID, client.Name, client.Secret, client.RedirectURIs, client.CreatedAt, client.UpdatedAt)
	if err != nil {
		return client, fmt.Errorf("db: failed to create client: %w", err)
	}

	return client, nil
}

type UpdateClientParams struct {
	Name         util.Optional[string]
	RedirectURIs util.Optional[[]string]
}

func (db *Database) UpdateClientByID(ctx context.Context, id uuid.UUID, params UpdateClientParams) error {
	var query strings.Builder
	var args []any
	argNum := 1

	query.WriteString("UPDATE tbl_client SET ")

	if params.Name.Some {
		query.WriteString(fmt.Sprintf("name = $%d, ", argNum))
		args = append(args, params.Name.Data)
		argNum++
	}

	if params.RedirectURIs.Some {
		query.WriteString(fmt.Sprintf("redirect_uris = $%d, ", argNum))
		args = append(args, params.RedirectURIs.Data)
		argNum++
	}

	if len(args) == 0 {
		// Nothing to update
		return nil
	}

	// Append updated_at
	query.WriteString(fmt.Sprintf("updated_at = $%d ", argNum))
	args = append(args, time.Now())
	argNum++

	// Where clause
	query.WriteString(fmt.Sprintf("WHERE id = $%d", argNum))
	args = append(args, id)

	_, err := db.Conn.Exec(ctx, query.String(), args...)
	return err
}

func (db *Database) DeleteClientByID(ctx context.Context, clientID uuid.UUID) error {
	_, err := db.Conn.Exec(ctx, "DELETE FROM tbl_client WHERE id = $1", clientID)
	return err
}
