package middleware

import (
	"net/http"

	"github.com/zaibon/shortcut/db/datastore"
)

func IsAdmin(db datastore.Querier) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())
			if user == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			isAdmin, err := db.IsAdmin(r.Context(), user.GUID.PgType())
			if err != nil || !isAdmin {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
