package websocket

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocketService управляет WebSocket-соединениями
type WebSocketService struct {
	Users map[string]*WSUser
	mu    sync.Mutex
	log   *slog.Logger
}

type WSUser struct {
	ID   string
	Conn *websocket.Conn
}

// NewWebSocketService создает новый WebSocketService
func New(log *slog.Logger) *WebSocketService {
	return &WebSocketService{
		Users: make(map[string]*WSUser),
		log:   log,
	}
}

// AddUser добавляет пользователя в сервис
func (s *WebSocketService) AddUser(userID string, conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Users[userID] = &WSUser{
		ID:   userID,
		Conn: conn,
	}
}

// RemoveUser удаляет пользователя из сервиса
func (s *WebSocketService) RemoveUser(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Users, userID)
}

// Send отправляет сообщение пользователю по его ID
func (s *WebSocketService) Send(userID, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if user, ok := s.Users[userID]; ok {
		err := user.Conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			return fmt.Errorf("oшибка при отправке сообщения %w", err)
		}
	}

	return nil
}
