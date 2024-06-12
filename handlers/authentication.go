package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"gitea.com/go-chi/session"
	"github.com/go-chi/chi/v5"

	"github.com/zaibon/shortcut/components"
	"github.com/zaibon/shortcut/domain"
	"github.com/zaibon/shortcut/middleware"
	"github.com/zaibon/shortcut/services"
	"github.com/zaibon/shortcut/views"
)

type AuthService interface {
	CreateUser(ctx context.Context, user *domain.User) error
	VerifyLogin(ctx context.Context, email, password string) (*domain.User, error)
}

type UsersHandler struct {
	svc AuthService
}

func NewUsersHandler(svc AuthService) *UsersHandler {
	return &UsersHandler{
		svc: svc,
	}
}

func (h *UsersHandler) Routes(r chi.Router) {
	r.Get("/login", h.loginPage)
	r.Get("/sign-up", h.signUp)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticated)
		r.Get("/my-account", h.myAccount)
	})

	r.Post("/auth/register", h.register)
	r.Post("/auth/login", h.login)
	r.Get("/logout", h.logout)
}

func (h *UsersHandler) loginPage(w http.ResponseWriter, r *http.Request) {
	data := views.LoginPageData{}
	Render(r.Context(), w, views.LoginPage(data))
}

func (h *UsersHandler) signUp(w http.ResponseWriter, r *http.Request) {
	data := views.SignUpPageData{}
	Render(r.Context(), w, views.SignUpPage(data))
}

func (h *UsersHandler) myAccount(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	Render(r.Context(), w, views.MyAccount(views.MyAccountData{
		User: user,
	}))
}

func (h *UsersHandler) login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.svc.VerifyLogin(r.Context(), email, password)
	if err != nil {
		w.WriteHeader(ErrorStatus(err))
		Render(r.Context(), w, components.LoginForm(components.LoginFormData{
			Email:    email,
			Password: password,
			Alerts:   []components.LoginAlert{loginAlerts(err)},
		}))
		return
	}

	sess := session.GetSession(r)
	sess.Set("user", user)

	HXRedirect(r.Context(), w, "/")
}

func (h *UsersHandler) register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("password-confirm")

	if len(name) == 0 || len(email) == 0 || len(password) == 0 {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	if password != confirmPassword {
		http.Error(w, "passwords do not match", http.StatusBadRequest)
		return
	}

	user := &domain.User{
		Name:     name,
		Email:    email,
		Password: password,
	}

	if err := h.svc.CreateUser(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sess := session.GetSession(r)
	sess.Set("user", user)

	HXRedirect(r.Context(), w, "/")
}

func (h *UsersHandler) logout(w http.ResponseWriter, r *http.Request) {
	sess := session.GetSession(r)
	id := sess.ID()

	sess.Delete("user")
	if err := sess.Flush(); err != nil {
		slog.Error("failed to flush sessions", "id", id)
	}
	if err := sess.Destroy(w, r); err != nil {
		slog.Error("failed to destroy sessions", "id", id)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func loginAlerts(err error) components.LoginAlert {
	if errors.Is(err, services.ErrInvalidCredentials) {
		return components.LoginAlert{
			Title: "Invalid credentials",
			Text:  "The password you entered is incorrect. Please try again.",
		}
	}

	if errors.Is(err, services.ErrInvalidCredentials) {
		return components.LoginAlert{
			Title: "Invalid credentials",
			Text:  "The password you entered is incorrect. Please try again.",
		}
	}

	return components.LoginAlert{
		Title: "Login failed",
		Text:  "An unexpected error occurred. Please try again.",
	}
}
