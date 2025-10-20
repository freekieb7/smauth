package web

import (
	"context"
	"errors"
	"log/slog"

	"github.com/freekieb7/smauth/internal/auth"
	"github.com/freekieb7/smauth/internal/http"
	"github.com/freekieb7/smauth/internal/oauth"
	"github.com/freekieb7/smauth/internal/session"
	"github.com/freekieb7/smauth/internal/util"
	"github.com/google/uuid"
)

type ContextKey string

const SessionContextKey ContextKey = "session"
const CSRFTokenSessionValueKey session.ValueKey = "csrf_token"
const AuthorizeSessionValueKey session.ValueKey = "authorize"
const ScopesSessionValueKey session.ValueKey = "scopes"

type APIHandler struct {
	Logger       *slog.Logger
	SessionStore *session.Store
	AuthStore    *auth.Store
	OAuthStore   *oauth.Store
	OAuthIssuer  *oauth.Issuer
}

func NewAPIHandler(logger *slog.Logger, sessionStore *session.Store, authStore *auth.Store, oauthStore *oauth.Store, oauthIssuer *oauth.Issuer) *APIHandler {
	return &APIHandler{
		Logger:       logger,
		SessionStore: sessionStore,
		AuthStore:    authStore,
		OAuthStore:   oauthStore,
		OAuthIssuer:  oauthIssuer,
	}
}

func (h *APIHandler) RegisterRoutes(router *http.Router) {
	noCacheMiddleware := NoCacheMiddleware()

	router.GET("/health", h.HealthCheck, noCacheMiddleware)

	router.Group("/api/users", func(group *http.Router) {
		group.POST("", h.HandleCreateUser)
		group.Group("/{user_id}", func(group *http.Router) {
			group.POST("/permissions", h.handleCreateUserPermissions)
		})
	})

	router.Group("/api/resource_servers", func(group *http.Router) {
		group.POST("", h.HandleCreateResourceServer)
	})

	router.Group("/api/clients", func(group *http.Router) {
		group.POST("", h.HandleCreateClient)
		group.Group("/{client_id}", func(group *http.Router) {
			group.POST("/permissions", h.handleCreateClientPermissions)
		})
	})
}

func (h *APIHandler) HealthCheck(ctx context.Context, req *http.Request, res *http.Response) error {
	return res.SendJSON(http.StatusOK, map[string]string{"status": "ok"})
}

type CreateUserForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *APIHandler) HandleCreateUser(ctx context.Context, req *http.Request, res *http.Response) error {
	var form CreateUserForm
	if err := req.DecodeJSON(&form); err != nil {
		h.Logger.Warn("failed to parse user creation form", "error", err)
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// Validate input
	if form.Email == "" || form.Password == "" {
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "email and password are required"})
	}

	// Check if user already exists
	_, err := h.AuthStore.GetUserByEmail(ctx, form.Email)
	if err == nil {
		return res.SendJSON(http.StatusConflict, map[string]string{"error": "user with this email already exists"})
	}
	if err != auth.ErrUserNotFound {
		h.Logger.Error("failed to check existing user", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	// Create new user
	newUser, err := h.AuthStore.NewUser(form.Email, form.Password, auth.RoleUser)
	if err != nil {
		h.Logger.Error("failed to create user", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	createdUser, err := h.AuthStore.SaveUser(ctx, newUser)
	if err != nil {
		h.Logger.Error("failed to create user", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return res.SendJSON(http.StatusCreated, map[string]string{"user_id": createdUser.ID.String()})
}

type CreateResourceServerForm struct {
	URL    string `json:"url"`
	Scopes []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"scopes"`
}

func (h *APIHandler) HandleCreateResourceServer(ctx context.Context, req *http.Request, res *http.Response) error {
	var form CreateResourceServerForm
	if err := req.DecodeJSON(&form); err != nil {
		h.Logger.Warn("failed to parse resource server creation form", "error", err)
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// Validate input
	if form.URL == "" {
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "resource server URL is required"})
	}

	_, err := h.OAuthStore.GetResourceServerByURL(ctx, form.URL)
	if err == nil {
		return res.SendJSON(http.StatusConflict, map[string]string{"error": "resource server with this URL already exists"})
	}
	if err != oauth.ErrResourceServerNotFound {
		h.Logger.Error("failed to check existing resource server", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	resourceServer, err := h.OAuthStore.NewResourceServer(form.URL)
	if err != nil {
		h.Logger.Error("failed to create resource server", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	resourceServer.Scopes = make([]oauth.Scope, len(form.Scopes))
	for i, scopeForm := range form.Scopes {
		scope, err := h.OAuthStore.NewScope(resourceServer.URL, scopeForm.Name, scopeForm.Description)
		if err != nil {
			h.Logger.Error("failed to create scope", "error", err)
			return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}

		resourceServer.Scopes[i] = scope
	}

	resourceServer, err = h.OAuthStore.SaveResourceServer(ctx, resourceServer)
	if err != nil {
		h.Logger.Error("failed to save resource server", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return res.SendJSON(http.StatusCreated, map[string]string{
		"resource_server_id": resourceServer.ID.String(),
	})
}

type CreateClientForm struct {
	Name         string   `json:"name"`
	RedirectURIs []string `json:"redirect_uris"`
}

func (h *APIHandler) HandleCreateClient(ctx context.Context, req *http.Request, res *http.Response) error {
	var form CreateClientForm
	if err := req.DecodeJSON(&form); err != nil {
		h.Logger.Warn("failed to parse client creation form", "error", err)
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// Validate input
	if form.Name == "" {
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "client name is required"})
	}

	if len(form.RedirectURIs) == 0 {
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "redirect URIs are required"})
	}

	client, err := h.OAuthStore.NewClient(form.Name, form.RedirectURIs)
	if err != nil {
		h.Logger.Error("failed to create client", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	client, err = h.OAuthStore.SaveClient(ctx, client)
	if err != nil {
		h.Logger.Error("failed to save client", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return res.SendJSON(http.StatusCreated, map[string]string{
		"client_id":     client.ID.String(),
		"client_secret": client.Secret,
	})
}

type CreatePermissionsForm struct {
	Permissions []string `json:"permissions"`
}

func (h *APIHandler) handleCreateClientPermissions(ctx context.Context, req *http.Request, res *http.Response) error {
	clientIDStr := req.URLPathValue("client_id")
	clientID, err := uuid.Parse(clientIDStr)
	if err != nil {
		h.Logger.Warn("invalid client ID format", "error", err)
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "invalid client ID"})
	}

	var form CreatePermissionsForm
	if err := req.DecodeJSON(&form); err != nil {
		h.Logger.Warn("failed to parse permissions form", "error", err)
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	client, err := h.OAuthStore.GetClient(ctx, clientID)
	if err != nil {
		if errors.Is(err, oauth.ErrClientNotFound) {
			return res.SendJSON(http.StatusNotFound, map[string]string{"error": "client not found"})
		}

		h.Logger.Error("failed to get client", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	if err := h.OAuthStore.AssignPermissions(ctx, util.Some(client.ID), util.None[uuid.UUID](), form.Permissions); err != nil {
		if errors.Is(err, oauth.ErrInvalidPermissionAssignment) {
			return res.SendJSON(http.StatusNotFound, map[string]string{"error": "invalid permission assignment"})
		}

		h.Logger.Error("failed to assign client permission", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return res.SendJSON(http.StatusCreated, map[string]string{"message": "client permissions created"})
}

func (h *APIHandler) handleCreateUserPermissions(ctx context.Context, req *http.Request, res *http.Response) error {
	userIDStr := req.URLPathValue("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.Logger.Warn("invalid user ID format", "error", err)
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
	}

	var form CreatePermissionsForm
	if err := req.DecodeJSON(&form); err != nil {
		h.Logger.Warn("failed to parse permissions form", "error", err)
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	user, err := h.AuthStore.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return res.SendJSON(http.StatusNotFound, map[string]string{"error": "user not found"})
		}

		h.Logger.Error("failed to get user", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	if err := h.OAuthStore.AssignPermissions(ctx, util.None[uuid.UUID](), util.Some(user.ID), form.Permissions); err != nil {
		if errors.Is(err, oauth.ErrInvalidPermissionAssignment) {
			return res.SendJSON(http.StatusNotFound, map[string]string{"error": "invalid permission assignment"})
		}

		h.Logger.Error("failed to assign user permission", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return res.SendJSON(http.StatusCreated, map[string]string{"message": "user permissions created"})
}
