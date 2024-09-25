package handler

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Alekseyt9/upscaler/internal/back/services/store"
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
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	memStore := store.NewMemoryStore()

	h := ServerHandler{
		store: memStore,
		ws:    mockWebSocket,
		opt:   handlerOptions,
		log:   log,
	}

	req := httptest.NewRequest("POST", "/login", nil)
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
	handlerOptions := HandlerOptions{
		JWTSecret: "mysecret",
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	memStore := store.NewMemoryStore()

	h := ServerHandler{
		store: memStore,
		ws:    &MockWebSocketService{},
		opt:   handlerOptions,
		log:   log,
	}

	server := httptest.NewServer(http.HandlerFunc(h.Login2))
	defer server.Close()

	url := "ws" + server.URL[len("http"):] + "/login2"
	dialer := websocket.Dialer{}

	header := http.Header{}
	conn, resp, err := dialer.Dial(url, header)
	require.NoError(t, err)
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

	defer conn.Close()
	err = conn.WriteMessage(websocket.TextMessage, []byte("test message"))
	require.NoError(t, err)
}
