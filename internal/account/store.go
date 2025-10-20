package account

import (
	"context"
	"errors"
	"net/mail"
	"time"

	"github.com/freekieb7/smauth/internal/database"
	"github.com/freekieb7/smauth/internal/telemetry"
	"github.com/freekieb7/smauth/internal/util"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidUser        = errors.New("invalid user")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrClientNotFound     = errors.New("client not found")
)

type Store struct {
	Logger *telemetry.Logger
	DB     *database.Database
}

func NewStore(logger *telemetry.Logger, db *database.Database) Store {
	return Store{
		Logger: logger,
		DB:     db,
	}
}

func (s *Store) NewUser(email string, password string, role UserRole) (User, error) {
	id := uuid.New()

	passwordHash, err := HashPassword(password)
	if err != nil {
		return User{}, err
	}

	user := User{
		ID:              id,
		Email:           email,
		PasswordHash:    passwordHash,
		Role:            role,
		IsEmailVerified: false,
		IsNew:           true,
	}

	return user, nil
}

func (s *Store) GetUser(ctx context.Context, userID uuid.UUID) (User, error) {
	var user User
	if err := s.DB.Conn.QueryRow(ctx, "SELECT id, email, password_hash, role, is_email_verified, created_at FROM tbl_user WHERE id=$1", userID).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.IsEmailVerified, &user.CreatedAt); err != nil {
		if err == database.ErrNoRows {
			return User{}, ErrUserNotFound
		}

		s.Logger.Warn("Failed to scan user", "user_id", userID)
		return User{}, err
	}

	return user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	if err := s.DB.Conn.QueryRow(ctx, "SELECT id, email, password_hash, role, is_email_verified, created_at FROM tbl_user WHERE email=$1", email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.IsEmailVerified, &user.CreatedAt); err != nil {
		if err == database.ErrNoRows {
			return User{}, ErrUserNotFound
		}

		s.Logger.Warn("Failed to scan user", "email", email)
		return User{}, err
	}

	return user, nil
}

func (s *Store) SaveUser(ctx context.Context, user User) (User, error) {
	// Validate user fields
	if user.Email == "" {
		return User{}, ErrInvalidUser
	}

	if _, err := mail.ParseAddress(user.Email); err != nil {
		return User{}, ErrInvalidUser
	}

	if !user.Role.IsValid() {
		return User{}, ErrInvalidUser
	}

	if user.IsNew {
		// Insert new user
		tx, err := s.DB.Conn.Begin(ctx)
		if err != nil {
			return User{}, err
		}
		defer tx.Rollback(ctx)

		createdAt := time.Now()
		if _, err := tx.Exec(ctx, "INSERT INTO tbl_account (id, type, created_at) VALUES ($1, $2, $3)", user.ID, AccountTypeUser, createdAt); err != nil {
			return User{}, err
		}

		if _, err := tx.Exec(ctx, "INSERT INTO tbl_user (id, email, password_hash, role, is_email_verified, created_at) VALUES ($1, $2, $3, $4, $5, $6)", user.ID, user.Email, user.PasswordHash, user.Role, user.IsEmailVerified, createdAt); err != nil {
			return User{}, err
		}

		if err := tx.Commit(ctx); err != nil {
			return User{}, err
		}

		user.IsNew = false
		user.CreatedAt = createdAt
		return user, nil
	}

	// Update existing user
	return User{}, errors.New("updating existing users not implemented")
}

func (s *Store) AdminExists(ctx context.Context) (bool, error) {
	var exists bool
	if err := s.DB.Conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM tbl_user WHERE role=$1)", RoleAdmin).Scan(&exists); err != nil {
		s.Logger.Warn("Failed to check for existing admin user")
		return false, err
	}
	return exists, nil
}

func (s *Store) ValidateCredentials(ctx context.Context, email string, password string) error {
	var passwordHash string
	if err := s.DB.Conn.QueryRow(ctx, "SELECT password_hash FROM tbl_user WHERE email=$1", email).Scan(&passwordHash); err != nil {
		if err == database.ErrNoRows {
			return ErrUserNotFound
		}

		s.Logger.Warn("Failed to scan password hash", "email", email)
		return err
	}

	if passwordHash == "" {
		s.Logger.Warn("Empty password hash for user", "email", email)
		return ErrInvalidCredentials
	}

	if err := ValidatePasswordWithHash(password, passwordHash); err != nil {
		s.Logger.Warn("Failed to compare password hashes", "email", email)
		return ErrInvalidCredentials
	}

	return nil
}

func (s *Store) NewClient(name string, redirectURIs []string) (Client, error) {
	secret, err := util.GenerateRandomString(32)
	if err != nil {
		return Client{}, err
	}

	return Client{
		ID:           uuid.New(),
		Secret:       secret,
		Name:         name,
		RedirectURIs: redirectURIs,
		IsNew:        true,
	}, nil
}

func (s *Store) GetClient(ctx context.Context, clientID uuid.UUID) (Client, error) {
	var client Client
	if err := s.DB.Conn.QueryRow(ctx, "SELECT id, secret, name, redirect_uris FROM tbl_client WHERE id=$1", clientID).Scan(&client.ID, &client.Secret, &client.Name, &client.RedirectURIs); err != nil {
		if errors.Is(err, database.ErrNoRows) {
			return Client{}, ErrClientNotFound
		}
		return Client{}, err
	}

	return client, nil
}

func (s *Store) SaveClient(ctx context.Context, client Client) (Client, error) {
	// Validate client fields
	if client.Name == "" {
		return Client{}, errors.New("client name cannot be empty")
	}
	if len(client.RedirectURIs) == 0 {
		return Client{}, errors.New("client must have at least one redirect URI")
	}

	if client.IsNew {
		// Insert new client
		tx, err := s.DB.Conn.Begin(ctx)
		if err != nil {
			return Client{}, err
		}
		defer tx.Rollback(ctx)

		if _, err := tx.Exec(ctx, "INSERT INTO tbl_account (id, type, created_at) VALUES ($1, $2, NOW())", client.ID, AccountTypeClient); err != nil {
			return Client{}, err
		}

		if _, err := tx.Exec(ctx, "INSERT INTO tbl_client (id, secret, name, redirect_uris, is_public, created_at) VALUES ($1, $2, $3, $4, $5, NOW())",
			client.ID, client.Secret, client.Name, client.RedirectURIs, client.IsPublic); err != nil {
			return Client{}, err
		}

		if err := tx.Commit(ctx); err != nil {
			return Client{}, err
		}

		client.IsNew = false
		return client, nil
	}

	// Update existing client
	return Client{}, errors.New("updating existing clients not implemented")
}

func (s *Store) DeleteClient(ctx context.Context, clientID uuid.UUID) error {
	// Start a transaction to delete from both client and account tables
	tx, err := s.DB.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Delete from tbl_client first (due to foreign key constraint)
	result, err := tx.Exec(ctx, "DELETE FROM tbl_client WHERE id = $1", clientID)
	if err != nil {
		s.Logger.Warn("Failed to delete client", "client_id", clientID, "error", err)
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrClientNotFound
	}

	// Delete from tbl_account
	_, err = tx.Exec(ctx, "DELETE FROM tbl_account WHERE id = $1", clientID)
	if err != nil {
		s.Logger.Warn("Failed to delete account for client", "client_id", clientID, "error", err)
		return err
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	s.Logger.Info("Client deleted successfully", "client_id", clientID)
	return nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func ValidatePasswordWithHash(password string, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
