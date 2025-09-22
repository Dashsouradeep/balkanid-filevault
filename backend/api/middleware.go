package api

import (
	"net/http"
	"strings"

	"github.com/Dashsouradeep/balkanid-filevault/backend/utils"
)

// AuthMiddleware validates JWT and attaches user_id to request context
func AuthMiddleware(next http.Handler, secret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "❌ Missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "❌ Invalid Authorization format", http.StatusUnauthorized)
			return
		}
		tokenStr := parts[1]

		claims, err := utils.ValidateAndGetClaims(tokenStr, secret)
		if err != nil {
			http.Error(w, "❌ Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// put user_id into context
		userID, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "❌ Invalid token payload", http.StatusUnauthorized)
			return
		}

		ctx := utils.WithUserID(r.Context(), int(userID))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
