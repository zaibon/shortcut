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
	ListConnectedProvider(ctx context.Context, userID domain.GUID) ([]domain.AccountProvider, error)
	Delete(ctx context.Context, guid domain.GUID) error
}

type URLUsageService interface {
	CountMonthlyURL(ctx context.Context, authorID domain.ID) (int64, error)
}

type UsersHandler struct {
	htmx         *htmx.HTMX
	auth         AuthService
	stripe       stripeService
	urlService   URLUsageService
	stripePubKey string
}

func NewUsersHandler(svc AuthService, stripe stripeService, urlService URLUsageService, stripePubKey string) *UsersHandler {
	return &UsersHandler{
		htmx:         htmx.New(),
		auth:         svc,
		stripe:       stripe,
		urlService:   urlService,
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
		r.Delete("/account", h.deleteAccount)
	})
}

func (h *UsersHandler) authPage(w http.ResponseWriter, r *http.Request) {
	templates.AuthPage().Render(r.Context(), w)
}

func (h *UsersHandler) myAccount(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	linkedProviders, err := h.auth.ListConnectedProvider(r.Context(), user.GUID)
	if err != nil {
		slog.Error("failed to get connected provider", "error", err)
		http.Error(w, "failed to get connected provider", http.StatusInternalServerError)
		return
	}

	count, err := h.urlService.CountMonthlyURL(r.Context(), user.ID)
	if err != nil {
		slog.Error("failed to get url count", "error", err)
	}

	limit := domain.FreePlanLimit
	stats := domain.SubscriptionStats{
		PlanName:        "Free",
		URLUsage:        int(count),
		URLLimit:        limit,
		Remaining:       limit - int(count),
		UsagePercentage: int((float64(count) / float64(limit)) * 100),
	}
	if stats.UsagePercentage > 100 {
		stats.UsagePercentage = 100
	}
	if stats.Remaining < 0 {
		stats.Remaining = 0
	}

	templates.AccountPage(*user, linkedProviders, stats).Render(r.Context(), w)
}

func (h *UsersHandler) deleteAccount(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if err := h.auth.Delete(r.Context(), user.GUID); err != nil {
		slog.Error("failed to delete user", "error", err)
		http.Error(w, "failed to delete account", http.StatusInternalServerError)
		return
	}

	sess := session.GetSession(r)
	if err := sess.Destroy(w, r); err != nil {
		slog.Error("failed to destroy sessions", "id", sess.ID())
	}

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
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
