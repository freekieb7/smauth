package web

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/freekieb7/smauth/internal"
	"github.com/freekieb7/smauth/internal/account"
	"github.com/freekieb7/smauth/internal/http"
	"github.com/freekieb7/smauth/internal/oauth"
	"github.com/freekieb7/smauth/internal/session"
	"github.com/freekieb7/smauth/internal/telemetry"
	"github.com/freekieb7/smauth/internal/util"
	"github.com/freekieb7/smauth/internal/web/ui"
	"github.com/freekieb7/smauth/internal/web/ui/template/layout"
	"github.com/freekieb7/smauth/internal/web/ui/template/page"
	"github.com/google/uuid"
)

// Constants ================================================================

const (
	// Microsoft OAuth URLs and settings
	microsoftAuthBaseURL = "https://login.microsoftonline.com"
	microsoftOAuthScopes = "openid profile email User.Read"

	// Session keys
	microsoftOAuthStateKey  = "microsoft_oauth_state"
	microsoftOAuthReturnKey = "microsoft_oauth_return_to"

	// Token request timeout
	tokenRequestTimeout = 5 * time.Second

	// Default redirect paths
	defaultHomePath  = "/dashboard"
	defaultLoginPath = "/login"

	// Common error messages
	errInternalServer      = "internal server error"
	errInvalidRequest      = "invalid request"
	errInvalidCredentials  = "invalid credentials"
	errUnauthorized        = "unauthorized"
	errSessionError        = "session_error"
	errTokenExchangeFailed = "token_exchange_failed"
)

// Common helper functions =================================================

// sendInternalError sends a standardized internal server error response
func (h *AuthHandler) sendInternalError(res *http.Response, logger *telemetry.Logger, message string, err error) error {
	logger.Error(message, "error", err)
	return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": errInternalServer})
}

// sendBadRequestError sends a standardized bad request error response
func (h *AuthHandler) sendBadRequestError(res *http.Response, message string) error {
	return res.SendJSON(http.StatusBadRequest, map[string]string{"error": message})
}

// sendUnauthorizedError sends a standardized unauthorized error response
func (h *AuthHandler) sendUnauthorizedError(res *http.Response, message string) error {
	return res.SendJSON(http.StatusUnauthorized, map[string]string{"error": message})
}

// redirectToLogin redirects to the login page with optional error and return URL
func (h *AuthHandler) redirectToLogin(res *http.Response, errorCode string, returnTo string) error {
	loginURL := defaultLoginPath
	params := url.Values{}

	if errorCode != "" {
		params.Set("error", errorCode)
	}
	if returnTo != "" {
		params.Set("returnTo", returnTo)
	}

	if len(params) > 0 {
		loginURL += "?" + params.Encode()
	}

	return res.SendRedirect(http.StatusSeeOther, loginURL)
}

// saveSessionWithError saves a session and handles errors consistently
func (h *AuthHandler) saveSessionWithError(ctx context.Context, sess session.Session, res *http.Response) (session.Session, error) {
	savedSess, err := h.SessionStore.SaveSession(ctx, sess)
	if err != nil {
		return sess, h.sendInternalError(res, h.Logger, "failed to save session", err)
	}
	return savedSess, nil
}

// validateEmailFormat validates basic email format
func (h *AuthHandler) validateEmailFormat(email string) bool {
	return email != "" && strings.Contains(email, "@") && len(email) > 3
}

// validateRequiredFields validates that required string fields are not empty
func (h *AuthHandler) validateRequiredFields(fields map[string]string) []string {
	var missing []string
	for fieldName, value := range fields {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, fieldName)
		}
	}
	return missing
}

// Main handler struct =====================================================

// AuthHandler handles all authentication-related HTTP requests including
// login, logout, OAuth flows, and token management.
type AuthHandler struct {
	Logger       *telemetry.Logger
	SessionStore *session.Store
	AccountStore *account.Store
	OAuthStore   *oauth.Store
	OAuthService *oauth.Service
	Config       internal.Config
}

// Request/Response types ==================================================

// LoginForm represents the structure of a login form submission
type LoginForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	ReturnTo string `json:"return_to"`
}

// GrantRequestForm represents the structure of an OAuth grant request
type GrantRequestForm struct {
	Authorize bool `json:"authorize"`
}

// TokensResponse represents the OAuth token response format
type TokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

// Microsoft OAuth types ==================================================

// MicrosoftTokenResponse represents the response from Microsoft's token endpoint
type MicrosoftTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// MicrosoftUserProfile represents a user profile from Microsoft Graph API
type MicrosoftUserProfile struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName"`
	GivenName         string `json:"givenName"`
	Surname           string `json:"surname"`
	UserPrincipalName string `json:"userPrincipalName"`
	Mail              string `json:"mail"`
}

// MicrosoftIDTokenClaims represents the claims in a Microsoft ID token
type MicrosoftIDTokenClaims struct {
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`
	Name              string `json:"name"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Sub               string `json:"sub"`
	Aud               string `json:"aud"`
	Iss               string `json:"iss"`
	Iat               int64  `json:"iat"`
	Exp               int64  `json:"exp"`
}

// Constructor =============================================================

// NewAuthHandler creates a new AuthHandler with the provided dependencies
func NewAuthHandler(logger *telemetry.Logger, sessionStore *session.Store, accountStore *account.Store, oauthStore *oauth.Store, oauthService *oauth.Service, config internal.Config) *AuthHandler {
	return &AuthHandler{
		Logger:       logger,
		SessionStore: sessionStore,
		AccountStore: accountStore,
		OAuthStore:   oauthStore,
		OAuthService: oauthService,
		Config:       config,
	}
}

// Route registration ======================================================

// RegisterRoutes registers all authentication-related routes with the router
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
		group.GET("/microsoft", h.HandleMicrosoftOAuth, sessionMiddleware)
		group.GET("/microsoft/callback", h.HandleMicrosoftOAuthCallback, sessionMiddleware)
	}, noCacheMiddleware)
}

// Login pages and handlers ===============================================

// ShowLoginPage displays the login page or redirects if user is already authenticated.
// It handles the optional returnTo parameter for post-login redirection.
func (h *AuthHandler) ShowLoginPage(ctx context.Context, req *http.Request, res *http.Response) error {
	sess := ctx.Value(SessionContextKey).(session.Session)

	returnTo := req.URLQueryParam("returnTo")

	if sess.AccountID.IsSet() {
		// User already logged in, redirect to home or dashboard
		if returnTo != "" {
			return res.SendRedirect(http.StatusSeeOther, returnTo)
		}
		return res.SendRedirect(http.StatusSeeOther, defaultHomePath)
	}

	csrfToken := sess.Values[CSRFTokenSessionValueKey].(string)

	return ui.Render(ctx, res, page.Login(page.LoginProps{RootProps: layout.RootProps{
		Title:     "Login",
		CSRFToken: csrfToken,
	},
		ReturnTo: returnTo,
	}))
}

// Login handlers ===========================================================

// HandleLogin processes login form submissions with comprehensive validation.
// It validates credentials, creates user sessions, and handles post-login redirection.
func (h *AuthHandler) HandleLogin(ctx context.Context, req *http.Request, res *http.Response) error {
	// Parse login form
	var form LoginForm
	if err := req.DecodeJSON(&form); err != nil {
		h.Logger.Warn("failed to parse login form", "error", err)
		return h.sendBadRequestError(res, errInvalidRequest)
	}

	// Validate required fields
	missing := h.validateRequiredFields(map[string]string{
		"email":    form.Email,
		"password": form.Password,
	})
	if len(missing) > 0 {
		h.Logger.Warn("missing required login fields", "missing", missing)
		return h.sendBadRequestError(res, "missing required fields: "+strings.Join(missing, ", "))
	}

	// Validate email format
	if !h.validateEmailFormat(form.Email) {
		h.Logger.Warn("invalid email format", "email", form.Email)
		return h.sendBadRequestError(res, "invalid email format")
	}

	err := h.AccountStore.ValidateCredentials(ctx, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, account.ErrUserNotFound) || errors.Is(err, account.ErrInvalidCredentials) {
			return h.sendUnauthorizedError(res, errInvalidCredentials)
		}
		return h.sendInternalError(res, h.Logger, "failed to validate credentials", err)
	}

	user, err := h.AccountStore.GetUserByEmail(ctx, form.Email)
	if err != nil {
		return h.sendInternalError(res, h.Logger, "failed to get user by email", err)
	}

	sess := ctx.Value(SessionContextKey).(session.Session)
	sess.AccountID = util.Some(user.ID)

	sess, err = h.saveSessionWithError(ctx, sess, res)
	if err != nil {
		return err // Error already handled in helper
	}

	sess, err = h.SessionStore.RegenerateSession(ctx, sess)
	if err != nil {
		return h.sendInternalError(res, h.Logger, "failed to regenerate session", err)
	}

	// Set session cookie
	h.SessionStore.SetCookie(ctx, res, sess.ID)

	h.Logger.Info("user logged in", "user_id", user.ID)

	redirectTo := defaultHomePath
	if form.ReturnTo != "" {
		redirectTo = form.ReturnTo
	}

	return res.SendRedirect(http.StatusSeeOther, redirectTo)
}

// Logout handlers =========================================================

// HandleLogout processes logout requests and clears user sessions
func (h *AuthHandler) HandleLogout(ctx context.Context, req *http.Request, res *http.Response) error {
	sess := ctx.Value(SessionContextKey).(session.Session)

	err := h.SessionStore.DeleteSession(ctx, sess.ID)
	if err != nil {
		return h.sendInternalError(res, h.Logger, "failed to delete session", err)
	}

	h.Logger.Info("user logged out", "user_id", sess.AccountID)

	// Clear session cookie
	h.SessionStore.ClearCookie(ctx, res)

	return res.SendJSON(http.StatusOK, map[string]string{"message": "logged out"})
}

// OAuth authorization handlers ============================================

// HandleAuthorize handles OAuth authorization requests and user consent
func (h *AuthHandler) HandleAuthorize(ctx context.Context, req *http.Request, res *http.Response) error {
	responseType := req.URLQueryParam("response_type")
	clientIDStr := req.URLQueryParam("client_id")
	redirectURI := req.URLQueryParam("redirect_uri")
	scopeParam := req.URLQueryParam("scope")
	state := req.URLQueryParam("state")
	codeChallenge := req.URLQueryParam("code_challenge")
	codeChallengeMethod := req.URLQueryParam("code_challenge_method")

	scopes := strings.Fields(scopeParam)

	// Validate input
	if responseType == "" || clientIDStr == "" || redirectURI == "" {
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "missing required parameters"})
	}

	// Require PKCE parameters for we only support PKCE flow
	if codeChallenge == "" || codeChallengeMethod == "" {
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "missing pkce parameters"})
	}

	clientID, err := uuid.Parse(clientIDStr)
	if err != nil {
		h.Logger.Warn("invalid client ID format", "error", err)
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "invalid client ID"})
	}

	sess := ctx.Value(SessionContextKey).(session.Session)
	if !sess.AccountID.IsSet() {
		// User not logged in, redirect to login page
		return res.SendRedirect(http.StatusSeeOther, "/login?returnTo="+url.QueryEscape(req.URL()))
	}

	grantScopes := false
	authorizeVal, exists := sess.Values[AuthorizeSessionValueKey]
	if exists {
		authorize, ok := authorizeVal.(bool)
		if !ok {
			h.Logger.Error("invalid authorize session value type", "type", authorizeVal)
			return RedirectOAuthResponse(res, redirectURI, state, "server_error")
		}

		delete(sess.Values, AuthorizeSessionValueKey)
		sess, err = h.SessionStore.SaveSession(ctx, sess)
		if err != nil {
			h.Logger.Error("failed to save session", "error", err)
			return RedirectOAuthResponse(res, redirectURI, state, "server_error")
		}

		if !authorize {
			// User denied authorization
			return RedirectOAuthResponse(res, redirectURI, state, "access_denied")
		}

		grantScopes = true
	}

	authorizationCode, client, ungrantedScopes, err := h.OAuthService.IssueCode(ctx, responseType, clientID, sess.AccountID.V, redirectURI, scopes, codeChallenge, codeChallengeMethod, grantScopes)
	if err != nil {
		if errors.Is(err, oauth.ErrUnsupportedResponseType) {
			h.Logger.Error("unsupported response type", "error", err)
			return RedirectOAuthResponse(res, redirectURI, state, "unsupported_response_type")
		}

		if errors.Is(err, oauth.ErrInvalidClientCredentials) {
			h.Logger.Error("invalid client credentials", "error", err)
			return RedirectOAuthResponse(res, redirectURI, state, "invalid_client")
		}

		if errors.Is(err, oauth.ErrInvalidPKCEParameters) {
			h.Logger.Error("invalid pkce parameters", "error", err)
			return RedirectOAuthResponse(res, redirectURI, state, "invalid_request")
		}

		// Let the user authorize the requested scopes
		if errors.Is(err, oauth.ErrRequestedUngrantedScopes) {
			csrfToken, err := util.GenerateRandomString(32)
			if err != nil {
				h.Logger.Error("failed to generate CSRF token", "error", err)
				return RedirectOAuthResponse(res, redirectURI, state, "server_error")
			}

			sess.Values[CSRFTokenSessionValueKey] = csrfToken
			sess, err = h.SessionStore.SaveSession(ctx, sess)
			if err != nil {
				h.Logger.Error("failed to save session", "error", err)
				return RedirectOAuthResponse(res, redirectURI, state, "server_error")
			}

			return ui.Render(ctx, res, page.Authorize(page.AuthorizeProps{RootProps: layout.RootProps{
				Title:     "Authorize",
				CSRFToken: csrfToken,
			},
				ApplicationName: client.Name,
				Scopes:          ungrantedScopes,
			}))
		}

		h.Logger.Error("Failed to issue authorization code", "error", err)
		return RedirectOAuthResponse(res, redirectURI, state, "server_error")
	}

	// Build redirect URL
	finalRedirectURI := redirectURI + "?code=" + url.QueryEscape(authorizationCode.Code)
	if state != "" {
		finalRedirectURI += "&state=" + url.QueryEscape(state)
	}

	h.Logger.Info("authorization granted", "user_id", sess.AccountID.V, "client_id", clientID, "scopes", scopes)
	return RedirectOAuthResponse(res, finalRedirectURI, state, "")
}

// OAuth helper functions ==================================================

// RedirectOAuthResponse handles OAuth redirect responses with proper error handling
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

// Helper functions for OAuth error formatting ============================

// JSONOAuthErrorResponse sends a standardized OAuth error response in JSON format
func JSONOAuthErrorResponse(res *http.Response, code http.StatusCode, status string, description string) error {
	return res.SendJSON(code, map[string]any{
		"error":             status,
		"error_description": description,
	})
}

// Token handlers ==========================================================

// HandleToken processes OAuth token requests
func (h *AuthHandler) HandleToken(ctx context.Context, req *http.Request, res *http.Response) error {
	h.Logger.Info("handling token request")

	grantType := req.FormValue("grant_type")
	clientIDStr := req.FormValue("client_id")
	clientSecret := req.FormValue("client_secret")
	scope := req.FormValue("scope")
	redirectURI := req.FormValue("redirect_uri")
	code := req.FormValue("code")
	codeVerifier := req.FormValue("code_verifier")
	refreshToken := req.FormValue("refresh_token")

	authHeader := req.Header("Authorization")
	if authHeader != "" {
		// Handle authorization header
		if after, ok := strings.CutPrefix(authHeader, "Basic "); ok {
			// Decode client credentials
			credentials, err := base64.StdEncoding.DecodeString(after)
			if err != nil {
				h.Logger.Warn("failed to decode authorization header", "error", err)
				return res.SendJSON(http.StatusUnauthorized, map[string]string{"error": "invalid client credentials"})
			}

			parts := strings.SplitN(string(credentials), ":", 2)
			if len(parts) != 2 {
				h.Logger.Warn("invalid authorization header format")
				return res.SendJSON(http.StatusUnauthorized, map[string]string{"error": "invalid client credentials"})
			}

			clientIDStr = parts[0]
			clientSecret = parts[1]
		}
	}

	clientID, err := uuid.Parse(clientIDStr)
	if err != nil {
		h.Logger.Warn("invalid client ID format", "error", err)
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "invalid client ID"})
	}
	scopes := strings.Fields(scope)

	tokens, err := h.OAuthService.IssueTokens(ctx, grantType, clientID, clientSecret, scopes, codeVerifier, code, redirectURI, refreshToken)
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

// Grant handlers ==========================================================

// HandleGrant processes OAuth authorization grant requests
func (h *AuthHandler) HandleGrant(ctx context.Context, req *http.Request, res *http.Response) error {
	var form GrantRequestForm
	if err := req.DecodeJSON(&form); err != nil {
		h.Logger.Warn("failed to parse grant request form", "error", err)
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	sess := ctx.Value(SessionContextKey).(session.Session)
	if !sess.AccountID.IsSet() {
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

	sess, err := h.SessionStore.SaveSession(ctx, sess)
	if err != nil {
		h.Logger.Error("failed to save session", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return res.SendJSON(http.StatusOK, map[string]string{"message": "grant processed"})
}

// Introspection handlers ==================================================

// HandleIntrospection handles OAuth token introspection requests
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

// Microsoft OAuth handlers ===============================================

// HandleMicrosoftOAuth initiates the Microsoft OAuth flow
func (h *AuthHandler) HandleMicrosoftOAuth(ctx context.Context, req *http.Request, res *http.Response) error {
	sess := ctx.Value(SessionContextKey).(session.Session)

	// Generate state for CSRF protection
	state, err := util.GenerateRandomString(32)
	if err != nil {
		h.Logger.Error("failed to generate OAuth state", "error", err)
		return h.redirectToLogin(res, "internal_error", "")
	}

	sess.Values[microsoftOAuthStateKey] = state

	// Store return URL if provided
	returnTo := req.URLQueryParam("returnTo")
	if returnTo != "" {
		sess.Values[microsoftOAuthReturnKey] = returnTo
	}

	// Save session
	if _, err := h.saveSessionWithError(ctx, sess, res); err != nil {
		return h.redirectToLogin(res, errSessionError, "")
	}

	// Build Microsoft OAuth authorization URL
	authURL := fmt.Sprintf("%s/%s/oauth2/v2.0/authorize", microsoftAuthBaseURL, h.Config.OAuth.Microsoft.TenantID)

	params := url.Values{}
	params.Set("client_id", h.Config.OAuth.Microsoft.ClientID)
	params.Set("response_type", "code")
	params.Set("redirect_uri", h.Config.OAuth.Microsoft.RedirectURI)
	params.Set("scope", microsoftOAuthScopes)
	params.Set("state", state)
	params.Set("response_mode", "query")

	fullAuthURL := authURL + "?" + params.Encode()

	return res.SendRedirect(http.StatusSeeOther, fullAuthURL)
}

// HandleMicrosoftOAuthCallback handles the callback from Microsoft OAuth
func (h *AuthHandler) HandleMicrosoftOAuthCallback(ctx context.Context, req *http.Request, res *http.Response) error {
	sess := ctx.Value(SessionContextKey).(session.Session)

	code := req.URLQueryParam("code")
	state := req.URLQueryParam("state")
	errorParam := req.URLQueryParam("error")

	// Check for OAuth errors
	if errorParam != "" {
		h.Logger.Warn("Microsoft OAuth returned error", "error", errorParam)
		return h.redirectToLogin(res, "oauth_denied", "")
	}

	if code == "" {
		h.Logger.Warn("missing code in microsoft oauth callback")
		return h.redirectToLogin(res, "oauth_error", "")
	}

	// Verify state parameter for CSRF protection
	expectedState, exists := sess.Values[microsoftOAuthStateKey]
	if !exists || expectedState != state {
		h.Logger.Warn("invalid state in microsoft oauth callback", "expected", expectedState, "received", state)
		return h.redirectToLogin(res, "invalid_state", "")
	}

	// Exchange authorization code for access token
	tokenResp, err := h.exchangeMicrosoftCodeForToken(code)
	if err != nil {
		h.Logger.Error("failed to exchange code for token", "error", err)
		return h.redirectToLogin(res, errTokenExchangeFailed, "")
	}

	// Extract user email from ID token instead of calling /me endpoint
	var email string
	var displayName string

	if tokenResp.IDToken != "" {
		claims, err := h.parseIDTokenClaims(tokenResp.IDToken)
		if err != nil {
			h.Logger.Error("failed to parse ID token claims", "error", err)
			return h.redirectToLogin(res, "token_parse_failed", "")
		}

		h.Logger.Debug("Microsoft ID token claims", "claims", claims)

		// Use email claim, fallback to preferred_username if email is not available
		email = claims.Email
		if email == "" {
			email = claims.PreferredUsername
		}
		displayName = claims.Name
	}

	// If no ID token or email extraction failed, show error
	if email == "" {
		h.Logger.Error("no email found in Microsoft ID token")
		return h.redirectToLogin(res, "no_email_found", "")
	}

	// Create or get user account
	user, err := h.createOrUpdateMicrosoftUser(ctx, email, displayName)
	if err != nil {
		h.Logger.Error("failed to create or update user", "error", err)
		return h.redirectToLogin(res, "user_creation_failed", "")
	}

	// Create session
	sess.AccountID = util.Some(user.ID)
	sess.Values[microsoftOAuthStateKey] = nil // Clear the state

	// Save session
	if _, err := h.saveSessionWithError(ctx, sess, res); err != nil {
		return h.redirectToLogin(res, errSessionError, "")
	}

	// Set session cookie
	h.SessionStore.SetCookie(ctx, res, sess.ID)

	// Get return URL from session or default to home
	returnTo := defaultHomePath
	if returnToValue, exists := sess.Values[microsoftOAuthReturnKey]; exists && returnToValue != nil {
		if returnToStr, ok := returnToValue.(string); ok && returnToStr != "" {
			returnTo = returnToStr
		}
	}

	// Clear the return URL from session
	sess.Values[microsoftOAuthReturnKey] = nil
	if _, err := h.SessionStore.SaveSession(ctx, sess); err != nil {
		h.Logger.Warn("failed to clear return URL from session", "error", err)
	}

	h.Logger.Info("Microsoft OAuth login successful", "user_id", user.ID, "email", user.Email)
	return res.SendRedirect(http.StatusSeeOther, returnTo)
}

// Microsoft OAuth helper functions =======================================

// parseIDTokenClaims parses a JWT ID token without signature verification
// Note: This is for extracting user claims only, not for security validation
func (h *AuthHandler) parseIDTokenClaims(idToken string) (*MicrosoftIDTokenClaims, error) {
	// Split the JWT into its three parts
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}

	// Decode the payload (second part)
	payload := parts[1]

	// Add padding if necessary for base64 decoding
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	payloadBytes, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var claims MicrosoftIDTokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWT claims: %w", err)
	}

	return &claims, nil
}

// exchangeMicrosoftCodeForToken exchanges the authorization code for an access token
func (h *AuthHandler) exchangeMicrosoftCodeForToken(code string) (*MicrosoftTokenResponse, error) {
	tokenURL := fmt.Sprintf("%s/%s/oauth2/v2.0/token", microsoftAuthBaseURL, h.Config.OAuth.Microsoft.TenantID)

	data := url.Values{}
	data.Set("client_id", h.Config.OAuth.Microsoft.ClientID)
	data.Set("client_secret", h.Config.OAuth.Microsoft.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", h.Config.OAuth.Microsoft.RedirectURI)

	client := http.NewClient(tokenRequestTimeout)
	resp, err := client.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to make token request: %w", err)
	}
	defer resp.Body.Close()

	if http.StatusCode(resp.StatusCode) != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp MicrosoftTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// createOrUpdateMicrosoftUser creates a new user or updates an existing one based on Microsoft email
func (h *AuthHandler) createOrUpdateMicrosoftUser(ctx context.Context, email, displayName string) (*account.User, error) {
	if email == "" {
		return nil, fmt.Errorf("no email provided for Microsoft user")
	}

	// Check if user already exists
	existingUser, err := h.AccountStore.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, account.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check for existing user: %w", err)
	}

	if err == nil {
		// User exists, update if needed
		updated := false

		// Mark email as verified since it's coming from Microsoft
		if !existingUser.IsEmailVerified {
			existingUser.IsEmailVerified = true
			updated = true
		}

		if updated {
			if updatedUser, err := h.AccountStore.SaveUser(ctx, existingUser); err != nil {
				return nil, fmt.Errorf("failed to update existing user: %w", err)
			} else {
				return &updatedUser, nil
			}
		}

		return &existingUser, nil
	}

	// Create new user
	newUser, err := h.AccountStore.NewUser(email, "", account.RoleUser) // Pre-hash password placeholder
	if err != nil {
		return nil, fmt.Errorf("failed to create new user object: %w", err)
	}

	newUser.IsEmailVerified = true // Mark email as verified since it's from Microsoft

	savedUser, err := h.AccountStore.SaveUser(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	return &savedUser, nil
}
