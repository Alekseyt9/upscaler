// Package websocket provides functionality for managing WebSocket connections,
// including adding and removing users and sending messages to connected clients.
package websocket

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocketService defines the interface for managing WebSocket connections.
type WebSocketService interface {
	// AddUser adds a new user with a WebSocket connection.
	AddUser(userID string, conn *websocket.Conn)

	// RemoveUser removes a user from the service and closes their WebSocket connection.
	RemoveUser(userID string)

	// Send sends a message to a user by their userID.
	// Returns an error if the message could not be sent.
	Send(userID, message string) error
}

// WebSocketService manages WebSocket connections for users.
type WebSocketServiceImpl struct {
	Users map[string]*WSUser // Map of user IDs to their WebSocket connections.
	mu    sync.Mutex         // Mutex for synchronizing access to the Users map.
	log   *slog.Logger       // Logger for logging information and errors.
}

// WSUser represents a user with a WebSocket connection.
type WSUser struct {
	ID   string          // User's unique identifier.
	Conn *websocket.Conn // WebSocket connection for the user.
}

// NewWebSocketService creates and returns a new WebSocketService instance.
// It initializes the Users map and sets up logging.
//
// Parameters:
//   - log: Logger for capturing logs and errors.
//
// Returns:
//   - A pointer to the newly created WebSocketService.
func New(log *slog.Logger) *WebSocketServiceImpl {
	return &WebSocketServiceImpl{
		Users: make(map[string]*WSUser),
		log:   log,
	}
}

// AddUser adds a new user to the WebSocketService.
// It associates the user's ID with their WebSocket connection.
//
// Parameters:
//   - userID: The unique identifier for the user.
//   - conn: The WebSocket connection to be associated with the user.
func (s *WebSocketServiceImpl) AddUser(userID string, conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Users[userID] = &WSUser{
		ID:   userID,
		Conn: conn,
	}
}

// RemoveUser removes a user from the WebSocketService by their userID.
// It closes the user's WebSocket connection and deletes their entry from the Users map.
//
// Parameters:
//   - userID: The unique identifier of the user to be removed.
func (s *WebSocketServiceImpl) RemoveUser(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Users, userID)
}

// Send sends a message to a user by their userID.
// It writes a message to the user's WebSocket connection.
//
// Parameters:
//   - userID: The unique identifier of the user to whom the message is being sent.
//   - message: The message to be sent to the user.
//
// Returns:
//   - An error if the message could not be sent or if the user is not connected.
func (s *WebSocketServiceImpl) Send(userID, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if user, ok := s.Users[userID]; ok {
		err := user.Conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			return fmt.Errorf("error sending message %w", err)
		}
	}

	return nil
}
