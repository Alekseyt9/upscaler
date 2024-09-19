// Package handler provides HTTP handlers for managing server-side operations.
package handler

import (
	"log/slog"

	"github.com/Alekseyt9/upscaler/internal/back/services/store"
	"github.com/Alekseyt9/upscaler/internal/back/services/userserv"
	"github.com/Alekseyt9/upscaler/internal/back/services/websocket"
	"github.com/Alekseyt9/upscaler/internal/common/services/s3store"
)

// ServerHandler is the main struct that handles server-side operations.
type ServerHandler struct {
	s3    s3store.S3Store             // S3 store for file handling operations.
	log   *slog.Logger                // Logger for capturing logs and error messages.
	store store.Store                 // Data store interface for interacting with the application's database.
	opt   HandlerOptions              // Options for the handler, such as the JWT secret.
	us    *userserv.UserService       // User service for managing user-related operations.
	ws    *websocket.WebSocketService // WebSocket service for handling WebSocket communications.
}

// HandlerOptions defines configuration options for ServerHandler.
type HandlerOptions struct {
	JWTSecret string // Secret key used for validating JWT tokens.
}

// New creates and returns a new instance of ServerHandler.
//
// Parameters:
//   - s3: Interface for interacting with the S3 storage system.
//   - log: Logger instance for capturing and outputting logs.
//   - store: Data store interface for managing application data.
//   - opt: HandlerOptions struct containing configuration details like the JWT secret.
//   - us: UserService for managing user-related operations and tasks.
//   - ws: WebSocketService for managing WebSocket connections and communication.
//
// Returns:
//   - A pointer to a newly created ServerHandler instance.
func New(s3 s3store.S3Store, log *slog.Logger, store store.Store,
	opt HandlerOptions, us *userserv.UserService, ws *websocket.WebSocketService) *ServerHandler {
	return &ServerHandler{
		s3:    s3,
		log:   log,
		store: store,
		opt:   opt,
		us:    us,
		ws:    ws,
	}
}
