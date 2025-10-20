package web

import (
	"context"
	"fmt"
	"net/url"
	"slices"

	"github.com/freekieb7/smauth/internal/account"
	"github.com/freekieb7/smauth/internal/http"
	"github.com/freekieb7/smauth/internal/oauth"
	"github.com/freekieb7/smauth/internal/session"
	"github.com/freekieb7/smauth/internal/telemetry"
	"github.com/freekieb7/smauth/internal/util"
)

func NoCacheMiddleware() http.MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(ctx context.Context, req *http.Request, res *http.Response) error {
			res.SetHeader("Cache-Control", "no-cache, no-store, must-revalidate")
			res.SetHeader("Pragma", "no-cache")
			res.SetHeader("Expires", "0")

			return next(ctx, req, res)
		})
	}
}

func SessionMiddleware(logger *telemetry.Logger, sessionStore *session.Store) http.MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(ctx context.Context, req *http.Request, res *http.Response) error {
			cookie, err := req.Cookie("SID")
			if err != nil || cookie.Value == "" {
				// No session cookie, proceed without session
				sess, err := sessionStore.NewSession()
				if err != nil {
					logger.Error("failed to create new session", "error", err)
					return res.Send(http.StatusInternalServerError)
				}

				sessionStore.SetCookie(ctx, res, sess.ID)

				ctx = context.WithValue(ctx, SessionContextKey, sess)
				return next(ctx, req, res)
			}

			sessionID := cookie.Value

			sess, err := sessionStore.GetSession(ctx, sessionID)
			if err != nil {
				if err != session.ErrSessionNotFound {
					logger.Error("failed to get session", "error", err)
					return res.Send(http.StatusInternalServerError)
				}

				// Session not found, create a new one
				sess, err = sessionStore.NewSession()
				if err != nil {
					logger.Error("failed to create new session", "error", err)
					return res.Send(http.StatusInternalServerError)
				}

				sessionStore.SetCookie(ctx, res, sess.ID)
			}

			ctx = context.WithValue(ctx, SessionContextKey, sess)
			return next(ctx, req, res)
		})
	}
}

func CSRFMiddleware(logger *telemetry.Logger, sessionStore *session.Store) http.MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(ctx context.Context, req *http.Request, res *http.Response) error {
			sessVal := ctx.Value(SessionContextKey)
			if sessVal == nil {
				logger.Warn("no session found in context for CSRF check")
				return res.Send(http.StatusForbidden)
			}

			sess, ok := sessVal.(session.Session)
			if !ok {
				logger.Warn("invalid session type in context for CSRF check")
				return res.Send(http.StatusForbidden)
			}

			switch req.Method() {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
				{
					token, err := util.GenerateRandomString(32)
					if err != nil {
						logger.Error("failed to generate CSRF token", "error", err)
						return res.Send(http.StatusInternalServerError)
					}

					sess.Values[CSRFTokenSessionValueKey] = token
					sess, err = sessionStore.SaveSession(ctx, sess)
					if err != nil {
						logger.Error("failed to save session with CSRF token", "error", err)
						return res.Send(http.StatusInternalServerError)
					}

					ctx = context.WithValue(ctx, SessionContextKey, sess)

					return next(ctx, req, res)
				}
			default:
				{
					token := req.Header("X-CSRF-Token")
					if token == "" {
						token = req.FormValue("csrf_token")
					}

					if token == "" {
						logger.Warn("missing CSRF token in request")
						return res.Send(http.StatusForbidden)
					}

					csrfToken, found := sess.Values[CSRFTokenSessionValueKey].(string)
					if !found || csrfToken == "" {
						logger.Warn("missing CSRF token in session")
						return res.Send(http.StatusForbidden)
					}

					if token != csrfToken {
						logger.Warn("invalid CSRF token", "expected", csrfToken, "got", token)
						return res.Send(http.StatusForbidden)
					}

					return next(ctx, req, res)
				}
			}

		})
	}
}

func AuthenticatedMiddleware(logger *telemetry.Logger) http.MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(ctx context.Context, req *http.Request, res *http.Response) error {
			sessVal := ctx.Value(SessionContextKey)
			if sessVal == nil {
				logger.Warn("no session found in context for authentication check")
				return res.Send(http.StatusUnauthorized)
			}

			sess, ok := sessVal.(session.Session)
			if !ok {
				logger.Warn("invalid session type in context for authentication check")
				return res.Send(http.StatusUnauthorized)
			}

			if !sess.AccountID.IsSet() {
				logger.Warn("unauthenticated access attempt")
				return res.Send(http.StatusUnauthorized)
			}

			return next(ctx, req, res)
		})
	}
}

func PageAuthenticatedMiddleware(logger *telemetry.Logger) http.MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(ctx context.Context, req *http.Request, res *http.Response) error {
			returnTo := req.URL()

			sessVal := ctx.Value(SessionContextKey)
			if sessVal == nil {
				logger.Warn("no session found in context for page authentication check")
				return res.SendRedirect(http.StatusSeeOther, fmt.Sprintf("/login?returnTo=%s", url.QueryEscape(returnTo)))
			}

			sess, ok := sessVal.(session.Session)
			if !ok {
				logger.Warn("invalid session type in context for page authentication check")
				return res.SendRedirect(http.StatusSeeOther, fmt.Sprintf("/login?returnTo=%s", url.QueryEscape(returnTo)))
			}

			if !sess.AccountID.IsSet() {
				logger.Warn("unauthenticated page access attempt")
				return res.SendRedirect(http.StatusSeeOther, fmt.Sprintf("/login?returnTo=%s", url.QueryEscape(returnTo)))
			}

			return next(ctx, req, res)
		})
	}
}

func AdminOnlyMiddleware(logger *telemetry.Logger, accountStore *account.Store) http.MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(ctx context.Context, req *http.Request, res *http.Response) error {
			sessVal := ctx.Value(SessionContextKey)
			if sessVal == nil {
				logger.Warn("no session found in context for admin authentication check")
				return res.Send(http.StatusUnauthorized)
			}

			sess, ok := sessVal.(session.Session)
			if !ok {
				logger.Warn("invalid session type in context for admin authentication check")
				return res.Send(http.StatusUnauthorized)
			}

			if !sess.AccountID.IsSet() {
				logger.Warn("unauthenticated access attempt to admin route")
				return res.Send(http.StatusUnauthorized)
			}

			user, err := accountStore.GetUser(ctx, sess.AccountID.V)
			if err != nil {
				logger.Error("failed to get user for admin authentication check", "error", err)
				return res.Send(http.StatusInternalServerError)
			}

			isAdmin := user.Role == account.RoleAdmin

			if !isAdmin {
				logger.Warn("non-admin user attempted to access admin route", "user_id", user.ID)
				return res.Send(http.StatusForbidden)
			}

			return next(ctx, req, res)
		})
	}
}

func AccessTokenMiddleware(logger *telemetry.Logger, sessionStore *session.Store, oauthService *oauth.Service) http.MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(ctx context.Context, req *http.Request, res *http.Response) error {
			// Get access token from request
			token := req.Header("Authorization")
			if token == "" {
				return res.Send(http.StatusUnauthorized)
			}

			// Validate access token format
			const bearerPrefix = "Bearer "
			if len(token) <= len(bearerPrefix) || token[:len(bearerPrefix)] != bearerPrefix {
				return res.Send(http.StatusUnauthorized)
			}

			token = token[len(bearerPrefix):]

			// Validate access token
			accessToken, err := oauthService.Store.GetAccessToken(ctx, token)
			if err != nil {
				logger.Error("failed to get access token", "error", err)
				return res.Send(http.StatusInternalServerError)
			}

			sess, err := sessionStore.NewSession()
			if err != nil {
				logger.Error("failed to create new session", "error", err)
				return res.Send(http.StatusInternalServerError)
			}

			sess.AccountID = util.Some(accessToken.AccountID)
			sess.Values[ScopesSessionValueKey] = accessToken.Scopes

			ctx = context.WithValue(ctx, SessionContextKey, sess)

			return next(ctx, req, res)
		})
	}
}

func ScopeProtectedMiddleware(logger *telemetry.Logger, requiredScopes ...string) http.MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(ctx context.Context, req *http.Request, res *http.Response) error {
			sessVal := ctx.Value(SessionContextKey)
			if sessVal == nil {
				logger.Warn("no session found in context for scope check")
				return res.Send(http.StatusForbidden)
			}

			sess, ok := sessVal.(session.Session)
			if !ok {
				logger.Warn("invalid session type in context for scope check")
				return res.Send(http.StatusForbidden)
			}

			scopesVal, found := sess.Values[ScopesSessionValueKey]
			if !found {
				logger.Warn("no scopes found in session for scope check")
				return res.Send(http.StatusForbidden)
			}

			scopes, ok := scopesVal.([]string)
			if !ok {
				logger.Warn("invalid scopes type in session for scope check")
				return res.Send(http.StatusForbidden)
			}

			for _, requiredScope := range requiredScopes {
				if !slices.Contains(scopes, requiredScope) {
					logger.Warn("insufficient scope", "required_scope", requiredScope)
					return res.Send(http.StatusForbidden)
				}
			}

			return next(ctx, req, res)
		})
	}
}
