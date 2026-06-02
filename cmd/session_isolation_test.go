package main

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
)

// spyStore implements scs.Store and counts Find calls to detect session lookups.
type spyStore struct {
	findCalls atomic.Int64
}

func (s *spyStore) Find(_ string) ([]byte, bool, error) {
	s.findCalls.Add(1)
	return nil, false, nil
}

func (s *spyStore) Commit(_ string, _ []byte, _ time.Time) error { return nil }
func (s *spyStore) Delete(_ string) error                        { return nil }

// buildTestRouter constructs a router with the same middleware split used in
// production: redirect hot path has no session middleware; everything else does.
func buildTestRouter(sm *scs.SessionManager) http.Handler {
	r := chi.NewRouter()

	// redirect hot path — no session store round-trip
	r.Group(func(r chi.Router) {
		r.Use(chiMiddleware.Recoverer)
		r.Get("/{shortID}", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://example.com", http.StatusFound)
		})
	})

	// session-required routes
	r.Group(func(r chi.Router) {
		r.Use(chiMiddleware.Recoverer)
		r.Use(sm.LoadAndSave)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		r.Post("/shorten", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	return r
}

// TestRedirectSkipsSessionStore verifies that GET /{shortID} never touches the
// session store, while other routes that need session support still do.
func TestRedirectSkipsSessionStore(t *testing.T) {
	spy := &spyStore{}
	sm := scs.New()
	sm.Store = spy
	sm.Cookie.Name = "test_session"

	router := buildTestRouter(sm)

	// A session cookie that would cause LoadAndSave to call Find if it were in
	// the middleware chain for the redirect route.
	sessionCookie := &http.Cookie{Name: "test_session", Value: "some-token-value"}

	t.Run("redirect does not query session store", func(t *testing.T) {
		before := spy.findCalls.Load()

		req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
		req.AddCookie(sessionCookie)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, before, spy.findCalls.Load(),
			"session store must not be queried for redirect requests (hot path latency)")
	})

	t.Run("index route queries session store", func(t *testing.T) {
		before := spy.findCalls.Load()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(sessionCookie)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Greater(t, spy.findCalls.Load(), before,
			"session store must be queried for routes that need authentication context")
	})

	t.Run("shorten route queries session store", func(t *testing.T) {
		before := spy.findCalls.Load()

		req := httptest.NewRequest(http.MethodPost, "/shorten", nil)
		req.AddCookie(sessionCookie)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Greater(t, spy.findCalls.Load(), before,
			"session store must be queried for authenticated routes")
	})
}
