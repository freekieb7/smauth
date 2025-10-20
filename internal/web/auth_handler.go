package web

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/freekieb7/smauth/internal/auth"
	"github.com/freekieb7/smauth/internal/http"
	"github.com/freekieb7/smauth/internal/oauth"
	"github.com/freekieb7/smauth/internal/session"
	"github.com/freekieb7/smauth/internal/util"
	"github.com/freekieb7/smauth/internal/web/ui"
	"github.com/freekieb7/smauth/internal/web/ui/template/layout"
	"github.com/freekieb7/smauth/internal/web/ui/template/page"
	"github.com/google/uuid"
)

type AuthHandler struct {
	Logger       *slog.Logger
	SessionStore *session.Store
	AuthStore    *auth.Store
	OAuthStore   *oauth.Store
	OAuthIssuer  *oauth.Issuer
}

func NewAuthHandler(logger *slog.Logger, sessionStore *session.Store, authStore *auth.Store, oauthStore *oauth.Store, oauthIssuer *oauth.Issuer) *AuthHandler {
	return &AuthHandler{
		Logger:       logger,
		SessionStore: sessionStore,
		AuthStore:    authStore,
		OAuthStore:   oauthStore,
		OAuthIssuer:  oauthIssuer,
	}
}

func (h *AuthHandler) RegisterRoutes(router *http.Router) {
	sessionMiddleware := SessionMiddleware(h.Logger, h.SessionStore)
	csrfMiddleware := CSRFMiddleware(h.Logger, h.SessionStore)
	noCacheMiddleware := NoCacheMiddleware()
	authenticatedMiddleware := AuthenticatedMiddleware(h.Logger)

	// Static files
	router.Static("/static", "internal/web/ui/static")

	// Login routes
	router.Group("/login", func(group *http.Router) {
		group.GET("", h.ShowLoginPage)
		group.POST("", h.HandleLogin)
	}, sessionMiddleware, csrfMiddleware)

	router.POST("/logout", h.HandleLogout, sessionMiddleware, csrfMiddleware, authenticatedMiddleware)
	// OAuth routes
	router.Group("/oauth", func(group *http.Router) {
		group.GET("/authorize", h.HandleAuthorize, sessionMiddleware, csrfMiddleware)
		group.POST("/token", h.HandleToken, sessionMiddleware)
		group.POST("/grant", h.HandleGrant, sessionMiddleware, csrfMiddleware)
		group.POST("/introspect", h.HandleIntrospection)
	}, noCacheMiddleware)
}

func (h *AuthHandler) ShowLoginPage(ctx context.Context, req *http.Request, res *http.Response) error {
	sess := ctx.Value(SessionContextKey).(session.Session)

	returnTo := req.URLQueryParam("returnTo")

	if sess.UserID.IsSet() {
		// User already logged in, redirect to home or dashboard
		return res.SendRedirect(http.StatusSeeOther, returnTo)
	}

	csrfToken := sess.Values[CSRFTokenSessionValueKey].(string)

	return ui.Render(ctx, res, page.Login(page.LoginProps{RootProps: layout.RootProps{
		Title:     "Login",
		CSRFToken: csrfToken,
	},
		ReturnTo: returnTo,
	}))
}

type LoginForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	ReturnTo string `json:"return_to"`
}

func (h *AuthHandler) HandleLogin(ctx context.Context, req *http.Request, res *http.Response) error {
	// Parse login form
	var form LoginForm
	if err := req.DecodeJSON(&form); err != nil {
		h.Logger.Warn("failed to parse login form", "error", err)
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	err := h.AuthStore.ValidateCredentials(ctx, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) || errors.Is(err, auth.ErrInvalidCredentials) {
			return res.SendJSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		}

		h.Logger.Error("failed to get user by email", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	user, err := h.AuthStore.GetUserByEmail(ctx, form.Email)
	if err != nil {
		h.Logger.Error("failed to get user by email", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	sess := ctx.Value(SessionContextKey).(session.Session)
	sess.UserID = util.Some(user.ID)

	sess, err = h.SessionStore.Save(ctx, sess)
	if err != nil {
		h.Logger.Error("failed to save session", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	h.Logger.Info("user logged in", "user_id", user.ID)

	redirectTo := "/authorize"
	if form.ReturnTo != "" {
		redirectTo = form.ReturnTo
	}

	return res.SendRedirect(http.StatusSeeOther, redirectTo)
}

func (h *AuthHandler) HandleLogout(ctx context.Context, req *http.Request, res *http.Response) error {
	sess := ctx.Value(SessionContextKey).(session.Session)

	err := h.SessionStore.Delete(ctx, sess.ID)
	if err != nil {
		h.Logger.Error("failed to delete session", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	h.Logger.Info("user logged out", "user_id", sess.UserID)

	// Clear session cookie
	res.SetCookie(http.Cookie{
		Name:    "SID",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	})

	return res.SendJSON(http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *AuthHandler) HandleAuthorize(ctx context.Context, req *http.Request, res *http.Response) error {
	responseTypeRaw := req.URLQueryParam("response_type")
	clientIDStrRaw := req.URLQueryParam("client_id")
	redirectURIRaw := req.URLQueryParam("redirect_uri")
	scopeRaw := req.URLQueryParam("scope")
	stateRaw := req.URLQueryParam("state")
	codeChallengeRaw := req.URLQueryParam("code_challenge")
	codeChallengeMethodRaw := req.URLQueryParam("code_challenge_method")

	scopes := strings.Fields(scopeRaw)

	// Validate input
	if responseTypeRaw == "" || clientIDStrRaw == "" || redirectURIRaw == "" {
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "missing required parameters"})
	}

	clientID, err := uuid.Parse(clientIDStrRaw)
	if err != nil {
		h.Logger.Warn("invalid client ID format", "error", err)
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "invalid client ID"})
	}

	sess := ctx.Value(SessionContextKey).(session.Session)
	if !sess.UserID.IsSet() {
		// User not logged in, redirect to login page
		return res.SendRedirect(http.StatusSeeOther, "/login?returnTo="+url.QueryEscape(req.URL()))
	}

	grantScopes := false
	authorizeVal, exists := sess.Values[AuthorizeSessionValueKey]
	if exists {
		authorize, ok := authorizeVal.(bool)
		if !ok {
			h.Logger.Error("invalid authorize session value type", "type", authorizeVal)
			return RedirectOAuthResponse(res, redirectURIRaw, stateRaw, "server_error")
		}

		delete(sess.Values, AuthorizeSessionValueKey)
		sess, err = h.SessionStore.Save(ctx, sess)
		if err != nil {
			h.Logger.Error("failed to save session", "error", err)
			return RedirectOAuthResponse(res, redirectURIRaw, stateRaw, "server_error")
		}

		if !authorize {
			// User denied authorization
			return RedirectOAuthResponse(res, redirectURIRaw, stateRaw, "access_denied")
		}

		grantScopes = true
	}

	authorizationCode, ungrantedScopes, err := h.OAuthIssuer.IssueCode(ctx, responseTypeRaw, clientID, sess.UserID.V, redirectURIRaw, scopes, codeChallengeRaw, codeChallengeMethodRaw, grantScopes)
	if err != nil {
		if errors.Is(err, oauth.ErrUnsupportedResponseType) {
			h.Logger.Error("unsupported response type", "error", err)
			return RedirectOAuthResponse(res, redirectURIRaw, stateRaw, "unsupported_response_type")
		}

		if errors.Is(err, oauth.ErrInvalidClientCredentials) {
			h.Logger.Error("invalid client credentials", "error", err)
			return RedirectOAuthResponse(res, redirectURIRaw, stateRaw, "invalid_client")
		}

		// Let the user authorize the requested scopes
		if errors.Is(err, oauth.ErrRequestedUngrantedScopes) {
			csrfToken, err := util.GenerateRandomString(32)
			if err != nil {
				h.Logger.Error("failed to generate CSRF token", "error", err)
				return RedirectOAuthResponse(res, redirectURIRaw, stateRaw, "server_error")
			}

			sess.Values[CSRFTokenSessionValueKey] = csrfToken
			sess, err = h.SessionStore.Save(ctx, sess)
			if err != nil {
				h.Logger.Error("failed to save session", "error", err)
				return RedirectOAuthResponse(res, redirectURIRaw, stateRaw, "server_error")
			}

			return ui.Render(ctx, res, page.Authorize(page.AuthorizeProps{RootProps: layout.RootProps{
				Title:     "Authorize",
				CSRFToken: csrfToken,
			},
				AuthorizeURL: req.URL(),
				Scopes:       ungrantedScopes,
			}))
		}

		h.Logger.Error("Failed to issue authorization code", "error", err)
		return RedirectOAuthResponse(res, redirectURIRaw, stateRaw, "server_error")
	}

	// Build redirect URL
	redirectURI := redirectURIRaw + "?code=" + url.QueryEscape(authorizationCode.Code)
	if stateRaw != "" {
		redirectURI += "&state=" + url.QueryEscape(stateRaw)
	}

	h.Logger.Info("authorization granted", "user_id", sess.UserID.V, "client_id", clientID, "scopes", scopes)
	return RedirectOAuthResponse(res, redirectURI, stateRaw, "")
	// return res.SendRedirect(http.StatusFound, redirectURI)
}

func RedirectOAuthResponse(res *http.Response, redirectURI string, state string, errCode string) error {
	if redirectURI == "" { // cannot redirect, return JSON
		return JSONOAuthErrorResponse(res, http.StatusBadRequest, errCode, "")
	}

	// best effort build
	if errCode != "" {
		redirectURI += "?error=" + url.QueryEscape(errCode)
	}
	if state != "" {
		redirectURI += "&state=" + url.QueryEscape(state)
	}

	return res.SendRedirect(http.StatusSeeOther, redirectURI)
}

// Helper functions for OAuth error formatting --------------------------------------------
func JSONOAuthErrorResponse(res *http.Response, code http.StatusCode, status string, description string) error {
	return res.SendJSON(code, map[string]any{
		"error":             status,
		"error_description": description,
	})
}

type TokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

func (h *AuthHandler) HandleToken(ctx context.Context, req *http.Request, res *http.Response) error {
	h.Logger.Info("handling token request")

	rawGrantType := req.FormValue("grant_type")
	rawClientIDStr := req.FormValue("client_id")
	rawClientSecret := req.FormValue("client_secret")
	rawScope := req.FormValue("scope")
	rawRedirectURI := req.FormValue("redirect_uri")
	rawCode := req.FormValue("code")
	rawRefreshToken := req.FormValue("refresh_token")

	authHeader := req.Header("Authorization")
	if authHeader != "" {
		// Handle authorization header
		authHeader := req.Header("Authorization")
		if strings.HasPrefix(authHeader, "Basic ") {
			// Decode client credentials
			credentials, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authHeader, "Basic "))
			if err != nil {
				h.Logger.Warn("failed to decode authorization header", "error", err)
				return res.SendJSON(http.StatusUnauthorized, map[string]string{"error": "invalid client credentials"})
			}

			parts := strings.SplitN(string(credentials), ":", 2)
			if len(parts) != 2 {
				h.Logger.Warn("invalid authorization header format")
				return res.SendJSON(http.StatusUnauthorized, map[string]string{"error": "invalid client credentials"})
			}

			rawClientIDStr = parts[0]
			rawClientSecret = parts[1]
		}
	}

	clientID, err := uuid.Parse(rawClientIDStr)
	if err != nil {
		h.Logger.Warn("invalid client ID format", "error", err)
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "invalid client ID"})
	}
	scopes := strings.Fields(rawScope)

	tokens, err := h.OAuthIssuer.IssueTokens(ctx, rawGrantType, clientID, rawClientSecret, scopes, rawCode, rawRedirectURI, rawRefreshToken)
	if err != nil {
		if errors.Is(err, oauth.ErrUnsupportedGrantType) {
			return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "unsupported grant type"})
		}

		if errors.Is(err, oauth.ErrInvalidClientCredentials) {
			return res.SendJSON(http.StatusUnauthorized, map[string]string{"error": "invalid client credentials"})
		}

		if errors.Is(err, oauth.ErrInvalidRefreshToken) {
			return res.SendJSON(http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
		}

		h.Logger.Error("failed to issue token", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return res.SendJSON(http.StatusOK, TokensResponse{
		AccessToken:  tokens.AccessToken.Token,
		RefreshToken: tokens.RefreshToken.Token,
		ExpiresIn:    time.Until(tokens.AccessToken.ExpiresAt).Milliseconds() / 1000,
		Scope:        strings.Join(tokens.AccessToken.Scopes, " "),
		TokenType:    "Bearer",
	})
}

type GrantRequestForm struct {
	Authorize bool `json:"authorize"`
}

func (h *AuthHandler) HandleGrant(ctx context.Context, req *http.Request, res *http.Response) error {
	var form GrantRequestForm
	if err := req.DecodeJSON(&form); err != nil {
		h.Logger.Warn("failed to parse grant request form", "error", err)
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	sess := ctx.Value(SessionContextKey).(session.Session)
	if !sess.UserID.IsSet() {
		return res.SendJSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Process the grant request
	if form.Authorize {
		// Handle authorization
		sess.Values[AuthorizeSessionValueKey] = true
	} else {
		// Handle denial
		sess.Values[AuthorizeSessionValueKey] = false
	}

	sess, err := h.SessionStore.Save(ctx, sess)
	if err != nil {
		h.Logger.Error("failed to save session", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return res.SendJSON(http.StatusOK, map[string]string{"message": "grant processed"})
}

func (h *AuthHandler) HandleIntrospection(ctx context.Context, req *http.Request, res *http.Response) error {
	token := req.FormValue("token")

	// Validate input
	if token == "" {
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "token is required"})
	}

	accessToken, err := h.OAuthStore.GetAccessToken(ctx, token)
	if err != nil {
		if errors.Is(err, oauth.ErrAccessTokenNotFound) {
			return res.SendJSON(http.StatusUnauthorized, map[string]string{"active": "false"})
		}

		h.Logger.Error("failed to get access token", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return res.SendJSON(http.StatusOK, map[string]any{
		"active": true,
		"scope":  strings.Join(accessToken.Scopes, " "),
	})
}
