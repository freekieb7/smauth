package web

import (
	"context"
	"log/slog"

	"github.com/freekieb7/smauth/internal/auth"
	"github.com/freekieb7/smauth/internal/http"
	"github.com/freekieb7/smauth/internal/session"
	"github.com/freekieb7/smauth/internal/util"
	"github.com/freekieb7/smauth/internal/web/ui"
	"github.com/freekieb7/smauth/internal/web/ui/template/layout"
	"github.com/freekieb7/smauth/internal/web/ui/template/page"
)

type AdminHandler struct {
	Logger       *slog.Logger
	SessionStore *session.Store
	AuthStore    *auth.Store
}

func NewAdminHandler(logger *slog.Logger, sessionStore *session.Store, authStore *auth.Store) AdminHandler {
	return AdminHandler{
		Logger:       logger,
		SessionStore: sessionStore,
		AuthStore:    authStore,
	}
}

func (h *AdminHandler) RegisterRoutes(router *http.Router) {
	router.Group("/admin", func(group *http.Router) {
		group.GET("", h.ShowHome)
		group.Group("/users", func(group *http.Router) {
			group.GET("", h.ShowUsers)
			group.POST("/new", h.NewUser)

		})
	}, SessionMiddleware(h.Logger, h.SessionStore), PageAuthenticatedMiddleware(h.Logger), CSRFMiddleware(h.Logger, h.SessionStore))
}

func (h *AdminHandler) ShowHome(ctx context.Context, req *http.Request, res *http.Response) error {
	return ui.Render(ctx, res, page.Home(page.HomeProps{
		AdminProps: h.LayoutAdminProps(ctx),
	}))
}

func (h *AdminHandler) ShowUsers(ctx context.Context, req *http.Request, res *http.Response) error {
	users, err := h.AuthStore.ListUsers(ctx)
	if err != nil {
		h.Logger.Warn("Failed to list users", "error", err)
		return err
	}

	return ui.Render(ctx, res, page.Users(page.UsersProps{
		AdminProps: h.LayoutAdminProps(ctx),
		Users:      users,
	}))
}

type NewUserForm struct {
	Email string    `form:"email"`
	Role  auth.Role `form:"role"`
}

func (h *AdminHandler) NewUser(ctx context.Context, req *http.Request, res *http.Response) error {
	var form NewUserForm
	if err := req.DecodeJSON(&form); err != nil {
		h.Logger.Warn("Failed to decode form", "error", err)
		return res.SendJSON(http.StatusBadRequest, map[string]string{"error": "invalid form"})
	}

	randomPassword, err := util.GenerateRandomString(32)
	if err != nil {
		h.Logger.Error("Failed to generate random password", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	user, err := h.AuthStore.NewUser(form.Email, randomPassword, form.Role)
	if err != nil {
		h.Logger.Error("Failed to create new user object", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	// Create new user
	_, err = h.AuthStore.SaveUser(ctx, user)
	if err != nil {
		h.Logger.Error("Failed to create user", "error", err)
		return res.SendJSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return res.SendJSON(http.StatusCreated, map[string]string{"message": "user created"})
}

func (h *AdminHandler) LayoutAdminProps(ctx context.Context) layout.AdminProps {
	sess := ctx.Value(SessionContextKey).(session.Session)
	csrfToken := sess.Values[CSRFTokenSessionValueKey].(string)

	return layout.AdminProps{
		RootProps: layout.RootProps{
			Title:     "Admin",
			CSRFToken: csrfToken,
		},
	}
}
