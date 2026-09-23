package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Akarshit7/competition-platform-go/internal/auth"
	"github.com/Akarshit7/competition-platform-go/internal/config"
	"github.com/Akarshit7/competition-platform-go/internal/response"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

func AuthMiddleware(cfg config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			response.Error(w, http.StatusUnauthorized, "missing authorization header")
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(w, http.StatusUnauthorized, "invalid authorization header")
			return
		}
		tokenStr := strings.TrimSpace(parts[1])
		claims, err := auth.VerifyToken(tokenStr, cfg.JwtTokenSecret)
		if err != nil {
			// also try admin token? admin uses same secret but different expiry claim, still same secret, so first check suffices
			response.Error(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminMiddleware checks role == admin. Requires AuthMiddleware before it, but can also verify admin token expiry secret if different.
func AdminMiddleware(cfg config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			response.Error(w, http.StatusUnauthorized, "missing authorization header")
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(w, http.StatusUnauthorized, "invalid authorization header")
			return
		}
		tokenStr := strings.TrimSpace(parts[1])
		claims, err := auth.VerifyToken(tokenStr, cfg.JwtTokenSecret)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if claims.Role != "admin" {
			response.Error(w, http.StatusForbidden, "admin access required")
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(r *http.Request) string {
	if v, ok := r.Context().Value(UserIDKey).(string); ok {
		return v
	}
	return ""
}

func GetRole(r *http.Request) string {
	if v, ok := r.Context().Value(RoleKey).(string); ok {
		return v
	}
	return ""
}
