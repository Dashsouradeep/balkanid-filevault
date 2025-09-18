package api

import (
	"net/http"
	"strings"

	"github.com/Dashsouradeep/balkanid-filevault/backend/utils"
)

func AuthMiddleware(secret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "❌ Missing token", http.StatusUnauthorized)
			return
		}

		// Expect "Bearer <token>"
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := utils.ValidateJWT(tokenStr, secret)
		if err != nil || !token.Valid {
			http.Error(w, "❌ Invalid token", http.StatusUnauthorized)
			return
		}

		// Token valid → proceed
		next(w, r)
	}
}
