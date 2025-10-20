package oauth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"slices"

	"github.com/freekieb7/smauth/internal/account"
	"github.com/freekieb7/smauth/internal/database"
	"github.com/freekieb7/smauth/internal/telemetry"
	"github.com/google/uuid"
)

var (
	ErrInvalidClientCredentials = errors.New("invalid client credentials")
	ErrUnsupportedGrantType     = errors.New("unsupported grant type")
	ErrUnsupportedResponseType  = errors.New("unsupported response type")
	ErrInvalidRedirectURI       = errors.New("invalid redirect_uri")
	ErrRequestedUngrantedScopes = errors.New("requested scopes include ungranted scopes")
	ErrInvalidRefreshToken      = errors.New("invalid refresh token")
	ErrInvalidPKCEParameters    = errors.New("PKCE parameters are missing")
)

type Service struct {
	Logger       *telemetry.Logger
	DB           *database.Database
	Store        *Store
	AccountStore *account.Store
}

func NewService(logger *telemetry.Logger, db *database.Database, store *Store, accountStore *account.Store) Service {
	return Service{
		Logger:       logger,
		DB:           db,
		Store:        store,
		AccountStore: accountStore,
	}
}

func (s *Service) IssueCode(ctx context.Context, responseType string, clientID, userID uuid.UUID, redirectURI string, scopes []string, codeChallenge string, codeMethod string, allowGrant bool) (AuthorizationCode, account.Client, []Scope, error) {
	switch responseType {
	case "code":
		// Valid response type
		client, err := s.AccountStore.GetClient(ctx, clientID)
		if err != nil {
			if errors.Is(err, account.ErrClientNotFound) {
				return AuthorizationCode{}, account.Client{}, nil, ErrInvalidClientCredentials
			}

			s.Logger.Error("failed to validate client credentials", "error", err)
			return AuthorizationCode{}, account.Client{}, nil, err
		}

		if !slices.Contains(client.RedirectURIs, redirectURI) {
			return AuthorizationCode{}, client, nil, ErrInvalidRedirectURI
		}

		// Validate PKCE (must prove challenge)
		if codeChallenge == "" || codeMethod == "" {
			return AuthorizationCode{}, client, nil, ErrInvalidPKCEParameters
		}

		if codeMethod != "S256" {
			return AuthorizationCode{}, client, nil, errors.New("unsupported code challenge method")
		}

		// Validate redirect URI
		if redirectURI == "" {
			return AuthorizationCode{}, client, nil, errors.New("redirect_uri is required")
		}

		// Validate scopes
		grantedScopes, ungrantedScopes, err := s.Store.FilterGrantedFromUngrantedScopes(ctx, userID, clientID, scopes)
		if err != nil {
			s.Logger.Error("failed to filter granted from ungranted scopes", "error", err)
			return AuthorizationCode{}, client, nil, err
		}

		if len(ungrantedScopes) > 0 {
			if !allowGrant {
				return AuthorizationCode{}, client, ungrantedScopes, ErrRequestedUngrantedScopes
			}

			// Grant the requested scopes
			if err := s.Store.Grant(ctx, userID, clientID, ungrantedScopes); err != nil {
				s.Logger.Error("failed to grant scopes", "user_id", userID, "client_id", clientID, "scopes", ungrantedScopes, "error", err)
				return AuthorizationCode{}, client, nil, err
			}
			grantedScopes = append(grantedScopes, ungrantedScopes...)
			ungrantedScopes = nil
		}

		scopes := make([]string, 0, len(grantedScopes))
		for _, scope := range grantedScopes {
			scopes = append(scopes, scope.Name)
		}

		// Create authorization code
		code, err := s.Store.NewAuthorizationCode(userID, clientID, redirectURI, scopes, codeChallenge, codeMethod)
		if err != nil {
			s.Logger.Error("failed to create authorization code", "error", err)
			return AuthorizationCode{}, client, nil, err
		}

		// Save authorization code
		code, err = s.Store.SaveAuthorizationCode(ctx, code)
		if err != nil {
			s.Logger.Error("failed to save authorization code", "error", err)
			return AuthorizationCode{}, client, nil, err
		}

		return code, client, nil, nil
	default:
		return AuthorizationCode{}, account.Client{}, nil, ErrUnsupportedResponseType
	}
}

type Tokens struct {
	AccessToken  AccessToken
	RefreshToken RefreshToken
}

func (s *Service) IssueTokens(ctx context.Context, grantType string, clientID uuid.UUID, clientSecret string, scopes []string, codeVerifier, code, redirectURI, refreshToken string) (Tokens, error) {
	switch grantType {
	case "client_credentials":
		// Validate client credentials
		client, err := s.AccountStore.GetClient(ctx, clientID)
		if err != nil {
			if errors.Is(err, account.ErrClientNotFound) {
				return Tokens{}, ErrInvalidClientCredentials
			}

			s.Logger.Error("failed to validate client credentials", "error", err)
			return Tokens{}, err
		}

		if client.Secret != clientSecret {
			return Tokens{}, ErrInvalidClientCredentials
		}

		// Validate scopes
		var grantedScopes []string

		if err := s.DB.Conn.QueryRow(ctx, `SELECT array_agg(scope.name) FROM tbl_account_permission permission
			JOIN tbl_scope scope ON permission.scope_id = scope.id
			WHERE permission.account_id = $1 AND scope.name = ANY($2::text[])`, client.ID, scopes).Scan(&grantedScopes); err != nil {
			if !errors.Is(err, database.ErrNoRows) {
				s.Logger.Error("failed to query granted scopes", "error", err)
				return Tokens{}, err
			}
		}

		accessToken, err := s.Store.NewAccessToken(client.ID, grantedScopes)
		if err != nil {
			s.Logger.Error("failed to create access token", "error", err)
			return Tokens{}, err
		}

		// Save access token
		accessToken, err = s.Store.SaveAccessToken(ctx, accessToken)
		if err != nil {
			s.Logger.Error("failed to save access token", "error", err)
			return Tokens{}, err
		}

		return Tokens{AccessToken: accessToken}, nil
	case "authorization_code":
		{
			client, err := s.AccountStore.GetClient(ctx, clientID)
			if err != nil {
				if errors.Is(err, account.ErrClientNotFound) {
					return Tokens{}, ErrInvalidClientCredentials
				}

				s.Logger.Error("failed to validate client credentials", "error", err)
				return Tokens{}, err
			}

			if !client.IsPublic && client.Secret != clientSecret {
				return Tokens{}, ErrInvalidClientCredentials
			}

			authCode, err := s.Store.GetAuthorizationCode(ctx, code)
			if err != nil {
				if errors.Is(err, ErrAuthorizationCodeNotFound) {
					return Tokens{}, errors.New("invalid authorization code")
				}

				s.Logger.Error("failed to retrieve authorization code", "error", err)
				return Tokens{}, err
			}

			if authCode.ClientID != clientID {
				return Tokens{}, ErrInvalidClientCredentials
			}

			if authCode.RedirectURI != redirectURI {
				return Tokens{}, ErrInvalidRedirectURI
			}

			// MISSING: PKCE validation
			if authCode.CodeChallenge != "" {
				if codeVerifier == "" {
					return Tokens{}, errors.New("code_verifier is required for PKCE")
				}

				// Validate PKCE challenge
				switch authCode.CodeMethod {
				case "S256":
					hash := sha256.Sum256([]byte(codeVerifier))
					expected := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
					if authCode.CodeChallenge != expected {
						return Tokens{}, errors.New("invalid code_verifier")
					}
				default:
					return Tokens{}, errors.New("unsupported code_challenge_method")
				}
			}

			// Mark authorization code as used
			if err := s.Store.MarkAuthorizationCodeAsUsed(ctx, authCode.Code); err != nil {
				s.Logger.Error("failed to mark authorization code as used", "error", err)
				return Tokens{}, err
			}

			accessToken, err := s.Store.NewAccessToken(authCode.UserID, authCode.Scopes)
			if err != nil {
				s.Logger.Error("failed to create access token", "error", err)
				return Tokens{}, err
			}

			// Save access token
			accessToken, err = s.Store.SaveAccessToken(ctx, accessToken)
			if err != nil {
				s.Logger.Error("failed to save access token", "error", err)
				return Tokens{}, err
			}

			var refreshToken RefreshToken

			if slices.Contains(authCode.Scopes, "offline_access") {
				refreshTokenChain, err := s.Store.NewRefreshTokenChain(clientID, authCode.UserID, authCode.Scopes)
				if err != nil {
					s.Logger.Error("failed to create refresh token chain", "error", err)
					return Tokens{}, err
				}

				// Save refresh token chain
				refreshTokenChain, err = s.Store.SaveRefreshTokenChain(ctx, refreshTokenChain)
				if err != nil {
					s.Logger.Error("failed to save refresh token chain", "error", err)
					return Tokens{}, err
				}

				refreshToken, err = s.Store.NewRefreshToken()
				if err != nil {
					s.Logger.Error("failed to create refresh token", "error", err)
					return Tokens{}, err
				}

				// Save refresh token
				refreshToken, err = s.Store.SaveRefreshToken(ctx, refreshTokenChain.ID, refreshToken)
				if err != nil {
					s.Logger.Error("failed to save refresh token", "error", err)
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
			client, err := s.AccountStore.GetClient(ctx, clientID)
			if err != nil {
				if errors.Is(err, account.ErrClientNotFound) {
					return Tokens{}, ErrInvalidClientCredentials
				}

				s.Logger.Error("failed to validate client credentials", "error", err)
				return Tokens{}, err
			}

			if !client.IsPublic && client.Secret != clientSecret {
				return Tokens{}, ErrInvalidClientCredentials
			}

			chain, err := s.Store.GetRefreshTokenChainByToken(ctx, refreshToken)
			if err != nil {
				if errors.Is(err, ErrRefreshTokenChainNotFound) {
					return Tokens{}, ErrInvalidRefreshToken
				}

				s.Logger.Error("failed to retrieve refresh token chain and token", "error", err)
				return Tokens{}, err
			}

			if chain.ClientID != clientID {
				return Tokens{}, ErrInvalidClientCredentials
			}

			accessToken, err := s.Store.NewAccessToken(chain.UserID, chain.Scopes)
			if err != nil {
				s.Logger.Error("failed to create access token", "error", err)
				return Tokens{}, err
			}

			// Save access token
			accessToken, err = s.Store.SaveAccessToken(ctx, accessToken)
			if err != nil {
				s.Logger.Error("failed to save access token", "error", err)
				return Tokens{}, err
			}

			newRefreshToken, err := s.Store.NewRefreshToken()
			if err != nil {
				s.Logger.Error("failed to create refresh token", "error", err)
				return Tokens{}, err
			}

			// Mark old refresh token as used
			if err := s.Store.MarkRefreshTokenAsUsed(ctx, refreshToken); err != nil {
				s.Logger.Error("failed to mark refresh token as used", "error", err)
				return Tokens{}, err
			}

			// Save refresh token
			newRefreshToken, err = s.Store.SaveRefreshToken(ctx, chain.ID, newRefreshToken)
			if err != nil {
				s.Logger.Error("failed to save refresh token", "error", err)
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
