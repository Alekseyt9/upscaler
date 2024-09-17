package run

import (
	"context"
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
	"github.com/Alekseyt9/upscaler/internal/back/services/store"
	"github.com/Alekseyt9/upscaler/internal/back/services/userserv"
	"github.com/Alekseyt9/upscaler/internal/common/services/s3store"
)

func Run(cfg *config.Config) error {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	store, err := store.NewPostgresStore(context.Background(), cfg.PgDataBaseDSN, log)
	if err != nil {
		return err
	}

	s3, err := s3store.New(s3store.S3Options{
		AccessKeyID:     cfg.S3AccessKeyID,
		SecretAccessKey: cfg.S3SecretAccessKey,
		BucketName:      cfg.S3BucketName,
	})
	if err != nil {
		return err
	}

	httpRouter, err := Router(cfg, log, store, s3)
	if err != nil {
		return err
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

	log.Info("Server shutdown completed")

	return nil
}

func Router(cfg *config.Config, log *slog.Logger, store store.Store, s3 s3store.S3Store) (http.Handler, error) {
	mux := http.NewServeMux()

	setupFileServer(mux, log)
	err := setupHandlers(mux, cfg, log, store, s3)
	if err != nil {
		return nil, err
	}
	handler := setupMiddlware(mux, log, cfg)

	return handler, nil
}

func setupMiddlware(h http.Handler, log *slog.Logger, cfg *config.Config) http.Handler {
	handler := logger.WithLogging(h, log)
	handler = jwtcheker.WithJWTCheck(handler, cfg.JWTSecret, log)
	return handler
}

func setupHandlers(mux *http.ServeMux, cfg *config.Config, log *slog.Logger, store store.Store, s3 s3store.S3Store) error {
	s3, err := s3store.New(s3store.S3Options{
		AccessKeyID:     cfg.S3AccessKeyID,
		SecretAccessKey: cfg.S3SecretAccessKey,
		BucketName:      cfg.S3BucketName,
	})
	if err != nil {
		return err
	}

	ho := handler.HandlerOptions{
		JWTSecret: cfg.JWTSecret,
	}
	us := userserv.New(store, s3)

	h := handler.New(s3, log, store, ho, us)
	mux.HandleFunc("/api/user/getuploadurls", h.GetUploadURLs)
	mux.HandleFunc("/api/user/completefilesupload", h.CompleteFilesUpload)
	mux.HandleFunc("/api/user/getstate", h.GetState)
	mux.HandleFunc("/api/auth/login", h.Login)

	return nil
}

func setupFileServer(mux *http.ServeMux, log *slog.Logger) {
	contentDir := filepath.Join("..", "..", "internal", "back", "content")
	log.Info("Serving files from", "contentDir", contentDir)

	fs := http.FileServer(http.Dir(contentDir))
	mux.Handle("/content/", http.StripPrefix("/content/", fs))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("..", "..", "internal", "back", "content", "index.html"))
	})
}
