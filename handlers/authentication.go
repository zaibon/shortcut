package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"gitea.com/go-chi/session"
	"github.com/donseba/go-htmx"
	"github.com/go-chi/chi/v5"

	"github.com/zaibon/shortcut/domain"
	"github.com/zaibon/shortcut/middleware"
	"github.com/zaibon/shortcut/templates"
)

type AuthService interface {
	// oauth
	InitiateOauthFlow(ctx context.Context, provider domain.OauthProvider) (string, error)
	IdentifyOauthUser(ctx context.Context, code string, provider domain.OauthProvider) (*domain.User, error)
	VerifyOauthState(ctx context.Context, state string) (bool, domain.OauthProvider, error)
}

type UsersHandler struct {
	htmx         *htmx.HTMX
	auth         AuthService
	stripe       stripeService
	stripePubKey string
}

func NewUsersHandler(svc AuthService, stripe stripeService, stripePubKey string) *UsersHandler {
	return &UsersHandler{
		htmx:         htmx.New(),
		auth:         svc,
		stripe:       stripe,
		stripePubKey: stripePubKey,
	}
}

func (h *UsersHandler) Routes(r chi.Router) {
	r.Get("/auth", h.authPage)
	r.Get("/logout", h.logout)

	// oauth
	r.Get("/oauth/login/{provider}", h.initiateOauthFlow)
	r.Get("/oauth/callback", h.oauthCallback)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticated)
		r.Get("/account", h.myAccount)
	})
}

func (h *UsersHandler) authPage(w http.ResponseWriter, r *http.Request) {
	templates.AuthPage().
		Render(r.Context(), w)
}

func (h *UsersHandler) myAccount(w http.ResponseWriter, r *http.Request) {
	// user := middleware.UserFromContext(r.Context())

	// subscription, err := h.stripe.GetSubscription(r.Context(), user)
	// if err != nil && !errors.Is(err, services.ErrNotSubscription) {
	// 	slog.Error("failed to get subscription detail", "err", err)
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// var customerPortalURL string
	// if errors.Is(err, services.ErrNotSubscription) {
	// 	subscription = nil
	// } else {
	// 	customerPortalURL, err = h.stripe.GenerateCustomerPortalURL(r.Context(), user)
	// 	if err != nil {
	// 		slog.Error("failed to generate customer portal URL", "err", err)
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	// }

	templates.AccountPage().
		Render(r.Context(), w)
	// Render(r.Context(), w, views.MyAccount(views.MyAccountData{
	// 	User:                  user,
	// 	Subscription:          subscription,
	// 	CustomterDashboardURL: customerPortalURL,
	// 	PricingData: components.PricingData{
	// 		User:     user,
	// 		StipeKey: h.stripePubKey,
	// 	},
	// }))
}

func (h *UsersHandler) logout(w http.ResponseWriter, r *http.Request) {
	sess := session.GetSession(r)
	id := sess.ID()

	if err := sess.Destroy(w, r); err != nil {
		slog.Error("failed to destroy sessions", "id", id)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *UsersHandler) initiateOauthFlow(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	if provider == "" {
		http.Error(w, "provider is required", http.StatusBadRequest)
		return
	}
	if !domain.IsValidProvider(provider) {
		http.Error(w, "invalid provider", http.StatusBadRequest)
		addFlash(w, r, "invalid provider", flashTypeError)
		return
	}

	authURL, err := h.auth.InitiateOauthFlow(r.Context(), domain.OauthProvider(provider))
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

	isValid, provider, err := h.auth.VerifyOauthState(r.Context(), state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !isValid {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	user, err := h.auth.IdentifyOauthUser(r.Context(), code, provider)
	if err != nil {
		w.WriteHeader(ErrorStatus(err))
		addFlash(w, r, err.Error(), flashTypeError)
		return
	}

	if err := session.GetSession(r).Set("user", user); err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
