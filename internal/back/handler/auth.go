package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

// POST
func (h *FrontHandler) Register(w http.ResponseWriter, r *http.Request) {
	// not implemented
}

// POST
func (h *FrontHandler) Login(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("jwt")
	if err != nil || cookie.Value == "" {
		_, err := h.createUserAndSetToken(w, r)
		if err != nil {
			h.log.Error("h.createUserAndSetToken", "error", err)
			http.Error(w, "h.createUserAndSetToken", http.StatusInternalServerError)
		}
		return
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.opt.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		_, err := h.createUserAndSetToken(w, r)
		if err != nil {
			h.log.Error("h.createUserAndSetToken", "error", err)
			http.Error(w, "h.createUserAndSetToken", http.StatusInternalServerError)
		}
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId := claims["userId"].(string)
		h.log.Info("User ID from token", "userID", userId)
	} else {
		_, err := h.createUserAndSetToken(w, r)
		if err != nil {
			h.log.Error("h.createUserAndSetToken", "error", err)
			http.Error(w, "h.createUserAndSetToken", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GET
func (h *FrontHandler) Login2(w http.ResponseWriter, r *http.Request) {
	var userID int64

	cookie, err := r.Cookie("jwt")
	if err != nil || cookie.Value == "" {
		userID, err = h.createUserAndSetToken(w, r)
		if err != nil {
			h.log.Error("h.createUserAndSetToken", "error", err)
			http.Error(w, "h.createUserAndSetToken", http.StatusInternalServerError)
			return
		}
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.opt.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		userID, err = h.createUserAndSetToken(w, r)
		if err != nil {
			h.log.Error("h.createUserAndSetToken", "error", err)
			http.Error(w, "h.createUserAndSetToken", http.StatusInternalServerError)
			return
		}
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr := claims["userId"].(string)
		userID, err = strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			h.log.Error("Ошибка при парсинге userId", "error", err)
			http.Error(w, "Ошибка при парсинге userId", http.StatusInternalServerError)
			return
		}
	} else {
		userID, err = h.createUserAndSetToken(w, r)
		if err != nil {
			h.log.Error("h.createUserAndSetToken", "error", err)
			http.Error(w, "h.createUserAndSetToken", http.StatusInternalServerError)
			return
		}
	}

	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("Ошибка при установлении WebSocket соединения", "error", err)
		http.Error(w, "Ошибка при установлении WebSocket соединения", http.StatusInternalServerError)
		return
	}

	userIDStr := strconv.FormatInt(userID, 10)
	h.ws.AddUser(userIDStr, conn)

	defer conn.Close()

	h.log.Info("Клиент подключен", "address", r.RemoteAddr)
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			h.log.Info("Ошибка при чтении сообщения", "error", err)
			break
		}
		h.log.Info("recieved message", "message", msg)
	}
	h.ws.RemoveUser(userIDStr)
	h.log.Info("Клиент отключен", "address", r.RemoteAddr)
}

func (h *FrontHandler) createUserAndSetToken(w http.ResponseWriter, r *http.Request) (int64, error) {
	id, err := h.store.CreateUser(r.Context())
	if err != nil {
		return 0, fmt.Errorf("store.CreateUser %w", err)
	}

	userId := strconv.FormatInt(id, 10)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userId,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.opt.JWTSecret))
	if err != nil {
		return 0, fmt.Errorf("oшибка генерации токена %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	return id, nil
}
