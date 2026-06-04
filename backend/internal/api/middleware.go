// Package api provides HTTP handlers and middleware.
package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/marvinf95/bitesense/backend/internal/auth"
)

type ctxKey int

const (
	ctxUserID ctxKey = iota
	ctxLocale
)

func UserIDFrom(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxUserID).(string)
	return v, ok
}

func LocaleFrom(ctx context.Context) string {
	if v, ok := ctx.Value(ctxLocale).(string); ok {
		return v
	}
	return "en"
}

// RequireAuth verifies the Bearer token and injects user_id + locale into the request context.
func RequireAuth(svc *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if !strings.HasPrefix(h, "Bearer ") {
				writeError(w, http.StatusUnauthorized, "missing bearer token")
				return
			}
			token := strings.TrimPrefix(h, "Bearer ")
			claims, err := svc.ParseAccess(token)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid token")
				return
			}
			ctx := context.WithValue(r.Context(), ctxUserID, claims.UserID)
			ctx = context.WithValue(ctx, ctxLocale, claims.Locale)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
