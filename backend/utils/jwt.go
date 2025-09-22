package utils

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var userIDKey = &struct{}{}

// WithUserID puts user_id into context
func WithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID retrieves user_id from context
func GetUserID(ctx context.Context) (int, bool) {
	val, ok := ctx.Value(userIDKey).(int)
	return val, ok
}

// GenerateJWT creates a JWT for a given user
func GenerateJWT(userID int, email, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateAndGetClaims parses and validates JWT, returns claims
func ValidateAndGetClaims(tokenStr, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("could not parse claims")
	}
	return claims, nil
}

// Backward compatibility wrapper
func ValidateJWT(tokenStr, secret string) (jwt.MapClaims, error) {
	return ValidateAndGetClaims(tokenStr, secret)
}

// ExtractToken takes "Bearer <token>" and returns "<token>"
func ExtractToken(authHeader string) string {
	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// Backward compatibility wrapper (old signature)
func GenerateJWTLegacy(userID int, email string) (string, error) {
	// You’ll need to pass your secret from env or config
	secret := "supersecret" // TODO: replace with os.Getenv("JWT_SECRET")
	return GenerateJWT(userID, email, secret)
}
