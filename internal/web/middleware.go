package web

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"slices"

	"github.com/freekieb7/smauth/internal/auth"
	"github.com/freekieb7/smauth/internal/http"
	"github.com/freekieb7/smauth/internal/session"
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

func SessionMiddleware(logger *slog.Logger, sessionStore *session.Store) http.MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(ctx context.Context, req *http.Request, res *http.Response) error {
			cookie, err := req.Cookie("SID")
			if err != nil || cookie.Value == "" {
				// No session cookie, proceed without session
				ses, err := sessionStore.New()
				if err != nil {
					logger.Error("failed to create new session", "error", err)
					return res.Send(http.StatusInternalServerError)
				}

				res.SetCookie(http.Cookie{
					Name:     "SID",
					Value:    ses.ID,
					Path:     "/",
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteLaxMode,
				})

				return next(ctx, req, res)
			}

			sessionID := cookie.Value

			session, err := sessionStore.Get(ctx, sessionID)
			if err != nil {
				logger.Error("failed to get session", "error", err)
				return res.Send(http.StatusInternalServerError)
			}

			ctx = context.WithValue(ctx, SessionContextKey, session)
			return next(ctx, req, res)
		})
	}
}

func CSRFMiddleware(logger *slog.Logger, sessionStore *session.Store) http.MiddlewareFunc {
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
					sess, err = sessionStore.Save(ctx, sess)
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

func AuthenticatedMiddleware(logger *slog.Logger) http.MiddlewareFunc {
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

			if !sess.UserID.IsSet() {
				logger.Warn("unauthenticated access attempt")
				return res.Send(http.StatusUnauthorized)
			}

			return next(ctx, req, res)
		})
	}
}

func PageAuthenticatedMiddleware(logger *slog.Logger) http.MiddlewareFunc {
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

			if !sess.UserID.IsSet() {
				logger.Warn("unauthenticated page access attempt")
				return res.SendRedirect(http.StatusSeeOther, fmt.Sprintf("/login?returnTo=%s", url.QueryEscape(returnTo)))
			}

			return next(ctx, req, res)
		})
	}
}

func AdminAuthenticatedMiddleware(logger *slog.Logger, authStore *auth.Store) http.MiddlewareFunc {
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

			if !sess.UserID.IsSet() {
				logger.Warn("unauthenticated access attempt to admin route")
				return res.Send(http.StatusUnauthorized)
			}

			user, err := authStore.GetUser(ctx, sess.UserID.V)
			if err != nil {
				logger.Error("failed to get user for admin authentication check", "error", err)
				return res.Send(http.StatusInternalServerError)
			}

			isAdmin := user.Role == auth.RoleAdmin

			if !isAdmin {
				logger.Warn("non-admin user attempted to access admin route", "user_id", sess.UserID)
				return res.Send(http.StatusForbidden)
			}

			return next(ctx, req, res)
		})
	}
}

func ScopeProtectedMiddleware(requiredScopes []string, logger *slog.Logger) http.MiddlewareFunc {
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
