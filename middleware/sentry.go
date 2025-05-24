package middleware

import (
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
)

func SentryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			// Check the concurrency guide for more details: https://docs.sentry.io/platforms/go/concurrency/
			hub = sentry.CurrentHub().Clone()
			ctx = sentry.SetHubOnContext(ctx, hub)
		}

		user := UserFromContext(ctx)
		if user != nil {
			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetUser(sentry.User{
					ID:       user.GUID.String(),
					Email:    user.Email,
					Username: user.Name,
					Name:     user.Name,
					Data: map[string]string{
						"oauth_provider": string(user.Provider),
					},
				})
			})
		}

		options := []sentry.SpanOption{
			// Set the OP based on values from https://develop.sentry.dev/sdk/performance/span-operations/
			sentry.WithOpName("http.server"),
			sentry.ContinueFromRequest(r),
			sentry.WithTransactionSource(sentry.SourceURL),
		}
		transaction := sentry.StartTransaction(ctx,
			fmt.Sprintf("%s %s", r.Method, r.URL.Path),
			options...,
		)

		defer transaction.Finish()

		next.ServeHTTP(w, r.WithContext(transaction.Context()))
	})
}
