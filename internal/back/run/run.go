package run

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Alekseyt9/upscaler/internal/back/config"
	"github.com/Alekseyt9/upscaler/internal/back/handler"
	"github.com/Alekseyt9/upscaler/internal/back/handler/middleware/logger"
	"github.com/Alekseyt9/upscaler/internal/back/services/s3store"
)

func Run(cfg *config.Config) error {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	httpRouter := Router(cfg, log)

	log.Info("Server started", "url", cfg.BackAddress)
	err := http.ListenAndServe(cfg.BackAddress, httpRouter)
	if err != nil {
		return err
	}

	return nil
}

func Router(cfg *config.Config, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	setupFileServer(mux, log)
	setupHandlers(mux, cfg, log)
	setupMiddlware(mux, log)
	handler := setupMiddlware(mux, log)

	return handler
}

func setupMiddlware(h http.Handler, log *slog.Logger) http.Handler {
	handler := logger.WithLogging(h, log)
	return handler
}

func setupHandlers(mux *http.ServeMux, cfg *config.Config, log *slog.Logger) error {
	s3, err := s3store.New(s3store.S3Options{
		AccessKeyID:     cfg.S3AccessKeyID,
		SecretAccessKey: cfg.S3SecretAccessKey,
		BucketName:      cfg.S3BucketName,
	})
	if err != nil {
		return err
	}
	h := handler.New(s3, log)

	mux.HandleFunc("/api/getuploadurls", h.GetUploadURLs)
	mux.HandleFunc("/api/completefilesupload", h.CompleFilesUpload)
	mux.HandleFunc("/api/getstate", h.GetState)

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
