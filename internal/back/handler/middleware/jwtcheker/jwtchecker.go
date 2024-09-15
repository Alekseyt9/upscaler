package jwtcheker

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDContextKey = contextKey("userID")

func WithJWTCheck(h http.Handler, JWTSecret string, log *slog.Logger) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/user/") {
			h.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("jwt")
		if err != nil || cookie.Value == "" {
			log.Error("WithJWTCheck", "error", "Токен не найден")
			http.Error(w, "Токен не найден", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			return []byte(JWTSecret), nil
		})
		if err != nil || !token.Valid {
			log.Error("WithJWTCheck", "error", "Неверный или истекший токен")
			http.Error(w, "Неверный или истекший токен", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			log.Error("WithJWTCheck", "error", "Неверный токен")
			http.Error(w, "Неверный токен", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["userId"].(string)
		if !ok {
			log.Error("WithJWTCheck", "error", "userId отсутствует в токене")
			http.Error(w, "userId отсутствует в токене", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func GetUserID(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(userIDContextKey).(string)
	return userID, ok
}
