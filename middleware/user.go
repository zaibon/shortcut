package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/google/uuid"

	"github.com/zaibon/shortcut/domain"
)

type contextKey string

var contextUser = contextKey("user")

type UserFetcher interface {
	GetUser(ctx context.Context, guid domain.GUID) (*domain.User, error)
}

func UserFromContext(ctx context.Context) *domain.User {
	if user, ok := ctx.Value(contextUser).(*domain.User); ok {
		return user
	}
	return nil
}

func WithUser(ctx context.Context, user domain.User) context.Context {
	return context.WithValue(ctx, contextUser, &user)
}

func UserContext(sessionManager *scs.SessionManager, store UserFetcher) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			userIDStr := sessionManager.GetString(r.Context(), "user_id")
			if userIDStr == "" {
				next.ServeHTTP(w, r)
				return
			}

			guid, err := uuid.Parse(userIDStr)
			if err != nil {
				slog.Error("failed to parse user id from session", "error", err)
				sessionManager.Remove(r.Context(), "user_id")
				next.ServeHTTP(w, r)
				return
			}

			user, err := store.GetUser(r.Context(), domain.GUID(guid))
			if err != nil {
				// User not found or DB error
				slog.Warn("failed to fetch user from session id", "guid", guid, "error", err)
				sessionManager.Remove(r.Context(), "user_id")
				next.ServeHTTP(w, r)
				return
			}

			// Populate context with user, regardless of status
			ctx := WithUser(r.Context(), *user)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

func Authenticated(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context())
		if user == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
