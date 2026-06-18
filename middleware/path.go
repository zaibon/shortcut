package middleware

import (
	"context"
	"net/http"
)

type pathContextKey string

var contextPath = pathContextKey("path")

func PathFromContext(ctx context.Context) string {
	if path, ok := ctx.Value(contextPath).(string); ok {
		return path
	}
	return ""
}

func WithPath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, contextPath, path)
}

func PathContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := WithPath(r.Context(), r.URL.Path)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
