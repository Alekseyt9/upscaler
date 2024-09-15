package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// POST
func (h *FrontHandler) Register(w http.ResponseWriter, r *http.Request) {
	// not implemented
}

// POST
func (h *FrontHandler) Login(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("jwt")
	if err != nil || cookie.Value == "" {
		h.createUserAndSetToken(w)
		return
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.opt.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		h.createUserAndSetToken(w)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId := claims["userId"].(string)
		w.Write([]byte("User ID from token: " + userId))
	} else {
		h.createUserAndSetToken(w)
	}
}

func (h *FrontHandler) createUserAndSetToken(w http.ResponseWriter) {
	id, err := h.store.CreateUser()
	if err != nil {
		h.log.Error("store.CreateUser", "error", err)
		http.Error(w, "store.CreateUser", http.StatusInternalServerError)
		return
	}

	userId := strconv.FormatInt(id, 10)
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
