package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Alekseyt9/upscaler/internal/back/services/store"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

type MockWebSocketService struct {
}

func (m *MockWebSocketService) AddUser(userID string, conn *websocket.Conn) {
}

func (m *MockWebSocketService) RemoveUser(userID string) {
}

func (s *MockWebSocketService) Send(userID, message string) error {
	return nil
}

func TestLogin(t *testing.T) {
	mockWebSocket := &MockWebSocketService{}
	handlerOptions := HandlerOptions{
		JWTSecret: "mysecret",
	}

	memStore := store.NewMemoryStore()

	h := ServerHandler{
		store: memStore,
		ws:    mockWebSocket,
		opt:   handlerOptions,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": "123",
		"exp":    time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(handlerOptions.JWTSecret))
	cookie := &http.Cookie{Name: "jwt", Value: tokenString}

	req := httptest.NewRequest("POST", "/login", nil)
	req.AddCookie(cookie)
	w := httptest.NewRecorder()

	h.Login(w, req)
	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest("POST", "/login", nil)
	w = httptest.NewRecorder()

	h.Login(w, req)
	resp = w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotEmpty(t, resp.Cookies())
}

func TestLogin2(t *testing.T) {
	mockWebSocket := new(MockWebSocketService)
	handlerOptions := HandlerOptions{
		JWTSecret: "mysecret",
	}

	memStore := store.NewMemoryStore()

	h := ServerHandler{
		store: memStore,
		ws:    mockWebSocket,
		opt:   handlerOptions,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": "123",
		"exp":    time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(handlerOptions.JWTSecret))
	cookie := &http.Cookie{Name: "jwt", Value: tokenString}

	req := httptest.NewRequest("GET", "/login2", nil)
	req.AddCookie(cookie)
	w := httptest.NewRecorder()

	/*
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}

		conn, _, err := upgrader.Upgrade(w, req, nil)
		require.NoError(t, err)
	*/

	h.Login2(w, req)
	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest("GET", "/login2", nil)
	w = httptest.NewRecorder()

	h.Login2(w, req)
	resp = w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
