// Package jwtcheker provides middleware for verifying JWT tokens in HTTP requests.
package jwtcheker

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDContextKey = contextKey("userID")

// WithJWTCheck is middleware that checks the JWT token in the request cookie.
// If the token is valid, it extracts the user ID from the token and stores it in the request context.
//
// Parameters:
//   - h: The HTTP handler to be wrapped by the middleware.
//   - JWTSecret: The secret key used to validate the JWT token.
//   - log: Logger for capturing and logging errors.
//
// Returns:
//   - An HTTP handler that performs the JWT check and calls the next handler if the token is valid.
func WithJWTCheck(h http.Handler, JWTSecret string, log *slog.Logger) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/user/") {
			h.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("jwt")
		if err != nil || cookie.Value == "" {
			log.Error("WithJWTCheck", "error", "Token not found")
			http.Error(w, "Token not found", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			return []byte(JWTSecret), nil
		})
		if err != nil || !token.Valid {
			log.Error("WithJWTCheck", "error", "Invalid or expired token")
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			log.Error("WithJWTCheck", "error", "Invalid token")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["userId"].(string)
		if !ok {
			log.Error("WithJWTCheck", "error", "userId not found in token")
			http.Error(w, "userId not found in token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// GetUserID extracts the user ID from the request context.
//
// Parameters:
//   - r: The HTTP request from which to extract the user ID.
//
// Returns:
//   - The user ID as int64 and a boolean indicating if the extraction was successful.
func GetUserID(r *http.Request) (int64, bool) {
	userIDStr, ok := r.Context().Value(userIDContextKey).(string)
	if !ok {
		return 0, false
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return 0, false
	}

	return userID, true
}
