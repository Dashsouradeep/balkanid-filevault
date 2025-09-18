package api

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Same secret you used in auth.go
var jwtSecret = []byte("supersecretkey")

// Extract user_id from JWT in Authorization header
func GetUserIDFromToken(r *http.Request) (int, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, false
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0, false
	}

	tokenStr := parts[1]
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, false
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, false
	}

	return int(userID), true
}
