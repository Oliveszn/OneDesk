package middleware

import (
	"net/http"
	"strings"

	"github.com/Oliveszn/OneDesk/internal/reqctx"
	"github.com/Oliveszn/OneDesk/internal/token"
)

func Auth(jwtService *token.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			tokenString := strings.TrimPrefix(header, "Bearer ")

			claims, err := jwtService.Parse(tokenString)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := reqctx.WithAuth(r.Context(), claims.UserID, claims.TenantID, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
