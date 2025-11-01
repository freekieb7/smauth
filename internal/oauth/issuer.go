package oauth

import (
	"context"
	"errors"
	"log/slog"
	"slices"

	"github.com/freekieb7/smauth/internal/database"
	"github.com/freekieb7/smauth/internal/util"
	"github.com/google/uuid"
)

var (
	ErrInvalidClientCredentials = errors.New("invalid client credentials")
	ErrUnsupportedGrantType     = errors.New("unsupported grant type")
	ErrUnsupportedResponseType  = errors.New("unsupported response type")
	ErrInvalidRedirectURI       = errors.New("invalid redirect_uri")
	ErrRequestedUngrantedScopes = errors.New("requested scopes include ungranted scopes")
	ErrInvalidRefreshToken      = errors.New("invalid refresh token")
)

type Issuer struct {
	Logger *slog.Logger
	DB     *database.Database
	Store  *Store
}

func NewIssuer(logger *slog.Logger, db *database.Database, store *Store) Issuer {
	return Issuer{
		Logger: logger,
		DB:     db,
		Store:  store,
	}
}

func (i *Issuer) IssueCode(ctx context.Context, responseType string, clientID, userID uuid.UUID, redirectURI string, scopes []string, codeChallenge string, codeMethod string, allowGrant bool) (AuthorizationCode, []Scope, error) {
	switch responseType {
	case "code":
		// Valid response type
		client, err := i.Store.GetClient(ctx, clientID)
		if err != nil {
			if errors.Is(err, ErrClientNotFound) {
				return AuthorizationCode{}, nil, ErrInvalidClientCredentials
			}

			i.Logger.Error("failed to validate client credentials", "error", err)
			return AuthorizationCode{}, nil, err
		}

		if !slices.Contains(client.RedirectURIs, redirectURI) {
			return AuthorizationCode{}, nil, ErrInvalidRedirectURI
		}

		// Validate PKCE (must prove challenge)
		if codeChallenge == "" || codeMethod == "" {
			return AuthorizationCode{}, nil, errors.New("PKCE parameters are missing")
		}

		if codeMethod != "S256" {
			return AuthorizationCode{}, nil, errors.New("unsupported code challenge method")
		}

		// Validate redirect URI
		if redirectURI == "" {
			return AuthorizationCode{}, nil, errors.New("redirect_uri is required")
		}

		// Validate scopes
		grantedScopes, ungrantedScopes, err := i.Store.FilterGrantedFromUngrantedScopes(ctx, userID, clientID, scopes)
		if err != nil {
			i.Logger.Error("failed to filter granted from ungranted scopes", "error", err)
			return AuthorizationCode{}, nil, err
		}

		if len(ungrantedScopes) > 0 {
			if !allowGrant {
				return AuthorizationCode{}, ungrantedScopes, ErrRequestedUngrantedScopes
			}

			// Grant the requested scopes
			if err := i.Store.Grant(ctx, userID, clientID, ungrantedScopes); err != nil {
				i.Logger.Error("failed to grant scopes", "user_id", userID, "client_id", clientID, "scopes", ungrantedScopes, "error", err)
				return AuthorizationCode{}, nil, err
			}
			grantedScopes = append(grantedScopes, ungrantedScopes...)
			ungrantedScopes = nil
		}

		scopes := make([]string, 0, len(grantedScopes))
		for _, scope := range grantedScopes {
			scopes = append(scopes, scope.Name)
		}

		// Create authorization code
		code, err := i.Store.NewAuthorizationCode(userID, clientID, redirectURI, scopes, codeChallenge, codeMethod)
		if err != nil {
			i.Logger.Error("failed to create authorization code", "error", err)
			return AuthorizationCode{}, nil, err
		}

		// Save authorization code
		code, err = i.Store.SaveAuthorizationCode(ctx, code)
		if err != nil {
			i.Logger.Error("failed to save authorization code", "error", err)
			return AuthorizationCode{}, nil, err
		}

		return code, nil, nil
	default:
		return AuthorizationCode{}, nil, ErrUnsupportedResponseType
	}
}

type Tokens struct {
	AccessToken  AccessToken
	RefreshToken RefreshToken
}

func (i *Issuer) IssueTokens(ctx context.Context, grantType string, clientID uuid.UUID, clientSecret string, scopes []string, code, redirectURI, refreshToken string) (Tokens, error) {
	switch grantType {
	case "client_credentials":
		// Validate client credentials
		client, err := i.Store.GetClient(ctx, clientID)
		if err != nil {
			if errors.Is(err, ErrClientNotFound) {
				return Tokens{}, ErrInvalidClientCredentials
			}

			i.Logger.Error("failed to validate client credentials", "error", err)
			return Tokens{}, err
		}

		if client.Secret != clientSecret {
			return Tokens{}, ErrInvalidClientCredentials
		}

		// Validate scopes
		var grantedScopes []string

		if err := i.DB.Conn.QueryRow(ctx, `SELECT array_agg(scope.name) FROM tbl_client_permission permission
			JOIN tbl_scope scope ON permission.scope_id = scope.id
			WHERE permission.client_id = $1 AND scope.name IN ($2::text[])`, clientID, scopes).Scan(&grantedScopes); err != nil {
			if !errors.Is(err, database.ErrNoRows) {
				i.Logger.Error("failed to query granted scopes", "error", err)
				return Tokens{}, err
			}
		}

		accessToken, err := i.Store.NewAccessToken(util.None[uuid.UUID](), util.Some(clientID), grantedScopes)
		if err != nil {
			i.Logger.Error("failed to create access token", "error", err)
			return Tokens{}, err
		}

		// Save access token
		accessToken, err = i.Store.SaveAccessToken(ctx, accessToken)
		if err != nil {
			i.Logger.Error("failed to save access token", "error", err)
			return Tokens{}, err
		}

		return Tokens{AccessToken: accessToken}, nil
	case "authorization_code":
		{
			client, err := i.Store.GetClient(ctx, clientID)
			if err != nil {
				if errors.Is(err, ErrClientNotFound) {
					return Tokens{}, ErrInvalidClientCredentials
				}

				i.Logger.Error("failed to validate client credentials", "error", err)
				return Tokens{}, err
			}

			if !client.IsPublic && client.Secret != clientSecret {
				return Tokens{}, ErrInvalidClientCredentials
			}

			authCode, err := i.Store.GetAuthorizationCode(ctx, code)
			if err != nil {
				if errors.Is(err, ErrAuthorizationCodeNotFound) {
					return Tokens{}, errors.New("invalid authorization code")
				}

				i.Logger.Error("failed to retrieve authorization code", "error", err)
				return Tokens{}, err
			}

			if authCode.ClientID != clientID {
				return Tokens{}, ErrInvalidClientCredentials
			}

			if authCode.RedirectURI != redirectURI {
				return Tokens{}, ErrInvalidRedirectURI
			}

			// Mark authorization code as used
			if err := i.Store.MarkAuthorizationCodeAsUsed(ctx, authCode.Code); err != nil {
				i.Logger.Error("failed to mark authorization code as used", "error", err)
				return Tokens{}, err
			}

			accessToken, err := i.Store.NewAccessToken(util.Some(authCode.UserID), util.Some(clientID), authCode.Scopes)
			if err != nil {
				i.Logger.Error("failed to create access token", "error", err)
				return Tokens{}, err
			}

			// Save access token
			accessToken, err = i.Store.SaveAccessToken(ctx, accessToken)
			if err != nil {
				i.Logger.Error("failed to save access token", "error", err)
				return Tokens{}, err
			}

			var refreshToken RefreshToken

			if slices.Contains(authCode.Scopes, "offline_access") {
				refreshTokenChain, err := i.Store.NewRefreshTokenChain(clientID, authCode.UserID, authCode.Scopes)
				if err != nil {
					i.Logger.Error("failed to create refresh token chain", "error", err)
					return Tokens{}, err
				}

				// Save refresh token chain
				refreshTokenChain, err = i.Store.SaveRefreshTokenChain(ctx, refreshTokenChain)
				if err != nil {
					i.Logger.Error("failed to save refresh token chain", "error", err)
					return Tokens{}, err
				}

				refreshToken, err = i.Store.NewRefreshToken()
				if err != nil {
					i.Logger.Error("failed to create refresh token", "error", err)
					return Tokens{}, err
				}

				// Save refresh token
				refreshToken, err = i.Store.SaveRefreshToken(ctx, refreshTokenChain.ID, refreshToken)
				if err != nil {
					i.Logger.Error("failed to save refresh token", "error", err)
					return Tokens{}, err
				}
			}

			return Tokens{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			}, nil
		}
	case "refresh_token":
		{
			client, err := i.Store.GetClient(ctx, clientID)
			if err != nil {
				if errors.Is(err, ErrClientNotFound) {
					return Tokens{}, ErrInvalidClientCredentials
				}

				i.Logger.Error("failed to validate client credentials", "error", err)
				return Tokens{}, err
			}

			if !client.IsPublic && client.Secret != clientSecret {
				return Tokens{}, ErrInvalidClientCredentials
			}

			chain, err := i.Store.GetRefreshTokenChainByToken(ctx, refreshToken)
			if err != nil {
				if errors.Is(err, ErrRefreshTokenChainNotFound) {
					return Tokens{}, ErrInvalidRefreshToken
				}

				i.Logger.Error("failed to retrieve refresh token chain and token", "error", err)
				return Tokens{}, err
			}

			if chain.ClientID != clientID {
				return Tokens{}, ErrInvalidClientCredentials
			}

			accessToken, err := i.Store.NewAccessToken(util.Some(chain.UserID), util.Some(clientID), chain.Scopes)
			if err != nil {
				i.Logger.Error("failed to create access token", "error", err)
				return Tokens{}, err
			}

			// Save access token
			accessToken, err = i.Store.SaveAccessToken(ctx, accessToken)
			if err != nil {
				i.Logger.Error("failed to save access token", "error", err)
				return Tokens{}, err
			}

			newRefreshToken, err := i.Store.NewRefreshToken()
			if err != nil {
				i.Logger.Error("failed to create refresh token", "error", err)
				return Tokens{}, err
			}

			// Mark old refresh token as used
			if err := i.Store.MarkRefreshTokenAsUsed(ctx, refreshToken); err != nil {
				i.Logger.Error("failed to mark refresh token as used", "error", err)
				return Tokens{}, err
			}

			// Save refresh token
			newRefreshToken, err = i.Store.SaveRefreshToken(ctx, chain.ID, newRefreshToken)
			if err != nil {
				i.Logger.Error("failed to save refresh token", "error", err)
				return Tokens{}, err
			}

			return Tokens{
				AccessToken:  accessToken,
				RefreshToken: newRefreshToken,
			}, nil
		}
	default:
		return Tokens{}, ErrUnsupportedGrantType
	}
}
