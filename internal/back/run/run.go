// Package run provides functionality to initialize and run the main application server.
package run

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Alekseyt9/upscaler/internal/back/config"
	"github.com/Alekseyt9/upscaler/internal/back/handler"
	"github.com/Alekseyt9/upscaler/internal/back/handler/middleware/jwtcheker"
	"github.com/Alekseyt9/upscaler/internal/back/handler/middleware/logger"
	"github.com/Alekseyt9/upscaler/internal/back/services/messagebroker"
	"github.com/Alekseyt9/upscaler/internal/back/services/store"
	"github.com/Alekseyt9/upscaler/internal/back/services/store/cache"
	"github.com/Alekseyt9/upscaler/internal/back/services/userserv"
	"github.com/Alekseyt9/upscaler/internal/back/services/websocket"
	"github.com/Alekseyt9/upscaler/internal/common/model"
	"github.com/Alekseyt9/upscaler/internal/common/services/s3store"
)

// Run initializes and starts the main application server, setting up the database,
// message broker, S3 storage, WebSocket service, user service, and HTTP server.
// It manages the server lifecycle and ensures a graceful shutdown on receiving termination signals.
//
// Parameters:
//   - cfg: The configuration settings for the application.
//   - log: A logger to capture and output log messages.
//
// Returns an error if any service fails to initialize or if there is an error during server execution.
func Run(cfg *config.Config, log *slog.Logger) error {
	pgstore, err := store.NewPostgresStore(context.Background(), cfg.PgDataBaseDSN, log)
	if err != nil {
		return fmt.Errorf("store.NewPostgresStore %w", err)
	}

	store, err := cache.NewCachedStore(pgstore, log)
	if err != nil {
		return fmt.Errorf("store.NewCachedStore %w", err)
	}

	s3, err := s3store.New(s3store.S3Options{
		AccessKeyID:     cfg.S3AccessKeyID,
		SecretAccessKey: cfg.S3SecretAccessKey,
		BucketName:      cfg.S3BucketName,
	})
	if err != nil {
		return fmt.Errorf("s3store.New %w", err)
	}

	ws := websocket.New(log)
	us := userserv.New(store, s3, ws)

	consumer, err := messagebroker.NewConsumer(us, log, model.BrokerOptions{
		Topic:         cfg.KafkaTopicResult,
		KafkaBrokers:  []string{cfg.KafkaAddr},
		ConsumerGroup: cfg.KafkeCunsumerGroup})
	if err != nil {
		return fmt.Errorf("messagebroker.NewConsumer %w", err)
	}

	producer, err := messagebroker.NewProducer(store, log, model.BrokerOptions{
		Topic:        cfg.KafkaTopic,
		KafkaBrokers: []string{cfg.KafkaAddr}})
	if err != nil {
		return fmt.Errorf("messagebroker.NewProducer %w", err)
	}

	httpRouter, err := Router(cfg, log, store, s3, us, ws)
	if err != nil {
		return fmt.Errorf("new Router %w", err)
	}

	server := &http.Server{
		Addr:    cfg.BackAddress,
		Handler: httpRouter,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Info("Server started", "url", cfg.BackAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed", "error", err)
			cancel()
		}
	}()

	<-stop
	log.Info("Shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Server shutdown failed", "error", err)
	}

	if err := store.Close(); err != nil {
		log.Error("Store shutdown failed", "error", err)
	}

	consumer.Close()
	producer.Close()

	log.Info("Server shutdown completed")

	return nil
}

// Router sets up the HTTP routes and applies middleware to the routes.
//
// Parameters:
//   - cfg: The configuration settings for the application.
//   - log: A logger for capturing log messages.
//   - store: The store interface for data handling.
//   - s3: The S3 store interface for file storage.
//   - us: The user service for handling user-related logic.
//   - ws: The WebSocket service for handling WebSocket connections.
//
// Returns:
//   - An http.Handler with all routes and middleware applied.
//   - An error if setting up handlers fails.
func Router(cfg *config.Config, log *slog.Logger, store store.Store, s3 s3store.S3Store,
	us userserv.UserService, ws websocket.WebSocketService) (http.Handler, error) {
	mux := http.NewServeMux()

	setupFileServer(mux, log)
	err := setupHandlers(mux, cfg, log, store, s3, us, ws)
	if err != nil {
		return nil, fmt.Errorf("setupHandlers %w", err)
	}
	handler := setupMiddlware(mux, log, cfg)

	return handler, nil
}

// setupMiddlware applies the necessary middleware to the HTTP handler chain.
//
// Parameters:
//   - h: The base HTTP handler to which middleware will be applied.
//   - log: A logger for capturing log messages.
//   - cfg: The configuration settings for the application.
//
// Returns:
//   - An http.Handler with middleware applied.
func setupMiddlware(h http.Handler, log *slog.Logger, cfg *config.Config) http.Handler {
	handler := logger.WithLogging(h, log)
	handler = jwtcheker.WithJWTCheck(handler, cfg.JWTSecret, log)
	return handler
}

// setupHandlers registers the HTTP routes with their corresponding handlers.
//
// Parameters:
//   - mux: The ServeMux to register the routes.
//   - cfg: The configuration settings for the application.
//   - log: A logger for capturing log messages.
//   - store: The store interface for data handling.
//   - s3: The S3 store interface for file storage.
//   - us: The user service for handling user-related logic.
//   - ws: The WebSocket service for handling WebSocket connections.
//
// Returns:
//   - An error if there is an issue setting up handlers.
func setupHandlers(mux *http.ServeMux, cfg *config.Config, log *slog.Logger, store store.Store,
	s3 s3store.S3Store, us userserv.UserService, ws websocket.WebSocketService) error {
	ho := handler.HandlerOptions{
		JWTSecret: cfg.JWTSecret,
	}

	h := handler.New(s3, log, store, ho, us, ws)
	mux.HandleFunc("/api/user/getuploadurls", h.GetUploadURLs)
	mux.HandleFunc("/api/user/completefilesupload", h.CompleteFilesUpload)
	mux.HandleFunc("/api/user/getstate", h.GetState)
	mux.HandleFunc("/api/auth/login", h.Login)
	mux.HandleFunc("/api/auth/login2", h.Login2)

	return nil
}

// setupFileServer sets up a file server to serve static files from a specified directory.
//
// Parameters:
//   - mux: The ServeMux to register the file server routes.
//   - log: A logger for capturing log messages.
func setupFileServer(mux *http.ServeMux, log *slog.Logger) {
	contentDir := filepath.Join("..", "..", "internal", "back", "content")
	log.Info("Serving files from", "contentDir", contentDir)

	fs := http.FileServer(http.Dir(contentDir))
	mux.Handle("/content/", http.StripPrefix("/content/", fs))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("..", "..", "internal", "back", "content", "index.html"))
	})
}
