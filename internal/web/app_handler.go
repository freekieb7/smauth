package web

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/freekieb7/smauth/internal/account"
	"github.com/freekieb7/smauth/internal/audit"
	"github.com/freekieb7/smauth/internal/http"
	"github.com/freekieb7/smauth/internal/oauth"
	"github.com/freekieb7/smauth/internal/session"
	"github.com/freekieb7/smauth/internal/telemetry"
	"github.com/freekieb7/smauth/internal/util"
	"github.com/freekieb7/smauth/internal/web/ui"
	"github.com/freekieb7/smauth/internal/web/ui/template/component"
	"github.com/freekieb7/smauth/internal/web/ui/template/layout"
	"github.com/freekieb7/smauth/internal/web/ui/template/page"
	"github.com/google/uuid"
)

// Constants ================================================================

const (
	// Default pagination settings
	defaultPageSize = 25
	maxPageSize     = 100

	// Admin paths
	adminUsersPath           = "/admin/users"
	adminClientsPath         = "/admin/clients"
	adminResourceServersPath = "/admin/resource-servers"
	adminAuditLogsPath       = "/admin/audit-logs"

	// Error codes for users
	errUserInvalidForm   = "invalid_form"
	errUserEmailRequired = "email_required"
	errUserRoleRequired  = "role_required"
	errUserInvalidRole   = "invalid_role"
	errUserExists        = "user_exists"
	errUserInternalError = "internal_error"

	// Error codes for clients
	errClientInvalidForm         = "invalid_form"
	errClientNameRequired        = "name_required"
	errClientRedirectURIRequired = "redirect_uris_required"
	errClientInvalidURI          = "invalid_uri"
	errClientExists              = "client_exists"
	errClientInternalError       = "internal_error"
	errClientIDRequired          = "client_id_required"
	errClientInvalidID           = "invalid_client_id"
	errClientNotFound            = "client_not_found"
	errClientDeleteFailed        = "delete_failed"

	// Error codes for resource servers
	errResourceServerInvalidForm   = "invalid_form"
	errResourceServerNameRequired  = "name_required"
	errResourceServerURLRequired   = "url_required"
	errResourceServerInvalidURL    = "invalid_url"
	errResourceServerExists        = "resource_server_exists"
	errResourceServerInternalError = "internal_error"
	errResourceServerIDRequired    = "id_required"
	errResourceServerInvalidID     = "invalid_id"
	errResourceServerNotFound      = "not_found"
	errResourceServerDeleteFailed  = "delete_failed"

	// Error codes for scopes
	errScopeNameRequired  = "scope_name_required"
	errScopeInvalidName   = "invalid_scope_name"
	errScopeExists        = "scope_exists"
	errScopeInternalError = "scope_internal_error"
	errScopeNotFound      = "scope_not_found"
	errScopeDeleteFailed  = "scope_delete_failed"

	// Success codes
	successUserCreated           = "user_created"
	successClientCreated         = "client_created"
	successClientDeleted         = "client_deleted"
	successResourceServerCreated = "resource_server_created"
	successResourceServerDeleted = "resource_server_deleted"
	successScopeCreated          = "scope_created"
	successScopeUpdated          = "scope_updated"
	successScopeDeleted          = "scope_deleted"
)

// Main handler struct =====================================================

// AppHandler handles all application-related HTTP requests including
// dashboard, admin panels, user management, client management, and audit logs.
type AppHandler struct {
	Logger                *telemetry.Logger
	SessionStore          *session.Store
	AccountStore          *account.Store
	AuditStore            *audit.Store
	AuditService          *audit.Service
	OAuthStore            *oauth.Store
	ClientService         *account.ClientService
	UserService           *account.UserService
	ResourceServerService *oauth.ResourceServerService
}

// Constructor =============================================================

// NewAppHandler creates a new AppHandler with the provided dependencies
func NewAppHandler(logger *telemetry.Logger, sessionStore *session.Store, accountStore *account.Store, auditStore *audit.Store, auditService *audit.Service, oauthStore *oauth.Store, clientService *account.ClientService, userService *account.UserService, resourceServerService *oauth.ResourceServerService) AppHandler {
	return AppHandler{
		Logger:                logger,
		SessionStore:          sessionStore,
		AccountStore:          accountStore,
		AuditStore:            auditStore,
		AuditService:          auditService,
		OAuthStore:            oauthStore,
		ClientService:         clientService,
		UserService:           userService,
		ResourceServerService: resourceServerService,
	}
}

// Common helper functions =================================================

// redirectWithError redirects to a path with an error code parameter
func (h *AppHandler) redirectWithError(res *http.Response, path, errorCode string) error {
	redirectURL := fmt.Sprintf("%s?error=%s", path, url.QueryEscape(errorCode))
	return res.SendRedirect(http.StatusSeeOther, redirectURL)
}

// redirectWithSuccess redirects to a path with a success code parameter
func (h *AppHandler) redirectWithSuccess(res *http.Response, path, successCode string) error {
	redirectURL := fmt.Sprintf("%s?success=%s", path, url.QueryEscape(successCode))
	return res.SendRedirect(http.StatusSeeOther, redirectURL)
}

// redirectWithClientInfo redirects to clients page with success and client details
func (h *AppHandler) redirectWithClientInfo(res *http.Response, clientID, clientSecret, clientName string) error {
	redirectURL := fmt.Sprintf("%s?success=%s&client_id=%s&client_secret=%s&client_name=%s",
		adminClientsPath,
		url.QueryEscape(successClientCreated),
		url.QueryEscape(clientID),
		url.QueryEscape(clientSecret),
		url.QueryEscape(clientName))
	return res.SendRedirect(http.StatusSeeOther, redirectURL)
}

// redirectWithResourceServerInfo redirects to resource servers page with success and resource server details
func (h *AppHandler) redirectWithResourceServerInfo(res *http.Response, resourceServerURL, resourceServerID string) error {
	redirectURL := fmt.Sprintf("%s?success=%s&resource_server_url=%s&resource_server_id=%s",
		adminResourceServersPath,
		url.QueryEscape(successResourceServerCreated),
		url.QueryEscape(resourceServerURL),
		url.QueryEscape(resourceServerID))
	return res.SendRedirect(http.StatusSeeOther, redirectURL)
}

// validateRequiredFields validates that required form fields are not empty
func (h *AppHandler) validateRequiredFields(fields map[string]string) []string {
	var missing []string
	for fieldName, value := range fields {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, fieldName)
		}
	}
	return missing
}

// parseFormOrRedirect parses form data and redirects with error if parsing fails
func (h *AppHandler) parseFormOrRedirect(req *http.Request, res *http.Response, redirectPath string) error {
	if err := req.Request.ParseForm(); err != nil {
		h.Logger.Warn("Failed to parse form", "error", err)
		return h.redirectWithError(res, redirectPath, errUserInvalidForm)
	}
	return nil
}

// validateUserRole validates that the provided role is valid
func (h *AppHandler) validateUserRole(role account.UserRole) bool {
	return role == account.RoleAdmin || role == account.RoleUser
}

// parsePaginationParams extracts and validates pagination parameters from request
func (h *AppHandler) parsePaginationParams(req *http.Request) (pageSize int, token, direction string) {
	pageSize = defaultPageSize
	if pageSizeParam := req.URLQueryParam("page_size"); pageSizeParam != "" {
		if ps, err := strconv.Atoi(pageSizeParam); err == nil && ps > 0 && ps <= maxPageSize {
			pageSize = ps
		}
	}

	token = req.URLQueryParam("token")
	direction = req.URLQueryParam("direction")
	if direction == "" {
		direction = "next"
	}

	return pageSize, token, direction
}

// Route registration ======================================================

// RegisterRoutes registers all application-related routes with the router
func (h *AppHandler) RegisterRoutes(router *http.Router) {
	router.GET("/", func(ctx context.Context, req *http.Request, res *http.Response) error {
		return res.SendRedirect(http.StatusSeeOther, "/dashboard")
	})

	router.Group("", func(group *http.Router) {
		group.GET("/dashboard", h.ShowDashboard)

		group.Group("/admin", func(group *http.Router) {
			group.Group("/users", func(group *http.Router) {
				group.GET("", h.ShowUsers)
				group.POST("/new", h.NewUser)
			})
			group.Group("/clients", func(group *http.Router) {
				group.GET("", h.ShowClients)
				group.POST("/new", h.NewClient)
				group.POST("/delete", h.DeleteClient)
			})
			group.Group("/resource-servers", func(group *http.Router) {
				group.GET("", h.ShowResourceServers)
				group.POST("/new", h.NewResourceServer)
				group.POST("/delete", h.DeleteResourceServer)
				group.GET("/{id}/edit", h.ShowEditResourceServer)
				group.POST("/{id}/scopes", h.CreateScope)
				group.POST("/{id}/scopes/edit", h.UpdateScope)
				group.POST("/{id}/scopes/delete", h.DeleteScope)
			})

			group.POST("/signoff", h.SignOff)
			group.GET("/audit-logs", h.ShowAuditLogs)
		}, AdminOnlyMiddleware(h.Logger, h.AccountStore))
	}, SessionMiddleware(h.Logger, h.SessionStore), CSRFMiddleware(h.Logger, h.SessionStore), PageAuthenticatedMiddleware(h.Logger))
}

// Dashboard handlers ======================================================

// ShowDashboard displays the main dashboard page for authenticated users
func (h *AppHandler) ShowDashboard(ctx context.Context, req *http.Request, res *http.Response) error {
	sess := ctx.Value(SessionContextKey).(session.Session)

	// Get the user account - this should be available since we're using authenticated middleware
	user, err := h.AccountStore.GetUser(ctx, sess.AccountID.V)
	if err != nil {
		h.Logger.Error("failed to get user for dashboard", "error", err, "user_id", sess.AccountID.V)
		return res.SendRedirect(http.StatusSeeOther, "/login")
	}

	appProps, err := h.LayoutAppProps(ctx, "Dashboard")
	if err != nil {
		h.Logger.Error("failed to get app props for dashboard", "error", err)
		return res.Send(http.StatusInternalServerError)
	}

	return ui.Render(ctx, res, page.Dashboard(page.DashboardProps{
		AppProps: appProps,
		User:     user,
	}))
}

// User management handlers ================================================

// ShowUsers displays the user management page with pagination and handles success/error messages
func (h *AppHandler) ShowUsers(ctx context.Context, req *http.Request, res *http.Response) error {
	sess := ctx.Value(SessionContextKey).(session.Session)

	h.AuditService.LogEvent(ctx, audit.LogEventRequest{
		Event:     audit.EventDataRead,
		AccountID: sess.AccountID.V,
		Resource:  "users",
		Action:    "view",
		Success:   true,
		IPAddress: req.RemoteAddr(),
		UserAgent: req.UserAgent(),
	})

	// Parse pagination parameters using helper
	pageSize, token, direction := h.parsePaginationParams(req)

	// Use the new paginated method
	listReq := account.ListUsersRequest{
		PageSize:  pageSize,
		Token:     token,
		Direction: direction,
	}

	result, err := h.UserService.ListUsersPaginated(ctx, listReq)
	if err != nil {
		h.Logger.Warn("Failed to list users", "error", err)
		return err
	}

	// Build pagination info
	pagination := component.PaginationInfo{
		PageSize:      pageSize,
		HasNext:       result.NextToken.IsSet(),
		HasPrev:       result.PrevToken.IsSet(),
		NextPageToken: result.NextToken.UnwrapOr(""),
		PrevPageToken: result.PrevToken.UnwrapOr(""),
		BasePath:      adminUsersPath,
	}

	// Check for success/error messages in query parameters
	successMsg := req.URLQueryParam("success")
	errorMsg := req.URLQueryParam("error")

	appProps, err := h.LayoutAppProps(ctx, "Users")
	if err != nil {
		h.Logger.Error("failed to get app props for users", "error", err)
		return res.Send(http.StatusInternalServerError)
	}

	props := page.UsersProps{
		AppProps:       appProps,
		Users:          result.Users,
		SuccessMessage: successMsg,
		ErrorMessage:   errorMsg,
		Pagination:     pagination,
	}

	return ui.Render(ctx, res, page.Users(props))
}

// NewUser creates a new user account with comprehensive validation and error handling
func (h *AppHandler) NewUser(ctx context.Context, req *http.Request, res *http.Response) error {
	// Parse form data
	if err := h.parseFormOrRedirect(req, res, adminUsersPath); err != nil {
		return err
	}

	// Extract form values
	email := req.FormValue("email")
	roleStr := req.FormValue("role")

	// Validate required fields
	missing := h.validateRequiredFields(map[string]string{
		"email": email,
		"role":  roleStr,
	})

	if len(missing) > 0 {
		for _, field := range missing {
			switch field {
			case "email":
				h.Logger.Warn("Email is required")
				return h.redirectWithError(res, adminUsersPath, errUserEmailRequired)
			case "role":
				h.Logger.Warn("Role is required")
				return h.redirectWithError(res, adminUsersPath, errUserRoleRequired)
			}
		}
	}

	role := account.UserRole(roleStr)
	// Validate role
	if !h.validateUserRole(role) {
		h.Logger.Warn("Invalid role", "role", role)
		return h.redirectWithError(res, adminUsersPath, errUserInvalidRole)
	}

	randomPassword, err := util.GenerateRandomString(32)
	if err != nil {
		h.Logger.Error("Failed to generate random password", "error", err)
		return h.redirectWithError(res, adminUsersPath, errUserInternalError)
	}

	user, err := h.AccountStore.NewUser(email, randomPassword, role)
	if err != nil {
		h.Logger.Error("Failed to create new user object", "error", err)
		return h.redirectWithError(res, adminUsersPath, errUserInternalError)
	}

	// Create new user
	_, err = h.AccountStore.SaveUser(ctx, user)
	if err != nil {
		h.Logger.Error("Failed to create user", "error", err)
		// Check for specific errors like duplicate email
		return h.redirectWithError(res, adminUsersPath, errUserExists)
	}

	// Log the action
	sess := ctx.Value(SessionContextKey).(session.Session)
	h.AuditService.LogEvent(ctx, audit.LogEventRequest{
		Event:     audit.EventDataWrite,
		AccountID: sess.AccountID.V,
		Resource:  "users",
		Action:    "create",
		Success:   true,
		IPAddress: req.RemoteAddr(),
		UserAgent: req.UserAgent(),
		Details: map[string]interface{}{
			"created_user_email": email,
			"created_user_role":  role,
		},
	})

	h.Logger.Info("User created successfully", "email", email, "role", role)
	return h.redirectWithSuccess(res, adminUsersPath, successUserCreated)
}

// Client management handlers ==============================================

// ShowClients displays the OAuth client management page with success/error messages
func (h *AppHandler) ShowClients(ctx context.Context, req *http.Request, res *http.Response) error {
	sess := ctx.Value(SessionContextKey).(session.Session)

	h.AuditService.LogEvent(ctx, audit.LogEventRequest{
		Event:     audit.EventDataRead,
		AccountID: sess.AccountID.V,
		Resource:  "clients",
		Action:    "view",
		Success:   true,
		IPAddress: req.RemoteAddr(),
		UserAgent: req.UserAgent(),
	})

	// Parse pagination parameters using helper
	pageSize, token, _ := h.parsePaginationParams(req) // direction not used for clients

	// Use the new paginated method
	listReq := account.ListClientsRequest{
		PageSize: pageSize,
		Token:    token,
	}

	result, err := h.ClientService.ListClientsPaginated(ctx, listReq)
	if err != nil {
		h.Logger.Warn("Failed to list clients", "error", err)
		return err
	}

	// Build pagination info
	pagination := component.PaginationInfo{
		PageSize:      pageSize,
		HasNext:       result.NextToken.E,
		HasPrev:       result.PrevToken.E,
		NextPageToken: result.NextToken.V,
		PrevPageToken: result.PrevToken.V,
		BasePath:      adminClientsPath,
	}

	// Check for success/error messages in query parameters
	successMsg := req.URLQueryParam("success")
	errorMsg := req.URLQueryParam("error")
	clientID := req.URLQueryParam("client_id")
	clientName := req.URLQueryParam("client_name")
	clientSecret := req.URLQueryParam("client_secret")

	appProps, err := h.LayoutAppProps(ctx, "Clients")
	if err != nil {
		h.Logger.Error("failed to get app props for clients", "error", err)
		return res.Send(http.StatusInternalServerError)
	}

	props := page.ClientsProps{
		AppProps:       appProps,
		Clients:        result.Clients,
		SuccessMessage: successMsg,
		ErrorMessage:   errorMsg,
		ClientID:       clientID,
		ClientName:     clientName,
		ClientSecret:   clientSecret,
		Pagination:     pagination,
	}

	return ui.Render(ctx, res, page.Clients(props))
}

// NewClient creates a new OAuth client with comprehensive validation and error handling
func (h *AppHandler) NewClient(ctx context.Context, req *http.Request, res *http.Response) error {
	// Parse form data
	if err := h.parseFormOrRedirect(req, res, adminClientsPath); err != nil {
		return err
	}

	// Extract form values
	name := req.FormValue("name")
	redirectURIsStr := req.FormValue("redirect_uris")

	// Validate required fields
	missing := h.validateRequiredFields(map[string]string{
		"name":          name,
		"redirect_uris": redirectURIsStr,
	})

	if len(missing) > 0 {
		for _, field := range missing {
			switch field {
			case "name":
				h.Logger.Warn("Name is required")
				return h.redirectWithError(res, adminClientsPath, errClientNameRequired)
			case "redirect_uris":
				h.Logger.Warn("Redirect URIs are required")
				return h.redirectWithError(res, adminClientsPath, errClientRedirectURIRequired)
			}
		}
	}

	// Parse redirect URIs (comma-separated or newline-separated)
	var redirectURIs []string
	// Split by newlines first, then by commas
	lines := strings.SplitSeq(redirectURIsStr, "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// If the line contains commas, split by commas too
			if strings.Contains(line, ",") {
				for uri := range strings.SplitSeq(line, ",") {
					uri = strings.TrimSpace(uri)
					if uri != "" {
						redirectURIs = append(redirectURIs, uri)
					}
				}
			} else {
				redirectURIs = append(redirectURIs, line)
			}
		}
	}

	if len(redirectURIs) == 0 {
		h.Logger.Warn("At least one redirect URI is required")
		return h.redirectWithError(res, adminClientsPath, errClientRedirectURIRequired)
	}

	client, err := h.AccountStore.NewClient(name, redirectURIs)
	if err != nil {
		h.Logger.Error("Failed to create new client object", "error", err)
		return h.redirectWithError(res, adminClientsPath, errClientInternalError)
	}

	// Create new client
	_, err = h.AccountStore.SaveClient(ctx, client)
	if err != nil {
		h.Logger.Error("Failed to create client", "error", err)
		return h.redirectWithError(res, adminClientsPath, errClientExists)
	}

	// Log the action
	sess := ctx.Value(SessionContextKey).(session.Session)
	h.AuditService.LogEvent(ctx, audit.LogEventRequest{
		Event:     audit.EventDataWrite,
		AccountID: sess.AccountID.V,
		Resource:  "clients",
		Action:    "create",
		Success:   true,
		IPAddress: req.RemoteAddr(),
		UserAgent: req.UserAgent(),
		Details: map[string]any{
			"created_client_name": name,
			"redirect_uris":       redirectURIs,
		},
	})

	h.Logger.Info("Client created successfully", "name", name, "redirect_uris", redirectURIs)
	return h.redirectWithClientInfo(res, client.ID.String(), client.Secret, name)
}

// DeleteClient removes an OAuth client with comprehensive validation and error handling
func (h *AppHandler) DeleteClient(ctx context.Context, req *http.Request, res *http.Response) error {
	// Parse form data
	if err := h.parseFormOrRedirect(req, res, adminClientsPath); err != nil {
		return err
	}

	// Extract client ID
	clientIDStr := req.FormValue("client_id")
	if clientIDStr == "" {
		h.Logger.Warn("Client ID is required")
		return h.redirectWithError(res, adminClientsPath, errClientIDRequired)
	}

	clientID, err := uuid.Parse(clientIDStr)
	if err != nil {
		h.Logger.Warn("Invalid client ID format", "client_id", clientIDStr, "error", err)
		return h.redirectWithError(res, adminClientsPath, errClientInvalidID)
	}

	// Get the client details for logging before deletion
	client, err := h.AccountStore.GetClient(ctx, clientID)
	if err != nil {
		if errors.Is(err, account.ErrClientNotFound) {
			h.Logger.Warn("Client not found for deletion", "client_id", clientID)
			return h.redirectWithError(res, adminClientsPath, errClientNotFound)
		}
		h.Logger.Error("Failed to get client for deletion", "client_id", clientID, "error", err)
		return h.redirectWithError(res, adminClientsPath, errClientInternalError)
	}

	// Delete the client
	err = h.AccountStore.DeleteClient(ctx, clientID)
	if err != nil {
		h.Logger.Error("Failed to delete client", "client_id", clientID, "error", err)
		return h.redirectWithError(res, adminClientsPath, errClientDeleteFailed)
	}

	// Log the action
	sess := ctx.Value(SessionContextKey).(session.Session)
	h.AuditService.LogEvent(ctx, audit.LogEventRequest{
		Event:     audit.EventDataDelete,
		AccountID: sess.AccountID.V,
		Resource:  "clients",
		Action:    "delete",
		Success:   true,
		IPAddress: req.RemoteAddr(),
		UserAgent: req.UserAgent(),
		Details: map[string]interface{}{
			"deleted_client_id":   clientID.String(),
			"deleted_client_name": client.Name,
		},
	})

	h.Logger.Info("Client deleted successfully", "client_id", clientID, "name", client.Name)
	return h.redirectWithSuccess(res, adminClientsPath, successClientDeleted)
}

// Resource server management handlers ====================================

// ShowResourceServers displays the resource server management page with success/error messages
func (h *AppHandler) ShowResourceServers(ctx context.Context, req *http.Request, res *http.Response) error {
	sess := ctx.Value(SessionContextKey).(session.Session)

	h.AuditService.LogEvent(ctx, audit.LogEventRequest{
		Event:     audit.EventDataRead,
		AccountID: sess.AccountID.V,
		Resource:  "resource servers",
		Action:    "view",
		Success:   true,
		IPAddress: req.RemoteAddr(),
		UserAgent: req.UserAgent(),
	})

	// Parse pagination parameters
	pageSize := 25 // default
	if pageSizeParam := req.URLQueryParam("page_size"); pageSizeParam != "" {
		if ps, err := strconv.Atoi(pageSizeParam); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	token := req.URLQueryParam("token")

	// Use the new paginated method
	listReq := oauth.ListResourceServersRequest{
		PageSize: pageSize,
		Token:    token,
	}

	result, err := h.ResourceServerService.ListResourceServersPaginated(ctx, listReq)
	if err != nil {
		h.Logger.Warn("Failed to list clients", "error", err)
		return err
	}

	// Build pagination info
	pagination := component.PaginationInfo{
		PageSize:      pageSize,
		HasNext:       result.NextToken.E,
		HasPrev:       result.PrevToken.E,
		NextPageToken: result.NextToken.V,
		PrevPageToken: result.PrevToken.V,
		BasePath:      "/admin/resource-servers",
	}

	// Check for success/error messages in query parameters
	successMsg := req.URLQueryParam("success")
	errorMsg := req.URLQueryParam("error")
	resourceServerURL := req.URLQueryParam("resource_server_url")
	resourceServerID := req.URLQueryParam("resource_server_id")

	appProps, err := h.LayoutAppProps(ctx, "Resource Servers")
	if err != nil {
		h.Logger.Error("failed to get app props for resource servers", "error", err)
		return res.Send(http.StatusInternalServerError)
	}

	props := page.ResourceServersProps{
		AppProps:          appProps,
		ResourceServers:   result.ResourceServers,
		SuccessMessage:    successMsg,
		ErrorMessage:      errorMsg,
		ResourceServerURL: resourceServerURL,
		ResourceServerID:  resourceServerID,
		Pagination:        pagination,
	}

	return ui.Render(ctx, res, page.ResourceServers(props))
}

// NewResourceServer creates a new resource server with comprehensive validation and error handling
func (h *AppHandler) NewResourceServer(ctx context.Context, req *http.Request, res *http.Response) error {
	// Parse form data
	if err := h.parseFormOrRedirect(req, res, adminResourceServersPath); err != nil {
		return err
	}

	// Extract form values
	formUrl := req.FormValue("url")

	// Validate required fields
	if formUrl == "" {
		h.Logger.Warn("URL is required")
		return h.redirectWithError(res, adminResourceServersPath, errResourceServerURLRequired)
	}

	// Check if url is already used
	_, err := h.OAuthStore.GetResourceServerByURL(ctx, formUrl)
	if err == nil {
		return h.redirectWithError(res, adminResourceServersPath, errResourceServerExists)
	}
	if !errors.Is(err, oauth.ErrResourceServerNotFound) {
		return h.redirectWithError(res, adminResourceServersPath, errResourceServerInternalError)
	}

	resourceServer, err := h.OAuthStore.NewResourceServer(formUrl)
	if err != nil {
		h.Logger.Error("Failed to create new resource server object", "error", err)
		return h.redirectWithError(res, adminResourceServersPath, errResourceServerInternalError)
	}

	// Create new resource server
	_, err = h.OAuthStore.SaveResourceServer(ctx, resourceServer)
	if err != nil {
		h.Logger.Error("Failed to create resource server", "error", err)
		return h.redirectWithError(res, adminResourceServersPath, errResourceServerInternalError)
	}

	// Log the action
	sess := ctx.Value(SessionContextKey).(session.Session)
	h.AuditService.LogEvent(ctx, audit.LogEventRequest{
		Event:     audit.EventDataWrite,
		AccountID: sess.AccountID.V,
		Resource:  "resource servers",
		Action:    "create",
		Success:   true,
		IPAddress: req.RemoteAddr(),
		UserAgent: req.UserAgent(),
		Details: map[string]any{
			"url": formUrl,
		},
	})

	h.Logger.Info("Resource Server created successfully", "url", formUrl)
	return h.redirectWithResourceServerInfo(res, resourceServer.URL, resourceServer.ID.String())
}

// DeleteResourceServer removes a resource server with comprehensive validation and error handling
func (h *AppHandler) DeleteResourceServer(ctx context.Context, req *http.Request, res *http.Response) error {
	// Parse form data
	if err := h.parseFormOrRedirect(req, res, adminResourceServersPath); err != nil {
		return err
	}

	// Extract Resource Server ID
	resourceServerIDStr := req.FormValue("resource_server_id")
	if resourceServerIDStr == "" {
		h.Logger.Warn("Resource Server ID is required")
		return h.redirectWithError(res, adminResourceServersPath, errResourceServerIDRequired)
	}

	resourceServerID, err := uuid.Parse(resourceServerIDStr)
	if err != nil {
		h.Logger.Warn("Invalid resource server ID format", "resource_server_id", resourceServerIDStr, "error", err)
		return h.redirectWithError(res, adminResourceServersPath, errResourceServerInvalidID)
	}

	// Get the resource server details for logging before deletion
	resourceServer, err := h.OAuthStore.GetResourceServerByID(ctx, resourceServerID)
	if err != nil {
		if errors.Is(err, oauth.ErrResourceServerNotFound) {
			h.Logger.Warn("Resource Server not found for deletion", "resource_server_id", resourceServerID)
			return h.redirectWithError(res, adminResourceServersPath, errResourceServerNotFound)
		}

		h.Logger.Error("Failed to get resource server for deletion", "resource_server_id", resourceServerID, "error", err)
		return h.redirectWithError(res, adminResourceServersPath, errResourceServerInternalError)
	}

	// Delete the Resource Server
	err = h.OAuthStore.DeleteResourceServer(ctx, resourceServerID)
	if err != nil {
		h.Logger.Error("Failed to delete resource server", "resource_server_id", resourceServer, "error", err)
		return h.redirectWithError(res, adminResourceServersPath, errResourceServerDeleteFailed)
	}

	// Log the action
	sess := ctx.Value(SessionContextKey).(session.Session)
	h.AuditService.LogEvent(ctx, audit.LogEventRequest{
		Event:     audit.EventDataDelete,
		AccountID: sess.AccountID.V,
		Resource:  "resource servers",
		Action:    "delete",
		Success:   true,
		IPAddress: req.RemoteAddr(),
		UserAgent: req.UserAgent(),
		Details: map[string]any{
			"deleted_resource_server_id":  resourceServer.ID.String(),
			"deleted_resource_server_url": resourceServer.URL,
		},
	})

	h.Logger.Info("Resource Server deleted successfully", "resource_server_id", resourceServer.ID, "name", resourceServer.URL)
	return h.redirectWithSuccess(res, adminResourceServersPath, successResourceServerDeleted)
}

// Audit log handlers =====================================================

// ShowAuditLogs displays the audit logs page with pagination and filtering
func (h *AppHandler) ShowAuditLogs(ctx context.Context, req *http.Request, res *http.Response) error {
	sess := ctx.Value(SessionContextKey).(session.Session)

	h.AuditService.LogEvent(ctx, audit.LogEventRequest{
		Event:     audit.EventDataRead,
		AccountID: sess.AccountID.V,
		Resource:  "audit_logs",
		Action:    "view",
		Success:   true,
		IPAddress: req.RemoteAddr(),
		UserAgent: req.UserAgent(),
	})

	// Parse pagination parameters
	pageSize := 25 // default
	if pageSizeParam := req.URLQueryParam("page_size"); pageSizeParam != "" {
		if ps, err := strconv.Atoi(pageSizeParam); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	token := req.URLQueryParam("token")

	// Use the new paginated method
	listReq := audit.ListLogEntriesRequest{
		PageSize: pageSize,
		Token:    token,
	}

	result, err := h.AuditService.ListLogEntriesPaginated(ctx, listReq)
	if err != nil {
		h.Logger.Warn("Failed to get audit logs", "error", err)
		return err
	}

	// Build pagination info
	pagination := component.PaginationInfo{
		PageSize:      pageSize,
		HasNext:       result.NextToken.E,
		HasPrev:       result.PrevToken.E,
		NextPageToken: result.NextToken.V,
		PrevPageToken: result.PrevToken.V,
		BasePath:      "/admin/audit-logs",
	}

	// Check for success/error messages in query parameters
	successMsg := req.URLQueryParam("success")
	errorMsg := req.URLQueryParam("error")

	appProps, err := h.LayoutAppProps(ctx, "Audit Logs")
	if err != nil {
		h.Logger.Error("failed to get app props for audit logs", "error", err)
		return res.Send(http.StatusInternalServerError)
	}

	return ui.Render(ctx, res, page.AuditLogs(page.AuditLogsProps{
		AppProps:       appProps,
		LogEntries:     result.LogEntries,
		SuccessMessage: successMsg,
		ErrorMessage:   errorMsg,
		Pagination:     pagination,
	}))
}

// Session management handlers ============================================

// SignOff handles user sign-off requests
func (h *AppHandler) SignOff(ctx context.Context, req *http.Request, res *http.Response) error {
	sess := ctx.Value(SessionContextKey).(session.Session)

	if err := h.SessionStore.DeleteSession(ctx, sess.ID); err != nil {
		h.Logger.Warn("Failed to delete session on sign off", "error", err, "session_id", sess.ID)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	h.SessionStore.ClearCookie(ctx, res)

	return res.SendJSON(http.StatusOK, map[string]string{"message": "signed off"})
}

// Layout helper functions ================================================

// LayoutAppProps provides common properties needed for app layout rendering
func (h *AppHandler) LayoutAppProps(ctx context.Context, pageTitle string) (layout.AppProps, error) {
	sess := ctx.Value(SessionContextKey).(session.Session)
	csrfToken := sess.Values[CSRFTokenSessionValueKey].(string)

	// Fetch user data
	user, err := h.AccountStore.GetUser(ctx, sess.AccountID.V)
	if err != nil {
		return layout.AppProps{}, err
	}

	return layout.AppProps{
		RootProps: layout.RootProps{
			Title:     pageTitle,
			CSRFToken: csrfToken,
		},
		User:      user,
		PageTitle: pageTitle,
	}, nil
}

func (h *AppHandler) ShowEditResourceServer(ctx context.Context, req *http.Request, res *http.Response) error {
	idStr := req.URLPathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return res.SendRedirect(http.StatusSeeOther, "/admin/resource-servers?error="+url.QueryEscape("Invalid resource server ID"))
	}

	resourceServer, err := h.OAuthStore.GetResourceServerByID(ctx, id)
	if err != nil {
		h.Logger.Error("failed to get resource server", "error", err, "id", id)
		return res.SendRedirect(http.StatusSeeOther, "/admin/resource-servers?error="+url.QueryEscape("Failed to load resource server"))
	}

	appProps, err := h.LayoutAppProps(ctx, "Edit Resource Server")
	if err != nil {
		h.Logger.Error("failed to get app props for edit resource server", "error", err)
		return res.Send(http.StatusInternalServerError)
	}

	props := page.ResourceServerEditProps{
		AppProps:       appProps,
		ResourceServer: resourceServer,
		SuccessMessage: req.URLQueryParam("success"),
		ErrorMessage:   req.URLQueryParam("error"),
	}

	return ui.Render(ctx, res, page.ResourceServerEdit(props))
}

func (h *AppHandler) CreateScope(ctx context.Context, req *http.Request, res *http.Response) error {
	idStr := req.URLPathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return res.SendRedirect(http.StatusSeeOther, "/admin/resource-servers?error="+url.QueryEscape("Invalid resource server ID"))
	}

	scopeName := req.FormValue("scope_name")
	scopeDescription := req.FormValue("scope_description")

	if scopeName == "" {
		return res.SendRedirect(http.StatusSeeOther, "/admin/resource-servers/"+idStr+"/edit?error="+url.QueryEscape("Scope name is required"))
	}

	resourceServer, err := h.OAuthStore.GetResourceServerByID(ctx, id)
	if err != nil {
		h.Logger.Error("failed to get resource server", "error", err, "id", id)
		return res.SendRedirect(http.StatusSeeOther, "/admin/resource-servers/"+idStr+"/edit?error="+url.QueryEscape("Failed to load resource server"))
	}

	scope := oauth.Scope{
		ID:          uuid.New(),
		Name:        resourceServer.URL + "/" + scopeName,
		Description: scopeDescription,
	}

	err = h.OAuthStore.CreateScope(ctx, resourceServer, scope)
	if err != nil {
		h.Logger.Error("failed to create scope", "error", err, "resource_server_id", id, "scope_name", scopeName)
		return res.SendRedirect(http.StatusSeeOther, "/admin/resource-servers/"+idStr+"/edit?error="+url.QueryEscape("Failed to create scope"))
	}

	h.Logger.Info("scope created", "resource_server_id", id, "scope_name", scopeName)
	return res.SendRedirect(http.StatusSeeOther, "/admin/resource-servers/"+idStr+"/edit?success="+url.QueryEscape("Scope '"+scopeName+"' created successfully"))
}

func (h *AppHandler) UpdateScope(ctx context.Context, req *http.Request, res *http.Response) error {
	idStr := req.URLPathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return res.SendRedirect(http.StatusSeeOther, "/admin/resource-servers?error="+url.QueryEscape("Invalid resource server ID"))
	}

	originalScopeName := req.FormValue("original_scope_name")
	scopeName := req.FormValue("scope_name")
	scopeDescription := req.FormValue("scope_description")

	if originalScopeName == "" || scopeName == "" {
		return res.SendRedirect(http.StatusSeeOther, "/admin/resource-servers/"+idStr+"/edit?error="+url.QueryEscape("Scope name is required"))
	}

	resourceServer, err := h.OAuthStore.GetResourceServerByID(ctx, id)
	if err != nil {
		h.Logger.Error("failed to get resource server", "error", err, "id", id)
		return res.SendRedirect(http.StatusSeeOther, "/admin/resource-servers/"+idStr+"/edit?error="+url.QueryEscape("Failed to load resource server"))
	}

	scope := oauth.Scope{
		Name:        resourceServer.URL + "/" + scopeName,
		Description: scopeDescription,
	}

	err = h.OAuthStore.UpdateScope(ctx, id, originalScopeName, scope)
	if err != nil {
		h.Logger.Error("failed to update scope", "error", err, "resource_server_id", id, "original_scope_name", originalScopeName, "scope_name", scopeName)
		return res.SendRedirect(http.StatusSeeOther, "/admin/resource-servers/"+idStr+"/edit?error="+url.QueryEscape("Failed to update scope"))
	}

	h.Logger.Info("scope updated", "resource_server_id", id, "original_scope_name", originalScopeName, "scope_name", scopeName)
	return res.SendRedirect(http.StatusSeeOther, "/admin/resource-servers/"+idStr+"/edit?success="+url.QueryEscape("Scope '"+scopeName+"' updated successfully"))
}

func (h *AppHandler) DeleteScope(ctx context.Context, req *http.Request, res *http.Response) error {
	idStr := req.URLPathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return res.SendRedirect(http.StatusSeeOther, "/admin/resource-servers?error="+url.QueryEscape("Invalid resource server ID"))
	}

	scopeName := req.FormValue("scope_name")
	if scopeName == "" {
		return res.SendRedirect(http.StatusSeeOther, "/admin/resource-servers/"+idStr+"/edit?error="+url.QueryEscape("Scope name is required"))
	}

	err = h.OAuthStore.DeleteScope(ctx, id, scopeName)
	if err != nil {
		h.Logger.Error("failed to delete scope", "error", err, "resource_server_id", id, "scope_name", scopeName)
		return res.SendRedirect(http.StatusSeeOther, "/admin/resource-servers/"+idStr+"/edit?error="+url.QueryEscape("Failed to delete scope"))
	}

	h.Logger.Info("scope deleted", "resource_server_id", id, "scope_name", scopeName)
	return res.SendRedirect(http.StatusSeeOther, "/admin/resource-servers/"+idStr+"/edit?success="+url.QueryEscape("Scope '"+scopeName+"' deleted successfully"))
}
