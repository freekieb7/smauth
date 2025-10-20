package oauth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/freekieb7/smauth/internal"
	"github.com/freekieb7/smauth/internal/database"
	"github.com/freekieb7/smauth/internal/util"
	"github.com/google/uuid"
)

var (
	ErrClientNotFound              = errors.New("oauth client not found")
	ErrResourceServerNotFound      = errors.New("resource server not found")
	ErrInvalidPermissionAssignment = errors.New("invalid permission assignment")
	ErrAccessTokenNotFound         = errors.New("access token not found")
	ErrAuthorizationCodeNotFound   = errors.New("authorization code not found")
	ErrRefreshTokenChainNotFound   = errors.New("refresh token chain not found")
	ErrRefreshTokenNotFound        = errors.New("refresh token not found")
)

type Store struct {
	Cfg    internal.OAuthConfig
	Logger *slog.Logger
	DB     *database.Database
}

func NewStore(cfg internal.OAuthConfig, logger *slog.Logger, db *database.Database) Store {
	return Store{
		Cfg:    cfg,
		Logger: logger,
		DB:     db,
	}
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
	if client.IsNew {
		if _, err := s.DB.Conn.Exec(ctx, "INSERT INTO tbl_client (id, secret, name, redirect_uris, is_public, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())",
			client.ID, client.Secret, client.Name, client.RedirectURIs, client.IsPublic); err != nil {
			return Client{}, err
		}

		client.IsNew = false
		return client, nil
	}

	// Update existing client
	if _, err := s.DB.Conn.Exec(ctx, "UPDATE tbl_client SET secret = $1, name = $2, redirect_uris = $3, is_public = $4, updated_at = NOW() WHERE id = $5",
		client.Secret, client.Name, client.RedirectURIs, client.IsPublic, client.ID); err != nil {
		return Client{}, err
	}

	return client, nil
}

func (s *Store) NewResourceServer(url string) (ResourceServer, error) {
	return ResourceServer{
		ID:    uuid.New(),
		URL:   url,
		IsNew: true,
	}, nil
}

func (s *Store) NewScope(url, name, description string) (Scope, error) {
	return Scope{
		ID:          uuid.New(),
		Name:        url + "/" + name,
		Description: description,
	}, nil
}

func (s *Store) GetResourceServerByURL(ctx context.Context, url string) (ResourceServer, error) {
	var resourceServer ResourceServer
	if err := s.DB.Conn.QueryRow(ctx, "SELECT id, url FROM tbl_resource_server WHERE url=$1", url).Scan(&resourceServer.ID, &resourceServer.URL); err != nil {
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
		if _, err := tx.Exec(ctx, "INSERT INTO tbl_resource_server (id, url, created_at, updated_at) VALUES ($1, $2, NOW(), NOW())",
			resourceServer.ID, resourceServer.URL); err != nil {
			return ResourceServer{}, err
		}

		var query strings.Builder
		args := []any{}
		argNum := 1

		query.WriteString("INSERT INTO tbl_scope (id, resource_server_id, name, description, created_at, updated_at) VALUES ")
		for i, scope := range resourceServer.Scopes {
			if i > 0 {
				query.WriteString(", ")
			}
			query.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d, NOW(), NOW())", argNum, argNum+1, argNum+2, argNum+3))
			args = append(args, scope.ID, resourceServer.ID, scope.Name, scope.Description)
			argNum += 4
		}

		if _, err := tx.Exec(ctx, query.String(), args...); err != nil {
			return ResourceServer{}, err
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

func (s *Store) NewAccessToken(userID, clientID util.Optional[uuid.UUID], scopes []string) (AccessToken, error) {
	token, err := util.GenerateRandomString(32)
	if err != nil {
		return AccessToken{}, err
	}

	if !clientID.IsSet() && !userID.IsSet() {
		return AccessToken{}, errors.New("either clientID or userID must be set")
	}

	return AccessToken{
		Token:     token,
		ClientID:  clientID,
		UserID:    userID,
		Scopes:    scopes,
		ExpiresAt: time.Now().Add(s.Cfg.TokenExpiration).UTC(),
		IsNew:     true,
	}, nil
}

func (s *Store) SaveAccessToken(ctx context.Context, accessToken AccessToken) (AccessToken, error) {
	if !accessToken.IsNew {
		return AccessToken{}, errors.New("updating access tokens is not supported")
	}

	if _, err := s.DB.Conn.Exec(ctx, "INSERT INTO tbl_access_token (token, client_id, user_id, scopes, expires_at, created_at) VALUES ($1, $2, $3, $4, $5, NOW())",
		accessToken.Token, accessToken.ClientID, accessToken.UserID, accessToken.Scopes, time.Now().Add(s.Cfg.TokenExpiration).UTC()); err != nil {
		return AccessToken{}, err
	}

	accessToken.IsNew = false
	return accessToken, nil
}

func (s *Store) GetAccessToken(ctx context.Context, token string) (AccessToken, error) {
	var accessToken AccessToken
	if err := s.DB.Conn.QueryRow(ctx, "SELECT token, client_id, user_id, scopes, expires_at FROM tbl_access_token WHERE token=$1 AND expires_at > NOW()", token).Scan(&accessToken.Token, &accessToken.ClientID, &accessToken.UserID, &accessToken.Scopes, &accessToken.ExpiresAt); err != nil {
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

func (s *Store) AssignPermissions(ctx context.Context, clientID, userID util.Optional[uuid.UUID], scopes []string) error {
	if len(scopes) == 0 {
		return nil
	}

	var scopeIDs []uuid.UUID
	if err := s.DB.Conn.QueryRow(ctx, "SELECT array_agg(id) FROM tbl_scope WHERE name = any($1::text[])", scopes).Scan(&scopeIDs); err != nil {
		s.Logger.Error("failed to get scope IDs", "error", err)
		return err
	}

	var query strings.Builder
	args := []any{}
	argNum := 1

	if userID.IsSet() {
		query.WriteString("INSERT INTO tbl_user_permission (id, user_id, scope_id, created_at) VALUES ")
	} else {
		query.WriteString("INSERT INTO tbl_client_permission (id, client_id, scope_id, created_at) VALUES ")
	}
	for i, scopeID := range scopeIDs {
		if i > 0 {
			query.WriteString(", ")
		}
		query.WriteString(fmt.Sprintf("($%d, $%d, $%d, NOW())", argNum, argNum+1, argNum+2))
		if userID.IsSet() {
			args = append(args, uuid.New(), userID.V, scopeID)
		} else {
			args = append(args, uuid.New(), clientID.V, scopeID)
		}
		argNum += 3
	}

	if userID.IsSet() {
		query.WriteString(" ON CONFLICT (user_id, scope_id) DO NOTHING")
	} else {
		query.WriteString(" ON CONFLICT (client_id, scope_id) DO NOTHING")
	}

	if _, err := s.DB.Conn.Exec(ctx, query.String(), args...); err != nil {
		s.Logger.Error("failed to assign permissions", "client_id", clientID, "user_id", userID, "scopes", scopes, "error", err)
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
