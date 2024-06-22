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
	"github.com/zaibon/shortcut/handlers/toast"
	"github.com/zaibon/shortcut/middleware"
	"github.com/zaibon/shortcut/services"
	"github.com/zaibon/shortcut/views"
)

type AuthService interface {
	CreateUser(ctx context.Context, user *domain.User) error
	UpdateUser(ctx context.Context, id domain.ID, user *domain.User) (*domain.User, error)
	UpdatePassword(ctx context.Context, id domain.ID, password string) error
	VerifyLogin(ctx context.Context, email, password string) (*domain.User, error)

	// oauth
	InitiateOauthFlow(ctx context.Context) (string, error)
	IdentifyOauthUser(ctx context.Context, code string) (*domain.User, error)
	VerifyOauthState(ctx context.Context, state string) (bool, error)
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
	// user/password auth
	r.Get("/login", h.loginPage)
	r.Get("/sign-up", h.signUp)
	r.Post("/auth/register", h.register)
	r.Post("/auth/login", h.login)
	r.Get("/logout", h.logout)

	// oauth
	r.Get("/oauth/callback", h.oauthCallback)
	r.Get("/oauth/login", h.initiateGoogleOauthFlow)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticated)
		r.Get("/my-account", h.myAccount)

		r.Post("/auth/edit-account", h.editAccount)
		r.Post("/auth/edit-password", h.editPassword)
	})
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
			Alerts: components.AlertListData{
				Alerts: []components.Alert{loginAlerts(err)},
			},
		}))
		return
	}

	if err := session.GetSession(r).Set("user", user); err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

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

	if err := session.GetSession(r).Set("user", user); err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	HXRedirect(r.Context(), w, "/")
}

func (h *UsersHandler) logout(w http.ResponseWriter, r *http.Request) {
	sess := session.GetSession(r)
	id := sess.ID()

	if err := sess.Destroy(w, r); err != nil {
		slog.Error("failed to destroy sessions", "id", id)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *UsersHandler) editAccount(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	name := r.FormValue("name")

	user := middleware.UserFromContext(r.Context())
	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if len(name) == 0 || len(email) == 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Render(r.Context(), w, views.MyAccount(views.MyAccountData{
			User: user,
			EditAlerts: components.AlertListData{
				Alerts: []components.Alert{
					{
						Title: "Edit failed",
						Text:  "Name and email are required",
					},
				},
			},
		}))
		return
	}

	user, err := h.svc.UpdateUser(r.Context(), user.ID, &domain.User{
		Name:  name,
		Email: email,
	})
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Render(r.Context(), w, views.MyAccount(views.MyAccountData{
			User: user,
			EditAlerts: components.AlertListData{
				Alerts: []components.Alert{
					{
						Title: "Edit failed",
						Text:  "An unexpected error occurred. Please try again.",
					},
				},
			},
		}))
		return
	}

	sess := session.GetSession(r)
	if err := sess.Set("user", user); err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	toast.Success(w, "Account updated", "Your account has been updated successfully")
	Render(r.Context(), w, components.EditAccountForm(components.EditAccountFormData{
		User: user,
	}))
}

func (h *UsersHandler) editPassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	oldPassword := r.FormValue("old_password")
	newPassword := r.FormValue("new_password")
	newPasswordConfirm := r.FormValue("new_password_confirm")

	data := components.EditPasswordFormData{
		OldPassword:     oldPassword,
		NewPassword:     newPassword,
		ConfirmPassword: newPasswordConfirm,
	}

	user := middleware.UserFromContext(r.Context())
	if user == nil {
		data.Alerts = components.AlertListData{
			Alerts: []components.Alert{{
				Title: "Edit failed",
				Text:  "User not found",
			}},
		}
		Render(r.Context(), w, components.EditPasswordForm(data))
		return
	}

	if len(oldPassword) == 0 || len(newPassword) == 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		data.Alerts = components.AlertListData{
			Alerts: []components.Alert{{
				Title: "Edit failed",
				Text:  "Old and new password are required",
			}},
		}
		Render(r.Context(), w, components.EditPasswordForm(data))
		return
	}

	if newPassword != newPasswordConfirm {
		w.WriteHeader(http.StatusUnprocessableEntity)
		data.Alerts = components.AlertListData{
			Alerts: []components.Alert{{
				Title: "Edit failed",
				Text:  "New passwords do not match",
			}},
		}
		Render(r.Context(), w, components.EditPasswordForm(data))
		return
	}

	if _, err := h.svc.VerifyLogin(r.Context(), user.Email, oldPassword); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		data.Alerts = components.AlertListData{
			Alerts: []components.Alert{{
				Title: "Edit failed",
				Text:  "Old password is incorrect",
			}},
		}
		Render(r.Context(), w, components.EditPasswordForm(data))
		return
	}

	if err := h.svc.UpdatePassword(r.Context(), user.ID, newPassword); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		data.Alerts = components.AlertListData{
			Alerts: []components.Alert{{
				Title: "Edit failed",
				Text:  "An unexpected error occurred. Please try again.",
			}},
		}
		Render(r.Context(), w, components.EditPasswordForm(data))
		return
	}

	sess := session.GetSession(r)
	if err := sess.Set("user", user); err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return

	}

	toast.Success(w, "Password updated", "Your password has been updated successfully")
	w.Header().Set("HX-Reswap", "none")
	Render(r.Context(), w, components.EditPasswordForm(components.EditPasswordFormData{}))
}

func (h *UsersHandler) initiateGoogleOauthFlow(w http.ResponseWriter, r *http.Request) {
	authURL, err := h.svc.InitiateOauthFlow(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *UsersHandler) oauthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" {
		http.Error(w, "code is required", http.StatusBadRequest)
		return
	}
	if state == "" {
		http.Error(w, "state is required", http.StatusBadRequest)
		return
	}

	isValid, err := h.svc.VerifyOauthState(r.Context(), state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !isValid {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	user, err := h.svc.IdentifyOauthUser(r.Context(), code)
	if err != nil {
		w.WriteHeader(ErrorStatus(err))
		Render(r.Context(), w, components.LoginForm(components.LoginFormData{
			Alerts: components.AlertListData{
				Alerts: []components.Alert{loginAlerts(err)},
			},
		}))
		return
	}

	if err := session.GetSession(r).Set("user", user); err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func loginAlerts(err error) components.Alert {
	if errors.Is(err, services.ErrInvalidCredentials) {
		return components.Alert{
			Title: "Invalid credentials",
			Text:  "The password you entered is incorrect. Please try again.",
		}
	}

	if errors.Is(err, services.ErrUserNotFound) {
		return components.Alert{
			Title: "User not found",
			Text:  "The email you entered is not associated with an account. Please try again.",
		}
	}

	return components.Alert{
		Title: "Login failed",
		Text:  "An unexpected error occurred. Please try again.",
	}
}
