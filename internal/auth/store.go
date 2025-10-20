package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/mail"

	"github.com/freekieb7/smauth/internal/database"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidUser        = errors.New("invalid user")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Store struct {
	Logger *slog.Logger
	DB     *database.Database
}

func NewStore(logger *slog.Logger, db *database.Database) Store {
	return Store{
		Logger: logger,
		DB:     db,
	}
}

func (s *Store) NewUser(email string, password string, role Role) (User, error) {
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

func (s *Store) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := s.DB.Conn.Query(ctx, "SELECT id, email, role, is_email_verified FROM tbl_user ORDER BY created_at DESC")
	if err != nil {
		s.Logger.Warn("Failed to query users")
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Email, &user.Role, &user.IsEmailVerified); err != nil {
			s.Logger.Warn("Failed to scan user", "error", err)
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		s.Logger.Warn("Error iterating over user rows", "error", err)
		return nil, err
	}

	return users, nil
}

func (s *Store) GetUser(ctx context.Context, userID uuid.UUID) (User, error) {
	var user User
	if err := s.DB.Conn.QueryRow(ctx, "SELECT email, password_hash, role, is_email_verified FROM tbl_user WHERE id=$1", userID).Scan(&user.Email, &user.PasswordHash, &user.Role, &user.IsEmailVerified); err != nil {
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
	if err := s.DB.Conn.QueryRow(ctx, "SELECT id, password_hash, role, is_email_verified FROM tbl_user WHERE email=$1", email).Scan(&user.ID, &user.PasswordHash, &user.Role, &user.IsEmailVerified); err != nil {
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
	if user.PasswordHash == "" {
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
		if _, err := s.DB.Conn.Exec(ctx, "INSERT INTO tbl_user (id, email, password_hash, role, is_email_verified, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())", user.ID, user.Email, user.PasswordHash, user.Role, user.IsEmailVerified); err != nil {
			return User{}, err
		}
		user.IsNew = false
		return user, nil
	}

	// Update existing user
	if _, err := s.DB.Conn.Exec(ctx, "UPDATE tbl_user SET email=$1, password_hash=$2, role=$3, updated_at=NOW() WHERE id=$4", user.Email, user.PasswordHash, user.Role, user.ID); err != nil {
		return User{}, err
	}
	return user, nil
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
	var user User
	if err := s.DB.Conn.QueryRow(ctx, "SELECT id, password_hash, is_email_verified FROM tbl_user WHERE email=$1", email).Scan(&user.ID, &user.PasswordHash, &user.IsEmailVerified); err != nil {
		if err == database.ErrNoRows {
			return ErrUserNotFound
		}

		s.Logger.Warn("Failed to scan password hash", "email", email)
		return err
	}

	if err := ValidatePasswordHash(password, user.PasswordHash); err != nil {
		s.Logger.Warn("Failed to compare password hashes", "email", email)
		return ErrInvalidCredentials
	}

	return nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func ValidatePasswordHash(password string, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
