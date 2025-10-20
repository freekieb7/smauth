package oauth

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/freekieb7/smauth/internal"
	"github.com/freekieb7/smauth/internal/database"
	"github.com/freekieb7/smauth/internal/telemetry"
	"github.com/freekieb7/smauth/internal/util"
	"github.com/google/uuid"
)

var (
	ErrResourceServerNotFound      = errors.New("resource server not found")
	ErrInvalidPermissionAssignment = errors.New("invalid permission assignment")
	ErrAccessTokenNotFound         = errors.New("access token not found")
	ErrAuthorizationCodeNotFound   = errors.New("authorization code not found")
	ErrRefreshTokenChainNotFound   = errors.New("refresh token chain not found")
	ErrRefreshTokenNotFound        = errors.New("refresh token not found")
)

type Store struct {
	Cfg    internal.OAuthConfig
	Logger *telemetry.Logger
	DB     *database.Database
}

func NewStore(cfg internal.OAuthConfig, logger *telemetry.Logger, db *database.Database) Store {
	return Store{
		Cfg:    cfg,
		Logger: logger,
		DB:     db,
	}
}

func (s *Store) NewResourceServer(url string) (ResourceServer, error) {
	return ResourceServer{
		ID:        uuid.New(),
		URL:       url,
		CreatedAt: time.Now(),
		IsNew:     true,
	}, nil
}

func (s *Store) NewScope(url, name, description string) (Scope, error) {
	return Scope{
		ID:          uuid.New(),
		Name:        url + "/" + name,
		Description: description,
	}, nil
}

func (s *Store) GetResourceServerByID(ctx context.Context, id uuid.UUID) (ResourceServer, error) {
	var resourceServer ResourceServer
	if err := s.DB.Conn.QueryRow(ctx, "SELECT id, url, created_at FROM tbl_resource_server WHERE id=$1", id).Scan(&resourceServer.ID, &resourceServer.URL, &resourceServer.CreatedAt); err != nil {
		if errors.Is(err, database.ErrNoRows) {
			return ResourceServer{}, ErrResourceServerNotFound
		}
		return ResourceServer{}, err
	}

	// Get scopes
	rows, err := s.DB.Conn.Query(ctx, "SELECT id, name, description FROM tbl_scope WHERE resource_server_id=$1", resourceServer.ID)
	if err != nil {
		return ResourceServer{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var scope Scope
		if err := rows.Scan(&scope.ID, &scope.Name, &scope.Description); err != nil {
			return ResourceServer{}, err
		}
		resourceServer.Scopes = append(resourceServer.Scopes, scope)
	}
	if err := rows.Err(); err != nil {
		return ResourceServer{}, err
	}

	return resourceServer, nil
}

func (s *Store) GetResourceServerByURL(ctx context.Context, url string) (ResourceServer, error) {
	var resourceServer ResourceServer
	if err := s.DB.Conn.QueryRow(ctx, "SELECT id, url, created_at FROM tbl_resource_server WHERE url=$1", url).Scan(&resourceServer.ID, &resourceServer.URL, &resourceServer.CreatedAt); err != nil {
		if errors.Is(err, database.ErrNoRows) {
			return ResourceServer{}, ErrResourceServerNotFound
		}
		return ResourceServer{}, err
	}

	// Get scopes
	rows, err := s.DB.Conn.Query(ctx, "SELECT id, name, description FROM tbl_scope WHERE resource_server_id=$1", resourceServer.ID)
	if err != nil {
		return ResourceServer{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var scope Scope
		if err := rows.Scan(&scope.ID, &scope.Name, &scope.Description); err != nil {
			return ResourceServer{}, err
		}
		resourceServer.Scopes = append(resourceServer.Scopes, scope)
	}
	if err := rows.Err(); err != nil {
		return ResourceServer{}, err
	}

	return resourceServer, nil
}

// When updating, only scopes can be updated
func (s *Store) SaveResourceServer(ctx context.Context, resourceServer ResourceServer) (ResourceServer, error) {
	for _, scope := range resourceServer.Scopes {
		if !strings.HasPrefix(scope.Name, resourceServer.URL) {
			return ResourceServer{}, fmt.Errorf("scope %s does not belong to resource server %s", scope.Name, resourceServer.URL)
		}
	}

	if resourceServer.IsNew {
		// Insert new resource server
		tx, err := s.DB.Conn.Begin(ctx)
		if err != nil {
			return ResourceServer{}, err
		}
		defer tx.Rollback(ctx)
		if _, err := tx.Exec(ctx, "INSERT INTO tbl_resource_server (id, url, created_at) VALUES ($1, $2, $3)",
			resourceServer.ID, resourceServer.URL, resourceServer.CreatedAt.UTC()); err != nil {
			return ResourceServer{}, err
		}

		if len(resourceServer.Scopes) > 0 {
			var query strings.Builder
			args := []any{}
			argNum := 1

			query.WriteString("INSERT INTO tbl_scope (id, resource_server_id, name, description, created_at) VALUES ")
			for i, scope := range resourceServer.Scopes {
				if i > 0 {
					query.WriteString(", ")
				}
				query.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d, NOW())", argNum, argNum+1, argNum+2, argNum+3))
				args = append(args, scope.ID, resourceServer.ID, scope.Name, scope.Description)
				argNum += 4
			}

			if _, err := tx.Exec(ctx, query.String(), args...); err != nil {
				return ResourceServer{}, err
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return ResourceServer{}, err
		}

		resourceServer.IsNew = false
		return resourceServer, nil
	}

	// Update existing resource server
	return ResourceServer{}, errors.New("update resource server is not supported")
}

func (s *Store) DeleteResourceServer(ctx context.Context, resourceServerID uuid.UUID) error {
	_, err := s.DB.Conn.Exec(ctx, "DELETE FROM tbl_resource_server WHERE id = $1", resourceServerID)
	if err != nil {
		s.Logger.Warn("Failed to delete resource server", "resource_server_id", resourceServerID, "error", err)
		return err
	}

	s.Logger.Info("Resource Server deleted successfully", "resource_server_id", resourceServerID)
	return nil
}

func (s *Store) NewAccessToken(accountID uuid.UUID, scopes []string) (AccessToken, error) {
	// Validate input parameters
	if accountID == uuid.Nil {
		return AccessToken{}, errors.New("account ID is required")
	}

	// Generate token
	token, err := util.GenerateRandomString(32)
	if err != nil {
		return AccessToken{}, err
	}

	return AccessToken{
		Token:     token,
		AccountID: accountID,
		Scopes:    scopes,
		ExpiresAt: time.Now().Add(s.Cfg.TokenExpiration),
		IsNew:     true,
	}, nil
}

func (s *Store) SaveAccessToken(ctx context.Context, accessToken AccessToken) (AccessToken, error) {
	if !accessToken.IsNew {
		return AccessToken{}, errors.New("updating access tokens is not supported")
	}

	// Validate access token fields
	if accessToken.Token == "" {
		return AccessToken{}, errors.New("access token is required")
	}
	if accessToken.AccountID == uuid.Nil {
		return AccessToken{}, errors.New("account ID is required")
	}
	if accessToken.Scopes == nil {
		accessToken.Scopes = []string{}
	}
	if accessToken.ExpiresAt.IsZero() {
		return AccessToken{}, errors.New("expiration time is required")
	}

	if _, err := s.DB.Conn.Exec(ctx, "INSERT INTO tbl_access_token (token, account_id, scopes, expires_at, created_at) VALUES ($1, $2, $3, $4, NOW())",
		accessToken.Token, accessToken.AccountID, accessToken.Scopes, accessToken.ExpiresAt.UTC()); err != nil {
		return AccessToken{}, err
	}

	accessToken.IsNew = false
	return accessToken, nil
}

func (s *Store) GetAccessToken(ctx context.Context, token string) (AccessToken, error) {
	var accessToken AccessToken
	if err := s.DB.Conn.QueryRow(ctx, "SELECT token, account_id, scopes, expires_at FROM tbl_access_token WHERE token=$1 AND expires_at > NOW()", token).Scan(&accessToken.Token, &accessToken.AccountID, &accessToken.Scopes, &accessToken.ExpiresAt); err != nil {
		if errors.Is(err, database.ErrNoRows) {
			return AccessToken{}, ErrAccessTokenNotFound
		}
		return AccessToken{}, err
	}

	return accessToken, nil
}

func (s *Store) NewAuthorizationCode(userID, clientID uuid.UUID, redirectURI string, scopes []string, codeChallenge string, codeMethod string) (AuthorizationCode, error) {
	code, err := util.GenerateRandomString(32)
	if err != nil {
		return AuthorizationCode{}, err
	}

	return AuthorizationCode{
		Code:          code,
		ClientID:      clientID,
		UserID:        userID,
		RedirectURI:   redirectURI,
		Scopes:        scopes,
		CodeChallenge: codeChallenge,
		CodeMethod:    codeMethod,
		IsNew:         true,
	}, nil
}

func (s *Store) SaveAuthorizationCode(ctx context.Context, code AuthorizationCode) (AuthorizationCode, error) {
	if code.IsNew {
		if _, err := s.DB.Conn.Exec(ctx, "INSERT INTO tbl_authorization_code (code, client_id, user_id, redirect_uri, scopes, code_challenge, code_method, expires_at, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())",
			code.Code, code.ClientID, code.UserID, code.RedirectURI, code.Scopes, code.CodeChallenge, code.CodeMethod, time.Now().Add(10*time.Minute).UTC()); err != nil {
			return AuthorizationCode{}, err
		}

		return code, nil
	}

	return AuthorizationCode{}, errors.New("updating authorization codes is not supported")
}

func (s *Store) MarkAuthorizationCodeAsUsed(ctx context.Context, codeStr string) error {
	if _, err := s.DB.Conn.Exec(ctx, "UPDATE tbl_authorization_code SET used_at = NOW() WHERE code = $1", codeStr); err != nil {
		return err
	}
	return nil
}

func (s *Store) GetAuthorizationCode(ctx context.Context, codeStr string) (AuthorizationCode, error) {
	var code AuthorizationCode
	if err := s.DB.Conn.QueryRow(ctx, "SELECT code, client_id, user_id, scopes, redirect_uri, code_challenge, code_method FROM tbl_authorization_code WHERE code=$1 AND expires_at > NOW() AND used_at IS NULL", codeStr).
		Scan(&code.Code, &code.ClientID, &code.UserID, &code.Scopes, &code.RedirectURI, &code.CodeChallenge, &code.CodeMethod); err != nil {
		if errors.Is(err, database.ErrNoRows) {
			return AuthorizationCode{}, ErrAuthorizationCodeNotFound
		}
		return AuthorizationCode{}, err
	}

	return code, nil
}

func (s *Store) NewRefreshTokenChain(clientID, userID uuid.UUID, scopes []string) (RefreshTokenChain, error) {
	return RefreshTokenChain{
		ID:       uuid.New(),
		ClientID: clientID,
		UserID:   userID,
		Scopes:   scopes,
		IsNew:    true,
	}, nil
}

func (s *Store) SaveRefreshTokenChain(ctx context.Context, chain RefreshTokenChain) (RefreshTokenChain, error) {
	if chain.IsNew {
		if _, err := s.DB.Conn.Exec(ctx, "INSERT INTO tbl_refresh_token_chain (id, client_id, user_id, scopes, created_at) VALUES ($1, $2, $3, $4, NOW())",
			chain.ID, chain.ClientID, chain.UserID, chain.Scopes); err != nil {
			return RefreshTokenChain{}, err
		}

		chain.IsNew = false
		return chain, nil
	}

	return RefreshTokenChain{}, errors.New("updating refresh token chains is not supported")
}

func (s *Store) GetRefreshTokenChain(ctx context.Context, chainID uuid.UUID) (RefreshTokenChain, error) {
	var chain RefreshTokenChain
	if err := s.DB.Conn.QueryRow(ctx, `SELECT id, client_id, user_id, scopes FROM tbl_refresh_token_chain WHERE id = $1`, chainID).
		Scan(&chain.ID, &chain.ClientID, &chain.UserID, &chain.Scopes); err != nil {
		if errors.Is(err, database.ErrNoRows) {
			return RefreshTokenChain{}, ErrRefreshTokenChainNotFound
		}
		return RefreshTokenChain{}, err
	}

	return chain, nil
}

func (s *Store) NewRefreshToken() (RefreshToken, error) {
	token, err := util.GenerateRandomString(32)
	if err != nil {
		return RefreshToken{}, err
	}

	return RefreshToken{
		Token: token,
		IsNew: true,
	}, nil
}

func (s *Store) SaveRefreshToken(ctx context.Context, chainID uuid.UUID, token RefreshToken) (RefreshToken, error) {
	if token.IsNew {
		if _, err := s.DB.Conn.Exec(ctx, "INSERT INTO tbl_refresh_token (token, chain_id, expires_at, used_at, created_at) VALUES ($1, $2, $3, NULL, NOW())",
			token.Token, chainID, time.Now().Add(24*time.Hour).UTC()); err != nil {
			return RefreshToken{}, err
		}

		token.IsNew = false
		return token, nil
	}

	return RefreshToken{}, errors.New("updating refresh tokens is not supported")
}

func (s *Store) GetRefreshTokenChainByToken(ctx context.Context, tokenStr string) (RefreshTokenChain, error) {
	var chain RefreshTokenChain
	if err := s.DB.Conn.QueryRow(ctx, `SELECT chain.id, chain.client_id, chain.user_id, chain.scopes
		FROM tbl_refresh_token token
		JOIN tbl_refresh_token_chain chain ON token.chain_id = chain.id
		WHERE token.token = $1 AND token.expires_at > NOW() AND token.used_at IS NULL`, tokenStr).
		Scan(&chain.ID, &chain.ClientID, &chain.UserID, &chain.Scopes); err != nil {
		if errors.Is(err, database.ErrNoRows) {
			return RefreshTokenChain{}, ErrRefreshTokenChainNotFound
		}
		return RefreshTokenChain{}, err
	}

	return chain, nil
}

func (s *Store) MarkRefreshTokenAsUsed(ctx context.Context, tokenStr string) error {
	if _, err := s.DB.Conn.Exec(ctx, "UPDATE tbl_refresh_token SET used_at = NOW() WHERE token = $1", tokenStr); err != nil {
		return err
	}
	return nil
}

func (s *Store) AssignPermissions(ctx context.Context, accountID uuid.UUID, scopes []string) error {
	if len(scopes) == 0 {
		return nil
	}

	var scopeIDs []uuid.UUID
	if err := s.DB.Conn.QueryRow(ctx, "SELECT array_agg(id) FROM tbl_scope WHERE name = any($1::text[])", scopes).Scan(&scopeIDs); err != nil {
		s.Logger.Error("failed to get scope IDs", "error", err)
		return err
	}

	if len(scopeIDs) == 0 {
		s.Logger.Error("no valid scopes found to assign", "account_id", accountID, "scopes", scopes)
		return ErrInvalidPermissionAssignment
	}

	var query strings.Builder
	args := []any{}
	argNum := 1

	query.WriteString("INSERT INTO tbl_account_permission (id, account_id, scope_id, created_at) VALUES ")
	for i, scopeID := range scopeIDs {
		if i > 0 {
			query.WriteString(", ")
		}
		query.WriteString(fmt.Sprintf("($%d, $%d, $%d, NOW())", argNum, argNum+1, argNum+2))
		args = append(args, uuid.New(), accountID, scopeID)
		argNum += 3
	}
	query.WriteString(" ON CONFLICT (account_id, scope_id) DO NOTHING")

	if _, err := s.DB.Conn.Exec(ctx, query.String(), args...); err != nil {
		s.Logger.Error("failed to assign permissions", "account_id", accountID, "scopes", scopes, "error", err)
		return err
	}

	return nil
}

// Removes unknown scopes from the requested scopes
func (s *Store) FilterGrantedFromUngrantedScopes(ctx context.Context, userID, clientID uuid.UUID, scopes []string) ([]Scope, []Scope, error) {
	// Process each resource server's scopes
	var knownScopes []Scope
	rows, err := s.DB.Conn.Query(ctx, `SELECT id, name, description FROM tbl_scope WHERE name = any($1::text[])`, scopes)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var scope Scope
		if err := rows.Scan(&scope.ID, &scope.Name, &scope.Description); err != nil {
			return nil, nil, err
		}
		knownScopes = append(knownScopes, scope)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	var grantedScopes []Scope
	var ungrantedScopes []Scope

	var knownScopeIds []uuid.UUID
	for _, scope := range knownScopes {
		knownScopeIds = append(knownScopeIds, scope.ID)
	}

	// Check which scopes are granted
	var grantedScopeIds []uuid.UUID
	if err := s.DB.Conn.QueryRow(ctx, `
		SELECT array_agg(scope_id)
		FROM tbl_grant
		WHERE user_id = $1
		AND client_id = $2
		AND scope_id = any($3::uuid[]);
	`, userID, clientID, knownScopeIds).Scan(&grantedScopeIds); err != nil {
		return nil, nil, err
	}

	// Determine ungranted scopes
	for _, scope := range knownScopes {
		if slices.Contains(grantedScopeIds, scope.ID) {
			grantedScopes = append(grantedScopes, scope)
		} else {
			ungrantedScopes = append(ungrantedScopes, scope)
		}
	}

	return grantedScopes, ungrantedScopes, nil
}

func (s *Store) Grant(ctx context.Context, userID, clientID uuid.UUID, scopes []Scope) error {
	var query strings.Builder
	args := []any{}
	argNum := 1

	query.WriteString("INSERT INTO tbl_grant (id, user_id, client_id, scope_id, created_at) VALUES ")
	for i, scope := range scopes {
		if i > 0 {
			query.WriteString(", ")
		}
		query.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d, NOW())", argNum, argNum+1, argNum+2, argNum+3))
		args = append(args, uuid.New(), userID, clientID, scope.ID)
		argNum += 4
	}

	query.WriteString(" ON CONFLICT (user_id, client_id, scope_id) DO NOTHING")
	if _, err := s.DB.Conn.Exec(ctx, query.String(), args...); err != nil {
		s.Logger.Error("failed to grant scopes to user", "user_id", userID, "client_id", clientID, "scopes", scopes, "error", err)
		return err
	}

	return nil
}

func (s *Store) CreateScope(ctx context.Context, resourceServer ResourceServer, scope Scope) error {
	if !strings.HasPrefix(scope.Name, resourceServer.URL) {
		return fmt.Errorf("scope %s does not belong to resource server %s", scope.Name, resourceServer.URL)
	}

	query := `
		INSERT INTO tbl_scope (id, resource_server_id, name, description, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`

	_, err := s.DB.Conn.Exec(ctx, query, scope.ID, resourceServer.ID, scope.Name, scope.Description)
	if err != nil {
		return fmt.Errorf("failed to create scope: %w", err)
	}

	return nil
}

func (s *Store) UpdateScope(ctx context.Context, resourceServerID uuid.UUID, originalScopeName string, scope Scope) error {
	query := `
		UPDATE tbl_scope 
		SET name = $1, description = $2
		WHERE resource_server_id = $3 AND name = $4
	`

	result, err := s.DB.Conn.Exec(ctx, query, scope.Name, scope.Description, resourceServerID, originalScopeName)
	if err != nil {
		return fmt.Errorf("failed to update scope: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("scope not found")
	}

	return nil
}

func (s *Store) DeleteScope(ctx context.Context, resourceServerID uuid.UUID, scopeName string) error {
	query := `
		DELETE FROM tbl_scope 
		WHERE resource_server_id = $1 AND name = $2
	`

	result, err := s.DB.Conn.Exec(ctx, query, resourceServerID, scopeName)
	if err != nil {
		return fmt.Errorf("failed to delete scope: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("scope not found")
	}

	return nil
}
