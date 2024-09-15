package jwtcheker

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func WithJWTCheck(h http.Handler, JWTSecret string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/user/") {
			h.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("jwt")
		if err != nil || cookie.Value == "" {
			http.Error(w, "Токен не найден", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			return []byte(JWTSecret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Неверный или истекший токен", http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
