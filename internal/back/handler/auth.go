package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

// Login handles the POST request for user login.
// It checks the JWT token from the cookie and validates it.
// If the token is missing, invalid, or expired, a new user is created,
// and a new JWT token is generated and set as a cookie.
func (h *ServerHandler) Login(w http.ResponseWriter, r *http.Request) {
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

// Login2 handles the GET request for user login and sets up a WebSocket connection.
// It checks the JWT token, validates it, and establishes a WebSocket connection if successful.
// If the JWT token is missing or invalid, it generates a new one.
func (h *ServerHandler) Login2(w http.ResponseWriter, r *http.Request) {
	var cookie *http.Cookie

	cookie, err := r.Cookie("jwt")
	if err != nil || cookie.Value == "" {
		cookie, err = h.createUserAndSetToken(w, r)
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
		cookie, err = h.createUserAndSetToken(w, r)
		if err != nil {
			h.log.Error("h.createUserAndSetToken", "error", err)
			http.Error(w, "h.createUserAndSetToken", http.StatusInternalServerError)
			return
		}
	}

	var userID int64
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr := claims["userId"].(string)
		userID, err = strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			h.log.Error("Error parsing userId", "error", err)
			http.Error(w, "Error parsing userId", http.StatusInternalServerError)
			return
		}
	}

	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	responseHeader := http.Header{}
	responseHeader.Add("Set-Cookie", cookie.String())

	conn, err := upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		h.log.Error("Error establishing WebSocket connection", "error", err)
		http.Error(w, "Error establishing WebSocket connection", http.StatusInternalServerError)
		return
	}

	userIDStr := strconv.FormatInt(userID, 10)
	h.ws.AddUser(userIDStr, conn)

	defer conn.Close()

	h.log.Info("Client connected", "address", r.RemoteAddr)
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			h.log.Info("Error reading message", "error", err)
			break
		}
		h.log.Info("Received message", "message", msg)
	}
	h.ws.RemoveUser(userIDStr)
	h.log.Info("Client disconnected", "address", r.RemoteAddr)
}

// createUserAndSetToken creates a new user in the system, generates a JWT token,
// and sets it as a cookie in the response.
func (h *ServerHandler) createUserAndSetToken(w http.ResponseWriter, r *http.Request) (*http.Cookie, error) {
	id, err := h.store.CreateUser(r.Context())
	if err != nil {
		return nil, fmt.Errorf("store.CreateUser %w", err)
	}

	userId := strconv.FormatInt(id, 10)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userId,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.opt.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("error generating token %w", err)
	}

	c := &http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, c)

	return c, nil
}
