package handler

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// POST
func (h *FrontHandler) Register(w http.ResponseWriter, r *http.Request) {

}

// POST
func (h *FrontHandler) Login(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("jwt")
	if err != nil || cookie.Value == "" {
		h.generateAndSetToken(w, "some_user_id")
		return
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.opt.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		h.generateAndSetToken(w, "some_user_id")
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId := claims["userId"].(string)
		w.Write([]byte("User ID from token: " + userId))
	} else {
		h.generateAndSetToken(w, "some_user_id")
	}
}

func (h *FrontHandler) generateAndSetToken(w http.ResponseWriter, userId string) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userId,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.opt.JWTSecret))
	if err != nil {
		http.Error(w, "Ошибка генерации токена", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	w.Write([]byte("Новый токен сгенерирован и установлен в куки"))
}
