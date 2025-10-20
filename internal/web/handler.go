package web

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/freekieb7/smauth/internal/database"
	"github.com/freekieb7/smauth/internal/util"
	"github.com/freekieb7/smauth/internal/web/view"
	"github.com/freekieb7/smauth/internal/web/view/component"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	Logger *slog.Logger
	DB     *database.Database
}

func NewHandler(logger *slog.Logger, db *database.Database) *Handler {
	return &Handler{
		Logger: logger,
		DB:     db,
	}
}

func (h *Handler) ShowUsers(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	users, err := h.DB.ListUsers(ctx)
	if err != nil {
		h.Logger.Error("Failed to fetch users", "error", err)
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return nil
	}

	viewUsers := make([]view.User, len(users))
	for i, u := range users {
		viewUsers[i] = view.User{
			ID:    u.ID,
			Email: u.Email,
			Name:  u.Name,
		}
	}

	return render(ctx, w, view.UsersPage(view.UsersPageProps{
		LayoutProps: component.LayoutProps{
			Title: "Users",
		},
		Users: viewUsers,
	}))
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return nil
	}

	// Check if user with the same email already exists
	_, err := h.DB.GetUserByEmail(ctx, req.Email)
	if err == nil {
		http.Error(w, "User with this email already exists", http.StatusConflict)
		return nil
	}

	// Create new user
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.Logger.Error("Failed to hash password", "error", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return nil
	}

	_, err = h.DB.CreateUser(ctx, database.CreateUserParams{
		Email:        req.Email,
		PasswordHash: string(bcryptHash),
		Name:         req.Name,
	})
	if err != nil {
		h.Logger.Error("Failed to create user", "error", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return nil
	}

	w.WriteHeader(http.StatusCreated)
	return nil
}

type UpdateUserRequest struct {
	UserID   string                `json:"id"`
	Email    util.Optional[string] `json:"email"`
	Password util.Optional[string] `json:"password"`
	Name     util.Optional[string] `json:"name"`
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return nil
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return nil
	}

	// Check if user exists
	user, err := h.DB.GetUserByID(ctx, userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return nil
	}

	passwordHash := util.None[string]()
	if req.Password.Some {
		bcryptHash, err := bcrypt.GenerateFromPassword([]byte(req.Password.Data), bcrypt.DefaultCost)
		if err != nil {
			h.Logger.Error("Failed to hash password", "error", err)
			http.Error(w, "Failed to update user", http.StatusInternalServerError)
			return nil
		}
		passwordHash = util.Some(string(bcryptHash))
	}

	err = h.DB.UpdateUserByID(ctx, user.ID, database.UpdateUserParams{
		Email:        req.Email,
		PasswordHash: passwordHash,
		Name:         req.Name,
	})
	if err != nil {
		h.Logger.Error("Failed to update user", "error", err)
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return nil
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

type DeleteUserRequest struct {
	UserID string `json:"id"`
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var req DeleteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return nil
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return nil
	}

	err = h.DB.DeleteUserByID(ctx, userID)
	if err != nil {
		h.Logger.Error("Failed to delete user", "error", err)
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return nil
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) ShowClients(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	clients, err := h.DB.ListClients(ctx)
	if err != nil {
		h.Logger.Error("Failed to list clients", "error", err)
		http.Error(w, "Failed to list clients", http.StatusInternalServerError)
		return nil
	}

	viewClients := make([]view.Client, len(clients))
	for i, c := range clients {
		viewClients[i] = view.Client{
			ID:           c.ID,
			Name:         c.Name,
			RedirectURIs: c.RedirectURIs,
			CreatedAt:    c.CreatedAt,
			UpdatedAt:    c.UpdatedAt,
		}
	}

	return render(ctx, w, view.ClientsPage(view.ClientsPageProps{
		LayoutProps: component.LayoutProps{
			Title: "Clients",
		},
		Clients: viewClients,
	}))
}

type CreateClientRequest struct {
	Name         string   `json:"name"`
	RedirectURIs []string `json:"redirect_uris"`
}

func (h *Handler) CreateClient(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var req CreateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return nil
	}

	// Generate a random secret for the client
	secret, err := util.GenerateRandomString(32)
	if err != nil {
		h.Logger.Error("Failed to generate client secret", "error", err)
		http.Error(w, "Failed to create client", http.StatusInternalServerError)
		return nil
	}

	_, err = h.DB.CreateClient(ctx, database.CreateClientParams{
		Name:         req.Name,
		Secret:       secret,
		RedirectURIs: req.RedirectURIs,
	})
	if err != nil {
		h.Logger.Error("Failed to create client", "error", err)
		http.Error(w, "Failed to create client", http.StatusInternalServerError)
		return nil
	}

	w.WriteHeader(http.StatusCreated)
	return nil
}

type UpdateClientRequest struct {
	ClientID     string                  `json:"id"`
	Name         util.Optional[string]   `json:"name"`
	RedirectURIs util.Optional[[]string] `json:"redirect_uris"`
}

func (h *Handler) UpdateClient(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var req UpdateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return nil
	}

	clientID, err := uuid.Parse(req.ClientID)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return nil
	}

	err = h.DB.UpdateClientByID(ctx, clientID, database.UpdateClientParams{
		Name:         req.Name,
		RedirectURIs: req.RedirectURIs,
	})
	if err != nil {
		h.Logger.Error("Failed to update client", "error", err)
		http.Error(w, "Failed to update client", http.StatusInternalServerError)
		return nil
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

type DeleteClientRequest struct {
	ClientID string `json:"id"`
}

func (h *Handler) DeleteClient(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var req DeleteClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return nil
	}

	clientID, err := uuid.Parse(req.ClientID)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return nil
	}

	err = h.DB.DeleteClientByID(ctx, clientID)
	if err != nil {
		h.Logger.Error("Failed to delete client", "error", err)
		http.Error(w, "Failed to delete client", http.StatusInternalServerError)
		return nil
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

type AuthorizeRequest struct {
	ResponseType string
	ClientID     string
	RedirectURI  string
	Scope        string
	State        string
}

func HandleAuthorize(db *database.Database) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return nil
		}

		var req AuthorizeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return nil
		}

		// Process the authorization request
		// ...

		w.Write([]byte("Hello, World!"))
		return nil
	}
}

type TokenRequest struct {
	GrantType    string
	Code         string
	RedirectURI  string
	ClientID     string
	ClientSecret string
}

func HandleToken(db *database.Database) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return nil
		}

		var req TokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return nil
		}

		// Process the token request
		// ...
		w.Write([]byte("Token endpoint"))
		return nil
	}
}

func render(ctx context.Context, w http.ResponseWriter, component templ.Component) error {
	w.Header().Set("Content-Type", "text/html")
	return component.Render(ctx, w)
}
