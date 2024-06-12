package middleware

import (
	"context"
	"net/http"

	"gitea.com/go-chi/session"

	"github.com/zaibon/shortcut/domain"
)

type contextKey string

var contextUser = contextKey("user")

func UserFromContext(ctx context.Context) *domain.User {
	if user, ok := ctx.Value(contextUser).(*domain.User); ok {
		return user
	}
	return nil
}

func WithUser(ctx context.Context, user domain.User) context.Context {
	return context.WithValue(ctx, contextUser, &user)
}

func UserContext(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		sess := session.GetSession(r)
		user, ok := sess.Get("user").(*domain.User)
		if ok {
			r = r.WithContext(WithUser(r.Context(), *user))
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func Authenticated(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		sess := session.GetSession(r)
		user, ok := sess.Get("user").(*domain.User)
		if ok && user != nil {
			next.ServeHTTP(w, r)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}

	return http.HandlerFunc(fn)
}
